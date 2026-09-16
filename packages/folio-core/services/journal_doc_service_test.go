package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

type knowledgeFixture struct {
	clock      *fakeClock
	logs       *fakeJournal
	docs       *fakeDocs
	tickets    *fakeTickets
	journalSvc *services.JournalService
	docSvc     *services.DocService
	searchSvc  *services.SearchService
	owner      ports.Actor
	viewer     ports.Actor
	outside    ports.Actor
	project    domain.ProjectID
}

func newKnowledgeFixture(t *testing.T) *knowledgeFixture {
	t.Helper()
	projects := newFakeProjects()
	projects.items["p001"] = domain.Project{ID: "p001", Slug: "api", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
		{UserID: "u-viewer", Role: domain.RoleViewer},
	}}
	guard := services.NewProjectGuard(projects)
	clock := &fakeClock{now: testNow}

	logs := newFakeJournal()
	docs := newFakeDocs()
	tickets := newFakeTickets()
	search := &fakeSearch{}

	return &knowledgeFixture{
		clock: clock, logs: logs, docs: docs, tickets: tickets, project: "p001",
		journalSvc: services.NewJournalService(logs, tickets, guard, clock, &seqIDs{prefix: "l"}, nopLogger{}),
		docSvc:     services.NewDocService(docs, tickets, guard, clock, &seqIDs{prefix: "d"}, nopLogger{}),
		searchSvc:  services.NewSearchService(search, guard, projects),
		owner:      ports.Actor{UserID: "u-owner"},
		viewer:     ports.Actor{UserID: "u-viewer"},
		outside:    ports.Actor{UserID: "u-stranger"},
	}
}

func TestCreateDocRejectsDuplicateSlug(t *testing.T) {
	f := newKnowledgeFixture(t)
	in := ports.CreateDocInput{ProjectID: f.project, Slug: "architecture", Title: "Architecture", Body: "hex"}

	if _, err := f.docSvc.CreateDoc(context.Background(), f.owner, in); err != nil {
		t.Fatal(err)
	}
	if _, err := f.docSvc.CreateDoc(context.Background(), f.owner, in); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a slug must be unique within a project, got %v", err)
	}
}

