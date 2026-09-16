package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListProjects(t *testing.T) {
	var gotPath, gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`{"projects":[{"id":"pr1","slug":"folio","name":"folio","descr":"A workspace","archived":false,"members":[{"user_id":"u1","email":"a@b.c","role":"owner"}],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}]}`))
	}))
	defer server.Close()

	projects, err := New(server.URL, "tok").ListProjects(false)
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if gotPath != "/api/folio/projects" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want none when archived is not requested", gotQuery)
	}
	if len(projects) != 1 || projects[0].Slug != "folio" {
		t.Fatalf("projects = %+v", projects)
	}
	if len(projects[0].Members) != 1 || projects[0].Members[0].Role != "owner" {
		t.Errorf("members = %+v", projects[0].Members)
	}
}

func TestListProjectsIncludingArchived(t *testing.T) {
	var gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"projects":[]}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, "tok").ListProjects(true); err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if gotQuery != "archived=true" {
		t.Errorf("query = %q, want archived=true", gotQuery)
	}
}

func TestGetProjectAcceptsAnIDOrSlug(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"pr1","slug":"folio","name":"folio","archived":false,"members":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	project, err := New(server.URL, "tok").GetProject("folio")
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}
	if gotPath != "/api/folio/projects/folio" {
		t.Errorf("path = %q", gotPath)
	}
	if project.Slug != "folio" {
		t.Errorf("slug = %q", project.Slug)
	}
}

func TestCreateProject(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"pr9","slug":"search-rewrite","name":"Search Rewrite","archived":false,"members":[{"user_id":"u1","role":"owner"}],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	project, err := New(server.URL, "tok").CreateProject(CreateProjectInput{Name: "Search Rewrite"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/folio/projects" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
	if gotBody["name"] != "Search Rewrite" {
		t.Errorf("body = %v", gotBody)
	}
	if _, sent := gotBody["slug"]; sent {
		t.Error("slug was sent; the server derives it from the name when omitted")
	}
	if project.Slug != "search-rewrite" {
		t.Errorf("slug = %q", project.Slug)
	}
}

func TestCreateProjectSendsAnExplicitSlug(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"pr9","slug":"custom","name":"Name","archived":false,"members":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, "tok").CreateProject(CreateProjectInput{Name: "Name", Slug: "custom"}); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if gotBody["slug"] != "custom" {
		t.Errorf("slug = %v, want custom", gotBody["slug"])
	}
}

func TestUpdateProject(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"pr1","slug":"folio","name":"Renamed","archived":true,"members":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	archived := true
	project, err := New(server.URL, "tok").UpdateProject("folio", ProjectInput{Archived: &archived})
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/api/folio/projects/folio" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
	if gotBody["archived"] != true {
		t.Errorf("body = %v", gotBody)
	}
	if _, sent := gotBody["name"]; sent {
		t.Error("name was sent; PATCH must omit untouched fields")
	}
	if !project.Archived {
		t.Error("expected the archived project back")
	}
}

func TestDeleteProject(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := New(server.URL, "tok").DeleteProject("folio"); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/folio/projects/folio" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
}

func TestAddMember(t *testing.T) {
	var gotPath string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"pr1","slug":"folio","name":"folio","archived":false,"members":[{"user_id":"u1","role":"owner"},{"user_id":"u2","email":"b@c.d","role":"editor"}],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	project, err := New(server.URL, "tok").AddMember("folio", "b@c.d", "editor")
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}
	if gotPath != "/api/folio/projects/folio/members" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["email"] != "b@c.d" || gotBody["role"] != "editor" {
		t.Errorf("body = %v", gotBody)
	}
	if len(project.Members) != 2 {
		t.Errorf("members = %d, want 2", len(project.Members))
	}
}

func TestRemoveMember(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"id":"pr1","slug":"folio","name":"folio","archived":false,"members":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, "tok").RemoveMember("folio", "u2"); err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/folio/projects/folio/members/u2" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
}

func TestSetMemberRole(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"pr1","slug":"folio","name":"folio","archived":false,"members":[],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, "tok").SetMemberRole("folio", "u2", "owner"); err != nil {
		t.Fatalf("SetMemberRole() error = %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/api/folio/projects/folio/members/u2" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
	if gotBody["role"] != "owner" {
		t.Errorf("body = %v", gotBody)
	}
}

func TestGetProjectReportsAnUnknownSlug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, "tok").GetProject("nope"); err == nil {
		t.Fatal("GetProject() error = nil, want an error")
	} else if err.Error() != "404: not found" {
		t.Errorf("err = %q", err.Error())
	}
}
