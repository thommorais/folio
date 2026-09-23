package httpapi

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
)

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

// searchAll spans every project the caller belongs to, which is what a command
// palette needs: it has no project in scope.
func (h *Handler) searchAll(e *core.RequestEvent) error {
	hits, err := h.search.SearchAll(e.Request.Context(), actorOf(e), searchQueryOf(e))
	if err != nil {
		return fail(e, err)
	}
	return respondHits(e, hits)
}

func (h *Handler) searchProject(e *core.RequestEvent) error {
	project, err := h.resolveProject(e)
	if err != nil {
		return fail(e, err)
	}
	hits, err := h.search.Search(e.Request.Context(), actorOf(e), project, searchQueryOf(e))
	if err != nil {
		return fail(e, err)
	}
	return respondHits(e, hits)
}

// searchQueryOf reads the query parameters both search routes share.
func searchQueryOf(e *core.RequestEvent) domain.SearchQuery {
	query := domain.SearchQuery{
		Text:   e.Request.URL.Query().Get("q"),
		Tags:   csv(e, "tags"),
		Limit:  queryInt(e, "limit"),
		Offset: queryInt(e, "offset"),
	}
	for _, k := range csv(e, "kind") {
		query.Kinds = append(query.Kinds, domain.SearchKind(k))
	}
	return query
}

func respondHits(e *core.RequestEvent, hits []domain.SearchHit) error {
	out := make([]searchHitView, 0, len(hits))
	for _, hit := range hits {
		out = append(out, toSearchHitView(hit))
	}
	return e.JSON(http.StatusOK, map[string]any{"hits": out})
}
