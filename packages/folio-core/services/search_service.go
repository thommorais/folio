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
	repo     ports.SearchRepository
	guard    ports.Guard
	projects ports.ProjectRepository
}

func NewSearchService(repo ports.SearchRepository, guard ports.Guard, projects ports.ProjectRepository) *SearchService {
	return &SearchService{repo: repo, guard: guard, projects: projects}
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

// SearchAll answers a query over every project the actor belongs to. The
// membership list is the authorisation: a project the actor cannot read never
// reaches the repository, so there is nothing to filter out afterwards.
func (s *SearchService) SearchAll(ctx context.Context, actor ports.Actor, q domain.SearchQuery) ([]domain.SearchHit, error) {
	q.Text = strings.TrimSpace(q.Text)
	if q.Text == "" && len(q.Tags) == 0 {
		return nil, domain.Invalid("q", "is required")
	}
	q.Limit = clampLimit(q.Limit)

	projects, err := s.projects.List(ctx, actor.UserID, true)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return []domain.SearchHit{}, nil
	}

	ids := make([]domain.ProjectID, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}

	hits, err := s.repo.SearchAcross(ctx, ids, q)
	if err != nil {
		return nil, err
	}
	if hits == nil {
		hits = []domain.SearchHit{}
	}
	return hits, nil
}
