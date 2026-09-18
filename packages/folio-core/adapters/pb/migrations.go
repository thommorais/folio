package pb

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain/rules"
)

// Register installs the folio schema. It is idempotent: an existing
// collection is left alone, so the app can call it on every boot.
//
// Access rules are expressed in PocketBase rule syntax and enforced by
// PocketBase itself for any direct REST access to these collections. The
// folio API additionally checks permissions in the service layer, so the
// rules here are the second line of defence rather than the only one.
// The rules reference journ_members, which cannot be resolved while that
// collection does not yet exist, so structure is created first and the access
// rules are applied in a second pass once every collection is present.
func Register(app core.App) error {
	if err := renameLogsToJournal(app); err != nil {
		return fmt.Errorf("rename journal: %w", err)
	}
	if err := ensureClients(app); err != nil {
		return fmt.Errorf("clients: %w", err)
	}
	if err := ensureDomains(app); err != nil {
		return fmt.Errorf("domains: %w", err)
	}
	if err := ensureProjects(app); err != nil {
		return fmt.Errorf("projects: %w", err)
	}
	if err := ensureMembers(app); err != nil {
		return fmt.Errorf("members: %w", err)
	}
	// Tickets come before plans, todos, logs and docs: each of those carries
	// a relation to this collection, so it has to exist first.
	if err := ensureTickets(app); err != nil {
		return fmt.Errorf("tickets: %w", err)
	}
	if err := ensurePlans(app); err != nil {
		return fmt.Errorf("plans: %w", err)
	}
	if err := ensureTodos(app); err != nil {
		return fmt.Errorf("todos: %w", err)
	}
	if err := ensureJournal(app); err != nil {
		return fmt.Errorf("journal: %w", err)
	}
	if err := ensureDocs(app); err != nil {
		return fmt.Errorf("docs: %w", err)
	}
	if err := ensureIssues(app); err != nil {
		return fmt.Errorf("issues: %w", err)
	}
	if err := ensureIssueLinks(app); err != nil {
		return fmt.Errorf("issue links: %w", err)
	}
	if err := ensureCycles(app); err != nil {
		return fmt.Errorf("cycles: %w", err)
	}

	if err := ensureWorkLogs(app); err != nil {
		return fmt.Errorf("work logs: %w", err)
	}
	if err := ensureEntries(app); err != nil {
		return fmt.Errorf("entries: %w", err)
	}
	if err := ensureTags(app); err != nil {
		return fmt.Errorf("tags: %w", err)
	}
	if err := ensureTagJoins(app); err != nil {
		return fmt.Errorf("tag joins: %w", err)
	}
	// Existing databases predate tickets: their collections were created by
	// an earlier Register and ensureX leaves them alone, so the new fields
	// are added in a separate pass.
	if err := alterForTickets(app); err != nil {
		return fmt.Errorf("alter: %w", err)
	}
	if err := alterForWayfinder(app); err != nil {
		return fmt.Errorf("alter wayfinder: %w", err)
	}
	if err := alterForJournalSlug(app); err != nil {
		return fmt.Errorf("alter journal slug: %w", err)
	}
	if err := backfillDomains(app); err != nil {
		return fmt.Errorf("backfill domains: %w", err)
	}
	if err := backfillMemberDomains(app); err != nil {
		return fmt.Errorf("backfill member domains: %w", err)
	}
	if err := backfillIssues(app); err != nil {
		return fmt.Errorf("backfill issues: %w", err)
	}
	if err := repointToIssues(app); err != nil {
		return fmt.Errorf("repoint to issues: %w", err)
	}
	if err := backfillEntries(app); err != nil {
		return fmt.Errorf("backfill entries: %w", err)
	}
	if err := backfillTags(app); err != nil {
		return fmt.Errorf("backfill tags: %w", err)
	}
	if err := relaxSupersededColumns(app); err != nil {
		return fmt.Errorf("relax superseded columns: %w", err)
	}
	if err := applyRules(app); err != nil {
		return fmt.Errorf("rules: %w", err)
	}
	return nil
}

