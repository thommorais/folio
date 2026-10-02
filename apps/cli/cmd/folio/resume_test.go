package main

import (
	"strings"
	"testing"

	"folio/cli/internal/client"
)

func TestDrift(t *testing.T) {
	handoff := client.JournalEntry{Branch: "feat/landing", Meta: map[string]any{"commit": "7601dc0aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}

	cases := []struct {
		name string
		here checkout
		want string
	}{
		{"same checkout", checkout{"feat/landing", "7601dc0aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, ""},
		{"other branch", checkout{"main", "7601dc0aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, "drift: here main @ 7601dc0, handoff feat/landing @ 7601dc0"},
		{"moved on", checkout{"feat/landing", "f113159bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}, "drift: here feat/landing @ f113159, handoff feat/landing @ 7601dc0"},
		{"outside a repo", checkout{}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := drift(handoff, c.here); got != c.want {
				t.Errorf("drift = %q, want %q", got, c.want)
			}
		})
	}
}

func TestResumeRendersWhatTheNextSessionNeeds(t *testing.T) {
	out := captureStdout(t, func() error {
		return renderResume(client.TicketResume{
			Ticket: client.Ticket{ID: "tk1", Title: "Landing page", Status: "in_progress", Priority: "medium", Body: "The spec."},
			Handoff: &client.JournalEntry{
				CreatedAt: "2026-10-01T20:00:00Z", Branch: "feat/landing", Body: "Next: coverage widget.\nTrap: empty seed.",
				Meta: map[string]any{"commit": "7601dc0aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "uncommitted": []any{"wip.go"}},
			},
			Logs:   []client.WorkLog{{ID: "e9", CreatedAt: "2026-10-01T21:00:00Z", Body: "Coverage renders."}},
			Open:   []client.Ticket{{ID: "c1", Kind: "ticket", Status: "open", Priority: "medium", Title: "Coverage", Blocked: true}},
			Closed: 8,
			Docs:   []client.Doc{{ID: "d1", Slug: "widget-store", Title: "Widget store"}},
		}, checkout{"feat/landing", "7601dc0aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
	})

	for _, want := range []string{
		"handoff 2026-10-01T20:00:00Z  feat/landing @ 7601dc0\n  Next: coverage widget.\n  Trap: empty seed.\n",
		"uncommitted at handoff: wip.go\n",
		"logs since handoff\n  e9  2026-10-01T21:00:00Z  Coverage renders.\n",
		"open\n  c1  ticket open         medium Coverage (blocked)\n",
		"closed: 8\n",
		"docs\n  d1  widget-store  Widget store\n",
		"\nThe spec.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "drift") {
		t.Errorf("no drift on the same checkout:\n%s", out)
	}
	if !strings.HasSuffix(out, "stop: folio stop tk1 -\n") {
		t.Errorf("want the stop command last:\n%s", out)
	}
	if strings.Index(out, "The spec.") < strings.Index(out, "docs") {
		t.Errorf("the body goes after the state:\n%s", out)
	}
}

func TestResumeWithoutAHandoffSaysSo(t *testing.T) {
	out := captureStdout(t, func() error {
		return renderResume(client.TicketResume{
			Ticket: client.Ticket{ID: "tk1", Title: "Landing page"},
			Logs:   []client.WorkLog{{ID: "e1", CreatedAt: "2026-10-01T21:00:00Z", Body: "Started."}},
		}, checkout{})
	})

	if !strings.Contains(out, "no handoff yet\n") || !strings.Contains(out, "recent logs\n  e1") {
		t.Errorf("out:\n%s", out)
	}
}

func TestResumeCallsTheResumeRoute(t *testing.T) {
	const ticket = `{"id":"abc123","kind":"ticket","title":"x","status":"open","tags":[],"depends_on":[]}`
	paths := slugMissServer(t, `{"issue":`+ticket+`,"handoff":null,"logs":[],"open":[],"closed":0,"plans":[],"docs":[]}`)

	if err := runCommand(t, resumeCommand(), "abc123", "-p", "folio"); err != nil {
		t.Fatal(err)
	}
	if got := *paths; len(got) != 2 || got[0] != "/api/folio/projects/folio/issues/abc123/resume" || got[1] != "/api/folio/issues/abc123/resume" {
		t.Errorf("requests = %v, want the slug route then the id route", got)
	}
}
