package client

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Doc struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	Kind      string   `json:"kind"`
	TicketID  string   `json:"issue_id,omitempty"`
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type DocInput struct {
	Kind     *string   `json:"kind,omitempty"`
	TicketID *string   `json:"issue_id,omitempty"`
	Slug     *string   `json:"slug,omitempty"`
	Title    *string   `json:"title,omitempty"`
	Body     *string   `json:"body,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
}

type DocFilter struct {
	TicketID string
	Tags     []string
	Search   string
	Limit    int
	Offset   int
}

func (f DocFilter) query() string {
	params := url.Values{}
	params.Set("kind", KindDoc)
	if f.TicketID != "" {
		params.Set("issue_id", f.TicketID)
	}
	if len(f.Tags) > 0 {
		params.Set("tags", strings.Join(f.Tags, ","))
	}
	if f.Search != "" {
		params.Set("q", f.Search)
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

func (c *Client) ListDocs(project string, filter DocFilter) ([]Doc, error) {
	var body struct {
		Entries []Doc `json:"entries"`
	}
	if err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/entries"+filter.query(), nil, &body); err != nil {
		return nil, err
	}
	return body.Entries, nil
}

func (c *Client) GetDoc(id string) (Doc, error) {
	var doc Doc
	err := c.do(http.MethodGet, "/api/folio/entries/"+id, nil, &doc)
	return doc, err
}

func (c *Client) GetDocBySlug(project, slug string) (Doc, error) {
	var doc Doc
	err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/entries/"+slug, nil, &doc)
	return doc, err
}

func (c *Client) CreateDoc(project string, in DocInput) (Doc, error) {
	var doc Doc
	err := c.do(http.MethodPost, "/api/folio/projects/"+project+"/entries", in, &doc)
	return doc, err
}

func (c *Client) UpdateDoc(id string, in DocInput) (Doc, error) {
	var doc Doc
	err := c.do(http.MethodPatch, "/api/folio/entries/"+id, in, &doc)
	return doc, err
}

func (c *Client) DeleteDoc(id string) error {
	return c.do(http.MethodDelete, "/api/folio/entries/"+id, nil, nil)
}
