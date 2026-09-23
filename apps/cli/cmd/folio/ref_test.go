package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func slugMissServer(t *testing.T, found string) *[]string {
	t.Helper()

	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/api/folio/projects/") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not found."}`))
			return
		}
		_, _ = w.Write([]byte(found))
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "")
	t.Cleanup(func() { flagProject = "" })

	return &paths
}

func runCommand(t *testing.T, cmd *cobra.Command, args ...string) error {
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

	cmd.SetOut(devnull)
	cmd.SetErr(devnull)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestIDResolvesWithAProjectSelected(t *testing.T) {
	const ticket = `{"id":"abc123","kind":"ticket","title":"x","status":"open","tags":[],"depends_on":[]}`
	const brief = `{"issue":` + ticket + `,"children":[],"plans":[],"journal":[],"cycles":[],"docs":[]}`
	const doc = `{"id":"abc123","slug":"x","title":"x","tags":[]}`

	cases := []struct {
		name     string
		cmd      func() *cobra.Command
		args     []string
		found    string
		fallback string
	}{
		{"ticket get", ticketCommand, []string{"get", "abc123"}, ticket, "/api/folio/issues/abc123"},
		{"ticket brief", ticketCommand, []string{"brief", "abc123"}, brief, "/api/folio/issues/abc123/brief"},
		{"doc get", docCommand, []string{"get", "abc123"}, doc, "/api/folio/entries/abc123"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			paths := slugMissServer(t, c.found)

			if err := runCommand(t, c.cmd(), append(c.args, "-p", "folio")...); err != nil {
				t.Fatalf("%s: %v", c.name, err)
			}
			if got := *paths; len(got) != 2 || got[1] != c.fallback {
				t.Errorf("requests = %v, want the slug route then %s", got, c.fallback)
			}
		})
	}
}

func TestMissingRefStillFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not found."}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Cleanup(func() { flagProject = "" })

	if err := runCommand(t, ticketCommand(), "get", "nope", "-p", "folio"); err == nil {
		t.Error("a ref that is neither a slug nor an id should fail")
	}
}
