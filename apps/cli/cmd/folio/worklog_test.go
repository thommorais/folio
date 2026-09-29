package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func workLogServer(t *testing.T) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, request{method: r.Method, path: r.URL.Path})
		_, _ = io.Copy(io.Discard, r.Body)

		switch r.URL.Path {
		case "/api/folio/issues/tk1":
			_, _ = w.Write([]byte(`{"id":"tk1","project_id":"pr1","kind":"ticket","tags":[],"depends_on":[]}`))
		case "/api/folio/plans/pl1":
			_, _ = w.Write([]byte(`{"id":"pl1","project_id":"pr1","tags":[]}`))
		case "/api/folio/projects/pr1/entries":
			if r.Method == http.MethodPost {
				_, _ = w.Write([]byte(`{"id":"e1"}`))
				return
			}
			_, _ = w.Write([]byte(`{"entries":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not found."}`))
		}
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "")

	return &got
}

func runWorkLog(t *testing.T, args ...string) error {
	t.Helper()

	return silenced(t, func() error {
		cmd := workLogCommand()
		cmd.SetArgs(args)
		return cmd.Execute()
	})
}

func TestWorkLogDerivesTheProjectFromTheTarget(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		lookup string
		method string
	}{
		{"write to a ticket", []string{"write", "a note", "--ticket", "tk1"}, "/api/folio/issues/tk1", http.MethodPost},
		{"write to a todo", []string{"write", "a note", "--todo", "tk1"}, "/api/folio/issues/tk1", http.MethodPost},
		{"write to a plan", []string{"write", "a note", "--plan", "pl1"}, "/api/folio/plans/pl1", http.MethodPost},
		{"list a ticket", []string{"list", "--ticket", "tk1"}, "/api/folio/issues/tk1", http.MethodGet},
		{"list a plan", []string{"list", "--plan", "pl1"}, "/api/folio/plans/pl1", http.MethodGet},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := workLogServer(t)

			if err := runWorkLog(t, tc.args...); err != nil {
				t.Fatalf("error = %v, want the project read from the target", err)
			}

			if len(*got) != 2 || (*got)[0].path != tc.lookup {
				t.Fatalf("requests = %+v, want a lookup of %s then the call", *got, tc.lookup)
			}
			if call := (*got)[1]; call.method != tc.method || call.path != "/api/folio/projects/pr1/entries" {
				t.Errorf("call = %+v", call)
			}
		})
	}
}

func TestWorkLogWithAProjectSelectedSkipsTheLookup(t *testing.T) {
	got := workLogServer(t)

	if err := runWorkLog(t, "write", "a note", "--ticket", "tk1", "-p", "pr1"); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 || (*got)[0].method != http.MethodPost {
		t.Errorf("requests = %+v, want one POST", *got)
	}
}
