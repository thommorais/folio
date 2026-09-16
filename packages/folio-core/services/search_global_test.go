package services_test

import (
	"context"
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

// globalFixture gives the actor two projects and leaves a third one that only
// a stranger belongs to, so scoping is observable.
type globalFixture struct {
	svc     *services.SearchService
	member  ports.Actor
	outside ports.Actor
	search  *fakeSearch
}

func newGlobalFixture(t *testing.T) *globalFixture {
	t.Helper()

	projects := newFakeProjects()
	projects.items["p001"] = domain.Project{ID: "p001", Slug: "api", Members: []domain.Member{
		{UserID: "u-member", Role: domain.RoleOwner},
	}}
	projects.items["p002"] = domain.Project{ID: "p002", Slug: "web", Members: []domain.Member{
		{UserID: "u-member", Role: domain.RoleViewer},
	}}
	projects.items["p003"] = domain.Project{ID: "p003", Slug: "secret", Members: []domain.Member{
		{UserID: "u-stranger", Role: domain.RoleOwner},
	}}

	search := &fakeSearch{hits: []domain.SearchHit{
		{Kind: domain.SearchKindJournal, ID: "h1", ProjectID: "p001", Title: "one"},
		{Kind: domain.SearchKindDoc, ID: "h2", ProjectID: "p002", Title: "two"},
		{Kind: domain.SearchKindJournal, ID: "h3", ProjectID: "p003", Title: "hidden"},
	}}

	return &globalFixture{
		svc:     services.NewSearchService(search, services.NewProjectGuard(projects), projects),
		member:  ports.Actor{UserID: "u-member"},
		outside: ports.Actor{UserID: "u-nobody"},
		search:  search,
	}
}

func TestSearchAllSpansEveryProjectTheActorReads(t *testing.T) {
	f := newGlobalFixture(t)

	hits, err := f.svc.SearchAll(context.Background(), f.member, domain.SearchQuery{Text: "x"})
	if err != nil {
		t.Fatal(err)
	}

	ids := make([]string, 0, len(hits))
	for _, h := range hits {
		ids = append(ids, h.ID)
	}

	if len(ids) != 2 {
		t.Fatalf("want 2 hits across the actor's two projects, got %v", ids)
	}
}

// The whole point of scoping: a project the actor does not belong to must not
// leak through a global query.
func TestSearchAllExcludesProjectsTheActorCannotRead(t *testing.T) {
	f := newGlobalFixture(t)

	hits, err := f.svc.SearchAll(context.Background(), f.member, domain.SearchQuery{Text: "x"})
	if err != nil {
		t.Fatal(err)
	}

	for _, h := range hits {
		if h.ProjectID == "p003" {
			t.Fatalf("hit from a project the actor cannot read: %+v", h)
		}
	}
}

func TestSearchAllWithNoProjectsReturnsEmptyNotError(t *testing.T) {
	f := newGlobalFixture(t)

	hits, err := f.svc.SearchAll(context.Background(), f.outside, domain.SearchQuery{Text: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("want no hits for an actor with no projects, got %d", len(hits))
	}
	if hits == nil {
		t.Fatal("want an empty slice rather than nil, so the API renders []")
	}
}

func TestSearchAllRejectsEmptyQuery(t *testing.T) {
	f := newGlobalFixture(t)

	_, err := f.svc.SearchAll(context.Background(), f.member, domain.SearchQuery{Text: "   "})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want a validation error for an empty query, got %v", err)
	}
}

func TestSearchAllCapsTheLimit(t *testing.T) {
	f := newGlobalFixture(t)

	if _, err := f.svc.SearchAll(context.Background(), f.member, domain.SearchQuery{Text: "x", Limit: 99999}); err != nil {
		t.Fatal(err)
	}
	if f.search.lastLimit != services.MaxPageSize {
		t.Fatalf("want the limit clamped to %d, got %d", services.MaxPageSize, f.search.lastLimit)
	}
}
