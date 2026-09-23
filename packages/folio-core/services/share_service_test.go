package services_test

import (
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

type shareFixture struct {
	*ticketFixture
	shares *fakeShares
	svc    *services.ShareService
}

func newShareFixture(t *testing.T) *shareFixture {
	t.Helper()
	f := newTicketFixture(t)
	shares := newFakeShares()
	return &shareFixture{
		ticketFixture: f,
		shares:        shares,
		svc: services.NewShareService(shares, f.issues, f.plans, f.issueSvc, f.planSvc, f.guard,
			&fakeClock{now: testNow}, &seqIDs{prefix: "sh"}, &seqTokens{}, nopLogger{}),
	}
}

func (f *shareFixture) plan(t *testing.T, title string, todos ...string) domain.Plan {
	t.Helper()
	in := ports.CreatePlanInput{ProjectID: f.project, Title: title}
	for _, todo := range todos {
		in.Todos = append(in.Todos, ports.CreateIssueInput{Title: todo})
	}
	plan, err := f.planSvc.CreatePlan(t.Context(), f.owner, in)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func TestSharedTicketOpensItsBriefWithoutAnActor(t *testing.T) {
	f := newShareFixture(t)
	ticket := f.ticket(t, f.project, "Checkout flow")
	if _, err := f.entrySvc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryJournal, IssueID: ticket.ID, Title: "Payments", Body: "picked stripe",
	}); err != nil {
		t.Fatal(err)
	}

	share, err := f.svc.ShareIssue(t.Context(), f.owner, ticket.ID, "  vendor  ")
	if err != nil {
		t.Fatal(err)
	}
	if share.Token == "" || share.CreatedBy != f.owner.UserID || share.Label != "vendor" {
		t.Fatalf("share = %+v, want a token, the caller as creator and a trimmed label", share)
	}

	item, err := f.svc.OpenShare(t.Context(), share.Token)
	if err != nil {
		t.Fatal(err)
	}
	if item.Brief == nil || item.Brief.Issue.ID != ticket.ID {
		t.Fatalf("opened %+v, want the ticket's brief", item)
	}
	if len(item.Brief.Journal) != 1 {
		t.Errorf("brief carries %d journal entries, want 1", len(item.Brief.Journal))
	}
	if item.Plan != nil {
		t.Error("an issue share opened a plan")
	}
}

func TestSharedPlanOpensWithItsTodos(t *testing.T) {
	f := newShareFixture(t)
	plan := f.plan(t, "Migrate auth", "drop sessions", "add tokens")

	share, err := f.svc.SharePlan(t.Context(), f.owner, plan.ID, "reviewer")
	if err != nil {
		t.Fatal(err)
	}
	item, err := f.svc.OpenShare(t.Context(), share.Token)
	if err != nil {
		t.Fatal(err)
	}
	if item.Plan == nil || item.Plan.ID != plan.ID {
		t.Fatalf("opened %+v, want the plan", item)
	}
	if item.Plan.Progress.Total != 2 || len(item.Todos) != 2 {
		t.Errorf("plan has progress %+v and %d todos, want 2 of each", item.Plan.Progress, len(item.Todos))
	}
	if item.Brief != nil {
		t.Error("a plan share opened an issue brief")
	}
}

func TestOnlyWritersCanShare(t *testing.T) {
	f := newShareFixture(t)
	ticket := f.ticket(t, f.project, "Checkout flow")
	plan := f.plan(t, "Migrate auth")

	for _, tc := range []struct {
		name  string
		actor ports.Actor
		want  error
	}{
		{"viewer", f.viewer, domain.ErrForbidden},
		{"outsider", f.outside, domain.ErrNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := f.svc.ShareIssue(t.Context(), tc.actor, ticket.ID, "x"); !errors.Is(err, tc.want) {
				t.Errorf("ShareIssue err = %v, want %v", err, tc.want)
			}
			if _, err := f.svc.SharePlan(t.Context(), tc.actor, plan.ID, "x"); !errors.Is(err, tc.want) {
				t.Errorf("SharePlan err = %v, want %v", err, tc.want)
			}
		})
	}
	if len(f.shares.items) != 0 {
		t.Errorf("%d shares written by callers who may not share", len(f.shares.items))
	}
}

