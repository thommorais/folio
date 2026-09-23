package pb_test

import (
	"context"
	"slices"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
)

// find returns the hit for a record id, so an assertion can name what it
// wanted rather than an index into a slice.
func find(hits []domain.SearchHit, id string) (domain.SearchHit, bool) {
	for _, h := range hits {
		if h.ID == id {
			return h, true
		}
	}
	return domain.SearchHit{}, false
}

func ids(hits []domain.SearchHit) []string {
	out := make([]string, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.ID)
	}
	return out
}

func search(t *testing.T, app core.App, project string, q domain.SearchQuery) []domain.SearchHit {
	t.Helper()

	hits, err := pb.NewSearchRepository(app).Search(context.Background(), domain.ProjectID(project), q)
	if err != nil {
		t.Fatalf("search %q: %v", q.Text, err)
	}
	return hits
}

// Every kind has to reach the index through its own trigger, and a collection
// that discriminates by a column has to report the right kind per row.
func TestIndexCoversEveryKind(t *testing.T) {
	s := setup(t)

	plan := newRecord(t, s.app, pb.ColPlans, map[string]any{
		"project": s.project.Id, "title": "Ship the realtime work",
		"goal": "keep subscriptions alive", "status": "active",
	})
	todo := newRecord(t, s.app, pb.ColIssues, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "todo",
		"slug": "wire-realtime", "title": "Wire realtime", "status": "open", "priority": "low",
	})
	doc := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "realtime-notes", "title": "Realtime notes", "body": "how subscriptions behave",
	})
	log := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "log",
		"body": "moved the realtime handler",
	})
	cycle := newRecord(t, s.app, pb.ColCycles, map[string]any{
		"project": s.project.Id, "issue": s.ticket.Id, "ordinal": 2,
		"phase": "act", "resolution": "realtime settled",
	})

	hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "realtime"})

	for _, want := range []struct {
		name string
		id   string
		kind domain.SearchKind
	}{
		{"plan", plan.Id, domain.SearchKindPlan},
		{"todo", todo.Id, domain.SearchKindTodo},
		{"doc", doc.Id, domain.SearchKindDoc},
		{"work log", log.Id, domain.SearchKindWorkLog},
		{"cycle resolution", cycle.Id, domain.SearchKindCycle},
	} {
		t.Run(want.name, func(t *testing.T) {
			hit, ok := find(hits, want.id)
			if !ok {
				t.Fatalf("not indexed; got %v", ids(hits))
			}
			if hit.Kind != want.kind {
				t.Errorf("kind = %q, want %q", hit.Kind, want.kind)
			}
		})
	}
}

// A log has no title, and a hit with an empty one is unreadable in a list.
func TestWorkLogGetsATitleAndACycleKeepsItsOrdinal(t *testing.T) {
	s := setup(t)

	log := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "log",
		"body": "rewrote the pagination",
	})
	cycle := newRecord(t, s.app, pb.ColCycles, map[string]any{
		"project": s.project.Id, "issue": s.ticket.Id, "ordinal": 3,
		"phase": "act", "resolution": "pagination settled",
	})

	hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "pagination"})

	if hit, ok := find(hits, log.Id); !ok || hit.Title != "Work log" {
		t.Errorf("log title = %q (found %v), want %q", hit.Title, ok, "Work log")
	}
	if hit, ok := find(hits, cycle.Id); !ok || hit.Title != "Cycle 3 resolution" {
		t.Errorf("cycle title = %q (found %v), want %q", hit.Title, ok, "Cycle 3 resolution")
	}
}

// A cycle is only searchable once it has closed, and closing one has to add it
// to the index that the insert trigger skipped.
func TestCycleEntersTheIndexOnlyOnceResolved(t *testing.T) {
	s := setup(t)

	cycle := newRecord(t, s.app, pb.ColCycles, map[string]any{
		"project": s.project.Id, "issue": s.ticket.Id, "ordinal": 1, "phase": "do",
	})

	// Searching the title rather than the body: an unresolved cycle has no
	// resolution text, so a body term would report nothing either way and the
	// assertion would hold even if the row had been indexed.
	if hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "cycle resolution"}); len(hits) != 0 {
		t.Fatalf("an unresolved cycle is indexed: %v", ids(hits))
	}

	cycle.Set("resolution", "throughput doubled")
	if err := s.app.Save(cycle); err != nil {
		t.Fatal(err)
	}

	if _, ok := find(search(t, s.app, s.project.Id, domain.SearchQuery{Text: "throughput"}), cycle.Id); !ok {
		t.Error("resolving a cycle did not index it")
	}
}

