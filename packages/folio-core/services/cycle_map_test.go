package services_test

import (
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func (f *ticketFixture) child(t *testing.T, parent domain.IssueID, w domain.WayfinderType, title string) domain.Issue {
	t.Helper()
	issue, err := f.issueSvc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: title, Wayfinder: w, ParentID: parent,
	})
	if err != nil {
		t.Fatal(err)
	}
	return issue
}

func TestAMapHoldsItsCycleInPlan(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	work := f.ticket(t, f.project, "Wayfinder view")
	theMap := f.child(t, work.ID, domain.WayfinderMap, "Plan the wayfinder view")
	question := f.child(t, theMap.ID, domain.WayfinderGrilling, "Tree or graph")

	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	linked, err := f.cycleSvc.SetCycleMap(ctx, f.owner, cycle.ID, theMap.ID)
	if err != nil {
		t.Fatal(err)
	}
	if linked.MapID != theMap.ID {
		t.Fatalf("map = %q, want %q", linked.MapID, theMap.ID)
	}

	if _, err := f.cycleSvc.AdvancePhase(ctx, f.owner, cycle.ID, domain.PhaseDo); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("an open decision must hold plan, got %v", err)
	}

	if _, err := f.issueSvc.UpdateIssue(ctx, f.owner, question.ID, ports.UpdateIssueInput{
		Status: ptr(domain.IssueDone), Resolution: ptr("A graph."),
	}); err != nil {
		t.Fatal(err)
	}

	advanced, err := f.cycleSvc.AdvancePhase(ctx, f.owner, cycle.ID, domain.PhaseDo)
	if err != nil {
		t.Fatalf("the way is clear, got %v", err)
	}
	if advanced.Phase != domain.PhaseDo {
		t.Fatalf("phase = %s", advanced.Phase)
	}
}

func TestSetCycleMapRefusesWhatIsNotAMap(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	work := f.ticket(t, f.project, "Wayfinder view")
	question := f.child(t, work.ID, domain.WayfinderGrilling, "Tree or graph")

	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.cycleSvc.SetCycleMap(ctx, f.owner, cycle.ID, question.ID); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
	if _, err := f.cycleSvc.SetCycleMap(ctx, f.viewer, cycle.ID, question.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a viewer cannot link a map, got %v", err)
	}
}

func TestDecisionTicketCannotOpenACycle(t *testing.T) {
	f := newTicketFixture(t)
	question := f.decision(t, domain.WayfinderResearch, "How dense does the graph get")

	if _, err := f.cycleSvc.OpenCycle(t.Context(), f.owner, question.ID); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}
