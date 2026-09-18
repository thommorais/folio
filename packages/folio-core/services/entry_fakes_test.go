package services_test

import (
	"context"
	"strings"

	"folio/folio-core/domain"
)

type fakeEntries struct {
	items map[domain.EntryID]domain.Entry
	order []domain.EntryID
}

func newFakeEntries() *fakeEntries {
	return &fakeEntries{items: map[domain.EntryID]domain.Entry{}}
}

func (r *fakeEntries) all() []domain.Entry {
	out := make([]domain.Entry, 0, len(r.order))
	for _, id := range r.order {
		if e, ok := r.items[id]; ok {
			out = append(out, e)
		}
	}
	return out
}

func (r *fakeEntries) List(_ context.Context, project domain.ProjectID, f domain.EntryFilter) ([]domain.Entry, error) {
	out := make([]domain.Entry, 0)
	for _, e := range r.all() {
		if e.ProjectID != project {
			continue
		}
		if f.Kind != "" && e.Kind != f.Kind {
			continue
		}
		if f.IssueID != "" && e.IssueID != f.IssueID {
			continue
		}
		if f.PlanID != "" && e.PlanID != f.PlanID {
			continue
		}
		if f.Branch != "" && e.Branch != f.Branch {
			continue
		}
		if q := strings.TrimSpace(f.Search); q != "" {
			if !strings.Contains(e.Title, q) && !strings.Contains(e.Body, q) {
				continue
			}
		}
		out = append(out, e)
	}
	if f.Offset > 0 {
		if f.Offset >= len(out) {
			return []domain.Entry{}, nil
		}
		out = out[f.Offset:]
	}
	if f.Limit > 0 && f.Limit < len(out) {
		out = out[:f.Limit]
	}
	return out, nil
}

func (r *fakeEntries) GetByID(_ context.Context, id domain.EntryID) (domain.Entry, error) {
	e, ok := r.items[id]
	if !ok {
		return domain.Entry{}, domain.ErrNotFound
	}
	return e, nil
}

func (r *fakeEntries) GetBySlug(_ context.Context, project domain.ProjectID, slug string) (domain.Entry, error) {
	for _, e := range r.all() {
		if e.ProjectID == project && e.Slug == slug {
			return e, nil
		}
	}
	return domain.Entry{}, domain.ErrNotFound
}

func (r *fakeEntries) Create(_ context.Context, e domain.Entry) (domain.Entry, error) {
	if _, exists := r.items[e.ID]; exists {
		return domain.Entry{}, domain.ErrConflict
	}
	r.items[e.ID] = e
	r.order = append(r.order, e.ID)
	return e, nil
}

func (r *fakeEntries) Update(_ context.Context, e domain.Entry) (domain.Entry, error) {
	if _, ok := r.items[e.ID]; !ok {
		return domain.Entry{}, domain.ErrNotFound
	}
	r.items[e.ID] = e
	return e, nil
}

func (r *fakeEntries) Delete(_ context.Context, id domain.EntryID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}