func renameLogsToJournal(app core.App) error {
	c, ok := find(app, "journ_logs")
	if !ok {
		return nil
	}
	if _, taken := find(app, ColJournal); taken {
		return nil
	}

	c.Name = ColJournal
	renamed := make([]string, 0, len(c.Indexes))
	for _, idx := range c.Indexes {
		renamed = append(renamed, strings.ReplaceAll(idx, "idx_journ_logs", "idx_journ_journal"))
	}
	c.Indexes = renamed

	return app.Save(c)
}

// alterForTickets brings a pre-ticket database up to date: it adds the ticket
// relation to every child collection and renames the log's free-text ticket
// key to external_ref. Both steps are no-ops once applied, so Register stays
// safe to call on every boot.
func alterForTickets(app core.App) error {
	tickets, err := app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return err
	}

	for _, name := range []string{ColPlans, ColTodos, ColJournal, ColDocs} {
		c, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			return err
		}
		changed := false

		// The log's "ticket" column held a free-text tracker key before the
		// ticket entity existed. It is renamed rather than replaced: keeping
		// the field's id makes PocketBase rename the underlying column, so
		// the values survive. Dropping and re-adding would silently empty it.
		if text, isText := c.Fields.GetByName("ticket").(*core.TextField); isText {
			text.Name = "external_ref"
			changed = true
		}

		// Only once "ticket" is free can the relation take the name.
		if c.Fields.GetByName("ticket") == nil {
			c.Fields.Add(ticketField(tickets))
			changed = true
		}

		if changed {
			if err := app.Save(c); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	return nil
}

func alterForWayfinder(app core.App) error {
	c, err := app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return err
	}
	changed := false

	if c.Fields.GetByName("parent") == nil {
		c.Fields.Add(&core.RelationField{Name: "parent", CollectionId: c.Id, CascadeDelete: false, MaxSelect: 1})
		changed = true
	}
	if c.Fields.GetByName("depends_on") == nil {
		c.Fields.Add(&core.JSONField{Name: "depends_on", MaxSize: 4000})
		changed = true
	}
	if c.Fields.GetByName("wayfinder") == nil {
		c.Fields.Add(&core.SelectField{Name: "wayfinder", MaxSelect: 1, Values: wayfinderValues})
		changed = true
	}

	if !changed {
		return nil
	}
	c.AddIndex("idx_journ_tickets_parent", false, "project, parent", "")
	c.AddIndex("idx_journ_tickets_wayfinder", false, "wayfinder", "")
	return app.Save(c)
}

// backfillDomains gives every existing project a client and a domain. Each
// project becomes its own client rather than sharing one placeholder: merging
// by hand can undo that, a shared placeholder could not.
func backfillDomains(app core.App) error {
	projects, err := app.FindCollectionByNameOrId(ColProjects)
	if err != nil {
		return err
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}
	if projects.Fields.GetByName("domain") == nil {
		projects.Fields.Add(&core.RelationField{Name: "domain", CollectionId: domains.Id, CascadeDelete: false, MaxSelect: 1})
		if err := app.Save(projects); err != nil {
			return fmt.Errorf("add domain field: %w", err)
		}
	}

	records, err := app.FindAllRecords(ColProjects)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}

	clientSlugs, err := existingSlugs(app, ColClients)
	if err != nil {
		return err
	}
	domainSlugs, err := existingSlugs(app, ColDomains)
	if err != nil {
		return err
	}

	clients, err := app.FindCollectionByNameOrId(ColClients)
	if err != nil {
		return err
	}

	for _, project := range records {
		if project.GetString("domain") != "" {
			continue
		}

		base := project.GetString("slug")
		if base == "" {
			base = rules.Slugify(project.GetString("name"))
		}

		clientSlug := uniqueSlug(base, clientSlugs)
		clientSlugs[clientSlug] = true
		client := core.NewRecord(clients)
		client.Set("slug", clientSlug)
		client.Set("name", project.GetString("name"))
		if err := app.Save(client); err != nil {
			return fmt.Errorf("create client for %s: %w", project.Id, err)
		}

		domainSlug := uniqueSlug(base, domainSlugs)
		domainSlugs[domainSlug] = true
		domain := core.NewRecord(domains)
		domain.Set("client", client.Id)
		domain.Set("slug", domainSlug)
		domain.Set("name", project.GetString("name"))
		domain.Set("descr", project.GetString("descr"))
		if err := app.Save(domain); err != nil {
			return fmt.Errorf("create domain for %s: %w", project.Id, err)
		}

		project.Set("domain", domain.Id)
		if err := app.Save(project); err != nil {
			return fmt.Errorf("link project %s to domain: %w", project.Id, err)
		}
	}

	return nil
}

