// Package ports declares the contracts between the folio domain and the world
// around it. Driven ports (repositories, clock, logger) are implemented by
// adapters; driving ports (use cases) are implemented by services and called
// by adapters such as the REST API, an MCP server or a CLI.
package ports

import (
	"context"
	"time"

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

type EntryRepository interface {
	List(ctx context.Context, project domain.ProjectID, f domain.EntryFilter) ([]domain.Entry, error)
	GetByID(ctx context.Context, id domain.EntryID) (domain.Entry, error)
	GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.Entry, error)
	Create(ctx context.Context, e domain.Entry) (domain.Entry, error)
	Update(ctx context.Context, e domain.Entry) (domain.Entry, error)
	Delete(ctx context.Context, id domain.EntryID) error
}

type TagTarget string

const (
	TagIssue TagTarget = "issue"
	TagEntry TagTarget = "entry"
)

type TagRepository interface {
	ListByDomain(ctx context.Context, domainID domain.DomainID) ([]domain.Tag, error)
	Ensure(ctx context.Context, domainID domain.DomainID, names []string) ([]domain.Tag, error)
	TagsOf(ctx context.Context, target TagTarget, ids []string) (map[string][]string, error)
	SetTags(ctx context.Context, target TagTarget, id string, domainID domain.DomainID, names []string) error
	IDsWithTags(ctx context.Context, target TagTarget, domainID domain.DomainID, names []string) ([]string, error)
}

type CycleRepository interface {
	ListByIssue(ctx context.Context, issue domain.IssueID) ([]domain.Cycle, error)
	GetByID(ctx context.Context, id domain.CycleID) (domain.Cycle, error)
	Create(ctx context.Context, c domain.Cycle) (domain.Cycle, error)
	Update(ctx context.Context, c domain.Cycle) (domain.Cycle, error)
	Delete(ctx context.Context, id domain.CycleID) error
}

type SearchRepository interface {
	Search(ctx context.Context, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error)
	// SearchAcross ranks one result set over several projects at once.
	// Running Search per project and merging would compare bm25 scores from
	// separate queries, which are not on a common scale.
	SearchAcross(ctx context.Context, projects []domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error)
}

type ShareRepository interface {
	List(ctx context.Context, project domain.ProjectID) ([]domain.Share, error)
	GetByID(ctx context.Context, id domain.ShareID) (domain.Share, error)
	GetByToken(ctx context.Context, token string) (domain.Share, error)
	Create(ctx context.Context, s domain.Share) (domain.Share, error)
	Delete(ctx context.Context, id domain.ShareID) error
	Touch(ctx context.Context, id domain.ShareID, at time.Time) error
}
