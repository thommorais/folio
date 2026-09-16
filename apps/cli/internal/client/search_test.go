package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearch(t *testing.T) {
	var gotPath, gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`{"hits":[{"kind":"log","id":"l1","project_id":"pr1","title":"Chose SQLite FTS5","snippet":"Search has to span logs…","tags":["decision"],"created_at":"2026-09-11T01:07:01Z"}]}`))
	}))
	defer server.Close()

	hits, err := New(server.URL, "tok").Search("folio", SearchQuery{Text: "fts5"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if gotPath != "/api/folio/projects/folio/search" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "q=fts5" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(hits) != 1 || hits[0].Kind != "log" {
		t.Fatalf("hits = %+v", hits)
	}
	if hits[0].Snippet == "" {
		t.Error("expected a snippet")
	}
}

func TestSearchSendsKindsAndTags(t *testing.T) {
	var gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"hits":[]}`))
	}))
	defer server.Close()

	_, err := New(server.URL, "tok").Search("folio", SearchQuery{
		Text:  "rules",
		Kinds: []string{"log", "doc"},
		Tags:  []string{"decision"},
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if gotQuery != "kind=log%2Cdoc&limit=5&q=rules&tags=decision" {
		t.Errorf("query = %q", gotQuery)
	}
}

// The API rejects a query with neither q nor tags, since it would scan the
// whole project. Failing locally keeps that from looking like a server fault.
func TestSearchRequiresATermOrTags(t *testing.T) {
	reached := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		_, _ = w.Write([]byte(`{"hits":[]}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, "tok").Search("folio", SearchQuery{}); err == nil {
		t.Fatal("Search() error = nil, want an error for an empty query")
	}
	if reached {
		t.Error("sent a request the API would reject")
	}
}
