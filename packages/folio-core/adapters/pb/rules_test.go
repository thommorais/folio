package pb_test

import (
	"os"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	_ "github.com/pocketbase/pocketbase/migrations"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
)

func newApp(t *testing.T) core.App {
	t.Helper()

	dir, err := os.MkdirTemp("", "folio-rules")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	app := core.NewBaseApp(core.BaseAppConfig{DataDir: dir})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })

	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	if err := pb.Register(app); err != nil {
		t.Fatal(err)
	}
	return app
}

func newUser(t *testing.T, app core.App, email string) *core.Record {
	t.Helper()

	users, err := app.FindCollectionByNameOrId(pb.ColUsers)
	if err != nil {
		t.Fatal(err)
	}
	r := core.NewRecord(users)
	r.Set("email", email)
	r.Set("password", "password12345")
	r.Set("verified", true)
	if err := app.Save(r); err != nil {
		t.Fatal(err)
	}
	return r
}

func newRecord(t *testing.T, app core.App, collection string, values map[string]any) *core.Record {
	t.Helper()

	c, err := app.FindCollectionByNameOrId(collection)
	if err != nil {
		t.Fatal(err)
	}
	r := core.NewRecord(c)
	for k, v := range values {
		r.Set(k, v)
	}
	if err := app.Save(r); err != nil {
		t.Fatalf("save %s: %v", collection, err)
	}
	return r
}

// scenario is a domain with one member of each role, plus a stranger.
type scenario struct {
	app                    core.App
	client, domain, other  *core.Record
	project, ticket, entry *core.Record
	owner, editor, viewer  *core.Record
	stranger               *core.Record
}

func setup(t *testing.T) scenario {
	t.Helper()

	app := newApp(t)
	s := scenario{app: app}

	s.owner = newUser(t, app, "owner@test.local")
	s.editor = newUser(t, app, "editor@test.local")
	s.viewer = newUser(t, app, "viewer@test.local")
	s.stranger = newUser(t, app, "stranger@test.local")

	s.client = newRecord(t, app, pb.ColClients, map[string]any{"slug": "acme", "name": "Acme"})
	s.domain = newRecord(t, app, pb.ColDomains, map[string]any{
		"client": s.client.Id, "slug": "web", "name": "Web",
	})
	s.other = newRecord(t, app, pb.ColDomains, map[string]any{
		"client": s.client.Id, "slug": "infra", "name": "Infra",
	})

	s.project = newRecord(t, app, pb.ColProjects, map[string]any{
		"slug": "redesign", "name": "Redesign", "domain": s.domain.Id,
	})

	for user, role := range map[*core.Record]string{
		s.owner: "owner", s.editor: "editor", s.viewer: "viewer",
	} {
		newRecord(t, app, pb.ColMembers, map[string]any{
			"domain": s.domain.Id, "project": s.project.Id, "user": user.Id, "role": role,
		})
	}

	s.ticket = newRecord(t, app, pb.ColTickets, map[string]any{
		"project": s.project.Id, "slug": "auth", "title": "Auth",
		"status": "open", "priority": "high",
	})
	s.entry = newRecord(t, app, pb.ColJournal, map[string]any{
		"project": s.project.Id, "slug": "note", "title": "Note",
	})

	return s
}

func info(user *core.Record, method string) *core.RequestInfo {
	return &core.RequestInfo{Auth: user, Method: method, Context: "default"}
}

// A rule that resolves to nothing denies everyone and is indistinguishable
// from one that correctly denies, so every case here pairs a denial with a
// grant.
func canView(t *testing.T, app core.App, record, user *core.Record) bool {
	t.Helper()

	ok, err := app.CanAccessRecord(record, info(user, "GET"), record.Collection().ViewRule)
	if err != nil {
		t.Fatalf("view rule on %s: %v", record.Collection().Name, err)
	}
	return ok
}

func TestMemberCanReadProjectContent(t *testing.T) {
	s := setup(t)

	for _, tc := range []struct {
		name   string
		record *core.Record
	}{
		{"project", s.project},
		{"ticket", s.ticket},
		{"journal", s.entry},
		{"domain", s.domain},
		{"client", s.client},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, user := range []*core.Record{s.owner, s.editor, s.viewer} {
				if !canView(t, s.app, tc.record, user) {
					t.Errorf("%s cannot read %s, rule resolves empty for a member",
						user.GetString("email"), tc.name)
				}
			}
			if canView(t, s.app, tc.record, s.stranger) {
				t.Errorf("stranger can read %s", tc.name)
			}
		})
	}
}

