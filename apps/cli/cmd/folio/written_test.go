package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
)

func writeServer(t *testing.T, record string) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte(record))
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "p1")
	t.Cleanup(func() { flagProject = "" })
}

func TestWritesPrintOnlyWhatTheServerDecided(t *testing.T) {
	const open = `{"id":"x1","project_id":"p1","kind":"ticket","slug":"s1","title":"The title","body":"The body","status":"open","priority":"high","tags":["cli"],"depends_on":[],"blocked":false}`
	const blocked = `{"id":"x1","project_id":"p1","kind":"todo","slug":"s1","title":"The title","status":"open","priority":"high","tags":["cli"],"depends_on":["x2"],"blocked":true}`
	const resolved = `{"id":"x1","project_id":"p1","kind":"ticket","slug":"s1","title":"The title","body":"The body","status":"done","priority":"high","tags":[],"depends_on":[],"resolution":"A graph."}`

	cases := []struct {
		name   string
		cmd    func() *cobra.Command
		args   []string
		record string
		want   string
	}{
		{"ticket create", ticketCommand, []string{"create", "The title", "--tags", "cli"}, open, "x1\nslug: s1\n"},
		{"ticket update", ticketCommand, []string{"update", "x1", "--priority", "high"}, open, "x1\n"},
		{"ticket update blocked", ticketCommand, []string{"update", "x1", "--depends-on", "x2"}, blocked, "x1\nblocked\n"},
		{"ticket resolve", ticketCommand, []string{"resolve", "x1", "A graph."}, resolved, "x1\nstatus: done\n"},
		{"todo create", todoCommand, []string{"create", "The title", "--tags", "cli"}, open, "x1\n"},
		{"todo done", todoCommand, []string{"done", "x1"}, open, "x1\n"},
		{"todo block", todoCommand, []string{"block", "x1", "--on", "x2"}, blocked, "x1\nblocked\n"},
		{"plan create", planCommand, []string{"create", "The title"}, `{"id":"x1","title":"The title","status":"draft","tags":[]}`, "x1\n"},
		{"journal write", journalCommand, []string{"write", "The title", "--branch", "", "--pr", ""}, `{"id":"x1","slug":"s1","title":"The title","tags":[]}`, "x1\nslug: s1\n"},
		{"journal update", journalCommand, []string{"update", "x1", "--title", "New"}, `{"id":"x1","slug":"s1","title":"New","tags":[]}`, "x1\n"},
		{"doc create", docCommand, []string{"create", "The title"}, `{"id":"x1","slug":"s1","title":"The title","tags":[]}`, "x1\nslug: s1\n"},
		{"doc update", docCommand, []string{"update", "x1", "--title", "New"}, `{"id":"x1","slug":"s1","title":"New","tags":[]}`, "x1\n"},
		{"kb add", knowledgeCommand, []string{"add", "The title", "--body", "b"}, `{"id":"x1","slug":"s1","title":"The title","tags":[]}`, "x1\nslug: s1\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			writeServer(t, c.record)

			got := captureStdout(t, func() error {
				cmd := c.cmd()
				cmd.SetArgs(c.args)
				return cmd.Execute()
			})
			if got != c.want {
				t.Errorf("printed %q, want %q", got, c.want)
			}
		})
	}
}
