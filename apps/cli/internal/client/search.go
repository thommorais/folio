package client

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type SearchHit struct {
	Kind      string   `json:"kind"`
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
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

func (c *Client) Search(project string, query SearchQuery) ([]SearchHit, error) {
	if strings.TrimSpace(query.Text) == "" && len(query.Tags) == 0 {
		return nil, errors.New("a search needs a term or --tags")
	}

	params := url.Values{}
	if query.Text != "" {
		params.Set("q", query.Text)
	}
	if len(query.Kinds) > 0 {
		params.Set("kind", strings.Join(query.Kinds, ","))
	}
	if len(query.Tags) > 0 {
		params.Set("tags", strings.Join(query.Tags, ","))
	}
	if query.Limit > 0 {
		params.Set("limit", strconv.Itoa(query.Limit))
	}
	if query.Offset > 0 {
		params.Set("offset", strconv.Itoa(query.Offset))
	}

	var body struct {
		Hits []SearchHit `json:"hits"`
	}
	if err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/search?"+params.Encode(), nil, &body); err != nil {
		return nil, err
	}
	return body.Hits, nil
}
