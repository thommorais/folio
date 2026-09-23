package pb_test

import (
	"context"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
)

func newKnowledge(t *testing.T, app core.App, k domain.Knowledge) domain.Knowledge {
	t.Helper()

	written, err := pb.NewKnowledgeRepository(app).Create(context.Background(), k)
	if err != nil {
		t.Fatalf("create knowledge: %v", err)
	}
	return written
}

// The point of the feature: a note written with no project answers a search
// made inside any project, because it is not fenced by one.
func TestKnowledgeAnswersASearchFromAnyProject(t *testing.T) {
	s := setup(t)

	note := newKnowledge(t, s.app, domain.Knowledge{
		ID: "know00000000001", Slug: "pocketbase-realtime-railway",
		Title: "Enable realtime in PocketBase on Railway",
		Body:  "Railway buffers the SSE response, so subscriptions hang.",
		Tags:  []string{"pocketbase"}, CreatedBy: domain.UserID(s.owner.Id),
	})

	// A second project, sharing nothing with the first.
	other := newRecord(t, s.app, pb.ColProjects, map[string]any{
		"slug": "unrelated", "name": "Unrelated", "domain": s.other.Id,
	})

	for _, project := range []struct{ name, id string }{
		{"the project it was written from", s.project.Id},
		{"an unrelated project", other.Id},
	} {
		t.Run(project.name, func(t *testing.T) {
			hits := search(t, s.app, project.id, domain.SearchQuery{Text: "railway"})
			hit, ok := find(hits, string(note.ID))
			if !ok {
				t.Fatalf("knowledge is not reachable from here; got %v", ids(hits))
			}
			if hit.Kind != domain.SearchKindKnowledge {
				t.Errorf("kind = %q, want %q", hit.Kind, domain.SearchKindKnowledge)
			}
			if hit.Slug != note.Slug {
				t.Errorf("slug = %q, want %q", hit.Slug, note.Slug)
			}
		})
	}
}

// Attaching a note to a project records where it came from. It must not turn
// into a fence: the note stays visible from everywhere.
func TestAttachedKnowledgeIsStillGlobal(t *testing.T) {
	s := setup(t)

	note := newKnowledge(t, s.app, domain.Knowledge{
		ID: "know00000000002", Slug: "attached-note", Title: "Attached note",
		Body: "learned while wiring the dashboard", ProjectID: domain.ProjectID(s.project.Id),
		CreatedBy: domain.UserID(s.owner.Id),
	})

	other := newRecord(t, s.app, pb.ColProjects, map[string]any{
		"slug": "elsewhere", "name": "Elsewhere", "domain": s.other.Id,
	})

	if _, ok := find(search(t, s.app, other.Id, domain.SearchQuery{Text: "dashboard"}), string(note.ID)); !ok {
		t.Error("attaching a project fenced the note off from other projects")
	}
}

// Project content must not become globally visible by sharing the index with
// knowledge: the scope column is doing real work in both directions.
func TestProjectContentStaysFencedAlongsideKnowledge(t *testing.T) {
	s := setup(t)

	newKnowledge(t, s.app, domain.Knowledge{
		ID: "know00000000003", Slug: "shared-tip", Title: "Shared tip",
		Body: "quarterly rollover notes", CreatedBy: domain.UserID(s.owner.Id),
	})
	secret := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "private", "title": "Quarterly figures", "body": "quarterly rollover figures",
	})

	other := newRecord(t, s.app, pb.ColProjects, map[string]any{
		"slug": "outside", "name": "Outside", "domain": s.other.Id,
	})

	hits := search(t, s.app, other.Id, domain.SearchQuery{Text: "quarterly rollover"})
	if _, leaked := find(hits, secret.Id); leaked {
		t.Errorf("a project doc leaked into another project's search: %v", ids(hits))
	}
	if len(hits) != 1 {
		t.Errorf("got %d hits, want only the knowledge note: %v", len(hits), ids(hits))
	}
}

// Knowledge writes its own tags column, so the tag filter has to work on it
// without the join table the other kinds use.
func TestKnowledgeTagsAreFilterable(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	note := newKnowledge(t, s.app, domain.Knowledge{
		ID: "know00000000004", Slug: "tagged-tip", Title: "Tagged tip",
		Body: "connection pooling advice", Tags: []string{"pocketbase", "railway"},
		CreatedBy: domain.UserID(s.owner.Id),
	})

	hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "pooling", Tags: []string{"railway"}})
	if _, ok := find(hits, string(note.ID)); !ok {
		t.Fatalf("tag filter missed the note: %v", ids(hits))
	}
	if hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "pooling", Tags: []string{"absent"}}); len(hits) != 0 {
		t.Errorf("a tag it does not carry still matched: %v", ids(hits))
	}

	// Retagging goes through the record itself here, so the record's own
	// update trigger is what has to refresh the index.
	note.Tags = []string{"sqlite"}
	if _, err := pb.NewKnowledgeRepository(s.app).Update(ctx, note); err != nil {
		t.Fatal(err)
	}
	if hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "pooling", Tags: []string{"railway"}}); len(hits) != 0 {
		t.Errorf("the removed tag still matches: %v", ids(hits))
	}
	if hits := search(t, s.app, s.project.Id, domain.SearchQuery{Text: "pooling", Tags: []string{"sqlite"}}); len(hits) != 1 {
		t.Errorf("the new tag does not match: %d hits", len(hits))
	}
}

// Deleting a project detaches its knowledge rather than destroying it: the
// note is the durable half of that relationship.
func TestDeletingAProjectKeepsItsKnowledge(t *testing.T) {
	s := setup(t)

	note := newKnowledge(t, s.app, domain.Knowledge{
		ID: "know00000000005", Slug: "outlives", Title: "Outlives the project",
		Body: "hard won lesson", ProjectID: domain.ProjectID(s.project.Id),
		CreatedBy: domain.UserID(s.owner.Id),
	})

	if err := s.app.Delete(s.project); err != nil {
		t.Fatal(err)
	}

	got, err := pb.NewKnowledgeRepository(s.app).GetByID(context.Background(), note.ID)
	if err != nil {
		t.Fatalf("the note died with its project: %v", err)
	}
	if got.ProjectID != "" {
		t.Errorf("project = %q, want it detached", got.ProjectID)
	}
}
