// Package httpapi is the driving adapter that exposes the folio use cases as
// a REST API mounted on PocketBase's router. It maps domain models to wire
// DTOs and domain errors to status codes; it holds no business logic.
package httpapi

import (
	"time"

	"folio/folio-core/domain"
)

// The view types are the API contract. They are deliberately separate from
// the domain models so a field rename inside the hexagon does not silently
// break every client.

type memberView struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Role   string `json:"role"`
}

type projectView struct {
	ID        string       `json:"id"`
	DomainID  string       `json:"domain_id,omitempty"`
	Slug      string       `json:"slug"`
	Name      string       `json:"name"`
	Descr     string       `json:"descr,omitempty"`
	Archived  bool         `json:"archived"`
	Members   []memberView `json:"members"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

func toProjectView(p domain.Project) projectView {
	members := make([]memberView, 0, len(p.Members))
	for _, m := range p.Members {
		members = append(members, memberView{
			UserID: string(m.UserID), Email: m.Email, Name: m.Name, Role: string(m.Role),
		})
	}
	return projectView{
		ID: string(p.ID), DomainID: string(p.DomainID), Slug: p.Slug, Name: p.Name, Descr: p.Descr,
		Archived: p.Archived, Members: members,
		CreatedAt: rfc3339(p.CreatedAt), UpdatedAt: rfc3339(p.UpdatedAt),
	}
}

type progressView struct {
	Total   int `json:"total"`
	Done    int `json:"done"`
	Percent int `json:"percent"`
}

type planView struct {
	ID        string       `json:"id"`
	ProjectID string       `json:"project_id"`
	TicketID  string       `json:"ticket_id,omitempty"`
	Title     string       `json:"title"`
	Goal      string       `json:"goal,omitempty"`
	Status    string       `json:"status"`
	Tags      []string     `json:"tags"`
	Progress  progressView `json:"progress"`
	CreatedBy string       `json:"created_by,omitempty"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

func toPlanView(p domain.Plan) planView {
	return planView{
		ID: string(p.ID), ProjectID: string(p.ProjectID), TicketID: string(p.TicketID), Title: p.Title, Goal: p.Goal,
		Status: string(p.Status), Tags: orEmpty(p.Tags),
		Progress:  progressView{Total: p.Progress.Total, Done: p.Progress.Done, Percent: p.Progress.Percent()},
		CreatedBy: string(p.CreatedBy),
		CreatedAt: rfc3339(p.CreatedAt), UpdatedAt: rfc3339(p.UpdatedAt),
	}
}

type ticketView struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"project_id"`
	ParentID    string       `json:"parent_id,omitempty"`
	Slug        string       `json:"slug"`
	Title       string       `json:"title"`
	Body        string       `json:"body"`
	Status      string       `json:"status"`
	Priority    string       `json:"priority"`
	Assignee    string       `json:"assignee,omitempty"`
	Tags        []string     `json:"tags"`
	ExternalRef string       `json:"external_ref,omitempty"`
	DependsOn   []string     `json:"depends_on"`
	Wayfinder   string       `json:"wayfinder,omitempty"`
	Blocked     bool         `json:"blocked"`
	Cycle       int          `json:"cycle,omitempty"`
	Phase       string       `json:"phase,omitempty"`
	Progress    progressView `json:"progress"`
	CreatedBy   string       `json:"created_by,omitempty"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

func toTicketView(t domain.Ticket) ticketView {
	return ticketView{
		ID: string(t.ID), ProjectID: string(t.ProjectID), ParentID: string(t.ParentID), Slug: t.Slug,
		Title: t.Title, Body: t.Body, Status: string(t.Status),
		Priority: string(t.Priority), Assignee: string(t.Assignee),
		Tags: orEmpty(t.Tags), ExternalRef: t.ExternalRef,
		DependsOn: fromTicketIDs(t.DependsOn), Wayfinder: string(t.Wayfinder), Blocked: t.Blocked,
		Cycle: t.Cycle, Phase: string(t.Phase),
		Progress:  progressView{Total: t.Progress.Total, Done: t.Progress.Done, Percent: t.Progress.Percent()},
		CreatedBy: string(t.CreatedBy),
		CreatedAt: rfc3339(t.CreatedAt), UpdatedAt: rfc3339(t.UpdatedAt),
	}
}

type cycleView struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	TicketID   string `json:"ticket_id"`
	Ordinal    int    `json:"ordinal"`
	Phase      string `json:"phase"`
	Resolution string `json:"resolution,omitempty"`
	CreatedBy  string `json:"created_by,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	ClosedAt   string `json:"closed_at,omitempty"`
}

func toCycleView(c domain.Cycle) cycleView {
	out := cycleView{
		ID: string(c.ID), ProjectID: string(c.ProjectID), TicketID: string(c.TicketID),
		Ordinal: c.Ordinal, Phase: string(c.Phase), Resolution: c.Resolution,
		CreatedBy: string(c.CreatedBy),
		CreatedAt: rfc3339(c.CreatedAt), UpdatedAt: rfc3339(c.UpdatedAt),
	}
	if c.ClosedAt != nil {
		out.ClosedAt = rfc3339(*c.ClosedAt)
	}
	return out
}

type ticketLogView struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	TicketID  string `json:"ticket_id"`
	CycleID   string `json:"cycle_id,omitempty"`
	Body      string `json:"body"`
	CreatedBy string `json:"created_by,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toTicketLogView(l domain.TicketLog) ticketLogView {
	return ticketLogView{
		ID: string(l.ID), ProjectID: string(l.ProjectID), TicketID: string(l.TicketID),
		CycleID: string(l.CycleID), Body: l.Body, CreatedBy: string(l.CreatedBy),
		CreatedAt: rfc3339(l.CreatedAt), UpdatedAt: rfc3339(l.UpdatedAt),
	}
}

