package domain

import "time"

type TicketLog struct {
	ID        TicketLogID
	ProjectID ProjectID
	TicketID  TicketID
	CycleID   CycleID
	Body      string
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TicketLogFilter struct {
	TicketID TicketID
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
	TodoID    TodoID
	Body      string
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TodoLogFilter struct {
	TodoID TodoID
	Search string
	Limit  int
	Offset int
}
