package services_test

import (
	"testing"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

type issueFixture struct {
	issues  *fakeIssues
	plans   *fakePlans
	svc     *services.IssueService
	owner   ports.Actor
	viewer  ports.Actor
	outside ports.Actor
	project domain.ProjectID
	other   domain.ProjectID
}

func newIssueFixture(t *testing.T) *issueFixture {
	t.Helper()

	projects := newFakeProjects()
	projects.items["p001"] = domain.Project{ID: "p001", Slug: "api", Name: "API", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
		{UserID: "u-viewer", Role: domain.RoleViewer},
	}}
	projects.items["p002"] = domain.Project{ID: "p002", Slug: "web", Name: "Web", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
	}}

	issues := newFakeIssues()
	plans := newFakePlans()
	guard := services.NewProjectGuard(projects)
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}

	return &issueFixture{
		issues: issues,
		plans:  plans,
		svc: services.NewIssueService(issues, plans, newFakeJournal(), newFakeDocs(),
			newFakeCycles(), guard, clock, &seqIDs{}, nopLogger{}),
		owner:   ports.Actor{UserID: "u-owner"},
		viewer:  ports.Actor{UserID: "u-viewer"},
		outside: ports.Actor{UserID: "u-nobody"},
		project: "p001",
		other:   "p002",
	}
}

func (f *issueFixture) create(t *testing.T, kind domain.IssueKind, title string) domain.Issue {
	t.Helper()
	issue, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: kind, Title: title,
	})
	if err != nil {
		t.Fatalf("create %s %q: %v", kind, title, err)
	}
	return issue
}

func TestTodoCanYieldATicket(t *testing.T) {
	f := newIssueFixture(t)

	todo := f.create(t, domain.IssueTodo, "Investigate the bug")
	ticket, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project,
		Kind:      domain.IssueTicket,
		Title:     "Fix the bug",
		ParentID:  todo.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if ticket.ParentID != todo.ID {
		t.Errorf("ticket parent is %q, want the todo %q", ticket.ParentID, todo.ID)
	}

	children, err := f.svc.Frontier(t.Context(), f.owner, todo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 1 || children[0].ID != ticket.ID {
		t.Errorf("todo's frontier has %d issues, want the spawned ticket", len(children))
	}
}

func TestTicketCanYieldATodo(t *testing.T) {
	f := newIssueFixture(t)

	ticket := f.create(t, domain.IssueTicket, "Ship the feature")
	todo, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTodo, Title: "Write the migration", ParentID: ticket.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if todo.ParentID != ticket.ID {
		t.Errorf("todo parent is %q, want the ticket", todo.ParentID)
	}
}

func TestBlockingAcrossKinds(t *testing.T) {
	f := newIssueFixture(t)

	blocker := f.create(t, domain.IssueTodo, "Blocker todo")
	blocked, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Blocked ticket",
		DependsOn: []domain.IssueID{blocker.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !blocked.Blocked {
		t.Error("ticket should be blocked by an open todo")
	}

	if _, err := f.svc.SetIssueStatus(t.Context(), f.owner, blocker.ID, domain.IssueDone); err != nil {
		t.Fatal(err)
	}
	after, err := f.svc.GetIssue(t.Context(), f.owner, blocked.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Blocked {
		t.Error("ticket should unblock once the todo is done")
	}
}

func TestLinkCannotCrossProjects(t *testing.T) {
	f := newIssueFixture(t)

	mine := f.create(t, domain.IssueTicket, "Mine")
	theirs, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.other, Kind: domain.IssueTicket, Title: "Theirs",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = f.svc.LinkIssues(t.Context(), f.owner, mine.ID, theirs.ID, domain.LinkBlocks)
	if err == nil {
		t.Error("a link across projects should be rejected even when the actor owns both")
	}
}

func TestParentCycleIsRejected(t *testing.T) {
	f := newIssueFixture(t)

	a := f.create(t, domain.IssueTicket, "A")
	b, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "B", ParentID: a.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := f.svc.LinkIssues(t.Context(), f.owner, a.ID, b.ID, domain.LinkParent); err == nil {
		t.Error("a -> b closes the parent loop and should be rejected")
	}
}

func TestViewerCannotWrite(t *testing.T) {
	f := newIssueFixture(t)

	_, err := f.svc.CreateIssue(t.Context(), f.viewer, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTodo, Title: "Nope",
	})
	if err == nil {
		t.Error("a viewer should not create issues")
	}
}

func TestOutsiderSeesNothing(t *testing.T) {
	f := newIssueFixture(t)
	issue := f.create(t, domain.IssueTicket, "Secret")

	if _, err := f.svc.GetIssue(t.Context(), f.outside, issue.ID); err == nil {
		t.Error("a non-member should not read an issue")
	}
}

func TestSizeIsValidatedAndScored(t *testing.T) {
	f := newIssueFixture(t)

	_, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Bad size", Size: 4,
	})
	if err == nil {
		t.Error("size 4 is off the scale and should be rejected")
	}

	cheap, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Cheap win",
		Size: domain.SizeXS, Priority: domain.PriorityHigh,
	})
	if err != nil {
		t.Fatal(err)
	}
	costly, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Big effort",
		Size: domain.SizeXL, Priority: domain.PriorityHigh,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cheap.Score() <= costly.Score() {
		t.Error("a small high-priority issue should outscore a large one")
	}
}

