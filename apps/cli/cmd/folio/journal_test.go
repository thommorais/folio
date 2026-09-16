package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
)

const writtenEntry = `{"id":"l1","project_id":"p1","title":"shipped it","slug":"shipped-it","tags":["cli"],"created_at":"2026-09-11T01:07:01Z","updated_at":"2026-09-11T01:07:01Z"}`

// journalServer records what the command sent, which is where a defaulted
// field shows up or fails to.
func journalServer(t *testing.T) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{method: r.Method, path: r.URL.Path}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		got = append(got, req)
		_, _ = w.Write([]byte(writtenEntry))
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "p1")

	return &got
}

// inRepo runs the rest of the test from inside a fresh git repository on the
// named branch, so detection has something real to read.
func inRepo(t *testing.T, branch string) {
	t.Helper()

	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "--initial-branch=" + branch},
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
}

// runJournal executes the journal tree, returning what reached stderr so a
// test can assert detection stays silent.
func runJournal(t *testing.T, args ...string) (string, error) {
	t.Helper()

	stdout := os.Stdout
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	os.Stdout = devnull

	stderr := os.Stderr
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = write

	t.Cleanup(func() {
		os.Stdout = stdout
		os.Stderr = stderr
		_ = devnull.Close()
	})

	cmd := journalCommand()
	cmd.SetOut(devnull)
	cmd.SetErr(write)
	cmd.SetArgs(args)
	runErr := cmd.Execute()

	_ = write.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, read)
	return buf.String(), runErr
}

func TestJournalWriteDefaultsBranch(t *testing.T) {
	t.Run("fills the branch from the checkout", func(t *testing.T) {
		inRepo(t, "feature-x")
		got := journalServer(t)

		if _, err := runJournal(t, "write", "shipped it"); err != nil {
			t.Fatalf("journal write: %v", err)
		}

		if len(*got) != 1 {
			t.Fatalf("requests = %d, want 1", len(*got))
		}
		if branch := (*got)[0].body["branch"]; branch != "feature-x" {
			t.Errorf("branch = %v, want feature-x", branch)
		}
	})

	t.Run("an explicit flag beats detection", func(t *testing.T) {
		inRepo(t, "feature-x")
		got := journalServer(t)

		if _, err := runJournal(t, "write", "shipped it", "--branch", "release"); err != nil {
			t.Fatalf("journal write: %v", err)
		}

		if branch := (*got)[0].body["branch"]; branch != "release" {
			t.Errorf("branch = %v, want release", branch)
		}
	})

	t.Run("an explicit empty branch suppresses detection", func(t *testing.T) {
		// Passing --branch "" is the only way to say "no branch" from inside a
		// checkout, so it has to reach the wire as empty rather than being
		// read as an absent flag and refilled.
		inRepo(t, "feature-x")
		got := journalServer(t)

		if _, err := runJournal(t, "write", "shipped it", "--branch", ""); err != nil {
			t.Fatalf("journal write: %v", err)
		}

		if branch, ok := (*got)[0].body["branch"]; ok && branch != "" {
			t.Errorf("branch = %v, want empty", branch)
		}
	})

	t.Run("succeeds outside a repository with no branch", func(t *testing.T) {
		t.Chdir(t.TempDir())
		got := journalServer(t)

		if _, err := runJournal(t, "write", "shipped it"); err != nil {
			t.Fatalf("journal write outside a repo: %v", err)
		}

		if branch, ok := (*got)[0].body["branch"]; ok && branch != "" {
			t.Errorf("branch = %v, want no branch outside a repo", branch)
		}
	})

	t.Run("says nothing on stderr about what it filled", func(t *testing.T) {
		inRepo(t, "feature-x")
		journalServer(t)

		stderr, err := runJournal(t, "write", "shipped it")
		if err != nil {
			t.Fatalf("journal write: %v", err)
		}
		if stderr != "" {
			t.Errorf("stderr = %q, want silence: detection does not announce itself", stderr)
		}
	})
}

func TestJournalUpdateDoesNotDefault(t *testing.T) {
	// Filling branch on an update would overwrite what an entry recorded with
	// wherever the shell happens to be now, rewriting history rather than
	// filling a gap.
	inRepo(t, "feature-x")
	got := journalServer(t)

	if _, err := runJournal(t, "update", "l1", "--title", "renamed"); err != nil {
		t.Fatalf("journal update: %v", err)
	}

	if len(*got) != 1 {
		t.Fatalf("requests = %d, want 1", len(*got))
	}
	if _, ok := (*got)[0].body["branch"]; ok {
		t.Errorf("body = %v, want no branch: update must not default", (*got)[0].body)
	}
}
