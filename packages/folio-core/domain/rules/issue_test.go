package rules_test

import (
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func issue(id string, status domain.IssueStatus) domain.Issue {
	return domain.Issue{ID: domain.IssueID(id), Kind: domain.IssueTicket, Status: status}
}

func TestApplyLinksDerivesParentAndBlocked(t *testing.T) {
	issues := []domain.Issue{
		issue("a", domain.IssueOpen),
		issue("b", domain.IssueOpen),
		issue("c", domain.IssueDone),
	}
	links := []domain.IssueLink{
		{From: "b", To: "a", Kind: domain.LinkParent},
		{From: "a", To: "b", Kind: domain.LinkBlocks},
		{From: "c", To: "a", Kind: domain.LinkBlocks},
	}

	rules.ApplyLinks(issues, links)

	if issues[1].ParentID != "a" {
		t.Errorf("b.ParentID = %q, want a", issues[1].ParentID)
	}
	if !issues[1].Blocked {
		t.Error("b should be blocked by open issue a")
	}
	if issues[0].Blocked {
		t.Error("a should not be blocked: its only blocker c is done")
	}
}

func TestApplyLinksIgnoresAbsentEnds(t *testing.T) {
	issues := []domain.Issue{issue("a", domain.IssueOpen)}
	rules.ApplyLinks(issues, []domain.IssueLink{
		{From: "ghost", To: "a", Kind: domain.LinkBlocks},
	})

	if issues[0].Blocked {
		t.Error("a blocker outside the set should not block")
	}
}

func TestApplyLinksIsIdempotent(t *testing.T) {
	issues := []domain.Issue{issue("a", domain.IssueOpen), issue("b", domain.IssueOpen)}
	links := []domain.IssueLink{{From: "a", To: "b", Kind: domain.LinkRelates}}

	rules.ApplyLinks(issues, links)
	rules.ApplyLinks(issues, links)

	if len(issues[0].RelatedTo) != 1 {
		t.Errorf("RelatedTo has %d entries after two passes, want 1", len(issues[0].RelatedTo))
	}
}

func TestCheckNoLinkCycle(t *testing.T) {
	links := []domain.IssueLink{
		{From: "b", To: "a", Kind: domain.LinkParent},
		{From: "c", To: "b", Kind: domain.LinkParent},
	}

	if err := rules.CheckNoLinkCycle("a", "c", domain.LinkParent, links); err == nil {
		t.Error("a -> c closes the loop a <- b <- c and should be rejected")
	}
	if err := rules.CheckNoLinkCycle("d", "a", domain.LinkParent, links); err != nil {
		t.Errorf("d -> a is acyclic: %v", err)
	}
	if err := rules.CheckNoLinkCycle("a", "a", domain.LinkParent, links); err == nil {
		t.Error("a self-link should be rejected")
	}
}

func TestCycleCheckIsPerKind(t *testing.T) {
	links := []domain.IssueLink{{From: "b", To: "a", Kind: domain.LinkParent}}

	if err := rules.CheckNoLinkCycle("a", "b", domain.LinkBlocks, links); err != nil {
		t.Errorf("a blocks-link must not see parent links: %v", err)
	}
}

func TestScoreRanksLowEffortHighValue(t *testing.T) {
	cheap := domain.Issue{Priority: domain.PriorityHigh, Size: domain.SizeXS}
	costly := domain.Issue{Priority: domain.PriorityHigh, Size: domain.SizeXL}
	unsized := domain.Issue{Priority: domain.PriorityHigh}

	if cheap.Score() <= costly.Score() {
		t.Error("a high-priority small issue should outrank a high-priority large one")
	}
	if unsized.Score() != 0 {
		t.Errorf("unsized scored %v, want 0", unsized.Score())
	}
}

func TestSortByScoreIsStable(t *testing.T) {
	issues := []domain.Issue{
		{ID: "low", Priority: domain.PriorityLow, Size: domain.SizeXL},
		{ID: "best", Priority: domain.PriorityHigh, Size: domain.SizeXS},
		{ID: "mid", Priority: domain.PriorityMedium, Size: domain.SizeM},
	}

	rules.SortByScore(issues)

	if issues[0].ID != "best" {
		t.Errorf("first is %q, want best", issues[0].ID)
	}
	if issues[len(issues)-1].ID != "low" {
		t.Errorf("last is %q, want low", issues[len(issues)-1].ID)
	}
}

func TestValidateIssueRejectsBadSize(t *testing.T) {
	base := domain.Issue{
		ProjectID: "p", Kind: domain.IssueTicket, Slug: "a-slug",
		Title: "Title", Status: domain.IssueOpen,
	}

	if err := rules.ValidateIssue(base); err != nil {
		t.Fatalf("valid issue rejected: %v", err)
	}

	sized := base
	sized.Size = 4
	if err := rules.ValidateIssue(sized); err == nil {
		t.Error("size 4 is not on the scale and should be rejected")
	}

	for _, ok := range []domain.Size{1, 2, 3, 5, 8} {
		sized.Size = ok
		if err := rules.ValidateIssue(sized); err != nil {
			t.Errorf("size %d rejected: %v", ok, err)
		}
	}
}

func TestCancelledIssueCannotReopen(t *testing.T) {
	if err := rules.CanTransitionIssue(domain.IssueCancelled, domain.IssueOpen); err == nil {
		t.Error("a cancelled issue should not reopen")
	}
	if err := rules.CanTransitionIssue(domain.IssueDone, domain.IssueOpen); err != nil {
		t.Errorf("a done issue should reopen: %v", err)
	}
}
