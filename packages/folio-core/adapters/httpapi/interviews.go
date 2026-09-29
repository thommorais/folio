package httpapi

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
)

type interviewView struct {
	ID          string                `json:"id"`
	ProjectID   string                `json:"project_id"`
	IssueID     string                `json:"issue_id"`
	Topic       string                `json:"topic"`
	State       domain.InterviewState `json:"state"`
	AgentStatus string                `json:"agent_status"`
	AgentSince  string                `json:"agent_since"`
	Handled     int                   `json:"handled"`
	FinishedAt  string                `json:"finished_at,omitempty"`
	CreatedBy   string                `json:"created_by,omitempty"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
	URL         string                `json:"url,omitempty"`
}

// pageURL joins the web app's origin and a path. An unconfigured base falls
// back to the origin the request came in on.
func pageURL(base string, r *http.Request, path string) string {
	if path == "" {
		return ""
	}
	if base = strings.TrimRight(base, "/"); base == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
		base = scheme + "://" + r.Host
	}
	return base + path
}

func (h *Handler) interviewURL(e *core.RequestEvent) (string, error) {
	path, err := h.interviews.PagePath(e.Request.Context(), actorOf(e), issueOf(e))
	if err != nil {
		return "", err
	}
	return pageURL(h.webURL, e.Request, path), nil
}

func (h *Handler) interviewViewOf(e *core.RequestEvent, i domain.Interview) (interviewView, error) {
	link, err := h.interviewURL(e)
	if err != nil {
		return interviewView{}, err
	}
	out := toInterviewView(i)
	out.URL = link
	return out, nil
}

func toInterviewView(i domain.Interview) interviewView {
	out := interviewView{
		ID: string(i.ID), ProjectID: string(i.ProjectID), IssueID: string(i.IssueID), Topic: i.Topic,
		State: i.State, AgentStatus: string(i.AgentStatus), AgentSince: rfc3339(i.AgentSince), Handled: i.Handled,
		CreatedBy: string(i.CreatedBy), CreatedAt: rfc3339(i.CreatedAt), UpdatedAt: rfc3339(i.UpdatedAt),
	}
	if out.State.Terms == nil {
		out.State.Terms = []domain.Term{}
	}
	if out.State.Questions == nil {
		out.State.Questions = []domain.Question{}
	}
	if i.FinishedAt != nil {
		out.FinishedAt = rfc3339(*i.FinishedAt)
	}
	return out
}

type sendView struct {
	Seq     int                 `json:"seq"`
	At      string              `json:"at"`
	Actions []domain.SendAction `json:"actions"`
}

func toSendView(e domain.InterviewEvent) sendView {
	actions := e.Actions
	if actions == nil {
		actions = []domain.SendAction{}
	}
	return sendView{Seq: e.Seq, At: rfc3339(e.At), Actions: actions}
}

func issueOf(e *core.RequestEvent) domain.IssueID {
	return domain.IssueID(e.Request.PathValue("issue"))
}

func (h *Handler) listInterviews(e *core.RequestEvent) error {
	list, err := h.interviews.ListInterviews(e.Request.Context(), actorOf(e), issueOf(e))
	if err != nil {
		return fail(e, err)
	}
	link, err := h.interviewURL(e)
	if err != nil {
		return fail(e, err)
	}
	out := make([]interviewView, 0, len(list))
	for _, i := range list {
		view := toInterviewView(i)
		view.URL = link
		out = append(out, view)
	}
	return e.JSON(http.StatusOK, map[string]any{"interviews": out})
}

func (h *Handler) startInterview(e *core.RequestEvent) error {
	interview, created, err := h.interviews.StartInterview(e.Request.Context(), actorOf(e), issueOf(e))
	if err != nil {
		return fail(e, err)
	}
	view, err := h.interviewViewOf(e, interview)
	if err != nil {
		return fail(e, err)
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	return e.JSON(status, view)
}

func (h *Handler) currentInterview(e *core.RequestEvent) error {
	interview, err := h.interviews.CurrentInterview(e.Request.Context(), actorOf(e), issueOf(e))
	if err != nil {
		return fail(e, err)
	}
	view, err := h.interviewViewOf(e, interview)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, view)
}

func (h *Handler) patchInterview(e *core.RequestEvent) error {
	raw, err := io.ReadAll(e.Request.Body)
	if err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return e.BadRequestError("the patch is empty", nil)
	}
	summary, err := h.interviews.PatchInterview(e.Request.Context(), actorOf(e), issueOf(e), raw)
	if err != nil {
		return fail(e, err)
	}
	view, err := h.interviewViewOf(e, summary.Interview)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, map[string]any{
		"round":     summary.Round,
		"added":     summary.Added,
		"answered":  summary.Answered,
		"handled":   summary.Handled,
		"interview": view,
	})
}

func (h *Handler) pendingSends(e *core.RequestEvent) error {
	ctx, actor := e.Request.Context(), actorOf(e)
	pending, err := h.interviews.PendingSends(ctx, actor, issueOf(e))
	if err != nil {
		return fail(e, err)
	}
	interview, err := h.interviews.CurrentInterview(ctx, actor, issueOf(e))
	if err != nil {
		return fail(e, err)
	}
	out := make([]sendView, 0, len(pending))
	for _, s := range pending {
		out = append(out, toSendView(s))
	}
	return e.JSON(http.StatusOK, map[string]any{"handled": interview.Handled, "sends": out})
}

type sendBody struct {
	Actions []domain.SendAction `json:"actions"`
}

func (h *Handler) sendToInterview(e *core.RequestEvent) error {
	var body sendBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	event, err := h.interviews.SendToInterview(e.Request.Context(), actorOf(e), issueOf(e), body.Actions)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusCreated, toSendView(event))
}

type finishBody struct {
	Answer string `json:"answer"`
	Doc    string `json:"doc"`
}

func (h *Handler) finishInterview(e *core.RequestEvent) error {
	var body finishBody
	if err := e.BindBody(&body); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	interview, err := h.interviews.FinishInterview(e.Request.Context(), actorOf(e), issueOf(e), body.Answer, body.Doc)
	if err != nil {
		return fail(e, err)
	}
	view, err := h.interviewViewOf(e, interview)
	if err != nil {
		return fail(e, err)
	}
	return e.JSON(http.StatusOK, view)
}
