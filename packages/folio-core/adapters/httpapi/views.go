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
	IssueID   string       `json:"issue_id,omitempty"`
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
		ID: string(p.ID), ProjectID: string(p.ProjectID), IssueID: string(p.IssueID), Title: p.Title, Goal: p.Goal,
		Status: string(p.Status), Tags: orEmpty(p.Tags),
		Progress:  progressView{Total: p.Progress.Total, Done: p.Progress.Done, Percent: p.Progress.Percent()},
		CreatedBy: string(p.CreatedBy),
		CreatedAt: rfc3339(p.CreatedAt), UpdatedAt: rfc3339(p.UpdatedAt),
	}
}

type cycleView struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	IssueID    string `json:"issue_id,omitempty"`
	Ordinal    int    `json:"ordinal"`
	Phase      string `json:"phase"`
	Resolution string `json:"resolution,omitempty"`
	MapID      string `json:"map_id,omitempty"`
	CreatedBy  string `json:"created_by,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	ClosedAt   string `json:"closed_at,omitempty"`
}

func toCycleView(c domain.Cycle) cycleView {
	out := cycleView{
		ID: string(c.ID), ProjectID: string(c.ProjectID), IssueID: string(c.IssueID),
		Ordinal: c.Ordinal, Phase: string(c.Phase), Resolution: c.Resolution, MapID: string(c.MapID),
		CreatedBy: string(c.CreatedBy),
		CreatedAt: rfc3339(c.CreatedAt), UpdatedAt: rfc3339(c.UpdatedAt),
	}
	if c.ClosedAt != nil {
		out.ClosedAt = rfc3339(*c.ClosedAt)
	}
	return out
}

type searchHitView struct {
	Kind        string   `json:"kind"`
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id"`
	ProjectSlug string   `json:"project_slug"`
	DomainSlug  string   `json:"domain_slug"`
	ClientSlug  string   `json:"client_slug"`
	Slug        string   `json:"slug,omitempty"`
	Title       string   `json:"title"`
	Snippet     string   `json:"snippet,omitempty"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"created_at"`
}

func toSearchHitView(h domain.SearchHit) searchHitView {
	return searchHitView{
		Kind: string(h.Kind), ID: h.ID, ProjectID: string(h.ProjectID),
		ProjectSlug: h.ProjectSlug, DomainSlug: h.DomainSlug, ClientSlug: h.ClientSlug, Slug: h.Slug,
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

type issueView struct {
	ID              string       `json:"id"`
	Kind            string       `json:"kind"`
	ProjectID       string       `json:"project_id"`
	ParentID        string       `json:"parent_id,omitempty"`
	PlanID          string       `json:"plan_id,omitempty"`
	Slug            string       `json:"slug"`
	Title           string       `json:"title"`
	Body            string       `json:"body"`
	Status          string       `json:"status"`
	Priority        string       `json:"priority"`
	Size            int          `json:"size,omitempty"`
	Score           float64      `json:"score"`
	Assignee        string       `json:"assignee,omitempty"`
	Tags            []string     `json:"tags"`
	Position        int          `json:"position,omitempty"`
	DueDate         string       `json:"due_date,omitempty"`
	ExternalRef     string       `json:"external_ref,omitempty"`
	DependsOn       []string     `json:"depends_on"`
	RelatedTo       []string     `json:"related_to"`
	Wayfinder       string       `json:"wayfinder,omitempty"`
	Resolution      string       `json:"resolution,omitempty"`
	ResolutionEntry string       `json:"resolution_entry_id,omitempty"`
	Blocked         bool         `json:"blocked"`
	Cycle           int          `json:"cycle,omitempty"`
	Phase           string       `json:"phase,omitempty"`
	Progress        progressView `json:"progress"`
	CreatedBy       string       `json:"created_by,omitempty"`
	CreatedAt       string       `json:"created_at"`
	UpdatedAt       string       `json:"updated_at"`
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
		Resolution: i.Resolution, ResolutionEntry: string(i.ResolutionEntry),
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
	Issue    issueView     `json:"issue"`
	Children []issueView   `json:"children"`
	Plans    []planView    `json:"plans"`
	Journal  []entryView   `json:"journal"`
	Docs     []entryView   `json:"docs"`
	Cycles   []cycleView   `json:"cycles"`
	Map      *mapBriefView `json:"map,omitempty"`
}

type mapBriefView struct {
	Issue    issueView   `json:"issue"`
	Open     int         `json:"open"`
	Frontier []issueView `json:"frontier"`
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
	journal := make([]entryView, 0, len(b.Journal))
	for _, j := range b.Journal {
		journal = append(journal, toEntryView(j))
	}
	docs := make([]entryView, 0, len(b.Docs))
	for _, d := range b.Docs {
		docs = append(docs, toEntryView(d))
	}
	cycles := make([]cycleView, 0, len(b.Cycles))
	for _, c := range b.Cycles {
		cycles = append(cycles, toCycleView(c))
	}
	out := issueBriefView{
		Issue: toIssueView(b.Issue), Children: children, Plans: plans,
		Journal: journal, Docs: docs, Cycles: cycles,
	}
	if b.Map != nil {
		out.Map = &mapBriefView{Issue: toIssueView(b.Map.Map), Open: b.Map.Open, Frontier: issueViews(b.Map.Frontier)}
	}
	return out
}

type entryView struct {
	ID          string         `json:"id"`
	Kind        string         `json:"kind"`
	ProjectID   string         `json:"project_id"`
	IssueID     string         `json:"issue_id,omitempty"`
	PlanID      string         `json:"plan_id,omitempty"`
	CycleID     string         `json:"cycle_id,omitempty"`
	Slug        string         `json:"slug,omitempty"`
	Title       string         `json:"title,omitempty"`
	Body        string         `json:"body"`
	Branch      string         `json:"branch,omitempty"`
	PR          string         `json:"pr,omitempty"`
	ExternalRef string         `json:"external_ref,omitempty"`
	Tags        []string       `json:"tags"`
	Meta        map[string]any `json:"meta,omitempty"`
	CreatedBy   string         `json:"created_by,omitempty"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

func toEntryView(e domain.Entry) entryView {
	return entryView{
		ID: string(e.ID), Kind: string(e.Kind), ProjectID: string(e.ProjectID),
		IssueID: string(e.IssueID), PlanID: string(e.PlanID), CycleID: string(e.CycleID),
		Slug: e.Slug, Title: e.Title, Body: e.Body,
		Branch: e.Branch, PR: e.PR, ExternalRef: e.ExternalRef,
		Tags: orEmpty(e.Tags), Meta: e.Meta, CreatedBy: string(e.CreatedBy),
		CreatedAt: rfc3339(e.CreatedAt), UpdatedAt: rfc3339(e.UpdatedAt),
	}
}

type knowledgeView struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	// ProjectID is omitted when the note belongs to no project, which is the
	// common case rather than an error.
	ProjectID string   `json:"project_id,omitempty"`
	Tags      []string `json:"tags"`
	CreatedBy string   `json:"created_by"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func toKnowledgeView(k domain.Knowledge) knowledgeView {
	return knowledgeView{
		ID: string(k.ID), Slug: k.Slug, Title: k.Title, Body: k.Body,
		ProjectID: string(k.ProjectID), Tags: orEmpty(k.Tags),
		CreatedBy: string(k.CreatedBy),
		CreatedAt: rfc3339(k.CreatedAt), UpdatedAt: rfc3339(k.UpdatedAt),
	}
}
