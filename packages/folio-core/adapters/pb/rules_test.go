package pb_test

import (
	"os"
	"testing"

	"github.com/pocketbase/dbx"
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
		newRecord(t, app, pb.ColProjectGrants, map[string]any{
			"project": s.project.Id, "user": user.Id, "role": role,
		})
	}

	s.ticket = newRecord(t, app, pb.ColIssues, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "ticket",
		"slug": "auth", "title": "Auth", "status": "open", "priority": "high",
	})
	s.entry = newRecord(t, app, pb.ColEntries, map[string]any{
		"domain": s.domain.Id, "project": s.project.Id, "kind": "journal",
		"slug": "note", "title": "Note",
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
		{"issue", s.ticket},
		{"entry", s.entry},
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

func TestProjectRepositoryLoadsEveryGrant(t *testing.T) {
	s := setup(t)
	newRecord(t, s.app, pb.ColClientMembers, map[string]any{"client": s.client.Id, "user": s.owner.Id, "role": "owner"})
	newRecord(t, s.app, pb.ColProjectGrants, map[string]any{"project": s.project.Id, "domain": s.domain.Id, "role": "editor"})
	newRecord(t, s.app, pb.ColProjectGrants, map[string]any{"project": s.project.Id, "user": s.stranger.Id, "role": "viewer"})

	project, err := pb.NewProjectRepository(s.app).GetByID(t.Context(), domain.ProjectID(s.project.Id))
	if err != nil {
		t.Fatal(err)
	}

	if len(project.ClientOwners) != 1 || string(project.ClientOwners[0]) != s.owner.Id {
		t.Errorf("client owners = %v, want the owner", project.ClientOwners)
	}
	if len(project.DomainGrants) != 1 || project.DomainGrants[0].Role != domain.RoleEditor ||
		len(project.DomainGrants[0].Members) != 3 {
		t.Errorf("domain grants = %+v, want one editor grant carrying the 3-member roster", project.DomainGrants)
	}
	if role, ok := project.RoleOf(domain.UserID(s.stranger.Id)); !ok || role != domain.RoleViewer {
		t.Errorf("stranger's personal grant = %q (found=%v), want viewer", role, ok)
	}
	for _, m := range project.Members {
		if string(m.UserID) == s.stranger.Id && m.Email != "stranger@test.local" {
			t.Errorf("stranger's grant carries email %q", m.Email)
		}
	}
}

func TestListFindsProjectsThroughEachGrant(t *testing.T) {
	s := setup(t)
	boss := newUser(t, s.app, "boss@test.local")
	solo := newUser(t, s.app, "solo@test.local")
	newRecord(t, s.app, pb.ColClientMembers, map[string]any{"client": s.client.Id, "user": boss.Id, "role": "owner"})
	newRecord(t, s.app, pb.ColProjectGrants, map[string]any{"project": s.project.Id, "domain": s.domain.Id, "role": "editor"})
	newRecord(t, s.app, pb.ColProjectGrants, map[string]any{"project": s.project.Id, "user": solo.Id, "role": "viewer"})
	newRecord(t, s.app, pb.ColProjects, map[string]any{"slug": "ungranted", "name": "Ungranted", "domain": s.domain.Id})

	repo := pb.NewProjectRepository(s.app)
	for _, tc := range []struct {
		who  *core.Record
		want int
	}{
		{boss, 2},
		{s.editor, 1},
		{solo, 1},
		{s.stranger, 0},
	} {
		listed, err := repo.List(t.Context(), domain.UserID(tc.who.Id), false)
		if err != nil {
			t.Fatal(err)
		}
		if len(listed) != tc.want {
			t.Errorf("%s lists %d projects, want %d", tc.who.GetString("email"), len(listed), tc.want)
		}
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
		Members:  []domain.Member{{UserID: domain.UserID(s.stranger.Id), Role: domain.RoleOwner}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if string(created.DomainID) != s.domain.Id {
		t.Errorf("project landed in domain %q, want %q", created.DomainID, s.domain.Id)
	}
	if role, ok := created.RoleOf(domain.UserID(s.stranger.Id)); !ok || role != domain.RoleOwner {
		t.Errorf("creator's personal grant = %q (found=%v), want owner", role, ok)
	}
	if len(created.DomainGrants) != 1 || created.DomainGrants[0].Role != domain.RoleEditor {
		t.Errorf("domain grants = %+v, want an editor grant to the project's domain", created.DomainGrants)
	}

	roster, err := s.app.FindAllRecords(pb.ColMembers, dbx.HashExp{"user": s.stranger.Id})
	if err != nil {
		t.Fatal(err)
	}
	if len(roster) != 0 {
		t.Error("creating a project in an existing domain added the creator to its roster")
	}
	domains, err := s.app.FindAllRecords(pb.ColDomains)
	if err != nil {
		t.Fatal(err)
	}
	if len(domains) != 2 {
		t.Errorf("domain count is %d, want 2: creating a project made a new one", len(domains))
	}
}

func TestCreateProjectWithoutDomainMakesCreatorClientOwner(t *testing.T) {
	s := setup(t)

	created, err := pb.NewProjectRepository(s.app).Create(t.Context(), domain.Project{
		ID: "proj00000000003", Slug: "solo", Name: "Solo",
		Members: []domain.Member{{UserID: domain.UserID(s.stranger.Id), Role: domain.RoleOwner}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(created.ClientOwners) != 1 || string(created.ClientOwners[0]) != s.stranger.Id {
		t.Errorf("client owners = %v, want the creator", created.ClientOwners)
	}
	if len(created.DomainGrants) != 1 || len(created.DomainGrants[0].Members) != 1 ||
		string(created.DomainGrants[0].Members[0]) != s.stranger.Id {
		t.Errorf("domain grants = %+v, want the new domain granted with the creator on its roster", created.DomainGrants)
	}
}

func TestMemberManagementWritesPersonalGrants(t *testing.T) {
	s := setup(t)
	repo := pb.NewProjectRepository(s.app)
	id := domain.ProjectID(s.project.Id)
	user := domain.UserID(s.stranger.Id)

	if err := repo.AddMember(t.Context(), id, user, domain.RoleViewer); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetMemberRole(t.Context(), id, user, domain.RoleEditor); err != nil {
		t.Fatal(err)
	}
	p, err := repo.GetByID(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	if role, ok := p.RoleOf(user); !ok || role != domain.RoleEditor {
		t.Fatalf("personal grant = %q (found=%v), want editor", role, ok)
	}

	if err := repo.RemoveMember(t.Context(), id, user); err != nil {
		t.Fatal(err)
	}
	p, err = repo.GetByID(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.RoleOf(user); ok {
		t.Error("personal grant survived RemoveMember")
	}
	roster, err := s.app.FindAllRecords(pb.ColMembers, dbx.HashExp{"user": s.stranger.Id})
	if err != nil {
		t.Fatal(err)
	}
	if len(roster) != 0 {
		t.Error("member management touched the domain roster")
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

func TestInterviewsAreWrittenOnlyThroughTheAPI(t *testing.T) {
	s := setup(t)
	interview := newRecord(t, s.app, pb.ColInterviews, map[string]any{
		"project": s.project.Id, "issue": s.ticket.Id, "topic": "Tree or graph", "agent_status": "waiting",
	})
	event := newRecord(t, s.app, pb.ColInterviewEvents, map[string]any{
		"project": s.project.Id, "interview": interview.Id, "seq": 1, "at": "2026-09-26 12:00:00.000Z",
	})

	for _, record := range []*core.Record{interview, event} {
		name := record.Collection().Name
		for _, user := range []*core.Record{s.owner, s.editor, s.viewer} {
			if !canView(t, s.app, record, user) {
				t.Errorf("%s cannot read %s", user.GetString("email"), name)
			}
		}
		if canView(t, s.app, record, s.stranger) {
			t.Errorf("stranger can read %s", name)
		}
		c := record.Collection()
		for label, rule := range map[string]*string{"create": c.CreateRule, "update": c.UpdateRule, "delete": c.DeleteRule} {
			if rule != nil {
				t.Errorf("%s has a %s rule %q: a write over REST would skip the interview rules", name, label, *rule)
			}
		}
	}
}