// backfillMemberDomains moves membership from the project to the domain. The
// project column is kept until the readers stop using it, so a rollback still
// finds its memberships. Two roles collapsing into one row keep the stronger:
// silently demoting someone is the worse failure.
func backfillMemberDomains(app core.App) error {
	members, err := app.FindCollectionByNameOrId(ColMembers)
	if err != nil {
		return err
	}
	domains, err := app.FindCollectionByNameOrId(ColDomains)
	if err != nil {
		return err
	}

	if members.Fields.GetByName("domain") == nil {
		members.Fields.Add(&core.RelationField{Name: "domain", CollectionId: domains.Id, CascadeDelete: true, MaxSelect: 1})
		// (project, user) would reject a user joining a second project in the
		// same domain.
		members.RemoveIndex("idx_journ_members_unique")
		if err := app.Save(members); err != nil {
			return fmt.Errorf("add domain field: %w", err)
		}
	}

	rows, err := app.FindAllRecords(ColMembers)
	if err != nil {
		return fmt.Errorf("load members: %w", err)
	}

	projectDomain := make(map[string]string)
	projects, err := app.FindAllRecords(ColProjects)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}
	for _, p := range projects {
		projectDomain[p.Id] = p.GetString("domain")
	}

	rank := map[string]int{"viewer": 0, "editor": 1, "owner": 2}
	type key struct{ domain, user string }
	seen := make(map[key]*core.Record)

	for _, row := range rows {
		domainID := row.GetString("domain")
		if domainID == "" {
			domainID = projectDomain[row.GetString("project")]
			if domainID == "" {
				return fmt.Errorf("member %s: project %s has no domain", row.Id, row.GetString("project"))
			}
			row.Set("domain", domainID)
			if err := app.Save(row); err != nil {
				return fmt.Errorf("set domain on member %s: %w", row.Id, err)
			}
		}

		k := key{domainID, row.GetString("user")}
		kept, ok := seen[k]
		if !ok {
			seen[k] = row
			continue
		}
		if rank[row.GetString("role")] > rank[kept.GetString("role")] {
			kept.Set("role", row.GetString("role"))
			if err := app.Save(kept); err != nil {
				return fmt.Errorf("promote member %s: %w", kept.Id, err)
			}
		}
		if err := app.Delete(row); err != nil {
			return fmt.Errorf("delete duplicate member %s: %w", row.Id, err)
		}
	}

	members, err = app.FindCollectionByNameOrId(ColMembers)
	if err != nil {
		return err
	}
	if field, ok := members.Fields.GetByName("domain").(*core.RelationField); ok && !field.Required {
		field.Required = true
		members.AddIndex("idx_journ_members_domain_unique", true, "domain, user", "")
		if err := app.Save(members); err != nil {
			return fmt.Errorf("require domain: %w", err)
		}
	}

	return nil
}

var ticketStatusToIssue = map[string]string{
	"open": "open", "in_progress": "in_progress", "blocked": "blocked",
	"closed": "done", "cancelled": "cancelled",
}

var todoStatusToIssue = map[string]string{
	"pending": "open", "in_progress": "in_progress", "blocked": "blocked",
	"done": "done", "cancelled": "cancelled",
}

