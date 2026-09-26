package rules_test

import (
	"errors"
	"strings"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func TestValidateDomain(t *testing.T) {
	if err := rules.ValidateDomain(domain.Domain{Slug: "web", Name: "Web"}); err != nil {
		t.Fatalf("valid domain rejected: %v", err)
	}
	for name, d := range map[string]domain.Domain{
		"bad slug":      {Slug: "Web Team", Name: "Web"},
		"empty name":    {Slug: "web", Name: ""},
		"long describe": {Slug: "web", Name: "Web", Descr: strings.Repeat("d", rules.DescrMaxLen+1)},
	} {
		if err := rules.ValidateDomain(d); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("%s: want validation error, got %v", name, err)
		}
	}
}

func TestCanAdminDomain(t *testing.T) {
	d := domain.Domain{
		ClientOwners: []domain.UserID{"boss"},
		Members: []domain.Member{
			{UserID: "lead", Role: domain.RoleOwner},
			{UserID: "dev", Role: domain.RoleEditor},
		},
	}
	for user, want := range map[domain.UserID]bool{"boss": true, "lead": true, "dev": false, "stranger": false} {
		if got := rules.CanAdminDomain(d, user); got != want {
			t.Errorf("CanAdminDomain(%s) = %v, want %v", user, got, want)
		}
	}
}
