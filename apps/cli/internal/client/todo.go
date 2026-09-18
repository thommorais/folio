package client

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Todo struct {
	ID        string   `json:"id"`
	Kind      string   `json:"kind"`
	ProjectID string   `json:"project_id"`
	TicketID  string   `json:"parent_id,omitempty"`
	PlanID    string   `json:"plan_id,omitempty"`
	Title     string   `json:"title"`
	Details   string   `json:"body,omitempty"`
	Size      int      `json:"size,omitempty"`
	Score     float64  `json:"score"`
	Status    string   `json:"status"`
	Priority  string   `json:"priority"`
	Tags      []string `json:"tags"`
	Position  int      `json:"position"`
	DependsOn []string `json:"depends_on"`
	DueDate   string   `json:"due_date,omitempty"`
	Blocked   bool     `json:"blocked"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// Pointers: an absent field means "leave alone" on PATCH, "default" on POST.
type TodoInput struct {
	Kind      *string   `json:"kind,omitempty"`
	TicketID  *string   `json:"parent_id,omitempty"`
	PlanID    *string   `json:"plan_id,omitempty"`
	Title     *string   `json:"title,omitempty"`
	Details   *string   `json:"body,omitempty"`
	Size      *int      `json:"size,omitempty"`
	Status    *string   `json:"status,omitempty"`
	Priority  *string   `json:"priority,omitempty"`
	Tags      *[]string `json:"tags,omitempty"`
	Position  *int      `json:"position,omitempty"`
	DependsOn *[]string `json:"depends_on,omitempty"`
	DueDate   *string   `json:"due_date,omitempty"`
}

type TodoFilter struct {
	TicketID string
	PlanID   string
	Status   []string
	Priority string
	Tags     []string
	Search   string
	Limit    int
	Offset   int
}

func (f TodoFilter) query() string {
	params := url.Values{}
	params.Set("kind", KindTodo)
	if f.TicketID != "" {
		params.Set("parent", f.TicketID)
	}
	if f.PlanID != "" {
		params.Set("plan_id", f.PlanID)
	}
	if len(f.Status) > 0 {
		params.Set("status", strings.Join(f.Status, ","))
	}
	if f.Priority != "" {
		params.Set("priority", f.Priority)
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

func (c *Client) ListTodos(project string, filter TodoFilter) ([]Todo, error) {
	var body struct {
		Issues []Todo `json:"issues"`
	}
	if err := c.do(http.MethodGet, "/api/folio/projects/"+project+"/issues"+filter.query(), nil, &body); err != nil {
		return nil, err
	}
	return body.Issues, nil
}

func (c *Client) GetTodo(id string) (Todo, error) {
	var todo Todo
	err := c.do(http.MethodGet, "/api/folio/issues/"+id, nil, &todo)
	return todo, err
}

func (c *Client) CreateTodo(project string, in TodoInput) (Todo, error) {
	if in.Kind == nil {
		kind := KindTodo
		in.Kind = &kind
	}
	var todo Todo
	err := c.do(http.MethodPost, "/api/folio/projects/"+project+"/issues", in, &todo)
	return todo, err
}

func (c *Client) UpdateTodo(id string, in TodoInput) (Todo, error) {
	var todo Todo
	err := c.do(http.MethodPatch, "/api/folio/issues/"+id, in, &todo)
	return todo, err
}

func (c *Client) DeleteTodo(id string) error {
	return c.do(http.MethodDelete, "/api/folio/issues/"+id, nil, nil)
}