// backfillIssues copies tickets and todos into journ_issues, keeping each
// record's id so existing references stay valid, and turns depends_on and
// parent into link rows.
func backfillIssues(app core.App) error {
	issues, err := app.FindCollectionByNameOrId(ColIssues)
	if err != nil {
		return err
	}
	links, err := app.FindCollectionByNameOrId(ColLinks)
	if err != nil {
		return err
	}

	projectDomain := make(map[string]string)
	projects, err := app.FindAllRecords(ColProjects)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}
	for _, p := range projects {
		projectDomain[p.Id] = p.GetString("domain")
	}

	migrated := make(map[string]bool)
	existing, err := app.FindAllRecords(ColIssues)
	if err != nil {
		return fmt.Errorf("load issues: %w", err)
	}
	for _, rec := range existing {
		migrated[rec.Id] = true
	}

	type pending struct {
		from, to, kind, domain string
	}
	var queued []pending

	// Todos have no slug of their own, and issues addresses every row by one.
	takenSlugs := make(map[string]map[string]bool)
	for _, rec := range existing {
		project := rec.GetString("project")
		if takenSlugs[project] == nil {
			takenSlugs[project] = make(map[string]bool)
		}
		takenSlugs[project][rec.GetString("slug")] = true
	}

	copyRow := func(src *core.Record, kind string, statuses map[string]string) error {
		if migrated[src.Id] {
			return nil
		}
		projectID := src.GetString("project")
		domainID := projectDomain[projectID]
		if domainID == "" {
			return fmt.Errorf("%s %s: project %s has no domain", kind, src.Id, projectID)
		}

		status, ok := statuses[src.GetString("status")]
		if !ok {
			return fmt.Errorf("%s %s: unmapped status %q", kind, src.Id, src.GetString("status"))
		}

		if takenSlugs[projectID] == nil {
			takenSlugs[projectID] = make(map[string]bool)
		}
		slug := src.GetString("slug")
		if slug == "" {
			slug = rules.Slugify(src.GetString("title"))
		}
		slug = uniqueSlug(slug, takenSlugs[projectID])
		takenSlugs[projectID][slug] = true

		rec := core.NewRecord(issues)
		rec.Id = src.Id
		rec.Set("domain", domainID)
		rec.Set("project", projectID)
		rec.Set("kind", kind)
		rec.Set("slug", slug)
		rec.Set("title", src.GetString("title"))
		rec.Set("status", status)
		rec.Set("priority", src.GetString("priority"))
		rec.Set("tags", src.GetString("tags"))
		rec.Set("created_by", src.GetString("created_by"))
		rec.Set("created", src.GetString("created"))
		rec.Set("updated", src.GetString("updated"))

		if kind == "ticket" {
			rec.Set("body", src.GetString("body"))
			rec.Set("wayfinder", src.GetString("wayfinder"))
			rec.Set("external_ref", src.GetString("external_ref"))
			rec.Set("assignee", src.GetString("assignee"))
		} else {
			rec.Set("body", src.GetString("details"))
			rec.Set("plan", src.GetString("plan"))
			rec.Set("position", src.GetInt("position"))
			if due := src.GetString("due_date"); due != "" {
				rec.Set("due_date", due)
			}
		}

		if err := app.Save(rec); err != nil {
			return fmt.Errorf("copy %s %s: %w", kind, src.Id, err)
		}
		migrated[src.Id] = true

		for _, dep := range strSlice(src, "depends_on") {
			queued = append(queued, pending{from: dep, to: src.Id, kind: "blocks", domain: domainID})
		}
		if parent := src.GetString("parent"); parent != "" {
			queued = append(queued, pending{from: src.Id, to: parent, kind: "parent", domain: domainID})
		}
		if kind == "todo" {
			if ticket := src.GetString("ticket"); ticket != "" {
				queued = append(queued, pending{from: src.Id, to: ticket, kind: "parent", domain: domainID})
			}
		}
		return nil
	}

	tickets, err := app.FindAllRecords(ColTickets)
	if err != nil {
		return fmt.Errorf("load tickets: %w", err)
	}
	for _, src := range tickets {
		if err := copyRow(src, "ticket", ticketStatusToIssue); err != nil {
			return err
		}
	}

	todos, err := app.FindAllRecords(ColTodos)
	if err != nil {
		return fmt.Errorf("load todos: %w", err)
	}
	for _, src := range todos {
		if err := copyRow(src, "todo", todoStatusToIssue); err != nil {
			return err
		}
	}

	// A link whose other end never migrated would fail the relation, so both
	// ends are checked once every row is in place.
	for _, q := range queued {
		if !migrated[q.from] || !migrated[q.to] {
			continue
		}
		dup, err := app.FindFirstRecordByFilter(
			ColLinks,
			"from = {:from} && to = {:to} && kind = {:kind}",
			dbx.Params{"from": q.from, "to": q.to, "kind": q.kind},
		)
		if err == nil && dup != nil {
			continue
		}
		rec := core.NewRecord(links)
		rec.Set("domain", q.domain)
		rec.Set("from", q.from)
		rec.Set("to", q.to)
		rec.Set("kind", q.kind)
		if err := app.Save(rec); err != nil {
			return fmt.Errorf("link %s->%s: %w", q.from, q.to, err)
		}
	}

	return nil
}

