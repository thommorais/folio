package client

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Ticket struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id"`
	ParentID    string   `json:"parent_id,omitempty"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Assignee    string   `json:"assignee,omitempty"`
	Tags        []string `json:"tags"`
	ExternalRef string   `json:"external_ref,omitempty"`
	DependsOn   []string `json:"depends_on"`
	Wayfinder   string   `json:"wayfinder,omitempty"`
	Blocked     bool     `json:"blocked"`
	Cycle       int      `json:"cycle,omitempty"`
	Phase       string   `json:"phase,omitempty"`
	Progress    Progress `json:"progress"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type TicketInput struct {
	ParentID    *string   `json:"parent_id,omitempty"`
	DependsOn   *[]string `json:"depends_on,omitempty"`
	Wayfinder   *string   `json:"wayfinder,omitempty"`
	Slug        *string   `json:"slug,omitempty"`
	Title       *string   `json:"title,omitempty"`
	Body        *string   `json:"body,omitempty"`
	Status      *string   `json:"status,omitempty"`
	Priority    *string   `json:"priority,omitempty"`
	Assignee    *string   `json:"assignee,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
	ExternalRef *string   `json:"external_ref,omitempty"`
}

type TicketFilter struct {
	Status   []string
	Priority string
	Assignee string
	Tags     []string
	Search   string
	Limit    int
	Offset   int
}

func (f TicketFilter) query() string {
	params := url.Values{}
	for key, value := range map[string]string{
		"priority": f.Priority,
		"assignee": f.Assignee,
		"q":        f.Search,
	} {
		if value != "" {
			params.Set(key, value)
		}
	}
	if len(f.Status) > 0 {
		params.Set("status", strings.Join(f.Status, ","))
	}
	if len(f.Tags) > 0 {
		params.Set("tags", strings.Join(f.Tags, ","))
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

func (c *Client) ListTickets(project string, filter TicketFilter) ([]Ticket, error) {
	var body struct {
		Tickets []Ticket `json:"tickets"`
	}
	if err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/tickets"+filter.query(), nil, &body); err != nil {
		return nil, err
	}
	return body.Tickets, nil
}

func (c *Client) TicketFrontier(id string) ([]Ticket, error) {
	var body struct {
		Tickets []Ticket `json:"tickets"`
	}
	if err := c.do(http.MethodGet, "/api/folio/tickets/"+id+"/frontier", nil, &body); err != nil {
		return nil, err
	}
	return body.Tickets, nil
}

func (c *Client) GetTicket(id string) (Ticket, error) {
	var ticket Ticket
	err := c.do(http.MethodGet, "/api/folio/tickets/"+id, nil, &ticket)
	return ticket, err
}

func (c *Client) GetTicketBySlug(project, slug string) (Ticket, error) {
	var ticket Ticket
	err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/tickets/"+slug, nil, &ticket)
	return ticket, err
}

func (c *Client) CreateTicket(project string, in TicketInput) (Ticket, error) {
	var ticket Ticket
	err := c.do(http.MethodPost, "/api/folio/projects/"+project+"/tickets", in, &ticket)
	return ticket, err
}

func (c *Client) UpdateTicket(id string, in TicketInput) (Ticket, error) {
	var ticket Ticket
	err := c.do(http.MethodPatch, "/api/folio/tickets/"+id, in, &ticket)
	return ticket, err
}

func (c *Client) DeleteTicket(id string) error {
	return c.do(http.MethodDelete, "/api/folio/tickets/"+id, nil, nil)
}

type TicketBrief struct {
	Ticket  Ticket         `json:"ticket"`
	Plans   []Plan         `json:"plans"`
	Todos   []Todo         `json:"todos"`
	Journal []JournalEntry `json:"journal"`
	Cycles  []Cycle        `json:"cycles"`
	Docs    []Doc          `json:"docs"`
}

func (c *Client) GetTicketBrief(id string, recentJournal int) (TicketBrief, error) {
	var brief TicketBrief
	err := c.do(http.MethodGet, "/api/folio/tickets/"+id+"/brief"+recentJournalQuery(recentJournal), nil, &brief)
	return brief, err
}

func (c *Client) GetTicketBriefBySlug(project, slug string, recentJournal int) (TicketBrief, error) {
	var brief TicketBrief
	err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/tickets/"+slug+"/brief"+recentJournalQuery(recentJournal), nil, &brief)
	return brief, err
}

func recentJournalQuery(n int) string {
	if n <= 0 {
		return ""
	}
	return "?recent_journal=" + strconv.Itoa(n)
}
