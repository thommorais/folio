package client

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type JournalEntry struct {
	ID          string         `json:"id"`
	ProjectID   string         `json:"project_id"`
	TicketID    string         `json:"ticket_id,omitempty"`
	PlanID      string         `json:"plan_id,omitempty"`
	TodoID      string         `json:"todo_id,omitempty"`
	Slug        string         `json:"slug"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	Branch      string         `json:"branch,omitempty"`
	PR          string         `json:"pr,omitempty"`
	ExternalRef string         `json:"external_ref,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
	Tags        []string       `json:"tags"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type LogInput struct {
	TicketID    *string   `json:"ticket_id,omitempty"`
	PlanID      *string   `json:"plan_id,omitempty"`
	TodoID      *string   `json:"todo_id,omitempty"`
	Slug        *string   `json:"slug,omitempty"`
	Title       *string   `json:"title,omitempty"`
	Body        *string   `json:"body,omitempty"`
	Branch      *string   `json:"branch,omitempty"`
	PR          *string   `json:"pr,omitempty"`
	ExternalRef *string   `json:"external_ref,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
}

type JournalFilter struct {
	TicketID    string
	PlanID      string
	TodoID      string
	Branch      string
	ExternalRef string
	Tags        []string
	Search      string
	Since       string
	Until       string
	Limit       int
	Offset      int
}

func (f JournalFilter) query() string {
	params := url.Values{}
	for key, value := range map[string]string{
		"ticket_id":    f.TicketID,
		"plan_id":      f.PlanID,
		"todo_id":      f.TodoID,
		"branch":       f.Branch,
		"external_ref": f.ExternalRef,
		"q":            f.Search,
		"since":        f.Since,
		"until":        f.Until,
	} {
		if value != "" {
			params.Set(key, value)
		}
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

func (c *Client) ListJournal(project string, filter JournalFilter) ([]JournalEntry, error) {
	var body struct {
		Journal []JournalEntry `json:"journal"`
	}
	if err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/journal"+filter.query(), nil, &body); err != nil {
		return nil, err
	}
	return body.Journal, nil
}

func (c *Client) GetJournalEntry(id string) (JournalEntry, error) {
	var entry JournalEntry
	err := c.do(http.MethodGet, "/api/folio/journal/"+id, nil, &entry)
	return entry, err
}

func (c *Client) GetJournalEntryBySlug(project, slug string) (JournalEntry, error) {
	var entry JournalEntry
	err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/journal/"+slug, nil, &entry)
	return entry, err
}

func (c *Client) WriteJournalEntry(project string, in LogInput) (JournalEntry, error) {
	var entry JournalEntry
	err := c.do(http.MethodPost, "/api/folio/projects/"+project+"/journal", in, &entry)
	return entry, err
}

func (c *Client) UpdateJournalEntry(id string, in LogInput) (JournalEntry, error) {
	var entry JournalEntry
	err := c.do(http.MethodPatch, "/api/folio/journal/"+id, in, &entry)
	return entry, err
}

func (c *Client) AppendJournalEntry(id, section string) (JournalEntry, error) {
	var entry JournalEntry
	body := struct {
		Section string `json:"section"`
	}{Section: section}
	err := c.do(http.MethodPost, "/api/folio/journal/"+id+"/append", body, &entry)
	return entry, err
}

func (c *Client) DeleteJournalEntry(id string) error {
	return c.do(http.MethodDelete, "/api/folio/journal/"+id, nil, nil)
}
