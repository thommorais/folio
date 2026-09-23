package services

import (
	"context"
	"fmt"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

// KnowledgeService is the one use case with no tenancy: knowledge is shared
// across the whole installation, so the only check is that the caller is
// signed in. A project named on a note is still checked against the actor's
// membership, because attaching to a project is a claim about that project.
type KnowledgeService struct {
	repo  ports.KnowledgeRepository
	guard ports.Guard
	clock ports.Clock
	ids   ports.IDGenerator
	log   ports.Logger
}

func NewKnowledgeService(repo ports.KnowledgeRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *KnowledgeService {
	return &KnowledgeService{repo: repo, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.KnowledgeUseCase = (*KnowledgeService)(nil)

// ensureSignedIn is the whole authorisation rule for knowledge. It is a named
// method rather than an inline check so that the absence of a membership test
// reads as deliberate at every call site.
func (s *KnowledgeService) ensureSignedIn(actor ports.Actor) error {
	if actor.UserID == "" && !actor.Superuser {
		return domain.ErrForbidden
	}
	return nil
}

func (s *KnowledgeService) ListKnowledge(ctx context.Context, actor ports.Actor, f domain.KnowledgeFilter) ([]domain.Knowledge, error) {
	if err := s.ensureSignedIn(actor); err != nil {
		return nil, err
	}
	f.Limit = clampLimit(f.Limit)
	items, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.Knowledge{}
	}
	return items, nil
}

func (s *KnowledgeService) GetKnowledge(ctx context.Context, actor ports.Actor, id domain.KnowledgeID) (domain.Knowledge, error) {
	if err := s.ensureSignedIn(actor); err != nil {
		return domain.Knowledge{}, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *KnowledgeService) GetKnowledgeBySlug(ctx context.Context, actor ports.Actor, slug string) (domain.Knowledge, error) {
	if err := s.ensureSignedIn(actor); err != nil {
		return domain.Knowledge{}, err
	}
	return s.repo.GetBySlug(ctx, strings.TrimSpace(slug))
}

func (s *KnowledgeService) WriteKnowledge(ctx context.Context, actor ports.Actor, in ports.WriteKnowledgeInput) (domain.Knowledge, error) {
	if err := s.ensureSignedIn(actor); err != nil {
		return domain.Knowledge{}, err
	}
	if err := s.checkProject(ctx, actor, in.ProjectID); err != nil {
		return domain.Knowledge{}, err
	}

	now := s.clock.Now()
	note := domain.Knowledge{
		ID:        domain.KnowledgeID(s.ids.NewID()),
		ProjectID: in.ProjectID,
		Title:     strings.TrimSpace(in.Title),
		Body:      in.Body,
		Tags:      in.Tags,
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	slug, err := s.freeSlug(ctx, in.Slug, note.Title)
	if err != nil {
		return domain.Knowledge{}, err
	}
	note.Slug = slug

	if err := rules.ValidateKnowledge(note); err != nil {
		return domain.Knowledge{}, err
	}
	return s.repo.Create(ctx, note)
}

func (s *KnowledgeService) UpdateKnowledge(ctx context.Context, actor ports.Actor, id domain.KnowledgeID, in ports.UpdateKnowledgeInput) (domain.Knowledge, error) {
	if err := s.ensureSignedIn(actor); err != nil {
		return domain.Knowledge{}, err
	}

	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Knowledge{}, err
	}

	if in.ProjectID != nil {
		if err := s.checkProject(ctx, actor, *in.ProjectID); err != nil {
			return domain.Knowledge{}, err
		}
		note.ProjectID = *in.ProjectID
	}
	if in.Title != nil {
		note.Title = strings.TrimSpace(*in.Title)
	}
	if in.Body != nil {
		note.Body = *in.Body
	}
	if in.Tags != nil {
		note.Tags = *in.Tags
	}
	if in.Slug != nil && strings.TrimSpace(*in.Slug) != note.Slug {
		slug, err := s.freeSlug(ctx, *in.Slug, note.Title)
		if err != nil {
			return domain.Knowledge{}, err
		}
		note.Slug = slug
	}

	// CreatedBy and CreatedAt are never touched: a shared note's author is
	// the only trace of where it came from, and an editor is not its author.
	note.UpdatedAt = s.clock.Now()

	if err := rules.ValidateKnowledge(note); err != nil {
		return domain.Knowledge{}, err
	}
	return s.repo.Update(ctx, note)
}

func (s *KnowledgeService) DeleteKnowledge(ctx context.Context, actor ports.Actor, id domain.KnowledgeID) error {
	if err := s.ensureSignedIn(actor); err != nil {
		return err
	}
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// checkProject refuses a project the actor cannot read. Without it, attaching
// a note to an arbitrary id would both dangle and confirm that the project
// exists, which is the leak the guard exists to prevent.
func (s *KnowledgeService) checkProject(ctx context.Context, actor ports.Actor, project domain.ProjectID) error {
	if project == "" {
		return nil
	}
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		if notFound(err) {
			return domain.Invalid("project", "does not exist")
		}
		return err
	}
	return nil
}

// freeSlug resolves the global slug namespace. An explicit request that is
// taken is a conflict the caller has to see; a derived one takes a suffix,
// because two notes titled the same is ordinary rather than an error.
func (s *KnowledgeService) freeSlug(ctx context.Context, requested, title string) (string, error) {
	if slug := strings.TrimSpace(requested); slug != "" {
		free, err := s.slugFree(ctx, slug)
		if err != nil {
			return "", err
		}
		if !free {
			return "", domain.ErrConflict
		}
		return slug, nil
	}

	base := rules.Slugify(title)
	if base == "" {
		base = "note"
	}
	if len(base) > 50 {
		base = strings.TrimRight(base[:50], "-")
	}

	candidate := base
	for n := 2; ; n++ {
		free, err := s.slugFree(ctx, candidate)
		if err != nil {
			return "", err
		}
		if free {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, n)
	}
}

func (s *KnowledgeService) slugFree(ctx context.Context, slug string) (bool, error) {
	_, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if notFound(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
