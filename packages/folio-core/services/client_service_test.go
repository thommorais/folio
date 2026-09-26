package services_test

import (
	"context"
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

func newClientFixture() (*services.ClientService, *fakeClients) {
	repo := newFakeClients()
	repo.items["c-acme"] = domain.Client{ID: "c-acme", Slug: "acme", Name: "Acme", Site: "https://acme.test", Owners: []domain.UserID{"u-boss"}}
	repo.visible["u-member"] = []domain.ClientID{"c-acme"}
	repo.visible["u-boss"] = []domain.ClientID{"c-acme"}
	return services.NewClientService(repo, &fakeClock{now: testNow}), repo
}

func TestUpdateClientChangesOnlyGivenFields(t *testing.T) {
	svc, _ := newClientFixture()

	got, err := svc.UpdateClient(context.Background(), ports.Actor{UserID: "u-boss"}, "acme", ports.UpdateClientInput{Name: ptr("Acme Corp")})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Acme Corp" || got.Slug != "acme" || got.Site != "https://acme.test" {
		t.Errorf("client = %+v, want only the name changed", got)
	}
	if !got.UpdatedAt.Equal(testNow) {
		t.Errorf("updated at %v, want the injected clock", got.UpdatedAt)
	}
}

func TestUpdateClientResolvesIDOrSlug(t *testing.T) {
	svc, _ := newClientFixture()
	boss := ports.Actor{UserID: "u-boss"}

	for _, ref := range []string{"c-acme", "acme"} {
		if _, err := svc.UpdateClient(context.Background(), boss, ref, ports.UpdateClientInput{Descr: ptr("x")}); err != nil {
			t.Errorf("ref %q: %v", ref, err)
		}
	}
}

func TestUpdateClientRequiresOwner(t *testing.T) {
	svc, _ := newClientFixture()
	in := ports.UpdateClientInput{Name: ptr("Hijacked")}

	if _, err := svc.UpdateClient(context.Background(), ports.Actor{UserID: "u-member"}, "acme", in); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("member who can see the client: err = %v, want ErrForbidden", err)
	}
	if _, err := svc.UpdateClient(context.Background(), ports.Actor{UserID: "u-stranger"}, "acme", in); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("stranger: err = %v, want ErrNotFound", err)
	}
	if _, err := svc.UpdateClient(context.Background(), ports.Actor{UserID: "u-root", Superuser: true}, "acme", in); err != nil {
		t.Errorf("superuser: %v", err)
	}
}

func TestUpdateClientValidates(t *testing.T) {
	svc, repo := newClientFixture()

	_, err := svc.UpdateClient(context.Background(), ports.Actor{UserID: "u-boss"}, "acme", ports.UpdateClientInput{Slug: ptr("Not A Slug")})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if repo.items["c-acme"].Slug != "acme" {
		t.Error("an invalid update was stored")
	}
}
