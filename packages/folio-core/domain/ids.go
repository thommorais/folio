package domain

// Named types (not aliases) for every identifier: passing a ProjectID where a
// TodoID is expected is a compile error, which catches wiring mistakes that
// string-typed IDs would let through silently.
type (
	ClientID  string
	DomainID  string
	ProjectID string
	IssueID   string
	EntryID   string
	TagID     string
	PlanID    string
	TicketID  string
	TodoID    string
	CycleID   string
	ShareID   string

	TicketLogID string
	PlanLogID   string
	TodoLogID   string
	JournalID   string
	DocID       string
	UserID      string
)
