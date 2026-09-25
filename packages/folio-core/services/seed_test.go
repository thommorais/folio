package services_test

import (
	"testing"

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
}
