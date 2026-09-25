package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func cycleServer(t *testing.T, project string) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{method: r.Method, path: r.URL.Path}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		got = append(got, req)

		switch {
		case strings.Contains(r.URL.Path, "/cycles"):
			if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/cycles") {
				w.WriteHeader(http.StatusCreated)
			}
			_, _ = w.Write([]byte(`{"id":"cy1","project_id":"pr1","issue_id":"tk1","ordinal":1,"phase":"plan"}`))
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"m1","project_id":"pr1","kind":"ticket","title":"Plan it","wayfinder":"map","tags":[],"depends_on":[]}`))
		default:
			_, _ = w.Write([]byte(`{"id":"tk1","project_id":"pr1","slug":"wayfinder-view","kind":"ticket","title":"Wayfinder view","status":"open","tags":[],"depends_on":[]}`))
		}
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", project)

	return &got
}

func runCycle(t *testing.T, args ...string) error {
	t.Helper()
	cmd := cycleCommand()
	cmd.SetArgs(args)
	return silenced(t, cmd.Execute)
}

func TestCycleCommandsAddressTheTicket(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		method string
		path   string
		body   map[string]any
	}{
		{[]string{"next", "tk1"}, http.MethodPost, "/api/folio/issues/tk1/cycles/current/next", nil},
		{[]string{"phase", "tk1", "do"}, http.MethodPatch, "/api/folio/issues/tk1/cycles/current", map[string]any{"phase": "do"}},
		{[]string{"resolve", "tk1", "Shipped"}, http.MethodPatch, "/api/folio/issues/tk1/cycles/current", map[string]any{"resolution": "Shipped"}},
		{[]string{"open", "tk1"}, http.MethodPost, "/api/folio/issues/tk1/cycles", nil},
		{[]string{"list", "tk1"}, http.MethodGet, "/api/folio/issues/tk1/cycles", nil},
	} {
		t.Run(tc.args[0], func(t *testing.T) {
			got := cycleServer(t, "")
			if err := runCycle(t, tc.args...); err != nil {
				t.Fatal(err)
			}
			if len(*got) != 1 {
				t.Fatalf("requests = %+v, want one", *got)
			}
			req := (*got)[0]
			if req.method != tc.method || req.path != tc.path {
				t.Errorf("got %s %s, want %s %s", req.method, req.path, tc.method, tc.path)
			}
			for key, want := range tc.body {
				if req.body[key] != want {
					t.Errorf("body[%s] = %v, want %v", key, req.body[key], want)
				}
			}
		})
	}
}

func TestCycleCommandResolvesASlug(t *testing.T) {
	got := cycleServer(t, "folio")
	if err := runCycle(t, "next", "wayfinder-view"); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 2 {
		t.Fatalf("requests = %+v", *got)
	}
	if lookup := (*got)[0]; lookup.path != "/api/folio/projects/folio/issues/wayfinder-view" {
		t.Errorf("lookup = %+v", lookup)
	}
	if next := (*got)[1]; next.path != "/api/folio/issues/tk1/cycles/current/next" {
		t.Errorf("next = %+v", next)
	}
}

func TestCycleOpenWithAMapCreatesAndLinksIt(t *testing.T) {
	got := cycleServer(t, "")
	if err := runCycle(t, "open", "tk1", "--map", "Plan the view"); err != nil {
		t.Fatal(err)
	}

	want := []struct{ method, path string }{
		{http.MethodGet, "/api/folio/issues/tk1"},
		{http.MethodPost, "/api/folio/issues/tk1/cycles"},
		{http.MethodPost, "/api/folio/projects/pr1/issues"},
		{http.MethodPatch, "/api/folio/issues/tk1/cycles/current"},
	}
	if len(*got) != len(want) {
		t.Fatalf("requests = %+v", *got)
	}
	for i, w := range want {
		if (*got)[i].method != w.method || (*got)[i].path != w.path {
			t.Errorf("request %d = %s %s, want %s %s", i, (*got)[i].method, (*got)[i].path, w.method, w.path)
		}
	}
	create, link := (*got)[2], (*got)[3]
	if create.body["wayfinder"] != "map" || create.body["parent_id"] != "tk1" || create.body["title"] != "Plan the view" {
		t.Errorf("create = %+v", create.body)
	}
	if link.body["map_id"] != "m1" {
		t.Errorf("link = %+v", link.body)
	}
}
