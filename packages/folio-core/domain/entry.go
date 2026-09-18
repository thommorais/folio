package domain

import "time"

type EntryKind string

const (
	EntryJournal EntryKind = "journal"
	EntryDoc     EntryKind = "doc"
	EntryLog     EntryKind = "log"
)

func (k EntryKind) Addressable() bool {
	return k == EntryJournal || k == EntryDoc
}

type Entry struct {
	ID        EntryID
	Kind      EntryKind
	ProjectID ProjectID
	IssueID   IssueID
	PlanID    PlanID
	CycleID   CycleID
	Slug      string
	Title     string
	Body      string
	Branch      string
	PR          string
	ExternalRef string
	Tags        []string
	Meta        map[string]any
	CreatedBy   UserID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EntryFilter struct {
	Kind        EntryKind
	IssueID     IssueID
	PlanID      PlanID
	CycleID     CycleID
	Branch      string
	ExternalRef string
	Tags        []string
	Search      string
	Since       *time.Time
	Until       *time.Time
	Limit       int
	Offset      int
}

type Tag struct {
	ID       TagID
	DomainID DomainID
	Slug     string
	Name     string
}
