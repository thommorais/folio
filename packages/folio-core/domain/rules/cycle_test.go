package rules_test

import (
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func validCycle() domain.Cycle {
	return domain.Cycle{ProjectID: "p1", IssueID: "tk1", Ordinal: 1, Phase: domain.PhasePlan}
}

func TestValidateCycle(t *testing.T) {
	t.Run("accepts a valid cycle", func(t *testing.T) {
		if err := rules.ValidateCycle(validCycle()); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("requires a ticket", func(t *testing.T) {
		bad := validCycle()
		bad.IssueID = ""
		if !errors.Is(rules.ValidateCycle(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("rejects an unknown phase", func(t *testing.T) {
		bad := validCycle()
		bad.Phase = "ponder"
		if !errors.Is(rules.ValidateCycle(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("rejects an ordinal below one", func(t *testing.T) {
		bad := validCycle()
		bad.Ordinal = 0
		if !errors.Is(rules.ValidateCycle(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})
}

func TestCanTransitionPhase(t *testing.T) {
	t.Run("walks the loop in order", func(t *testing.T) {
		for _, step := range []struct{ from, to domain.Phase }{
			{domain.PhasePlan, domain.PhaseDo},
			{domain.PhaseDo, domain.PhaseCheck},
			{domain.PhaseCheck, domain.PhaseAct},
		} {
			if err := rules.CanTransitionPhase(step.from, step.to); err != nil {
				t.Errorf("%s -> %s: want nil, got %v", step.from, step.to, err)
			}
		}
	})

	t.Run("rejects skipping a phase", func(t *testing.T) {
		if !errors.Is(rules.CanTransitionPhase(domain.PhasePlan, domain.PhaseCheck), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("rejects going backwards", func(t *testing.T) {
		if !errors.Is(rules.CanTransitionPhase(domain.PhaseCheck, domain.PhaseDo), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("rejects an unknown phase", func(t *testing.T) {
		if !errors.Is(rules.CanTransitionPhase(domain.PhaseAct, "ponder"), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("allows staying put", func(t *testing.T) {
		if err := rules.CanTransitionPhase(domain.PhaseDo, domain.PhaseDo); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}

func TestCurrentCycle(t *testing.T) {
	t.Run("returns the highest ordinal", func(t *testing.T) {
		got, ok := rules.CurrentCycle([]domain.Cycle{
			{ID: "c1", Ordinal: 1, Phase: domain.PhaseAct},
			{ID: "c3", Ordinal: 3, Phase: domain.PhaseDo},
			{ID: "c2", Ordinal: 2, Phase: domain.PhaseAct},
		})
		if !ok {
			t.Fatal("want a cycle")
		}
		if got.ID != "c3" {
			t.Errorf("current = %q, want c3", got.ID)
		}
	})

	t.Run("reports none for an empty set", func(t *testing.T) {
		if _, ok := rules.CurrentCycle(nil); ok {
			t.Fatal("want no cycle")
		}
	})
}

func TestNextOrdinal(t *testing.T) {
	t.Run("starts at one", func(t *testing.T) {
		if got := rules.NextOrdinal(nil); got != 1 {
			t.Errorf("next = %d, want 1", got)
		}
	})

	t.Run("follows the highest ordinal", func(t *testing.T) {
		got := rules.NextOrdinal([]domain.Cycle{{Ordinal: 1}, {Ordinal: 3}, {Ordinal: 2}})
		if got != 4 {
			t.Errorf("next = %d, want 4", got)
		}
	})
}

func TestCheckClosable(t *testing.T) {
	t.Run("allows closing when the current cycle carries a resolution", func(t *testing.T) {
		cycles := []domain.Cycle{{Ordinal: 1, Phase: domain.PhaseAct, Resolution: "Shipped behind a flag"}}
		if err := rules.CheckClosableIssue(domain.IssueDone, cycles); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("refuses to close without a resolution", func(t *testing.T) {
		cycles := []domain.Cycle{{Ordinal: 1, Phase: domain.PhaseAct}}
		if !errors.Is(rules.CheckClosableIssue(domain.IssueDone, cycles), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("refuses to close on a blank resolution", func(t *testing.T) {
		cycles := []domain.Cycle{{Ordinal: 1, Phase: domain.PhaseAct, Resolution: "   "}}
		if !errors.Is(rules.CheckClosableIssue(domain.IssueDone, cycles), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("reads the current cycle, not an earlier one", func(t *testing.T) {
		cycles := []domain.Cycle{
			{Ordinal: 1, Phase: domain.PhaseAct, Resolution: "Shipped"},
			{Ordinal: 2, Phase: domain.PhaseDo},
		}
		if !errors.Is(rules.CheckClosableIssue(domain.IssueDone, cycles), domain.ErrValidation) {
			t.Fatal("a stale resolution from cycle 1 must not close cycle 2")
		}
	})

	t.Run("requires a resolution to cancel too", func(t *testing.T) {
		cycles := []domain.Cycle{{Ordinal: 1, Phase: domain.PhaseDo}}
		if !errors.Is(rules.CheckClosableIssue(domain.IssueCancelled, cycles), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("ignores a ticket with no cycles at all", func(t *testing.T) {
		if err := rules.CheckClosableIssue(domain.IssueDone, nil); err != nil {
			t.Fatalf("a ticket that never opened a cycle must still close, got %v", err)
		}
	})

	t.Run("leaves a non-terminal status alone", func(t *testing.T) {
		cycles := []domain.Cycle{{Ordinal: 1, Phase: domain.PhaseDo}}
		if err := rules.CheckClosableIssue(domain.IssueInProgress, cycles); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}

func TestNextPhase(t *testing.T) {
	for from, want := range map[domain.Phase]domain.Phase{
		domain.PhasePlan: domain.PhaseDo, domain.PhaseDo: domain.PhaseCheck, domain.PhaseCheck: domain.PhaseAct,
	} {
		got, err := rules.NextPhase(from)
		if err != nil || got != want {
			t.Errorf("after %s: got %s, %v; want %s", from, got, err, want)
		}
	}
	if _, err := rules.NextPhase(domain.PhaseAct); !errors.Is(err, domain.ErrValidation) {
		t.Error("act is the last phase and has no next")
	}
}
