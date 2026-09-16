package domain

import "time"

// Doc is durable project knowledge: a spec, a convention, an architecture
// note. Unlike a JournalEntry it is meant to be edited and kept current.
type Doc struct {
	ID        DocID
	ProjectID ProjectID
	// TicketID is optional: an empty value means the doc sits directly under
	// the project rather than under one of its tickets.
	TicketID  TicketID
	Slug      string
	Title     string
	Body      string
	Tags      []string
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DocFilter narrows a doc listing. Zero values mean "no restriction".
type DocFilter struct {
	TicketID TicketID
	Tags     []string
	Search   string
	Limit    int
	Offset   int
}

// SearchKind identifies which collection a search hit came from.
type SearchKind string

const (
	SearchKindJournal SearchKind = "journal"
	SearchKindDoc     SearchKind = "doc"
	SearchKindTodo    SearchKind = "todo"
	SearchKindPlan    SearchKind = "plan"
	SearchKindTicket  SearchKind = "ticket"
	SearchKindWorkLog SearchKind = "worklog"
	SearchKindCycle   SearchKind = "resolution"
)

// SearchHit is one result of a cross-collection search, flattened so a client
// can render a mixed list without knowing each source type.
type SearchHit struct {
	Kind      SearchKind
	ID        string
	ProjectID ProjectID
	// ProjectSlug names the hit's project, which a global search needs and a
	// project-scoped one already knows.
	ProjectSlug string
	Title       string
	Snippet     string
	Tags        []string
	CreatedAt   time.Time
}

// SearchQuery asks for matches across the kinds listed; empty Kinds means all.
type SearchQuery struct {
	Text   string
	Kinds  []SearchKind
	Tags   []string
	Limit  int
	Offset int
}
