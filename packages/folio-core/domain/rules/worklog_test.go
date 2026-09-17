package rules_test

import (
	"errors"
	"strings"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func TestValidateTicketLog(t *testing.T) {
	valid := domain.TicketLog{ProjectID: "p1", IssueID: "tk1", Body: "Mapbox rejects feature-state in a filter."}

	t.Run("accepts a valid entry", func(t *testing.T) {
		if err := rules.ValidateTicketLog(valid); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("requires a ticket", func(t *testing.T) {
		bad := valid
		bad.IssueID = ""
		if !errors.Is(rules.ValidateTicketLog(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("requires a body", func(t *testing.T) {
		bad := valid
		bad.Body = "   "
		if !errors.Is(rules.ValidateTicketLog(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("rejects a body over the limit", func(t *testing.T) {
		bad := valid
		bad.Body = strings.Repeat("x", rules.BodyMaxLen+1)
		if !errors.Is(rules.ValidateTicketLog(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("accepts an entry with no cycle", func(t *testing.T) {
		ok := valid
		ok.CycleID = ""
		if err := rules.ValidateTicketLog(ok); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}

func TestValidatePlanLog(t *testing.T) {
	valid := domain.PlanLog{ProjectID: "p1", PlanID: "pl1", Body: "Benchmarks pending."}

	t.Run("accepts a valid entry", func(t *testing.T) {
		if err := rules.ValidatePlanLog(valid); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("requires a plan", func(t *testing.T) {
		bad := valid
		bad.PlanID = ""
		if !errors.Is(rules.ValidatePlanLog(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("requires a body", func(t *testing.T) {
		bad := valid
		bad.Body = ""
		if !errors.Is(rules.ValidatePlanLog(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})
}

func TestValidateTodoLog(t *testing.T) {
	valid := domain.TodoLog{ProjectID: "p1", IssueID: "t1", Body: "Blocked on the upstream fix."}

	t.Run("accepts a valid entry", func(t *testing.T) {
		if err := rules.ValidateTodoLog(valid); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("requires a todo", func(t *testing.T) {
		bad := valid
		bad.IssueID = ""
		if !errors.Is(rules.ValidateTodoLog(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("requires a body", func(t *testing.T) {
		bad := valid
		bad.Body = ""
		if !errors.Is(rules.ValidateTodoLog(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})
}
