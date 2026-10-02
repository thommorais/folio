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

func TestBriefHonoursRecentJournal(t *testing.T) {
	dir, err := os.MkdirTemp("", "folio-brief")
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
	ticket := record(t, app, pb.ColIssues, map[string]any{
		"domain": dom.Id, "project": project.Id, "kind": "ticket", "slug": "auth", "title": "Auth",
		"status": "open", "priority": "high", "created_by": owner.Id,
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

	for i := range 3 {
		body := fmt.Sprintf(`{"kind":"journal","issue_id":%q,"title":"entry %d"}`, ticket.Id, i)
		if res := do(http.MethodPost, "/api/folio/projects/redesign/entries", body); res.Code != http.StatusCreated {
			t.Fatalf("write entry = %d %s", res.Code, res.Body)
		}
	}

	for _, path := range []string{
		"/api/folio/issues/" + ticket.Id + "/brief?recent_journal=1",
		"/api/folio/projects/redesign/issues/auth/brief?recent_journal=1",
	} {
		res := do(http.MethodGet, path, "")
		if res.Code != http.StatusOK {
			t.Fatalf("%s = %d %s", path, res.Code, res.Body)
		}
		var brief struct {
			Journal []struct {
				Title string `json:"title"`
			} `json:"journal"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &brief); err != nil {
			t.Fatal(err)
		}
		if len(brief.Journal) != 1 || brief.Journal[0].Title != "entry 2" {
			t.Errorf("%s: journal = %+v, want only the newest, entry 2", path, brief.Journal)
		}
	}

	res := do(http.MethodGet, "/api/folio/projects/redesign/entries?kind=journal", "")
	var list struct {
		Entries []struct {
			Title string `json:"title"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	titles := []string{}
	for _, e := range list.Entries {
		titles = append(titles, e.Title)
	}
	if strings.Join(titles, ",") != "entry 2,entry 1,entry 0" {
		t.Errorf("entries = %v, want newest first", titles)
	}
}
