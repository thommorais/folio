package services_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	"folio/folio-core/domain"
)

type seqTokens struct{ n int }

func (g *seqTokens) NewToken() string {
	g.n++
	return fmt.Sprintf("tok%040d", g.n)
}

type fakeShares struct {
	items map[domain.ShareID]domain.Share
}

func newFakeShares() *fakeShares { return &fakeShares{items: map[domain.ShareID]domain.Share{}} }

func (r *fakeShares) List(_ context.Context, project domain.ProjectID) ([]domain.Share, error) {
	out := []domain.Share{}
	for _, s := range r.items {
		if s.ProjectID == project {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakeShares) GetByID(_ context.Context, id domain.ShareID) (domain.Share, error) {
	s, ok := r.items[id]
	if !ok {
		return domain.Share{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *fakeShares) GetByToken(_ context.Context, token string) (domain.Share, error) {
	for _, s := range r.items {
		if s.Token == token {
			return s, nil
		}
	}
	return domain.Share{}, domain.ErrNotFound
}

func (r *fakeShares) Create(_ context.Context, s domain.Share) (domain.Share, error) {
	r.items[s.ID] = s
	return s, nil
}

func (r *fakeShares) Delete(_ context.Context, id domain.ShareID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *fakeShares) Touch(_ context.Context, id domain.ShareID, at time.Time) error {
	s, ok := r.items[id]
	if !ok {
		return domain.ErrNotFound
	}
	s.LastAccessedAt = &at
	r.items[id] = s
	return nil
}
