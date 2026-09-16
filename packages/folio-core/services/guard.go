// Package services holds the use case implementations: the orchestration
// between the pure rules in domain/rules and the driven ports. Every exported
// method takes a ports.Actor and authorises the call itself, so a new driving
// adapter cannot accidentally skip a permission check.
package services

import (
	"context"
	"errors"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// ProjectGuard resolves project permissions from membership.
type ProjectGuard struct {
	projects ports.ProjectRepository
}

func NewProjectGuard(projects ports.ProjectRepository) *ProjectGuard {
	return &ProjectGuard{projects: projects}
}

var _ ports.Guard = (*ProjectGuard)(nil)

// ensure loads the project and checks the actor's role against want. A
// non-member gets ErrNotFound rather than ErrForbidden: telling a stranger
// that a project exists is itself a leak. A member lacking the level gets
// ErrForbidden, since they already know it exists.
func (g *ProjectGuard) ensure(ctx context.Context, actor ports.Actor, id domain.ProjectID, allow func(domain.Role) bool) (domain.Project, error) {
	project, err := g.projects.GetByID(ctx, id)
	if err != nil {
		return domain.Project{}, err
	}
	if actor.Superuser {
		return project, nil
	}
	role, member := project.RoleOf(actor.UserID)
	if !member {
		return domain.Project{}, domain.ErrNotFound
	}
	if !allow(role) {
		return domain.Project{}, domain.ErrForbidden
	}
	return project, nil
}

func (g *ProjectGuard) EnsureRead(ctx context.Context, actor ports.Actor, id domain.ProjectID) (domain.Project, error) {
	return g.ensure(ctx, actor, id, func(domain.Role) bool { return true })
}

func (g *ProjectGuard) EnsureWrite(ctx context.Context, actor ports.Actor, id domain.ProjectID) (domain.Project, error) {
	return g.ensure(ctx, actor, id, domain.Role.CanWrite)
}

func (g *ProjectGuard) EnsureAdmin(ctx context.Context, actor ports.Actor, id domain.ProjectID) (domain.Project, error) {
	return g.ensure(ctx, actor, id, domain.Role.CanAdmin)
}

// notFound reports whether err means the record is absent.
func notFound(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}
