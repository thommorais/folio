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

func TestResumeOverTheAPI(t *testing.T) {
	dir, err := os.MkdirTemp("", "folio-resume")
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

	handoff := fmt.Sprintf(`{"kind":"handoff","issue_id":%q,"body":"Next: login form.","branch":"feat/auth","meta":{"commit":"abc123"}}`, ticket.Id)
	if res := do(http.MethodPost, "/api/folio/projects/redesign/entries", handoff); res.Code != http.StatusCreated {
		t.Fatalf("write handoff = %d %s", res.Code, res.Body)
	}
	doc := fmt.Sprintf(`{"kind":"doc","issue_id":%q,"title":"Auth notes","body":"a long body"}`, ticket.Id)
	if res := do(http.MethodPost, "/api/folio/projects/redesign/entries", doc); res.Code != http.StatusCreated {
		t.Fatalf("write doc = %d %s", res.Code, res.Body)
	}

	for _, path := range []string{
		"/api/folio/issues/" + ticket.Id + "/resume",
		"/api/folio/projects/redesign/issues/auth/resume",
	} {
		res := do(http.MethodGet, path, "")
		if res.Code != http.StatusOK {
			t.Fatalf("%s = %d %s", path, res.Code, res.Body)
		}
		var resume struct {
			Issue   struct{ ID string } `json:"issue"`
			Handoff *struct {
				Body   string         `json:"body"`
				Branch string         `json:"branch"`
				Meta   map[string]any `json:"meta"`
			} `json:"handoff"`
			Logs   []json.RawMessage `json:"logs"`
			Open   []json.RawMessage `json:"open"`
			Closed int               `json:"closed"`
			Docs   []map[string]any  `json:"docs"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &resume); err != nil {
			t.Fatal(err)
		}
		if resume.Issue.ID != ticket.Id {
			t.Errorf("%s: issue = %q", path, resume.Issue.ID)
		}
		if resume.Handoff == nil || resume.Handoff.Body != "Next: login form." || resume.Handoff.Branch != "feat/auth" || resume.Handoff.Meta["commit"] != "abc123" {
			t.Errorf("%s: handoff = %+v", path, resume.Handoff)
		}
		if resume.Logs == nil || resume.Open == nil {
			t.Errorf("%s: logs and open must be arrays, got %s", path, res.Body)
		}
		if len(resume.Docs) != 1 || resume.Docs[0]["title"] != "Auth notes" {
			t.Errorf("%s: docs = %v", path, resume.Docs)
		}
		if _, ok := resume.Docs[0]["body"]; ok {
			t.Errorf("%s: a doc in a resume is a pointer, not its body", path)
		}
	}
}
