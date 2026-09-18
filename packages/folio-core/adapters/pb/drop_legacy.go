package pb

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// The collections issues and entries replaced. They are spelled out rather
// than referenced as constants: nothing else in the schema knows these names
// any more, and this file is the last thing that should.
const (
	legacyTickets    = "journ_tickets"
	legacyTodos      = "journ_todos"
	legacyJournal    = "journ_journal"
	legacyDocs       = "journ_docs"
	legacyTicketLogs = "journ_ticket_logs"
	legacyPlanLogs   = "journ_plan_logs"
	legacyTodoLogs   = "journ_todo_logs"
)

// dropLegacy removes the superseded collections from a database that still
// carries them. A database created after they were removed from the schema has
// none of them, so this is a no-op there.
//
// The backfills that copied these rows into issues and entries are gone, so
// this cannot assume the copy happened. Every source row is checked for its
// counterpart first, and a single row without one aborts the drop: the row
// would be unrecoverable afterwards, and a partial schema is easier to fix
// than lost data.
func dropLegacy(app core.App) error {
	if err := verifyCopied(app); err != nil {
		return err
	}

	// Relations pointing at these collections go first: PocketBase refuses to
	// delete a collection another one still references.
	if err := dropSupersededColumns(app); err != nil {
		return err
	}

	// Children before parents, so no drop is blocked by a live reference.
	for _, name := range []string{
		legacyTicketLogs, legacyPlanLogs, legacyTodoLogs,
		legacyJournal, legacyDocs, legacyTodos, legacyTickets,
	} {
		c, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			continue
		}
		if err := app.Delete(c); err != nil {
			return fmt.Errorf("drop %s: %w", name, err)
		}
	}

	return nil
}

// verifyCopied refuses the drop unless every legacy row already exists in the
// collection that replaced it. The backfill copied each row under its own id,
// so the id is what proves a counterpart.
func verifyCopied(app core.App) error {
	for _, target := range []struct{ source, into string }{
		{legacyTickets, ColIssues},
		{legacyTodos, ColIssues},
		{legacyJournal, ColEntries},
		{legacyDocs, ColEntries},
		{legacyTicketLogs, ColEntries},
		{legacyPlanLogs, ColEntries},
		{legacyTodoLogs, ColEntries},
	} {
		// A collection that is already gone has nothing left to verify.
		if _, ok := find(app, target.source); !ok {
			continue
		}

		rows, err := app.FindAllRecords(target.source)
		if err != nil {
			return fmt.Errorf("read %s: %w", target.source, err)
		}
		for _, row := range rows {
			if _, err := app.FindRecordById(target.into, row.Id); err != nil {
				return fmt.Errorf(
					"%s %s has no counterpart in %s: migrate it before the drop",
					target.source, row.Id, target.into,
				)
			}
		}
	}

	return nil
}

// dropSupersededColumns removes the relations that pointed at the retired
// collections. Each was replaced by one pointing at issues, and the repositories
// stopped reading them, but PocketBase will not delete a collection while a
// relation still targets it.
func dropSupersededColumns(app core.App) error {
	for _, target := range []struct{ collection, field string }{
		{ColPlans, "ticket"},
		{ColCycles, "ticket"},
		{ColEntries, "issue_todo"},
		{ColEntries, "todo"},
	} {
		c, err := app.FindCollectionByNameOrId(target.collection)
		if err != nil {
			continue
		}
		if c.Fields.GetByName(target.field) == nil {
			continue
		}
		c.Fields.RemoveByName(target.field)

		// An index over the dropped column would fail to rebuild.
		column := regexp.MustCompile(`\b` + regexp.QuoteMeta(target.field) + `\b`)
		kept := make([]string, 0, len(c.Indexes))
		for _, idx := range c.Indexes {
			if open := strings.LastIndex(idx, "("); open < 0 || !column.MatchString(idx[open:]) {
				kept = append(kept, idx)
			}
		}
		c.Indexes = kept

		if err := app.Save(c); err != nil {
			return fmt.Errorf("drop %s.%s: %w", target.collection, target.field, err)
		}
	}

	return nil
}
