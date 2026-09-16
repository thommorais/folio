package pb

import (
	"fmt"
	"strings"

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
	if err := ensureCycles(app); err != nil {
		return fmt.Errorf("cycles: %w", err)
	}
	if err := ensureWorkLogs(app); err != nil {
		return fmt.Errorf("work logs: %w", err)
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
