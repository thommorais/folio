package ports

import (
	"context"
	"time"

	"folio/folio-core/domain"
)

// Clock is injected so services never call time.Now directly, which keeps
// timestamp behaviour deterministic under test.
type Clock interface {
	Now() time.Time
}

// IDGenerator produces storage-compatible identifiers. Services generate IDs
// up front rather than letting storage assign them, so a created record has a
// stable ID before it is written.
type IDGenerator interface {
	NewID() string
}

// TokenGenerator produces share-link secrets. It is separate from IDGenerator
// because a record ID is not secret and is far too short to be one.
type TokenGenerator interface {
	NewToken() string
}

type Logger interface {
	Debug(msg string, fields map[string]any)
	Info(msg string, fields map[string]any)
	Warn(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

// Actor is the authenticated caller. Adapters build it from whatever their
// transport authenticated with; services only ever see this.
type Actor struct {
	UserID domain.UserID
	Email  string
	// Superuser bypasses project membership checks. PocketBase superusers get
	// it; regular tokens never do.
	Superuser bool
}

// Guard resolves what an actor may do with a project. Services call it before
// every operation so the permission rule lives in exactly one place.
type Guard interface {
	// EnsureRead returns the project when the actor may read it, else
	// domain.ErrForbidden (or domain.ErrNotFound if it does not exist).
	EnsureRead(ctx context.Context, actor Actor, id domain.ProjectID) (domain.Project, error)
	EnsureWrite(ctx context.Context, actor Actor, id domain.ProjectID) (domain.Project, error)
	EnsureAdmin(ctx context.Context, actor Actor, id domain.ProjectID) (domain.Project, error)
}
