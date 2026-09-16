package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

type ticketFixture struct {
	tickets    *fakeTickets
	todos      *fakeTodos
	plans      *fakePlans
	journal    *fakeJournal
	cycles     *fakeCycles
	ticketLogs *fakeTicketLogs
	docs       *fakeDocs
	ticketSvc  *services.TicketService
	todoSvc    *services.TodoService
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
	// A second project the actor also owns, to catch a ticket reference that
	// crosses the boundary between two projects the caller can write to.
	projects.items["p002"] = domain.Project{ID: "p002", Slug: "web", Name: "Web", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
	}}

	tickets := newFakeTickets()
	todos := newFakeTodos()
	plans := newFakePlans()
	journal := newFakeJournal()
	cycles := newFakeCycles()
	ticketLogs := newFakeTicketLogs()
	planLogs := newFakePlanLogs()
	todoLogs := newFakeTodoLogs()
	docs := newFakeDocs()
	guard := services.NewProjectGuard(projects)
	clock := &fakeClock{now: testNow}

	todoSvc := services.NewTodoService(todos, plans, tickets, guard, clock, &seqIDs{prefix: "t"}, nopLogger{})

	return &ticketFixture{
		tickets: tickets, todos: todos, plans: plans, journal: journal, docs: docs, cycles: cycles, ticketLogs: ticketLogs,
		ticketSvc:  services.NewTicketService(tickets, todos, plans, journal, docs, cycles, guard, clock, &seqIDs{prefix: "tk"}, nopLogger{}),
		todoSvc:    todoSvc,
		planSvc:    services.NewPlanService(plans, todos, tickets, todoSvc, guard, clock, &seqIDs{prefix: "pl"}, nopLogger{}),
		docSvc:     services.NewDocService(docs, tickets, guard, clock, &seqIDs{prefix: "d"}, nopLogger{}),
		journalSvc: services.NewJournalService(journal, tickets, guard, clock, &seqIDs{prefix: "l"}, nopLogger{}),
		cycleSvc:   services.NewCycleService(cycles, tickets, guard, clock, &seqIDs{prefix: "cy"}, nopLogger{}),
		workLogSvc: services.NewWorkLogService(ticketLogs, planLogs, todoLogs, tickets, plans, todos, cycles, guard, clock, &seqIDs{prefix: "wl"}, nopLogger{}),
		owner:      ports.Actor{UserID: "u-owner"},
		viewer:     ports.Actor{UserID: "u-viewer"},
		outside:    ports.Actor{UserID: "u-stranger"},
		project:    "p001",
		other:      "p002",
	}
}

func (f *ticketFixture) ticket(t *testing.T, project domain.ProjectID, title string) domain.Ticket {
	t.Helper()
	ticket, err := f.ticketSvc.CreateTicket(context.Background(), f.owner, ports.CreateTicketInput{
		ProjectID: project, Title: title,
	})
	if err != nil {
		t.Fatal(err)
	}
	return ticket
}

func TestCreateTicketDerivesSlugAndDefaults(t *testing.T) {
	f := newTicketFixture(t)

	ticket := f.ticket(t, f.project, "Tokens expire an hour early")

	if ticket.Slug != "tokens-expire-an-hour-early" {
		t.Errorf("slug = %q", ticket.Slug)
	}
	if ticket.Status != domain.TicketOpen {
		t.Errorf("status = %q, want open", ticket.Status)
	}
	if ticket.Priority != domain.PriorityMedium {
		t.Errorf("priority = %q, want medium", ticket.Priority)
	}
	if ticket.CreatedBy != "u-owner" {
		t.Errorf("created_by = %q", ticket.CreatedBy)
	}
}

func TestCreateTicketRejectsDuplicateSlugInSameProject(t *testing.T) {
	f := newTicketFixture(t)
	in := ports.CreateTicketInput{ProjectID: f.project, Slug: "auth-bug", Title: "Auth bug"}

	if _, err := f.ticketSvc.CreateTicket(context.Background(), f.owner, in); err != nil {
		t.Fatal(err)
	}
	if _, err := f.ticketSvc.CreateTicket(context.Background(), f.owner, in); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
}

