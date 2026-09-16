package domain

import "time"

// Role is a member's permission level on a project. Every project has at least
// one owner; the API refuses to remove the last one.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// Member is a user's membership in a project.
type Member struct {
	UserID UserID
	Role   Role
	Email  string
	Name   string
}

// CanWrite reports whether the role may mutate project content.
func (r Role) CanWrite() bool {
	return r == RoleOwner || r == RoleEditor
}

// CanAdmin reports whether the role may change membership or delete the project.
func (r Role) CanAdmin() bool {
	return r == RoleOwner
}

type Project struct {
	ID        ProjectID
	Slug      string
	Name      string
	Descr     string
	Archived  bool
	Members   []Member
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RoleOf returns the role the given user holds on the project.
func (p Project) RoleOf(user UserID) (Role, bool) {
	for _, m := range p.Members {
		if m.UserID == user {
			return m.Role, true
		}
	}
	return "", false
}

// Owners returns every member holding the owner role.
func (p Project) Owners() []Member {
	out := make([]Member, 0, 1)
	for _, m := range p.Members {
		if m.Role == RoleOwner {
			out = append(out, m)
		}
	}
	return out
}