type planLogView struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	PlanID    string `json:"plan_id"`
	Body      string `json:"body"`
	CreatedBy string `json:"created_by,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toPlanLogView(l domain.PlanLog) planLogView {
	return planLogView{
		ID: string(l.ID), ProjectID: string(l.ProjectID), PlanID: string(l.PlanID),
		Body: l.Body, CreatedBy: string(l.CreatedBy),
		CreatedAt: rfc3339(l.CreatedAt), UpdatedAt: rfc3339(l.UpdatedAt),
	}
}

type todoLogView struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	TodoID    string `json:"todo_id"`
	Body      string `json:"body"`
	CreatedBy string `json:"created_by,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toTodoLogView(l domain.TodoLog) todoLogView {
	return todoLogView{
		ID: string(l.ID), ProjectID: string(l.ProjectID), TodoID: string(l.TodoID),
		Body: l.Body, CreatedBy: string(l.CreatedBy),
		CreatedAt: rfc3339(l.CreatedAt), UpdatedAt: rfc3339(l.UpdatedAt),
	}
}

type todoView struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	TicketID  string   `json:"ticket_id,omitempty"`
	PlanID    string   `json:"plan_id,omitempty"`
	Title     string   `json:"title"`
	Details   string   `json:"details,omitempty"`
	Status    string   `json:"status"`
	Priority  string   `json:"priority"`
	Tags      []string `json:"tags"`
	Position  int      `json:"position"`
	DependsOn []string `json:"depends_on"`
	DueDate   string   `json:"due_date,omitempty"`
	// Blocked is derived from the dependencies' statuses, not stored.
	Blocked   bool   `json:"blocked"`
	CreatedBy string `json:"created_by,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toTodoView(t domain.Todo) todoView {
	deps := make([]string, 0, len(t.DependsOn))
	for _, d := range t.DependsOn {
		deps = append(deps, string(d))
	}
	v := todoView{
		ID: string(t.ID), ProjectID: string(t.ProjectID), TicketID: string(t.TicketID), PlanID: string(t.PlanID),
		Title: t.Title, Details: t.Details, Status: string(t.Status), Priority: string(t.Priority),
		Tags: orEmpty(t.Tags), Position: t.Position, DependsOn: deps, Blocked: t.Blocked,
		CreatedBy: string(t.CreatedBy),
		CreatedAt: rfc3339(t.CreatedAt), UpdatedAt: rfc3339(t.UpdatedAt),
	}
	if t.DueDate != nil {
		v.DueDate = rfc3339(*t.DueDate)
	}
	return v
}

type journalView struct {
	ID          string         `json:"id"`
	ProjectID   string         `json:"project_id"`
	TicketID    string         `json:"ticket_id,omitempty"`
	PlanID      string         `json:"plan_id,omitempty"`
	TodoID      string         `json:"todo_id,omitempty"`
	Slug        string         `json:"slug"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	Branch      string         `json:"branch,omitempty"`
	PR          string         `json:"pr,omitempty"`
	ExternalRef string         `json:"external_ref,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
	Tags        []string       `json:"tags"`
	CreatedBy   string         `json:"created_by,omitempty"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

