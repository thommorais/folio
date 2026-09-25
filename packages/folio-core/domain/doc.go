package domain

import "time"

// Doc is durable project knowledge: a spec, a convention, an architecture
// note. Unlike a JournalEntry it is meant to be edited and kept current.
type Doc struct {
	ID        DocID
	ProjectID ProjectID
	// IssueID is optional: an empty value means the doc sits directly under
	// the project rather than under one of its issues.
	IssueID   IssueID
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
	IssueID IssueID
	Tags    []string
	Search  string
	Limit   int
	Offset  int
}

// SearchKind identifies which collection a search hit came from.
type SearchKind string

const (
	SearchKindJournal  SearchKind = "journal"
	SearchKindDoc      SearchKind = "doc"
	SearchKindTodo     SearchKind = "todo"
	SearchKindPlan     SearchKind = "plan"
	SearchKindTicket   SearchKind = "ticket"
	SearchKindWorkLog  SearchKind = "worklog"
	SearchKindCycle    SearchKind = "resolution"
	SearchKindDecision SearchKind = "decision"
	// SearchKindKnowledge is the one kind not fenced by a project: a note
	// answers any caller's search.
	SearchKindKnowledge SearchKind = "knowledge"
)

// SearchHit is one result of a cross-collection search, flattened so a client
// can render a mixed list without knowing each source type.
type SearchHit struct {
	Kind      SearchKind
	ID        string
	ProjectID ProjectID
	// ProjectSlug names the hit's project, which a global search needs and a
	// project-scoped one already knows. ClientSlug and DomainSlug complete the
	// path a web client routes by; all three are empty for a hit that belongs
	// to no project, which only knowledge can be.
	ProjectSlug string
	ClientSlug  string
	DomainSlug  string
	// Slug addresses the record in its own routes. Empty for kinds that have
	// none, such as todos and plans.
	Slug      string
	Title     string
	Snippet   string
	Tags      []string
	CreatedAt time.Time
}

// SearchQuery asks for matches across the kinds listed; empty Kinds means all.
type SearchQuery struct {
	Text   string
	Kinds  []SearchKind
	Tags   []string
	Limit  int
	Offset int
}