func TestKindFilterSeparatesTicketsAndTodos(t *testing.T) {
	f := newIssueFixture(t)
	f.create(t, domain.IssueTicket, "A ticket")
	f.create(t, domain.IssueTodo, "A todo")

	todos, err := f.svc.ListIssues(t.Context(), f.owner, f.project, domain.IssueFilter{Kind: domain.IssueTodo})
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 1 || todos[0].Kind != domain.IssueTodo {
		t.Errorf("kind filter returned %d issues, want 1 todo", len(todos))
	}

	all, err := f.svc.ListIssues(t.Context(), f.owner, f.project, domain.IssueFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("unfiltered listing returned %d issues, want 2", len(all))
	}
}

func TestSlugCollidesAcrossKinds(t *testing.T) {
	f := newIssueFixture(t)
	f.create(t, domain.IssueTicket, "Same name")

	_, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTodo, Title: "Same name",
	})
	if err == nil {
		t.Error("a todo must not reuse a ticket's slug: they now share one namespace")
	}
}

func TestBatchKeepsGoodItems(t *testing.T) {
	f := newIssueFixture(t)

	result, err := f.svc.CreateIssues(t.Context(), f.owner, f.project, []ports.CreateIssueInput{
		{Kind: domain.IssueTodo, Title: "Good one"},
		{Kind: domain.IssueTodo, Title: "Bad size", Size: 7},
		{Kind: domain.IssueTodo, Title: "Another good"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Created) != 2 {
		t.Errorf("created %d issues, want 2", len(result.Created))
	}
	if len(result.Errors) != 1 || result.Errors[0].Index != 1 {
		t.Errorf("errors = %v, want one at index 1", result.Errors)
	}
}

func TestDeleteDetachesRatherThanCascades(t *testing.T) {
	f := newIssueFixture(t)

	parent := f.create(t, domain.IssueTicket, "Parent")
	child, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTodo, Title: "Child", ParentID: parent.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := f.svc.DeleteIssue(t.Context(), f.owner, parent.ID); err != nil {
		t.Fatal(err)
	}
	survivor, err := f.svc.GetIssue(t.Context(), f.owner, child.ID)
	if err != nil {
		t.Fatalf("child did not survive its parent: %v", err)
	}
	if survivor.ParentID != "" {
		t.Errorf("child still points at the deleted parent %q", survivor.ParentID)
	}
}
