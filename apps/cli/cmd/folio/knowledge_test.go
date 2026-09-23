package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const aNote = `{"id":"k1","slug":"realtime-on-railway","title":"Realtime on Railway","body":"disable buffering","tags":["pocketbase"],"created_by":"u1","created_at":"2026-09-23T10:00:00Z","updated_at":"2026-09-23T10:00:00Z"}`

// knowledgeServer records every request and answers each one with the same
// note, since these tests assert on what was sent rather than what came back.
func knowledgeServer(t *testing.T, body string) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{method: r.Method, path: r.URL.Path + queryOf(r)}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		got = append(got, req)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")

	return &got
}

func queryOf(r *http.Request) string {
	if r.URL.RawQuery == "" {
		return ""
	}
	return "?" + r.URL.RawQuery
}

func runKnowledge(t *testing.T, args ...string) error {
	t.Helper()

	stdout := os.Stdout
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	os.Stdout = devnull
	t.Cleanup(func() {
		os.Stdout = stdout
		_ = devnull.Close()
	})

	cmd := knowledgeCommand()
	cmd.SetOut(devnull)
	cmd.SetErr(devnull)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// Knowledge has no project segment anywhere: that is the difference from every
// other resource, and a route that grew one would still look plausible.
func TestKnowledgeRoutesCarryNoProject(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		method string
		path   string
	}{
		{"list", []string{"list"}, http.MethodGet, "/api/folio/knowledge"},
		{"add", []string{"add", "A note"}, http.MethodPost, "/api/folio/knowledge"},
		{"get", []string{"get", "realtime-on-railway"}, http.MethodGet, "/api/folio/knowledge/realtime-on-railway"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := knowledgeServer(t, listOrNote(tc.name))
			if err := runKnowledge(t, tc.args...); err != nil {
				t.Fatal(err)
			}
			if len(*got) == 0 {
				t.Fatal("no request was sent")
			}
			first := (*got)[0]
			if first.method != tc.method || first.path != tc.path {
				t.Errorf("sent %s %s, want %s %s", first.method, first.path, tc.method, tc.path)
			}
			if strings.Contains(first.path, "/projects/") {
				t.Errorf("route is scoped to a project: %s", first.path)
			}
		})
	}
}

func listOrNote(name string) string {
	if name == "list" {
		return `{"knowledge":[` + aNote + `]}`
	}
	return aNote
}

// Tags here are free-form: the shared helper checks them against the domain
// vocabulary, which knowledge has none of, so a new tag must go through.
func TestKnowledgeAcceptsUnknownTags(t *testing.T) {
	got := knowledgeServer(t, aNote)

	if err := runKnowledge(t, "add", "A note", "--tags", "pocketbase, railway"); err != nil {
		t.Fatalf("a tag outside the project vocabulary was rejected: %v", err)
	}

	body := (*got)[0].body
	tags, ok := body["tags"].([]any)
	if !ok {
		t.Fatalf("tags were not sent: %v", body)
	}
	if len(tags) != 2 || tags[0] != "pocketbase" || tags[1] != "railway" {
		t.Errorf("sent tags %v, want [pocketbase railway] trimmed", tags)
	}
}

// A slug addresses a note, but update and delete take an id, so the command
// resolves one into the other rather than sending a slug to an id route.
func TestKnowledgeUpdateResolvesASlugToItsID(t *testing.T) {
	got := knowledgeServer(t, aNote)

	if err := runKnowledge(t, "update", "realtime-on-railway", "--title", "Renamed"); err != nil {
		t.Fatal(err)
	}

	if len(*got) != 2 {
		t.Fatalf("sent %d requests, want a lookup then a patch: %v", len(*got), *got)
	}
	lookup, patch := (*got)[0], (*got)[1]
	if lookup.method != http.MethodGet || lookup.path != "/api/folio/knowledge/realtime-on-railway" {
		t.Errorf("lookup was %s %s", lookup.method, lookup.path)
	}
	if patch.method != http.MethodPatch || patch.path != "/api/folio/knowledge/k1" {
		t.Errorf("patched %s %s, want PATCH by id", patch.method, patch.path)
	}
	if patch.body["title"] != "Renamed" {
		t.Errorf("patch body %v, want the new title", patch.body)
	}
}

// An update with no flags would otherwise send an empty patch and report
// success without changing anything.
func TestKnowledgeUpdateNeedsAField(t *testing.T) {
	got := knowledgeServer(t, aNote)

	err := runKnowledge(t, "update", "realtime-on-railway")
	if err == nil {
		t.Fatal("an update with no fields was accepted")
	}
	if len(*got) != 0 {
		t.Errorf("it still called the API: %v", *got)
	}
}

// --unattached and --project are different questions: an empty project means
// "no restriction", so it cannot ask for notes that have none.
func TestKnowledgeListFilters(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"unattached", []string{"list", "--unattached"}, "unattached=true"},
		{"by project", []string{"list", "--project", "p1"}, "project=p1"},
		{"by tag", []string{"list", "--tags", "pocketbase"}, "tags=pocketbase"},
		{"by text", []string{"list", "-q", "railway"}, "q=railway"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := knowledgeServer(t, `{"knowledge":[]}`)
			if err := runKnowledge(t, tc.args...); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains((*got)[0].path, tc.want) {
				t.Errorf("sent %s, want it to carry %s", (*got)[0].path, tc.want)
			}
		})
	}
}
