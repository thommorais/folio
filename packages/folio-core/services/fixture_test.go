package services_test

import (
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

type ticketFixture struct {
	issues     *fakeIssues
	plans      *fakePlans
	entries    *fakeEntries
	cycles     *fakeCycles
	issueSvc   *services.IssueService
	planSvc    *services.PlanService
	entrySvc   *services.EntryService
	cycleSvc   *services.CycleService
	owner      ports.Actor
	viewer     ports.Actor
	outside    ports.Actor
	project    domain.ProjectID
	other      domain.ProjectID
}

func newTicketFixture(t *testing.T) *ticketFixture {
	t.Helper()

	projects := newFakeProjects()
	projects.items["p001"] = domain.Project{ID: "p001", Slug: "api", Name: "API", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
		{UserID: "u-viewer", Role: domain.RoleViewer},
	}}
	// A second project the actor also owns, to catch an issue reference that
	// crosses the boundary between two projects the caller can write to.
	projects.items["p002"] = domain.Project{ID: "p002", Slug: "web", Name: "Web", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
	}}

	issues := newFakeIssues()
	plans := newFakePlans()
	entries := newFakeEntries()
	cycles := newFakeCycles()
	guard := services.NewProjectGuard(projects)
	clock := &fakeClock{now: testNow}

	issueSvc := services.NewIssueService(issues, plans, entries, cycles, guard, clock, &seqIDs{prefix: "is"}, nopLogger{})

	return &ticketFixture{
		issues: issues, plans: plans, entries: entries, cycles: cycles,
		issueSvc: issueSvc,
		planSvc:  services.NewPlanService(plans, issues, issueSvc, guard, clock, &seqIDs{prefix: "pl"}, nopLogger{}),
		entrySvc: services.NewEntryService(entries, issues, plans, guard, clock, &seqIDs{prefix: "e"}, nopLogger{}),
		cycleSvc: services.NewCycleService(cycles, issues, guard, clock, &seqIDs{prefix: "cy"}, nopLogger{}),
		owner:      ports.Actor{UserID: "u-owner"},
		viewer:     ports.Actor{UserID: "u-viewer"},
		outside:    ports.Actor{UserID: "u-stranger"},
		project:    "p001",
		other:      "p002",
	}
}

func (f *ticketFixture) ticket(t *testing.T, project domain.ProjectID, title string) domain.Issue {
	t.Helper()
	issue, err := f.issueSvc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: project, Kind: domain.IssueTicket, Title: title,
	})
	if err != nil {
		t.Fatal(err)
	}
	return issue
}