// repointToIssues gives each child collection an issue relation pointing at
// journ_issues. PocketBase refuses to retarget an existing relation, so the
// new field is added alongside, filled from the old one, and the old field is
// left in place for the rollback window. The stored ids are already valid:
// backfillIssues copied every ticket and todo under its own id.
func repointToIssues(app core.App) error {
	issues, err := app.FindCollectionByNameOrId(ColIssues)
	if err != nil {
		return err
	}

	type source struct {
		collection string
		from       string
		cascade    bool
	}
	for _, src := range []source{
		{ColPlans, "ticket", false},
		{ColJournal, "ticket", false},
		{ColDocs, "ticket", false},
		{ColCycles, "ticket", true},
		{ColTicketLogs, "ticket", true},
		{ColTodoLogs, "todo", true},
	} {
		c, err := app.FindCollectionByNameOrId(src.collection)
		if err != nil {
			return err
		}
		if c.Fields.GetByName("issue") == nil {
			c.Fields.Add(&core.RelationField{
				Name: "issue", CollectionId: issues.Id,
				CascadeDelete: src.cascade, MaxSelect: 1,
			})
			c.AddIndex("idx_"+src.collection+"_issue", false, "issue", "")
			if err := app.Save(c); err != nil {
				return fmt.Errorf("add issue field to %s: %w", src.collection, err)
			}
		}

		rows, err := app.FindAllRecords(src.collection)
		if err != nil {
			return fmt.Errorf("load %s: %w", src.collection, err)
		}
		for _, row := range rows {
			if row.GetString("issue") != "" {
				continue
			}
			ref := row.GetString(src.from)
			if ref == "" {
				continue
			}
			if _, err := app.FindRecordById(ColIssues, ref); err != nil {
				continue
			}
			row.Set("issue", ref)
			if err := app.Save(row); err != nil {
				return fmt.Errorf("fill issue on %s %s: %w", src.collection, row.Id, err)
			}
		}
	}

	j, err := app.FindCollectionByNameOrId(ColJournal)
	if err != nil {
		return err
	}
	if j.Fields.GetByName("issue_todo") == nil && j.Fields.GetByName("todo") != nil {
		j.Fields.Add(&core.RelationField{Name: "issue_todo", CollectionId: issues.Id, MaxSelect: 1})
		if err := app.Save(j); err != nil {
			return fmt.Errorf("add issue_todo to journal: %w", err)
		}
		rows, err := app.FindAllRecords(ColJournal)
		if err != nil {
			return fmt.Errorf("load journal: %w", err)
		}
		for _, row := range rows {
			ref := row.GetString("todo")
			if ref == "" || row.GetString("issue_todo") != "" {
				continue
			}
			if _, err := app.FindRecordById(ColIssues, ref); err != nil {
				continue
			}
			row.Set("issue_todo", ref)
			if err := app.Save(row); err != nil {
				return fmt.Errorf("fill issue_todo on %s: %w", row.Id, err)
			}
		}
	}

	return nil
}

