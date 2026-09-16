package httpapi

import (
	"errors"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// Handler wires the folio use cases onto PocketBase's router. It depends only
// on ports, so the same handler works against any adapter set.
type Handler struct {
	projects ports.ProjectUseCase
	plans    ports.PlanUseCase
	tickets  ports.TicketUseCase
	todos    ports.TodoUseCase
	journal  ports.JournalUseCase
	cycles   ports.CycleUseCase
	workLogs ports.WorkLogUseCase
	docs     ports.DocUseCase
	search   ports.SearchUseCase
}

type Deps struct {
	Projects ports.ProjectUseCase
	Plans    ports.PlanUseCase
	Tickets  ports.TicketUseCase
	Todos    ports.TodoUseCase
	Journal  ports.JournalUseCase
	Cycles   ports.CycleUseCase
	WorkLogs ports.WorkLogUseCase
	Docs     ports.DocUseCase
	Search   ports.SearchUseCase
}

func New(d Deps) *Handler {
	return &Handler{
		projects: d.Projects, plans: d.Plans, tickets: d.Tickets, todos: d.Todos,
		journal: d.Journal, cycles: d.Cycles, workLogs: d.WorkLogs, docs: d.Docs, search: d.Search,
	}
}

// BasePath is where the API mounts. PocketBase owns /api/collections and the
// rest of its own surface, so folio takes its own namespace.
const BasePath = "/api/folio"

// Mount registers every route. Auth is PocketBase's: RequireAuth rejects
// anonymous callers before a handler runs, and actorOf reads the record it
// attached to the request.
func (h *Handler) Mount(e *core.ServeEvent) {
	g := e.Router.Group(BasePath)
	g.Bind(apis.RequireAuth())

	g.GET("/projects", h.listProjects)
	g.POST("/projects", h.createProject)
	g.GET("/projects/{project}", h.getProject)
	g.PATCH("/projects/{project}", h.updateProject)
	g.DELETE("/projects/{project}", h.deleteProject)

	g.POST("/projects/{project}/members", h.addMember)
	g.PATCH("/projects/{project}/members/{user}", h.setMemberRole)
	g.DELETE("/projects/{project}/members/{user}", h.removeMember)

	g.GET("/projects/{project}/plans", h.listPlans)
	g.POST("/projects/{project}/plans", h.createPlan)
	g.GET("/plans/{plan}", h.getPlan)
	g.PATCH("/plans/{plan}", h.updatePlan)
	g.DELETE("/plans/{plan}", h.deletePlan)

	g.GET("/projects/{project}/tickets", h.listTickets)
	g.POST("/projects/{project}/tickets", h.createTicket)
	// The slug route is registered before the ID route so a project-scoped
	// slug lookup is not shadowed by it.
	g.GET("/projects/{project}/tickets/{slug}", h.getTicketBySlug)
	g.GET("/projects/{project}/tickets/{slug}/brief", h.getTicketBriefBySlug)
	g.GET("/tickets/{ticket}", h.getTicket)
	g.GET("/tickets/{ticket}/brief", h.getTicketBrief)
	g.GET("/tickets/{ticket}/frontier", h.ticketFrontier)
	g.GET("/tickets/{ticket}/cycles", h.listCycles)
	g.POST("/tickets/{ticket}/cycles", h.openCycle)
	g.PATCH("/cycles/{cycle}", h.updateCycle)
	g.GET("/tickets/{ticket}/logs", h.listTicketLogs)
	g.POST("/tickets/{ticket}/logs", h.writeTicketLog)
	g.DELETE("/ticket-logs/{log}", h.deleteTicketLog)
	g.GET("/plans/{plan}/logs", h.listPlanLogs)
	g.POST("/plans/{plan}/logs", h.writePlanLog)
	g.DELETE("/plan-logs/{log}", h.deletePlanLog)
	g.GET("/todos/{todo}/logs", h.listTodoLogs)
	g.POST("/todos/{todo}/logs", h.writeTodoLog)
	g.DELETE("/todo-logs/{log}", h.deleteTodoLog)
	g.PATCH("/tickets/{ticket}", h.updateTicket)
	g.DELETE("/tickets/{ticket}", h.deleteTicket)

	g.GET("/projects/{project}/todos", h.listTodos)
	g.POST("/projects/{project}/todos", h.createTodos)
	g.GET("/todos/{todo}", h.getTodo)
	g.PATCH("/todos/{todo}", h.updateTodo)
	g.DELETE("/todos/{todo}", h.deleteTodo)

	g.GET("/projects/{project}/journal", h.listJournal)
	g.POST("/projects/{project}/journal", h.writeJournalEntry)
	g.GET("/projects/{project}/journal/{slug}", h.getJournalEntryBySlug)
	g.GET("/journal/{entry}", h.getJournalEntry)
	g.PATCH("/journal/{entry}", h.updateJournalEntry)
	g.POST("/journal/{entry}/append", h.appendJournalEntry)
	g.DELETE("/journal/{entry}", h.deleteJournalEntry)

	g.GET("/projects/{project}/docs", h.listDocs)
	g.POST("/projects/{project}/docs", h.createDoc)
	g.GET("/projects/{project}/docs/{slug}", h.getDocBySlug)
	g.GET("/docs/{doc}", h.getDoc)
	g.PATCH("/docs/{doc}", h.updateDoc)
	g.DELETE("/docs/{doc}", h.deleteDoc)

	g.GET("/projects/{project}/search", h.searchProject)
}

// actorOf builds the domain actor from the record PocketBase authenticated.
func actorOf(e *core.RequestEvent) ports.Actor {
	if e.Auth == nil {
		return ports.Actor{}
	}
	return ports.Actor{
		UserID:    domain.UserID(e.Auth.Id),
		Email:     e.Auth.Email(),
		Superuser: e.HasSuperuserAuth(),
	}
}

// fail maps a domain error to the matching HTTP status. Keeping this in one
// place is what lets the services stay transport-agnostic.
func fail(e *core.RequestEvent, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return e.NotFoundError("not found", nil)
	case errors.Is(err, domain.ErrForbidden):
		return e.ForbiddenError("you do not have access to this resource", nil)
	case errors.Is(err, domain.ErrConflict):
		return e.Error(409, "already exists", nil)
	case errors.Is(err, domain.ErrValidation):
		var v domain.ValidationError
		if errors.As(err, &v) {
			return e.BadRequestError(v.Error(), map[string]any{"field": v.Field, "reason": v.Reason})
		}
		return e.BadRequestError(err.Error(), nil)
	default:
		return e.InternalServerError("request failed", err)
	}
}

// resolveProject turns the {project} path segment, which may be an ID or a
// slug, into an ID the use cases can take.
func (h *Handler) resolveProject(e *core.RequestEvent) (domain.ProjectID, error) {
	ref := e.Request.PathValue("project")
	project, err := h.projects.GetProject(e.Request.Context(), actorOf(e), ref)
	if err != nil {
		return "", err
	}
	return project.ID, nil
}

func queryInt(e *core.RequestEvent, key string) int {
	n, err := strconv.Atoi(e.Request.URL.Query().Get(key))
	if err != nil {
		return 0
	}
	return n
}

// csv splits a repeatable comma-separated query parameter, e.g.
// ?status=pending,done or ?status=pending&status=done.
func csv(e *core.RequestEvent, key string) []string {
	values := e.Request.URL.Query()[key]
	out := make([]string, 0, len(values))
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			if p := strings.TrimSpace(part); p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func queryBool(e *core.RequestEvent, key string) bool {
	v := strings.ToLower(e.Request.URL.Query().Get(key))
	return v == "1" || v == "true" || v == "yes"
}
