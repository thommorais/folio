package services_test

import (
	"context"
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func TestClientOwnerAdministersProjectWithoutPersonalGrant(t *testing.T) {
	f := newProjectFixture(t)
	p := f.repo.items[f.project.ID]
	p.ClientOwners = []domain.UserID{"u-boss"}
	f.repo.items[p.ID] = p

	if _, err := f.guard.EnsureAdmin(context.Background(), ports.Actor{UserID: "u-boss"}, p.ID); err != nil {
		t.Fatalf("client owner denied admin: %v", err)
	}
}

func TestDomainGrantReachesDomainMembers(t *testing.T) {
	f := newProjectFixture(t)
	p := f.repo.items[f.project.ID]
	p.DomainGrants = []domain.DomainGrant{{DomainID: "d-web", Role: domain.RoleEditor, Members: []domain.UserID{"u-team"}}}
	f.repo.items[p.ID] = p
	team := ports.Actor{UserID: "u-team"}

	if _, err := f.guard.EnsureWrite(context.Background(), team, p.ID); err != nil {
		t.Fatalf("domain member denied the domain's editor grant: %v", err)
	}
	if _, err := f.guard.EnsureAdmin(context.Background(), team, p.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("editor grant allowed admin, err = %v", err)
	}

	listed, err := f.svc.ListProjects(context.Background(), team, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("domain member lists %d projects, want 1", len(listed))
	}
}
