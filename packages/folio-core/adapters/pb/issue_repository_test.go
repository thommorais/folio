package pb_test

import (
	"testing"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func newIssue(t *testing.T, s scenario, kind domain.IssueKind, slug string) domain.Issue {
	t.Helper()

	repo := pb.NewIssueRepository(s.app)
	created, err := repo.Create(t.Context(), domain.Issue{
		Kind:      kind,
		ProjectID: domain.ProjectID(s.project.Id),
		Slug:      slug,
		Title:     slug,
		Status:    domain.IssueOpen,
		Priority:  domain.PriorityMedium,
	})
	if err != nil {
		t.Fatalf("create %s %s: %v", kind, slug, err)
	}
	return created
}

func TestIssueRoundTrip(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)

	due := s.ticket.GetDateTime("created").Time()
	created, err := repo.Create(t.Context(), domain.Issue{
		Kind:      domain.IssueTodo,
		ProjectID: domain.ProjectID(s.project.Id),
		Slug:      "write-tests",
		Title:     "Write tests",
		Body:      "details here",
		Status:    domain.IssueOpen,
		Priority:  domain.PriorityHigh,
		Size:      domain.SizeM,
		Tags:      []string{"backend"},
		Position:  3,
		DueDate:   &due,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != domain.IssueTodo {
		t.Errorf("kind = %q, want todo", got.Kind)
	}
	if got.Size != domain.SizeM {
		t.Errorf("size = %d, want 3", got.Size)
	}
	if got.Position != 3 {
		t.Errorf("position = %d, want 3", got.Position)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "backend" {
		t.Errorf("tags = %v, want [backend]", got.Tags)
	}
	if got.DueDate == nil {
		t.Error("due date did not survive the round trip")
	}
}

func TestTicketsAndTodosShareOneCollection(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)

	newIssue(t, s, domain.IssueTicket, "a-ticket")
	newIssue(t, s, domain.IssueTodo, "a-todo")

	all, err := repo.List(t.Context(), domain.ProjectID(s.project.Id), domain.IssueFilter{})
	if err != nil {
		t.Fatal(err)
	}
	// The fixture seeds one ticket of its own.
	if len(all) != 3 {
		t.Fatalf("listed %d issues, want 3", len(all))
	}

	todos, err := repo.List(t.Context(), domain.ProjectID(s.project.Id), domain.IssueFilter{Kind: domain.IssueTodo})
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 1 || todos[0].Kind != domain.IssueTodo {
		t.Errorf("kind filter returned %d rows, want 1 todo", len(todos))
	}
}

func TestLinksAcrossKinds(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)

	ticket := newIssue(t, s, domain.IssueTicket, "parent-ticket")
	todo := newIssue(t, s, domain.IssueTodo, "child-todo")

	if err := repo.Link(t.Context(), todo.ID, ticket.ID, domain.LinkParent); err != nil {
		t.Fatal(err)
	}
	// A todo yielding a ticket is the direction the old schema could not express.
	spawned := newIssue(t, s, domain.IssueTicket, "spawned-ticket")
	if err := repo.Link(t.Context(), spawned.ID, todo.ID, domain.LinkParent); err != nil {
		t.Fatal(err)
	}

	children, err := repo.ListByParent(t.Context(), ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 1 || children[0].ID != todo.ID {
		t.Errorf("ticket has %d children, want the todo", len(children))
	}

	under, err := repo.ListByParent(t.Context(), todo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(under) != 1 || under[0].ID != spawned.ID {
		t.Errorf("todo has %d children, want the spawned ticket", len(under))
	}
}

func TestLinkIsIdempotentAndRemovable(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)

	a := newIssue(t, s, domain.IssueTicket, "issue-a")
	b := newIssue(t, s, domain.IssueTicket, "issue-b")

	for range 2 {
		if err := repo.Link(t.Context(), a.ID, b.ID, domain.LinkBlocks); err != nil {
			t.Fatal(err)
		}
	}
	links, err := repo.Links(t.Context(), a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("linking twice produced %d rows, want 1", len(links))
	}

	if err := repo.Unlink(t.Context(), a.ID, b.ID, domain.LinkBlocks); err != nil {
		t.Fatal(err)
	}
	links, err = repo.Links(t.Context(), a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Errorf("unlink left %d rows", len(links))
	}
}

func TestSetLinksReplacesTheSet(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)

	subject := newIssue(t, s, domain.IssueTicket, "subject")
	a := newIssue(t, s, domain.IssueTicket, "dep-a")
	b := newIssue(t, s, domain.IssueTicket, "dep-b")
	c := newIssue(t, s, domain.IssueTicket, "dep-c")

	if err := repo.SetLinks(t.Context(), subject.ID, domain.LinkBlocks, []domain.IssueID{a.ID, b.ID}); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetLinks(t.Context(), subject.ID, domain.LinkBlocks, []domain.IssueID{b.ID, c.ID}); err != nil {
		t.Fatal(err)
	}

	links, err := repo.Links(t.Context(), subject.ID)
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[domain.IssueID]bool)
	for _, l := range links {
		got[l.To] = true
	}
	if got[a.ID] || !got[b.ID] || !got[c.ID] {
		t.Errorf("links are %v, want b and c only", got)
	}
}

func TestBlockedIsDerivedFromLinks(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)

	blocker := newIssue(t, s, domain.IssueTicket, "blocker")
	blocked := newIssue(t, s, domain.IssueTicket, "blocked-one")
	if err := repo.Link(t.Context(), blocker.ID, blocked.ID, domain.LinkBlocks); err != nil {
		t.Fatal(err)
	}

	load := func() []domain.Issue {
		issues, err := repo.List(t.Context(), domain.ProjectID(s.project.Id), domain.IssueFilter{})
		if err != nil {
			t.Fatal(err)
		}
		links, err := repo.LinksOfProject(t.Context(), domain.ProjectID(s.project.Id))
		if err != nil {
			t.Fatal(err)
		}
		rules.ApplyLinks(issues, links)
		return issues
	}

	find := func(issues []domain.Issue, id domain.IssueID) domain.Issue {
		for _, i := range issues {
			if i.ID == id {
				return i
			}
		}
		t.Fatalf("issue %s missing from the listing", id)
		return domain.Issue{}
	}

	if !find(load(), blocked.ID).Blocked {
		t.Fatal("issue is not blocked while its blocker is open")
	}

	blocker.Status = domain.IssueDone
	if _, err := repo.Update(t.Context(), blocker); err != nil {
		t.Fatal(err)
	}
	if find(load(), blocked.ID).Blocked {
		t.Error("issue is still blocked after its blocker closed")
	}
}
