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

type todoBody struct {
	TicketID  *string   `json:"ticket_id"`
	PlanID    *string   `json:"plan_id"`
	Title     *string   `json:"title"`
	Details   *string   `json:"details"`
	Status    *string   `json:"status"`
	Priority  *string   `json:"priority"`
	Tags      *[]string `json:"tags"`
	Position  *int      `json:"position"`
	DependsOn *[]string `json:"depends_on"`
	DueDate   *string   `json:"due_date"`
}

// toCreateInput reads the body as a creation request, where absent fields
// simply take the service's defaults.
func (b todoBody) toCreateInput() ports.CreateTodoInput {
	in := ports.CreateTodoInput{DueDate: b.DueDate}
	if b.TicketID != nil {
		in.TicketID = domain.TicketID(*b.TicketID)
	}
	if b.PlanID != nil {
		in.PlanID = domain.PlanID(*b.PlanID)
	}
	if b.Title != nil {
		in.Title = *b.Title
	}
	if b.Details != nil {
		in.Details = *b.Details
	}
	if b.Status != nil {
		in.Status = domain.TodoStatus(*b.Status)
	}
	if b.Priority != nil {
		in.Priority = domain.Priority(*b.Priority)
	}
	if b.Tags != nil {
		in.Tags = *b.Tags
	}
	if b.DependsOn != nil {
		in.DependsOn = toTodoIDs(*b.DependsOn)
	}
	return in
}

// toUpdateInput reads the body as a PATCH, where an absent field means
// "leave it alone" rather than "clear it".
func (b todoBody) toUpdateInput() ports.UpdateTodoInput {
	in := ports.UpdateTodoInput{
		Title: b.Title, Details: b.Details, Tags: b.Tags,
		Position: b.Position, DueDate: b.DueDate,
	}
	if b.TicketID != nil {
		id := domain.TicketID(*b.TicketID)
		in.TicketID = &id
	}
	if b.PlanID != nil {
		id := domain.PlanID(*b.PlanID)
		in.PlanID = &id
	}
	if b.Status != nil {
		s := domain.TodoStatus(*b.Status)
		in.Status = &s
	}
	if b.Priority != nil {
		p := domain.Priority(*b.Priority)
		in.Priority = &p
	}
	if b.DependsOn != nil {
		deps := toTodoIDs(*b.DependsOn)
		in.DependsOn = &deps
	}
	return in
}

func toTodoIDs(raw []string) []domain.TodoID {
	out := make([]domain.TodoID, 0, len(raw))
	for _, s := range raw {
		out = append(out, domain.TodoID(s))
	}
	return out
}

type createPlanBody struct {
	TicketID string     `json:"ticket_id"`
	Title    string     `json:"title"`
	Goal     string     `json:"goal"`
	Status   string     `json:"status"`
	Tags     []string   `json:"tags"`
	Todos    []todoBody `json:"todos"`
}

// createPlan accepts the plan and its first todos together: an agent drafting
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

	todos := make([]ports.CreateTodoInput, 0, len(body.Todos))
	for _, t := range body.Todos {
		todos = append(todos, t.toCreateInput())
	}

	plan, err := h.plans.CreatePlan(e.Request.Context(), actorOf(e), ports.CreatePlanInput{
		ProjectID: project, TicketID: domain.TicketID(body.TicketID),
		Title: body.Title, Goal: body.Goal,
		Status: domain.PlanStatus(body.Status), Tags: body.Tags, Todos: todos,
	})
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toPlanView(plan))
}

type updatePlanBody struct {
	TicketID *string   `json:"ticket_id"`
	Title    *string   `json:"title"`
	Goal     *string   `json:"goal"`
	Status   *string   `json:"status"`
	Tags     *[]string `json:"tags"`
}

func (h *Handler) updatePlan(e *core.RequestEvent) error {
	var body updatePlanBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	in := ports.UpdatePlanInput{Title: body.Title, Goal: body.Goal, Tags: body.Tags}
	if body.TicketID != nil {
		id := domain.TicketID(*body.TicketID)
		in.TicketID = &id
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

func (h *Handler) listTodos(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	filter := domain.TodoFilter{
		PlanID:   domain.PlanID(e.Request.URL.Query().Get("plan_id")),
		TicketID: domain.TicketID(e.Request.URL.Query().Get("ticket_id")),
		Priority: domain.Priority(e.Request.URL.Query().Get("priority")),
		Tags:     csv(e, "tags"),
		Search:   e.Request.URL.Query().Get("q"),
		Limit:    queryInt(e, "limit"),
		Offset:   queryInt(e, "offset"),
	}
	for _, s := range csv(e, "status") {
		filter.Status = append(filter.Status, domain.TodoStatus(s))
	}

	todos, err := h.todos.ListTodos(e.Request.Context(), actorOf(e), project, filter)
	if err != nil {
		return fail(e, err)
	}
	out := make([]todoView, 0, len(todos))
	for _, t := range todos {
		out = append(out, toTodoView(t))
	}
	return e.JSON(http.StatusOK, map[string]any{"todos": out})
}

func (h *Handler) getTodo(e *core.RequestEvent) error {
	todo, err := h.todos.GetTodo(e.Request.Context(), actorOf(e), domain.TodoID(e.Request.PathValue("todo")))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toTodoView(todo))
}

// createTodosBody accepts either a single todo or a batch, so an agent
// triaging several steps does not need one request per step.
type createTodosBody struct {
	todoBody
	Items []todoBody `json:"items"`
}

func (h *Handler) createTodos(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	var body createTodosBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	if len(body.Items) == 0 {
		todo, err := h.todos.CreateTodo(e.Request.Context(), actorOf(e), withProject(body.todoBody.toCreateInput(), project))
		if err != nil {
			return fail(e, err)
		}
		return e.JSON(http.StatusCreated, toTodoView(todo))
	}

	items := make([]ports.CreateTodoInput, 0, len(body.Items))
	for _, item := range body.Items {
		items = append(items, withProject(item.toCreateInput(), project))
	}
	res, err := h.todos.CreateTodos(e.Request.Context(), actorOf(e), project, items)
	if err != nil {
		return fail(e, err)
	}

	created := make([]todoView, 0, len(res.Created))
	for _, t := range res.Created {
		created = append(created, toTodoView(t))
	}
	failures := make([]batchErrorView, 0, len(res.Errors))
	for _, be := range res.Errors {
		failures = append(failures, batchErrorView{Index: be.Index, Title: be.Title, Reason: be.Reason})
	}
	// 207 because a batch can partly succeed; the body says which items did.
	status := http.StatusCreated
	if len(failures) > 0 {
		status = http.StatusMultiStatus
	}
	return e.JSON(status, map[string]any{"created": created, "errors": failures})
}

func withProject(in ports.CreateTodoInput, project domain.ProjectID) ports.CreateTodoInput {
	in.ProjectID = project
	return in
}

func (h *Handler) updateTodo(e *core.RequestEvent) error {
	var body todoBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	todo, err := h.todos.UpdateTodo(e.Request.Context(), actorOf(e), domain.TodoID(e.Request.PathValue("todo")), body.toUpdateInput())
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toTodoView(todo))
}

func (h *Handler) deleteTodo(e *core.RequestEvent) error {
	if err := h.todos.DeleteTodo(e.Request.Context(), actorOf(e), domain.TodoID(e.Request.PathValue("todo"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}