// An edit that removes the matching text has to remove the row, or the index
// keeps answering for content that no longer exists.
func TestEditingAndDeletingKeepTheIndexHonest(t *testing.T) {
	s := setup(t)

	doc := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "caching", "title": "Caching", "body": "memoize the resolver",
	})

	if _, ok := find(search(t, s.app, s.project.Id, domain.SearchQuery{Text: "memoize"}), doc.Id); !ok {
		t.Fatal("not indexed on insert")
	}

	doc.Set("body", "drop the resolver entirely")
	if err := s.app.Save(doc); err != nil {
		t.Fatal(err)
	}

	if hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "memoize"}); len(hits) != 0 {
		t.Errorf("stale text still matches after an edit: %v", ids(hits))
	}
	if _, ok := find(search(t, s.app, s.project.Id, domain.SearchQuery{Text: "entirely"}), doc.Id); !ok {
		t.Error("the new text is not indexed")
	}

	if err := s.app.Delete(doc); err != nil {
		t.Fatal(err)
	}
	if hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "entirely"}); len(hits) != 0 {
		t.Errorf("a deleted record still matches: %v", ids(hits))
	}
}

// The point of the index over the old substring scan: a title match outranks
// a body match instead of the newer record simply winning.
func TestTitleOutranksBody(t *testing.T) {
	s := setup(t)

	body := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "mentions-webhooks", "title": "Deploy notes",
		"body": "webhooks are configured in the dashboard",
	})
	// Created second, so recency ordering alone would put this one first
	// regardless of where the term appears.
	titled := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "webhooks", "title": "Webhooks", "body": "unrelated prose",
	})

	hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "webhooks"})
	got := ids(hits)

	if len(got) < 2 {
		t.Fatalf("want both hits, got %v", got)
	}
	if got[0] != titled.Id {
		t.Errorf("ranked %v first, want the title match %s (body match is %s)", got[0], titled.Id, body.Id)
	}
}

// Identifiers are what an agent actually searches for. This pins recall, not
// the tokenizer: `tokenchars '_-'` is a latency choice, and without it
// pb_hooks indexes as two adjacent tokens that a phrase query still matches.
// Its evidence is a benchmark, not this test.
func TestIdentifiersSurviveTokenizing(t *testing.T) {
	s := setup(t)

	doc := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "railway", "title": "PocketBase on Railway",
		"body": "register pb_hooks and pass x-forwarded-for through the proxy",
	})

	for _, term := range []string{"pb_hooks", "x-forwarded-for", "PB_HOOKS"} {
		t.Run(term, func(t *testing.T) {
			if _, ok := find(search(t, s.app, s.project.Id, domain.SearchQuery{Text: term}), doc.Id); !ok {
				t.Errorf("%q does not match the indexed identifier", term)
			}
		})
	}

	// Porter stems, so a query in another form still finds the record.
	if _, ok := find(search(t, s.app, s.project.Id, domain.SearchQuery{Text: "registered"}), doc.Id); !ok {
		t.Error(`"registered" does not match "register": stemming is off`)
	}
}

// Filters narrow a ranked result rather than being applied to a separate scan.
func TestKindAndTagFilters(t *testing.T) {
	s := setup(t)

	doc := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "indexing", "title": "Indexing", "body": "how search works",
		"tags": []string{"decision"},
	})
	newRecord(t, s.app, pb.ColIssues, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "todo",
		"slug": "indexing-todo", "title": "Indexing", "status": "open", "priority": "low",
	})

	byKind := search(t, s.app, s.project.Id, domain.SearchQuery{
		Text: "indexing", Kinds: []domain.SearchKind{domain.SearchKindDoc},
	})
	if got := ids(byKind); !slices.Equal(got, []string{doc.Id}) {
		t.Errorf("kind filter gave %v, want only the doc %s", got, doc.Id)
	}

	byTag := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "indexing", Tags: []string{"decision"}})
	if got := ids(byTag); !slices.Equal(got, []string{doc.Id}) {
		t.Errorf("tag filter gave %v, want only the tagged doc %s", got, doc.Id)
	}
}

// A project's search must not leak another project's content.
func TestSearchIsScopedToTheProject(t *testing.T) {
	s := setup(t)

	other := newRecord(t, s.app, pb.ColProjects, map[string]any{
		"slug": "other", "name": "Other", "domain": s.domain.Id,
	})
	newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": other.Id, "kind": "doc",
		"slug": "secret", "title": "Quarterly secret", "body": "not yours",
	})
	mine := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "mine", "title": "Quarterly plan", "body": "mine",
	})

	if got := ids(search(t, s.app, s.project.Id, domain.SearchQuery{Text: "quarterly"})); !slices.Equal(got, []string{mine.Id}) {
		t.Errorf("got %v, want only %s: the other project leaked", got, mine.Id)
	}
}

// Rows written before the index existed still have to be findable.
func TestBackfillIndexesPreexistingRows(t *testing.T) {
	s := setup(t)

	doc := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "legacy", "title": "Legacy note", "body": "written before the index",
	})

	// Simulate the pre-index state: drop the table and rebuild from scratch,
	// which is what a deploy onto an existing database does.
	if _, err := s.app.DB().NewQuery("DROP TABLE " + pb.SearchIndex).Execute(); err != nil {
		t.Fatal(err)
	}
	if err := pb.Register(s.app); err != nil {
		t.Fatalf("re-register: %v", err)
	}

	if _, ok := find(search(t, s.app, s.project.Id, domain.SearchQuery{Text: "legacy"}), doc.Id); !ok {
		t.Error("backfill did not index a row that predated the index")
	}
}
