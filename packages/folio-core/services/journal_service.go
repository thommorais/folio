package services

import (
	"context"
	"fmt"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

// Paging bounds. The work log is the project's memory and grows without
// bound, so every listing is clamped: an unbounded read would eventually time
// out or blow a context window.
const (
	DefaultPageSize = 50
	MaxPageSize     = 500
)

// clampLimit applies the default when unset and the ceiling when over.
func clampLimit(limit int) int {
	if limit <= 0 {
		return DefaultPageSize
	}
	if limit > MaxPageSize {
		return MaxPageSize
	}
	return limit
}

// JournalService manages the work log: titled, searchable entries describing what
// was built, how, and where it stands.
type JournalService struct {
	repo    ports.JournalRepository
	issues ports.IssueRepository
	guard   ports.Guard
	clock   ports.Clock
	ids     ports.IDGenerator
	log     ports.Logger
}

func NewJournalService(repo ports.JournalRepository, issues ports.IssueRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *JournalService {
	return &JournalService{repo: repo, issues: issues, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.JournalUseCase = (*JournalService)(nil)

func (s *JournalService) ListJournal(ctx context.Context, actor ports.Actor, project domain.ProjectID, f domain.JournalFilter) ([]domain.JournalEntry, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	f.Limit = clampLimit(f.Limit)
	entries, err := s.repo.List(ctx, project, f)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []domain.JournalEntry{}
	}
	return entries, nil
}

func (s *JournalService) GetJournalEntry(ctx context.Context, actor ports.Actor, id domain.JournalID) (domain.JournalEntry, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.JournalEntry{}, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, entry.ProjectID); err != nil {
		return domain.JournalEntry{}, err
	}
	return entry, nil
}

func (s *JournalService) GetJournalEntryBySlug(ctx context.Context, actor ports.Actor, project domain.ProjectID, slug string) (domain.JournalEntry, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return domain.JournalEntry{}, err
	}
	return s.repo.GetBySlug(ctx, project, slug)
}

func (s *JournalService) WriteJournalEntry(ctx context.Context, actor ports.Actor, in ports.WriteJournalInput) (domain.JournalEntry, error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, in.ProjectID); err != nil {
		return domain.JournalEntry{}, err
	}

	if err := issueScope(ctx, s.issues, in.IssueID, in.ProjectID); err != nil {
		return domain.JournalEntry{}, err
	}

	slug, err := s.freeSlug(ctx, in.ProjectID, in.Slug, in.Title)
	if err != nil {
		return domain.JournalEntry{}, err
	}

	now := s.clock.Now()
	entry := domain.JournalEntry{
		ID:          domain.JournalID(s.ids.NewID()),
		ProjectID:   in.ProjectID,
		Slug:        slug,
		IssueID:     in.IssueID,
		PlanID:      in.PlanID,
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
	if err := rules.ValidateJournalEntry(entry); err != nil {
		return domain.JournalEntry{}, err
	}
	return s.repo.Create(ctx, entry)
}

// freeSlug conflicts on an explicit slug, the way a doc does, but suffixes a
// derived one: journal titles repeat by nature, and refusing the second "fix
// the build" would refuse to record work that happened.
func (s *JournalService) freeSlug(ctx context.Context, project domain.ProjectID, requested, title string) (string, error) {
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
	// Leaves room for a suffix inside the 60-character column.
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

func (s *JournalService) slugFree(ctx context.Context, project domain.ProjectID, slug string) (bool, error) {
	_, err := s.repo.GetBySlug(ctx, project, slug)
	if err != nil {
		if notFound(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (s *JournalService) UpdateJournalEntry(ctx context.Context, actor ports.Actor, id domain.JournalID, in ports.UpdateJournalInput) (domain.JournalEntry, error) {
	entry, err := s.writable(ctx, actor, id)
	if err != nil {
		return domain.JournalEntry{}, err
	}

	if in.IssueID != nil {
		if err := issueScope(ctx, s.issues, *in.IssueID, entry.ProjectID); err != nil {
			return domain.JournalEntry{}, err
		}
		entry.IssueID = *in.IssueID
	}
	if in.PlanID != nil {
		entry.PlanID = *in.PlanID
	}
	if in.IssueID != nil {
		entry.IssueID = *in.IssueID
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
	return s.save(ctx, entry)
}

// AppendToJournalEntry adds a section to an entry's body. An agent recording progress
// on work it already wrote up should not have to read, splice and resend the
// whole body just to add a paragraph.
func (s *JournalService) AppendToJournalEntry(ctx context.Context, actor ports.Actor, id domain.JournalID, section string) (domain.JournalEntry, error) {
	text := strings.TrimSpace(section)
	if text == "" {
		return domain.JournalEntry{}, domain.Invalid("section", "is required")
	}
	entry, err := s.writable(ctx, actor, id)
	if err != nil {
		return domain.JournalEntry{}, err
	}

	if existing := strings.TrimRight(entry.Body, "\n"); existing == "" {
		entry.Body = text
	} else {
		// A blank line between sections keeps the body valid markdown.
		entry.Body = existing + "\n\n" + text
	}
	return s.save(ctx, entry)
}

func (s *JournalService) DeleteJournalEntry(ctx context.Context, actor ports.Actor, id domain.JournalID) error {
	if _, err := s.writable(ctx, actor, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// writable loads an entry and checks the actor may change it.
func (s *JournalService) writable(ctx context.Context, actor ports.Actor, id domain.JournalID) (domain.JournalEntry, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.JournalEntry{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, entry.ProjectID); err != nil {
		return domain.JournalEntry{}, err
	}
	return entry, nil
}

// save validates and persists an edit. CreatedAt is left untouched: an edit
// records when the entry changed, never when the work happened.
func (s *JournalService) save(ctx context.Context, entry domain.JournalEntry) (domain.JournalEntry, error) {
	if err := rules.ValidateJournalEntry(entry); err != nil {
		return domain.JournalEntry{}, err
	}
	entry.UpdatedAt = s.clock.Now()
	return s.repo.Update(ctx, entry)
}
