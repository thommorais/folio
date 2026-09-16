package client

import "net/http"

type Member struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Role   string `json:"role"`
}

type Project struct {
	ID        string   `json:"id"`
	Slug      string   `json:"slug"`
	Name      string   `json:"name"`
	Descr     string   `json:"descr,omitempty"`
	Archived  bool     `json:"archived"`
	Members   []Member `json:"members"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func (c *Client) ListProjects(includeArchived bool) ([]Project, error) {
	path := "/api/folio/projects"
	if includeArchived {
		path += "?archived=true"
	}

	var body struct {
		Projects []Project `json:"projects"`
	}
	if err := c.do(http.MethodGet, path, nil, &body); err != nil {
		return nil, err
	}
	return body.Projects, nil
}

func (c *Client) GetProject(ref string) (Project, error) {
	var project Project
	err := c.do(http.MethodGet, "/api/folio/projects/"+ref, nil, &project)
	return project, err
}

type CreateProjectInput struct {
	Slug  string `json:"slug,omitempty"`
	Name  string `json:"name"`
	Descr string `json:"descr,omitempty"`
}

type ProjectInput struct {
	Name     *string `json:"name,omitempty"`
	Descr    *string `json:"descr,omitempty"`
	Archived *bool   `json:"archived,omitempty"`
}

func (c *Client) CreateProject(in CreateProjectInput) (Project, error) {
	var project Project
	err := c.do(http.MethodPost, "/api/folio/projects", in, &project)
	return project, err
}

func (c *Client) UpdateProject(ref string, in ProjectInput) (Project, error) {
	var project Project
	err := c.do(http.MethodPatch, "/api/folio/projects/"+ref, in, &project)
	return project, err
}

func (c *Client) DeleteProject(ref string) error {
	return c.do(http.MethodDelete, "/api/folio/projects/"+ref, nil, nil)
}

func (c *Client) AddMember(ref, email, role string) (Project, error) {
	var project Project
	body := struct {
		Email string `json:"email"`
		Role  string `json:"role,omitempty"`
	}{Email: email, Role: role}
	err := c.do(http.MethodPost, "/api/folio/projects/"+ref+"/members", body, &project)
	return project, err
}

func (c *Client) SetMemberRole(ref, user, role string) (Project, error) {
	var project Project
	body := struct {
		Role string `json:"role"`
	}{Role: role}
	err := c.do(http.MethodPatch, "/api/folio/projects/"+ref+"/members/"+user, body, &project)
	return project, err
}

func (c *Client) RemoveMember(ref, user string) (Project, error) {
	var project Project
	err := c.do(http.MethodDelete, "/api/folio/projects/"+ref+"/members/"+user, nil, &project)
	return project, err
}
