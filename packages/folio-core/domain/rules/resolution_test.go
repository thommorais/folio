package rules_test

import (
	"errors"
	"strings"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func TestCheckResolvedIssue(t *testing.T) {
	decision := func(w domain.WayfinderType, resolution string) domain.Issue {
		return domain.Issue{Kind: domain.IssueTicket, Wayfinder: w, Resolution: resolution}
	}

	for _, w := range []domain.WayfinderType{domain.WayfinderResearch, domain.WayfinderPrototype, domain.WayfinderGrilling, domain.WayfinderTask} {
		t.Run("refuses to close a "+string(w)+" ticket without a resolution", func(t *testing.T) {
			for _, to := range []domain.IssueStatus{domain.IssueDone, domain.IssueCancelled} {
				if !errors.Is(rules.CheckResolvedIssue(decision(w, ""), to), domain.ErrValidation) {
					t.Fatalf("want validation error closing to %s", to)
				}
			}
		})
	}

	t.Run("refuses a blank resolution", func(t *testing.T) {
		if !errors.Is(rules.CheckResolvedIssue(decision(domain.WayfinderGrilling, "  "), domain.IssueDone), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("allows closing with a resolution", func(t *testing.T) {
		if err := rules.CheckResolvedIssue(decision(domain.WayfinderGrilling, "A graph"), domain.IssueDone); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("lets a map close freely", func(t *testing.T) {
		if err := rules.CheckResolvedIssue(decision(domain.WayfinderMap, ""), domain.IssueDone); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("lets a ticket outside wayfinder close freely", func(t *testing.T) {
		if err := rules.CheckResolvedIssue(decision("", ""), domain.IssueDone); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("leaves a non-terminal status alone", func(t *testing.T) {
		if err := rules.CheckResolvedIssue(decision(domain.WayfinderGrilling, ""), domain.IssueInProgress); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}

func TestValidateIssueBoundsTheResolution(t *testing.T) {
	issue := domain.Issue{ProjectID: "p1", Kind: domain.IssueTicket, Slug: "a", Title: "A", Status: domain.IssueDone}

	issue.Resolution = strings.Repeat("x", rules.TitleMaxLen)
	if err := rules.ValidateIssue(issue); err != nil {
		t.Fatalf("a resolution at the limit should pass: %v", err)
	}

	issue.Resolution = strings.Repeat("x", rules.TitleMaxLen+1)
	if !errors.Is(rules.ValidateIssue(issue), domain.ErrValidation) {
		t.Fatal("a resolution past the limit should be rejected")
	}
}

func TestValidateResolutionEntry(t *testing.T) {
	entry := domain.Entry{ProjectID: "p1", Kind: domain.EntryResolution, IssueID: "tk1", Body: "We chose a graph."}

	if err := rules.ValidateEntry(entry); err != nil {
		t.Fatalf("want nil, got %v", err)
	}

	detached := entry
	detached.IssueID = ""
	if !errors.Is(rules.ValidateEntry(detached), domain.ErrValidation) {
		t.Fatal("a resolution must attach to an issue")
	}

	onPlan := detached
	onPlan.PlanID = "pl1"
	if !errors.Is(rules.ValidateEntry(onPlan), domain.ErrValidation) {
		t.Fatal("a plan is not something a resolution can answer")
	}

	empty := entry
	empty.Body = ""
	if !errors.Is(rules.ValidateEntry(empty), domain.ErrValidation) {
		t.Fatal("a resolution without a body says nothing")
	}
}

func TestApplyIssueStatus(t *testing.T) {
	answered := domain.Issue{Status: domain.IssueDone, Resolution: "A graph.", ResolutionEntry: "e1"}

	t.Run("reopening clears the answer", func(t *testing.T) {
		got := rules.ApplyIssueStatus(answered, domain.IssueOpen)
		if got.Status != domain.IssueOpen || got.Resolution != "" || got.ResolutionEntry != "" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("staying closed keeps the answer", func(t *testing.T) {
		got := rules.ApplyIssueStatus(answered, domain.IssueCancelled)
		if got.Status != domain.IssueCancelled || got.Resolution != "A graph." || got.ResolutionEntry != "e1" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("an open issue keeps what it carries", func(t *testing.T) {
		open := domain.Issue{Status: domain.IssueOpen, Resolution: "draft"}
		if got := rules.ApplyIssueStatus(open, domain.IssueDone); got.Status != domain.IssueDone || got.Resolution != "draft" {
			t.Fatalf("got %+v", got)
		}
	})
}
