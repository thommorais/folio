package ports

import (
	"context"

	"folio/folio-core/domain"
)

type ProjectUseCase interface {
	ListProjects(ctx context.Context, actor Actor, includeArchived bool) ([]domain.Project, error)
	GetProject(ctx context.Context, actor Actor, ref string) (domain.Project, error)
	CreateProject(ctx context.Context, actor Actor, in CreateProjectInput) (domain.Project, error)
	UpdateProject(ctx context.Context, actor Actor, id domain.ProjectID, in UpdateProjectInput) (domain.Project, error)
	DeleteProject(ctx context.Context, actor Actor, id domain.ProjectID) error

	AddMember(ctx context.Context, actor Actor, id domain.ProjectID, email string, role domain.Role) (domain.Project, error)
	RemoveMember(ctx context.Context, actor Actor, id domain.ProjectID, user domain.UserID) (domain.Project, error)
	SetMemberRole(ctx context.Context, actor Actor, id domain.ProjectID, user domain.UserID, role domain.Role) (domain.Project, error)
}

type CreateProjectInput struct {
	DomainID domain.DomainID
	Slug     string
	Name     string
	Descr    string
}

type UpdateProjectInput struct {
	Name     *string
	Descr    *string
	Archived *bool
}

type PlanUseCase interface {
	ListPlans(ctx context.Context, actor Actor, project domain.ProjectID, statuses []domain.PlanStatus) ([]domain.Plan, error)
	GetPlan(ctx context.Context, actor Actor, id domain.PlanID) (domain.Plan, error)
	CreatePlan(ctx context.Context, actor Actor, in CreatePlanInput) (domain.Plan, error)
	UpdatePlan(ctx context.Context, actor Actor, id domain.PlanID, in UpdatePlanInput) (domain.Plan, error)
	DeletePlan(ctx context.Context, actor Actor, id domain.PlanID) error
}

type CreatePlanInput struct {
	ProjectID domain.ProjectID
	IssueID   domain.IssueID
	Title     string
	Goal      string
	Status    domain.PlanStatus
	Tags      []string
	Todos     []CreateIssueInput
}

type UpdatePlanInput struct {
	IssueID *domain.IssueID
	Title   *string
	Goal    *string
	Status  *domain.PlanStatus
	Tags    *[]string
}

type IssueUseCase interface {
	ListIssues(ctx context.Context, actor Actor, project domain.ProjectID, f domain.IssueFilter) ([]domain.Issue, error)
	GetIssue(ctx context.Context, actor Actor, id domain.IssueID) (domain.Issue, error)
	GetIssueBySlug(ctx context.Context, actor Actor, project domain.ProjectID, slug string) (domain.Issue, error)
	CreateIssue(ctx context.Context, actor Actor, in CreateIssueInput) (domain.Issue, error)
	CreateIssues(ctx context.Context, actor Actor, project domain.ProjectID, in []CreateIssueInput) (BatchResult[domain.Issue], error)
	UpdateIssue(ctx context.Context, actor Actor, id domain.IssueID, in UpdateIssueInput) (domain.Issue, error)
	SetIssueStatus(ctx context.Context, actor Actor, id domain.IssueID, status domain.IssueStatus) (domain.Issue, error)
	DeleteIssue(ctx context.Context, actor Actor, id domain.IssueID) error

	LinkIssues(ctx context.Context, actor Actor, from, to domain.IssueID, kind domain.LinkKind) error
	UnlinkIssues(ctx context.Context, actor Actor, from, to domain.IssueID, kind domain.LinkKind) error

	Frontier(ctx context.Context, actor Actor, mapID domain.IssueID) ([]domain.Issue, error)
	GetIssueBrief(ctx context.Context, actor Actor, id domain.IssueID, in BriefOptions) (domain.IssueBrief, error)
	GetIssueBriefBySlug(ctx context.Context, actor Actor, project domain.ProjectID, slug string, in BriefOptions) (domain.IssueBrief, error)
}

type CreateIssueInput struct {
	ProjectID   domain.ProjectID
	Kind        domain.IssueKind
	ParentID    domain.IssueID
	PlanID      domain.PlanID
	Slug        string
	Title       string
	Body        string
	Status      domain.IssueStatus
	Priority    domain.Priority
	Size        domain.Size
	Assignee    domain.UserID
	Tags        []string
	DueDate     *string
	DependsOn   []domain.IssueID
	Wayfinder   domain.WayfinderType
	ExternalRef string
	Resolution  string
}

type UpdateIssueInput struct {
	Kind            *domain.IssueKind
	ParentID        *domain.IssueID
	PlanID          *domain.PlanID
	Slug            *string
	Title           *string
	Body            *string
	Status          *domain.IssueStatus
	Priority        *domain.Priority
	Size            *domain.Size
	Assignee        *domain.UserID
	Tags            *[]string
	Position        *int
	DueDate         *string
	DependsOn       *[]domain.IssueID
	Wayfinder       *domain.WayfinderType
	Resolution      *string
	ResolutionEntry *domain.EntryID
	ExternalRef     *string
}

