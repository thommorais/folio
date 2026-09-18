package pb_test

import (
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/pb"
)

const (
	legacyTickets = "journ_tickets"
	legacyJournal = "journ_journal"
)

// seedLegacy recreates enough of the retired schema to stand in for a database
// that predates the drop. The real ones carried more columns; what matters here
// is the name, a project relation and the id, which is what the drop checks.
func seedLegacy(t *testing.T, app core.App, name string) *core.Collection {
	t.Helper()

	projects, err := app.FindCollectionByNameOrId(pb.ColProjects)
	if err != nil {
		t.Fatal(err)
	}

	c := core.NewBaseCollection(name)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		&core.TextField{Name: "title", Required: true, Max: 200},
	)
	if err := app.Save(c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestRegisterDropsALegacyCollectionOnceItsRowsWereCopied(t *testing.T) {
	s := setup(t)
	seedLegacy(t, s.app, legacyTickets)

	// A row that already exists in issues under the same id: the shape the
	// backfill left behind.
	copied := newRecord(t, s.app, legacyTickets, map[string]any{
		"project": s.project.Id, "title": "Already copied",
	})
	issues, err := s.app.FindCollectionByNameOrId(pb.ColIssues)
	if err != nil {
		t.Fatal(err)
	}
	issue := core.NewRecord(issues)
	// The id has to be set before the first save: PocketBase refuses to change
	// a primary key afterwards, and the id is what proves the counterpart.
	issue.Id = copied.Id
	issue.Set("domain", s.domain.Id)
	issue.Set("project", s.project.Id)
	issue.Set("kind", "ticket")
	issue.Set("slug", "already-copied")
	issue.Set("title", "Already copied")
	issue.Set("status", "open")
	issue.Set("priority", "medium")
	if err := s.app.Save(issue); err != nil {
		t.Fatal(err)
	}

	if err := pb.Register(s.app); err != nil {
		t.Fatal(err)
	}

	if _, err := s.app.FindCollectionByNameOrId(legacyTickets); err == nil {
		t.Error("journ_tickets survived the drop")
	}
}

func TestRegisterRefusesToDropARowThatWasNeverCopied(t *testing.T) {
	s := setup(t)
	seedLegacy(t, s.app, legacyTickets)

	orphan := newRecord(t, s.app, legacyTickets, map[string]any{
		"project": s.project.Id, "title": "Never copied",
	})

	err := pb.Register(s.app)
	if err == nil {
		t.Fatal("dropped while a row had no counterpart")
	}
	if !strings.Contains(err.Error(), orphan.Id) {
		t.Errorf("error does not name the uncopied row: %v", err)
	}

	if _, err := s.app.FindCollectionByNameOrId(legacyTickets); err != nil {
		t.Error("journ_tickets was dropped despite the refusal")
	}
}

func TestRegisterDropsAnEmptyLegacyCollection(t *testing.T) {
	s := setup(t)
	seedLegacy(t, s.app, legacyJournal)

	if err := pb.Register(s.app); err != nil {
		t.Fatal(err)
	}

	if _, err := s.app.FindCollectionByNameOrId(legacyJournal); err == nil {
		t.Error("an empty journ_journal survived the drop")
	}
}

func TestRegisterLeavesTheLiveCollectionsAlone(t *testing.T) {
	s := setup(t)
	seedLegacy(t, s.app, legacyTickets)

	if err := pb.Register(s.app); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		pb.ColIssues, pb.ColEntries, pb.ColLinks, pb.ColTags,
		pb.ColIssueTags, pb.ColEntryTags, pb.ColPlans, pb.ColCycles,
		pb.ColProjects, pb.ColDomains, pb.ColClients, pb.ColMembers,
	} {
		if _, err := s.app.FindCollectionByNameOrId(name); err != nil {
			t.Errorf("%s was dropped: %v", name, err)
		}
	}
}

// Register runs on every boot, so the drop has to survive finding nothing left
// to do.
func TestRegisterIsIdempotentOnADatabaseWithNoLegacyCollections(t *testing.T) {
	s := setup(t)

	for range 2 {
		if err := pb.Register(s.app); err != nil {
			t.Fatalf("second run failed: %v", err)
		}
	}
}
