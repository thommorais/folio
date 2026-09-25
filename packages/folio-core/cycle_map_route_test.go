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

func TestCycleMapOverTheAPI(t *testing.T) {
	dir, err := os.MkdirTemp("", "folio-cycle-map")
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
	idOf := func(res *httptest.ResponseRecorder, want int) string {
		t.Helper()
		if res.Code != want {
			t.Fatalf("status = %d %s, want %d", res.Code, res.Body, want)
		}
		var out struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out.ID
	}

	work := idOf(do(http.MethodPost, "/api/folio/projects/redesign/issues", `{"kind":"ticket","title":"Wayfinder view"}`), http.StatusCreated)
	theMap := idOf(do(http.MethodPost, "/api/folio/projects/redesign/issues",
		fmt.Sprintf(`{"kind":"ticket","title":"Plan the view","wayfinder":"map","parent_id":%q}`, work)), http.StatusCreated)
	idOf(do(http.MethodPost, "/api/folio/projects/redesign/issues",
		fmt.Sprintf(`{"kind":"ticket","title":"Tree or graph","wayfinder":"grilling","parent_id":%q}`, theMap)), http.StatusCreated)
	cycle := idOf(do(http.MethodPost, "/api/folio/issues/"+work+"/cycles", ``), http.StatusCreated)

	res := do(http.MethodPatch, "/api/folio/cycles/"+cycle, fmt.Sprintf(`{"map_id":%q}`, theMap))
	if res.Code != http.StatusOK {
		t.Fatalf("set map = %d %s", res.Code, res.Body)
	}
	var linked struct {
		MapID string `json:"map_id"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &linked); err != nil {
		t.Fatal(err)
	}
	if linked.MapID != theMap {
		t.Fatalf("map_id = %q, want %q", linked.MapID, theMap)
	}

	if res := do(http.MethodPatch, "/api/folio/cycles/"+cycle, `{"phase":"do"}`); res.Code != http.StatusBadRequest {
		t.Fatalf("leave plan with an open decision = %d %s, want 400", res.Code, res.Body)
	}
}
