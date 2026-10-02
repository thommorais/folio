package rules_test

import (
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func TestValidateHandoffEntry(t *testing.T) {
	entry := domain.Entry{ProjectID: "p1", Kind: domain.EntryHandoff, IssueID: "tk1", Body: "Next: wire the coverage widget."}

	if err := rules.ValidateEntry(entry); err != nil {
		t.Fatalf("want nil, got %v", err)
	}

	detached := entry
	detached.IssueID = ""
	if !errors.Is(rules.ValidateEntry(detached), domain.ErrValidation) {
		t.Fatal("a handoff must attach to an issue")
	}

	onPlan := entry
	onPlan.PlanID = "pl1"
	if !errors.Is(rules.ValidateEntry(onPlan), domain.ErrValidation) {
		t.Fatal("a handoff resumes an issue, not a plan")
	}

	empty := entry
	empty.Body = " \n"
	if !errors.Is(rules.ValidateEntry(empty), domain.ErrValidation) {
		t.Fatal("a handoff without a body leaves nothing to resume from")
	}
}