func toJournalView(e domain.JournalEntry) journalView {
	return journalView{
		ID: string(e.ID), ProjectID: string(e.ProjectID), TicketID: string(e.TicketID),
		PlanID: string(e.PlanID), TodoID: string(e.TodoID), Slug: e.Slug, Title: e.Title, Body: e.Body,
		Branch: e.Branch, PR: e.PR, ExternalRef: e.ExternalRef,
		Meta: e.Meta, Tags: orEmpty(e.Tags), CreatedBy: string(e.CreatedBy),
		CreatedAt: rfc3339(e.CreatedAt), UpdatedAt: rfc3339(e.UpdatedAt),
	}
}

type docView struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	TicketID  string   `json:"ticket_id,omitempty"`
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Tags      []string `json:"tags"`
	CreatedBy string   `json:"created_by,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func toDocView(d domain.Doc) docView {
	return docView{
		ID: string(d.ID), ProjectID: string(d.ProjectID), TicketID: string(d.TicketID), Slug: d.Slug, Title: d.Title,
		Body: d.Body, Tags: orEmpty(d.Tags), CreatedBy: string(d.CreatedBy),
		CreatedAt: rfc3339(d.CreatedAt), UpdatedAt: rfc3339(d.UpdatedAt),
	}
}

type searchHitView struct {
	Kind      string   `json:"kind"`
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	Title     string   `json:"title"`
	Snippet   string   `json:"snippet,omitempty"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
}

func toSearchHitView(h domain.SearchHit) searchHitView {
	return searchHitView{
		Kind: string(h.Kind), ID: h.ID, ProjectID: string(h.ProjectID),
		Title: h.Title, Snippet: h.Snippet, Tags: orEmpty(h.Tags),
		CreatedAt: rfc3339(h.CreatedAt),
	}
}

// batchErrorView reports one rejected item of a batch write.
type batchErrorView struct {
	Index  int    `json:"index"`
	Title  string `json:"title,omitempty"`
	Reason string `json:"reason"`
}

func rfc3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// orEmpty keeps JSON arrays as [] rather than null, so clients can iterate
// without a nil check.
func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

type ticketBriefView struct {
	Ticket  ticketView    `json:"ticket"`
	Plans   []planView    `json:"plans"`
	Todos   []todoView    `json:"todos"`
	Journal []journalView `json:"journal"`
	Docs    []docView     `json:"docs"`
	Cycles  []cycleView   `json:"cycles"`
}

func toTicketBriefView(b domain.TicketBrief) ticketBriefView {
	out := ticketBriefView{
		Ticket:  toTicketView(b.Ticket),
		Plans:   make([]planView, 0, len(b.Plans)),
		Todos:   make([]todoView, 0, len(b.Todos)),
		Journal: make([]journalView, 0, len(b.Journal)),
		Docs:    make([]docView, 0, len(b.Docs)),
		Cycles:  make([]cycleView, 0, len(b.Cycles)),
	}
	for _, p := range b.Plans {
		out.Plans = append(out.Plans, toPlanView(p))
	}
	for _, t := range b.Todos {
		out.Todos = append(out.Todos, toTodoView(t))
	}
	for _, e := range b.Journal {
		out.Journal = append(out.Journal, toJournalView(e))
	}
	for _, d := range b.Docs {
		out.Docs = append(out.Docs, toDocView(d))
	}
	for _, c := range b.Cycles {
		out.Cycles = append(out.Cycles, toCycleView(c))
	}
	return out
}

