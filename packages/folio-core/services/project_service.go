package services

import (
	"context"
	"errors"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type ProjectService struct {
	repo  ports.ProjectRepository
	guard ports.Guard
	clock ports.Clock
	ids   ports.IDGenerator
	log   ports.Logger
}

func NewProjectService(repo ports.ProjectRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *ProjectService {
	return &ProjectService{repo: repo, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.ProjectUseCase = (*ProjectService)(nil)

func (s *ProjectService) ListProjects(ctx context.Context, actor ports.Actor, includeArchived bool) ([]domain.Project, error) {
	return s.repo.List(ctx, actor.UserID, includeArchived)
}

// GetProject accepts either an ID or a slug, so an agent that only knows the
// project by the name a human used can still address it.
func (s *ProjectService) GetProject(ctx context.Context, actor ports.Actor, ref string) (domain.Project, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return domain.Project{}, err
	}
	return s.guard.EnsureRead(ctx, actor, project.ID)
}

func (s *ProjectService) resolve(ctx context.Context, ref string) (domain.Project, error) {
	project, err := s.repo.GetByID(ctx, domain.ProjectID(ref))
	if err == nil {
		return project, nil
	}
	if !notFound(err) {
		return domain.Project{}, err
	}
	return s.repo.GetBySlug(ctx, ref)
}

func (s *ProjectService) CreateProject(ctx context.Context, actor ports.Actor, in ports.CreateProjectInput) (domain.Project, error) {
	now := s.clock.Now()
	project := domain.Project{
		ID:       domain.ProjectID(s.ids.NewID()),
		DomainID: in.DomainID,
		Slug:     strings.TrimSpace(in.Slug),
		Name:     strings.TrimSpace(in.Name),
		Descr:    in.Descr,
		// The creator becomes the first owner; without this a project would be
		// born unadministrable.
		Members:   []domain.Member{{UserID: actor.UserID, Role: domain.RoleOwner, Email: actor.Email}},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidateProject(project); err != nil {
		return domain.Project{}, err
	}
	created, err := s.repo.Create(ctx, project)
	if err != nil {
		return domain.Project{}, err
	}
	s.log.Info("project created", map[string]any{"project": string(created.ID), "slug": created.Slug})
	return created, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, actor ports.Actor, id domain.ProjectID, in ports.UpdateProjectInput) (domain.Project, error) {
	project, err := s.guard.EnsureWrite(ctx, actor, id)
	if err != nil {
		return domain.Project{}, err
	}
	if in.Name != nil {
		project.Name = strings.TrimSpace(*in.Name)
	}
	if in.Descr != nil {
		project.Descr = *in.Descr
	}
	if in.Archived != nil {
		project.Archived = *in.Archived
	}
	if err := rules.ValidateProject(project); err != nil {
		return domain.Project{}, err
	}
	project.UpdatedAt = s.clock.Now()
	return s.repo.Update(ctx, project)
}

func (s *ProjectService) DeleteProject(ctx context.Context, actor ports.Actor, id domain.ProjectID) error {
	if _, err := s.guard.EnsureAdmin(ctx, actor, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *ProjectService) AddMember(ctx context.Context, actor ports.Actor, id domain.ProjectID, email string, role domain.Role) (domain.Project, error) {
	if _, err := s.guard.EnsureAdmin(ctx, actor, id); err != nil {
		return domain.Project{}, err
	}
	if err := rules.ValidateRole(role); err != nil {
		return domain.Project{}, err
	}
	user, err := s.repo.FindUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if notFound(err) {
			return domain.Project{}, errors.Join(domain.ErrNotFound, errors.New("no account for "+email))
		}
		return domain.Project{}, err
	}
	if err := s.repo.AddMember(ctx, id, user, role); err != nil {
		return domain.Project{}, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) RemoveMember(ctx context.Context, actor ports.Actor, id domain.ProjectID, user domain.UserID) (domain.Project, error) {
	project, err := s.guard.EnsureAdmin(ctx, actor, id)
	if err != nil {
		return domain.Project{}, err
	}
	if err := rules.CanRemoveMember(project, user); err != nil {
		return domain.Project{}, err
	}
	if err := s.repo.RemoveMember(ctx, id, user); err != nil {
		return domain.Project{}, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) SetMemberRole(ctx context.Context, actor ports.Actor, id domain.ProjectID, user domain.UserID, role domain.Role) (domain.Project, error) {
	project, err := s.guard.EnsureAdmin(ctx, actor, id)
	if err != nil {
		return domain.Project{}, err
	}
	if err := rules.ValidateRole(role); err != nil {
		return domain.Project{}, err
	}
	// Demoting the last owner orphans the project as surely as removing them.
	if role != domain.RoleOwner {
		if err := rules.CanRemoveMember(project, user); err != nil {
			return domain.Project{}, err
		}
	}
	if err := s.repo.SetMemberRole(ctx, id, user, role); err != nil {
		return domain.Project{}, err
	}
	return s.repo.GetByID(ctx, id)
}
