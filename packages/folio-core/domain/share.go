package domain

import "time"

// Share is a read-only link to one issue or one plan for someone without an
// account. The token is the whole credential, so there is no expiry: a link
// stops working only when it is revoked, which deletes it.
type Share struct {
	ID        ShareID
	ProjectID ProjectID
	// Exactly one of IssueID and PlanID is set.
	IssueID IssueID
	PlanID  PlanID
	// Label records who the link was given to, so a list of links can be
	// pruned without guessing.
	Label          string
	Token          string
	CreatedBy      UserID
	CreatedAt      time.Time
	LastAccessedAt *time.Time
}

// SharedItem is what a share link opens. Brief is set for an issue share;
// Plan and Todos for a plan share.
type SharedItem struct {
	Share Share
	Brief *IssueBrief
	Plan  *Plan
	Todos []Issue
}
