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

func TestInterviewRunsOverTheAPI(t *testing.T) {
	dir, err := os.MkdirTemp("", "folio-interview")
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
	decode := func(res *httptest.ResponseRecorder, into any) {
		t.Helper()
		if err := json.Unmarshal(res.Body.Bytes(), into); err != nil {
			t.Fatalf("decode %s: %v", res.Body, err)
		}
	}

	var ticket struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	res := do(http.MethodPost, "/api/folio/projects/redesign/issues", `{"kind":"ticket","title":"Tree or graph","wayfinder":"grilling"}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("create ticket = %d %s", res.Code, res.Body)
	}
	decode(res, &ticket)
	base := "/api/folio/issues/" + ticket.ID + "/interview"

	if res := do(http.MethodGet, base, ""); res.Code != http.StatusNotFound {
		t.Fatalf("show before start = %d %s, want 404", res.Code, res.Body)
	}

	type interview struct {
		ID          string `json:"id"`
		IssueID     string `json:"issue_id"`
		Topic       string `json:"topic"`
		AgentStatus string `json:"agent_status"`
		Handled     int    `json:"handled"`
		FinishedAt  string `json:"finished_at"`
		URL         string `json:"url"`
		State       struct {
			Terms     []any `json:"terms"`
			Questions []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"questions"`
		} `json:"state"`
	}
	var started interview
	res = do(http.MethodPost, base, "")
	if res.Code != http.StatusCreated {
		t.Fatalf("start = %d %s", res.Code, res.Body)
	}
	decode(res, &started)
	if started.IssueID != ticket.ID || started.Topic != "Tree or graph" || started.AgentStatus != "waiting" || started.State.Terms == nil {
		t.Fatalf("started = %+v", started)
	}
	if want := "http://example.com/acme/web/redesign/interview/" + ticket.Slug; started.URL != want {
		t.Fatalf("url = %q, want %q", started.URL, want)
	}
	var resumed interview
	res = do(http.MethodPost, base, "")
	if res.Code != http.StatusOK {
		t.Fatalf("a second start resumes = %d %s, want 200", res.Code, res.Body)
	}
	decode(res, &resumed)
	if resumed.ID != started.ID {
		t.Fatalf("resumed %s, want %s", resumed.ID, started.ID)
	}

	round := `{"questions":[
		{"id":"q1","round":1,"title":"Tree or graph","options":[{"k":"a","text":"Tree"},{"k":"b","text":"Graph"}],"rec":{"option":"b","why":"Edges show blockers."}},
		{"id":"q2","round":1,"title":"What do we call it","rec":{"text":"interview","why":"Plain word."}}
	]}`
	var summary struct {
		Round    int `json:"round"`
		Added    int `json:"added"`
		Answered int `json:"answered"`
		Handled  int `json:"handled"`
	}
	res = do(http.MethodPatch, base, round)
	if res.Code != http.StatusOK {
		t.Fatalf("patch = %d %s", res.Code, res.Body)
	}
	decode(res, &summary)
	if summary.Round != 1 || summary.Added != 2 || summary.Answered != 0 || summary.Handled != 0 {
		t.Fatalf("summary = %+v", summary)
	}
	if res := do(http.MethodPatch, base, `{"questions":[{"id":"q1","bogus":1}]}`); res.Code != http.StatusBadRequest {
		t.Fatalf("a patch with an unknown field = %d %s, want 400", res.Code, res.Body)
	}
	if res := do(http.MethodPatch, base, ""); res.Code != http.StatusBadRequest {
		t.Fatalf("an empty patch = %d %s, want 400", res.Code, res.Body)
	}

	type pendingView struct {
		Handled int `json:"handled"`
		Sends   []struct {
			Seq     int              `json:"seq"`
			At      string           `json:"at"`
			Actions []map[string]any `json:"actions"`
		} `json:"sends"`
	}
	var pending pendingView
	decode(do(http.MethodGet, base+"/pending", ""), &pending)
	if pending.Sends == nil || len(pending.Sends) != 0 {
		t.Fatalf("nothing sent yet, got %+v", pending)
	}

	var send struct {
		Seq     int              `json:"seq"`
		Actions []map[string]any `json:"actions"`
	}
	res = do(http.MethodPost, base+"/sends", `{"actions":[{"type":"answer","q":"q1","kind":"accept","option":"b"},{"type":"defer","q":"q2"}]}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("send = %d %s", res.Code, res.Body)
	}
	decode(res, &send)
	if send.Seq != 1 || len(send.Actions) != 2 {
		t.Fatalf("send = %+v", send)
	}
	if res := do(http.MethodPost, base+"/sends", `{"actions":[{"type":"defer","q":"q9"}]}`); res.Code != http.StatusBadRequest {
		t.Fatalf("a Send naming no question = %d %s, want 400", res.Code, res.Body)
	}

	decode(do(http.MethodGet, base+"/pending", ""), &pending)
	if len(pending.Sends) != 1 || pending.Sends[0].Seq != 1 || pending.Sends[0].At == "" || pending.Sends[0].Actions[0]["q"] != "q1" {
		t.Fatalf("pending = %+v", pending)
	}
	var working interview
	decode(do(http.MethodGet, base, ""), &working)
	if working.AgentStatus != "working" {
		t.Fatalf("reading pending marks the agent working, got %q", working.AgentStatus)
	}

	settle := `{"questions":[{"id":"q1","status":"answered","answer":{"kind":"accept","option":"b"}},{"id":"q2","status":"deferred"}],"agent":{"handled":1}}`
	res = do(http.MethodPatch, base, settle)
	if res.Code != http.StatusOK {
		t.Fatalf("settle = %d %s", res.Code, res.Body)
	}
	decode(res, &summary)
	if summary.Answered != 1 || summary.Handled != 1 {
		t.Fatalf("settle summary = %+v", summary)
	}

	if res := do(http.MethodPost, base+"/finish", `{"answer":"A graph."}`); res.Code != http.StatusBadRequest {
		t.Fatalf("finish without a doc = %d %s, want 400", res.Code, res.Body)
	}
	var finished interview
	res = do(http.MethodPost, base+"/finish", `{"answer":"A graph.","doc":"# Tree or graph\n\nA graph."}`)
	if res.Code != http.StatusOK {
		t.Fatalf("finish = %d %s", res.Code, res.Body)
	}
	decode(res, &finished)
	if finished.FinishedAt == "" {
		t.Fatalf("finished = %+v", finished)
	}

	var closed struct {
		Status     string `json:"status"`
		Resolution string `json:"resolution"`
	}
	decode(do(http.MethodGet, "/api/folio/issues/"+ticket.ID, ""), &closed)
	if closed.Status != "done" || closed.Resolution != "A graph." {
		t.Fatalf("ticket after finish = %+v", closed)
	}
	if res := do(http.MethodGet, base, ""); res.Code != http.StatusNotFound {
		t.Fatalf("show after finish = %d %s, want 404", res.Code, res.Body)
	}

	var list struct {
		Interviews []interview `json:"interviews"`
	}
	decode(do(http.MethodGet, "/api/folio/issues/"+ticket.ID+"/interviews", ""), &list)
	if len(list.Interviews) != 1 || list.Interviews[0].ID != started.ID || list.Interviews[0].FinishedAt == "" {
		t.Fatalf("list = %+v", list)
	}
}
