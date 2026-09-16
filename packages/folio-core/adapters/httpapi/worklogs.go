package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
)

type workLogBody struct {
	Body string `json:"body"`
}

func (h *Handler) listTicketLogs(e *core.RequestEvent) error {
	f := domain.TicketLogFilter{
		CycleID: domain.CycleID(e.Request.URL.Query().Get("cycle")),
		Search:  e.Request.URL.Query().Get("q"),
		Limit:   queryInt(e, "limit"),
		Offset:  queryInt(e, "offset"),
	}
	entries, err := h.workLogs.ListTicketLogs(e.Request.Context(), actorOf(e), domain.TicketID(e.Request.PathValue("ticket")), f)
	if err != nil {
		return fail(e, err)
	}
	out := make([]ticketLogView, 0, len(entries))
	for _, l := range entries {
		out = append(out, toTicketLogView(l))
	}
	return e.JSON(http.StatusOK, map[string]any{"logs": out})
}

func (h *Handler) writeTicketLog(e *core.RequestEvent) error {
	var body workLogBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	entry, err := h.workLogs.WriteTicketLog(e.Request.Context(), actorOf(e), domain.TicketID(e.Request.PathValue("ticket")), body.Body)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toTicketLogView(entry))
}

func (h *Handler) deleteTicketLog(e *core.RequestEvent) error {
	if err := h.workLogs.DeleteTicketLog(e.Request.Context(), actorOf(e), domain.TicketLogID(e.Request.PathValue("log"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}

func (h *Handler) listPlanLogs(e *core.RequestEvent) error {
	f := domain.PlanLogFilter{
		Search: e.Request.URL.Query().Get("q"),
		Limit:  queryInt(e, "limit"),
		Offset: queryInt(e, "offset"),
	}
	entries, err := h.workLogs.ListPlanLogs(e.Request.Context(), actorOf(e), domain.PlanID(e.Request.PathValue("plan")), f)
	if err != nil {
		return fail(e, err)
	}
	out := make([]planLogView, 0, len(entries))
	for _, l := range entries {
		out = append(out, toPlanLogView(l))
	}
	return e.JSON(http.StatusOK, map[string]any{"logs": out})
}

func (h *Handler) writePlanLog(e *core.RequestEvent) error {
	var body workLogBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	entry, err := h.workLogs.WritePlanLog(e.Request.Context(), actorOf(e), domain.PlanID(e.Request.PathValue("plan")), body.Body)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toPlanLogView(entry))
}

func (h *Handler) deletePlanLog(e *core.RequestEvent) error {
	if err := h.workLogs.DeletePlanLog(e.Request.Context(), actorOf(e), domain.PlanLogID(e.Request.PathValue("log"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}

func (h *Handler) listTodoLogs(e *core.RequestEvent) error {
	f := domain.TodoLogFilter{
		Search: e.Request.URL.Query().Get("q"),
		Limit:  queryInt(e, "limit"),
		Offset: queryInt(e, "offset"),
	}
	entries, err := h.workLogs.ListTodoLogs(e.Request.Context(), actorOf(e), domain.TodoID(e.Request.PathValue("todo")), f)
	if err != nil {
		return fail(e, err)
	}
	out := make([]todoLogView, 0, len(entries))
	for _, l := range entries {
		out = append(out, toTodoLogView(l))
	}
	return e.JSON(http.StatusOK, map[string]any{"logs": out})
}

func (h *Handler) writeTodoLog(e *core.RequestEvent) error {
	var body workLogBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	entry, err := h.workLogs.WriteTodoLog(e.Request.Context(), actorOf(e), domain.TodoID(e.Request.PathValue("todo")), body.Body)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toTodoLogView(entry))
}

func (h *Handler) deleteTodoLog(e *core.RequestEvent) error {
	if err := h.workLogs.DeleteTodoLog(e.Request.Context(), actorOf(e), domain.TodoLogID(e.Request.PathValue("log"))); err != nil {
		return fail(e, err)
	}
	return e.NoContent(http.StatusNoContent)
}
