package client

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Ticket struct {
	ID              string   `json:"id"`
	Kind            string   `json:"kind"`
	ProjectID       string   `json:"project_id"`
	ParentID        string   `json:"parent_id,omitempty"`
	PlanID          string   `json:"plan_id,omitempty"`
	Size            int      `json:"size,omitempty"`
	Score           float64  `json:"score"`
	RelatedTo       []string `json:"related_to"`
	Slug            string   `json:"slug"`
	Title           string   `json:"title"`
	Body            string   `json:"body"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority"`
	Assignee        string   `json:"assignee,omitempty"`
	Tags            []string `json:"tags"`
	ExternalRef     string   `json:"external_ref,omitempty"`
	DependsOn       []string `json:"depends_on"`
	Wayfinder       string   `json:"wayfinder,omitempty"`
	Resolution      string   `json:"resolution,omitempty"`
	ResolutionEntry string   `json:"resolution_entry_id,omitempty"`
	Blocked         bool     `json:"blocked"`
	Cycle           int      `json:"cycle,omitempty"`
	Phase           string   `json:"phase,omitempty"`
	Progress        Progress `json:"progress"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type TicketInput struct {
	Kind            *string   `json:"kind,omitempty"`
	ParentID        *string   `json:"parent_id,omitempty"`
	PlanID          *string   `json:"plan_id,omitempty"`
	Size            *int      `json:"size,omitempty"`
	DependsOn       *[]string `json:"depends_on,omitempty"`
	Wayfinder       *string   `json:"wayfinder,omitempty"`
	Slug            *string   `json:"slug,omitempty"`
	Title           *string   `json:"title,omitempty"`
	Body            *string   `json:"body,omitempty"`
	Status          *string   `json:"status,omitempty"`
	Priority        *string   `json:"priority,omitempty"`
	Assignee        *string   `json:"assignee,omitempty"`
	Tags            *[]string `json:"tags,omitempty"`
	ExternalRef     *string   `json:"external_ref,omitempty"`
	Resolution      *string   `json:"resolution,omitempty"`
	ResolutionEntry *string   `json:"resolution_entry_id,omitempty"`
}

type TicketFilter struct {
	Kind     string
	ParentID string
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
		"kind":     f.Kind,
		"parent":   f.ParentID,
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
	if filter.Kind == "" {
		filter.Kind = KindTicket
	}
	var body struct {
		Issues []Ticket `json:"issues"`
	}
	if err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/issues"+filter.query(), nil, &body); err != nil {
		return nil, err
	}
	return body.Issues, nil
}

const (
	KindTicket     = "ticket"
	KindTodo       = "todo"
	KindJournal    = "journal"
	KindDoc        = "doc"
	KindLog        = "log"
	KindResolution = "resolution"
)

func (c *Client) TicketFrontier(id string) ([]Ticket, error) {
	var body struct {
		Issues []Ticket `json:"issues"`
	}
	if err := c.do(http.MethodGet, "/api/folio/issues/"+id+"/frontier", nil, &body); err != nil {
		return nil, err
	}
	return body.Issues, nil
}

func (c *Client) GetTicket(id string) (Ticket, error) {
	var ticket Ticket
	err := c.do(http.MethodGet, "/api/folio/issues/"+id, nil, &ticket)
	return ticket, err
}

func (c *Client) GetTicketBySlug(project, slug string) (Ticket, error) {
	var ticket Ticket
	err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/issues/"+slug, nil, &ticket)
	return ticket, err
}

func (c *Client) CreateTicket(project string, in TicketInput) (Ticket, error) {
	if in.Kind == nil {
		kind := KindTicket
		in.Kind = &kind
	}
	var ticket Ticket
	err := c.do(http.MethodPost, "/api/folio/projects/"+project+"/issues", in, &ticket)
	return ticket, err
}

func (c *Client) UpdateTicket(id string, in TicketInput) (Ticket, error) {
	var ticket Ticket
	err := c.do(http.MethodPatch, "/api/folio/issues/"+id, in, &ticket)
	return ticket, err
}

type Resolution struct {
	Answer string
	Detail string
	Status string
}

func (c *Client) ResolveTicket(project, id string, r Resolution) (Ticket, error) {
	status := r.Status
	if status == "" {
		status = "done"
	}
	in := TicketInput{Status: &status, Resolution: &r.Answer}

	var detail WorkLog
	if r.Detail != "" {
		kind := KindResolution
		if err := c.do(http.MethodPost, "/api/folio/projects/"+project+"/entries",
			LogInput{Kind: &kind, TicketID: &id, Body: &r.Detail}, &detail); err != nil {
			return Ticket{}, err
		}
		in.ResolutionEntry = &detail.ID
	}

	ticket, err := c.UpdateTicket(id, in)
	if err != nil && detail.ID != "" {
		_ = c.DeleteWorkLog(detail.ID)
	}
	return ticket, err
}

func (c *Client) DeleteTicket(id string) error {
	return c.do(http.MethodDelete, "/api/folio/issues/"+id, nil, nil)
}

type TicketBrief struct {
	Ticket   Ticket         `json:"issue"`
	Children []Ticket       `json:"children"`
	Plans    []Plan         `json:"plans"`
	Journal  []JournalEntry `json:"journal"`
	Cycles   []Cycle        `json:"cycles"`
	Docs     []Doc          `json:"docs"`
	Map      *MapBrief      `json:"map,omitempty"`
}

type MapBrief struct {
	Ticket   Ticket   `json:"issue"`
	Open     int      `json:"open"`
	Frontier []Ticket `json:"frontier"`
}

func (c *Client) GetTicketBrief(id string, recentJournal int) (TicketBrief, error) {
	var brief TicketBrief
	err := c.do(http.MethodGet, "/api/folio/issues/"+id+"/brief"+recentJournalQuery(recentJournal), nil, &brief)
	return brief, err
}

func (c *Client) GetTicketBriefBySlug(project, slug string, recentJournal int) (TicketBrief, error) {
	var brief TicketBrief
	err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/issues/"+slug+"/brief"+recentJournalQuery(recentJournal), nil, &brief)
	return brief, err
}

func recentJournalQuery(n int) string {
	if n <= 0 {
		return ""
	}
	return "?recent_journal=" + strconv.Itoa(n)
}
