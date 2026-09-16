package pb

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
)

// SearchIndex is the FTS5 table every searchable kind is mirrored into.
const SearchIndex = "journ_search"

// indexed describes one searchable collection: where its text lives and which
// kind a hit from it reports. Title and Body are column names in the source
// table; a source with no title of its own leaves Title empty.
type indexed struct {
	Collection string
	Kind       domain.SearchKind
	Title      string
	Body       string
	// Fixed title for sources that have none, such as work logs.
	StaticTitle string
	// Extra SQL predicate a row must satisfy to be indexed at all.
	Where string
	// Set when the collection addresses records by slug in its routes.
	HasSlug bool
}

func (s indexed) slugExpr(alias string) string {
	if s.HasSlug {
		return alias + ".slug"
	}
	return "''"
}

// sources is the whole searchable surface. Adding a kind here gives it an
// index, triggers and a backfill; nothing else needs to change.
var sources = []indexed{
	{Collection: ColJournal, Kind: domain.SearchKindJournal, Title: "title", Body: "body", HasSlug: true},
	{Collection: ColDocs, Kind: domain.SearchKindDoc, Title: "title", Body: "body", HasSlug: true},
	{Collection: ColTodos, Kind: domain.SearchKindTodo, Title: "title", Body: "details"},
	{Collection: ColPlans, Kind: domain.SearchKindPlan, Title: "title", Body: "goal"},
	{Collection: ColTickets, Kind: domain.SearchKindTicket, Title: "title", Body: "body", HasSlug: true},

	{Collection: ColTicketLogs, Kind: domain.SearchKindWorkLog, Body: "body", StaticTitle: "Ticket work log"},
	{Collection: ColPlanLogs, Kind: domain.SearchKindWorkLog, Body: "body", StaticTitle: "Plan work log"},
	{Collection: ColTodoLogs, Kind: domain.SearchKindWorkLog, Body: "body", StaticTitle: "Todo work log"},

	// A cycle is only searchable once it has closed with a resolution.
	{
		Collection:  ColCycles,
		Kind:        domain.SearchKindCycle,
		Body:        "resolution",
		StaticTitle: "resolution",
		Where:       "resolution != ''",
	},
}

// titleExpr is the SQL that produces a row's indexed title, which is either a
// column or a constant.
func (s indexed) titleExpr(alias string) string {
	if s.Title != "" {
		return alias + "." + s.Title
	}
	return "'" + s.StaticTitle + "'"
}

// tagsExpr is the SQL for a row's tags. Work logs and cycles carry none, and
// an empty JSON array keeps the indexed shape uniform.
func (s indexed) tagsExpr(alias string) string {
	switch s.Collection {
	case ColTicketLogs, ColPlanLogs, ColTodoLogs, ColCycles:
		return "'[]'"
	default:
		return "COALESCE(" + alias + ".tags, '[]')"
	}
}

