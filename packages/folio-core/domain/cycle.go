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
	IssueID    IssueID
	Ordinal    int
	Phase      Phase
	Resolution string
	MapID      IssueID
	CreatedBy  UserID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ClosedAt   *time.Time
}

func (c Cycle) IsClosed() bool { return c.ClosedAt != nil }

type CycleFilter struct {
	IssueID IssueID
	Limit    int
	Offset   int
}
