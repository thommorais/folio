package services_test

import (
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func (f *ticketFixture) decision(t *testing.T, w domain.WayfinderType, title string) domain.Issue {
	t.Helper()
	issue, err := f.issueSvc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: title, Wayfinder: w,
	})
	if err != nil {
		t.Fatal(err)
	}
	return issue
}

func ptr[T any](v T) *T { return &v }

func TestDecisionTicketNeedsAResolutionToClose(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.decision(t, domain.WayfinderGrilling, "Tree or graph")

	if _, err := f.issueSvc.SetIssueStatus(t.Context(), f.owner, ticket.ID, domain.IssueDone); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}

	got, err := f.issueSvc.GetIssue(t.Context(), f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.IssueOpen {
		t.Fatalf("a refused close must leave the ticket open, got %s", got.Status)
	}
}

func TestResolutionClosesADecisionInOneCall(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.decision(t, domain.WayfinderGrilling, "Tree or graph")

	detail, err := f.entrySvc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryResolution, IssueID: ticket.ID,
		Body: "The tree hides blockers, the board hides depth.",
	})
	if err != nil {
		t.Fatal(err)
	}

	closed, err := f.issueSvc.UpdateIssue(t.Context(), f.owner, ticket.ID, ports.UpdateIssueInput{
		Status:          ptr(domain.IssueDone),
		Resolution:      ptr("  A graph.  "),
		ResolutionEntry: ptr(detail.ID),
	})
	if err != nil {
		t.Fatal(err)
	}
	if closed.Status != domain.IssueDone || closed.Resolution != "A graph." || closed.ResolutionEntry != detail.ID {
		t.Fatalf("got status %s, resolution %q, entry %q", closed.Status, closed.Resolution, closed.ResolutionEntry)
	}
}

func TestMapAndPlainTicketsCloseWithoutAResolution(t *testing.T) {
	f := newTicketFixture(t)

	for _, ticket := range []domain.Issue{
		f.decision(t, domain.WayfinderMap, "The map"),
		f.ticket(t, f.project, "Mobile nav"),
	} {
		if _, err := f.issueSvc.SetIssueStatus(t.Context(), f.owner, ticket.ID, domain.IssueDone); err != nil {
			t.Errorf("%s: %v", ticket.Title, err)
		}
	}
}

func TestReopeningClearsTheResolution(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.decision(t, domain.WayfinderResearch, "Which queue")

	detail, err := f.entrySvc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryResolution, IssueID: ticket.ID, Body: "SQS, for the dead letter queue.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.issueSvc.UpdateIssue(t.Context(), f.owner, ticket.ID, ports.UpdateIssueInput{
		Status: ptr(domain.IssueDone), Resolution: ptr("SQS"), ResolutionEntry: ptr(detail.ID),
	}); err != nil {
		t.Fatal(err)
	}

	reopened, err := f.issueSvc.SetIssueStatus(t.Context(), f.owner, ticket.ID, domain.IssueOpen)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Resolution != "" || reopened.ResolutionEntry != "" {
		t.Fatalf("want both cleared, got %q and %q", reopened.Resolution, reopened.ResolutionEntry)
	}
}

func TestDecisionTicketCannotBeCreatedClosedWithoutAResolution(t *testing.T) {
	f := newTicketFixture(t)

	if _, err := f.issueSvc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Embed a graph library",
		Wayfinder: domain.WayfinderTask, Status: domain.IssueCancelled,
	}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}

	created, err := f.issueSvc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Embed a graph library",
		Wayfinder: domain.WayfinderTask, Status: domain.IssueCancelled, Resolution: " Out of scope: no map needs it. ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Resolution != "Out of scope: no map needs it." {
		t.Fatalf("got %q", created.Resolution)
	}
}
