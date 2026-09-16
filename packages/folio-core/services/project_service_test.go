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

var testNow = time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

type projectFixture struct {
	repo    *fakeProjects
	svc     *services.ProjectService
	guard   ports.Guard
	owner   ports.Actor
	editor  ports.Actor
	viewer  ports.Actor
	outside ports.Actor
	project domain.Project
}

func newProjectFixture(t *testing.T) *projectFixture {
	t.Helper()
	repo := newFakeProjects()
	repo.emails["editor@example.com"] = "u-editor"
	repo.emails["newbie@example.com"] = "u-newbie"

	project := domain.Project{
		ID: "p001", Slug: "api", Name: "API",
		Members: []domain.Member{
			{UserID: "u-owner", Role: domain.RoleOwner},
			{UserID: "u-editor", Role: domain.RoleEditor},
			{UserID: "u-viewer", Role: domain.RoleViewer},
		},
		CreatedAt: testNow, UpdatedAt: testNow,
	}
	repo.items[project.ID] = project

	guard := services.NewProjectGuard(repo)
	svc := services.NewProjectService(repo, guard, &fakeClock{now: testNow}, &seqIDs{prefix: "p"}, nopLogger{})

	return &projectFixture{
		repo: repo, svc: svc, guard: guard, project: project,
		owner:   ports.Actor{UserID: "u-owner"},
		editor:  ports.Actor{UserID: "u-editor"},
		viewer:  ports.Actor{UserID: "u-viewer"},
		outside: ports.Actor{UserID: "u-stranger"},
	}
}

