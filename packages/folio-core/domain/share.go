package domain

import "time"

type Share struct {
	ID             ShareID
	ProjectID      ProjectID
	IssueID        IssueID
	PlanID         PlanID
	Label          string
	Token          string
	CreatedBy      UserID
	CreatedAt      time.Time
	LastAccessedAt *time.Time
}

type SharedItem struct {
	Share Share
	Brief *IssueBrief
	Plan  *Plan
	Todos []Issue
}
