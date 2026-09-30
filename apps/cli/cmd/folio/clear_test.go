package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func patchServer(t *testing.T) *map[string]any {
	t.Helper()

	sent := map[string]any{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			raw, _ := io.ReadAll(r.Body)
			sent = map[string]any{}
			_ = json.Unmarshal(raw, &sent)
		}
		_, _ = w.Write([]byte(`{"id":"x","slug":"x","title":"x","tags":[],"depends_on":[]}`))
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "")

	return &sent
}

func runTree(t *testing.T, cmd *cobra.Command, args ...string) error {
	t.Helper()

	return silenced(t, func() error {
		cmd.SetArgs(args)
		return cmd.Execute()
	})
}

func TestUpdateCommandsClearAFieldPassedEmpty(t *testing.T) {
	cases := []struct {
		name  string
		tree  func() *cobra.Command
		args  []string
		empty []string
		lists []string
	}{
		{
			"ticket", ticketCommand,
			[]string{"update", "tk1", "--body", "", "--assignee", "", "--parent", "", "--external-ref", "", "--wayfinder", "", "--depends-on", "", "--tags", ""},
			[]string{"body", "assignee", "parent_id", "external_ref", "wayfinder"},
			[]string{"depends_on", "tags"},
		},
		{
			"todo", todoCommand,
			[]string{"update", "tk1", "--details", "", "--plan", "", "--ticket", "", "--due", "", "--tags", ""},
			[]string{"body", "plan_id", "parent_id", "due_date"},
			[]string{"tags"},
		},
		{
			"plan", planCommand,
			[]string{"update", "pl1", "--goal", "", "--ticket", "", "--tags", ""},
			[]string{"goal", "issue_id"},
			[]string{"tags"},
		},
		{
			"doc", docCommand,
			[]string{"update", "d1", "--body", "", "--ticket", "", "--tags", ""},
			[]string{"body", "issue_id"},
			[]string{"tags"},
		},
		{
			"journal", journalCommand,
			[]string{"update", "j1", "--body", "", "--branch", "", "--pr", "", "--ticket", "", "--external-ref", "", "--tags", ""},
			[]string{"body", "branch", "pr", "issue_id", "external_ref"},
			[]string{"tags"},
		},
		{
			"knowledge", knowledgeCommand,
			[]string{"update", "n1", "--body", "", "--project", "", "--tags", ""},
			[]string{"body", "project_id"},
			[]string{"tags"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sent := patchServer(t)

			if err := runTree(t, tc.tree(), tc.args...); err != nil {
				t.Fatal(err)
			}

			for _, key := range tc.empty {
				if value, ok := (*sent)[key]; !ok || value != "" {
					t.Errorf("%s = %v (present %v), want an empty string sent", key, value, ok)
				}
			}
			for _, key := range tc.lists {
				list, ok := (*sent)[key].([]any)
				if !ok || len(list) != 0 {
					t.Errorf("%s = %v, want an empty list sent", key, (*sent)[key])
				}
			}
		})
	}
}

func TestUpdateSendsOnlyTheFlagsPassed(t *testing.T) {
	sent := patchServer(t)

	if err := runTree(t, ticketCommand(), "update", "tk1", "--status", "done"); err != nil {
		t.Fatal(err)
	}
	if len(*sent) != 1 || (*sent)["status"] != "done" {
		t.Errorf("sent = %v, want only status", *sent)
	}
}

func TestUpdateWithNoFlagsStillRefuses(t *testing.T) {
	patchServer(t)

	err := runTree(t, ticketCommand(), "update", "tk1")
	if err == nil || !strings.Contains(err.Error(), "nothing to update") {
		t.Errorf("error = %v, want nothing to update", err)
	}
}
