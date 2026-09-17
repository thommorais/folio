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

type fakeJournal struct {
	items  []domain.JournalEntry
	failOn string
	// lastFilter records what the service actually asked storage for, so
	// tests can assert on clamping the service applies before the call.
	lastFilter domain.JournalFilter
}

func newFakeJournal() *fakeJournal { return &fakeJournal{} }

func (r *fakeJournal) List(_ context.Context, project domain.ProjectID, f domain.JournalFilter) ([]domain.JournalEntry, error) {
	r.lastFilter = f
	if r.failOn == "List" {
		return nil, fmt.Errorf("storage exploded")
	}
	out := []domain.JournalEntry{}
	for _, e := range r.items {
		if e.ProjectID != project {
			continue
		}
		if f.PlanID != "" && e.PlanID != f.PlanID {
			continue
		}
		if f.IssueID != "" && e.IssueID != f.IssueID {
			continue
		}
		if f.Branch != "" && e.Branch != f.Branch {
			continue
		}
		if f.ExternalRef != "" && e.ExternalRef != f.ExternalRef {
			continue
		}
		if f.IssueID != "" && e.IssueID != f.IssueID {
			continue
		}
		if f.Search != "" {
			q := strings.ToLower(f.Search)
			if !strings.Contains(strings.ToLower(e.Title), q) && !strings.Contains(strings.ToLower(e.Body), q) {
				continue
			}
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *fakeJournal) GetByID(_ context.Context, id domain.JournalID) (domain.JournalEntry, error) {
	for _, e := range r.items {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.JournalEntry{}, domain.ErrNotFound
}

func (r *fakeJournal) GetBySlug(_ context.Context, project domain.ProjectID, slug string) (domain.JournalEntry, error) {
	for _, e := range r.items {
		if e.ProjectID == project && e.Slug == slug {
			return e, nil
		}
	}
	return domain.JournalEntry{}, domain.ErrNotFound
}

func (r *fakeJournal) Create(_ context.Context, e domain.JournalEntry) (domain.JournalEntry, error) {
	if r.failOn == "Create" {
		return domain.JournalEntry{}, fmt.Errorf("storage exploded")
	}
	r.items = append(r.items, e)
	return e, nil
}

func (r *fakeJournal) Update(_ context.Context, e domain.JournalEntry) (domain.JournalEntry, error) {
	for i, existing := range r.items {
		if existing.ID == e.ID {
			r.items[i] = e
			return e, nil
		}
	}
	return domain.JournalEntry{}, domain.ErrNotFound
}

func (r *fakeJournal) Delete(_ context.Context, id domain.JournalID) error {
	for i, e := range r.items {
		if e.ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}

type fakeDocs struct {
	items map[domain.DocID]domain.Doc
}

func newFakeDocs() *fakeDocs { return &fakeDocs{items: map[domain.DocID]domain.Doc{}} }

func (r *fakeDocs) List(_ context.Context, project domain.ProjectID, f domain.DocFilter) ([]domain.Doc, error) {
	out := []domain.Doc{}
	for _, d := range r.items {
		if f.IssueID != "" && d.IssueID != f.IssueID {
			continue
		}
		if d.ProjectID != project {
			continue
		}
		if f.Search != "" && !strings.Contains(strings.ToLower(d.Title), strings.ToLower(f.Search)) {
			continue
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakeDocs) GetByID(_ context.Context, id domain.DocID) (domain.Doc, error) {
	d, ok := r.items[id]
	if !ok {
		return domain.Doc{}, domain.ErrNotFound
	}
	return d, nil
}

func (r *fakeDocs) GetBySlug(_ context.Context, project domain.ProjectID, slug string) (domain.Doc, error) {
	for _, d := range r.items {
		if d.ProjectID == project && d.Slug == slug {
			return d, nil
		}
	}
	return domain.Doc{}, domain.ErrNotFound
}

func (r *fakeDocs) Create(_ context.Context, d domain.Doc) (domain.Doc, error) {
	r.items[d.ID] = d
	return d, nil
}

func (r *fakeDocs) Update(_ context.Context, d domain.Doc) (domain.Doc, error) {
	if _, ok := r.items[d.ID]; !ok {
		return domain.Doc{}, domain.ErrNotFound
	}
	r.items[d.ID] = d
	return d, nil
}

func (r *fakeDocs) Delete(_ context.Context, id domain.DocID) error {
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

type fakeTicketLogs struct {
	items map[domain.TicketLogID]domain.TicketLog
}

func newFakeTicketLogs() *fakeTicketLogs {
	return &fakeTicketLogs{items: map[domain.TicketLogID]domain.TicketLog{}}
}

func (r *fakeTicketLogs) List(_ context.Context, project domain.ProjectID, f domain.TicketLogFilter) ([]domain.TicketLog, error) {
	out := []domain.TicketLog{}
	for _, l := range r.items {
		if l.ProjectID != project {
			continue
		}
		if f.IssueID != "" && l.IssueID != f.IssueID {
			continue
		}
		if f.CycleID != "" && l.CycleID != f.CycleID {
			continue
		}
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakeTicketLogs) GetByID(_ context.Context, id domain.TicketLogID) (domain.TicketLog, error) {
	l, ok := r.items[id]
	if !ok {
		return domain.TicketLog{}, domain.ErrNotFound
	}
	return l, nil
}

func (r *fakeTicketLogs) Create(_ context.Context, l domain.TicketLog) (domain.TicketLog, error) {
	r.items[l.ID] = l
	return l, nil
}

func (r *fakeTicketLogs) Update(_ context.Context, l domain.TicketLog) (domain.TicketLog, error) {
	if _, ok := r.items[l.ID]; !ok {
		return domain.TicketLog{}, domain.ErrNotFound
	}
	r.items[l.ID] = l
	return l, nil
}

func (r *fakeTicketLogs) Delete(_ context.Context, id domain.TicketLogID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

type fakePlanLogs struct {
	items map[domain.PlanLogID]domain.PlanLog
}

func newFakePlanLogs() *fakePlanLogs {
	return &fakePlanLogs{items: map[domain.PlanLogID]domain.PlanLog{}}
}

func (r *fakePlanLogs) List(_ context.Context, project domain.ProjectID, f domain.PlanLogFilter) ([]domain.PlanLog, error) {
	out := []domain.PlanLog{}
	for _, l := range r.items {
		if l.ProjectID != project {
			continue
		}
		if f.PlanID != "" && l.PlanID != f.PlanID {
			continue
		}
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakePlanLogs) GetByID(_ context.Context, id domain.PlanLogID) (domain.PlanLog, error) {
	l, ok := r.items[id]
	if !ok {
		return domain.PlanLog{}, domain.ErrNotFound
	}
	return l, nil
}

func (r *fakePlanLogs) Create(_ context.Context, l domain.PlanLog) (domain.PlanLog, error) {
	r.items[l.ID] = l
	return l, nil
}

func (r *fakePlanLogs) Update(_ context.Context, l domain.PlanLog) (domain.PlanLog, error) {
	if _, ok := r.items[l.ID]; !ok {
		return domain.PlanLog{}, domain.ErrNotFound
	}
	r.items[l.ID] = l
	return l, nil
}

func (r *fakePlanLogs) Delete(_ context.Context, id domain.PlanLogID) error {
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

type fakeTodoLogs struct {
	items map[domain.TodoLogID]domain.TodoLog
}

func newFakeTodoLogs() *fakeTodoLogs {
	return &fakeTodoLogs{items: map[domain.TodoLogID]domain.TodoLog{}}
}

func (r *fakeTodoLogs) List(_ context.Context, project domain.ProjectID, f domain.TodoLogFilter) ([]domain.TodoLog, error) {
	out := []domain.TodoLog{}
	for _, l := range r.items {
		if l.ProjectID != project {
			continue
		}
		if f.IssueID != "" && l.IssueID != f.IssueID {
			continue
		}
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakeTodoLogs) GetByID(_ context.Context, id domain.TodoLogID) (domain.TodoLog, error) {
	l, ok := r.items[id]
	if !ok {
		return domain.TodoLog{}, domain.ErrNotFound
	}
	return l, nil
}

func (r *fakeTodoLogs) Create(_ context.Context, l domain.TodoLog) (domain.TodoLog, error) {
	r.items[l.ID] = l
	return l, nil
}

func (r *fakeTodoLogs) Update(_ context.Context, l domain.TodoLog) (domain.TodoLog, error) {
	if _, ok := r.items[l.ID]; !ok {
		return domain.TodoLog{}, domain.ErrNotFound
	}
	r.items[l.ID] = l
	return l, nil
}

func (r *fakeTodoLogs) Delete(_ context.Context, id domain.TodoLogID) error {
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
	_ ports.JournalRepository   = (*fakeJournal)(nil)
	_ ports.DocRepository       = (*fakeDocs)(nil)
	_ ports.IssueRepository     = (*fakeIssues)(nil)
	_ ports.CycleRepository     = (*fakeCycles)(nil)
	_ ports.TicketLogRepository = (*fakeTicketLogs)(nil)
	_ ports.PlanLogRepository   = (*fakePlanLogs)(nil)
	_ ports.TodoLogRepository   = (*fakeTodoLogs)(nil)
	_ ports.SearchRepository    = (*fakeSearch)(nil)
	_ ports.Clock               = (*fakeClock)(nil)
	_ ports.IDGenerator         = (*seqIDs)(nil)
	_ ports.Logger              = nopLogger{}
)
