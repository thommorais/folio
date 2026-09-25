package rules_test

import (
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func TestCheckCycleMap(t *testing.T) {
	cycle := domain.Cycle{ProjectID: "p1", IssueID: "tk1", Ordinal: 1, Phase: domain.PhasePlan}
	theMap := domain.Issue{ID: "m1", ProjectID: "p1", Kind: domain.IssueTicket, Wayfinder: domain.WayfinderMap}

	t.Run("links a map to a cycle in plan", func(t *testing.T) {
		if err := rules.CheckCycleMap(cycle, theMap); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("refuses a ticket that is not a map", func(t *testing.T) {
		grilling := theMap
		grilling.Wayfinder = domain.WayfinderGrilling
		if !errors.Is(rules.CheckCycleMap(cycle, grilling), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("refuses a map from another project", func(t *testing.T) {
		elsewhere := theMap
		elsewhere.ProjectID = "p2"
		if !errors.Is(rules.CheckCycleMap(cycle, elsewhere), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("refuses the ticket that owns the cycle", func(t *testing.T) {
		self := theMap
		self.ID = cycle.IssueID
		if !errors.Is(rules.CheckCycleMap(cycle, self), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("refuses a cycle past plan", func(t *testing.T) {
		doing := cycle
		doing.Phase = domain.PhaseDo
		if !errors.Is(rules.CheckCycleMap(doing, theMap), domain.ErrValidation) {
			t.Fatal("a map plans a cycle, so it has to arrive during plan")
		}
	})
}

func TestCheckPlanClear(t *testing.T) {
	planned := domain.Cycle{Phase: domain.PhasePlan, MapID: "m1"}
	decision := func(status domain.IssueStatus) domain.Issue {
		return domain.Issue{Kind: domain.IssueTicket, Wayfinder: domain.WayfinderGrilling, Status: status}
	}

	t.Run("holds plan while a decision is open", func(t *testing.T) {
		for _, status := range []domain.IssueStatus{domain.IssueOpen, domain.IssueInProgress, domain.IssueBlocked} {
			decisions := []domain.Issue{decision(domain.IssueDone), decision(status)}
			if !errors.Is(rules.CheckPlanClear(planned, domain.PhaseDo, decisions), domain.ErrValidation) {
				t.Errorf("a %s decision should hold plan", status)
			}
		}
	})

	t.Run("lets plan go once every decision is closed", func(t *testing.T) {
		decisions := []domain.Issue{decision(domain.IssueDone), decision(domain.IssueCancelled)}
		if err := rules.CheckPlanClear(planned, domain.PhaseDo, decisions); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("ignores todos under the map", func(t *testing.T) {
		step := domain.Issue{Kind: domain.IssueTodo, Status: domain.IssueOpen}
		if err := rules.CheckPlanClear(planned, domain.PhaseDo, []domain.Issue{step}); err != nil {
			t.Fatalf("a todo is a step, not a decision: %v", err)
		}
	})

	t.Run("does not apply without a map", func(t *testing.T) {
		unmapped := planned
		unmapped.MapID = ""
		if err := rules.CheckPlanClear(unmapped, domain.PhaseDo, []domain.Issue{decision(domain.IssueOpen)}); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("does not apply past plan", func(t *testing.T) {
		doing := planned
		doing.Phase = domain.PhaseDo
		if err := rules.CheckPlanClear(doing, domain.PhaseCheck, []domain.Issue{decision(domain.IssueOpen)}); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}