func TestCreateDocDerivesSlugFromTitle(t *testing.T) {
	f := newKnowledgeFixture(t)

	got, err := f.docSvc.CreateDoc(context.Background(), f.owner, ports.CreateDocInput{
		ProjectID: f.project, Title: "Testing Strategy & Tools", Body: "x",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got.Slug != "testing-strategy-tools" {
		t.Fatalf("want a slug derived from the title, got %q", got.Slug)
	}
}

func TestUpdateDocChecksSlugCollision(t *testing.T) {
	f := newKnowledgeFixture(t)
	first, _ := f.docSvc.CreateDoc(context.Background(), f.owner, ports.CreateDocInput{ProjectID: f.project, Slug: "one", Title: "One", Body: "x"})
	_, _ = f.docSvc.CreateDoc(context.Background(), f.owner, ports.CreateDocInput{ProjectID: f.project, Slug: "two", Title: "Two", Body: "x"})

	taken := "two"
	if _, err := f.docSvc.UpdateDoc(context.Background(), f.owner, first.ID, ports.UpdateDocInput{Slug: &taken}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want conflict, got %v", err)
	}

	// Keeping its own slug must not read as a collision with itself.
	same := "one"
	if _, err := f.docSvc.UpdateDoc(context.Background(), f.owner, first.ID, ports.UpdateDocInput{Slug: &same}); err != nil {
		t.Fatalf("re-setting a doc's own slug must be allowed, got %v", err)
	}
}

func TestGetDocBySlug(t *testing.T) {
	f := newKnowledgeFixture(t)
	if _, err := f.docSvc.CreateDoc(context.Background(), f.owner, ports.CreateDocInput{ProjectID: f.project, Slug: "conventions", Title: "Conventions", Body: "x"}); err != nil {
		t.Fatal(err)
	}

	got, err := f.docSvc.GetDocBySlug(context.Background(), f.viewer, f.project, "conventions")
	if err != nil {
		t.Fatalf("get by slug: %v", err)
	}
	if got.Title != "Conventions" {
		t.Fatalf("got %q", got.Title)
	}

	if _, err := f.docSvc.GetDocBySlug(context.Background(), f.outside, f.project, "conventions"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-member must get not-found, got %v", err)
	}
}

func TestDocWritesRequireWriteAccess(t *testing.T) {
	f := newKnowledgeFixture(t)
	doc, _ := f.docSvc.CreateDoc(context.Background(), f.owner, ports.CreateDocInput{ProjectID: f.project, Slug: "x", Title: "X", Body: "x"})

	title := "Nope"
	if _, err := f.docSvc.UpdateDoc(context.Background(), f.viewer, doc.ID, ports.UpdateDocInput{Title: &title}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not edit, got %v", err)
	}
	if err := f.docSvc.DeleteDoc(context.Background(), f.viewer, doc.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not delete, got %v", err)
	}
}

func TestSearchRequiresProjectAccessAndCapsLimit(t *testing.T) {
	f := newKnowledgeFixture(t)

	if _, err := f.searchSvc.Search(context.Background(), f.outside, f.project, domain.SearchQuery{Text: "x"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-member must get not-found, got %v", err)
	}

	got, err := f.searchSvc.Search(context.Background(), f.viewer, f.project, domain.SearchQuery{Text: "x", Limit: 99999})
	if err != nil {
		t.Fatalf("viewer search: %v", err)
	}
	if got == nil {
		t.Fatal("search must return an empty slice rather than nil, so clients can range over it")
	}
}

func TestSearchRejectsEmptyQuery(t *testing.T) {
	f := newKnowledgeFixture(t)

	if _, err := f.searchSvc.Search(context.Background(), f.owner, f.project, domain.SearchQuery{Text: "  "}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("an empty search would scan the project, want validation error, got %v", err)
	}
}

// The work log documents what was built, how and where it stands. Entries are
// titled, searchable and editable, and their dates come from the server clock
// so the project reads back chronologically.

func TestWriteJournalEntryRecordsAuthorAndClock(t *testing.T) {
	f := newKnowledgeFixture(t)

	got, err := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{
		ProjectID:   f.project,
		Title:       "Restored GA4 pageview tracking",
		Body:        "The config call was deleted in #4484, so nothing fired a hit.",
		Branch:      "release/r378-ga-pageview-fix",
		PR:          "4873",
		ExternalRef: "XWWP-4420",
		Tags:        []string{"analytics", "decision"},
	})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if got.CreatedBy != "u-owner" {
		t.Fatalf("entry must record its author, got %q", got.CreatedBy)
	}
	if !got.CreatedAt.Equal(testNow) || !got.UpdatedAt.Equal(testNow) {
		t.Fatalf("dates must come from the injected clock, got %v / %v", got.CreatedAt, got.UpdatedAt)
	}
	if got.Branch != "release/r378-ga-pageview-fix" || got.PR != "4873" || got.ExternalRef != "XWWP-4420" {
		t.Fatalf("work refs not stored: %+v", got)
	}
}

func TestWriteJournalEntryRequiresATitle(t *testing.T) {
	f := newKnowledgeFixture(t)

	// A body alone is not a log: without a title it cannot be found again.
	if _, err := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{
		ProjectID: f.project, Title: "  ", Body: "lots of detail",
	}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestWriteJournalEntryAcceptsAStub(t *testing.T) {
	f := newKnowledgeFixture(t)

	// Work is often logged before it is finished; an empty body is allowed.
	if _, err := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{
		ProjectID: f.project, Title: "Investigating the OOM",
	}); err != nil {
		t.Fatalf("a stub entry must be allowed, got %v", err)
	}
}

func TestWriteJournalEntryRequiresWriteAccess(t *testing.T) {
	f := newKnowledgeFixture(t)
	in := ports.WriteJournalInput{ProjectID: f.project, Title: "x"}

	if _, err := f.journalSvc.WriteJournalEntry(context.Background(), f.viewer, in); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a viewer must not write logs, got %v", err)
	}
	if _, err := f.journalSvc.WriteJournalEntry(context.Background(), f.outside, in); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-member must get not-found, got %v", err)
	}
}

