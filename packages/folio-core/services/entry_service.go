package services

import (
	"context"
	"fmt"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type EntryService struct {
	repo   ports.EntryRepository
	issues ports.IssueRepository
	plans  ports.PlanRepository
	guard  ports.Guard
	clock  ports.Clock
	ids    ports.IDGenerator
	log    ports.Logger
}

func NewEntryService(repo ports.EntryRepository, issues ports.IssueRepository, plans ports.PlanRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *EntryService {
	return &EntryService{repo: repo, issues: issues, plans: plans, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.EntryUseCase = (*EntryService)(nil)

func (s *EntryService) ListEntries(ctx context.Context, actor ports.Actor, project domain.ProjectID, f domain.EntryFilter) ([]domain.Entry, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	f.Limit = clampLimit(f.Limit)
	entries, err := s.repo.List(ctx, project, f)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []domain.Entry{}
	}
	return entries, nil
}

func (s *EntryService) GetEntry(ctx context.Context, actor ports.Actor, id domain.EntryID) (domain.Entry, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Entry{}, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, entry.ProjectID); err != nil {
		return domain.Entry{}, err
	}
	return entry, nil
}

func (s *EntryService) GetEntryBySlug(ctx context.Context, actor ports.Actor, project domain.ProjectID, slug string) (domain.Entry, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return domain.Entry{}, err
	}
	return s.repo.GetBySlug(ctx, project, slug)
}

func (s *EntryService) WriteEntry(ctx context.Context, actor ports.Actor, in ports.WriteEntryInput) (domain.Entry, error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, in.ProjectID); err != nil {
		return domain.Entry{}, err
	}
	kind := in.Kind
	if kind == "" {
		kind = domain.EntryJournal
	}
	if err := issueScope(ctx, s.issues, in.IssueID, in.ProjectID); err != nil {
		return domain.Entry{}, err
	}
	if err := s.checkPlan(ctx, in.ProjectID, in.PlanID); err != nil {
		return domain.Entry{}, err
	}

	now := s.clock.Now()
	entry := domain.Entry{
		ID:          domain.EntryID(s.ids.NewID()),
		Kind:        kind,
		ProjectID:   in.ProjectID,
		IssueID:     in.IssueID,
		PlanID:      in.PlanID,
		CycleID:     in.CycleID,
		Title:       strings.TrimSpace(in.Title),
		Body:        in.Body,
		Branch:      strings.TrimSpace(in.Branch),
		PR:          strings.TrimSpace(in.PR),
		ExternalRef: strings.TrimSpace(in.ExternalRef),
		Tags:        in.Tags,
		Meta:        in.Meta,
		CreatedBy:   actor.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if kind.Addressable() {
		slug, err := s.freeSlug(ctx, in.ProjectID, in.Slug, entry.Title)
		if err != nil {
			return domain.Entry{}, err
		}
		entry.Slug = slug
	}
	if err := rules.ValidateEntry(entry); err != nil {
		return domain.Entry{}, err
	}
	return s.repo.Create(ctx, entry)
}

func (s *EntryService) checkPlan(ctx context.Context, project domain.ProjectID, plan domain.PlanID) error {
	if plan == "" {
		return nil
	}
	found, err := s.plans.GetByID(ctx, plan)
	if err != nil {
		if notFound(err) {
			return domain.Invalid("plan", "does not exist")
		}
		return err
	}
	if found.ProjectID != project {
		return domain.Invalid("plan", "belongs to a different project")
	}
	return nil
}

func (s *EntryService) freeSlug(ctx context.Context, project domain.ProjectID, requested, title string) (string, error) {
	if slug := strings.TrimSpace(requested); slug != "" {
		free, err := s.slugFree(ctx, project, slug)
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
		base = "entry"
	}
	if len(base) > 50 {
		base = strings.TrimRight(base[:50], "-")
	}

	candidate := base
	for n := 2; ; n++ {
		free, err := s.slugFree(ctx, project, candidate)
		if err != nil {
			return "", err
		}
		if free {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, n)
	}
}

func (s *EntryService) slugFree(ctx context.Context, project domain.ProjectID, slug string) (bool, error) {
	_, err := s.repo.GetBySlug(ctx, project, slug)
	if err != nil {
		if notFound(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (s *EntryService) UpdateEntry(ctx context.Context, actor ports.Actor, id domain.EntryID, in ports.UpdateEntryInput) (domain.Entry, error) {
	entry, err := s.writable(ctx, actor, id)
	if err != nil {
		return domain.Entry{}, err
	}

	if in.IssueID != nil {
		if err := issueScope(ctx, s.issues, *in.IssueID, entry.ProjectID); err != nil {
			return domain.Entry{}, err
		}
		entry.IssueID = *in.IssueID
	}
	if in.PlanID != nil {
		if err := s.checkPlan(ctx, entry.ProjectID, *in.PlanID); err != nil {
			return domain.Entry{}, err
		}
		entry.PlanID = *in.PlanID
	}
	if in.Title != nil {
		entry.Title = strings.TrimSpace(*in.Title)
	}
	if in.Body != nil {
		entry.Body = *in.Body
	}
	if in.Branch != nil {
		entry.Branch = strings.TrimSpace(*in.Branch)
	}
	if in.PR != nil {
		entry.PR = strings.TrimSpace(*in.PR)
	}
	if in.ExternalRef != nil {
		entry.ExternalRef = strings.TrimSpace(*in.ExternalRef)
	}
	if in.Tags != nil {
		entry.Tags = *in.Tags
	}
	if in.Meta != nil {
		entry.Meta = *in.Meta
	}
	if in.Slug != nil && entry.Kind.Addressable() {
		slug := strings.TrimSpace(*in.Slug)
		if slug != entry.Slug {
			free, err := s.slugFree(ctx, entry.ProjectID, slug)
			if err != nil {
				return domain.Entry{}, err
			}
			if !free {
				return domain.Entry{}, domain.ErrConflict
			}
			entry.Slug = slug
		}
	}

	if err := rules.ValidateEntry(entry); err != nil {
		return domain.Entry{}, err
	}
	return s.save(ctx, entry)
}

func (s *EntryService) AppendToEntry(ctx context.Context, actor ports.Actor, id domain.EntryID, section string) (domain.Entry, error) {
	text := strings.TrimSpace(section)
	if text == "" {
		return domain.Entry{}, domain.Invalid("section", "is required")
	}
	entry, err := s.writable(ctx, actor, id)
	if err != nil {
		return domain.Entry{}, err
	}

	if existing := strings.TrimRight(entry.Body, "\n"); existing == "" {
		entry.Body = text
	} else {
		entry.Body = existing + "\n\n" + text
	}
	return s.save(ctx, entry)
}

func (s *EntryService) DeleteEntry(ctx context.Context, actor ports.Actor, id domain.EntryID) error {
	if _, err := s.writable(ctx, actor, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *EntryService) writable(ctx context.Context, actor ports.Actor, id domain.EntryID) (domain.Entry, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Entry{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, entry.ProjectID); err != nil {
		return domain.Entry{}, err
	}
	return entry, nil
}

func (s *EntryService) save(ctx context.Context, entry domain.Entry) (domain.Entry, error) {
	entry.UpdatedAt = s.clock.Now()
	return s.repo.Update(ctx, entry)
}
