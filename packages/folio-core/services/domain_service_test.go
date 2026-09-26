package services_test

import (
	"context"
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

func newDomainFixture() (*services.DomainService, *fakeDomains) {
	clients := newFakeClients()
	clients.items["c-acme"] = domain.Client{ID: "c-acme", Slug: "acme", Name: "Acme"}
	clients.items["c-globex"] = domain.Client{ID: "c-globex", Slug: "globex", Name: "Globex"}

	domains := newFakeDomains()
	domains.items["d-web"] = domain.Domain{
		ID: "d-web", ClientID: "c-acme", ClientSlug: "acme", Slug: "web", Name: "Web", Descr: "front end",
		ClientOwners: []domain.UserID{"u-boss"},
		Members: []domain.Member{
			{UserID: "u-lead", Role: domain.RoleOwner},
			{UserID: "u-dev", Role: domain.RoleEditor},
		},
	}
	domains.items["d-ops"] = domain.Domain{ID: "d-ops", ClientID: "c-globex", ClientSlug: "globex", Slug: "ops", Name: "Ops"}
	domains.visible["u-dev"] = []domain.DomainID{"d-web"}
	return services.NewDomainService(domains, clients, &fakeClock{now: testNow}), domains
}

func TestUpdateDomainChangesOnlyGivenFields(t *testing.T) {
	svc, _ := newDomainFixture()

	got, err := svc.UpdateDomain(context.Background(), ports.Actor{UserID: "u-lead"}, "acme", "web", ports.UpdateDomainInput{Name: ptr("Web Team")})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Web Team" || got.Slug != "web" || got.Descr != "front end" {
		t.Errorf("domain = %+v, want only the name changed", got)
	}
	if !got.UpdatedAt.Equal(testNow) {
		t.Errorf("updated at %v, want the injected clock", got.UpdatedAt)
	}
}

func TestUpdateDomainResolvesIDsOrSlugs(t *testing.T) {
	svc, _ := newDomainFixture()
	boss := ports.Actor{UserID: "u-boss"}

	for _, ref := range [][2]string{{"acme", "web"}, {"c-acme", "d-web"}, {"acme", "d-web"}} {
		if _, err := svc.UpdateDomain(context.Background(), boss, ref[0], ref[1], ports.UpdateDomainInput{Descr: ptr("x")}); err != nil {
			t.Errorf("refs %v: %v", ref, err)
		}
	}
	if _, err := svc.UpdateDomain(context.Background(), ports.Actor{Superuser: true}, "acme", "d-ops", ports.UpdateDomainInput{Descr: ptr("x")}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("a domain id under another client: err = %v, want ErrNotFound", err)
	}
}

func TestUpdateDomainRequiresRosterOwnerOrClientOwner(t *testing.T) {
	svc, _ := newDomainFixture()
	in := ports.UpdateDomainInput{Name: ptr("Hijacked")}

	for _, tc := range []struct {
		name  string
		actor ports.Actor
		want  error
	}{
		{"client owner", ports.Actor{UserID: "u-boss"}, nil},
		{"roster owner", ports.Actor{UserID: "u-lead"}, nil},
		{"superuser", ports.Actor{UserID: "u-root", Superuser: true}, nil},
		{"roster editor", ports.Actor{UserID: "u-dev"}, domain.ErrForbidden},
		{"stranger", ports.Actor{UserID: "u-stranger"}, domain.ErrNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.UpdateDomain(context.Background(), tc.actor, "acme", "web", in)
			if tc.want == nil && err != nil {
				t.Fatalf("err = %v, want success", err)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestUpdateDomainValidates(t *testing.T) {
	svc, domains := newDomainFixture()

	_, err := svc.UpdateDomain(context.Background(), ports.Actor{UserID: "u-boss"}, "acme", "web", ports.UpdateDomainInput{Slug: ptr("Not A Slug")})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if domains.items["d-web"].Slug != "web" {
		t.Error("an invalid update was stored")
	}
}
