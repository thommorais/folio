package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/ports"
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

type updateClientBody struct {
	Slug  *string `json:"slug"`
	Name  *string `json:"name"`
	Site  *string `json:"site"`
	Logo  *string `json:"logo"`
	Descr *string `json:"descr"`
}

func (h *Handler) updateClient(e *core.RequestEvent) error {
	var body updateClientBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	client, err := h.clients.UpdateClient(e.Request.Context(), actorOf(e), e.Request.PathValue("client"), ports.UpdateClientInput{
		Slug: body.Slug, Name: body.Name, Site: body.Site, Logo: body.Logo, Descr: body.Descr,
	})
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, toClientView(client))
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
