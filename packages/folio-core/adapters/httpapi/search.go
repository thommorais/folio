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
