package client

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Progress struct {
	Total   int `json:"total"`
	Done    int `json:"done"`
	Percent int `json:"percent"`
}

type Plan struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	TicketID  string   `json:"ticket_id,omitempty"`
	Title     string   `json:"title"`
	Goal      string   `json:"goal,omitempty"`
	Status    string   `json:"status"`
	Tags      []string `json:"tags"`
	Progress  Progress `json:"progress"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type PlanInput struct {
	TicketID *string   `json:"ticket_id,omitempty"`
	Title    *string   `json:"title,omitempty"`
	Goal     *string   `json:"goal,omitempty"`
	Status   *string   `json:"status,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
}

type PlanFilter struct {
	TicketID string
	Status   []string
	Limit    int
	Offset   int
}

func (f PlanFilter) query() string {
	params := url.Values{}
	if f.TicketID != "" {
		params.Set("ticket_id", f.TicketID)
	}
	if len(f.Status) > 0 {
		params.Set("status", strings.Join(f.Status, ","))
	}
	if f.Limit > 0 {
		params.Set("limit", strconv.Itoa(f.Limit))
	}
	if f.Offset > 0 {
		params.Set("offset", strconv.Itoa(f.Offset))
	}
	if len(params) == 0 {
		return ""
	}
	return "?" + params.Encode()
}

func (c *Client) ListPlans(project string, filter PlanFilter) ([]Plan, error) {
	var body struct {
		Plans []Plan `json:"plans"`
	}
	if err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/plans"+filter.query(), nil, &body); err != nil {
		return nil, err
	}
	return body.Plans, nil
}

func (c *Client) GetPlan(id string) (Plan, error) {
	var plan Plan
	err := c.do(http.MethodGet, "/api/folio/plans/"+id, nil, &plan)
	return plan, err
}

func (c *Client) CreatePlan(project string, in PlanInput) (Plan, error) {
	var plan Plan
	err := c.do(http.MethodPost, "/api/folio/projects/"+project+"/plans", in, &plan)
	return plan, err
}

func (c *Client) UpdatePlan(id string, in PlanInput) (Plan, error) {
	var plan Plan
	err := c.do(http.MethodPatch, "/api/folio/plans/"+id, in, &plan)
	return plan, err
}

func (c *Client) DeletePlan(id string) error {
	return c.do(http.MethodDelete, "/api/folio/plans/"+id, nil, nil)
}
