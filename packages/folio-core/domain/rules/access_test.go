package rules_test

import (
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

const u domain.UserID = "u"

func TestEffectiveRole(t *testing.T) {
	cases := []struct {
		name   string
		access domain.ProjectAccess
		want   domain.Role
		member bool
	}{
		{
			name:   "no grant reaches the user",
			access: domain.ProjectAccess{},
		},
		{
			name:   "a client owner owns every project",
			access: domain.ProjectAccess{ClientOwners: []domain.UserID{u}},
			want:   domain.RoleOwner, member: true,
		},
		{
			name: "a domain member gets the domain's grant",
			access: domain.ProjectAccess{DomainGrants: []domain.DomainGrant{
				{DomainID: "d", Role: domain.RoleEditor, Members: []domain.UserID{u}},
			}},
			want: domain.RoleEditor, member: true,
		},
		{
			name: "another domain's grant does not reach the user",
			access: domain.ProjectAccess{DomainGrants: []domain.DomainGrant{
				{DomainID: "d", Role: domain.RoleEditor, Members: []domain.UserID{"other"}},
			}},
		},
		{
			name:   "a personal grant alone is enough",
			access: domain.ProjectAccess{Members: []domain.Member{{UserID: u, Role: domain.RoleViewer}}},
			want:   domain.RoleViewer, member: true,
		},
		{
			name: "a personal grant raises a domain grant",
			access: domain.ProjectAccess{
				DomainGrants: []domain.DomainGrant{{DomainID: "d", Role: domain.RoleViewer, Members: []domain.UserID{u}}},
				Members:      []domain.Member{{UserID: u, Role: domain.RoleEditor}},
			},
			want: domain.RoleEditor, member: true,
		},
		{
			name: "a personal grant does not lower a domain grant",
			access: domain.ProjectAccess{
				DomainGrants: []domain.DomainGrant{{DomainID: "d", Role: domain.RoleEditor, Members: []domain.UserID{u}}},
				Members:      []domain.Member{{UserID: u, Role: domain.RoleViewer}},
			},
			want: domain.RoleEditor, member: true,
		},
		{
			name: "the highest of several domain grants wins",
			access: domain.ProjectAccess{DomainGrants: []domain.DomainGrant{
				{DomainID: "a", Role: domain.RoleViewer, Members: []domain.UserID{u}},
				{DomainID: "b", Role: domain.RoleOwner, Members: []domain.UserID{u}},
				{DomainID: "c", Role: domain.RoleEditor, Members: []domain.UserID{u}},
			}},
			want: domain.RoleOwner, member: true,
		},
		{
			name: "client ownership outranks a lower personal grant",
			access: domain.ProjectAccess{
				ClientOwners: []domain.UserID{u},
				Members:      []domain.Member{{UserID: u, Role: domain.RoleViewer}},
			},
			want: domain.RoleOwner, member: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, member := rules.EffectiveRole(tc.access, u)
			if got != tc.want || member != tc.member {
				t.Errorf("EffectiveRole = (%q, %v), want (%q, %v)", got, member, tc.want, tc.member)
			}
		})
	}
}
