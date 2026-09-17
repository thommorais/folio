package services

import (
	"context"
	"strings"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

const DefaultRecentJournal = 10

// issueScope resolves the issue a child entity is being attached to. The
// project guard alone is not enough: a caller with write access to project A
// could otherwise name an issue in project B and pull the child across the
// tenancy boundary. An empty id means "no issue", which is always legal.
func issueScope(ctx context.Context, issues ports.IssueRepository, id domain.IssueID, project domain.ProjectID) error {
	if id == "" {
		return nil
	}
	issue, err := issues.GetByID(ctx, id)
	if err != nil {
		if notFound(err) {
			return domain.Invalid("issue", "does not exist")
		}
		return err
	}
	if issue.ProjectID != project {
		return domain.Invalid("issue", "belongs to a different project")
	}
	return nil
}

func defaultPriority(p domain.Priority) domain.Priority {
	if p == "" {
		return domain.PriorityMedium
	}
	return p
}

// parseDue accepts RFC 3339, treating an empty string as "clear the date" and
// nil as "leave it alone".
func parseDue(raw *string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return nil, domain.Invalid("due_date", "must be an RFC 3339 timestamp, e.g. 2026-12-01T09:00:00Z")
	}
	return &t, nil
}
