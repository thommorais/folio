package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func (h *Handler) listIssues(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	filter := domain.IssueFilter{
		Kind:     domain.IssueKind(e.Request.URL.Query().Get("kind")),
		ParentID: domain.IssueID(e.Request.URL.Query().Get("parent")),
		PlanID:   domain.PlanID(e.Request.URL.Query().Get("plan")),
		Priority: domain.Priority(e.Request.URL.Query().Get("priority")),
		Assignee: domain.UserID(e.Request.URL.Query().Get("assignee")),
		Tags:     csv(e, "tags"),
		Search:   e.Request.URL.Query().Get("q"),
		Limit:    queryInt(e, "limit"),
		Offset:   queryInt(e, "offset"),
	}
	for _, s := range csv(e, "status") {
		filter.Status = append(filter.Status, domain.IssueStatus(s))
	}

	issues, err := h.issues.ListIssues(e.Request.Context(), actorOf(e), project, filter)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, map[string]any{"issues": issueViews(issues)})
}

func issueViews(issues []domain.Issue) []issueView {
	out := make([]issueView, 0, len(issues))
	for _, i := range issues {
		out = append(out, toIssueView(i))
	}
	return out
}

func (h *Handler) getIssue(e *core.RequestEvent) error {
	issue, err := h.issues.GetIssue(e.Request.Context(), actorOf(e), domain.IssueID(e.Request.PathValue("issue")))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toIssueView(issue))
}

func (h *Handler) getIssueBySlug(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	issue, err := h.issues.GetIssueBySlug(e.Request.Context(), actorOf(e), project, e.Request.PathValue("slug"))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toIssueView(issue))
}

func (h *Handler) getIssueBrief(e *core.RequestEvent) error {
	in := ports.BriefOptions{RecentJournal: queryInt(e, "recent_journal")}
	brief, err := h.issues.GetIssueBrief(e.Request.Context(), actorOf(e), domain.IssueID(e.Request.PathValue("issue")), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toIssueBriefView(brief))
}

func (h *Handler) getIssueBriefBySlug(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	in := ports.BriefOptions{RecentJournal: queryInt(e, "recent_journal")}
	brief, err := h.issues.GetIssueBriefBySlug(e.Request.Context(), actorOf(e), project, e.Request.PathValue("slug"), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toIssueBriefView(brief))
}

func (h *Handler) issueFrontier(e *core.RequestEvent) error {
	issues, err := h.issues.Frontier(e.Request.Context(), actorOf(e), domain.IssueID(e.Request.PathValue("issue")))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, map[string]any{"issues": issueViews(issues)})
}

type issueBody struct {
	Kind            *string   `json:"kind"`
	ParentID        *string   `json:"parent_id"`
	PlanID          *string   `json:"plan_id"`
	Slug            *string   `json:"slug"`
	Title           *string   `json:"title"`
	Body            *string   `json:"body"`
	Status          *string   `json:"status"`
	Priority        *string   `json:"priority"`
	Size            *int      `json:"size"`
	Assignee        *string   `json:"assignee"`
	Tags            *[]string `json:"tags"`
	Position        *int      `json:"position"`
	DueDate         *string   `json:"due_date"`
	DependsOn       *[]string `json:"depends_on"`
	Wayfinder       *string   `json:"wayfinder"`
	ExternalRef     *string   `json:"external_ref"`
	Resolution      *string   `json:"resolution"`
	ResolutionEntry *string   `json:"resolution_entry_id"`
}

func (b issueBody) toCreate(project domain.ProjectID) ports.CreateIssueInput {
	in := ports.CreateIssueInput{ProjectID: project, DueDate: b.DueDate}
	if b.Kind != nil {
		in.Kind = domain.IssueKind(*b.Kind)
	}
	if b.ParentID != nil {
		in.ParentID = domain.IssueID(*b.ParentID)
	}
	if b.PlanID != nil {
		in.PlanID = domain.PlanID(*b.PlanID)
	}
	if b.Slug != nil {
		in.Slug = *b.Slug
	}
	if b.Title != nil {
		in.Title = *b.Title
	}
	if b.Body != nil {
		in.Body = *b.Body
	}
	if b.Status != nil {
		in.Status = domain.IssueStatus(*b.Status)
	}
	if b.Priority != nil {
		in.Priority = domain.Priority(*b.Priority)
	}
	if b.Size != nil {
		in.Size = domain.Size(*b.Size)
	}
	if b.Assignee != nil {
		in.Assignee = domain.UserID(*b.Assignee)
	}
	if b.Tags != nil {
		in.Tags = *b.Tags
	}
	if b.DependsOn != nil {
		in.DependsOn = toIssueIDs(*b.DependsOn)
	}
	if b.Wayfinder != nil {
		in.Wayfinder = domain.WayfinderType(*b.Wayfinder)
	}
	if b.ExternalRef != nil {
		in.ExternalRef = *b.ExternalRef
	}
	if b.Resolution != nil {
		in.Resolution = *b.Resolution
	}
	return in
}

