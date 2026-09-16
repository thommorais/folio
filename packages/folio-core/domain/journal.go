package domain

import "time"

// JournalEntry documents a piece of work: what was built, how, where it stands
// and why it was done that way. It is the project's memory, written for
// whoever picks the work up next, including the agent itself on a later run.
//
// An entry is titled and editable rather than a timestamped event: the state
// of a piece of work changes as it progresses, and the record should follow
// it instead of accumulating corrections in a stream.
type JournalEntry struct {
	ID        JournalID
	ProjectID ProjectID
	// Slug addresses the entry within its project and survives a title edit.
	Slug string
	// PlanID and TodoID are optional back-references to the work the entry
	// documents.
	PlanID PlanID
	TodoID TodoID
	// TicketID is optional: an empty value means the entry sits directly
	// under the project rather than under one of its tickets.
	TicketID TicketID
	Title    string
	// Body is markdown: the problem, the approach, the decisions, whatever
	// the next person needs. Length is deliberately generous.
	Body string
	// Branch, PR and ExternalRef anchor the entry to the work as it happened.
	// An agent knows its branch from git, so these cost nothing to fill and
	// make the entry findable from a code reference later. ExternalRef names
	// the work in another tracker and is distinct from TicketID, which points
	// at a folio ticket.
	Branch      string
	PR          string
	ExternalRef string
	Tags        []string
	// Meta is arbitrary structured context an adapter wants to preserve.
	Meta      map[string]any
	CreatedBy UserID
	// CreatedAt and UpdatedAt are set from the server clock, never by the
	// caller: the date an entry carries has to be trustworthy for reading the
	// project back chronologically.
	CreatedAt time.Time
	UpdatedAt time.Time
}

// JournalFilter narrows a log query. Zero values mean "no restriction".
type JournalFilter struct {
	PlanID      PlanID
	TodoID      TodoID
	TicketID    TicketID
	Branch      string
	ExternalRef string
	Tags        []string
	// Search matches the title and the body.
	Search string
	Since  *time.Time
	Until  *time.Time
	Limit  int
	Offset int
}
