package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func (h *Handler) listEntries(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	filter := domain.EntryFilter{
		Kind:        domain.EntryKind(e.Request.URL.Query().Get("kind")),
		IssueID:     domain.IssueID(e.Request.URL.Query().Get("issue_id")),
		PlanID:      domain.PlanID(e.Request.URL.Query().Get("plan_id")),
		CycleID:     domain.CycleID(e.Request.URL.Query().Get("cycle_id")),
		Branch:      e.Request.URL.Query().Get("branch"),
		ExternalRef: e.Request.URL.Query().Get("external_ref"),
		Tags:        csv(e, "tags"),
		Search:      e.Request.URL.Query().Get("q"),
		Limit:       queryInt(e, "limit"),
		Offset:      queryInt(e, "offset"),
	}
	if since, ok := queryTime(e, "since"); ok {
		filter.Since = since
	}
	if until, ok := queryTime(e, "until"); ok {
		filter.Until = until
	}

	entries, err := h.entries.ListEntries(e.Request.Context(), actorOf(e), project, filter)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, map[string]any{"entries": entryViews(entries)})
}

func entryViews(entries []domain.Entry) []entryView {
	out := make([]entryView, 0, len(entries))
	for _, entry := range entries {
		out = append(out, toEntryView(entry))
	}
	return out
}

func (h *Handler) getEntry(e *core.RequestEvent) error {
	entry, err := h.entries.GetEntry(e.Request.Context(), actorOf(e), domain.EntryID(e.Request.PathValue("entry")))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toEntryView(entry))
}

func (h *Handler) getEntryBySlug(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	entry, err := h.entries.GetEntryBySlug(e.Request.Context(), actorOf(e), project, e.Request.PathValue("slug"))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toEntryView(entry))
}

type entryBody struct {
	Kind        *string         `json:"kind"`
	IssueID     *string         `json:"issue_id"`
	PlanID      *string         `json:"plan_id"`
	CycleID     *string         `json:"cycle_id"`
	Slug        *string         `json:"slug"`
	Title       *string         `json:"title"`
	Body        *string         `json:"body"`
	Branch      *string         `json:"branch"`
	PR          *string         `json:"pr"`
	ExternalRef *string         `json:"external_ref"`
	Tags        *[]string       `json:"tags"`
	Meta        *map[string]any `json:"meta"`
}

func (h *Handler) writeEntry(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	var body entryBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	in := ports.WriteEntryInput{ProjectID: project}
	if body.Kind != nil {
		in.Kind = domain.EntryKind(*body.Kind)
	}
	if body.IssueID != nil {
		in.IssueID = domain.IssueID(*body.IssueID)
	}
	if body.PlanID != nil {
		in.PlanID = domain.PlanID(*body.PlanID)
	}
	if body.CycleID != nil {
		in.CycleID = domain.CycleID(*body.CycleID)
	}
	if body.Slug != nil {
		in.Slug = *body.Slug
	}
	if body.Title != nil {
		in.Title = *body.Title
	}
	if body.Body != nil {
		in.Body = *body.Body
	}
	if body.Branch != nil {
		in.Branch = *body.Branch
	}
	if body.PR != nil {
		in.PR = *body.PR
	}
	if body.ExternalRef != nil {
		in.ExternalRef = *body.ExternalRef
	}
	if body.Tags != nil {
		in.Tags = *body.Tags
	}
	if body.Meta != nil {
		in.Meta = *body.Meta
	}

	entry, err := h.entries.WriteEntry(e.Request.Context(), actorOf(e), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toEntryView(entry))
}

func (h *Handler) updateEntry(e *core.RequestEvent) error {
	var body entryBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	in := ports.UpdateEntryInput{
		Slug: body.Slug, Title: body.Title, Body: body.Body,
		Branch: body.Branch, PR: body.PR, ExternalRef: body.ExternalRef,
		Tags: body.Tags, Meta: body.Meta,
	}
	if body.IssueID != nil {
		issue := domain.IssueID(*body.IssueID)
		in.IssueID = &issue
	}
	if body.PlanID != nil {
		plan := domain.PlanID(*body.PlanID)
		in.PlanID = &plan
	}

	entry, err := h.entries.UpdateEntry(e.Request.Context(), actorOf(e), domain.EntryID(e.Request.PathValue("entry")), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toEntryView(entry))
}

type appendEntryBody struct {
	Section string `json:"section"`
}

func (h *Handler) appendEntry(e *core.RequestEvent) error {
	var body appendEntryBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	entry, err := h.entries.AppendToEntry(e.Request.Context(), actorOf(e), domain.EntryID(e.Request.PathValue("entry")), body.Section)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toEntryView(entry))
}

func (h *Handler) deleteEntry(e *core.RequestEvent) error {
	if err := h.entries.DeleteEntry(e.Request.Context(), actorOf(e), domain.EntryID(e.Request.PathValue("entry"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}
