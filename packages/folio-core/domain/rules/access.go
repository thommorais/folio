package rules

import (
	"slices"

	"folio/folio-core/domain"
)

var roleRank = map[domain.Role]int{
	domain.RoleViewer: 1,
	domain.RoleEditor: 2,
	domain.RoleOwner:  3,
}

func EffectiveRole(access domain.ProjectAccess, user domain.UserID) (domain.Role, bool) {
	if slices.Contains(access.ClientOwners, user) {
		return domain.RoleOwner, true
	}

	var best domain.Role
	raise := func(r domain.Role) {
		if roleRank[r] > roleRank[best] {
			best = r
		}
	}
	for _, g := range access.DomainGrants {
		if slices.Contains(g.Members, user) {
			raise(g.Role)
		}
	}
	for _, m := range access.Members {
		if m.UserID == user {
			raise(m.Role)
		}
	}
	return best, best != ""
}

func CanCreateProjectIn(d domain.Domain, user domain.UserID) bool {
	if slices.Contains(d.ClientOwners, user) {
		return true
	}
	return slices.ContainsFunc(d.Members, func(m domain.Member) bool { return m.UserID == user })
}

func CanAdminClient(c domain.Client, user domain.UserID) bool {
	return slices.Contains(c.Owners, user)
}
