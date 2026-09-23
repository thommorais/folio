package client

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Knowledge is a note that belongs to no project, so unlike every other
// resource here its routes carry no project segment.
type Knowledge struct {
	ID        string   `json:"id"`
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	ProjectID string   `json:"project_id,omitempty"`
	Tags      []string `json:"tags"`
	CreatedBy string   `json:"created_by"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type KnowledgeInput struct {
	ProjectID *string   `json:"project_id,omitempty"`
	Slug      *string   `json:"slug,omitempty"`
	Title     *string   `json:"title,omitempty"`
	Body      *string   `json:"body,omitempty"`
	Tags      *[]string `json:"tags,omitempty"`
}

type KnowledgeFilter struct {
	ProjectID string
	// Unattached asks for notes belonging to no project, which an empty
	// ProjectID cannot express: that means "no restriction".
	Unattached bool
	Tags       []string
	Search     string
	Limit      int
	Offset     int
}

func (f KnowledgeFilter) query() string {
	params := url.Values{}
	if f.ProjectID != "" {
		params.Set("project", f.ProjectID)
	}
	if f.Unattached {
		params.Set("unattached", "true")
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

func (c *Client) ListKnowledge(filter KnowledgeFilter) ([]Knowledge, error) {
	var body struct {
		Knowledge []Knowledge `json:"knowledge"`
	}
	if err := c.do(http.MethodGet, "/api/folio/knowledge"+filter.query(), nil, &body); err != nil {
		return nil, err
	}
	return body.Knowledge, nil
}

// GetKnowledge takes an id or a slug: the slug namespace is global, so no
// project is needed to resolve one.
func (c *Client) GetKnowledge(ref string) (Knowledge, error) {
	var note Knowledge
	err := c.do(http.MethodGet, "/api/folio/knowledge/"+url.PathEscape(ref), nil, &note)
	return note, err
}

func (c *Client) CreateKnowledge(in KnowledgeInput) (Knowledge, error) {
	var note Knowledge
	err := c.do(http.MethodPost, "/api/folio/knowledge", in, &note)
	return note, err
}

func (c *Client) UpdateKnowledge(id string, in KnowledgeInput) (Knowledge, error) {
	var note Knowledge
	err := c.do(http.MethodPatch, "/api/folio/knowledge/"+url.PathEscape(id), in, &note)
	return note, err
}

func (c *Client) DeleteKnowledge(id string) error {
	return c.do(http.MethodDelete, "/api/folio/knowledge/"+url.PathEscape(id), nil, nil)
}
