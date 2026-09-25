package folio_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	folio "folio/folio-core"
	"folio/folio-core/adapters/httpapi"
	"folio/folio-core/adapters/pb"
)

func TestResolutionClosesADecisionOverTheAPI(t *testing.T) {
	dir, err := os.MkdirTemp("", "folio-resolution")
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

	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	httpapi.New(folio.New(app, nil).Deps()).Mount(&core.ServeEvent{App: app, Router: router})
	mux, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	token, err := owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}

	do := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	var ticket struct {
		ID string `json:"id"`
	}
	res := do(http.MethodPost, "/api/folio/projects/redesign/issues", `{"kind":"ticket","title":"Tree or graph","wayfinder":"grilling"}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("create ticket = %d %s", res.Code, res.Body)
	}
	if err := json.Unmarshal(res.Body.Bytes(), &ticket); err != nil {
		t.Fatal(err)
	}

	if res := do(http.MethodPatch, "/api/folio/issues/"+ticket.ID, `{"status":"done"}`); res.Code != http.StatusBadRequest {
		t.Fatalf("close without a resolution = %d %s, want 400", res.Code, res.Body)
	}

	var entry struct {
		ID   string `json:"id"`
		Kind string `json:"kind"`
	}
	res = do(http.MethodPost, "/api/folio/projects/redesign/entries",
		fmt.Sprintf(`{"kind":"resolution","issue_id":%q,"body":"The tree hides blockers."}`, ticket.ID))
	if res.Code != http.StatusCreated {
		t.Fatalf("write resolution = %d %s", res.Code, res.Body)
	}
	if err := json.Unmarshal(res.Body.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Kind != "resolution" {
		t.Fatalf("entry kind = %q, want resolution", entry.Kind)
	}

	res = do(http.MethodPatch, "/api/folio/issues/"+ticket.ID,
		fmt.Sprintf(`{"status":"done","resolution":"A graph.","resolution_entry_id":%q}`, entry.ID))
	if res.Code != http.StatusOK {
		t.Fatalf("close with a resolution = %d %s", res.Code, res.Body)
	}

	var closed struct {
		Status          string `json:"status"`
		Resolution      string `json:"resolution"`
		ResolutionEntry string `json:"resolution_entry_id"`
	}
	if err := json.Unmarshal(do(http.MethodGet, "/api/folio/issues/"+ticket.ID, "").Body.Bytes(), &closed); err != nil {
		t.Fatal(err)
	}
	if closed.Status != "done" || closed.Resolution != "A graph." || closed.ResolutionEntry != entry.ID {
		t.Fatalf("got %+v", closed)
	}

	res = do(http.MethodPost, "/api/folio/projects/redesign/issues",
		`{"kind":"ticket","title":"Embed a graph library","wayfinder":"task","status":"cancelled","resolution":"Out of scope."}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("create a closed decision with a resolution = %d %s", res.Code, res.Body)
	}
}
