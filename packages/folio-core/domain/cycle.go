package domain

import "time"

type Phase string

const (
	PhasePlan  Phase = "plan"
	PhaseDo    Phase = "do"
	PhaseCheck Phase = "check"
	PhaseAct   Phase = "act"
)

type Cycle struct {
	ID         CycleID
	ProjectID  ProjectID
	TicketID   TicketID
	Ordinal    int
	Phase      Phase
	Resolution string
	CreatedBy  UserID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ClosedAt   *time.Time
}

func (c Cycle) IsClosed() bool { return c.ClosedAt != nil }

type CycleFilter struct {
	TicketID TicketID
	Limit    int
	Offset   int
}
