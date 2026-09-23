package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// Knowledge routes carry no project segment. Every other resource is addressed
// under /projects/{project}, because that is where its permission comes from;
// a note has none, so it sits at the top level.
func (h *Handler) listKnowledge(e *core.RequestEvent) error {
	filter := domain.KnowledgeFilter{
		ProjectID:  domain.ProjectID(e.Request.URL.Query().Get("project")),
		Unattached: e.Request.URL.Query().Get("unattached") == "true",
		Tags:       csv(e, "tags"),
		Search:     e.Request.URL.Query().Get("q"),
		Limit:      queryInt(e, "limit"),
		Offset:     queryInt(e, "offset"),
	}

	notes, err := h.knowledge.ListKnowledge(e.Request.Context(), actorOf(e), filter)
	if err != nil {
		return fail(e, err)
	}
	out := make([]knowledgeView, 0, len(notes))
	for _, note := range notes {
		out = append(out, toKnowledgeView(note))
	}
	return e.JSON(http.StatusOK, map[string]any{"knowledge": out})
}

// getKnowledge accepts an id or a slug. The slug namespace is global, so
// unlike a doc it needs no project to disambiguate.
func (h *Handler) getKnowledge(e *core.RequestEvent) error {
	ref := e.Request.PathValue("knowledge")

	note, err := h.knowledge.GetKnowledge(e.Request.Context(), actorOf(e), domain.KnowledgeID(ref))
	if err != nil {
		bySlug, slugErr := h.knowledge.GetKnowledgeBySlug(e.Request.Context(), actorOf(e), ref)
		if slugErr != nil {
			return fail(e, err)
		}
		note = bySlug
	}
	return e.JSON(http.StatusOK, toKnowledgeView(note))
}

func (h *Handler) createKnowledge(e *core.RequestEvent) error {
	var body knowledgeBody
	if err := e.BindBody(&body); err != nil {
		return fail(e, domain.Invalid("body", "is not valid JSON"))
	}

	note, err := h.knowledge.WriteKnowledge(e.Request.Context(), actorOf(e), body.write())
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toKnowledgeView(note))
}

func (h *Handler) updateKnowledge(e *core.RequestEvent) error {
	var body knowledgeBody
	if err := e.BindBody(&body); err != nil {
		return fail(e, domain.Invalid("body", "is not valid JSON"))
	}

	note, err := h.knowledge.UpdateKnowledge(e.Request.Context(), actorOf(e),
		domain.KnowledgeID(e.Request.PathValue("knowledge")), body.update())
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toKnowledgeView(note))
}

func (h *Handler) deleteKnowledge(e *core.RequestEvent) error {
	err := h.knowledge.DeleteKnowledge(e.Request.Context(), actorOf(e),
		domain.KnowledgeID(e.Request.PathValue("knowledge")))
	if err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}

// knowledgeBody is the wire shape for both a create and an update. The fields
// are pointers so an update can tell "not supplied" from "set to empty".
type knowledgeBody struct {
	ProjectID *string   `json:"project_id"`
	Slug      *string   `json:"slug"`
	Title     *string   `json:"title"`
	Body      *string   `json:"body"`
	Tags      *[]string `json:"tags"`
}

func (b knowledgeBody) write() ports.WriteKnowledgeInput {
	in := ports.WriteKnowledgeInput{}
	if b.ProjectID != nil {
		in.ProjectID = domain.ProjectID(*b.ProjectID)
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
	if b.Tags != nil {
		in.Tags = *b.Tags
	}
	return in
}

func (b knowledgeBody) update() ports.UpdateKnowledgeInput {
	in := ports.UpdateKnowledgeInput{Slug: b.Slug, Title: b.Title, Body: b.Body, Tags: b.Tags}
	if b.ProjectID != nil {
		id := domain.ProjectID(*b.ProjectID)
		in.ProjectID = &id
	}
	return in
}