type issueView struct {
	ID          string       `json:"id"`
	Kind        string       `json:"kind"`
	ProjectID   string       `json:"project_id"`
	ParentID    string       `json:"parent_id,omitempty"`
	PlanID      string       `json:"plan_id,omitempty"`
	Slug        string       `json:"slug"`
	Title       string       `json:"title"`
	Body        string       `json:"body"`
	Status      string       `json:"status"`
	Priority    string       `json:"priority"`
	Size        int          `json:"size,omitempty"`
	Score       float64      `json:"score"`
	Assignee    string       `json:"assignee,omitempty"`
	Tags        []string     `json:"tags"`
	Position    int          `json:"position,omitempty"`
	DueDate     string       `json:"due_date,omitempty"`
	ExternalRef string       `json:"external_ref,omitempty"`
	DependsOn   []string     `json:"depends_on"`
	RelatedTo   []string     `json:"related_to"`
	Wayfinder   string       `json:"wayfinder,omitempty"`
	Blocked     bool         `json:"blocked"`
	Cycle       int          `json:"cycle,omitempty"`
	Phase       string       `json:"phase,omitempty"`
	Progress    progressView `json:"progress"`
	CreatedBy   string       `json:"created_by,omitempty"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

func toIssueView(i domain.Issue) issueView {
	due := ""
	if i.DueDate != nil {
		due = rfc3339(*i.DueDate)
	}
	return issueView{
		ID: string(i.ID), Kind: string(i.Kind), ProjectID: string(i.ProjectID),
		ParentID: string(i.ParentID), PlanID: string(i.PlanID), Slug: i.Slug,
		Title: i.Title, Body: i.Body, Status: string(i.Status),
		Priority: string(i.Priority), Size: int(i.Size), Score: i.Score(),
		Assignee: string(i.Assignee), Tags: orEmpty(i.Tags),
		Position: i.Position, DueDate: due, ExternalRef: i.ExternalRef,
		DependsOn: fromIssueIDs(i.DependsOn), RelatedTo: fromIssueIDs(i.RelatedTo),
		Wayfinder: string(i.Wayfinder), Blocked: i.Blocked,
		Cycle: i.Cycle, Phase: string(i.Phase),
		Progress:  progressView{Total: i.Progress.Total, Done: i.Progress.Done, Percent: i.Progress.Percent()},
		CreatedBy: string(i.CreatedBy),
		CreatedAt: rfc3339(i.CreatedAt), UpdatedAt: rfc3339(i.UpdatedAt),
	}
}

func fromIssueIDs(ids []domain.IssueID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

func toIssueIDs(raw []string) []domain.IssueID {
	out := make([]domain.IssueID, 0, len(raw))
	for _, s := range raw {
		out = append(out, domain.IssueID(s))
	}
	return out
}

type issueBriefView struct {
	Issue    issueView      `json:"issue"`
	Children []issueView    `json:"children"`
	Plans    []planView     `json:"plans"`
	Journal  []journalView  `json:"journal"`
	Docs     []docView      `json:"docs"`
	Cycles   []cycleView    `json:"cycles"`
}

func toIssueBriefView(b domain.IssueBrief) issueBriefView {
	children := make([]issueView, 0, len(b.Children))
	for _, c := range b.Children {
		children = append(children, toIssueView(c))
	}
	plans := make([]planView, 0, len(b.Plans))
	for _, p := range b.Plans {
		plans = append(plans, toPlanView(p))
	}
	journal := make([]journalView, 0, len(b.Journal))
	for _, j := range b.Journal {
		journal = append(journal, toJournalView(j))
	}
	docs := make([]docView, 0, len(b.Docs))
	for _, d := range b.Docs {
		docs = append(docs, toDocView(d))
	}
	cycles := make([]cycleView, 0, len(b.Cycles))
	for _, c := range b.Cycles {
		cycles = append(cycles, toCycleView(c))
	}
	return issueBriefView{
		Issue: toIssueView(b.Issue), Children: children, Plans: plans,
		Journal: journal, Docs: docs, Cycles: cycles,
	}
}
