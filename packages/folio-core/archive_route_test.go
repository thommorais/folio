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

	folio "folio/folio-core"
	"folio/folio-core/adapters/httpapi"
	"folio/folio-core/adapters/pb"
)

func TestArchiveAndRestoreOverTheAPI(t *testing.T) {
	dir, err := os.MkdirTemp("", "folio-archive")
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
	record(t, app, pb.ColProjectGrants, map[string]any{"project": project.Id, "user": owner.Id, "role": "owner"})

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

	listed := func(query string) []string {
		t.Helper()
		res := do(http.MethodGet, "/api/folio/projects/redesign/issues"+query, "")
		if res.Code != http.StatusOK {
			t.Fatalf("list%s = %d %s", query, res.Code, res.Body)
		}
		var body struct {
			Issues []struct {
				Slug     string `json:"slug"`
				Archived bool   `json:"archived"`
			} `json:"issues"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		out := make([]string, 0, len(body.Issues))
		for _, i := range body.Issues {
			out = append(out, i.Slug)
		}
		return out
	}

	var created struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	res := do(http.MethodPost, "/api/folio/projects/redesign/issues", `{"kind":"ticket","title":"Old plan"}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("create = %d %s", res.Code, res.Body)
	}
	if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	if res := do(http.MethodPatch, "/api/folio/issues/"+created.ID, `{"archived":true}`); res.Code != http.StatusOK {
		t.Fatalf("archive = %d %s", res.Code, res.Body)
	}
	if got := listed(""); len(got) != 0 {
		t.Errorf("default list after archive = %v, want none", got)
	}
	if got := listed("?archived=only"); len(got) != 1 || got[0] != created.Slug {
		t.Errorf("archived list = %v, want [%s]", got, created.Slug)
	}
	if got := listed("?archived=any"); len(got) != 1 {
		t.Errorf("any list = %v, want one", got)
	}

	if res := do(http.MethodPatch, "/api/folio/issues/"+created.ID, `{"archived":false}`); res.Code != http.StatusOK {
		t.Fatalf("restore = %d %s", res.Code, res.Body)
	}
	if got := listed(""); len(got) != 1 {
		t.Errorf("default list after restore = %v, want one", got)
	}

	if res := do(http.MethodGet, "/api/folio/projects/redesign/issues?archived=nope", ""); res.Code != http.StatusBadRequest {
		t.Errorf("bad archived value = %d, want 400", res.Code)
	}
}
