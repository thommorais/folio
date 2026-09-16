package services

import (
	"context"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// SearchService is a thin use case: it authorises, clamps and delegates. The
// matching itself belongs to the storage adapter, which can push it into the
// database instead of fanning out over the other repositories.
type SearchService struct {
	repo  ports.SearchRepository
	guard ports.Guard
}

func NewSearchService(repo ports.SearchRepository, guard ports.Guard) *SearchService {
	return &SearchService{repo: repo, guard: guard}
}

var _ ports.SearchUseCase = (*SearchService)(nil)

func (s *SearchService) Search(ctx context.Context, actor ports.Actor, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	q.Text = strings.TrimSpace(q.Text)
	// An empty query would degenerate into a full scan of the project.
	if q.Text == "" && len(q.Tags) == 0 {
		return nil, domain.Invalid("q", "is required")
	}
	q.Limit = clampLimit(q.Limit)

	hits, err := s.repo.Search(ctx, project, q)
	if err != nil {
		return nil, err
	}
	if hits == nil {
		hits = []domain.SearchHit{}
	}
	return hits, nil
}
