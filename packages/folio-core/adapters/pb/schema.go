package pb

import (
	"github.com/pocketbase/pocketbase/core"
)

// Membership is modelled as its own collection rather than a multi-relation
// on the project, because a member carries a role. That also lets every other
// collection express its access rule as a single subquery against members,
// so ownership is checked the same way everywhere.
//
// Rules traverse the back-relation from the record's project to its
// membership rows (journ_members_via_project). An earlier version joined
// @collection.journ_members with two separate conditions, which let one row
// satisfy the project match and a different row satisfy the user match, and
// made journ_members' own rule reference journ_members. Both read as empty
// rather than as an error, so every direct collection listing returned [].
const (
	// ?= because a domain has many membership rows and only one has to belong
	// to the caller.
	memberOfDomain = "project.domain.journ_members_via_domain.user ?= @request.auth.id"
	// An aliased @collection join is rejected on create, where the record has
	// no id yet, so writes traverse the same relations as reads.
	writerOfDomain = "project.domain.journ_members_via_domain.user ?= @request.auth.id && project.domain.journ_members_via_domain.role ?!= 'viewer'"
)

func strPtr(s string) *string { return &s }

func autodates() []core.Field {
	return []core.Field{
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	}
}

func ensureClients(app core.App) error {
	if _, ok := find(app, ColClients); ok {
		return nil
	}
	c := core.NewBaseCollection(ColClients)
	c.Fields.Add(
		&core.TextField{Name: "slug", Required: true, Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`, Presentable: true},
		&core.TextField{Name: "name", Required: true, Max: 120, Presentable: true},
		&core.TextField{Name: "site", Max: 300},
		&core.TextField{Name: "logo", Max: 300},
		&core.TextField{Name: "descr", Max: 2000},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_clients_slug", true, "slug", "")

	return app.Save(c)
}

func ensureDomains(app core.App) error {
	if _, ok := find(app, ColDomains); ok {
		return nil
	}
	clients, err := app.FindCollectionByNameOrId(ColClients)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColDomains)
	c.Fields.Add(
		&core.RelationField{Name: "client", Required: true, CollectionId: clients.Id, CascadeDelete: true, MaxSelect: 1},
		&core.TextField{Name: "slug", Required: true, Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`, Presentable: true},
		&core.TextField{Name: "name", Required: true, Max: 120, Presentable: true},
		&core.TextField{Name: "descr", Max: 2000},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_domains_slug", true, "client, slug", "")
	c.AddIndex("idx_journ_domains_client", false, "client", "")

	return app.Save(c)
}

func ensureProjects(app core.App) error {
	if _, ok := find(app, ColProjects); ok {
		return nil
	}
	c := core.NewBaseCollection(ColProjects)
	c.Fields.Add(
		&core.TextField{Name: "slug", Required: true, Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`, Presentable: true},
		&core.TextField{Name: "name", Required: true, Max: 120, Presentable: true},
		&core.TextField{Name: "descr", Max: 2000},
		&core.BoolField{Name: "archived"},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_projects_slug", true, "slug", "")

	return app.Save(c)
}

func ensureMembers(app core.App) error {
	if _, ok := find(app, ColMembers); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColMembers)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "user", Required: true, CollectionId: users.Id, CascadeDelete: true, MaxSelect: 1},
		&core.SelectField{Name: "role", Required: true, MaxSelect: 1, Values: []string{"owner", "editor", "viewer"}},
	)
	c.Fields.Add(autodates()...)
	// One row per user per project: the unique index is what stops a
	// duplicate membership from producing two conflicting roles.
	c.AddIndex("idx_journ_members_unique", true, "project, user", "")
	c.AddIndex("idx_journ_members_user", false, "user", "")

	return app.Save(c)
}

func ensurePlans(app core.App) error {
	if _, ok := find(app, ColPlans); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	tickets, err := app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColPlans)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		ticketField(tickets),
		&core.TextField{Name: "title", Required: true, Max: 200, Presentable: true},
		&core.TextField{Name: "goal", Max: 2000},
		&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"draft", "active", "done", "abandoned"}},
		&core.JSONField{Name: "tags", MaxSize: 4000},
		&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_plans_project", false, "project", "")

	return app.Save(c)
}

func ensureTickets(app core.App) error {
	if _, ok := find(app, ColTickets); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColTickets)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		&core.TextField{Name: "slug", Required: true, Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`},
		&core.TextField{Name: "title", Required: true, Max: 200, Presentable: true},
		&core.EditorField{Name: "body", MaxSize: 500000},
		&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"open", "in_progress", "blocked", "closed", "cancelled"}},
		&core.SelectField{Name: "priority", Required: true, MaxSelect: 1, Values: []string{"low", "medium", "high"}},
		// The assignee is not cascade-deleted: losing the account should not
		// take the ticket with it.
		&core.RelationField{Name: "assignee", CollectionId: users.Id, CascadeDelete: false, MaxSelect: 1},
		&core.JSONField{Name: "tags", MaxSize: 4000},
		&core.TextField{Name: "external_ref", Max: 200},
		&core.JSONField{Name: "depends_on", MaxSize: 4000},
		&core.SelectField{Name: "wayfinder", MaxSelect: 1, Values: wayfinderValues},
		&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
	)
	c.Fields.Add(autodates()...)
	// Slugs address a ticket within its project, so uniqueness is per project.
	c.AddIndex("idx_journ_tickets_slug", true, "project, slug", "")
	c.AddIndex("idx_journ_tickets_status", false, "project, status", "")
	c.AddIndex("idx_journ_tickets_assignee", false, "assignee", "")

	return app.Save(c)
}

