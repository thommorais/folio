package services_test

import (
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

func TestSeedWritesTheDemoAndReruns(t *testing.T) {
	projects := newFakeProjects()
	issues := newFakeIssues()
	plans := newFakePlans()
	entries := newFakeEntries()
	cycles := newFakeCycles()
	guard := services.NewProjectGuard(projects)
	clock := &fakeClock{now: testNow}

	issueSvc := services.NewIssueService(issues, plans, entries, cycles, guard, clock, &seqIDs{prefix: "is"}, nopLogger{})
	uc := services.SeedUseCases{
		Projects: services.NewProjectService(projects, guard, clock, &seqIDs{prefix: "pr"}, nopLogger{}),
		Plans:    services.NewPlanService(plans, issues, issueSvc, guard, clock, &seqIDs{prefix: "pl"}, nopLogger{}),
		Issues:   issueSvc,
		Entries:  services.NewEntryService(entries, issues, plans, guard, clock, &seqIDs{prefix: "e"}, nopLogger{}),
		Cycles:   services.NewCycleService(cycles, issues, guard, clock, &seqIDs{prefix: "cy"}, nopLogger{}),
	}
	actor := ports.Actor{UserID: "u-owner"}

	report, err := services.Seed(t.Context(), uc, actor)
	if err != nil {
		t.Fatal(err)
	}
	if report.Projects == 0 || report.Tickets == 0 || report.Cycles == 0 {
		t.Fatalf("report = %+v", report)
	}

	if _, err := services.Seed(t.Context(), uc, actor); err != nil {
		t.Fatalf("a second run must be a no-op, got %v", err)
	}

	project, err := uc.Projects.GetProject(t.Context(), actor, "folio")
	if err != nil {
		t.Fatal(err)
	}
	work, err := uc.Issues.GetIssueBySlug(t.Context(), actor, project.ID, "make-the-work-legible")
	if err != nil {
		t.Fatal(err)
	}
	if work.Wayfinder != "" {
		t.Fatalf("the work ticket is not itself a map, got wayfinder %q", work.Wayfinder)
	}
	current, err := uc.Cycles.CurrentCycle(t.Context(), actor, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Ordinal != 2 || current.MapID == "" {
		t.Fatalf("cycle 2 should be planned by a map, got %+v", current)
	}
	theMap, err := uc.Issues.GetIssue(t.Context(), actor, current.MapID)
	if err != nil {
		t.Fatal(err)
	}
	if theMap.Wayfinder != domain.WayfinderMap || theMap.ParentID != work.ID {
		t.Fatalf("map = %+v, want a map under the work ticket", theMap)
	}
}
