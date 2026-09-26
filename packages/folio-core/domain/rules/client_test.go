package rules_test

import (
	"errors"
	"strings"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func TestValidateClient(t *testing.T) {
	ok := domain.Client{Slug: "acme", Name: "Acme"}
	if err := rules.ValidateClient(ok); err != nil {
		t.Fatalf("valid client rejected: %v", err)
	}

	for name, c := range map[string]domain.Client{
		"bad slug":      {Slug: "Acme Co", Name: "Acme"},
		"empty name":    {Slug: "acme", Name: " "},
		"long site":     {Slug: "acme", Name: "Acme", Site: strings.Repeat("s", 301)},
		"long logo":     {Slug: "acme", Name: "Acme", Logo: strings.Repeat("l", 301)},
		"long describe": {Slug: "acme", Name: "Acme", Descr: strings.Repeat("d", rules.DescrMaxLen+1)},
	} {
		if err := rules.ValidateClient(c); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("%s: want validation error, got %v", name, err)
		}
	}
}

func TestCanAdminClient(t *testing.T) {
	c := domain.Client{Owners: []domain.UserID{"boss"}}
	if !rules.CanAdminClient(c, "boss") {
		t.Error("client owner cannot administer the client")
	}
	if rules.CanAdminClient(c, "someone") {
		t.Error("a non-owner can administer the client")
	}
}