type BriefOptions struct {
	RecentJournal int
}

type EntryUseCase interface {
	ListEntries(ctx context.Context, actor Actor, project domain.ProjectID, f domain.EntryFilter) ([]domain.Entry, error)
	GetEntry(ctx context.Context, actor Actor, id domain.EntryID) (domain.Entry, error)
	GetEntryBySlug(ctx context.Context, actor Actor, project domain.ProjectID, slug string) (domain.Entry, error)
	WriteEntry(ctx context.Context, actor Actor, in WriteEntryInput) (domain.Entry, error)
	UpdateEntry(ctx context.Context, actor Actor, id domain.EntryID, in UpdateEntryInput) (domain.Entry, error)
	AppendToEntry(ctx context.Context, actor Actor, id domain.EntryID, section string) (domain.Entry, error)
	DeleteEntry(ctx context.Context, actor Actor, id domain.EntryID) error
}

type WriteEntryInput struct {
	ProjectID   domain.ProjectID
	Kind        domain.EntryKind
	IssueID     domain.IssueID
	PlanID      domain.PlanID
	CycleID     domain.CycleID
	Slug        string
	Title       string
	Body        string
	Branch      string
	PR          string
	ExternalRef string
	Tags        []string
	Meta        map[string]any
}

type UpdateEntryInput struct {
	IssueID     *domain.IssueID
	PlanID      *domain.PlanID
	Slug        *string
	Title       *string
	Body        *string
	Branch      *string
	PR          *string
	ExternalRef *string
	Tags        *[]string
	Meta        *map[string]any
}

type CycleUseCase interface {
	ListCycles(ctx context.Context, actor Actor, issue domain.IssueID) ([]domain.Cycle, error)
	OpenCycle(ctx context.Context, actor Actor, issue domain.IssueID) (domain.Cycle, error)
	AdvancePhase(ctx context.Context, actor Actor, id domain.CycleID, phase domain.Phase) (domain.Cycle, error)
	ResolveCycle(ctx context.Context, actor Actor, id domain.CycleID, resolution string) (domain.Cycle, error)
	SetCycleMap(ctx context.Context, actor Actor, id domain.CycleID, mapID domain.IssueID) (domain.Cycle, error)
}

// BatchResult reports a partial success: what was written, and why the rest
// was not.
type BatchResult[T any] struct {
	Created []T
	Errors  []BatchError
}

type BatchError struct {
	Index  int
	Title  string
	Reason string
}

type SearchUseCase interface {
	Search(ctx context.Context, actor Actor, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error)
	// SearchAll spans every project the actor can read.
	SearchAll(ctx context.Context, actor Actor, q domain.SearchQuery) ([]domain.SearchHit, error)
}

type ShareUseCase interface {
	ShareIssue(ctx context.Context, actor Actor, issue domain.IssueID, label string) (domain.Share, error)
	SharePlan(ctx context.Context, actor Actor, plan domain.PlanID, label string) (domain.Share, error)
	ListShares(ctx context.Context, actor Actor, project domain.ProjectID) ([]domain.Share, error)
	RevokeShare(ctx context.Context, actor Actor, id domain.ShareID) error
	OpenShare(ctx context.Context, token string) (domain.SharedItem, error)
}

// KnowledgeUseCase is deliberately the one surface with no project in its
// signatures: a note is readable and writable by any authenticated user, and
// the optional project on it is a label rather than a permission.
type KnowledgeUseCase interface {
	ListKnowledge(ctx context.Context, actor Actor, f domain.KnowledgeFilter) ([]domain.Knowledge, error)
	GetKnowledge(ctx context.Context, actor Actor, id domain.KnowledgeID) (domain.Knowledge, error)
	GetKnowledgeBySlug(ctx context.Context, actor Actor, slug string) (domain.Knowledge, error)
	WriteKnowledge(ctx context.Context, actor Actor, in WriteKnowledgeInput) (domain.Knowledge, error)
	UpdateKnowledge(ctx context.Context, actor Actor, id domain.KnowledgeID, in UpdateKnowledgeInput) (domain.Knowledge, error)
	DeleteKnowledge(ctx context.Context, actor Actor, id domain.KnowledgeID) error
}

type WriteKnowledgeInput struct {
	// ProjectID is optional. When set it is checked against the actor's
	// membership, so naming a project is not a way to learn one exists.
	ProjectID domain.ProjectID
	// Slug is derived from the title when empty, and a clash is resolved with
	// a suffix; an explicitly requested one that is taken is a conflict.
	Slug  string
	Title string
	Body  string
	Tags  []string
}

// UpdateKnowledgeInput leaves a nil field alone, so a caller can change one
// field without reading the record first.
type UpdateKnowledgeInput struct {
	ProjectID *domain.ProjectID
	Slug      *string
	Title     *string
	Body      *string
	Tags      *[]string
}
