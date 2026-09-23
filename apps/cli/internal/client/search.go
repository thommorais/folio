package client

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type SearchHit struct {
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	// ProjectSlug names the hit's project, which a global search needs to
	// render and a project-scoped one already knows.
	ProjectSlug string `json:"project_slug"`
	// Slug addresses the record in its own routes, where it has one.
	Slug      string   `json:"slug,omitempty"`
	Title     string   `json:"title"`
	Snippet   string   `json:"snippet,omitempty"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
}

type SearchQuery struct {
	Text   string
	Kinds  []string
	Tags   []string
	Limit  int
	Offset int
}

// query is shared by both routes, which take the same parameters and differ
// only in whether a project scopes them.
func (q SearchQuery) query() (string, error) {
	if strings.TrimSpace(q.Text) == "" && len(q.Tags) == 0 {
		return "", errors.New("a search needs a term or --tags")
	}

	params := url.Values{}
	if q.Text != "" {
		params.Set("q", q.Text)
	}
	if len(q.Kinds) > 0 {
		params.Set("kind", strings.Join(q.Kinds, ","))
	}
	if len(q.Tags) > 0 {
		params.Set("tags", strings.Join(q.Tags, ","))
	}
	if q.Limit > 0 {
		params.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Offset > 0 {
		params.Set("offset", strconv.Itoa(q.Offset))
	}
	return "?" + params.Encode(), nil
}

func (c *Client) hits(path string, query SearchQuery) ([]SearchHit, error) {
	q, err := query.query()
	if err != nil {
		return nil, err
	}

	var body struct {
		Hits []SearchHit `json:"hits"`
	}
	if err := c.do(http.MethodGet, path+q, nil, &body); err != nil {
		return nil, err
	}
	return body.Hits, nil
}

// SearchAll spans every project the caller can read, ranked together.
func (c *Client) SearchAll(query SearchQuery) ([]SearchHit, error) {
	return c.hits("/api/folio/search", query)
}

func (c *Client) Search(project string, query SearchQuery) ([]SearchHit, error) {
	return c.hits("/api/folio/projects/"+project+"/search", query)
}
