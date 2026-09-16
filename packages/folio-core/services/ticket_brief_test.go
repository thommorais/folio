package services_test

import (
	"context"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

func TestTicketBriefCarriesEveryKind(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Ship search")

	if _, err := f.planSvc.CreatePlan(ctx, f.owner, ports.CreatePlanInput{
		ProjectID: f.project, TicketID: ticket.ID, Title: "Index the corpus",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
		ProjectID: f.project, TicketID: ticket.ID, Title: "Add the table",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.journalSvc.WriteJournalEntry(ctx, f.owner, ports.WriteJournalInput{
		ProjectID: f.project, TicketID: ticket.ID, Title: "Picked FTS5",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.docSvc.CreateDoc(ctx, f.owner, ports.CreateDocInput{
		ProjectID: f.project, TicketID: ticket.ID, Title: "Query syntax",
	}); err != nil {
		t.Fatal(err)
	}

	brief, err := f.ticketSvc.GetTicketBrief(ctx, f.owner, ticket.ID, ports.BriefOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if brief.Ticket.Title != "Ship search" {
		t.Errorf("ticket = %q", brief.Ticket.Title)
	}
	for _, c := range []struct {
		kind string
		n    int
	}{
		{"plans", len(brief.Plans)},
		{"todos", len(brief.Todos)},
		{"journal", len(brief.Journal)},
		{"docs", len(brief.Docs)},
	} {
		if c.n != 1 {
			t.Errorf("%s = %d, want 1", c.kind, c.n)
		}
	}
}

func TestTicketBriefPutsOpenTodosFirst(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Ship search")

	for _, title := range []string{"first", "second", "third"} {
		if _, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
			ProjectID: f.project, TicketID: ticket.ID, Title: title,
		}); err != nil {
			t.Fatal(err)
		}
	}
	todos, err := f.todos.ListByTicket(ctx, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Close the first, so a brief that preserved insertion order would fail.
	if _, err := f.todoSvc.SetTodoStatus(ctx, f.owner, todos[0].ID, domain.TodoDone); err != nil {
		t.Fatal(err)
	}

	brief, err := f.ticketSvc.GetTicketBrief(ctx, f.owner, ticket.ID, ports.BriefOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(brief.Todos) != 3 {
		t.Fatalf("todos = %d, want 3", len(brief.Todos))
	}
	if brief.Todos[0].Status.IsTerminal() {
		t.Errorf("first todo is %s, want an open one", brief.Todos[0].Status)
	}
	if !brief.Todos[2].Status.IsTerminal() {
		t.Errorf("last todo is %s, want the done one", brief.Todos[2].Status)
	}
}

func TestTicketBriefCapsTheLogs(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Ship search")

	for _, title := range []string{"one", "two", "three"} {
		if _, err := f.journalSvc.WriteJournalEntry(ctx, f.owner, ports.WriteJournalInput{
			ProjectID: f.project, TicketID: ticket.ID, Title: title,
		}); err != nil {
			t.Fatal(err)
		}
	}

	// The fake ignores Limit, so the assertion is on the bound the service
	// passes down rather than on the row count it would return.
	if _, err := f.ticketSvc.GetTicketBrief(ctx, f.owner, ticket.ID, ports.BriefOptions{RecentJournal: 2}); err != nil {
		t.Fatal(err)
	}
	if got := f.journal.lastFilter.Limit; got != 2 {
		t.Errorf("limit = %d, want the requested 2", got)
	}

	if _, err := f.ticketSvc.GetTicketBrief(ctx, f.owner, ticket.ID, ports.BriefOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := f.journal.lastFilter.Limit; got != services.DefaultRecentJournal {
		t.Errorf("default limit = %d, want %d", got, services.DefaultRecentJournal)
	}
}

func TestTicketBriefHidesAnotherProjectFromANonMember(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.project, "Ship search")

	_, err := f.ticketSvc.GetTicketBrief(context.Background(), f.outside, ticket.ID, ports.BriefOptions{})
	if err == nil {
		t.Fatal("want an error for a non-member")
	}
}

func TestTicketBriefWithNoChildrenReturnsEmptySlices(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.project, "Nothing filed yet")

	brief, err := f.ticketSvc.GetTicketBrief(context.Background(), f.owner, ticket.ID, ports.BriefOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// Nil marshals to JSON null, which forces every client to branch.
	if brief.Plans == nil || brief.Todos == nil || brief.Journal == nil || brief.Docs == nil {
		t.Errorf("want empty slices, got plans=%v todos=%v logs=%v docs=%v",
			brief.Plans, brief.Todos, brief.Journal, brief.Docs)
	}
}
