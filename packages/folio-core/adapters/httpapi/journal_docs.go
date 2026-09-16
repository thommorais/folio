package httpapi

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func (h *Handler) listJournal(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	filter := domain.JournalFilter{
		PlanID:      domain.PlanID(e.Request.URL.Query().Get("plan_id")),
		TodoID:      domain.TodoID(e.Request.URL.Query().Get("todo_id")),
		TicketID:    domain.TicketID(e.Request.URL.Query().Get("ticket_id")),
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

	entries, err := h.journal.ListJournal(e.Request.Context(), actorOf(e), project, filter)
	if err != nil {
		return fail(e, err)
	}
	out := make([]journalView, 0, len(entries))
	for _, entry := range entries {
		out = append(out, toJournalView(entry))
	}
	return e.JSON(http.StatusOK, map[string]any{"journal": out})
}

// queryTime parses an RFC 3339 query parameter, reporting whether one was
// both present and valid. An unparseable value is ignored rather than
// failing the read: a log query should degrade, not 400.
func queryTime(e *core.RequestEvent, key string) (*time.Time, bool) {
	raw := e.Request.URL.Query().Get(key)
	if raw == "" {
		return nil, false
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, false
	}
	return &t, true
}

func (h *Handler) getJournalEntryBySlug(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	entry, err := h.journal.GetJournalEntryBySlug(e.Request.Context(), actorOf(e), project, e.Request.PathValue("slug"))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toJournalView(entry))
}

func (h *Handler) getJournalEntry(e *core.RequestEvent) error {
	entry, err := h.journal.GetJournalEntry(e.Request.Context(), actorOf(e), domain.JournalID(e.Request.PathValue("entry")))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toJournalView(entry))
}

type logBody struct {
	TicketID    *string         `json:"ticket_id"`
	PlanID      *string         `json:"plan_id"`
	TodoID      *string         `json:"todo_id"`
	Slug        *string         `json:"slug"`
	Title       *string         `json:"title"`
	Body        *string         `json:"body"`
	Branch      *string         `json:"branch"`
	PR          *string         `json:"pr"`
	ExternalRef *string         `json:"external_ref"`
	Tags        *[]string       `json:"tags"`
	Meta        *map[string]any `json:"meta"`
}

func (h *Handler) writeJournalEntry(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	var body logBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	in := ports.WriteJournalInput{ProjectID: project}
	if body.TicketID != nil {
		in.TicketID = domain.TicketID(*body.TicketID)
	}
	if body.PlanID != nil {
		in.PlanID = domain.PlanID(*body.PlanID)
	}
	if body.TodoID != nil {
		in.TodoID = domain.TodoID(*body.TodoID)
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

	entry, err := h.journal.WriteJournalEntry(e.Request.Context(), actorOf(e), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toJournalView(entry))
}

func (h *Handler) updateJournalEntry(e *core.RequestEvent) error {
	var body logBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	in := ports.UpdateJournalInput{
		Title: body.Title, Body: body.Body, Branch: body.Branch,
		PR: body.PR, ExternalRef: body.ExternalRef, Tags: body.Tags, Meta: body.Meta,
	}
	if body.TicketID != nil {
		id := domain.TicketID(*body.TicketID)
		in.TicketID = &id
	}
	if body.PlanID != nil {
		id := domain.PlanID(*body.PlanID)
		in.PlanID = &id
	}
	if body.TodoID != nil {
		id := domain.TodoID(*body.TodoID)
		in.TodoID = &id
	}

	entry, err := h.journal.UpdateJournalEntry(e.Request.Context(), actorOf(e), domain.JournalID(e.Request.PathValue("entry")), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toJournalView(entry))
}

type appendJournalBody struct {
	Section string `json:"section"`
}

// appendLog adds a section to an existing entry, so recording progress on
// work already written up does not mean resending the whole body.
func (h *Handler) appendJournalEntry(e *core.RequestEvent) error {
	var body appendJournalBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	entry, err := h.journal.AppendToJournalEntry(e.Request.Context(), actorOf(e), domain.JournalID(e.Request.PathValue("entry")), body.Section)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toJournalView(entry))
}

