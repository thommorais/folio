package services

import (
	"context"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type DocService struct {
	repo    ports.DocRepository
	tickets ports.TicketRepository
	guard   ports.Guard
	clock   ports.Clock
	ids     ports.IDGenerator
	log     ports.Logger
}

func NewDocService(repo ports.DocRepository, tickets ports.TicketRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *DocService {
	return &DocService{repo: repo, tickets: tickets, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.DocUseCase = (*DocService)(nil)

func (s *DocService) ListDocs(ctx context.Context, actor ports.Actor, project domain.ProjectID, f domain.DocFilter) ([]domain.Doc, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	f.Limit = clampLimit(f.Limit)
	docs, err := s.repo.List(ctx, project, f)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []domain.Doc{}
	}
	return docs, nil
}

func (s *DocService) GetDoc(ctx context.Context, actor ports.Actor, id domain.DocID) (domain.Doc, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Doc{}, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, doc.ProjectID); err != nil {
		return domain.Doc{}, err
	}
	return doc, nil
}

func (s *DocService) GetDocBySlug(ctx context.Context, actor ports.Actor, project domain.ProjectID, slug string) (domain.Doc, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return domain.Doc{}, err
	}
	return s.repo.GetBySlug(ctx, project, slug)
}

func (s *DocService) CreateDoc(ctx context.Context, actor ports.Actor, in ports.CreateDocInput) (domain.Doc, error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, in.ProjectID); err != nil {
		return domain.Doc{}, err
	}

	if err := ticketScope(ctx, s.tickets, in.TicketID, in.ProjectID); err != nil {
		return domain.Doc{}, err
	}

	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = rules.Slugify(in.Title)
	}
	now := s.clock.Now()
	doc := domain.Doc{
		ID:        domain.DocID(s.ids.NewID()),
		ProjectID: in.ProjectID,
		TicketID:  in.TicketID,
		Slug:      slug,
		Title:     strings.TrimSpace(in.Title),
		Body:      in.Body,
		Tags:      in.Tags,
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidateDoc(doc); err != nil {
		return domain.Doc{}, err
	}
	if err := s.slugFree(ctx, doc.ProjectID, doc.Slug, ""); err != nil {
		return domain.Doc{}, err
	}
	return s.repo.Create(ctx, doc)
}

// slugFree rejects a slug already used in the project. except is the ID of the
// doc being updated, so keeping its own slug is not a collision with itself.
func (s *DocService) slugFree(ctx context.Context, project domain.ProjectID, slug string, except domain.DocID) error {
	existing, err := s.repo.GetBySlug(ctx, project, slug)
	if err != nil {
		if notFound(err) {
			return nil
		}
		return err
	}
	if existing.ID == except {
		return nil
	}
	return domain.ErrConflict
}

func (s *DocService) UpdateDoc(ctx context.Context, actor ports.Actor, id domain.DocID, in ports.UpdateDocInput) (domain.Doc, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Doc{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, doc.ProjectID); err != nil {
		return domain.Doc{}, err
	}

	if in.TicketID != nil {
		if err := ticketScope(ctx, s.tickets, *in.TicketID, doc.ProjectID); err != nil {
			return domain.Doc{}, err
		}
		doc.TicketID = *in.TicketID
	}
	if in.Slug != nil {
		doc.Slug = strings.TrimSpace(*in.Slug)
	}
	if in.Title != nil {
		doc.Title = strings.TrimSpace(*in.Title)
	}
	if in.Body != nil {
		doc.Body = *in.Body
	}
	if in.Tags != nil {
		doc.Tags = *in.Tags
	}
	if err := rules.ValidateDoc(doc); err != nil {
		return domain.Doc{}, err
	}
	if in.Slug != nil {
		if err := s.slugFree(ctx, doc.ProjectID, doc.Slug, doc.ID); err != nil {
			return domain.Doc{}, err
		}
	}
	doc.UpdatedAt = s.clock.Now()
	return s.repo.Update(ctx, doc)
}

func (s *DocService) DeleteDoc(ctx context.Context, actor ports.Actor, id domain.DocID) error {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, doc.ProjectID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
