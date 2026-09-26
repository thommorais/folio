package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

func (h *Handler) listClients(e *core.RequestEvent) error {
	clients, err := h.clients.ListClients(e.Request.Context(), actorOf(e))
	if err != nil {
		return fail(e, err)
	}
	out := make([]clientView, 0, len(clients))
	for _, c := range clients {
		out = append(out, toClientView(c))
	}
	return e.JSON(http.StatusOK, map[string]any{"clients": out})
}

func (h *Handler) listDomains(e *core.RequestEvent) error {
	domains, err := h.domains.ListDomains(e.Request.Context(), actorOf(e))
	if err != nil {
		return fail(e, err)
	}
	out := make([]domainView, 0, len(domains))
	for _, d := range domains {
		out = append(out, toDomainView(d))
	}
	return e.JSON(http.StatusOK, map[string]any{"domains": out})
}
