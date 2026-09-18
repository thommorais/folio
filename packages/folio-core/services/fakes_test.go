package services_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// In-memory doubles for the driven ports. They enforce only what storage
// enforces (existence, uniqueness), never business rules, so a test failing
// here means the service under test let something through.

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type seqIDs struct {
	prefix string
	n      int
}

func (g *seqIDs) NewID() string {
	g.n++
	return fmt.Sprintf("%s%03d", g.prefix, g.n)
}

type nopLogger struct{}

func (nopLogger) Debug(string, map[string]any) {}
func (nopLogger) Info(string, map[string]any)  {}
func (nopLogger) Warn(string, map[string]any)  {}
func (nopLogger) Error(string, map[string]any) {}

type fakeProjects struct {
	items  map[domain.ProjectID]domain.Project
	emails map[string]domain.UserID
	// failOn forces the named method to return an error, to test propagation.
	failOn string
}

func newFakeProjects() *fakeProjects {
	return &fakeProjects{
		items:  map[domain.ProjectID]domain.Project{},
		emails: map[string]domain.UserID{},
	}
}

func (r *fakeProjects) fail(method string) error {
	if r.failOn == method {
		return fmt.Errorf("storage exploded in %s", method)
	}
	return nil
}

func (r *fakeProjects) List(_ context.Context, actor domain.UserID, includeArchived bool) ([]domain.Project, error) {
	if err := r.fail("List"); err != nil {
		return nil, err
	}
	out := []domain.Project{}
	for _, p := range r.items {
		if _, member := p.RoleOf(actor); !member {
			continue
		}
		if p.Archived && !includeArchived {
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakeProjects) GetByID(_ context.Context, id domain.ProjectID) (domain.Project, error) {
	if err := r.fail("GetByID"); err != nil {
		return domain.Project{}, err
	}
	p, ok := r.items[id]
	if !ok {
		return domain.Project{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *fakeProjects) GetBySlug(_ context.Context, slug string) (domain.Project, error) {
	for _, p := range r.items {
		if p.Slug == slug {
			return p, nil
		}
	}
	return domain.Project{}, domain.ErrNotFound
}

func (r *fakeProjects) Create(_ context.Context, p domain.Project) (domain.Project, error) {
	if err := r.fail("Create"); err != nil {
		return domain.Project{}, err
	}
	for _, existing := range r.items {
		if existing.Slug == p.Slug {
			return domain.Project{}, domain.ErrConflict
		}
	}
	r.items[p.ID] = p
	return p, nil
}

func (r *fakeProjects) Update(_ context.Context, p domain.Project) (domain.Project, error) {
	if err := r.fail("Update"); err != nil {
		return domain.Project{}, err
	}
	if _, ok := r.items[p.ID]; !ok {
		return domain.Project{}, domain.ErrNotFound
	}
	r.items[p.ID] = p
	return p, nil
}

func (r *fakeProjects) Delete(_ context.Context, id domain.ProjectID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *fakeProjects) AddMember(_ context.Context, id domain.ProjectID, user domain.UserID, role domain.Role) error {
	p, ok := r.items[id]
	if !ok {
		return domain.ErrNotFound
	}
	for i, m := range p.Members {
		if m.UserID == user {
			p.Members[i].Role = role
			r.items[id] = p
			return nil
		}
	}
	p.Members = append(p.Members, domain.Member{UserID: user, Role: role})
	r.items[id] = p
	return nil
}

func (r *fakeProjects) RemoveMember(_ context.Context, id domain.ProjectID, user domain.UserID) error {
	p, ok := r.items[id]
	if !ok {
		return domain.ErrNotFound
	}
	kept := make([]domain.Member, 0, len(p.Members))
	for _, m := range p.Members {
		if m.UserID != user {
			kept = append(kept, m)
		}
	}
	p.Members = kept
	r.items[id] = p
	return nil
}

func (r *fakeProjects) SetMemberRole(ctx context.Context, id domain.ProjectID, user domain.UserID, role domain.Role) error {
	return r.AddMember(ctx, id, user, role)
}

func (r *fakeProjects) FindUserByEmail(_ context.Context, email string) (domain.UserID, error) {
	id, ok := r.emails[strings.ToLower(email)]
	if !ok {
		return "", domain.ErrNotFound
	}
	return id, nil
}

type fakePlans struct {
	items  map[domain.PlanID]domain.Plan
	failOn string
}

func newFakePlans() *fakePlans { return &fakePlans{items: map[domain.PlanID]domain.Plan{}} }

func (r *fakePlans) List(_ context.Context, project domain.ProjectID, statuses []domain.PlanStatus) ([]domain.Plan, error) {
	if r.failOn == "List" {
		return nil, fmt.Errorf("storage exploded")
	}
	allow := map[domain.PlanStatus]bool{}
	for _, s := range statuses {
		allow[s] = true
	}
	out := []domain.Plan{}
	for _, p := range r.items {
		if p.ProjectID != project {
			continue
		}
		if len(allow) > 0 && !allow[p.Status] {
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakePlans) ListByIssue(_ context.Context, issue domain.IssueID) ([]domain.Plan, error) {
	out := []domain.Plan{}
	for _, p := range r.items {
		if p.IssueID == issue {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakePlans) GetByID(_ context.Context, id domain.PlanID) (domain.Plan, error) {
	p, ok := r.items[id]
	if !ok {
		return domain.Plan{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *fakePlans) Create(_ context.Context, p domain.Plan) (domain.Plan, error) {
	if r.failOn == "Create" {
		return domain.Plan{}, fmt.Errorf("storage exploded")
	}
	r.items[p.ID] = p
	return p, nil
}

func (r *fakePlans) Update(_ context.Context, p domain.Plan) (domain.Plan, error) {
	if _, ok := r.items[p.ID]; !ok {
		return domain.Plan{}, domain.ErrNotFound
	}
	r.items[p.ID] = p
	return p, nil
}

func (r *fakePlans) Delete(_ context.Context, id domain.PlanID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

type fakeSearch struct {
	hits []domain.SearchHit
}

func (r *fakeSearch) Search(_ context.Context, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	out := []domain.SearchHit{}
	for _, h := range r.hits {
		if h.ProjectID == project {
			out = append(out, h)
		}
	}
	return out, nil
}

type fakeCycles struct {
	items map[domain.CycleID]domain.Cycle
}

func newFakeCycles() *fakeCycles {
	return &fakeCycles{items: map[domain.CycleID]domain.Cycle{}}
}

func (r *fakeCycles) ListByIssue(_ context.Context, issue domain.IssueID) ([]domain.Cycle, error) {
	out := []domain.Cycle{}
	for _, c := range r.items {
		if c.IssueID == issue {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ordinal < out[j].Ordinal })
	return out, nil
}

func (r *fakeCycles) GetByID(_ context.Context, id domain.CycleID) (domain.Cycle, error) {
	c, ok := r.items[id]
	if !ok {
		return domain.Cycle{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *fakeCycles) Create(_ context.Context, c domain.Cycle) (domain.Cycle, error) {
	r.items[c.ID] = c
	return c, nil
}

func (r *fakeCycles) Update(_ context.Context, c domain.Cycle) (domain.Cycle, error) {
	if _, ok := r.items[c.ID]; !ok {
		return domain.Cycle{}, domain.ErrNotFound
	}
	r.items[c.ID] = c
	return c, nil
}

func (r *fakeCycles) Delete(_ context.Context, id domain.CycleID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

// compile-time checks that the doubles satisfy the ports they stand in for.
var (
	_ ports.ProjectRepository   = (*fakeProjects)(nil)
	_ ports.PlanRepository      = (*fakePlans)(nil)
	_ ports.IssueRepository     = (*fakeIssues)(nil)
	_ ports.EntryRepository     = (*fakeEntries)(nil)
	_ ports.CycleRepository     = (*fakeCycles)(nil)
	_ ports.SearchRepository    = (*fakeSearch)(nil)
	_ ports.Clock               = (*fakeClock)(nil)
	_ ports.IDGenerator         = (*seqIDs)(nil)
	_ ports.Logger              = nopLogger{}
)
