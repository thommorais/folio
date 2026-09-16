package client

import (
	"net/http"
	"net/url"
	"strconv"
)

type WorkLog struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	TicketID  string `json:"ticket_id,omitempty"`
	PlanID    string `json:"plan_id,omitempty"`
	TodoID    string `json:"todo_id,omitempty"`
	CycleID   string `json:"cycle_id,omitempty"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type WorkLogFilter struct {
	Cycle  string
	Search string
	Limit  int
}

func (f WorkLogFilter) query() string {
	values := url.Values{}
	if f.Cycle != "" {
		values.Set("cycle", f.Cycle)
	}
	if f.Search != "" {
		values.Set("q", f.Search)
	}
	if f.Limit > 0 {
		values.Set("limit", strconv.Itoa(f.Limit))
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func (c *Client) listWorkLogs(path string, f WorkLogFilter) ([]WorkLog, error) {
	var body struct {
		Logs []WorkLog `json:"logs"`
	}
	if err := c.do(http.MethodGet, path+f.query(), nil, &body); err != nil {
		return nil, err
	}
	return body.Logs, nil
}

func (c *Client) writeWorkLog(path, text string) (WorkLog, error) {
	var entry WorkLog
	err := c.do(http.MethodPost, path, map[string]string{"body": text}, &entry)
	return entry, err
}

func (c *Client) ListTicketLogs(ticket string, f WorkLogFilter) ([]WorkLog, error) {
	return c.listWorkLogs("/api/folio/tickets/"+ticket+"/logs", f)
}

func (c *Client) WriteTicketLog(ticket, body string) (WorkLog, error) {
	return c.writeWorkLog("/api/folio/tickets/"+ticket+"/logs", body)
}

func (c *Client) ListPlanLogs(plan string, f WorkLogFilter) ([]WorkLog, error) {
	return c.listWorkLogs("/api/folio/plans/"+plan+"/logs", f)
}

func (c *Client) WritePlanLog(plan, body string) (WorkLog, error) {
	return c.writeWorkLog("/api/folio/plans/"+plan+"/logs", body)
}

func (c *Client) ListTodoLogs(todo string, f WorkLogFilter) ([]WorkLog, error) {
	return c.listWorkLogs("/api/folio/todos/"+todo+"/logs", f)
}

func (c *Client) WriteTodoLog(todo, body string) (WorkLog, error) {
	return c.writeWorkLog("/api/folio/todos/"+todo+"/logs", body)
}

func (c *Client) DeleteWorkLog(kind, id string) error {
	return c.do(http.MethodDelete, "/api/folio/"+kind+"-logs/"+id, nil, nil)
}
