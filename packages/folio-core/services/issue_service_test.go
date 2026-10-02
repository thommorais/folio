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
		svc: services.NewIssueService(issues, plans, newFakeEntries(), newFakeCycles(), guard, clock, &seqIDs{}, nopLogger{}),
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

func TestRejectedLinkLeavesNoIssue(t *testing.T) {
	f := newIssueFixture(t)

	theirs, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.other, Kind: domain.IssueTicket, Title: "Theirs",
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]ports.CreateIssueInput{
		"missing parent":            {ParentID: "nope"},
		"parent in another project": {ParentID: theirs.ID},
		"missing blocker":           {DependsOn: []domain.IssueID{"nope"}},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			in.ProjectID, in.Kind, in.Title = f.project, domain.IssueTicket, name
			if _, err := f.svc.CreateIssue(t.Context(), f.owner, in); err == nil {
				t.Fatal("create should be rejected")
			}
			for _, issue := range f.issues.items {
				if issue.Title == name {
					t.Errorf("rejected create still stored %q", issue.ID)
				}
			}
		})
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

func archive(t *testing.T, f *issueFixture, id domain.IssueID, archived bool) {
	t.Helper()
	if _, err := f.svc.UpdateIssue(t.Context(), f.owner, id, ports.UpdateIssueInput{Archived: &archived}); err != nil {
		t.Fatalf("set archived=%v: %v", archived, err)
	}
}

func listed(t *testing.T, f *issueFixture, view domain.ArchiveView) []domain.Issue {
	t.Helper()
	issues, err := f.svc.ListIssues(t.Context(), f.owner, f.project, domain.IssueFilter{Archive: view})
	if err != nil {
		t.Fatal(err)
	}
	return issues
}

func TestArchivedIssueLeavesTheDefaultListing(t *testing.T) {
	f := newIssueFixture(t)
	keep := f.create(t, domain.IssueTicket, "Keep")
	shelved := f.create(t, domain.IssueTicket, "Shelved")

	archive(t, f, shelved.ID, true)

	live := listed(t, f, domain.ArchiveLive)
	if len(live) != 1 || live[0].ID != keep.ID {
		t.Errorf("default listing = %v, want only %q", live, keep.ID)
	}
}

func TestArchiveViewListsOnlyArchivedIssues(t *testing.T) {
	f := newIssueFixture(t)
	f.create(t, domain.IssueTicket, "Keep")
	shelved := f.create(t, domain.IssueTicket, "Shelved")

	archive(t, f, shelved.ID, true)

	only := listed(t, f, domain.ArchiveOnly)
	if len(only) != 1 || only[0].ID != shelved.ID || !only[0].Archived {
		t.Errorf("archived listing = %v, want only %q marked archived", only, shelved.ID)
	}
}

func TestRestoredIssueReturnsToTheDefaultListing(t *testing.T) {
	f := newIssueFixture(t)
	shelved := f.create(t, domain.IssueTicket, "Shelved")

	archive(t, f, shelved.ID, true)
	archive(t, f, shelved.ID, false)

	if got := listed(t, f, domain.ArchiveLive); len(got) != 1 {
		t.Errorf("default listing has %d issues after restore, want 1", len(got))
	}
	if got := listed(t, f, domain.ArchiveOnly); len(got) != 0 {
		t.Errorf("archived listing has %d issues after restore, want 0", len(got))
	}
}

func TestArchivedIssueIsStillReadableByID(t *testing.T) {
	f := newIssueFixture(t)
	shelved := f.create(t, domain.IssueTicket, "Shelved")
	archive(t, f, shelved.ID, true)

	got, err := f.svc.GetIssue(t.Context(), f.owner, shelved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Archived {
		t.Error("a fetched archived issue must report archived, or the UI cannot offer Restore")
	}
}

func TestArchiveLeavesStatusAlone(t *testing.T) {
	f := newIssueFixture(t)
	issue := f.create(t, domain.IssueTodo, "Half done")
	if _, err := f.svc.UpdateIssue(t.Context(), f.owner, issue.ID, ports.UpdateIssueInput{Status: ptr(domain.IssueInProgress)}); err != nil {
		t.Fatal(err)
	}

	archive(t, f, issue.ID, true)
	archive(t, f, issue.ID, false)

	got, err := f.svc.GetIssue(t.Context(), f.owner, issue.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.IssueInProgress {
		t.Errorf("status after archive and restore = %q, want %q", got.Status, domain.IssueInProgress)
	}
}

func TestArchivedChildLeavesTheBrief(t *testing.T) {
	f := newIssueFixture(t)
	parent := f.create(t, domain.IssueTicket, "Parent")
	kept, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTodo, Title: "Kept", ParentID: parent.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	shelved, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTodo, Title: "Shelved", ParentID: parent.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	archive(t, f, shelved.ID, true)

	brief, err := f.svc.GetIssueBrief(t.Context(), f.owner, parent.ID, ports.BriefOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(brief.Children) != 1 || brief.Children[0].ID != kept.ID {
		t.Errorf("brief children = %v, want only %q", brief.Children, kept.ID)
	}
}

func TestArchivedBlockerNoLongerBlocks(t *testing.T) {
	f := newIssueFixture(t)
	blocker := f.create(t, domain.IssueTodo, "Blocker")
	blocked, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTodo, Title: "Blocked", DependsOn: []domain.IssueID{blocker.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	archive(t, f, blocker.ID, true)

	got, err := f.svc.GetIssue(t.Context(), f.owner, blocked.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Blocked {
		t.Error("an archived blocker is out of the graph and must not hold what it blocks")
	}

	archive(t, f, blocker.ID, false)

	got, err = f.svc.GetIssue(t.Context(), f.owner, blocked.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Blocked {
		t.Error("restoring an open blocker must block again")
	}
}

func TestViewerCannotArchive(t *testing.T) {
	f := newIssueFixture(t)
	issue := f.create(t, domain.IssueTicket, "Mine")
	archived := true

	_, err := f.svc.UpdateIssue(t.Context(), f.viewer, issue.ID, ports.UpdateIssueInput{Archived: &archived})
	if err == nil {
		t.Error("a viewer archived an issue")
	}
}

func TestArchivedChildDoesNotCountTowardProgress(t *testing.T) {
	f := newIssueFixture(t)
	parent := f.create(t, domain.IssueTicket, "Parent")
	for _, title := range []string{"Done one", "Shelved one"} {
		if _, err := f.svc.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
			ProjectID: f.project, Kind: domain.IssueTodo, Title: title, ParentID: parent.ID,
		}); err != nil {
			t.Fatal(err)
		}
	}
	issues := listed(t, f, domain.ArchiveLive)
	for _, i := range issues {
		switch i.Title {
		case "Done one":
			if _, err := f.svc.SetIssueStatus(t.Context(), f.owner, i.ID, domain.IssueDone); err != nil {
				t.Fatal(err)
			}
		case "Shelved one":
			archive(t, f, i.ID, true)
		}
	}

	got, err := f.svc.GetIssue(t.Context(), f.owner, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Progress.Total != 1 || got.Progress.Done != 1 {
		t.Errorf("progress = %d/%d, want 1/1 with the archived child left out", got.Progress.Done, got.Progress.Total)
	}
}