func (h *Handler) deleteJournalEntry(e *core.RequestEvent) error {
	if err := h.journal.DeleteJournalEntry(e.Request.Context(), actorOf(e), domain.JournalID(e.Request.PathValue("entry"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}

func (h *Handler) listDocs(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	docs, err := h.docs.ListDocs(e.Request.Context(), actorOf(e), project, domain.DocFilter{
		TicketID: domain.TicketID(e.Request.URL.Query().Get("ticket_id")),
		Tags:     csv(e, "tags"),
		Search:   e.Request.URL.Query().Get("q"),
		Limit:    queryInt(e, "limit"),
		Offset:   queryInt(e, "offset"),
	})
	if err != nil {
		return fail(e, err)
	}
	out := make([]docView, 0, len(docs))
	for _, d := range docs {
		out = append(out, toDocView(d))
	}
	return e.JSON(http.StatusOK, map[string]any{"docs": out})
}

func (h *Handler) getDoc(e *core.RequestEvent) error {
	doc, err := h.docs.GetDoc(e.Request.Context(), actorOf(e), domain.DocID(e.Request.PathValue("doc")))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toDocView(doc))
}

func (h *Handler) getDocBySlug(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	doc, err := h.docs.GetDocBySlug(e.Request.Context(), actorOf(e), project, e.Request.PathValue("slug"))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toDocView(doc))
}

type docBody struct {
	TicketID *string   `json:"ticket_id"`
	Slug     *string   `json:"slug"`
	Title    *string   `json:"title"`
	Body     *string   `json:"body"`
	Tags     *[]string `json:"tags"`
}

func (h *Handler) createDoc(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	var body docBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	in := ports.CreateDocInput{ProjectID: project}
	if body.TicketID != nil {
		in.TicketID = domain.TicketID(*body.TicketID)
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
	if body.Tags != nil {
		in.Tags = *body.Tags
	}

	doc, err := h.docs.CreateDoc(e.Request.Context(), actorOf(e), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toDocView(doc))
}

func (h *Handler) updateDoc(e *core.RequestEvent) error {
	var body docBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	in := ports.UpdateDocInput{Slug: body.Slug, Title: body.Title, Body: body.Body, Tags: body.Tags}
	if body.TicketID != nil {
		id := domain.TicketID(*body.TicketID)
		in.TicketID = &id
	}

	doc, err := h.docs.UpdateDoc(e.Request.Context(), actorOf(e), domain.DocID(e.Request.PathValue("doc")), in)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toDocView(doc))
}

func (h *Handler) deleteDoc(e *core.RequestEvent) error {
	if err := h.docs.DeleteDoc(e.Request.Context(), actorOf(e), domain.DocID(e.Request.PathValue("doc"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}

// searchAll spans every project the caller belongs to, which is what the
// command palette needs: it has no project in scope.
func (h *Handler) searchAll(e *core.RequestEvent) error {
	query := domain.SearchQuery{
		Text:   e.Request.URL.Query().Get("q"),
		Tags:   csv(e, "tags"),
		Limit:  queryInt(e, "limit"),
		Offset: queryInt(e, "offset"),
	}
	for _, k := range csv(e, "kind") {
		query.Kinds = append(query.Kinds, domain.SearchKind(k))
	}

	hits, err := h.search.SearchAll(e.Request.Context(), actorOf(e), query)
	if err != nil {
		return fail(e, err)
	}
	out := make([]searchHitView, 0, len(hits))
	for _, hit := range hits {
		out = append(out, toSearchHitView(hit))
	}
	return e.JSON(http.StatusOK, map[string]any{"hits": out})
}

func (h *Handler) searchProject(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	query := domain.SearchQuery{
		Text:   e.Request.URL.Query().Get("q"),
		Tags:   csv(e, "tags"),
		Limit:  queryInt(e, "limit"),
		Offset: queryInt(e, "offset"),
	}
	for _, k := range csv(e, "kind") {
		query.Kinds = append(query.Kinds, domain.SearchKind(k))
	}

	hits, err := h.search.Search(e.Request.Context(), actorOf(e), project, query)
	if err != nil {
		return fail(e, err)
	}
	out := make([]searchHitView, 0, len(hits))
	for _, hit := range hits {
		out = append(out, toSearchHitView(hit))
	}
	return e.JSON(http.StatusOK, map[string]any{"hits": out})
}
