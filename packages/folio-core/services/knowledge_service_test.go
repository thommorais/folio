package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

type knowledgeFixture struct {
	svc     *services.KnowledgeService
	repo    *fakeKnowledge
	author  ports.Actor
	someone ports.Actor
	// project is one the author belongs to; closed is one they do not.
	project domain.ProjectID
	closed  domain.ProjectID
}

func newKnowledgeFixture(t *testing.T) *knowledgeFixture {
	t.Helper()

	projects := newFakeProjects()
	projects.items["p001"] = domain.Project{ID: "p001", Slug: "api", Members: []domain.Member{
		{UserID: "u-author", Role: domain.RoleEditor},
		{UserID: "u-someone", Role: domain.RoleViewer},
	}}
	projects.items["p002"] = domain.Project{ID: "p002", Slug: "closed", Members: []domain.Member{
		{UserID: "u-stranger", Role: domain.RoleOwner},
	}}

	repo := newFakeKnowledge()
	return &knowledgeFixture{
		svc: services.NewKnowledgeService(repo, services.NewProjectGuard(projects),
			&fakeClock{now: time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)},
			&seqIDs{prefix: "k"}, nopLogger{}),
		repo:    repo,
		author:  ports.Actor{UserID: "u-author"},
		someone: ports.Actor{UserID: "u-someone"},
		project: "p001",
		closed:  "p002",
	}
}

func (f *knowledgeFixture) write(t *testing.T, actor ports.Actor, in ports.WriteKnowledgeInput) domain.Knowledge {
	t.Helper()

	k, err := f.svc.WriteKnowledge(context.Background(), actor, in)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	return k
}

// The defining rule: knowledge is not fenced by project membership, so a user
// who shares no project with the author still reads and edits it.
func TestKnowledgeIsReadableAndWritableByAnyAuthenticatedUser(t *testing.T) {
	f := newKnowledgeFixture(t)
	stranger := ports.Actor{UserID: "u-stranger"}

	written := f.write(t, f.author, ports.WriteKnowledgeInput{
		Title: "Realtime on Railway", Body: "disable buffering",
	})

	got, err := f.svc.GetKnowledge(context.Background(), stranger, written.ID)
	if err != nil {
		t.Fatalf("a stranger cannot read knowledge: %v", err)
	}
	if got.ID != written.ID {
		t.Fatalf("got %q, want %q", got.ID, written.ID)
	}

	body := "disable buffering and set the health check"
	if _, err := f.svc.UpdateKnowledge(context.Background(), stranger, written.ID,
		ports.UpdateKnowledgeInput{Body: &body}); err != nil {
		t.Errorf("a stranger cannot edit knowledge: %v", err)
	}
	if err := f.svc.DeleteKnowledge(context.Background(), stranger, written.ID); err != nil {
		t.Errorf("a stranger cannot delete knowledge: %v", err)
	}
}

// Open to every user still means every *authenticated* user.
func TestKnowledgeRefusesAnAnonymousActor(t *testing.T) {
	f := newKnowledgeFixture(t)
	anon := ports.Actor{}

	_, err := f.svc.WriteKnowledge(context.Background(), anon, ports.WriteKnowledgeInput{Title: "x"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("write by an anonymous actor: got %v, want ErrForbidden", err)
	}

	written := f.write(t, f.author, ports.WriteKnowledgeInput{Title: "Realtime"})
	if _, err := f.svc.GetKnowledge(context.Background(), anon, written.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("read by an anonymous actor: got %v, want ErrForbidden", err)
	}
}

// The author is recorded on write and never moves afterwards, since it is the
// only trace of where a shared note came from.
func TestKnowledgeKeepsItsOriginalAuthor(t *testing.T) {
	f := newKnowledgeFixture(t)

	written := f.write(t, f.author, ports.WriteKnowledgeInput{Title: "Realtime"})
	if written.CreatedBy != f.author.UserID {
		t.Fatalf("created_by = %q, want %q", written.CreatedBy, f.author.UserID)
	}

	title := "Realtime, revised"
	edited, err := f.svc.UpdateKnowledge(context.Background(), f.someone, written.ID,
		ports.UpdateKnowledgeInput{Title: &title})
	if err != nil {
		t.Fatal(err)
	}
	if edited.CreatedBy != f.author.UserID {
		t.Errorf("an edit moved created_by to %q, want %q", edited.CreatedBy, f.author.UserID)
	}
	if !edited.UpdatedAt.After(written.CreatedAt) && !edited.UpdatedAt.Equal(written.CreatedAt) {
		t.Errorf("updated_at went backwards: %v before %v", edited.UpdatedAt, written.CreatedAt)
	}
}

// A project is optional, and naming one must not become a back door: a caller
// can only attach knowledge to a project they can actually read.
func TestKnowledgeProjectIsOptionalAndChecked(t *testing.T) {
	f := newKnowledgeFixture(t)

	loose := f.write(t, f.author, ports.WriteKnowledgeInput{Title: "General tip"})
	if loose.ProjectID != "" {
		t.Errorf("project = %q, want none", loose.ProjectID)
	}

	attached := f.write(t, f.author, ports.WriteKnowledgeInput{
		Title: "API tip", ProjectID: f.project,
	})
	if attached.ProjectID != f.project {
		t.Errorf("project = %q, want %q", attached.ProjectID, f.project)
	}

	_, err := f.svc.WriteKnowledge(context.Background(), f.author, ports.WriteKnowledgeInput{
		Title: "Sneaky", ProjectID: f.closed,
	})
	if err == nil {
		t.Error("attached knowledge to a project the actor cannot read")
	}

	_, err = f.svc.WriteKnowledge(context.Background(), f.author, ports.WriteKnowledgeInput{
		Title: "Ghost", ProjectID: "p-does-not-exist",
	})
	if err == nil {
		t.Error("attached knowledge to a project that does not exist")
	}
}

// Slugs address a note in a CLI, and knowledge is not project-scoped, so the
// namespace is global and a clash has to be resolved rather than collide.
func TestKnowledgeSlugsAreGlobalAndUnique(t *testing.T) {
	f := newKnowledgeFixture(t)

	first := f.write(t, f.author, ports.WriteKnowledgeInput{Title: "Realtime on Railway"})
	if first.Slug != "realtime-on-railway" {
		t.Fatalf("slug = %q, want it derived from the title", first.Slug)
	}

	second := f.write(t, f.someone, ports.WriteKnowledgeInput{Title: "Realtime on Railway"})
	if second.Slug == first.Slug {
		t.Errorf("two notes share the slug %q", second.Slug)
	}

	_, err := f.svc.WriteKnowledge(context.Background(), f.author, ports.WriteKnowledgeInput{
		Title: "Another", Slug: first.Slug,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("an explicitly requested duplicate slug: got %v, want ErrConflict", err)
	}
}

func TestKnowledgeListingIsClamped(t *testing.T) {
	f := newKnowledgeFixture(t)
	f.write(t, f.author, ports.WriteKnowledgeInput{Title: "One"})

	if _, err := f.svc.ListKnowledge(context.Background(), f.author,
		domain.KnowledgeFilter{Limit: 100000}); err != nil {
		t.Fatal(err)
	}
	if f.repo.lastLimit <= 0 || f.repo.lastLimit > services.MaxPageSize {
		t.Errorf("limit reached the repository as %d, want it clamped to %d",
			f.repo.lastLimit, services.MaxPageSize)
	}
}
