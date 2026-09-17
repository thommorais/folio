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
	IssueID  domain.IssueID
	Title     string
	Goal      string
	Status    domain.PlanStatus
	Tags      []string
	Todos     []CreateIssueInput
}

type UpdatePlanInput struct {
	IssueID *domain.IssueID
	Title    *string
	Goal     *string
	Status   *domain.PlanStatus
	Tags     *[]string
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
}

type UpdateIssueInput struct {
	Kind        *domain.IssueKind
	ParentID    *domain.IssueID
	PlanID      *domain.PlanID
	Slug        *string
	Title       *string
	Body        *string
	Status      *domain.IssueStatus
	Priority    *domain.Priority
	Size        *domain.Size
	Assignee    *domain.UserID
	Tags        *[]string
	Position    *int
	DueDate     *string
	DependsOn   *[]domain.IssueID
	Wayfinder   *domain.WayfinderType
	ExternalRef *string
}

type BriefOptions struct {
	RecentJournal int
}

type CycleUseCase interface {
	ListCycles(ctx context.Context, actor Actor, issue domain.IssueID) ([]domain.Cycle, error)
	OpenCycle(ctx context.Context, actor Actor, issue domain.IssueID) (domain.Cycle, error)
	AdvancePhase(ctx context.Context, actor Actor, id domain.CycleID, phase domain.Phase) (domain.Cycle, error)
	ResolveCycle(ctx context.Context, actor Actor, id domain.CycleID, resolution string) (domain.Cycle, error)
}

type WorkLogUseCase interface {
	ListTicketLogs(ctx context.Context, actor Actor, issue domain.IssueID, f domain.TicketLogFilter) ([]domain.TicketLog, error)
	WriteTicketLog(ctx context.Context, actor Actor, issue domain.IssueID, body string) (domain.TicketLog, error)
	DeleteTicketLog(ctx context.Context, actor Actor, id domain.TicketLogID) error

	ListPlanLogs(ctx context.Context, actor Actor, plan domain.PlanID, f domain.PlanLogFilter) ([]domain.PlanLog, error)
	WritePlanLog(ctx context.Context, actor Actor, plan domain.PlanID, body string) (domain.PlanLog, error)
	DeletePlanLog(ctx context.Context, actor Actor, id domain.PlanLogID) error

	ListTodoLogs(ctx context.Context, actor Actor, issue domain.IssueID, f domain.TodoLogFilter) ([]domain.TodoLog, error)
	WriteTodoLog(ctx context.Context, actor Actor, issue domain.IssueID, body string) (domain.TodoLog, error)
	DeleteTodoLog(ctx context.Context, actor Actor, id domain.TodoLogID) error
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

type JournalUseCase interface {
	ListJournal(ctx context.Context, actor Actor, project domain.ProjectID, f domain.JournalFilter) ([]domain.JournalEntry, error)
	GetJournalEntry(ctx context.Context, actor Actor, id domain.JournalID) (domain.JournalEntry, error)
	GetJournalEntryBySlug(ctx context.Context, actor Actor, project domain.ProjectID, slug string) (domain.JournalEntry, error)
	WriteJournalEntry(ctx context.Context, actor Actor, in WriteJournalInput) (domain.JournalEntry, error)
	UpdateJournalEntry(ctx context.Context, actor Actor, id domain.JournalID, in UpdateJournalInput) (domain.JournalEntry, error)
	AppendToJournalEntry(ctx context.Context, actor Actor, id domain.JournalID, section string) (domain.JournalEntry, error)
	DeleteJournalEntry(ctx context.Context, actor Actor, id domain.JournalID) error
}

type WriteJournalInput struct {
	ProjectID   domain.ProjectID
	IssueID     domain.IssueID
	PlanID      domain.PlanID
	Slug        string
	Title       string
	Body        string
	Branch      string
	PR          string
	ExternalRef string
	Tags        []string
	Meta        map[string]any
}

type UpdateJournalInput struct {
	IssueID     *domain.IssueID
	PlanID      *domain.PlanID
	Title       *string
	Body        *string
	Branch      *string
	PR          *string
	ExternalRef *string
	Tags        *[]string
	Meta        *map[string]any
}

type DocUseCase interface {
	ListDocs(ctx context.Context, actor Actor, project domain.ProjectID, f domain.DocFilter) ([]domain.Doc, error)
	GetDoc(ctx context.Context, actor Actor, id domain.DocID) (domain.Doc, error)
	GetDocBySlug(ctx context.Context, actor Actor, project domain.ProjectID, slug string) (domain.Doc, error)
	CreateDoc(ctx context.Context, actor Actor, in CreateDocInput) (domain.Doc, error)
	UpdateDoc(ctx context.Context, actor Actor, id domain.DocID, in UpdateDocInput) (domain.Doc, error)
	DeleteDoc(ctx context.Context, actor Actor, id domain.DocID) error
}

type CreateDocInput struct {
	ProjectID domain.ProjectID
	IssueID  domain.IssueID
	Slug      string
	Title     string
	Body      string
	Tags      []string
}

type UpdateDocInput struct {
	IssueID *domain.IssueID
	Slug     *string
	Title    *string
	Body     *string
	Tags     *[]string
}

type SearchUseCase interface {
	Search(ctx context.Context, actor Actor, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error)
}
