// Package git reads metadata a journal entry can infer from the working
// directory. Every lookup answers with an empty string rather than an error,
// so a missing repo, a detached HEAD or an absent gh never fails the write it
// is decorating.
package git

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

// gh pr view is an HTTPS call, and an installed but offline or expired gh
// hangs rather than failing, blocking the write behind it.
const ghTimeout = 2 * time.Second

// Branch is the checked out branch, or empty outside a repository and on a
// detached HEAD.
func Branch(dir string) string {
	branch := run(context.Background(), dir, "git", "rev-parse", "--abbrev-ref", "HEAD")

	// A detached HEAD exits zero with the literal string "HEAD".
	if branch == "HEAD" {
		return ""
	}
	return branch
}

// PR is the url of the pull request open for the current branch, stored in
// full so an entry stays resolvable when it is read outside the repository it
// was written in. gh is optional enrichment rather than a dependency, so
// anything short of an answer is empty.
func PR(dir string) string {
	if _, err := exec.LookPath("gh"); err != nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), ghTimeout)
	defer cancel()

	out := run(ctx, dir, "gh", "pr", "view", "--json", "url")
	if out == "" {
		return ""
	}

	var view struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(out), &view); err != nil {
		return ""
	}
	return view.URL
}

// run returns trimmed stdout, or empty for any failure at all.
func run(ctx context.Context, dir, name string, args ...string) string {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
