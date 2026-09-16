package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTodos(t *testing.T) {
	t.Run("sends the token and parses the envelope", func(t *testing.T) {
		var gotAuth, gotPath, gotQuery string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			gotPath = r.URL.Path
			gotQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{"todos":[{"id":"t1","project_id":"p1","title":"write it","status":"pending","priority":"high","tags":["x"],"position":1,"depends_on":[],"blocked":false,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}]}`))
		}))
		defer server.Close()

		todos, err := New(server.URL, "tok").ListTodos("folio", TodoFilter{Status: []string{"pending"}, Limit: 5})
		if err != nil {
			t.Fatalf("ListTodos() error = %v", err)
		}

		if gotAuth != "tok" {
			t.Errorf("Authorization = %q, want %q", gotAuth, "tok")
		}
		if gotPath != "/api/folio/projects/folio/todos" {
			t.Errorf("path = %q", gotPath)
		}
		if gotQuery != "limit=5&status=pending" {
			t.Errorf("query = %q, want %q", gotQuery, "limit=5&status=pending")
		}
		if len(todos) != 1 {
			t.Fatalf("len(todos) = %d, want 1", len(todos))
		}
		if todos[0].Title != "write it" || todos[0].Priority != "high" {
			t.Errorf("todo = %+v", todos[0])
		}
	})

	t.Run("reports the API error message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"you do not have access to this resource"}`))
		}))
		defer server.Close()

		_, err := New(server.URL, "tok").ListTodos("folio", TodoFilter{})
		if err == nil {
			t.Fatal("ListTodos() error = nil, want an error")
		}
		if got := err.Error(); got != "403: you do not have access to this resource" {
			t.Errorf("err = %q", got)
		}
	})

	t.Run("refuses to send without a token", func(t *testing.T) {
		_, err := New("http://127.0.0.1:1", "").ListTodos("folio", TodoFilter{})
		if err == nil {
			t.Fatal("ListTodos() error = nil, want a missing-token error")
		}
	})
}

func TestCreateTodo(t *testing.T) {
	var gotBody map[string]any
	var gotMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"t9","project_id":"p1","title":"ship it","status":"pending","priority":"medium","tags":[],"position":1,"depends_on":[],"blocked":false,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	todo, err := New(server.URL, "tok").CreateTodo("folio", TodoInput{Title: strptr("ship it")})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody["title"] != "ship it" {
		t.Errorf("body = %v", gotBody)
	}
	if _, sent := gotBody["status"]; sent {
		t.Error("status was sent; an unset field must be omitted so the server applies its default")
	}
	if todo.ID != "t9" {
		t.Errorf("id = %q", todo.ID)
	}
}

func TestUpdateTodo(t *testing.T) {
	var gotBody map[string]any
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"t1","project_id":"p1","title":"kept","status":"done","priority":"low","tags":[],"position":1,"depends_on":[],"blocked":false,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	todo, err := New(server.URL, "tok").UpdateTodo("t1", TodoInput{Status: strptr("done")})
	if err != nil {
		t.Fatalf("UpdateTodo() error = %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotPath != "/api/folio/todos/t1" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["status"] != "done" {
		t.Errorf("body = %v", gotBody)
	}
	if _, sent := gotBody["title"]; sent {
		t.Error("title was sent; PATCH must omit untouched fields rather than clear them")
	}
	if todo.Status != "done" {
		t.Errorf("status = %q", todo.Status)
	}
}

func TestDeleteTodo(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := New(server.URL, "tok").DeleteTodo("t1"); err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/folio/todos/t1" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestGetTodo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/folio/todos/t1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"t1","project_id":"p1","title":"one","status":"pending","priority":"low","tags":[],"position":1,"depends_on":[],"blocked":true,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`))
	}))
	defer server.Close()

	todo, err := New(server.URL, "tok").GetTodo("t1")
	if err != nil {
		t.Fatalf("GetTodo() error = %v", err)
	}
	if !todo.Blocked {
		t.Error("blocked = false, want true")
	}
}

func strptr(s string) *string { return &s }
