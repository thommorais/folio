package domain

import "time"

type TicketStatus string

const (
	TicketOpen       TicketStatus = "open"
	TicketInProgress TicketStatus = "in_progress"
	TicketBlocked    TicketStatus = "blocked"
	TicketClosed     TicketStatus = "closed"
	TicketCancelled  TicketStatus = "cancelled"
)

func (s TicketStatus) IsTerminal() bool {
	return s == TicketClosed || s == TicketCancelled
}

type WayfinderType string

const (
	WayfinderMap       WayfinderType = "map"
	WayfinderResearch  WayfinderType = "research"
	WayfinderPrototype WayfinderType = "prototype"
	WayfinderGrilling  WayfinderType = "grilling"
	WayfinderTask      WayfinderType = "task"
)

type Ticket struct {
	ID          TicketID
	ProjectID   ProjectID
	ParentID    TicketID
	Slug        string
	Title       string
	Body        string
	Status      TicketStatus
	Priority    Priority
	Assignee    UserID
	Tags        []string
	ExternalRef string
	DependsOn   []TicketID
	Wayfinder   WayfinderType
	CreatedBy   UserID
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Progress Progress
	Blocked  bool
	Cycle    int
	Phase    Phase
}

type TicketFilter struct {
	ParentID TicketID
	Status   []TicketStatus
	Priority Priority
	Assignee UserID
	Tags     []string
	Search   string
	Limit    int
	Offset   int
}

type TicketBrief struct {
	Ticket  Ticket
	Plans   []Plan
	Todos   []Todo
	Journal []JournalEntry
	Docs    []Doc
	Cycles  []Cycle
}
