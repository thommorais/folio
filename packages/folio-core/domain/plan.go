package domain

import "time"

// PlanStatus tracks a plan through its life. A plan is the agent's stated
// intent for a piece of work; the todos under it are the steps.
type PlanStatus string

const (
	PlanDraft     PlanStatus = "draft"
	PlanActive    PlanStatus = "active"
	PlanDone      PlanStatus = "done"
	PlanAbandoned PlanStatus = "abandoned"
)

// IsTerminal reports whether the plan has reached an end state.
func (s PlanStatus) IsTerminal() bool {
	return s == PlanDone || s == PlanAbandoned
}

type Plan struct {
	ID        PlanID
	ProjectID ProjectID
	// TicketID is optional: an empty value means the plan sits directly under
	// the project rather than under one of its tickets.
	TicketID  TicketID
	Title     string
	Goal      string
	Status    PlanStatus
	Tags      []string
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time

	// Progress is derived from the plan's todos on read, never persisted.
	Progress Progress
}

// Progress counts a plan's todos by completion, for clients that render a bar
// without fetching the whole list.
type Progress struct {
	Total int
	Done  int
}

// Percent returns completion 0-100; an empty plan reports 0.
func (p Progress) Percent() int {
	if p.Total == 0 {
		return 0
	}
	return p.Done * 100 / p.Total
}