var wayfinderValues = []string{"map", "research", "prototype", "grilling", "task"}

// ticketField is the nullable back-reference every child collection carries.
// Deleting a ticket detaches its children rather than destroying them, so the
// relation must not cascade.
func ticketField(tickets *core.Collection) *core.RelationField {
	return &core.RelationField{Name: "ticket", CollectionId: tickets.Id, CascadeDelete: false, MaxSelect: 1}
}

func ensureIssues(app core.App) error {
	if _, ok := find(app, ColIssues); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}
	plans, err := app.FindCollectionByNameOrId(ColPlans)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColIssues)
	c.Fields.Add(
		&core.RelationField{Name: "domain", Required: true, CollectionId: domains.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		&core.SelectField{Name: "kind", Required: true, MaxSelect: 1, Values: []string{"ticket", "todo"}},
		&core.RelationField{Name: "plan", CollectionId: plans.Id, CascadeDelete: false, MaxSelect: 1},
		&core.TextField{Name: "slug", Required: true, Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`},
		&core.TextField{Name: "title", Required: true, Max: 200, Presentable: true},
		&core.EditorField{Name: "body", MaxSize: 500000},
		&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: issueStatuses},
		&core.SelectField{Name: "priority", Required: true, MaxSelect: 1, Values: []string{"low", "medium", "high"}},
		&core.NumberField{Name: "size", OnlyInt: true},
		&core.RelationField{Name: "assignee", CollectionId: users.Id, CascadeDelete: false, MaxSelect: 1},
		&core.JSONField{Name: "tags", MaxSize: 4000},
		&core.NumberField{Name: "position", OnlyInt: true},
		&core.DateField{Name: "due_date"},
		&core.SelectField{Name: "wayfinder", MaxSelect: 1, Values: wayfinderValues},
		&core.TextField{Name: "external_ref", Max: 200},
		&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_issues_slug", true, "project, slug", "")
	c.AddIndex("idx_journ_issues_status", false, "project, kind, status", "")
	c.AddIndex("idx_journ_issues_plan", false, "plan", "")
	c.AddIndex("idx_journ_issues_assignee", false, "assignee", "")
	c.AddIndex("idx_journ_issues_domain", false, "domain", "")

	return app.Save(c)
}

var issueStatuses = []string{"open", "in_progress", "blocked", "done", "cancelled"}

func ensureIssueLinks(app core.App) error {
	if _, ok := find(app, ColLinks); ok {
		return nil
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}
	issues, err := app.FindCollectionByNameOrId(ColIssues)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColLinks)
	c.Fields.Add(
		&core.RelationField{Name: "domain", Required: true, CollectionId: domains.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "from", Required: true, CollectionId: issues.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "to", Required: true, CollectionId: issues.Id, CascadeDelete: true, MaxSelect: 1},
		&core.SelectField{Name: "kind", Required: true, MaxSelect: 1, Values: []string{"blocks", "relates", "parent"}},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_links_unique", true, "[[from]], [[to]], kind", "")
	c.AddIndex("idx_journ_links_to", false, "[[to]], kind", "")

	return app.Save(c)
}

func ensureTodos(app core.App) error {
	if _, ok := find(app, ColTodos); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	plans, err := app.FindCollectionByNameOrId(ColPlans)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	tickets, err := app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColTodos)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		ticketField(tickets),
		// Deleting a plan detaches its todos rather than destroying them, so
		// this relation must not cascade.
		&core.RelationField{Name: "plan", CollectionId: plans.Id, CascadeDelete: false, MaxSelect: 1},
		&core.TextField{Name: "title", Required: true, Max: 200, Presentable: true},
		&core.TextField{Name: "details", Max: 2000},
		&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"pending", "in_progress", "done", "blocked", "cancelled"}},
		&core.SelectField{Name: "priority", Required: true, MaxSelect: 1, Values: []string{"low", "medium", "high"}},
		&core.JSONField{Name: "tags", MaxSize: 4000},
		&core.NumberField{Name: "position"},
		&core.JSONField{Name: "depends_on", MaxSize: 4000},
		&core.DateField{Name: "due_date"},
		&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_todos_project", false, "project", "")
	c.AddIndex("idx_journ_todos_plan", false, "plan", "")
	c.AddIndex("idx_journ_todos_status", false, "project, status", "")

	return app.Save(c)
}

func ensureJournal(app core.App) error {
	if _, ok := find(app, ColJournal); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	plans, err := app.FindCollectionByNameOrId(ColPlans)
	if err != nil {
		return err
	}
	todos, err := app.FindCollectionByNameOrId(ColTodos)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	tickets, err := app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColJournal)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		ticketField(tickets),
		&core.RelationField{Name: "plan", CollectionId: plans.Id, CascadeDelete: false, MaxSelect: 1},
		&core.RelationField{Name: "todo", CollectionId: todos.Id, CascadeDelete: false, MaxSelect: 1},
		&core.TextField{Name: "slug", Required: true, Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`},
		&core.TextField{Name: "title", Required: true, Max: 200, Presentable: true},
		&core.EditorField{Name: "body", MaxSize: 500000},
		&core.TextField{Name: "branch", Max: 200},
		&core.TextField{Name: "pr", Max: 200},
		&core.TextField{Name: "external_ref", Max: 200},
		&core.JSONField{Name: "meta", MaxSize: 100000},
		&core.JSONField{Name: "tags", MaxSize: 4000},
		&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_journal_project_created", false, "project, created", "")
	c.AddIndex("idx_journ_journal_plan", false, "plan", "")
	c.AddIndex("idx_journ_journal_todo", false, "todo", "")
	c.AddIndex("idx_journ_journal_branch", false, "branch", "")
	c.AddIndex("idx_journ_journal_ticket", false, "ticket", "")
	c.AddIndex("idx_journ_journal_external_ref", false, "external_ref", "")
	c.AddIndex("idx_journ_journal_slug", true, "project, slug", "")

	return app.Save(c)
}

func ensureEntries(app core.App) error {
	if _, ok := find(app, ColEntries); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}
	issues, err := app.FindCollectionByNameOrId(ColIssues)
	if err != nil {
		return err
	}
	plans, err := app.FindCollectionByNameOrId(ColPlans)
	if err != nil {
		return err
	}
	cycles, err := app.FindCollectionByNameOrId(ColCycles)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColEntries)
	c.Fields.Add(
		&core.RelationField{Name: "domain", Required: true, CollectionId: domains.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		&core.SelectField{Name: "kind", Required: true, MaxSelect: 1, Values: entryKinds},
		&core.RelationField{Name: "issue", CollectionId: issues.Id, CascadeDelete: false, MaxSelect: 1},
		&core.RelationField{Name: "plan", CollectionId: plans.Id, CascadeDelete: false, MaxSelect: 1},
		&core.RelationField{Name: "cycle", CollectionId: cycles.Id, CascadeDelete: false, MaxSelect: 1},
		&core.TextField{Name: "slug", Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`},
		&core.TextField{Name: "title", Max: 200, Presentable: true},
		&core.EditorField{Name: "body", MaxSize: 500000},
		&core.TextField{Name: "branch", Max: 200},
		&core.TextField{Name: "pr", Max: 200},
		&core.TextField{Name: "external_ref", Max: 200},
		&core.JSONField{Name: "meta", MaxSize: 100000},
		&core.JSONField{Name: "tags", MaxSize: 4000},
		&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_entries_slug", true, "project, slug", "slug != ''")
	c.AddIndex("idx_journ_entries_kind", false, "project, kind, created", "")
	c.AddIndex("idx_journ_entries_issue", false, "issue", "")
	c.AddIndex("idx_journ_entries_plan", false, "plan", "")
	c.AddIndex("idx_journ_entries_branch", false, "branch", "")
	c.AddIndex("idx_journ_entries_external_ref", false, "external_ref", "")

	return app.Save(c)
}

var entryKinds = []string{"journal", "doc", "log"}

func ensureTags(app core.App) error {
	if _, ok := find(app, ColTags); ok {
		return nil
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColTags)
	c.Fields.Add(
		&core.RelationField{Name: "domain", Required: true, CollectionId: domains.Id, CascadeDelete: true, MaxSelect: 1},
		&core.TextField{Name: "slug", Required: true, Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`, Presentable: true},
		&core.TextField{Name: "name", Required: true, Max: 60, Presentable: true},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_tags_slug", true, "domain, slug", "")

	return app.Save(c)
}

func ensureTagJoins(app core.App) error {
	tags, err := app.FindCollectionByNameOrId(ColTags)
	if err != nil {
		return err
	}

	if _, ok := find(app, ColIssueTags); !ok {
		issues, err := app.FindCollectionByNameOrId(ColIssues)
		if err != nil {
			return err
		}
		c := core.NewBaseCollection(ColIssueTags)
		c.Fields.Add(
			&core.RelationField{Name: "issue", Required: true, CollectionId: issues.Id, CascadeDelete: true, MaxSelect: 1},
			&core.RelationField{Name: "tag", Required: true, CollectionId: tags.Id, CascadeDelete: true, MaxSelect: 1},
		)
		c.Fields.Add(autodates()...)
		c.AddIndex("idx_journ_issue_tags_unique", true, "issue, tag", "")
		c.AddIndex("idx_journ_issue_tags_tag", false, "tag", "")
		if err := app.Save(c); err != nil {
			return err
		}
	}

	if _, ok := find(app, ColEntryTags); !ok {
		entries, err := app.FindCollectionByNameOrId(ColEntries)
		if err != nil {
			return err
		}
		c := core.NewBaseCollection(ColEntryTags)
		c.Fields.Add(
			&core.RelationField{Name: "entry", Required: true, CollectionId: entries.Id, CascadeDelete: true, MaxSelect: 1},
			&core.RelationField{Name: "tag", Required: true, CollectionId: tags.Id, CascadeDelete: true, MaxSelect: 1},
		)
		c.Fields.Add(autodates()...)
		c.AddIndex("idx_journ_entry_tags_unique", true, "entry, tag", "")
		c.AddIndex("idx_journ_entry_tags_tag", false, "tag", "")
		if err := app.Save(c); err != nil {
			return err
		}
	}

	return nil
}

func ensureCycles(app core.App) error {
	if _, ok := find(app, ColCycles); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}
	tickets, err := app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColCycles)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		&core.RelationField{Name: "ticket", Required: true, CollectionId: tickets.Id, CascadeDelete: true, MaxSelect: 1},
		&core.NumberField{Name: "ordinal", Required: true},
		&core.SelectField{Name: "phase", Required: true, MaxSelect: 1, Values: []string{"plan", "do", "check", "act"}},
		&core.TextField{Name: "resolution", Max: 2000},
		&core.DateField{Name: "closed_at"},
		&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
	)
	c.Fields.Add(autodates()...)
	c.AddIndex("idx_journ_cycles_ticket", false, "ticket, ordinal", "")

	return app.Save(c)
}

