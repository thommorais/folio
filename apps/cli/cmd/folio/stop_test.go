package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func stopServer(t *testing.T) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{method: r.Method, path: r.URL.Path}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		got = append(got, req)

		switch r.URL.Path {
		case "/api/folio/issues/tk1":
			_, _ = w.Write([]byte(`{"id":"tk1","project_id":"pr1","kind":"ticket","tags":[],"depends_on":[]}`))
		case "/api/folio/projects/pr1/entries":
			_, _ = w.Write([]byte(`{"id":"e1","kind":"handoff"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not found."}`))
		}
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "")

	return &got
}

func stopRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "--initial-branch=feat/landing"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"commit", "--allow-empty", "-m", "first"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	t.Chdir(dir)
	return dir
}

func runStop(t *testing.T, args ...string) error {
	t.Helper()

	return silenced(t, func() error {
		cmd := stopCommand()
		cmd.SetArgs(args)
		return cmd.Execute()
	})
}

func TestStopWritesAHandoffWithTheCommit(t *testing.T) {
	got := stopServer(t)
	dir := stopRepo(t)
	head, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}

	if err := runStop(t, "tk1", "Next: wire the coverage widget."); err != nil {
		t.Fatal(err)
	}

	if len(*got) != 2 {
		t.Fatalf("requests = %+v, want a ticket lookup then the write", *got)
	}
	write := (*got)[1]
	if write.method != http.MethodPost || write.path != "/api/folio/projects/pr1/entries" {
		t.Fatalf("write = %+v", write)
	}
	body := write.body
	if body["kind"] != "handoff" || body["issue_id"] != "tk1" || body["body"] != "Next: wire the coverage widget." {
		t.Errorf("body = %v", body)
	}
	if body["branch"] != "feat/landing" {
		t.Errorf("branch = %v, want feat/landing", body["branch"])
	}
	meta, _ := body["meta"].(map[string]any)
	if meta["commit"] != string(head[:40]) {
		t.Errorf("meta = %v, want the HEAD sha", meta)
	}
}

func TestStopRefusesUncommittedWork(t *testing.T) {
	got := stopServer(t)
	dir := stopRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "wip.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runStop(t, "tk1", "Next: finish it.")
	if err == nil {
		t.Fatal("want a refusal on a dirty tree")
	}
	if code := exitCode(err); code != exitValidation {
		t.Errorf("exit = %d, want %d", code, exitValidation)
	}
	for _, r := range *got {
		if r.method == http.MethodPost {
			t.Fatalf("wrote %+v despite uncommitted work", r)
		}
	}
}

func TestStopAllowsUncommittedWorkWhenAsked(t *testing.T) {
	got := stopServer(t)
	dir := stopRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "wip.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runStop(t, "tk1", "Next: finish it.", "--allow-dirty"); err != nil {
		t.Fatal(err)
	}

	meta, _ := (*got)[len(*got)-1].body["meta"].(map[string]any)
	dirty, _ := meta["uncommitted"].([]any)
	if len(dirty) != 1 || dirty[0] != "wip.go" {
		t.Errorf("meta = %v, want the uncommitted paths recorded", meta)
	}
}

func TestStopNeedsANote(t *testing.T) {
	stopServer(t)
	stopRepo(t)

	if err := runStop(t, "tk1"); err == nil {
		t.Fatal("want a refusal without a note")
	}
}
