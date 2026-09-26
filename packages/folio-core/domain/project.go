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
	ID       ProjectID
	DomainID DomainID
	Slug     string
	Name     string
	Descr    string
	Archived bool

	ClientOwners []UserID
	DomainGrants []DomainGrant
	Members      []Member

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Client is a company or a person.
type Client struct {
	ID        ClientID
	Slug      string
	Name      string
	Site      string
	Logo      string
	Descr     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Domain is the root of permissions: its members reach every project under it.
type Domain struct {
	ID        DomainID
	ClientID  ClientID
	Slug      string
	Name      string
	Descr     string
	Members   []Member
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p Project) Access() ProjectAccess {
	return ProjectAccess{ClientOwners: p.ClientOwners, DomainGrants: p.DomainGrants, Members: p.Members}
}

// RoleOf returns the user's personal grant, ignoring client and domain grants.
func (p Project) RoleOf(user UserID) (Role, bool) {
	for _, m := range p.Members {
		if m.UserID == user {
			return m.Role, true
		}
	}
	return "", false
}

// Owners returns every personal grant holding the owner role.
func (p Project) Owners() []Member {
	out := make([]Member, 0, 1)
	for _, m := range p.Members {
		if m.Role == RoleOwner {
			out = append(out, m)
		}
	}
	return out
}