func backfillEntries(app core.App) error {
	entries, err := app.FindCollectionByNameOrId(ColEntries)
	if err != nil {
		return err
	}

	projectDomain := make(map[string]string)
	projects, err := app.FindAllRecords(ColProjects)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}
	for _, p := range projects {
		projectDomain[p.Id] = p.GetString("domain")
	}

	migrated := make(map[string]bool)
	existing, err := app.FindAllRecords(ColEntries)
	if err != nil {
		return fmt.Errorf("load entries: %w", err)
	}
	for _, rec := range existing {
		migrated[rec.Id] = true
	}

	copyRow := func(src *core.Record, kind string, rich bool) error {
		if migrated[src.Id] {
			return nil
		}
		projectID := src.GetString("project")
		domainID := projectDomain[projectID]
		if domainID == "" {
			return fmt.Errorf("%s %s: project %s has no domain", kind, src.Id, projectID)
		}

		rec := core.NewRecord(entries)
		rec.Id = src.Id
		rec.Set("domain", domainID)
		rec.Set("project", projectID)
		rec.Set("kind", kind)
		rec.Set("body", src.GetString("body"))
		rec.Set("created_by", src.GetString("created_by"))
		rec.Set("created", src.GetString("created"))
		rec.Set("updated", src.GetString("updated"))

		if issue := src.GetString("issue"); issue != "" {
			rec.Set("issue", issue)
		}
		if plan := src.GetString("plan"); plan != "" {
			rec.Set("plan", plan)
		}
		if cycle := src.GetString("cycle"); cycle != "" {
			rec.Set("cycle", cycle)
		}
		if rich {
			rec.Set("slug", src.GetString("slug"))
			rec.Set("title", src.GetString("title"))
			rec.Set("tags", src.GetString("tags"))
			rec.Set("branch", src.GetString("branch"))
			rec.Set("pr", src.GetString("pr"))
			rec.Set("external_ref", src.GetString("external_ref"))
			rec.Set("meta", src.GetString("meta"))
		}

		if err := app.Save(rec); err != nil {
			return fmt.Errorf("copy %s %s: %w", kind, src.Id, err)
		}
		migrated[src.Id] = true
		return nil
	}

	for _, src := range []struct {
		collection string
		kind       string
		rich       bool
	}{
		{ColJournal, "journal", true},
		{ColDocs, "doc", true},
		{ColTicketLogs, "log", false},
		{ColPlanLogs, "log", false},
		{ColTodoLogs, "log", false},
	} {
		rows, err := app.FindAllRecords(src.collection)
		if err != nil {
			return fmt.Errorf("load %s: %w", src.collection, err)
		}
		for _, row := range rows {
			if err := copyRow(row, src.kind, src.rich); err != nil {
				return err
			}
		}
	}

	return nil
}

func backfillTags(app core.App) error {
	tags, err := app.FindCollectionByNameOrId(ColTags)
	if err != nil {
		return err
	}

	known := make(map[string]string)
	rows, err := app.FindAllRecords(ColTags)
	if err != nil {
		return fmt.Errorf("load tags: %w", err)
	}
	for _, row := range rows {
		known[row.GetString("domain")+"/"+row.GetString("slug")] = row.Id
	}

	tagID := func(domainID, name string) (string, error) {
		slug := rules.Slugify(name)
		if slug == "" {
			return "", nil
		}
		key := domainID + "/" + slug
		if id, ok := known[key]; ok {
			return id, nil
		}
		rec := core.NewRecord(tags)
		rec.Set("domain", domainID)
		rec.Set("slug", slug)
		rec.Set("name", name)
		if err := app.Save(rec); err != nil {
			return "", fmt.Errorf("create tag %q: %w", name, err)
		}
		known[key] = rec.Id
		return rec.Id, nil
	}

	for _, src := range []struct{ collection, join, field string }{
		{ColIssues, ColIssueTags, "issue"},
		{ColEntries, ColEntryTags, "entry"},
	} {
		join, err := app.FindCollectionByNameOrId(src.join)
		if err != nil {
			return err
		}
		linked := make(map[string]bool)
		existing, err := app.FindAllRecords(src.join)
		if err != nil {
			return fmt.Errorf("load %s: %w", src.join, err)
		}
		for _, row := range existing {
			linked[row.GetString(src.field)+"/"+row.GetString("tag")] = true
		}

		records, err := app.FindAllRecords(src.collection)
		if err != nil {
			return fmt.Errorf("load %s: %w", src.collection, err)
		}
		for _, rec := range records {
			names := strSlice(rec, "tags")
			if len(names) == 0 {
				continue
			}
			domainID := rec.GetString("domain")
			if domainID == "" {
				return fmt.Errorf("%s %s has no domain", src.collection, rec.Id)
			}
			for _, name := range names {
				id, err := tagID(domainID, name)
				if err != nil {
					return err
				}
				if id == "" || linked[rec.Id+"/"+id] {
					continue
				}
				row := core.NewRecord(join)
				row.Set(src.field, rec.Id)
				row.Set("tag", id)
				if err := app.Save(row); err != nil {
					return fmt.Errorf("link tag on %s: %w", rec.Id, err)
				}
				linked[rec.Id+"/"+id] = true
			}
		}
	}

	return nil
}