func TestShareNeedsALabel(t *testing.T) {
	f := newShareFixture(t)
	ticket := f.ticket(t, f.project, "Checkout flow")

	if _, err := f.svc.ShareIssue(t.Context(), f.owner, ticket.ID, "   "); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("err = %v, want a validation error", err)
	}
}

func TestSharingAMissingTargetIsNotFound(t *testing.T) {
	f := newShareFixture(t)

	if _, err := f.svc.ShareIssue(t.Context(), f.owner, "nope", "x"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ShareIssue err = %v, want not found", err)
	}
	if _, err := f.svc.SharePlan(t.Context(), f.owner, "nope", "x"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("SharePlan err = %v, want not found", err)
	}
}

func TestUnknownOrRevokedTokenOpensNothing(t *testing.T) {
	f := newShareFixture(t)
	ticket := f.ticket(t, f.project, "Checkout flow")
	share, err := f.svc.ShareIssue(t.Context(), f.owner, ticket.ID, "vendor")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.svc.RevokeShare(t.Context(), f.owner, share.ID); err != nil {
		t.Fatal(err)
	}

	for _, token := range []string{share.Token, "", "made-up"} {
		if _, err := f.svc.OpenShare(t.Context(), token); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("OpenShare(%q) err = %v, want not found", token, err)
		}
	}
}

func TestOpeningAShareRecordsTheVisit(t *testing.T) {
	f := newShareFixture(t)
	ticket := f.ticket(t, f.project, "Checkout flow")
	share, err := f.svc.ShareIssue(t.Context(), f.owner, ticket.ID, "vendor")
	if err != nil {
		t.Fatal(err)
	}
	if share.LastAccessedAt != nil {
		t.Fatal("a new share reports a visit")
	}

	if _, err := f.svc.OpenShare(t.Context(), share.Token); err != nil {
		t.Fatal(err)
	}
	listed, err := f.svc.ListShares(t.Context(), f.owner, f.project)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].LastAccessedAt == nil || !listed[0].LastAccessedAt.Equal(testNow) {
		t.Errorf("listed %+v, want one share visited at %v", listed, testNow)
	}
}

func TestSharesAreListedToMembersOnly(t *testing.T) {
	f := newShareFixture(t)
	ticket := f.ticket(t, f.project, "Checkout flow")
	if _, err := f.svc.ShareIssue(t.Context(), f.owner, ticket.ID, "vendor"); err != nil {
		t.Fatal(err)
	}

	listed, err := f.svc.ListShares(t.Context(), f.viewer, f.project)
	if err != nil || len(listed) != 1 {
		t.Errorf("viewer listed %d shares (err %v), want 1", len(listed), err)
	}
	if _, err := f.svc.ListShares(t.Context(), f.outside, f.project); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("outsider err = %v, want not found", err)
	}
}

func TestOnlyWritersCanRevoke(t *testing.T) {
	f := newShareFixture(t)
	ticket := f.ticket(t, f.project, "Checkout flow")
	share, err := f.svc.ShareIssue(t.Context(), f.owner, ticket.ID, "vendor")
	if err != nil {
		t.Fatal(err)
	}

	if err := f.svc.RevokeShare(t.Context(), f.viewer, share.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("viewer err = %v, want forbidden", err)
	}
	if err := f.svc.RevokeShare(t.Context(), f.outside, share.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("outsider err = %v, want not found", err)
	}
	if _, err := f.svc.OpenShare(t.Context(), share.Token); err != nil {
		t.Errorf("share stopped opening after refused revokes: %v", err)
	}
}
