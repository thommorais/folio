package client

import (
	"encoding/json"
	"net/http"
)

type Interview struct {
	ID          string          `json:"id"`
	ProjectID   string          `json:"project_id"`
	TicketID    string          `json:"issue_id"`
	Topic       string          `json:"topic"`
	State       json.RawMessage `json:"state"`
	AgentStatus string          `json:"agent_status"`
	AgentSince  string          `json:"agent_since,omitempty"`
	Handled     int             `json:"handled"`
	FinishedAt  string          `json:"finished_at,omitempty"`
	CreatedBy   string          `json:"created_by,omitempty"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	URL         string          `json:"url,omitempty"`
}

type InterviewQuestion struct {
	ID     string `json:"id"`
	Round  int    `json:"round"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func (i Interview) Questions() ([]InterviewQuestion, error) {
	var state struct {
		Questions []InterviewQuestion `json:"questions"`
	}
	if len(i.State) == 0 {
		return nil, nil
	}
	err := json.Unmarshal(i.State, &state)
	return state.Questions, err
}

type InterviewPatchSummary struct {
	Interview Interview `json:"interview"`
	Round     int       `json:"round"`
	Added     int       `json:"added"`
	Answered  int       `json:"answered"`
	Handled   int       `json:"handled"`
}

type InterviewSend struct {
	Seq     int               `json:"seq"`
	At      string            `json:"at"`
	Actions []json.RawMessage `json:"actions"`
}

type InterviewPending struct {
	Handled int             `json:"handled"`
	Sends   []InterviewSend `json:"sends"`
}

func interviewPath(ticket string) string {
	return "/api/folio/issues/" + ticket + "/interview"
}

func (c *Client) StartInterview(ticket string) (Interview, error) {
	var out Interview
	err := c.do(http.MethodPost, interviewPath(ticket), struct{}{}, &out)
	return out, err
}

func (c *Client) CurrentInterview(ticket string) (Interview, error) {
	var out Interview
	err := c.do(http.MethodGet, interviewPath(ticket), nil, &out)
	return out, err
}

func (c *Client) ListInterviews(ticket string) ([]Interview, error) {
	var body struct {
		Interviews []Interview `json:"interviews"`
	}
	if err := c.do(http.MethodGet, "/api/folio/issues/"+ticket+"/interviews", nil, &body); err != nil {
		return nil, err
	}
	return body.Interviews, nil
}

func (c *Client) PatchInterview(ticket string, patch json.RawMessage) (InterviewPatchSummary, error) {
	var out InterviewPatchSummary
	err := c.do(http.MethodPatch, interviewPath(ticket), patch, &out)
	return out, err
}

func (c *Client) PendingSends(ticket string) (InterviewPending, error) {
	var out InterviewPending
	err := c.do(http.MethodGet, interviewPath(ticket)+"/pending", nil, &out)
	return out, err
}

func (c *Client) FinishInterview(ticket, answer, doc string) (Interview, error) {
	var out Interview
	body := map[string]string{"answer": answer, "doc": doc}
	err := c.do(http.MethodPost, interviewPath(ticket)+"/finish", body, &out)
	return out, err
}

func (c *Client) BaseURL() string {
	return c.baseURL
}
