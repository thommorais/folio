package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func (h *Handler) listPlans(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	statuses := make([]domain.PlanStatus, 0)
	for _, s := range csv(e, "status") {
		statuses = append(statuses, domain.PlanStatus(s))
	}

	plans, err := h.plans.ListPlans(e.Request.Context(), actorOf(e), project, statuses)
	if err != nil {
		return fail(e, err)
	}
	out := make([]planView, 0, len(plans))
	for _, p := range plans {
		out = append(out, toPlanView(p))
	}
	return e.JSON(http.StatusOK, map[string]any{"plans": out})
}

func (h *Handler) getPlan(e *core.RequestEvent) error {
	plan, err := h.plans.GetPlan(e.Request.Context(), actorOf(e), domain.PlanID(e.Request.PathValue("plan")))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toPlanView(plan))
}

type createPlanBody struct {
	IssueID string      `json:"issue_id"`
	Title   string      `json:"title"`
	Goal    string      `json:"goal"`
	Status  string      `json:"status"`
	Tags    []string    `json:"tags"`
	Issues  []issueBody `json:"issues"`
}

type updatePlanBody struct {
	IssueID *string   `json:"issue_id"`
	Title   *string   `json:"title"`
	Goal    *string   `json:"goal"`
	Status  *string   `json:"status"`
	Tags    *[]string `json:"tags"`
}

// createPlan accepts the plan and its first issues together: an agent drafting
// a plan knows the steps at the same moment.
func (h *Handler) createPlan(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	var body createPlanBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	issues := make([]ports.CreateIssueInput, 0, len(body.Issues))
	for _, t := range body.Issues {
		issues = append(issues, t.toCreate(project))
	}

	plan, err := h.plans.CreatePlan(e.Request.Context(), actorOf(e), ports.CreatePlanInput{
		ProjectID: project, IssueID: domain.IssueID(body.IssueID),
		Title: body.Title, Goal: body.Goal,
		Status: domain.PlanStatus(body.Status), Tags: body.Tags, Todos: issues,
	})
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toPlanView(plan))
}

func (h *Handler) updatePlan(e *core.RequestEvent) error {
	var body updatePlanBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	in := ports.UpdatePlanInput{Title: body.Title, Goal: body.Goal, Tags: body.Tags}
	if body.IssueID != nil {
		id := domain.IssueID(*body.IssueID)
		in.IssueID = &id
	}
	if body.Status != nil {
		s := domain.PlanStatus(*body.Status)
		in.Status = &s
	}

	plan, err := h.plans.UpdatePlan(e.Request.Context(), actorOf(e), domain.PlanID(e.Request.PathValue("plan")), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toPlanView(plan))
}

func (h *Handler) deletePlan(e *core.RequestEvent) error {
	if err := h.plans.DeletePlan(e.Request.Context(), actorOf(e), domain.PlanID(e.Request.PathValue("plan"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}