func ensureWorkLogs(app core.App) error {
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}
	tickets, err := app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return err
	}
	plans, err := app.FindCollectionByNameOrId(ColPlans)
	if err != nil {
		return err
	}
	todos, err := app.FindCollectionByNameOrId(ColTodos)
	if err != nil {
		return err
	}
	cycles, err := app.FindCollectionByNameOrId(ColCycles)
	if err != nil {
		return err
	}

	projectField := func() *core.RelationField {
		return &core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1}
	}
	bodyField := func() *core.EditorField {
		return &core.EditorField{Name: "body", MaxSize: 500000}
	}
	authorField := func() *core.RelationField {
		return &core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1}
	}

	if _, ok := find(app, ColTicketLogs); !ok {
		c := core.NewBaseCollection(ColTicketLogs)
		c.Fields.Add(
			projectField(),
			&core.RelationField{Name: "ticket", Required: true, CollectionId: tickets.Id, CascadeDelete: true, MaxSelect: 1},
			&core.RelationField{Name: "cycle", CollectionId: cycles.Id, CascadeDelete: false, MaxSelect: 1},
			bodyField(),
			authorField(),
		)
		c.Fields.Add(autodates()...)
		c.AddIndex("idx_journ_ticket_logs_ticket", false, "ticket, created", "")
		c.AddIndex("idx_journ_ticket_logs_cycle", false, "cycle", "")
		if err := app.Save(c); err != nil {
			return err
		}
	}

	if _, ok := find(app, ColPlanLogs); !ok {
		c := core.NewBaseCollection(ColPlanLogs)
		c.Fields.Add(
			projectField(),
			&core.RelationField{Name: "plan", Required: true, CollectionId: plans.Id, CascadeDelete: true, MaxSelect: 1},
			bodyField(),
			authorField(),
		)
		c.Fields.Add(autodates()...)
		c.AddIndex("idx_journ_plan_logs_plan", false, "plan, created", "")
		if err := app.Save(c); err != nil {
			return err
		}
	}

	if _, ok := find(app, ColTodoLogs); !ok {
		c := core.NewBaseCollection(ColTodoLogs)
		c.Fields.Add(
			projectField(),
			&core.RelationField{Name: "todo", Required: true, CollectionId: todos.Id, CascadeDelete: true, MaxSelect: 1},
			bodyField(),
			authorField(),
		)
		c.Fields.Add(autodates()...)
		c.AddIndex("idx_journ_todo_logs_todo", false, "todo, created", "")
		if err := app.Save(c); err != nil {
			return err
		}
	}

	return nil
}

