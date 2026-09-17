// Package ports declares the contracts between the folio domain and the world
// around it. Driven ports (repositories, clock, logger) are implemented by
// adapters; driving ports (use cases) are implemented by services and called
// by adapters such as the REST API, an MCP server or a CLI.
package ports

import (
	"context"

	"folio/folio-core/domain"
)

// ProjectRepository stores projects and their membership. Implementations are
// responsible for tenancy: every method takes the acting user, and must not
// return projects that user cannot see.
type ProjectRepository interface {
	List(ctx context.Context, actor domain.UserID, includeArchived bool) ([]domain.Project, error)
	GetByID(ctx context.Context, id domain.ProjectID) (domain.Project, error)
	GetBySlug(ctx context.Context, slug string) (domain.Project, error)
	Create(ctx context.Context, p domain.Project) (domain.Project, error)
	Update(ctx context.Context, p domain.Project) (domain.Project, error)
	Delete(ctx context.Context, id domain.ProjectID) error

	AddMember(ctx context.Context, id domain.ProjectID, user domain.UserID, role domain.Role) error
	RemoveMember(ctx context.Context, id domain.ProjectID, user domain.UserID) error
	SetMemberRole(ctx context.Context, id domain.ProjectID, user domain.UserID, role domain.Role) error
	// FindUserByEmail resolves an invite target. Returns domain.ErrNotFound
	// when no account exists for the address.
	FindUserByEmail(ctx context.Context, email string) (domain.UserID, error)
}

type PlanRepository interface {
	List(ctx context.Context, project domain.ProjectID, statuses []domain.PlanStatus) ([]domain.Plan, error)
	// ListByTicket returns every plan under a ticket, so deleting the ticket
	// can detach them.
	ListByIssue(ctx context.Context, issue domain.IssueID) ([]domain.Plan, error)
	GetByID(ctx context.Context, id domain.PlanID) (domain.Plan, error)
	Create(ctx context.Context, p domain.Plan) (domain.Plan, error)
	Update(ctx context.Context, p domain.Plan) (domain.Plan, error)
	Delete(ctx context.Context, id domain.PlanID) error
}

type IssueRepository interface {
	List(ctx context.Context, project domain.ProjectID, f domain.IssueFilter) ([]domain.Issue, error)
	ListByParent(ctx context.Context, parent domain.IssueID) ([]domain.Issue, error)
	ListByPlan(ctx context.Context, plan domain.PlanID) ([]domain.Issue, error)
	GetByID(ctx context.Context, id domain.IssueID) (domain.Issue, error)
	GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.Issue, error)
	Create(ctx context.Context, i domain.Issue) (domain.Issue, error)
	Update(ctx context.Context, i domain.Issue) (domain.Issue, error)
	Delete(ctx context.Context, id domain.IssueID) error

	Links(ctx context.Context, id domain.IssueID) ([]domain.IssueLink, error)
	LinksOfProject(ctx context.Context, project domain.ProjectID) ([]domain.IssueLink, error)
	Link(ctx context.Context, from, to domain.IssueID, kind domain.LinkKind) error
	Unlink(ctx context.Context, from, to domain.IssueID, kind domain.LinkKind) error
	SetLinks(ctx context.Context, from domain.IssueID, kind domain.LinkKind, to []domain.IssueID) error
}

// JournalRepository stores the work log. Entries are editable: a log documents
// the state of a piece of work, and that state changes as the work proceeds.
type JournalRepository interface {
	List(ctx context.Context, project domain.ProjectID, f domain.JournalFilter) ([]domain.JournalEntry, error)
	GetByID(ctx context.Context, id domain.JournalID) (domain.JournalEntry, error)
	GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.JournalEntry, error)
	Create(ctx context.Context, e domain.JournalEntry) (domain.JournalEntry, error)
	Update(ctx context.Context, e domain.JournalEntry) (domain.JournalEntry, error)
	Delete(ctx context.Context, id domain.JournalID) error
}

type CycleRepository interface {
	ListByIssue(ctx context.Context, issue domain.IssueID) ([]domain.Cycle, error)
	GetByID(ctx context.Context, id domain.CycleID) (domain.Cycle, error)
	Create(ctx context.Context, c domain.Cycle) (domain.Cycle, error)
	Update(ctx context.Context, c domain.Cycle) (domain.Cycle, error)
	Delete(ctx context.Context, id domain.CycleID) error
}

type TicketLogRepository interface {
	List(ctx context.Context, project domain.ProjectID, f domain.TicketLogFilter) ([]domain.TicketLog, error)
	GetByID(ctx context.Context, id domain.TicketLogID) (domain.TicketLog, error)
	Create(ctx context.Context, l domain.TicketLog) (domain.TicketLog, error)
	Update(ctx context.Context, l domain.TicketLog) (domain.TicketLog, error)
	Delete(ctx context.Context, id domain.TicketLogID) error
}

type PlanLogRepository interface {
	List(ctx context.Context, project domain.ProjectID, f domain.PlanLogFilter) ([]domain.PlanLog, error)
	GetByID(ctx context.Context, id domain.PlanLogID) (domain.PlanLog, error)
	Create(ctx context.Context, l domain.PlanLog) (domain.PlanLog, error)
	Update(ctx context.Context, l domain.PlanLog) (domain.PlanLog, error)
	Delete(ctx context.Context, id domain.PlanLogID) error
}

type TodoLogRepository interface {
	List(ctx context.Context, project domain.ProjectID, f domain.TodoLogFilter) ([]domain.TodoLog, error)
	GetByID(ctx context.Context, id domain.TodoLogID) (domain.TodoLog, error)
	Create(ctx context.Context, l domain.TodoLog) (domain.TodoLog, error)
	Update(ctx context.Context, l domain.TodoLog) (domain.TodoLog, error)
	Delete(ctx context.Context, id domain.TodoLogID) error
}

type DocRepository interface {
	List(ctx context.Context, project domain.ProjectID, f domain.DocFilter) ([]domain.Doc, error)
	GetByID(ctx context.Context, id domain.DocID) (domain.Doc, error)
	GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.Doc, error)
	Create(ctx context.Context, d domain.Doc) (domain.Doc, error)
	Update(ctx context.Context, d domain.Doc) (domain.Doc, error)
	Delete(ctx context.Context, id domain.DocID) error
}

// SearchRepository runs the cross-collection text search. It is a separate
// port because the efficient implementation is one query per storage engine,
// not a fan-out over the other repositories.
type SearchRepository interface {
	Search(ctx context.Context, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error)
}
