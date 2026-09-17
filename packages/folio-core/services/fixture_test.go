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
	journal    *fakeJournal
	cycles     *fakeCycles
	ticketLogs *fakeTicketLogs
	docs       *fakeDocs
	issueSvc   *services.IssueService
	planSvc    *services.PlanService
	docSvc     *services.DocService
	journalSvc *services.JournalService
	cycleSvc   *services.CycleService
	workLogSvc *services.WorkLogService
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
	journal := newFakeJournal()
	cycles := newFakeCycles()
	ticketLogs := newFakeTicketLogs()
	planLogs := newFakePlanLogs()
	todoLogs := newFakeTodoLogs()
	docs := newFakeDocs()
	guard := services.NewProjectGuard(projects)
	clock := &fakeClock{now: testNow}

	issueSvc := services.NewIssueService(issues, plans, journal, docs, cycles, guard, clock, &seqIDs{prefix: "is"}, nopLogger{})

	return &ticketFixture{
		issues: issues, plans: plans, journal: journal, docs: docs, cycles: cycles, ticketLogs: ticketLogs,
		issueSvc:   issueSvc,
		planSvc:    services.NewPlanService(plans, issues, issueSvc, guard, clock, &seqIDs{prefix: "pl"}, nopLogger{}),
		docSvc:     services.NewDocService(docs, issues, guard, clock, &seqIDs{prefix: "d"}, nopLogger{}),
		journalSvc: services.NewJournalService(journal, issues, guard, clock, &seqIDs{prefix: "l"}, nopLogger{}),
		cycleSvc:   services.NewCycleService(cycles, issues, guard, clock, &seqIDs{prefix: "cy"}, nopLogger{}),
		workLogSvc: services.NewWorkLogService(ticketLogs, planLogs, todoLogs, issues, plans, cycles, guard, clock, &seqIDs{prefix: "wl"}, nopLogger{}),
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
