package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// repo builds a real git repository in a temp dir. The helper shells out to
// git, so a fake would test the fake rather than the parsing.
func repo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "--initial-branch=main"},
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
	return dir
}

func TestBranch(t *testing.T) {
	t.Run("reads the checked out branch", func(t *testing.T) {
		if got := Branch(repo(t)); got != "main" {
			t.Errorf("Branch = %q, want main", got)
		}
	})

	t.Run("is empty outside a repository", func(t *testing.T) {
		if got := Branch(t.TempDir()); got != "" {
			t.Errorf("Branch = %q, want empty outside a repo", got)
		}
	})

	t.Run("is empty on a detached HEAD", func(t *testing.T) {
		// git rev-parse --abbrev-ref prints the literal string HEAD here
		// rather than failing, so storing it would persist "HEAD" as if it
		// were a branch name.
		dir := repo(t)
		checkout := exec.Command("git", "checkout", "--detach", "HEAD")
		checkout.Dir = dir
		if out, err := checkout.CombinedOutput(); err != nil {
			t.Fatalf("git checkout --detach: %v\n%s", err, out)
		}

		if got := Branch(dir); got != "" {
			t.Errorf("Branch = %q, want empty on a detached HEAD", got)
		}
	})

	t.Run("is empty when the directory does not exist", func(t *testing.T) {
		if got := Branch(filepath.Join(t.TempDir(), "gone")); got != "" {
			t.Errorf("Branch = %q, want empty", got)
		}
	})
}

func TestPRValue(t *testing.T) {
	// gh is stubbed because the real one needs a GitHub remote and auth, and
	// what is under test is which field of its output is stored.
	t.Run("stores the url rather than the number", func(t *testing.T) {
		dir := repo(t)
		stubGH(t, `{"number":5282,"url":"https://github.com/o/r/pull/5282"}`, 0)

		if got := PR(dir); got != "https://github.com/o/r/pull/5282" {
			t.Errorf("PR = %q, want the url", got)
		}
	})

	t.Run("is empty when the branch has no pull request", func(t *testing.T) {
		// gh exits non-zero with a message on stderr in this case.
		dir := repo(t)
		stubGH(t, "", 1)

		if got := PR(dir); got != "" {
			t.Errorf("PR = %q, want empty", got)
		}
	})

	t.Run("is empty when gh answers with no url", func(t *testing.T) {
		dir := repo(t)
		stubGH(t, `{"number":5282}`, 0)

		if got := PR(dir); got != "" {
			t.Errorf("PR = %q, want empty", got)
		}
	})
}

// stubGH puts a fake gh at the front of PATH, printing out and exiting with
// code. git has to stay reachable, so the real PATH is kept behind it.
func stubGH(t *testing.T, out string, code int) {
	t.Helper()

	dir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s' " + strconv.Quote(out) + "\nexit " + strconv.Itoa(code) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0o755); err != nil {
		t.Fatalf("write gh stub: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestPR(t *testing.T) {
	t.Run("is empty when gh is not installed", func(t *testing.T) {
		// PATH without gh is the CI case: a missing binary must read as no PR
		// rather than as a failure. The repo is built first, since emptying
		// PATH would otherwise hide git from the helper too.
		dir := repo(t)
		t.Setenv("PATH", t.TempDir())

		if got := PR(dir); got != "" {
			t.Errorf("PR = %q, want empty without gh", got)
		}
	})

	t.Run("is empty outside a repository", func(t *testing.T) {
		if got := PR(t.TempDir()); got != "" {
			t.Errorf("PR = %q, want empty outside a repo", got)
		}
	})
}