type createIssuesBody struct {
	Issues []issueBody `json:"issues"`
}

func (h *Handler) createIssues(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}

	var batch createIssuesBody
	if err := e.BindBody(&batch); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	// A single object and a batch share the route: an agent filing one issue
	// should not have to wrap it.
	if len(batch.Issues) == 0 {
		var single issueBody
		if err := e.BindBody(&single); err != nil {
			return e.BadRequestError("invalid request body", err)
		}
		if single.Title == nil {
			return e.BadRequestError("title is required", nil)
		}
		issue, err := h.issues.CreateIssue(e.Request.Context(), actorOf(e), single.toCreate(project))
		if err != nil {
			return fail(e, err)
		}
		return e.JSON(http.StatusCreated, toIssueView(issue))
	}

	in := make([]ports.CreateIssueInput, 0, len(batch.Issues))
	for _, item := range batch.Issues {
		in = append(in, item.toCreate(project))
	}
	result, err := h.issues.CreateIssues(e.Request.Context(), actorOf(e), project, in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, map[string]any{
		"issues": issueViews(result.Created),
		"errors": result.Errors,
	})
}

func (h *Handler) updateIssue(e *core.RequestEvent) error {
	var body issueBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	in := ports.UpdateIssueInput{
		Slug: body.Slug, Title: body.Title, Body: body.Body,
		Tags: body.Tags, Position: body.Position, DueDate: body.DueDate,
		ExternalRef: body.ExternalRef, Resolution: body.Resolution,
	}
	if body.Kind != nil {
		kind := domain.IssueKind(*body.Kind)
		in.Kind = &kind
	}
	if body.ParentID != nil {
		parent := domain.IssueID(*body.ParentID)
		in.ParentID = &parent
	}
	if body.PlanID != nil {
		plan := domain.PlanID(*body.PlanID)
		in.PlanID = &plan
	}
	if body.Status != nil {
		status := domain.IssueStatus(*body.Status)
		in.Status = &status
	}
	if body.Priority != nil {
		priority := domain.Priority(*body.Priority)
		in.Priority = &priority
	}
	if body.Size != nil {
		size := domain.Size(*body.Size)
		in.Size = &size
	}
	if body.Assignee != nil {
		assignee := domain.UserID(*body.Assignee)
		in.Assignee = &assignee
	}
	if body.DependsOn != nil {
		deps := toIssueIDs(*body.DependsOn)
		in.DependsOn = &deps
	}
	if body.Wayfinder != nil {
		wayfinder := domain.WayfinderType(*body.Wayfinder)
		in.Wayfinder = &wayfinder
	}
	if body.ResolutionEntry != nil {
		entry := domain.EntryID(*body.ResolutionEntry)
		in.ResolutionEntry = &entry
	}

	issue, err := h.issues.UpdateIssue(e.Request.Context(), actorOf(e), domain.IssueID(e.Request.PathValue("issue")), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toIssueView(issue))
}

func (h *Handler) deleteIssue(e *core.RequestEvent) error {
	if err := h.issues.DeleteIssue(e.Request.Context(), actorOf(e), domain.IssueID(e.Request.PathValue("issue"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}

type linkBody struct {
	To   string `json:"to"`
	Kind string `json:"kind"`
}

func (h *Handler) linkIssue(e *core.RequestEvent) error {
	var body linkBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	from := domain.IssueID(e.Request.PathValue("issue"))
	if err := h.issues.LinkIssues(e.Request.Context(), actorOf(e), from, domain.IssueID(body.To), domain.LinkKind(body.Kind)); err != nil {
		return fail(e, err)
	}
	issue, err := h.issues.GetIssue(e.Request.Context(), actorOf(e), from)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toIssueView(issue))
}

func (h *Handler) unlinkIssue(e *core.RequestEvent) error {
	from := domain.IssueID(e.Request.PathValue("issue"))
	to := domain.IssueID(e.Request.PathValue("to"))
	kind := domain.LinkKind(e.Request.URL.Query().Get("kind"))
	if kind == "" {
		kind = domain.LinkBlocks
	}
	if err := h.issues.UnlinkIssues(e.Request.Context(), actorOf(e), from, to, kind); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}