// ensureSearchIndex creates the FTS5 table, the triggers that keep it current
// and a backfill for rows that predate it.
//
// The index is keyed by the record's own id rather than rowid. PocketBase
// tables declare `id TEXT PRIMARY KEY`, so rowid is implicit, and SQLite only
// promises rowid stability across VACUUM for INTEGER PRIMARY KEY tables. An
// external-content table keyed on rowid could therefore survive a vacuum
// pointing at the wrong records.
func ensureSearchIndex(app core.App) error {
	db := app.DB()

	// Contentless-delete keeps the index self-sufficient: the delete trigger
	// can remove a row by its own columns without reading the source table.
	create := fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS %s USING fts5(
			kind UNINDEXED,
			rec_id UNINDEXED,
			slug UNINDEXED,
			project UNINDEXED,
			tags UNINDEXED,
			created UNINDEXED,
			title,
			body,
			tokenize='porter unicode61'
		)`, SearchIndex)

	// An index built by an earlier version has fewer columns, and FTS5 cannot
	// alter one in place, so it is dropped and rebuilt from the sources.
	var existing string
	if err := db.NewQuery(`SELECT COALESCE(max(sql), '') FROM sqlite_master WHERE name = {:name}`).
		Bind(dbx.Params{"name": SearchIndex}).Row(&existing); err != nil {
		return fmt.Errorf("inspect %s: %w", SearchIndex, err)
	}
	if existing != "" && !strings.Contains(existing, "slug") {
		if _, err := db.NewQuery(fmt.Sprintf(`DROP TABLE %s`, SearchIndex)).Execute(); err != nil {
			return fmt.Errorf("drop stale %s: %w", SearchIndex, err)
		}
	}

	if _, err := db.NewQuery(create).Execute(); err != nil {
		return fmt.Errorf("create %s: %w", SearchIndex, err)
	}

	for _, source := range sources {
		if err := ensureTriggers(app, source); err != nil {
			return fmt.Errorf("%s triggers: %w", source.Collection, err)
		}
	}

	return backfillSearchIndex(app)
}

// ensureTriggers mirrors one collection into the index on insert, update and
// delete. An update deletes and reinserts rather than patching, since FTS5
// cannot update a contentless row in place.
func ensureTriggers(app core.App, s indexed) error {
	db := app.DB()
	prefix := "fts_" + s.Collection

	// The insert is written as a SELECT ... WHERE so the same statement serves
	// both triggers: on update a row that stopped qualifying simply inserts
	// nothing, and the preceding delete has already removed it.
	condition := "1"
	if s.Where != "" {
		condition = strings.ReplaceAll(s.Where, s.Body, "new."+s.Body)
	}

	insert := fmt.Sprintf(
		`INSERT INTO %s(kind, rec_id, slug, project, tags, created, title, body) SELECT '%s', new.id, %s, new.project, %s, new.created, %s, new.%s WHERE %s;`,
		SearchIndex, s.Kind, s.slugExpr("new"), s.tagsExpr("new"), s.titleExpr("new"), s.Body, condition,
	)
	remove := fmt.Sprintf(`DELETE FROM %s WHERE kind = '%s' AND rec_id = old.id;`, SearchIndex, s.Kind)

	statements := []string{
		fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_ai`, prefix),
		fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_ad`, prefix),
		fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_au`, prefix),

		fmt.Sprintf(`CREATE TRIGGER %s_ai AFTER INSERT ON %s BEGIN %s END`,
			prefix, s.Collection, insert),

		fmt.Sprintf(`CREATE TRIGGER %s_ad AFTER DELETE ON %s BEGIN %s END`,
			prefix, s.Collection, remove),

		fmt.Sprintf(`CREATE TRIGGER %s_au AFTER UPDATE ON %s BEGIN %s %s END`,
			prefix, s.Collection, remove, insert),
	}

	for _, sql := range statements {
		if _, err := db.NewQuery(sql).Execute(); err != nil {
			return fmt.Errorf("%s: %w", sql, err)
		}
	}

	return nil
}

// backfillSearchIndex fills the index from rows that existed before it did.
// It rebuilds from empty rather than diffing, which keeps it correct if a
// trigger was ever missing, and is cheap at the size these projects reach.
func backfillSearchIndex(app core.App) error {
	db := app.DB()

	var indexed int
	if err := db.NewQuery(fmt.Sprintf(`SELECT count(*) FROM %s`, SearchIndex)).Row(&indexed); err != nil {
		return fmt.Errorf("count index: %w", err)
	}

	var stored int
	for _, source := range sources {
		var n int
		query := fmt.Sprintf(`SELECT count(*) FROM %s`, source.Collection)
		if source.Where != "" {
			query += " WHERE " + source.Where
		}
		if err := db.NewQuery(query).Row(&n); err != nil {
			return fmt.Errorf("count %s: %w", source.Collection, err)
		}
		stored += n
	}

	if indexed == stored {
		return nil
	}

	if _, err := db.NewQuery(fmt.Sprintf(`DELETE FROM %s`, SearchIndex)).Execute(); err != nil {
		return fmt.Errorf("clear index: %w", err)
	}

	for _, source := range sources {
		fill := fmt.Sprintf(
			`INSERT INTO %s(kind, rec_id, slug, project, tags, created, title, body) SELECT '%s', s.id, %s, s.project, %s, s.created, %s, s.%s FROM %s s`,
			SearchIndex, source.Kind, source.slugExpr("s"), source.tagsExpr("s"), source.titleExpr("s"), source.Body, source.Collection,
		)
		if source.Where != "" {
			fill += " WHERE s." + source.Where
		}
		if _, err := db.NewQuery(fill).Execute(); err != nil {
			return fmt.Errorf("backfill %s: %w", source.Collection, err)
		}
	}

	return nil
}
