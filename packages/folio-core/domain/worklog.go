package domain

import "time"

type TicketLog struct {
	ID        TicketLogID
	ProjectID ProjectID
	IssueID   IssueID
	CycleID   CycleID
	Body      string
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TicketLogFilter struct {
	IssueID IssueID
	CycleID  CycleID
	Search   string
	Limit    int
	Offset   int
}

type PlanLog struct {
	ID        PlanLogID
	ProjectID ProjectID
	PlanID    PlanID
	Body      string
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PlanLogFilter struct {
	PlanID PlanID
	Search string
	Limit  int
	Offset int
}

type TodoLog struct {
	ID        TodoLogID
	ProjectID ProjectID
	IssueID   IssueID
	Body      string
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TodoLogFilter struct {
	IssueID IssueID
	Search string
	Limit  int
	Offset int
}
