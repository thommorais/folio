package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

func (h *Handler) listProjects(e *core.RequestEvent) error {
	projects, err := h.projects.ListProjects(e.Request.Context(), actorOf(e), queryBool(e, "archived"))
	if err != nil {
		return fail(e, err)
	}
	out := make([]projectView, 0, len(projects))
	for _, p := range projects {
		out = append(out, toProjectView(p))
	}
	return e.JSON(http.StatusOK, map[string]any{"projects": out})
}

func (h *Handler) getProject(e *core.RequestEvent) error {
	project, err := h.projects.GetProject(e.Request.Context(), actorOf(e), e.Request.PathValue("project"))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toProjectView(project))
}

type createProjectBody struct {
	DomainID string `json:"domain_id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Descr    string `json:"descr"`
}

func (h *Handler) createProject(e *core.RequestEvent) error {
	var body createProjectBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	// A missing slug is derived from the name, so a caller that only has a
	// human-readable name still gets a stable address.
	if body.Slug == "" {
		body.Slug = rules.Slugify(body.Name)
	}

	project, err := h.projects.CreateProject(e.Request.Context(), actorOf(e), ports.CreateProjectInput{
		DomainID: domain.DomainID(body.DomainID),
		Slug:     body.Slug, Name: body.Name, Descr: body.Descr,
	})
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toProjectView(project))
}

type updateProjectBody struct {
	Name     *string `json:"name"`
	Descr    *string `json:"descr"`
	Archived *bool   `json:"archived"`
}

func (h *Handler) updateProject(e *core.RequestEvent) error {
	id, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	var body updateProjectBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	project, err := h.projects.UpdateProject(e.Request.Context(), actorOf(e), id, ports.UpdateProjectInput{
		Name: body.Name, Descr: body.Descr, Archived: body.Archived,
	})
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toProjectView(project))
}

func (h *Handler) deleteProject(e *core.RequestEvent) error {
	id, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	if err := h.projects.DeleteProject(e.Request.Context(), actorOf(e), id); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}

type memberBody struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *Handler) addMember(e *core.RequestEvent) error {
	id, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	var body memberBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	if body.Role == "" {
		body.Role = string(domain.RoleEditor)
	}

	project, err := h.projects.AddMember(e.Request.Context(), actorOf(e), id, body.Email, domain.Role(body.Role))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toProjectView(project))
}

func (h *Handler) setMemberRole(e *core.RequestEvent) error {
	id, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	var body memberBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	user := domain.UserID(e.Request.PathValue("user"))
	project, err := h.projects.SetMemberRole(e.Request.Context(), actorOf(e), id, user, domain.Role(body.Role))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toProjectView(project))
}

func (h *Handler) removeMember(e *core.RequestEvent) error {
	id, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	user := domain.UserID(e.Request.PathValue("user"))
	project, err := h.projects.RemoveMember(e.Request.Context(), actorOf(e), id, user)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toProjectView(project))
}
