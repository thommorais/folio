package folio_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	_ "github.com/pocketbase/pocketbase/migrations"

	folio "folio/folio-core"
	"folio/folio-core/adapters/httpapi"
	"folio/folio-core/adapters/pb"
)

func record(t *testing.T, app core.App, collection string, values map[string]any) *core.Record {
	t.Helper()
	c, err := app.FindCollectionByNameOrId(collection)
	if err != nil {
		t.Fatal(err)
	}
	r := core.NewRecord(c)
	for k, v := range values {
		r.Set(k, v)
	}
	if err := app.Save(r); err != nil {
		t.Fatalf("save %s: %v", collection, err)
	}
	return r
}

func TestShareLinkOpensWithoutAnAccount(t *testing.T) {
	dir, err := os.MkdirTemp("", "folio-share")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: dir})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	if err := folio.Migrate(app); err != nil {
		t.Fatal(err)
	}

	owner := record(t, app, pb.ColUsers, map[string]any{"email": "owner@test.local", "password": "password12345", "verified": true})
	client := record(t, app, pb.ColClients, map[string]any{"slug": "acme", "name": "Acme"})
	dom := record(t, app, pb.ColDomains, map[string]any{"client": client.Id, "slug": "web", "name": "Web"})
	project := record(t, app, pb.ColProjects, map[string]any{"slug": "redesign", "name": "Redesign", "domain": dom.Id})
	record(t, app, pb.ColMembers, map[string]any{"domain": dom.Id, "project": project.Id, "user": owner.Id, "role": "owner"})
	ticket := record(t, app, pb.ColIssues, map[string]any{
		"domain": dom.Id, "project": project.Id, "kind": "ticket", "slug": "auth", "title": "Auth",
		"body": "use passkeys", "status": "open", "priority": "high", "created_by": owner.Id, "assignee": owner.Id,
	})

	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	httpapi.New(folio.New(app, nil).Deps()).Mount(&core.ServeEvent{App: app, Router: router})
	mux, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	ownerToken, err := owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}

	do := func(method, path, auth, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	if res := do(http.MethodPost, "/api/folio/issues/"+ticket.Id+"/shares", "", `{"label":"vendor"}`); res.Code != http.StatusUnauthorized {
		t.Errorf("anonymous share = %d, want 401", res.Code)
	}
	res := do(http.MethodPost, "/api/folio/issues/"+ticket.Id+"/shares", ownerToken, `{"label":"vendor"}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("share = %d %s", res.Code, res.Body)
	}
	var share struct {
		ID    string `json:"id"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &share); err != nil {
		t.Fatal(err)
	}

	res = do(http.MethodGet, httpapi.SharePath+"/"+share.Token, "", "")
	if res.Code != http.StatusOK {
		t.Fatalf("open = %d %s", res.Code, res.Body)
	}
	body := res.Body.String()
	if !strings.Contains(body, "use passkeys") || !strings.Contains(body, `"kind":"issue"`) {
		t.Errorf("opened share lacks the ticket: %s", body)
	}
	for _, leaked := range []string{owner.Id, project.Id, share.Token} {
		if strings.Contains(body, leaked) {
			t.Errorf("opened share exposes %q: %s", leaked, body)
		}
	}
	if got := res.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}

	if res := do(http.MethodDelete, "/api/folio/shares/"+share.ID, ownerToken, ""); res.Code != http.StatusNoContent {
		t.Fatalf("revoke = %d %s", res.Code, res.Body)
	}
	if res := do(http.MethodGet, httpapi.SharePath+"/"+share.Token, "", ""); res.Code != http.StatusNotFound {
		t.Errorf("revoked link = %d, want 404", res.Code)
	}
}
