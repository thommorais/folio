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

func TestCurrentCycleAndNextPhase(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	work := f.ticket(t, f.project, "Mobile nav")

	if _, err := f.cycleSvc.CurrentCycle(ctx, f.owner, work.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no cycle yet: want not found, got %v", err)
	}

	first, err := f.cycleSvc.OpenCycle(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []domain.Phase{domain.PhaseDo, domain.PhaseCheck, domain.PhaseAct} {
		got, err := f.cycleSvc.NextPhase(ctx, f.owner, first.ID)
		if err != nil || got.Phase != want {
			t.Fatalf("got %s, %v; want %s", got.Phase, err, want)
		}
	}
	if _, err := f.cycleSvc.NextPhase(ctx, f.owner, first.ID); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("past act: want validation error, got %v", err)
	}
	if _, err := f.cycleSvc.NextPhase(ctx, f.outside, first.ID); errors.Is(err, domain.ErrValidation) || err == nil {
		t.Fatalf("an outsider must be refused before the phase is judged, got %v", err)
	}

	if _, err := f.cycleSvc.ResolveCycle(ctx, f.owner, first.ID, "Shipped"); err != nil {
		t.Fatal(err)
	}
	second, err := f.cycleSvc.OpenCycle(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	current, err := f.cycleSvc.CurrentCycle(ctx, f.owner, work.ID)
	if err != nil || current.ID != second.ID {
		t.Fatalf("current = %+v, %v; want cycle 2", current, err)
	}
	if _, err := f.cycleSvc.CurrentCycle(ctx, f.outside, work.ID); err == nil {
		t.Fatal("an outsider must not read the cycle")
	}
}

func TestAMapHoldsItsCycleOpenToo(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	work := f.ticket(t, f.project, "Wayfinder view")
	theMap := f.child(t, work.ID, domain.WayfinderMap, "Plan the wayfinder view")
	question := f.child(t, theMap.ID, domain.WayfinderGrilling, "Tree or graph")

	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.cycleSvc.SetCycleMap(ctx, f.owner, cycle.ID, theMap.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := f.cycleSvc.ResolveCycle(ctx, f.owner, cycle.ID, "Dropped it"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("resolving past an open decision must be refused, got %v", err)
	}

	if _, err := f.issueSvc.UpdateIssue(ctx, f.owner, question.ID, ports.UpdateIssueInput{
		Status: ptr(domain.IssueCancelled), Resolution: ptr("Out of scope"),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.cycleSvc.ResolveCycle(ctx, f.owner, cycle.ID, "Dropped it"); err != nil {
		t.Fatalf("every decision is closed, got %v", err)
	}
}

func TestBriefCarriesTheCurrentCyclesPlan(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	work := f.ticket(t, f.project, "Wayfinder view")
	theMap := f.child(t, work.ID, domain.WayfinderMap, "Plan the wayfinder view")
	first := f.child(t, theMap.ID, domain.WayfinderGrilling, "Tree or graph")
	f.child(t, theMap.ID, domain.WayfinderResearch, "How dense")
	if _, err := f.issueSvc.CreateIssue(ctx, f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Prototype the rail", Wayfinder: domain.WayfinderPrototype,
		ParentID: theMap.ID, DependsOn: []domain.IssueID{first.ID},
	}); err != nil {
		t.Fatal(err)
	}

	if brief, err := f.issueSvc.GetIssueBrief(ctx, f.owner, work.ID, ports.BriefOptions{}); err != nil || brief.Map != nil {
		t.Fatalf("no cycle yet: want no plan, got %+v, %v", brief.Map, err)
	}

	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.cycleSvc.SetCycleMap(ctx, f.owner, cycle.ID, theMap.ID); err != nil {
		t.Fatal(err)
	}

	brief, err := f.issueSvc.GetIssueBrief(ctx, f.owner, work.ID, ports.BriefOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if brief.Map == nil || brief.Map.Map.ID != theMap.ID {
		t.Fatalf("plan = %+v, want the cycle's map", brief.Map)
	}
	if brief.Map.Open != 3 {
		t.Errorf("open = %d, want 3", brief.Map.Open)
	}
	titles := []string{}
	for _, next := range brief.Map.Frontier {
		titles = append(titles, next.Title)
	}
	if len(titles) != 2 || titles[0] != "Tree or graph" || titles[1] != "How dense" {
		t.Errorf("frontier = %v, want the two unblocked decisions, oldest first", titles)
	}
}
