package pb_test

import (
	"testing"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func TestIssueTagsRoundTripThroughTheJoin(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)

	created, err := repo.Create(t.Context(), domain.Issue{
		Kind:      domain.IssueTicket,
		ProjectID: domain.ProjectID(s.project.Id),
		Slug:      "tagged",
		Title:     "Tagged",
		Status:    domain.IssueOpen,
		Priority:  domain.PriorityMedium,
		Tags:      []string{"backend", "search"},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tags) != 2 {
		t.Fatalf("read back %d tags, want 2: %v", len(got.Tags), got.Tags)
	}

	rows, err := s.app.FindAllRecords(pb.ColIssueTags)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Errorf("join has %d rows, want 2", len(rows))
	}
}

func TestSetTagsReplacesTheSet(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)

	created, err := repo.Create(t.Context(), domain.Issue{
		Kind: domain.IssueTicket, ProjectID: domain.ProjectID(s.project.Id),
		Slug: "shifting", Title: "Shifting", Status: domain.IssueOpen,
		Priority: domain.PriorityMedium, Tags: []string{"a", "b"},
	})
	if err != nil {
		t.Fatal(err)
	}

	created.Tags = []string{"b", "c"}
	if _, err := repo.Update(t.Context(), created); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, tag := range got.Tags {
		have[tag] = true
	}
	if have["a"] || !have["b"] || !have["c"] {
		t.Errorf("tags = %v, want b and c only", got.Tags)
	}
}

func TestTagsAreScopedPerDomain(t *testing.T) {
	s := setup(t)
	tags := pb.NewTagRepository(s.app)

	first, err := tags.Ensure(t.Context(), domain.DomainID(s.domain.Id), []string{"bug"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := tags.Ensure(t.Context(), domain.DomainID(s.other.Id), []string{"bug"})
	if err != nil {
		t.Fatal(err)
	}

	if first[0].ID == second[0].ID {
		t.Error("the same tag name in two domains should be two rows")
	}

	inDomain, err := tags.ListByDomain(t.Context(), domain.DomainID(s.domain.Id))
	if err != nil {
		t.Fatal(err)
	}
	if len(inDomain) != 1 {
		t.Errorf("domain has %d tags, want 1", len(inDomain))
	}
}

func TestEnsureIsIdempotentAndSlugifies(t *testing.T) {
	s := setup(t)
	tags := pb.NewTagRepository(s.app)

	first, err := tags.Ensure(t.Context(), domain.DomainID(s.domain.Id), []string{"Back End"})
	if err != nil {
		t.Fatal(err)
	}
	if first[0].Slug != "back-end" {
		t.Errorf("slug = %q, want back-end", first[0].Slug)
	}

	again, err := tags.Ensure(t.Context(), domain.DomainID(s.domain.Id), []string{"Back End"})
	if err != nil {
		t.Fatal(err)
	}
	if again[0].ID != first[0].ID {
		t.Error("ensuring the same name twice created a second tag")
	}
}

func TestFilterByTagRequiresAll(t *testing.T) {
	s := setup(t)
	repo := pb.NewIssueRepository(s.app)
	project := domain.ProjectID(s.project.Id)

	mk := func(slug string, tags []string) {
		if _, err := repo.Create(t.Context(), domain.Issue{
			Kind: domain.IssueTicket, ProjectID: project, Slug: slug, Title: slug,
			Status: domain.IssueOpen, Priority: domain.PriorityMedium, Tags: tags,
		}); err != nil {
			t.Fatal(err)
		}
	}
	mk("both", []string{"backend", "search"})
	mk("one", []string{"backend"})
	mk("none", nil)

	got, err := repo.List(t.Context(), project, domain.IssueFilter{Tags: []string{"backend", "search"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Slug != "both" {
		t.Errorf("filter returned %d issues, want only the one carrying both tags", len(got))
	}
}

func TestEntryTagsUseTheirOwnJoin(t *testing.T) {
	s := setup(t)
	entries := pb.NewEntryRepository(s.app)

	created, err := entries.Create(t.Context(), domain.Entry{
		Kind: domain.EntryDoc, ProjectID: domain.ProjectID(s.project.Id),
		Slug: "a-doc", Title: "A doc", Body: "b", Tags: []string{"architecture"},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := entries.GetByID(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "architecture" {
		t.Errorf("tags = %v, want [architecture]", got.Tags)
	}

	issueRows, err := s.app.FindAllRecords(pb.ColIssueTags)
	if err != nil {
		t.Fatal(err)
	}
	if len(issueRows) != 0 {
		t.Errorf("an entry tag landed in the issue join: %d rows", len(issueRows))
	}
}

var _ ports.TagRepository = (*pb.TagRepository)(nil)
