package services_test

import (
	"context"
	"strings"

	"folio/folio-core/domain"
)

type fakeIssues struct {
	items map[domain.IssueID]domain.Issue
	links []domain.IssueLink
	order []domain.IssueID
}

func newFakeIssues() *fakeIssues {
	return &fakeIssues{items: map[domain.IssueID]domain.Issue{}}
}

func (r *fakeIssues) all() []domain.Issue {
	out := make([]domain.Issue, 0, len(r.order))
	for _, id := range r.order {
		if issue, ok := r.items[id]; ok {
			out = append(out, issue)
		}
	}
	return out
}

func (r *fakeIssues) List(_ context.Context, project domain.ProjectID, f domain.IssueFilter) ([]domain.Issue, error) {
	out := make([]domain.Issue, 0, len(r.items))
	for _, issue := range r.all() {
		if issue.ProjectID != project {
			continue
		}
		if f.Kind != "" && issue.Kind != f.Kind {
			continue
		}
		if f.PlanID != "" && issue.PlanID != f.PlanID {
			continue
		}
		if f.Priority != "" && issue.Priority != f.Priority {
			continue
		}
		if f.Assignee != "" && issue.Assignee != f.Assignee {
			continue
		}
		if len(f.Status) > 0 {
			match := false
			for _, s := range f.Status {
				if issue.Status == s {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if q := strings.TrimSpace(f.Search); q != "" {
			if !strings.Contains(issue.Title, q) && !strings.Contains(issue.Body, q) {
				continue
			}
		}
		out = append(out, issue)
	}
	if f.ParentID != "" {
		kept := out[:0]
		for _, issue := range out {
			for _, l := range r.links {
				if l.Kind == domain.LinkParent && l.From == issue.ID && l.To == f.ParentID {
					kept = append(kept, issue)
					break
				}
			}
		}
		out = kept
	}
	if f.Offset > 0 {
		if f.Offset >= len(out) {
			return []domain.Issue{}, nil
		}
		out = out[f.Offset:]
	}
	if f.Limit > 0 && f.Limit < len(out) {
		out = out[:f.Limit]
	}
	return out, nil
}

func (r *fakeIssues) ListByParent(_ context.Context, parent domain.IssueID) ([]domain.Issue, error) {
	out := make([]domain.Issue, 0)
	for _, l := range r.links {
		if l.Kind == domain.LinkParent && l.To == parent {
			if issue, ok := r.items[l.From]; ok {
				out = append(out, issue)
			}
		}
	}
	return out, nil
}

func (r *fakeIssues) ListByPlan(_ context.Context, plan domain.PlanID) ([]domain.Issue, error) {
	out := make([]domain.Issue, 0)
	for _, issue := range r.all() {
		if issue.PlanID == plan {
			out = append(out, issue)
		}
	}
	return out, nil
}

func (r *fakeIssues) GetByID(_ context.Context, id domain.IssueID) (domain.Issue, error) {
	issue, ok := r.items[id]
	if !ok {
		return domain.Issue{}, domain.ErrNotFound
	}
	return issue, nil
}

func (r *fakeIssues) GetBySlug(_ context.Context, project domain.ProjectID, slug string) (domain.Issue, error) {
	for _, issue := range r.all() {
		if issue.ProjectID == project && issue.Slug == slug {
			return issue, nil
		}
	}
	return domain.Issue{}, domain.ErrNotFound
}

func (r *fakeIssues) Create(_ context.Context, i domain.Issue) (domain.Issue, error) {
	if _, exists := r.items[i.ID]; exists {
		return domain.Issue{}, domain.ErrConflict
	}
	r.items[i.ID] = i
	r.order = append(r.order, i.ID)
	return i, nil
}

func (r *fakeIssues) Update(_ context.Context, i domain.Issue) (domain.Issue, error) {
	if _, ok := r.items[i.ID]; !ok {
		return domain.Issue{}, domain.ErrNotFound
	}
	r.items[i.ID] = i
	return i, nil
}

func (r *fakeIssues) Delete(_ context.Context, id domain.IssueID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	kept := r.links[:0]
	for _, l := range r.links {
		if l.From != id && l.To != id {
			kept = append(kept, l)
		}
	}
	r.links = kept
	return nil
}

func (r *fakeIssues) Links(_ context.Context, id domain.IssueID) ([]domain.IssueLink, error) {
	out := make([]domain.IssueLink, 0)
	for _, l := range r.links {
		if l.From == id || l.To == id {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *fakeIssues) LinksOfProject(_ context.Context, project domain.ProjectID) ([]domain.IssueLink, error) {
	out := make([]domain.IssueLink, 0)
	for _, l := range r.links {
		from, ok := r.items[l.From]
		if ok && from.ProjectID == project {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *fakeIssues) Link(_ context.Context, from, to domain.IssueID, kind domain.LinkKind) error {
	for _, l := range r.links {
		if l.From == from && l.To == to && l.Kind == kind {
			return nil
		}
	}
	r.links = append(r.links, domain.IssueLink{From: from, To: to, Kind: kind})
	return nil
}

func (r *fakeIssues) Unlink(_ context.Context, from, to domain.IssueID, kind domain.LinkKind) error {
	for i, l := range r.links {
		if l.From == from && l.To == to && l.Kind == kind {
			r.links = append(r.links[:i], r.links[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *fakeIssues) SetLinks(ctx context.Context, from domain.IssueID, kind domain.LinkKind, to []domain.IssueID) error {
	want := make(map[domain.IssueID]bool, len(to))
	for _, id := range to {
		want[id] = true
	}
	kept := r.links[:0]
	for _, l := range r.links {
		if l.From == from && l.Kind == kind {
			if want[l.To] {
				delete(want, l.To)
				kept = append(kept, l)
			}
			continue
		}
		kept = append(kept, l)
	}
	r.links = kept
	for id := range want {
		if err := r.Link(ctx, from, id, kind); err != nil {
			return err
		}
	}
	return nil
}
