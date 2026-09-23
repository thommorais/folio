package services_test

import (
	"context"
	"sort"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type fakeKnowledge struct {
	items map[domain.KnowledgeID]domain.Knowledge
	// lastLimit is what the service passed down, which is how the clamping
	// test sees a limit the caller never asked for.
	lastLimit int
}

func newFakeKnowledge() *fakeKnowledge {
	return &fakeKnowledge{items: map[domain.KnowledgeID]domain.Knowledge{}}
}

var _ ports.KnowledgeRepository = (*fakeKnowledge)(nil)

func (r *fakeKnowledge) List(_ context.Context, f domain.KnowledgeFilter) ([]domain.Knowledge, error) {
	r.lastLimit = f.Limit

	out := []domain.Knowledge{}
	for _, k := range r.items {
		if f.Unattached && k.ProjectID != "" {
			continue
		}
		if f.ProjectID != "" && k.ProjectID != f.ProjectID {
			continue
		}
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakeKnowledge) GetByID(_ context.Context, id domain.KnowledgeID) (domain.Knowledge, error) {
	k, ok := r.items[id]
	if !ok {
		return domain.Knowledge{}, domain.ErrNotFound
	}
	return k, nil
}

func (r *fakeKnowledge) GetBySlug(_ context.Context, slug string) (domain.Knowledge, error) {
	for _, k := range r.items {
		if k.Slug == slug {
			return k, nil
		}
	}
	return domain.Knowledge{}, domain.ErrNotFound
}

func (r *fakeKnowledge) Create(_ context.Context, k domain.Knowledge) (domain.Knowledge, error) {
	r.items[k.ID] = k
	return k, nil
}

func (r *fakeKnowledge) Update(_ context.Context, k domain.Knowledge) (domain.Knowledge, error) {
	if _, ok := r.items[k.ID]; !ok {
		return domain.Knowledge{}, domain.ErrNotFound
	}
	r.items[k.ID] = k
	return k, nil
}

func (r *fakeKnowledge) Delete(_ context.Context, id domain.KnowledgeID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}
