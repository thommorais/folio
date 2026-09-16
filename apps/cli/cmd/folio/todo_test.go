package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// request captures what a command actually sent, which is the only thing that
// distinguishes a shortcut from the update call it is meant to match.
type request struct {
	method string
	path   string
	body   map[string]any
}

// todoServer stands in for the API and records every request. The todo it
// returns is the same regardless of the patch, since these tests assert on
// what was sent rather than on what came back.
func todoServer(t *testing.T, todo string) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{method: r.Method, path: r.URL.Path}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		got = append(got, req)
		_, _ = w.Write([]byte(todo))
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")

	return &got
}

// run executes the todo command tree with stdout discarded, so a test reads
// the recorded requests rather than the rendered table.
func run(t *testing.T, args ...string) error {
	t.Helper()

	stdout := os.Stdout
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	os.Stdout = devnull
	t.Cleanup(func() {
		os.Stdout = stdout
		_ = devnull.Close()
	})

	cmd := todoCommand()
	cmd.SetOut(devnull)
	cmd.SetErr(devnull)
	cmd.SetArgs(args)
	return cmd.Execute()
}

const pendingTodo = `{"id":"t1","project_id":"p1","title":"write it","status":"pending","priority":"high","tags":["cli"],"position":1,"depends_on":[],"blocked":false,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`

func TestTodoStatusShortcuts(t *testing.T) {
	// The plan keeps --status alongside the shortcuts, so the two paths have
	// to stay in agreement: same method, same path, same patch.
	shortcuts := map[string]string{
		"start":  "in_progress",
		"done":   "done",
		"cancel": "cancelled",
	}

	for name, status := range shortcuts {
		t.Run(name+" patches the status", func(t *testing.T) {
			got := todoServer(t, pendingTodo)

			if err := run(t, name, "t1"); err != nil {
				t.Fatalf("todo %s: %v", name, err)
			}

			if len(*got) != 1 {
				t.Fatalf("requests = %d, want 1", len(*got))
			}
			req := (*got)[0]
			if req.method != http.MethodPatch {
				t.Errorf("method = %q, want PATCH", req.method)
			}
			if req.path != "/api/folio/todos/t1" {
				t.Errorf("path = %q", req.path)
			}
			if len(req.body) != 1 || req.body["status"] != status {
				t.Errorf("body = %v, want only status=%q", req.body, status)
			}
		})

		t.Run(name+" matches the equivalent update", func(t *testing.T) {
			viaShortcut := todoServer(t, pendingTodo)
			if err := run(t, name, "t1"); err != nil {
				t.Fatalf("todo %s: %v", name, err)
			}

			viaUpdate := todoServer(t, pendingTodo)
			if err := run(t, "update", "t1", "--status", status); err != nil {
				t.Fatalf("todo update --status %s: %v", status, err)
			}

			if len(*viaShortcut) != 1 || len(*viaUpdate) != 1 {
				t.Fatalf("requests = %d and %d, want 1 each", len(*viaShortcut), len(*viaUpdate))
			}
			if (*viaShortcut)[0].method != (*viaUpdate)[0].method ||
				(*viaShortcut)[0].path != (*viaUpdate)[0].path ||
				(*viaShortcut)[0].body["status"] != (*viaUpdate)[0].body["status"] {
				t.Errorf("shortcut %+v != update %+v", (*viaShortcut)[0], (*viaUpdate)[0])
			}
		})

		t.Run(name+" takes exactly one id", func(t *testing.T) {
			got := todoServer(t, pendingTodo)

			if err := run(t, name, "t1", "t2"); err == nil {
				t.Error("two ids: error = nil, want an arity error")
			}
			if err := run(t, name); err == nil {
				t.Error("no id: error = nil, want an arity error")
			}
			if len(*got) != 0 {
				t.Errorf("requests = %d, want 0: a rejected call must not reach the API", len(*got))
			}
		})
	}
}

