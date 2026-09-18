package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPlans(t *testing.T) {
	var gotPath, gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`{"plans":[{"id":"p1","project_id":"pr1","title":"Ship search","status":"active","tags":["search"],"progress":{"total":5,"done":2,"percent":40},"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}]}`))
	}))
	defer server.Close()

	plans, err := New(server.URL, "tok").ListPlans("folio", PlanFilter{Status: []string{"active", "draft"}})
	if err != nil {
		t.Fatalf("ListPlans() error = %v", err)
	}
	if gotPath != "/api/folio/projects/folio/plans" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "status=active%2Cdraft" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(plans) != 1 || plans[0].Title != "Ship search" {
		t.Fatalf("plans = %+v", plans)
	}
	if plans[0].Progress.Percent != 40 {
		t.Errorf("percent = %d, want 40", plans[0].Progress.Percent)
	}
}

func TestCreatePlan(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"p9","project_id":"pr1","title":"New","status":"draft","tags":[],"progress":{"total":0,"done":0,"percent":0},"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	plan, err := New(server.URL, "tok").CreatePlan("folio", PlanInput{Title: strptr("New")})
	if err != nil {
		t.Fatalf("CreatePlan() error = %v", err)
	}
	if gotBody["title"] != "New" {
		t.Errorf("body = %v", gotBody)
	}
	if _, sent := gotBody["status"]; sent {
		t.Error("status was sent; an unset field must be omitted so the server defaults it")
	}
	if plan.ID != "p9" {
		t.Errorf("id = %q", plan.ID)
	}
}

func TestUpdatePlan(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"p1","project_id":"pr1","title":"Kept","status":"done","tags":[],"progress":{"total":1,"done":1,"percent":100},"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	plan, err := New(server.URL, "tok").UpdatePlan("p1", PlanInput{Status: strptr("done")})
	if err != nil {
		t.Fatalf("UpdatePlan() error = %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/api/folio/plans/p1" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
	if _, sent := gotBody["title"]; sent {
		t.Error("title was sent; PATCH must omit untouched fields")
	}
	if plan.Status != "done" {
		t.Errorf("status = %q", plan.Status)
	}
}

func TestDeletePlan(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := New(server.URL, "tok").DeletePlan("p1"); err != nil {
		t.Fatalf("DeletePlan() error = %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/folio/plans/p1" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
}

func TestListJournal(t *testing.T) {
	var gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"entries":[{"id":"l1","project_id":"pr1","title":"Chose FTS5","body":"## Context","branch":"feat/search","ticket":"J-12","tags":["decision"],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}]}`))
	}))
	defer server.Close()

	logs, err := New(server.URL, "tok").ListJournal("folio", JournalFilter{Branch: "feat/search", Search: "fts5", Limit: 5})
	if err != nil {
		t.Fatalf("ListJournal() error = %v", err)
	}
	if gotQuery != "branch=feat%2Fsearch&kind=journal&limit=5&q=fts5" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(logs) != 1 || logs[0].Branch != "feat/search" {
		t.Fatalf("logs = %+v", logs)
	}
}

func TestWriteJournalEntry(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"l9","project_id":"pr1","title":"Entry","body":"text","tags":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	entry, err := New(server.URL, "tok").WriteJournalEntry("folio", LogInput{Title: strptr("Entry"), Body: strptr("text")})
	if err != nil {
		t.Fatalf("WriteJournalEntry() error = %v", err)
	}
	if gotBody["title"] != "Entry" || gotBody["body"] != "text" {
		t.Errorf("body = %v", gotBody)
	}
	if entry.ID != "l9" {
		t.Errorf("id = %q", entry.ID)
	}
}

func TestAppendJournalEntry(t *testing.T) {
	var gotPath string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"l1","project_id":"pr1","title":"Entry","body":"text\n\nmore","tags":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	entry, err := New(server.URL, "tok").AppendJournalEntry("l1", "more")
	if err != nil {
		t.Fatalf("AppendJournalEntry() error = %v", err)
	}
	if gotPath != "/api/folio/entries/l1/append" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["section"] != "more" {
		t.Errorf("body = %v", gotBody)
	}
	if entry.Body == "" {
		t.Error("expected the updated body back")
	}
}

func TestListDocs(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"entries":[{"id":"d1","project_id":"pr1","slug":"architecture","title":"Architecture","body":"Hexagonal.","tags":["reference"],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}]}`))
	}))
	defer server.Close()

	docs, err := New(server.URL, "tok").ListDocs("folio", DocFilter{})
	if err != nil {
		t.Fatalf("ListDocs() error = %v", err)
	}
	if gotPath != "/api/folio/projects/folio/entries" {
		t.Errorf("path = %q", gotPath)
	}
	if len(docs) != 1 || docs[0].Slug != "architecture" {
		t.Fatalf("docs = %+v", docs)
	}
}

func TestGetDocBySlug(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"d1","project_id":"pr1","slug":"architecture","title":"Architecture","body":"Hexagonal.","tags":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	doc, err := New(server.URL, "tok").GetDocBySlug("folio", "architecture")
	if err != nil {
		t.Fatalf("GetDocBySlug() error = %v", err)
	}
	if gotPath != "/api/folio/projects/folio/entries/architecture" {
		t.Errorf("path = %q", gotPath)
	}
	if doc.Title != "Architecture" {
		t.Errorf("title = %q", doc.Title)
	}
}

func TestCreateDoc(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"d9","project_id":"pr1","slug":"new-doc","title":"New Doc","body":"","tags":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	doc, err := New(server.URL, "tok").CreateDoc("folio", DocInput{Title: strptr("New Doc")})
	if err != nil {
		t.Fatalf("CreateDoc() error = %v", err)
	}
	if gotBody["title"] != "New Doc" {
		t.Errorf("body = %v", gotBody)
	}
	if _, sent := gotBody["slug"]; sent {
		t.Error("slug was sent; the server derives it from the title when omitted")
	}
	if doc.Slug != "new-doc" {
		t.Errorf("slug = %q", doc.Slug)
	}
}

func TestEntityErrorsCarryTheAPIMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"you do not have access to this resource"}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, "tok").ListPlans("folio", PlanFilter{}); err == nil {
		t.Fatal("ListPlans() error = nil, want an error")
	} else if err.Error() != "403: you do not have access to this resource" {
		t.Errorf("err = %q", err.Error())
	}
}
