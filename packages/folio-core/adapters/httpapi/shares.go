package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
)

type shareView struct {
	ID             string `json:"id"`
	ProjectID      string `json:"project_id"`
	IssueID        string `json:"issue_id,omitempty"`
	PlanID         string `json:"plan_id,omitempty"`
	Label          string `json:"label"`
	Token          string `json:"token"`
	CreatedBy      string `json:"created_by,omitempty"`
	CreatedAt      string `json:"created_at"`
	LastAccessedAt string `json:"last_accessed_at,omitempty"`
}

func toShareView(s domain.Share) shareView {
	out := shareView{
		ID: string(s.ID), ProjectID: string(s.ProjectID), IssueID: string(s.IssueID), PlanID: string(s.PlanID),
		Label: s.Label, Token: s.Token, CreatedBy: string(s.CreatedBy), CreatedAt: rfc3339(s.CreatedAt),
	}
	if s.LastAccessedAt != nil {
		out.LastAccessedAt = rfc3339(*s.LastAccessedAt)
	}
	return out
}

type sharedView struct {
	Kind  string          `json:"kind"`
	Label string          `json:"label"`
	Brief *issueBriefView `json:"brief,omitempty"`
	Plan  *planView       `json:"plan,omitempty"`
	Todos []issueView     `json:"todos,omitempty"`
}

func toSharedView(item domain.SharedItem) sharedView {
	out := sharedView{Label: item.Share.Label}
	if item.Brief != nil {
		b := *item.Brief
		b.Issue = anonIssue(b.Issue)
		b.Children = anonIssues(b.Children)
		b.Plans = anonPlans(b.Plans)
		b.Journal = anonEntries(b.Journal)
		b.Docs = anonEntries(b.Docs)
		cycles := make([]domain.Cycle, len(b.Cycles))
		for i, c := range b.Cycles {
			c.ProjectID, c.CreatedBy = "", ""
			cycles[i] = c
		}
		b.Cycles = cycles
		view := toIssueBriefView(b)
		out.Kind, out.Brief = "issue", &view
		return out
	}
	plan := toPlanView(anonPlans([]domain.Plan{*item.Plan})[0])
	todos := make([]issueView, 0, len(item.Todos))
	for _, t := range anonIssues(item.Todos) {
		todos = append(todos, toIssueView(t))
	}
	out.Kind, out.Plan, out.Todos = "plan", &plan, todos
	return out
}

func anonIssue(i domain.Issue) domain.Issue {
	i.ProjectID, i.CreatedBy, i.Assignee = "", "", ""
	return i
}

func anonIssues(in []domain.Issue) []domain.Issue {
	out := make([]domain.Issue, len(in))
	for i, v := range in {
		out[i] = anonIssue(v)
	}
	return out
}

func anonPlans(in []domain.Plan) []domain.Plan {
	out := make([]domain.Plan, len(in))
	for i, p := range in {
		p.ProjectID, p.CreatedBy = "", ""
		out[i] = p
	}
	return out
}

func anonEntries(in []domain.Entry) []domain.Entry {
	out := make([]domain.Entry, len(in))
	for i, e := range in {
		e.ProjectID, e.CreatedBy = "", ""
		out[i] = e
	}
	return out
}

func (h *Handler) openShare(e *core.RequestEvent) error {
	item, err := h.shares.OpenShare(e.Request.Context(), e.Request.PathValue("token"))
	if err != nil {
		return fail(e, err)
	}
	e.Response.Header().Set("Cache-Control", "no-store")
	return e.JSON(http.StatusOK, toSharedView(item))
}

func (h *Handler) listShares(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	shares, err := h.shares.ListShares(e.Request.Context(), actorOf(e), project)
	if err != nil {
		return fail(e, err)
	}
	out := make([]shareView, 0, len(shares))
	for _, s := range shares {
		out = append(out, toShareView(s))
	}
	return e.JSON(http.StatusOK, map[string]any{"shares": out})
}

type shareBody struct {
	Label string `json:"label"`
}

func (h *Handler) shareIssue(e *core.RequestEvent) error {
	var body shareBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid body", err)
	}
	share, err := h.shares.ShareIssue(e.Request.Context(), actorOf(e), domain.IssueID(e.Request.PathValue("issue")), body.Label)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toShareView(share))
}

func (h *Handler) sharePlan(e *core.RequestEvent) error {
	var body shareBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid body", err)
	}
	share, err := h.shares.SharePlan(e.Request.Context(), actorOf(e), domain.PlanID(e.Request.PathValue("plan")), body.Label)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toShareView(share))
}

func (h *Handler) revokeShare(e *core.RequestEvent) error {
	if err := h.shares.RevokeShare(e.Request.Context(), actorOf(e), domain.ShareID(e.Request.PathValue("share"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}
