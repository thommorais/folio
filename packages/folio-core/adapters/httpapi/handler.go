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
	issues   ports.IssueUseCase
	entries  ports.EntryUseCase
	cycles   ports.CycleUseCase
	search   ports.SearchUseCase
	shares   ports.ShareUseCase
}

type Deps struct {
	Projects ports.ProjectUseCase
	Plans    ports.PlanUseCase
	Issues   ports.IssueUseCase
	Entries  ports.EntryUseCase
	Cycles   ports.CycleUseCase
	Search   ports.SearchUseCase
	Shares   ports.ShareUseCase
}

func New(d Deps) *Handler {
	return &Handler{
		projects: d.Projects, plans: d.Plans, issues: d.Issues, entries: d.Entries, cycles: d.Cycles, search: d.Search, shares: d.Shares,
	}
}

// BasePath is where the API mounts. PocketBase owns /api/collections and the
// rest of its own surface, so folio takes its own namespace.
const BasePath = "/api/folio"

const SharePath = "/api/share"

// Mount registers every route. Auth is PocketBase's: RequireAuth rejects
// anonymous callers before a handler runs, and actorOf reads the record it
// attached to the request.
func (h *Handler) Mount(e *core.ServeEvent) {
	e.Router.GET(SharePath+"/{token}", h.openShare)

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

	// The slug route is registered before the ID route so a project-scoped
	// slug lookup is not shadowed by it.
	g.GET("/issues/{issue}/cycles", h.listCycles)
	g.POST("/issues/{issue}/cycles", h.openCycle)
	g.PATCH("/cycles/{cycle}", h.updateCycle)

	g.GET("/projects/{project}/issues", h.listIssues)
	g.POST("/projects/{project}/issues", h.createIssues)
	g.GET("/projects/{project}/issues/{slug}", h.getIssueBySlug)
	g.GET("/projects/{project}/issues/{slug}/brief", h.getIssueBriefBySlug)
	g.GET("/issues/{issue}", h.getIssue)
	g.GET("/issues/{issue}/brief", h.getIssueBrief)
	g.GET("/issues/{issue}/frontier", h.issueFrontier)
	g.PATCH("/issues/{issue}", h.updateIssue)
	g.DELETE("/issues/{issue}", h.deleteIssue)
	g.POST("/issues/{issue}/links", h.linkIssue)
	g.DELETE("/issues/{issue}/links/{to}", h.unlinkIssue)




	g.GET("/projects/{project}/entries", h.listEntries)
	g.POST("/projects/{project}/entries", h.writeEntry)
	g.GET("/projects/{project}/entries/{slug}", h.getEntryBySlug)
	g.GET("/entries/{entry}", h.getEntry)
	g.PATCH("/entries/{entry}", h.updateEntry)
	g.POST("/entries/{entry}/append", h.appendEntry)
	g.DELETE("/entries/{entry}", h.deleteEntry)

	g.GET("/projects/{project}/search", h.searchProject)

	g.GET("/projects/{project}/shares", h.listShares)
	g.POST("/issues/{issue}/shares", h.shareIssue)
	g.POST("/plans/{plan}/shares", h.sharePlan)
	g.DELETE("/shares/{share}", h.revokeShare)
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