func TestCreateTicketAllowsSameSlugInAnotherProject(t *testing.T) {
	f := newTicketFixture(t)
	in := ports.CreateTicketInput{ProjectID: f.project, Slug: "auth-bug", Title: "Auth bug"}

	if _, err := f.ticketSvc.CreateTicket(context.Background(), f.owner, in); err != nil {
		t.Fatal(err)
	}
	in.ProjectID = f.other
	if _, err := f.ticketSvc.CreateTicket(context.Background(), f.owner, in); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestCreateTicketRefusesAViewer(t *testing.T) {
	f := newTicketFixture(t)

	_, err := f.ticketSvc.CreateTicket(context.Background(), f.viewer, ports.CreateTicketInput{
		ProjectID: f.project, Title: "Nope",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("want forbidden, got %v", err)
	}
}

func TestGetTicketHidesAnotherProjectFromANonMember(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.project, "Internal")

	if _, err := f.ticketSvc.GetTicket(context.Background(), f.outside, ticket.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want not found, got %v", err)
	}
}

func TestGetTicketBySlug(t *testing.T) {
	f := newTicketFixture(t)
	f.ticket(t, f.project, "Tokens expire an hour early")

	got, err := f.ticketSvc.GetTicketBySlug(context.Background(), f.owner, f.project, "tokens-expire-an-hour-early")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Tokens expire an hour early" {
		t.Errorf("title = %q", got.Title)
	}
}