func TestUpdateJournalEntryTracksTheEditButKeepsCreatedAt(t *testing.T) {
	f := newKnowledgeFixture(t)
	entry, err := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{
		ProjectID: f.project, Title: "Search rewrite", Body: "first pass",
	})
	if err != nil {
		t.Fatal(err)
	}

	later := testNow.Add(48 * time.Hour)
	f.clock.now = later
	body := "first pass, then FTS5"
	got, err := f.journalSvc.UpdateJournalEntry(context.Background(), f.owner, entry.ID, ports.UpdateJournalInput{Body: &body})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !got.CreatedAt.Equal(testNow) {
		t.Fatalf("editing must not move the original date, got %v", got.CreatedAt)
	}
	if !got.UpdatedAt.Equal(later) {
		t.Fatalf("want the edit recorded at %v, got %v", later, got.UpdatedAt)
	}
	if got.Title != "Search rewrite" {
		t.Fatalf("an omitted field must survive the patch, got %q", got.Title)
	}
}

func TestAppendToJournalEntryAddsASection(t *testing.T) {
	f := newKnowledgeFixture(t)
	entry, err := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{
		ProjectID: f.project, Title: "Search rewrite", Body: "Chose FTS5.",
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := f.journalSvc.AppendToJournalEntry(context.Background(), f.owner, entry.ID, "Ported to r379, clean cherry-pick.")
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if !strings.Contains(got.Body, "Chose FTS5.") {
		t.Fatal("appending must not drop the existing body")
	}
	if !strings.Contains(got.Body, "Ported to r379") {
		t.Fatal("the new section is missing")
	}
	if strings.Contains(got.Body, "Chose FTS5.Ported") {
		t.Fatal("sections must be separated, not concatenated")
	}
}

func TestAppendToAnEmptyLogDoesNotLeadWithBlankLines(t *testing.T) {
	f := newKnowledgeFixture(t)
	entry, _ := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{ProjectID: f.project, Title: "Stub"})

	got, err := f.journalSvc.AppendToJournalEntry(context.Background(), f.owner, entry.ID, "First finding.")
	if err != nil {
		t.Fatal(err)
	}
	if got.Body != "First finding." {
		t.Fatalf("want the section alone, got %q", got.Body)
	}
}

func TestAppendToJournalEntryRejectsEmptyText(t *testing.T) {
	f := newKnowledgeFixture(t)
	entry, _ := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{ProjectID: f.project, Title: "Stub"})

	if _, err := f.journalSvc.AppendToJournalEntry(context.Background(), f.owner, entry.ID, "   "); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestLogReadsAreOpenToViewers(t *testing.T) {
	f := newKnowledgeFixture(t)
	entry, err := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{
		ProjectID: f.project, Title: "Visible to the team",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.journalSvc.GetJournalEntry(context.Background(), f.viewer, entry.ID); err != nil {
		t.Fatalf("a viewer must be able to read a log: %v", err)
	}
	got, err := f.journalSvc.ListJournal(context.Background(), f.viewer, f.project, domain.JournalFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 entry, got %d", len(got))
	}
	if _, err := f.journalSvc.GetJournalEntry(context.Background(), f.outside, entry.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-member must get not-found, got %v", err)
	}
}

func TestEditingAndDeletingRequireWriteAccess(t *testing.T) {
	f := newKnowledgeFixture(t)
	entry, _ := f.journalSvc.WriteJournalEntry(context.Background(), f.owner, ports.WriteJournalInput{ProjectID: f.project, Title: "x"})

	title := "hijacked"
	if _, err := f.journalSvc.UpdateJournalEntry(context.Background(), f.viewer, entry.ID, ports.UpdateJournalInput{Title: &title}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not edit, got %v", err)
	}
	if _, err := f.journalSvc.AppendToJournalEntry(context.Background(), f.viewer, entry.ID, "sneaky"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not append, got %v", err)
	}
	if err := f.journalSvc.DeleteJournalEntry(context.Background(), f.viewer, entry.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not delete, got %v", err)
	}
	if err := f.journalSvc.DeleteJournalEntry(context.Background(), f.owner, entry.ID); err != nil {
		t.Fatalf("owner delete: %v", err)
	}
}

func TestListJournalCapsTheLimit(t *testing.T) {
	f := newKnowledgeFixture(t)

	// The log is the project's memory and grows without bound; an unbounded
	// read would eventually blow a context window.
	if _, err := f.journalSvc.ListJournal(context.Background(), f.owner, f.project, domain.JournalFilter{Limit: 100000}); err != nil {
		t.Fatal(err)
	}
	if got := f.logs.lastFilter.Limit; got != services.MaxPageSize {
		t.Fatalf("want the limit capped at %d, got %d", services.MaxPageSize, got)
	}
}
