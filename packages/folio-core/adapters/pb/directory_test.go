package pb_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
)

func TestClientListFollowsVisibility(t *testing.T) {
	s := setup(t)
	c := castGrants(t, s)
	elsewhere := newRecord(t, s.app, pb.ColClients, map[string]any{"slug": "globex", "name": "Globex"})
	newRecord(t, s.app, pb.ColDomains, map[string]any{"client": elsewhere.Id, "slug": "ops", "name": "Ops"})

	repo := pb.NewClientRepository(s.app)
	for _, tc := range []struct {
		name string
		user *core.Record
		want []string
	}{
		{"plain client member", c.clientMember, []string{"acme"}},
		{"domain roster member", c.teamEditor, []string{"acme"}},
		{"personal grant only", s.viewer, []string{"acme"}},
		{"stranger", s.stranger, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clients, err := repo.List(t.Context(), domain.UserID(tc.user.Id))
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, cl := range clients {
				got = append(got, cl.Slug)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("clients = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDomainListFollowsVisibility(t *testing.T) {
	s := setup(t)
	c := castGrants(t, s)

	repo := pb.NewDomainRepository(s.app)
	for _, tc := range []struct {
		name string
		user *core.Record
		want []string
	}{
		{"client owner", c.boss, []string{"infra", "readers", "web"}},
		{"plain client member", c.clientMember, nil},
		{"domain roster member", c.teamEditor, []string{"infra"}},
		{"personal grant only", s.viewer, []string{"web"}},
		{"stranger", s.stranger, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			domains, err := repo.List(t.Context(), domain.UserID(tc.user.Id))
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, d := range domains {
				got = append(got, d.Slug)
				if d.ClientSlug != "acme" {
					t.Errorf("domain %s carries client slug %q, want acme", d.Slug, d.ClientSlug)
				}
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("domains = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClientRepositoryUpdates(t *testing.T) {
	s := setup(t)
	c := castGrants(t, s)
	newRecord(t, s.app, pb.ColClients, map[string]any{"slug": "globex", "name": "Globex"})
	repo := pb.NewClientRepository(s.app)

	got, err := repo.GetBySlug(t.Context(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Owners) != 1 || string(got.Owners[0]) != c.boss.Id {
		t.Errorf("owners = %v, want the boss only", got.Owners)
	}

	got.Name, got.Site = "Acme Corp", "https://acme.test"
	updated, err := repo.Update(t.Context(), got)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Acme Corp" || updated.Site != "https://acme.test" {
		t.Errorf("updated = %+v", updated)
	}

	got.Slug = "globex"
	if _, err := repo.Update(t.Context(), got); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("taking another client's slug: err = %v, want ErrConflict", err)
	}
	if _, err := repo.GetByID(t.Context(), "missing0000000"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("missing client: err = %v, want ErrNotFound", err)
	}
}
