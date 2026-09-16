package domain

import "time"

type TodoStatus string

const (
	TodoPending    TodoStatus = "pending"
	TodoInProgress TodoStatus = "in_progress"
	TodoDone       TodoStatus = "done"
	TodoBlocked    TodoStatus = "blocked"
	TodoCancelled  TodoStatus = "cancelled"
)

// IsTerminal reports whether no further work is expected on the todo.
func (s TodoStatus) IsTerminal() bool {
	return s == TodoDone || s == TodoCancelled
}

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Todo struct {
	ID        TodoID
	ProjectID ProjectID
	// PlanID is optional: a todo can stand alone, outside any plan.
	PlanID PlanID
	// TicketID is optional: an empty value means the todo sits directly under
	// the project rather than under one of its tickets.
	TicketID TicketID
	Title    string
	Details  string
	Status   TodoStatus
	Priority Priority
	Tags     []string
	// Position orders todos within their plan; gaps are allowed.
	Position  int
	DependsOn []TodoID
	DueDate   *time.Time
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time

	// Blocked is derived from DependsOn against the full set on read. It is
	// never persisted, and is distinct from the explicit TodoBlocked status
	// which the agent sets by hand.
	Blocked bool
}

// TodoFilter narrows a todo listing. Zero values mean "no restriction".
type TodoFilter struct {
	PlanID   PlanID
	TicketID TicketID
	Status   []TodoStatus
	Priority Priority
	Tags     []string
	Search   string
	Limit    int
	Offset   int
}
