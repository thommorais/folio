package httpapi

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
)

func (h *Handler) listCycles(e *core.RequestEvent) error {
	cycles, err := h.cycles.ListCycles(e.Request.Context(), actorOf(e), domain.IssueID(e.Request.PathValue("issue")))
	if err != nil {
		return fail(e, err)
	}
	out := make([]cycleView, 0, len(cycles))
	for _, c := range cycles {
		out = append(out, toCycleView(c))
	}
	return e.JSON(http.StatusOK, map[string]any{"cycles": out})
}

func (h *Handler) openCycle(e *core.RequestEvent) error {
	cycle, err := h.cycles.OpenCycle(e.Request.Context(), actorOf(e), domain.IssueID(e.Request.PathValue("issue")))
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toCycleView(cycle))
}

type cycleBody struct {
	Phase      *string `json:"phase"`
	Resolution *string `json:"resolution"`
	MapID      *string `json:"map_id"`
}

func (h *Handler) updateCycle(e *core.RequestEvent) error {
	var body cycleBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	id := domain.CycleID(e.Request.PathValue("cycle"))

	if body.MapID != nil {
		cycle, err := h.cycles.SetCycleMap(e.Request.Context(), actorOf(e), id, domain.IssueID(*body.MapID))
		if err != nil {
			return fail(e, err)
		}
		if body.Phase == nil && body.Resolution == nil {
			return e.JSON(http.StatusOK, toCycleView(cycle))
		}
	}
	if body.Resolution != nil {
		cycle, err := h.cycles.ResolveCycle(e.Request.Context(), actorOf(e), id, *body.Resolution)
		if err != nil {
			return fail(e, err)
		}
		return e.JSON(http.StatusOK, toCycleView(cycle))
	}
	if body.Phase != nil {
		cycle, err := h.cycles.AdvancePhase(e.Request.Context(), actorOf(e), id, domain.Phase(*body.Phase))
		if err != nil {
			return fail(e, err)
		}
		return e.JSON(http.StatusOK, toCycleView(cycle))
	}
	return e.BadRequestError("nothing to update: pass phase, resolution or map_id", nil)
}