func TestTicketProgressCountsItsTodos(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.project, "Ship search")

	for _, title := range []string{"one", "two", "three"} {
		if _, err := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{
			ProjectID: f.project, TicketID: ticket.ID, Title: title,
		}); err != nil {
			t.Fatal(err)
		}
	}
	todos, err := f.todos.ListByTicket(context.Background(), ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.todoSvc.SetTodoStatus(context.Background(), f.owner, todos[0].ID, domain.TodoDone); err != nil {
		t.Fatal(err)
	}

	got, err := f.ticketSvc.GetTicket(context.Background(), f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Progress.Total != 3 || got.Progress.Done != 1 {
		t.Fatalf("progress = %d/%d, want 1/3", got.Progress.Done, got.Progress.Total)
	}
}

func TestAPlansTodosInheritItsTicket(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.project, "Ship search")

	if _, err := f.planSvc.CreatePlan(context.Background(), f.owner, ports.CreatePlanInput{
		ProjectID: f.project,
		TicketID:  ticket.ID,
		Title:     "Index strategy",
		Todos: []ports.CreateTodoInput{
			{Title: "Compare FTS5 and trigram"},
			{Title: "Benchmark both"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	todos, err := f.todos.ListByTicket(context.Background(), ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 2 {
		t.Fatalf("got %d todos under the ticket, want 2", len(todos))
	}
}

func TestChildrenRejectATicketFromAnotherProject(t *testing.T) {
	f := newTicketFixture(t)
	// The actor owns both projects, so only the ticket check can stop this.
	foreign := f.ticket(t, f.other, "Elsewhere")
	ctx := context.Background()

	t.Run("todo", func(t *testing.T) {
		_, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
			ProjectID: f.project, TicketID: foreign.ID, Title: "Sneak",
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("plan", func(t *testing.T) {
		_, err := f.planSvc.CreatePlan(ctx, f.owner, ports.CreatePlanInput{
			ProjectID: f.project, TicketID: foreign.ID, Title: "Sneak",
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("doc", func(t *testing.T) {
		_, err := f.docSvc.CreateDoc(ctx, f.owner, ports.CreateDocInput{
			ProjectID: f.project, TicketID: foreign.ID, Title: "Sneak",
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("log", func(t *testing.T) {
		_, err := f.journalSvc.WriteJournalEntry(ctx, f.owner, ports.WriteJournalInput{
			ProjectID: f.project, TicketID: foreign.ID, Title: "Sneak",
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})
}

func TestChildrenRejectATicketThatDoesNotExist(t *testing.T) {
	f := newTicketFixture(t)

	_, err := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{
		ProjectID: f.project, TicketID: "tk-nope", Title: "Orphan",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestSetTicketStatusRefusesToRevivACancelledTicket(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.project, "Abandoned work")
	ctx := context.Background()

	if _, err := f.ticketSvc.SetTicketStatus(ctx, f.owner, ticket.ID, domain.TicketCancelled); err != nil {
		t.Fatal(err)
	}
	if _, err := f.ticketSvc.SetTicketStatus(ctx, f.owner, ticket.ID, domain.TicketOpen); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestDeleteTicketDetachesItsChildren(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.project, "Ship search")
	ctx := context.Background()

	todo, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
		ProjectID: f.project, TicketID: ticket.ID, Title: "Write the repository",
	})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := f.planSvc.CreatePlan(ctx, f.owner, ports.CreatePlanInput{
		ProjectID: f.project, TicketID: ticket.ID, Title: "Index strategy",
	})
	if err != nil {
		t.Fatal(err)
	}
	doc, err := f.docSvc.CreateDoc(ctx, f.owner, ports.CreateDocInput{
		ProjectID: f.project, TicketID: ticket.ID, Title: "Search notes",
	})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := f.journalSvc.WriteJournalEntry(ctx, f.owner, ports.WriteJournalInput{
		ProjectID: f.project, TicketID: ticket.ID, Title: "Picked FTS5",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := f.ticketSvc.DeleteTicket(ctx, f.owner, ticket.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := f.ticketSvc.GetTicket(ctx, f.owner, ticket.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ticket still present: %v", err)
	}

	gotTodo, err := f.todoSvc.GetTodo(ctx, f.owner, todo.ID)
	if err != nil {
		t.Fatalf("todo was deleted with the ticket: %v", err)
	}
	if gotTodo.TicketID != "" {
		t.Errorf("todo ticket = %q, want empty", gotTodo.TicketID)
	}

	gotPlan, err := f.planSvc.GetPlan(ctx, f.owner, plan.ID)
	if err != nil {
		t.Fatalf("plan was deleted with the ticket: %v", err)
	}
	if gotPlan.TicketID != "" {
		t.Errorf("plan ticket = %q, want empty", gotPlan.TicketID)
	}

	gotDoc, err := f.docSvc.GetDoc(ctx, f.owner, doc.ID)
	if err != nil {
		t.Fatalf("doc was deleted with the ticket: %v", err)
	}
	if gotDoc.TicketID != "" {
		t.Errorf("doc ticket = %q, want empty", gotDoc.TicketID)
	}

	gotEntry, err := f.journalSvc.GetJournalEntry(ctx, f.owner, entry.ID)
	if err != nil {
		t.Fatalf("log was deleted with the ticket: %v", err)
	}
	if gotEntry.TicketID != "" {
		t.Errorf("log ticket = %q, want empty", gotEntry.TicketID)
	}
}

func TestListTicketsScopesToTheProject(t *testing.T) {
	f := newTicketFixture(t)
	f.ticket(t, f.project, "Mine")
	f.ticket(t, f.other, "Theirs")

	got, err := f.ticketSvc.ListTickets(context.Background(), f.owner, f.project, domain.TicketFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "Mine" {
		t.Fatalf("got %d tickets, want just the project's own", len(got))
	}
}

func TestListTicketsFiltersByStatus(t *testing.T) {
	f := newTicketFixture(t)
	open := f.ticket(t, f.project, "Still open")
	closed := f.ticket(t, f.project, "Done with")
	ctx := context.Background()

	if _, err := f.ticketSvc.SetTicketStatus(ctx, f.owner, closed.ID, domain.TicketClosed); err != nil {
		t.Fatal(err)
	}

	got, err := f.ticketSvc.ListTickets(ctx, f.owner, f.project, domain.TicketFilter{
		Status: []domain.TicketStatus{domain.TicketOpen},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != open.ID {
		t.Fatalf("got %d tickets, want only the open one", len(got))
	}
}

func TestCreateTicketCarriesGraphFields(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	parent := f.ticket(t, f.project, "The map")
	blocker := f.ticket(t, f.project, "Decide the shape")

	child, err := f.ticketSvc.CreateTicket(ctx, f.owner, ports.CreateTicketInput{
		ProjectID: f.project, ParentID: parent.ID, Title: "Depends on the shape",
		DependsOn: []domain.TicketID{blocker.ID}, Wayfinder: domain.WayfinderGrilling,
	})
	if err != nil {
		t.Fatal(err)
	}
	if child.ParentID != parent.ID {
		t.Errorf("parent = %q, want %q", child.ParentID, parent.ID)
	}
	if child.Wayfinder != domain.WayfinderGrilling {
		t.Errorf("wayfinder = %q, want grilling", child.Wayfinder)
	}
	if !child.Blocked {
		t.Error("a ticket created behind an open blocker must report blocked")
	}
}

func TestCreateTicketRejectsAParentFromAnotherProject(t *testing.T) {
	f := newTicketFixture(t)
	foreign := f.ticket(t, f.other, "Elsewhere")

	_, err := f.ticketSvc.CreateTicket(context.Background(), f.owner, ports.CreateTicketInput{
		ProjectID: f.project, ParentID: foreign.ID, Title: "Sneak",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestUpdateTicketRejectsADependencyCycle(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	a := f.ticket(t, f.project, "A")
	b, err := f.ticketSvc.CreateTicket(ctx, f.owner, ports.CreateTicketInput{
		ProjectID: f.project, Title: "B", DependsOn: []domain.TicketID{a.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	deps := []domain.TicketID{b.ID}
	if _, err := f.ticketSvc.UpdateTicket(ctx, f.owner, a.ID, ports.UpdateTicketInput{DependsOn: &deps}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestFrontier(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	mapTicket := f.ticket(t, f.project, "The map")

	child := func(title string, in ports.CreateTicketInput) domain.Ticket {
		t.Helper()
		in.ProjectID, in.ParentID, in.Title = f.project, mapTicket.ID, title
		ticket, err := f.ticketSvc.CreateTicket(ctx, f.owner, in)
		if err != nil {
			t.Fatal(err)
		}
		return ticket
	}

	at := func(ticket domain.Ticket, offset time.Duration) domain.Ticket {
		t.Helper()
		stored := f.tickets.items[ticket.ID]
		stored.CreatedAt = testNow.Add(offset)
		f.tickets.items[ticket.ID] = stored
		return stored
	}

	takeable := child("Takeable", ports.CreateTicketInput{})
	blocker := child("Blocker", ports.CreateTicketInput{})
	child("Blocked", ports.CreateTicketInput{DependsOn: []domain.TicketID{blocker.ID}})
	child("Claimed", ports.CreateTicketInput{Assignee: "u-owner"})
	closed := child("Closed", ports.CreateTicketInput{})
	if _, err := f.ticketSvc.SetTicketStatus(ctx, f.owner, closed.ID, domain.TicketClosed); err != nil {
		t.Fatal(err)
	}

	at(takeable, 2*time.Hour)
	at(blocker, time.Hour)

	frontier, err := f.ticketSvc.Frontier(ctx, f.owner, mapTicket.ID)
	if err != nil {
		t.Fatal(err)
	}

	got := make([]domain.TicketID, 0, len(frontier))
	for _, ticket := range frontier {
		got = append(got, ticket.ID)
	}
	want := []domain.TicketID{blocker.ID, takeable.ID}
	if len(got) != len(want) {
		t.Fatalf("frontier = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("frontier = %v, want %v", got, want)
		}
	}
}

func TestReadsReportBlocked(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	blocker := f.ticket(t, f.project, "Blocker")
	blocked, err := f.ticketSvc.CreateTicket(ctx, f.owner, ports.CreateTicketInput{
		ProjectID: f.project, Title: "Blocked", DependsOn: []domain.TicketID{blocker.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("list", func(t *testing.T) {
		tickets, err := f.ticketSvc.ListTickets(ctx, f.owner, f.project, domain.TicketFilter{})
		if err != nil {
			t.Fatal(err)
		}
		for _, ticket := range tickets {
			if ticket.ID == blocked.ID && !ticket.Blocked {
				t.Error("a ticket behind an open blocker must list as blocked")
			}
			if ticket.ID == blocker.ID && ticket.Blocked {
				t.Error("a ticket with no dependencies must not list as blocked")
			}
		}
	})

	t.Run("get", func(t *testing.T) {
		got, err := f.ticketSvc.GetTicket(ctx, f.owner, blocked.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Blocked {
			t.Error("a ticket behind an open blocker must read as blocked")
		}
	})

	t.Run("clears once the blocker closes", func(t *testing.T) {
		if _, err := f.ticketSvc.SetTicketStatus(ctx, f.owner, blocker.ID, domain.TicketClosed); err != nil {
			t.Fatal(err)
		}
		got, err := f.ticketSvc.GetTicket(ctx, f.owner, blocked.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Blocked {
			t.Error("a closed blocker must not block")
		}
	})
}