func TestCreateProjectMakesCallerOwner(t *testing.T) {
	f := newProjectFixture(t)

	got, err := f.svc.CreateProject(context.Background(), f.outside, ports.CreateProjectInput{Slug: "new-thing", Name: "New Thing"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	role, ok := got.RoleOf("u-stranger")
	if !ok || role != domain.RoleOwner {
		t.Fatalf("creator must become owner, got role %q present=%v", role, ok)
	}
	if got.ID == "" {
		t.Fatal("created project must have an id")
	}
	if !got.CreatedAt.Equal(testNow) {
		t.Fatalf("want injected clock time, got %v", got.CreatedAt)
	}
}

func TestCreateProjectValidates(t *testing.T) {
	f := newProjectFixture(t)

	_, err := f.svc.CreateProject(context.Background(), f.owner, ports.CreateProjectInput{Slug: "Bad Slug", Name: "x"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestCreateProjectRejectsDuplicateSlug(t *testing.T) {
	f := newProjectFixture(t)

	_, err := f.svc.CreateProject(context.Background(), f.owner, ports.CreateProjectInput{Slug: "api", Name: "Another"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
}

func TestListProjectsOnlyReturnsMemberships(t *testing.T) {
	f := newProjectFixture(t)
	f.repo.items["p002"] = domain.Project{ID: "p002", Slug: "other", Name: "Other",
		Members: []domain.Member{{UserID: "u-someone", Role: domain.RoleOwner}}}

	got, err := f.svc.ListProjects(context.Background(), f.viewer, false)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].ID != "p001" {
		t.Fatalf("viewer should see only their project, got %+v", got)
	}
}

func TestGetProjectResolvesSlugOrID(t *testing.T) {
	f := newProjectFixture(t)

	bySlug, err := f.svc.GetProject(context.Background(), f.viewer, "api")
	if err != nil {
		t.Fatalf("by slug: %v", err)
	}
	byID, err := f.svc.GetProject(context.Background(), f.viewer, "p001")
	if err != nil {
		t.Fatalf("by id: %v", err)
	}
	if bySlug.ID != byID.ID {
		t.Fatal("slug and id must resolve to the same project")
	}
}

func TestNonMemberIsDeniedNotLeaked(t *testing.T) {
	f := newProjectFixture(t)

	_, err := f.svc.GetProject(context.Background(), f.outside, "api")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a non-member must get not-found rather than a forbidden that confirms existence, got %v", err)
	}
}

func TestUpdateProjectRequiresWrite(t *testing.T) {
	f := newProjectFixture(t)
	name := "Renamed"

	if _, err := f.svc.UpdateProject(context.Background(), f.viewer, "p001", ports.UpdateProjectInput{Name: &name}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not write, got %v", err)
	}

	got, err := f.svc.UpdateProject(context.Background(), f.editor, "p001", ports.UpdateProjectInput{Name: &name})
	if err != nil {
		t.Fatalf("editor update: %v", err)
	}
	if got.Name != "Renamed" {
		t.Fatalf("want Renamed, got %q", got.Name)
	}
}

func TestUpdateProjectLeavesOmittedFields(t *testing.T) {
	f := newProjectFixture(t)
	archived := true

	got, err := f.svc.UpdateProject(context.Background(), f.owner, "p001", ports.UpdateProjectInput{Archived: &archived})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Name != "API" {
		t.Fatalf("omitted name must be preserved, got %q", got.Name)
	}
	if !got.Archived {
		t.Fatal("archived should be set")
	}
}

func TestDeleteProjectRequiresOwner(t *testing.T) {
	f := newProjectFixture(t)

	if err := f.svc.DeleteProject(context.Background(), f.editor, "p001"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("editor must not delete, got %v", err)
	}
	if err := f.svc.DeleteProject(context.Background(), f.owner, "p001"); err != nil {
		t.Fatalf("owner delete: %v", err)
	}
}

func TestAddMemberRequiresOwnerAndKnownEmail(t *testing.T) {
	f := newProjectFixture(t)

	if _, err := f.svc.AddMember(context.Background(), f.editor, "p001", "newbie@example.com", domain.RoleEditor); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("editor must not add members, got %v", err)
	}

	if _, err := f.svc.AddMember(context.Background(), f.owner, "p001", "ghost@example.com", domain.RoleEditor); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown email must be not-found, got %v", err)
	}

	got, err := f.svc.AddMember(context.Background(), f.owner, "p001", "newbie@example.com", domain.RoleEditor)
	if err != nil {
		t.Fatalf("add member: %v", err)
	}
	if role, ok := got.RoleOf("u-newbie"); !ok || role != domain.RoleEditor {
		t.Fatalf("want newbie as editor, got %q present=%v", role, ok)
	}
}

func TestAddMemberRejectsUnknownRole(t *testing.T) {
	f := newProjectFixture(t)

	if _, err := f.svc.AddMember(context.Background(), f.owner, "p001", "newbie@example.com", "admin"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestCannotRemoveLastOwner(t *testing.T) {
	f := newProjectFixture(t)

	if _, err := f.svc.RemoveMember(context.Background(), f.owner, "p001", "u-owner"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("removing the only owner must fail, got %v", err)
	}

	if _, err := f.svc.AddMember(context.Background(), f.owner, "p001", "editor@example.com", domain.RoleOwner); err != nil {
		t.Fatalf("promote: %v", err)
	}
	if _, err := f.svc.RemoveMember(context.Background(), f.owner, "p001", "u-owner"); err != nil {
		t.Fatalf("removing an owner while another remains must work, got %v", err)
	}
}

func TestCannotDemoteLastOwner(t *testing.T) {
	f := newProjectFixture(t)

	if _, err := f.svc.SetMemberRole(context.Background(), f.owner, "p001", "u-owner", domain.RoleViewer); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("demoting the only owner would orphan the project, got %v", err)
	}
}

func TestSuperuserBypassesMembership(t *testing.T) {
	f := newProjectFixture(t)
	root := ports.Actor{UserID: "u-root", Superuser: true}

	if _, err := f.svc.GetProject(context.Background(), root, "api"); err != nil {
		t.Fatalf("superuser read: %v", err)
	}
	if err := f.svc.DeleteProject(context.Background(), root, "p001"); err != nil {
		t.Fatalf("superuser delete: %v", err)
	}
}

func TestStorageErrorsPropagate(t *testing.T) {
	f := newProjectFixture(t)
	f.repo.failOn = "List"

	if _, err := f.svc.ListProjects(context.Background(), f.owner, false); err == nil {
		t.Fatal("a storage failure must not be swallowed")
	}
}
