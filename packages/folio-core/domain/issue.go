package domain

import "time"

type IssueKind string

const (
	IssueTicket IssueKind = "ticket"
	IssueTodo   IssueKind = "todo"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type WayfinderType string

const (
	WayfinderMap       WayfinderType = "map"
	WayfinderResearch  WayfinderType = "research"
	WayfinderPrototype WayfinderType = "prototype"
	WayfinderGrilling  WayfinderType = "grilling"
	WayfinderTask      WayfinderType = "task"
)

type IssueStatus string

const (
	IssueOpen       IssueStatus = "open"
	IssueInProgress IssueStatus = "in_progress"
	IssueBlocked    IssueStatus = "blocked"
	IssueDone       IssueStatus = "done"
	IssueCancelled  IssueStatus = "cancelled"
)

func (s IssueStatus) IsTerminal() bool {
	return s == IssueDone || s == IssueCancelled
}

type Size int

const (
	SizeXS Size = 1
	SizeS  Size = 2
	SizeM  Size = 3
	SizeL  Size = 5
	SizeXL Size = 8
)

var sizes = map[Size]bool{SizeXS: true, SizeS: true, SizeM: true, SizeL: true, SizeXL: true}

func (s Size) Valid() bool { return sizes[s] }

type LinkKind string

const (
	LinkBlocks  LinkKind = "blocks"
	LinkRelates LinkKind = "relates"
	LinkParent  LinkKind = "parent"
)

// From is the subject: for LinkParent it is the child, for LinkBlocks the blocker.
type IssueLink struct {
	ID       string
	DomainID DomainID
	From     IssueID
	To       IssueID
	Kind     LinkKind
}

type Issue struct {
	ID          IssueID
	Kind        IssueKind
	ProjectID   ProjectID
	PlanID      PlanID
	Slug        string
	Title       string
	Body        string
	Status      IssueStatus
	Priority    Priority
	Size        Size
	Assignee    UserID
	Tags        []string
	Position    int
	DueDate     *time.Time
	Wayfinder   WayfinderType
	ExternalRef string
	CreatedBy   UserID
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Derived on read, never persisted.
	ParentID  IssueID
	DependsOn []IssueID
	RelatedTo []IssueID
	Blocked   bool
	Progress  Progress
	Cycle     int
	Phase     Phase
}

// An unsized issue scores 0 and sorts last: an estimate never made should not
// win by default.
func (i Issue) Score() float64 {
	if !i.Size.Valid() {
		return 0
	}
	weight := map[Priority]float64{PriorityLow: 1, PriorityMedium: 2, PriorityHigh: 3}[i.Priority]
	return weight / float64(i.Size)
}

type IssueFilter struct {
	Kind     IssueKind
	ParentID IssueID
	PlanID   PlanID
	Status   []IssueStatus
	Priority Priority
	Assignee UserID
	Tags     []string
	Search   string
	Limit    int
	Offset   int
}

type IssueBrief struct {
	Issue    Issue
	Children []Issue
	Plans    []Plan
	Journal  []JournalEntry
	Docs     []Doc
	Cycles   []Cycle
}
