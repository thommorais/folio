package services_test

import (
	"context"
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func TestOpenCycle(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")

	first, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Ordinal != 1 {
		t.Errorf("ordinal = %d, want 1", first.Ordinal)
	}
	if first.Phase != domain.PhasePlan {
		t.Errorf("phase = %q, want plan", first.Phase)
	}

	t.Run("refuses a second cycle while one is unresolved", func(t *testing.T) {
		if _, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("want conflict, got %v", err)
		}
	})

	t.Run("opens the next once the current resolves", func(t *testing.T) {
		if _, err := f.cycleSvc.ResolveCycle(ctx, f.owner, first.ID, "Shipped behind a flag"); err != nil {
			t.Fatal(err)
		}
		second, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
		if err != nil {
			t.Fatal(err)
		}
		if second.Ordinal != 2 {
			t.Errorf("ordinal = %d, want 2", second.Ordinal)
		}
		if second.Phase != domain.PhasePlan {
			t.Errorf("a new cycle starts at plan, got %q", second.Phase)
		}
	})
}

func TestAdvancePhase(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")
	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("advances one step", func(t *testing.T) {
		got, err := f.cycleSvc.AdvancePhase(ctx, f.owner, cycle.ID, domain.PhaseDo)
		if err != nil {
			t.Fatal(err)
		}
		if got.Phase != domain.PhaseDo {
			t.Errorf("phase = %q, want do", got.Phase)
		}
	})

	t.Run("refuses to skip", func(t *testing.T) {
		if _, err := f.cycleSvc.AdvancePhase(ctx, f.owner, cycle.ID, domain.PhaseAct); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})
}

func TestResolveCycle(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")
	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("rejects a blank resolution", func(t *testing.T) {
		if _, err := f.cycleSvc.ResolveCycle(ctx, f.owner, cycle.ID, "   "); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("records the resolution and closes the cycle", func(t *testing.T) {
		got, err := f.cycleSvc.ResolveCycle(ctx, f.owner, cycle.ID, "Shipped behind a flag")
		if err != nil {
			t.Fatal(err)
		}
		if got.Resolution != "Shipped behind a flag" {
			t.Errorf("resolution = %q", got.Resolution)
		}
		if !got.IsClosed() {
			t.Error("a resolved cycle must be closed")
		}
	})
}

func TestCyclePermissions(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")

	t.Run("a viewer cannot open a cycle", func(t *testing.T) {
		if _, err := f.cycleSvc.OpenCycle(ctx, f.viewer, ticket.ID); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("want forbidden, got %v", err)
		}
	})

	t.Run("a stranger cannot read cycles", func(t *testing.T) {
		if _, err := f.cycleSvc.ListCycles(ctx, f.outside, ticket.ID); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want not found, got %v", err)
		}
	})
}

func TestClosingATicketRequiresAResolution(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")
	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.issueSvc.SetIssueStatus(ctx, f.owner, ticket.ID, domain.IssueDone); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}

	if _, err := f.cycleSvc.ResolveCycle(ctx, f.owner, cycle.ID, "Shipped behind a flag"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.issueSvc.SetIssueStatus(ctx, f.owner, ticket.ID, domain.IssueDone); err != nil {
		t.Fatalf("a resolved cycle must let the ticket close, got %v", err)
	}
}

func TestATicketWithNoCyclesStillCloses(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.project, "Never used PDCA")

	if _, err := f.issueSvc.SetIssueStatus(context.Background(), f.owner, ticket.ID, domain.IssueDone); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

var _ = ports.Actor{}

func TestTicketReportsItsCurrentCycle(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")

	t.Run("reports nothing before any cycle opens", func(t *testing.T) {
		got, err := f.issueSvc.GetIssue(ctx, f.owner, ticket.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Cycle != 0 || got.Phase != "" {
			t.Errorf("cycle = %d phase = %q, want 0 and empty", got.Cycle, got.Phase)
		}
	})

	first, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("reports the open cycle and its phase", func(t *testing.T) {
		if _, err := f.cycleSvc.AdvancePhase(ctx, f.owner, first.ID, domain.PhaseDo); err != nil {
			t.Fatal(err)
		}
		got, err := f.issueSvc.GetIssue(ctx, f.owner, ticket.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Cycle != 1 || got.Phase != domain.PhaseDo {
			t.Errorf("cycle = %d phase = %q, want 1 and do", got.Cycle, got.Phase)
		}
	})

	t.Run("follows the latest cycle after a reopen", func(t *testing.T) {
		if _, err := f.cycleSvc.ResolveCycle(ctx, f.owner, first.ID, "Shipped"); err != nil {
			t.Fatal(err)
		}
		if _, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID); err != nil {
			t.Fatal(err)
		}
		got, err := f.issueSvc.GetIssue(ctx, f.owner, ticket.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Cycle != 2 || got.Phase != domain.PhasePlan {
			t.Errorf("cycle = %d phase = %q, want 2 and plan", got.Cycle, got.Phase)
		}
	})
}

func TestUpdateTicketReturnsTheCurrentCycle(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()
	ticket := f.ticket(t, f.project, "Mobile nav")
	if _, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID); err != nil {
		t.Fatal(err)
	}

	got, err := f.issueSvc.UpdateIssue(ctx, f.owner, ticket.ID, ports.UpdateIssueInput{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Cycle != 1 || got.Phase != domain.PhasePlan {
		t.Errorf("cycle = %d phase = %q, want 1 and plan", got.Cycle, got.Phase)
	}
}
