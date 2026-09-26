package services_test

import (
	"context"
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func TestCreateProjectIntoDomainRequiresBelongingToIt(t *testing.T) {
	f := newProjectFixture(t)
	f.domains.items["d-web"] = domain.Domain{
		ID: "d-web", ClientOwners: []domain.UserID{"u-boss"},
		Members: []domain.Member{{UserID: "u-team", Role: domain.RoleEditor}},
	}
	ctx := context.Background()

	for _, tc := range []struct {
		name  string
		actor ports.Actor
		ok    bool
	}{
		{"domain member", ports.Actor{UserID: "u-team"}, true},
		{"client owner", ports.Actor{UserID: "u-boss"}, true},
		{"superuser", ports.Actor{UserID: "u-root", Superuser: true}, true},
		{"stranger", f.outside, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.svc.CreateProject(ctx, tc.actor, ports.CreateProjectInput{
				DomainID: "d-web", Slug: "in-web-" + string(tc.actor.UserID), Name: "In Web",
			})
			if tc.ok && err != nil {
				t.Fatalf("create = %v, want success", err)
			}
			if !tc.ok && !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("create = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestCreateProjectIntoMissingDomainIsNotFound(t *testing.T) {
	f := newProjectFixture(t)

	_, err := f.svc.CreateProject(context.Background(), f.owner, ports.CreateProjectInput{DomainID: "d-ghost", Slug: "ghost", Name: "Ghost"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("create = %v, want ErrNotFound", err)
	}
}
