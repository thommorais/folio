package services_test

import (
	"context"
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func TestWriteTicketLog(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")

	t.Run("rejects an empty body", func(t *testing.T) {
		if _, err := f.workLogSvc.WriteTicketLog(ctx, f.owner, ticket.ID, "  "); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("writes with no cycle when none is open", func(t *testing.T) {
		got, err := f.workLogSvc.WriteTicketLog(ctx, f.owner, ticket.ID, "Started looking")
		if err != nil {
			t.Fatal(err)
		}
		if got.CycleID != "" {
			t.Errorf("cycle = %q, want empty", got.CycleID)
		}
	})

	t.Run("stamps the current cycle when one is open", func(t *testing.T) {
		cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
		if err != nil {
			t.Fatal(err)
		}
		got, err := f.workLogSvc.WriteTicketLog(ctx, f.owner, ticket.ID, "Mapbox rejects feature-state in a filter")
		if err != nil {
			t.Fatal(err)
		}
		if got.CycleID != cycle.ID {
			t.Errorf("cycle = %q, want %q", got.CycleID, cycle.ID)
		}
	})

	t.Run("a viewer cannot write", func(t *testing.T) {
		if _, err := f.workLogSvc.WriteTicketLog(ctx, f.viewer, ticket.ID, "Sneak"); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("want forbidden, got %v", err)
		}
	})
}

func TestListTicketLogs(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")

	if _, err := f.workLogSvc.WriteTicketLog(ctx, f.owner, ticket.ID, "Before the cycle"); err != nil {
		t.Fatal(err)
	}
	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.workLogSvc.WriteTicketLog(ctx, f.owner, ticket.ID, "During cycle one"); err != nil {
		t.Fatal(err)
	}

	t.Run("returns every entry on the ticket", func(t *testing.T) {
		got, err := f.workLogSvc.ListTicketLogs(ctx, f.owner, ticket.ID, domain.TicketLogFilter{})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Fatalf("got %d entries, want 2", len(got))
		}
	})

	t.Run("narrows to one cycle", func(t *testing.T) {
		got, err := f.workLogSvc.ListTicketLogs(ctx, f.owner, ticket.ID, domain.TicketLogFilter{CycleID: cycle.ID})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].Body != "During cycle one" {
			t.Fatalf("got %+v, want only the cycle-one entry", got)
		}
	})

	t.Run("a stranger cannot read", func(t *testing.T) {
		if _, err := f.workLogSvc.ListTicketLogs(ctx, f.outside, ticket.ID, domain.TicketLogFilter{}); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want not found, got %v", err)
		}
	})
}

func TestPlanAndTodoLogs(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")

	plan, err := f.planSvc.CreatePlan(ctx, f.owner, ports.CreatePlanInput{
		ProjectID: f.project, IssueID: ticket.ID, Title: "Ship the nav",
	})
	if err != nil {
		t.Fatal(err)
	}
	todo, err := f.issueSvc.CreateIssue(ctx, f.owner, ports.CreateIssueInput{
		ProjectID: f.project, ParentID: ticket.ID, Title: "Add the hamburger",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("a plan log records against its plan", func(t *testing.T) {
		got, err := f.workLogSvc.WritePlanLog(ctx, f.owner, plan.ID, "Approach changed")
		if err != nil {
			t.Fatal(err)
		}
		if got.PlanID != plan.ID {
			t.Errorf("plan = %q, want %q", got.PlanID, plan.ID)
		}
	})

	t.Run("a todo log records against its todo", func(t *testing.T) {
		got, err := f.workLogSvc.WriteTodoLog(ctx, f.owner, todo.ID, "Blocked upstream")
		if err != nil {
			t.Fatal(err)
		}
		if got.IssueID != todo.ID {
			t.Errorf("todo = %q, want %q", got.IssueID, todo.ID)
		}
	})

	t.Run("both reject an empty body", func(t *testing.T) {
		if _, err := f.workLogSvc.WritePlanLog(ctx, f.owner, plan.ID, ""); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("plan: want validation error, got %v", err)
		}
		if _, err := f.workLogSvc.WriteTodoLog(ctx, f.owner, todo.ID, ""); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("todo: want validation error, got %v", err)
		}
	})
}