func TestTodoBlock(t *testing.T) {
	// block writes DependsOn rather than the blocked status, so Blocked stays
	// derived from the dependency the way domain/todo.go describes.
	t.Run("appends to depends_on rather than the status", func(t *testing.T) {
		got := todoServer(t, `{"id":"t1","project_id":"p1","title":"write it","status":"pending","priority":"high","tags":["cli"],"position":1,"depends_on":["t9"],"blocked":true,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`)

		if err := run(t, "block", "t1", "--on", "t2"); err != nil {
			t.Fatalf("todo block: %v", err)
		}

		if len(*got) != 2 {
			t.Fatalf("requests = %d, want a GET then a PATCH", len(*got))
		}
		if (*got)[0].method != http.MethodGet {
			t.Errorf("first method = %q, want GET: block must read the current set before replacing it", (*got)[0].method)
		}

		patch := (*got)[1]
		if patch.method != http.MethodPatch {
			t.Fatalf("second method = %q, want PATCH", patch.method)
		}
		if _, ok := patch.body["status"]; ok {
			t.Errorf("body = %v, want no status: block records a dependency, not the status", patch.body)
		}

		deps, ok := patch.body["depends_on"].([]any)
		if !ok {
			t.Fatalf("depends_on = %v, want a list", patch.body["depends_on"])
		}
		if len(deps) != 2 || deps[0] != "t9" || deps[1] != "t2" {
			t.Errorf("depends_on = %v, want the existing t9 kept and t2 appended", deps)
		}
	})

	t.Run("does not duplicate a blocker already recorded", func(t *testing.T) {
		got := todoServer(t, `{"id":"t1","project_id":"p1","title":"write it","status":"pending","priority":"high","tags":["cli"],"position":1,"depends_on":["t2"],"blocked":true,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`)

		if err := run(t, "block", "t1", "--on", "t2"); err != nil {
			t.Fatalf("todo block: %v", err)
		}

		deps := (*got)[1].body["depends_on"].([]any)
		if len(deps) != 1 || deps[0] != "t2" {
			t.Errorf("depends_on = %v, want t2 once", deps)
		}
	})

	t.Run("rejects a todo blocking itself", func(t *testing.T) {
		got := todoServer(t, pendingTodo)

		err := run(t, "block", "t1", "--on", "t1")
		if err == nil {
			t.Fatal("error = nil, want a self-dependency error")
		}
		if !strings.Contains(err.Error(), "cannot block itself") {
			t.Errorf("err = %q", err)
		}
		if len(*got) != 0 {
			t.Errorf("requests = %d, want 0", len(*got))
		}
	})

	t.Run("--off drops one blocker and keeps the rest", func(t *testing.T) {
		got := todoServer(t, `{"id":"t1","project_id":"p1","title":"write it","status":"pending","priority":"high","tags":["cli"],"position":1,"depends_on":["t2","t3"],"blocked":true,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`)

		if err := run(t, "block", "t1", "--off", "t2"); err != nil {
			t.Fatalf("todo block --off: %v", err)
		}

		deps, ok := (*got)[1].body["depends_on"].([]any)
		if !ok {
			t.Fatalf("depends_on = %v, want a list", (*got)[1].body["depends_on"])
		}
		if len(deps) != 1 || deps[0] != "t3" {
			t.Errorf("depends_on = %v, want t2 dropped and t3 kept", deps)
		}
	})

	t.Run("--off the last blocker sends an empty list", func(t *testing.T) {
		// The empty case is the one that has to reach the wire: it is what
		// takes a todo back to unblocked, and an omitted field would instead
		// leave the dependency in place.
		got := todoServer(t, `{"id":"t1","project_id":"p1","title":"write it","status":"pending","priority":"high","tags":["cli"],"position":1,"depends_on":["t2"],"blocked":true,"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`)

		if err := run(t, "block", "t1", "--off", "t2"); err != nil {
			t.Fatalf("todo block --off: %v", err)
		}

		raw, ok := (*got)[1].body["depends_on"]
		if !ok {
			t.Fatal("depends_on absent from the patch, want an empty list")
		}
		if deps := raw.([]any); len(deps) != 0 {
			t.Errorf("depends_on = %v, want empty", deps)
		}
	})

	t.Run("requires exactly one of --on or --off", func(t *testing.T) {
		got := todoServer(t, pendingTodo)

		if err := run(t, "block", "t1"); err == nil {
			t.Error("neither flag: error = nil, want an error")
		}
		if err := run(t, "block", "t1", "--on", "t2", "--off", "t3"); err == nil {
			t.Error("both flags: error = nil, want an error")
		}
		if len(*got) != 0 {
			t.Errorf("requests = %d, want 0", len(*got))
		}
	})
}
