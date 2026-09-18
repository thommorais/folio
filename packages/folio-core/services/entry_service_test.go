package services_test

import (
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

type entryFixture struct {
	entries *fakeEntries
	issues  *fakeIssues
	svc     *services.EntryService
	issueSv *services.IssueService
	owner   ports.Actor
	viewer  ports.Actor
	outside ports.Actor
	project domain.ProjectID
	other   domain.ProjectID
}

func newEntryFixture(t *testing.T) *entryFixture {
	t.Helper()

	projects := newFakeProjects()
	projects.items["p001"] = domain.Project{ID: "p001", Slug: "api", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
		{UserID: "u-viewer", Role: domain.RoleViewer},
	}}
	projects.items["p002"] = domain.Project{ID: "p002", Slug: "web", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
	}}

	entries := newFakeEntries()
	issues := newFakeIssues()
	plans := newFakePlans()
	guard := services.NewProjectGuard(projects)
	clock := &fakeClock{now: testNow}

	return &entryFixture{
		entries: entries,
		issues:  issues,
		svc:     services.NewEntryService(entries, issues, plans, guard, clock, &seqIDs{prefix: "e"}, nopLogger{}),
		issueSv: services.NewIssueService(issues, plans, entries, newFakeCycles(), guard, clock, &seqIDs{prefix: "is"}, nopLogger{}),
		owner:   ports.Actor{UserID: "u-owner"},
		viewer:  ports.Actor{UserID: "u-viewer"},
		outside: ports.Actor{UserID: "u-stranger"},
		project: "p001",
		other:   "p002",
	}
}

func TestJournalDocAndLogShareOneCollection(t *testing.T) {
	f := newEntryFixture(t)

	for _, kind := range []domain.EntryKind{domain.EntryJournal, domain.EntryDoc} {
		if _, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
			ProjectID: f.project, Kind: kind, Title: string(kind) + " one", Body: "text",
		}); err != nil {
			t.Fatalf("write %s: %v", kind, err)
		}
	}

	issue, err := f.issueSv.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Work",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryLog, IssueID: issue.ID, Body: "progress",
	}); err != nil {
		t.Fatalf("write log: %v", err)
	}

	all, err := f.svc.ListEntries(t.Context(), f.owner, f.project, domain.EntryFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("listed %d entries, want 3", len(all))
	}

	docs, err := f.svc.ListEntries(t.Context(), f.owner, f.project, domain.EntryFilter{Kind: domain.EntryDoc})
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || docs[0].Kind != domain.EntryDoc {
		t.Errorf("kind filter returned %d rows, want 1 doc", len(docs))
	}
}

func TestLogNeedsNoSlugButJournalDoes(t *testing.T) {
	f := newEntryFixture(t)

	issue, err := f.issueSv.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.project, Kind: domain.IssueTicket, Title: "Work",
	})
	if err != nil {
		t.Fatal(err)
	}

	log, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryLog, IssueID: issue.ID, Body: "just a body",
	})
	if err != nil {
		t.Fatal(err)
	}
	if log.Slug != "" || log.Title != "" {
		t.Errorf("log got slug %q title %q, want both empty", log.Slug, log.Title)
	}

	entry, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryJournal, Title: "Wired the cache", Body: "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.Slug != "wired-the-cache" {
		t.Errorf("slug = %q, want wired-the-cache", entry.Slug)
	}
}

func TestLogMustAttachToWork(t *testing.T) {
	f := newEntryFixture(t)

	_, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryLog, Body: "floating",
	})
	if err == nil {
		t.Error("a log with no issue and no plan should be rejected")
	}
}

func TestJournalAndDocShareASlugNamespace(t *testing.T) {
	f := newEntryFixture(t)

	if _, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryJournal, Slug: "architecture", Title: "Arch", Body: "b",
	}); err != nil {
		t.Fatal(err)
	}

	_, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryDoc, Slug: "architecture", Title: "Arch doc", Body: "b",
	})
	if err == nil {
		t.Error("a doc must not reuse a journal slug: they share one namespace now")
	}
}

func TestSlugDerivationAvoidsCollision(t *testing.T) {
	f := newEntryFixture(t)

	first, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryJournal, Title: "Same title", Body: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryDoc, Title: "Same title", Body: "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Slug == second.Slug {
		t.Errorf("both entries got slug %q", first.Slug)
	}
	if second.Slug != "same-title-2" {
		t.Errorf("second slug = %q, want same-title-2", second.Slug)
	}
}

func TestAppendKeepsMarkdownValid(t *testing.T) {
	f := newEntryFixture(t)

	entry, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryJournal, Title: "Notes", Body: "first",
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := f.svc.AppendToEntry(t.Context(), f.owner, entry.ID, "second")
	if err != nil {
		t.Fatal(err)
	}
	if got.Body != "first\n\nsecond" {
		t.Errorf("body = %q, want a blank line between sections", got.Body)
	}
}

func TestEntryCannotCrossProjects(t *testing.T) {
	f := newEntryFixture(t)

	theirs, err := f.issueSv.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{
		ProjectID: f.other, Kind: domain.IssueTicket, Title: "Theirs",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryJournal, Title: "Mine", Body: "b", IssueID: theirs.ID,
	})
	if err == nil {
		t.Error("an entry must not attach to an issue in another project")
	}
}

func TestViewerCannotWriteEntries(t *testing.T) {
	f := newEntryFixture(t)

	if _, err := f.svc.WriteEntry(t.Context(), f.viewer, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryJournal, Title: "Nope", Body: "b",
	}); err == nil {
		t.Error("a viewer should not write entries")
	}
}

func TestOutsiderCannotReadEntries(t *testing.T) {
	f := newEntryFixture(t)

	entry, err := f.svc.WriteEntry(t.Context(), f.owner, ports.WriteEntryInput{
		ProjectID: f.project, Kind: domain.EntryJournal, Title: "Secret", Body: "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.GetEntry(t.Context(), f.outside, entry.ID); err == nil {
		t.Error("a non-member should not read an entry")
	}
}