func ensureDocs(app core.App) error {
	if _, ok := find(app, ColDocs); ok {
		return nil
	}
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	users, err := app.FindCollectionByNameOrId(ColUsers)
	if err != nil {
		return err
	}

	tickets, err := app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return err
	}

	c := core.NewBaseCollection(ColDocs)
	c.Fields.Add(
		&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true, MaxSelect: 1},
		ticketField(tickets),
		&core.TextField{Name: "slug", Required: true, Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`},
		&core.TextField{Name: "title", Required: true, Max: 200, Presentable: true},
		&core.EditorField{Name: "body", MaxSize: 500000},
		&core.JSONField{Name: "tags", MaxSize: 4000},
		&core.RelationField{Name: "created_by", CollectionId: users.Id, MaxSelect: 1},
	)
	c.Fields.Add(autodates()...)
	// Slugs address a doc within its project, so uniqueness is per project.
	c.AddIndex("idx_journ_docs_slug", true, "project, slug", "")

	return app.Save(c)
}

// applyRules sets the access rules once every collection exists. Splitting
// this from creation is what lets a rule reference journ_members: PocketBase
// resolves @collection references at save time, so the target must already be
// there.
//
// These rules govern direct REST access to the collections. The folio API
// checks the same permissions in its service layer, so they are a second line
// of defence rather than the only one.
func applyRules(app core.App) error {
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	// Creating a project is open to any authenticated user, who becomes its
	// first owner through the membership row written alongside.
	memberOfThis := "domain.journ_members_via_domain.user ?= @request.auth.id"
	ownerOfThis := "domain.journ_members_via_domain.user ?= @request.auth.id && domain.journ_members_via_domain.role ?= 'owner'"
	projects.ListRule = strPtr(memberOfThis)
	projects.ViewRule = strPtr(memberOfThis)
	projects.CreateRule = strPtr("@request.auth.id != ''")
	projects.UpdateRule = strPtr(ownerOfThis)
	projects.DeleteRule = strPtr(ownerOfThis)
	if err := app.Save(projects); err != nil {
		return err
	}

	// The domain holds the membership rows, so it checks them directly.
	clients, err := app.FindCollectionByNameOrId(ColClients)
	if err != nil {
		return err
	}
	clientVisible := "journ_domains_via_client.journ_members_via_domain.user ?= @request.auth.id"
	clients.ListRule = strPtr(clientVisible)
	clients.ViewRule = strPtr(clientVisible)
	clients.CreateRule = strPtr("@request.auth.id != ''")
	if err := app.Save(clients); err != nil {
		return err
	}

	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}
	domainVisible := "journ_members_via_domain.user ?= @request.auth.id"
	domains.ListRule = strPtr(domainVisible)
	domains.ViewRule = strPtr(domainVisible)
	domains.CreateRule = strPtr("@request.auth.id != ''")
	if err := app.Save(domains); err != nil {
		return err
	}

	members, err := app.FindCollectionByNameOrId(ColMembers)
	if err != nil {
		return err
	}
	// A membership row already names its user, so reads need no join. Listing
	// the rest of a project's roster goes through the folio API, which checks
	// membership in the service layer.
	ownRow := "user = @request.auth.id"
	ownerOfOwnDomain := "domain.journ_members_via_domain.user ?= @request.auth.id && domain.journ_members_via_domain.role ?= 'owner'"
	members.ListRule = strPtr(ownRow)
	members.ViewRule = strPtr(ownRow)
	members.CreateRule = strPtr(ownerOfOwnDomain)
	members.UpdateRule = strPtr(ownerOfOwnDomain)
	members.DeleteRule = strPtr(ownerOfOwnDomain)
	if err := app.Save(members); err != nil {
		return err
	}

	// The join and link rows carry no project of their own, so they are
	// reached through the row they attach to.
	for _, join := range []struct{ collection, via string }{
		{ColLinks, "from"},
		{ColIssueTags, "issue"},
		{ColEntryTags, "entry"},
	} {
		c, err := app.FindCollectionByNameOrId(join.collection)
		if err != nil {
			return err
		}
		read := join.via + ".project.domain.journ_members_via_domain.user ?= @request.auth.id"
		write := read + " && " + join.via + ".project.domain.journ_members_via_domain.role ?!= 'viewer'"
		c.ListRule = strPtr(read)
		c.ViewRule = strPtr(read)
		c.CreateRule = strPtr(write)
		c.UpdateRule = strPtr(write)
		c.DeleteRule = strPtr(write)
		if err := app.Save(c); err != nil {
			return err
		}
	}

	tags, err := app.FindCollectionByNameOrId(ColTags)
	if err != nil {
		return err
	}
	tagVisible := "domain.journ_members_via_domain.user ?= @request.auth.id"
	tagWritable := tagVisible + " && domain.journ_members_via_domain.role ?!= 'viewer'"
	tags.ListRule = strPtr(tagVisible)
	tags.ViewRule = strPtr(tagVisible)
	tags.CreateRule = strPtr(tagWritable)
	tags.UpdateRule = strPtr(tagWritable)
	tags.DeleteRule = strPtr(tagWritable)
	if err := app.Save(tags); err != nil {
		return err
	}

	for _, name := range []string{ColTickets, ColPlans, ColTodos, ColDocs, ColJournal, ColIssues, ColEntries, ColCycles, ColTicketLogs, ColPlanLogs, ColTodoLogs} {
		c, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			return err
		}
		c.ListRule = strPtr(memberOfDomain)
		c.ViewRule = strPtr(memberOfDomain)
		c.CreateRule = strPtr(writerOfDomain)
		c.UpdateRule = strPtr(writerOfDomain)
		c.DeleteRule = strPtr(writerOfDomain)
		if err := app.Save(c); err != nil {
			return err
		}
	}

	return nil
}