// relaxSupersededColumns clears the required flag on the columns the issue and
// entry collections replaced. They are kept for the rollback window, but
// nothing writes them any more, so a required one rejects every insert.
func relaxSupersededColumns(app core.App) error {
	for _, target := range []struct{ collection, field string }{
		{ColCycles, "ticket"},
		{ColTicketLogs, "ticket"},
		{ColTodoLogs, "todo"},
		{ColPlanLogs, "plan"},
	} {
		c, err := app.FindCollectionByNameOrId(target.collection)
		if err != nil {
			return err
		}
		field, ok := c.Fields.GetByName(target.field).(*core.RelationField)
		if !ok || !field.Required {
			continue
		}
		field.Required = false
		if err := app.Save(c); err != nil {
			return fmt.Errorf("relax %s.%s: %w", target.collection, target.field, err)
		}
	}
	return nil
}

func existingSlugs(app core.App, collection string) (map[string]bool, error) {
	records, err := app.FindAllRecords(collection)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", collection, err)
	}
	taken := make(map[string]bool, len(records))
	for _, record := range records {
		if slug := record.GetString("slug"); slug != "" {
			taken[slug] = true
		}
	}
	return taken, nil
}

// find returns the collection if it already exists.
func find(app core.App, name string) (*core.Collection, bool) {
	c, err := app.FindCollectionByNameOrId(name)
	if err != nil || c == nil {
		return nil, false
	}
	return c, true
}

// alterForJournalSlug backfills a slug onto entries that predate the column.
// The field is added nullable, filled, and only then made required and
// unique: a required unique column cannot be added in one step over rows that
// all hold an empty value.
func alterForJournalSlug(app core.App) error {
	c, err := app.FindCollectionByNameOrId(ColJournal)
	if err != nil {
		return err
	}
	if c.Fields.GetByName("slug") != nil {
		return nil
	}

	c.Fields.Add(&core.TextField{Name: "slug", Max: 60, Pattern: `^[a-z0-9]+(-[a-z0-9]+)*$`})
	if err := app.Save(c); err != nil {
		return fmt.Errorf("add slug field: %w", err)
	}

	records, err := app.FindAllRecords(ColJournal)
	if err != nil {
		return fmt.Errorf("load journal: %w", err)
	}

	taken := make(map[string]map[string]bool)
	for _, record := range records {
		project := record.GetString("project")
		if taken[project] == nil {
			taken[project] = make(map[string]bool)
		}
		if slug := record.GetString("slug"); slug != "" {
			taken[project][slug] = true
		}
	}

	for _, record := range records {
		if record.GetString("slug") != "" {
			continue
		}
		project := record.GetString("project")
		slug := uniqueSlug(rules.Slugify(record.GetString("title")), taken[project])

		taken[project][slug] = true
		record.Set("slug", slug)
		if err := app.Save(record); err != nil {
			return fmt.Errorf("backfill slug for %s: %w", record.Id, err)
		}
	}

	c, err = app.FindCollectionByNameOrId(ColJournal)
	if err != nil {
		return err
	}
	if field, ok := c.Fields.GetByName("slug").(*core.TextField); ok {
		field.Required = true
	}
	c.AddIndex("idx_journ_journal_slug", true, "project, slug", "")

	return app.Save(c)
}

// uniqueSlug takes the lowest free numeric suffix. A title that slugifies to
// nothing (emoji, CJK) falls back to "entry".
func uniqueSlug(base string, taken map[string]bool) string {
	if base == "" {
		base = "entry"
	}
	// Leaves room for a suffix inside the 60-character column.
	if len(base) > 50 {
		base = strings.TrimRight(base[:50], "-")
	}
	if !taken[base] {
		return base
	}
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if !taken[candidate] {
			return candidate
		}
	}
}
