package domain

import "time"

// Knowledge is a durable note that outlives the project it was learned on: a
// tip, a snippet, a fix worth keeping, a page worth not losing.
//
// It sits outside the tenancy every other record obeys. Any authenticated user
// reads and writes all of it, because the things worth keeping here are things
// like how to get realtime working behind a proxy, and copying that into each
// project would be the only alternative.
type Knowledge struct {
	ID KnowledgeID
	// ProjectID is optional and is a note of where this was learned, not a
	// fence around who may read it. A project's deletion detaches its
	// knowledge rather than destroying it.
	ProjectID ProjectID
	Slug      string
	Title     string
	Body      string
	Tags      []string
	// CreatedBy is always set, though nothing filters on it: attribution here
	// tells a reader whom to ask, it does not decide who may look.
	CreatedBy UserID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// KnowledgeFilter narrows a listing. Zero values mean "no restriction", and an
// unset ProjectID therefore lists everything rather than only the unattached
// notes; Unattached asks for those.
type KnowledgeFilter struct {
	ProjectID  ProjectID
	Unattached bool
	Tags       []string
	Search     string
	Limit      int
	Offset     int
}