func TestUnrelatedDomainIsHidden(t *testing.T) {
	s := setup(t)

	if canView(t, s.app, s.other, s.owner) {
		t.Error("owner can see a domain holding none of their projects")
	}
}

func TestProjectRepositoryResolvesRosterThroughDomain(t *testing.T) {
	s := setup(t)

	repo := pb.NewProjectRepository(s.app)
	project, err := repo.GetByID(t.Context(), domain.ProjectID(s.project.Id))
	if err != nil {
		t.Fatal(err)
	}

	if len(project.Members) != 3 {
		t.Fatalf("roster has %d members, want 3", len(project.Members))
	}
	role, ok := project.RoleOf(domain.UserID(s.editor.Id))
	if !ok || role != domain.RoleEditor {
		t.Errorf("editor resolved to %q (found=%v), want editor", role, ok)
	}
	if _, ok := project.RoleOf(domain.UserID(s.stranger.Id)); ok {
		t.Error("stranger appears in the roster")
	}
}

func TestSiblingProjectSharesRoster(t *testing.T) {
	s := setup(t)

	sibling := newRecord(t, s.app, pb.ColProjects, map[string]any{
		"slug": "second", "name": "Second", "domain": s.domain.Id,
	})

	repo := pb.NewProjectRepository(s.app)
	project, err := repo.GetByID(t.Context(), domain.ProjectID(sibling.Id))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := project.RoleOf(domain.UserID(s.owner.Id)); !ok {
		t.Error("owner of the domain is not a member of its sibling project")
	}

	listed, err := repo.List(t.Context(), domain.UserID(s.owner.Id), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 {
		t.Errorf("List returned %d projects, want 2", len(listed))
	}
}

func TestCreateProjectIntoExistingDomain(t *testing.T) {
	s := setup(t)

	repo := pb.NewProjectRepository(s.app)
	created, err := repo.Create(t.Context(), domain.Project{
		ID:       "proj00000000002",
		DomainID: domain.DomainID(s.domain.Id),
		Slug:     "second",
		Name:     "Second",
	})
	if err != nil {
		t.Fatal(err)
	}

	if string(created.DomainID) != s.domain.Id {
		t.Errorf("project landed in domain %q, want %q", created.DomainID, s.domain.Id)
	}
	if _, ok := created.RoleOf(domain.UserID(s.owner.Id)); !ok {
		t.Error("project does not inherit the domain roster")
	}

	domains, err := s.app.FindAllRecords(pb.ColDomains)
	if err != nil {
		t.Fatal(err)
	}
	if len(domains) != 2 {
		t.Errorf("domain count is %d, want 2: creating a project made a new one", len(domains))
	}
}

func TestEveryContentCollectionHasRules(t *testing.T) {
	app := newApp(t)

	for _, name := range []string{
		pb.ColIssues, pb.ColEntries, pb.ColLinks, pb.ColTags,
		pb.ColIssueTags, pb.ColEntryTags, pb.ColPlans, pb.ColCycles,
	} {
		c, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for label, rule := range map[string]*string{
			"list": c.ListRule, "view": c.ViewRule, "create": c.CreateRule,
			"update": c.UpdateRule, "delete": c.DeleteRule,
		} {
			if rule == nil {
				t.Errorf("%s has no %s rule, so only superusers can reach it", name, label)
			}
		}
	}
}

func TestMemberReadsIssuesAndEntriesOverRest(t *testing.T) {
	s := setup(t)

	issue := newRecord(t, s.app, pb.ColIssues, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "ticket",
		"slug": "an-issue", "title": "An issue", "status": "open", "priority": "medium",
	})
	entry := newRecord(t, s.app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "doc",
		"slug": "a-doc", "title": "A doc",
	})
	tag := newRecord(t, s.app, pb.ColTags, map[string]any{
		"domain": s.domain.Id, "slug": "backend", "name": "backend",
	})
	link := newRecord(t, s.app, pb.ColIssueTags, map[string]any{
		"issue": issue.Id, "tag": tag.Id,
	})

	for _, tc := range []struct {
		name   string
		record *core.Record
	}{
		{"issue", issue},
		{"entry", entry},
		{"tag", tag},
		{"issue tag link", link},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !canView(t, s.app, tc.record, s.editor) {
				t.Errorf("a member cannot read the %s", tc.name)
			}
			if canView(t, s.app, tc.record, s.stranger) {
				t.Errorf("a stranger can read the %s", tc.name)
			}
		})
	}
}
