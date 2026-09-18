package client

import (
	"net/http"
	"net/url"
	"strconv"
)

type WorkLog struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	ProjectID string `json:"project_id"`
	TicketID  string `json:"issue_id,omitempty"`
	PlanID    string `json:"plan_id,omitempty"`
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

func (f WorkLogFilter) query(target, id string) string {
	values := url.Values{}
	values.Set("kind", KindLog)
	values.Set(target, id)
	if f.Cycle != "" {
		values.Set("cycle_id", f.Cycle)
	}
	if f.Search != "" {
		values.Set("q", f.Search)
	}
	if f.Limit > 0 {
		values.Set("limit", strconv.Itoa(f.Limit))
	}
	return "?" + values.Encode()
}

func (c *Client) listWorkLogs(project, target, id string, f WorkLogFilter) ([]WorkLog, error) {
	var body struct {
		Entries []WorkLog `json:"entries"`
	}
	if err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/entries"+f.query(target, id), nil, &body); err != nil {
		return nil, err
	}
	return body.Entries, nil
}

func (c *Client) writeWorkLog(project, target, id, text string) (WorkLog, error) {
	kind := KindLog
	in := LogInput{Kind: &kind, Body: &text}
	if target == "issue_id" {
		in.TicketID = &id
	} else {
		in.PlanID = &id
	}

	var entry WorkLog
	err := c.do(http.MethodPost, "/api/folio/projects/"+project+"/entries", in, &entry)
	return entry, err
}

func (c *Client) ListIssueLogs(project, issue string, f WorkLogFilter) ([]WorkLog, error) {
	return c.listWorkLogs(project, "issue_id", issue, f)
}

func (c *Client) WriteIssueLog(project, issue, body string) (WorkLog, error) {
	return c.writeWorkLog(project, "issue_id", issue, body)
}

func (c *Client) ListPlanLogs(project, plan string, f WorkLogFilter) ([]WorkLog, error) {
	return c.listWorkLogs(project, "plan_id", plan, f)
}

func (c *Client) WritePlanLog(project, plan, body string) (WorkLog, error) {
	return c.writeWorkLog(project, "plan_id", plan, body)
}

func (c *Client) DeleteWorkLog(id string) error {
	return c.do(http.MethodDelete, "/api/folio/entries/"+id, nil, nil)
}
