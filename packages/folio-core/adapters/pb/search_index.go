package pb

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/ports"
)

// SearchIndex is the FTS5 table every searchable kind is mirrored into.
const SearchIndex = "journ_search"

// indexColumns is the table body, kept apart from the CREATE so the same text
// can be compared against what SQLite stored. SQLite drops `IF NOT EXISTS`
// when it records the statement and keeps the rest verbatim, so an exact
// compare tells us whether the live index still matches this definition.
//
// The tokenizer is porter wrapping unicode61, which stems English so "hosting"
// finds "hosted". `tokenchars '_-'` keeps snake_case and kebab-case
// identifiers as single tokens; without it `x-forwarded-for` becomes a
// three-token phrase query and pays for position checks on every candidate.
// A dot is deliberately not a token char: rules.FTSQuery splits on it, and the
// two have to agree or dotted terms stop matching.
const indexColumns = `(
	source UNINDEXED,
	kind UNINDEXED,
	rec_id UNINDEXED,
	slug UNINDEXED,
	project UNINDEXED,
	tags UNINDEXED,
	created UNINDEXED,
	title,
	body,
	tokenize='porter unicode61 remove_diacritics 2 tokenchars ''_-'''
)`

// indexDefinition is what SQLite stores for the index as it exists now.
func indexDefinition() string {
	return "CREATE VIRTUAL TABLE " + SearchIndex + " USING fts5" + indexColumns
}

// indexed describes one searchable collection. Every field but Collection and
// Where is a SQL expression in which `{a}` stands for the row being read,
// which is `new` inside a trigger and the table alias during a backfill.
//
// A collection maps to several search kinds when it discriminates by a column,
// so Kind is an expression rather than a constant: journ_entries carries
// journal, doc and log rows, and journ_issues carries tickets and todos.
type indexed struct {
	Collection string
	Kind       string
	Title      string
	Body       string
	Slug       string
	Tags       string
	// Where is an extra predicate, in the same `{a}` form, that a row must
	// satisfy to be indexed at all.
	Where string
	// TagTarget names which join table carries this collection's tags, when
	// one does. Tags are written after the record itself and never touch its
	// row, so the record's own triggers cannot see them: the join table needs
	// triggers of its own that rebuild the index row.
	TagTarget ports.TagTarget
}

// tagJoinOf reuses the repository's mapping so the index and the writes agree
// on which table holds a target's tags.
func (s indexed) tagJoinOf() (collection, ref string) {
	return tagJoin(s.TagTarget)
}

// tagsFrom aggregates a record's tag slugs as the JSON array the index stores,
// which is the shape the tag filter's LIKE expects.
func tagsFrom(collection, ref string) string {
	return fmt.Sprintf(
		`(SELECT COALESCE(json_group_array(t.slug), '[]') FROM %s j JOIN %s t ON t.id = j.tag WHERE j.%s = {a}.id)`,
		collection, ColTags, ref,
	)
}

// tagsExpr is the SQL for a row's tags, from its join table when it has one.
func (s indexed) tagsExpr() string {
	if s.TagTarget != "" {
		return tagsFrom(s.tagJoinOf())
	}
	if s.Tags != "" {
		return s.Tags
	}
	return noTags
}

// expand resolves the `{a}` alias in a fragment.
func expand(fragment, alias string) string {
	return strings.ReplaceAll(fragment, "{a}", alias)
}

const (
	noSlug = "''"
	noTags = "'[]'"
)

// sources is the whole searchable surface. Adding a kind here gives it an
// index, triggers and a backfill; nothing else needs to change.
var sources = []indexed{
	{
		Collection: ColEntries,
		// A log has no title of its own, so the index supplies one; a journal
		// entry or doc that was saved without one falls back the same way.
		Kind:      `CASE {a}.kind WHEN 'log' THEN 'worklog' ELSE {a}.kind END`,
		Title:     `CASE WHEN {a}.title = '' THEN 'Work log' ELSE {a}.title END`,
		Body:      `{a}.body`,
		Slug:      `{a}.slug`,
		TagTarget: ports.TagEntry,
	},
	{
		Collection: ColIssues,
		Kind:       `{a}.kind`,
		Title:      `{a}.title`,
		Body:       `{a}.body`,
		Slug:       `{a}.slug`,
		TagTarget:  ports.TagIssue,
	},
	{
		Collection: ColPlans,
		Kind:       `'plan'`,
		Title:      `{a}.title`,
		Body:       `{a}.goal`,
		Slug:       noSlug,
		// A plan has a tags column, but nothing writes it and there is no
		// join table, so a plan carries no tags rather than a stale array.
		Tags: noTags,
	},
	{
		Collection: ColCycles,
		Kind:       `'resolution'`,
		// The ordinal is what names a cycle, and the index is the only place
		// that can read it at write time, so it is baked into the title.
		Title: `'Cycle ' || {a}.ordinal || ' resolution'`,
		Body:  `{a}.resolution`,
		Slug:  noSlug,
		Tags:  noTags,
		// A cycle is only searchable once it has closed with a resolution.
		Where: `{a}.resolution != ''`,
	},
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

	// An index built by an earlier version has different columns or a
	// different tokenizer, and FTS5 cannot alter one in place, so it is
	// dropped and rebuilt from the sources.
	var existing string
	if err := db.NewQuery(`SELECT COALESCE(max(sql), '') FROM sqlite_master WHERE name = {:name}`).
		Bind(dbx.Params{"name": SearchIndex}).Row(&existing); err != nil {
		return fmt.Errorf("inspect %s: %w", SearchIndex, err)
	}
	if existing != "" && existing != indexDefinition() {
		if _, err := db.NewQuery(`DROP TABLE ` + SearchIndex).Execute(); err != nil {
			return fmt.Errorf("drop stale %s: %w", SearchIndex, err)
		}
	}

	// Contentless-delete keeps the index self-sufficient: the delete trigger
	// can remove a row by its own columns without reading the source table.
	create := "CREATE VIRTUAL TABLE IF NOT EXISTS " + SearchIndex + " USING fts5" + indexColumns
	if _, err := db.NewQuery(create).Execute(); err != nil {
		return fmt.Errorf("create %s: %w", SearchIndex, err)
	}

	for _, source := range sources {
		if err := ensureTriggers(app, source); err != nil {
			return fmt.Errorf("%s triggers: %w", source.Collection, err)
		}
		if err := ensureTagTriggers(app, source); err != nil {
			return fmt.Errorf("%s tag triggers: %w", source.Collection, err)
		}
	}

	return backfillSearchIndex(app)
}

// indexColumnList is the insert target, shared by the triggers and the
// backfill so the two cannot drift apart.
const indexColumnList = "source, kind, rec_id, slug, project, tags, created, title, body"

// selectFor builds the SELECT that produces one index row from a source row.
func (s indexed) selectFor(alias string) string {
	return fmt.Sprintf(`SELECT '%s', %s, %s.id, %s, %s.project, %s, %s.created, %s, %s`,
		s.Collection,
		expand(s.Kind, alias),
		alias,
		expand(s.Slug, alias),
		alias,
		expand(s.tagsExpr(), alias),
		alias,
		expand(s.Title, alias),
		expand(s.Body, alias),
	)
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
		condition = expand(s.Where, "new")
	}

	insert := fmt.Sprintf(`INSERT INTO %s(%s) %s WHERE %s;`,
		SearchIndex, indexColumnList, s.selectFor("new"), condition)

	// Deleting by source and id rather than by kind: a collection that
	// discriminates by a column can change a row's kind on update, and the
	// delete has to remove the old row whatever kind it was filed under.
	remove := fmt.Sprintf(`DELETE FROM %s WHERE source = '%s' AND rec_id = old.id;`,
		SearchIndex, s.Collection)

	statements := []string{
		fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_ai`, prefix),
		fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_ad`, prefix),
		fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_au`, prefix),

		fmt.Sprintf(`CREATE TRIGGER %s_ai AFTER INSERT ON %s BEGIN %s END`, prefix, s.Collection, insert),
		fmt.Sprintf(`CREATE TRIGGER %s_ad AFTER DELETE ON %s BEGIN %s END`, prefix, s.Collection, remove),
		fmt.Sprintf(`CREATE TRIGGER %s_au AFTER UPDATE ON %s BEGIN %s %s END`, prefix, s.Collection, remove, insert),
	}

	for _, sql := range statements {
		if _, err := db.NewQuery(sql).Execute(); err != nil {
			return fmt.Errorf("%s: %w", sql, err)
		}
	}

	return nil
}

// ensureTagTriggers rebuilds a record's index row when its tags change.
//
// Tagging writes to a join table, not to the record, so nothing about the
// record's own row changes and its update trigger never fires. These triggers
// watch the join table instead and re-read the record, which is why the insert
// they run selects from the source collection rather than from `new`.
func ensureTagTriggers(app core.App, s indexed) error {
	if s.TagTarget == "" {
		return nil
	}

	db := app.DB()
	joinTable, ref := s.tagJoinOf()
	prefix := "fts_" + joinTable

	// The row is rebuilt for whichever record the join row points at, which is
	// new.<ref> on insert and old.<ref> on delete.
	rebuild := func(alias string) string {
		record := alias + "." + ref

		remove := fmt.Sprintf(`DELETE FROM %s WHERE source = '%s' AND rec_id = %s;`,
			SearchIndex, s.Collection, record)

		insert := fmt.Sprintf(`INSERT INTO %s(%s) %s FROM %s src WHERE src.id = %s`,
			SearchIndex, indexColumnList, s.selectFor("src"), s.Collection, record)
		if s.Where != "" {
			insert += " AND " + expand(s.Where, "src")
		}

		return remove + insert + ";"
	}

	statements := []string{
		fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_ai`, prefix),
		fmt.Sprintf(`DROP TRIGGER IF EXISTS %s_ad`, prefix),

		fmt.Sprintf(`CREATE TRIGGER %s_ai AFTER INSERT ON %s BEGIN %s END`,
			prefix, joinTable, rebuild("new")),
		fmt.Sprintf(`CREATE TRIGGER %s_ad AFTER DELETE ON %s BEGIN %s END`,
			prefix, joinTable, rebuild("old")),
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
	if err := db.NewQuery(`SELECT count(*) FROM ` + SearchIndex).Row(&indexed); err != nil {
		return fmt.Errorf("count index: %w", err)
	}

	var stored int
	for _, source := range sources {
		var n int
		query := `SELECT count(*) FROM ` + source.Collection
		if source.Where != "" {
			query += " WHERE " + expand(source.Where, source.Collection)
		}
		if err := db.NewQuery(query).Row(&n); err != nil {
			return fmt.Errorf("count %s: %w", source.Collection, err)
		}
		stored += n
	}

	if indexed == stored {
		return nil
	}

	if _, err := db.NewQuery(`DELETE FROM ` + SearchIndex).Execute(); err != nil {
		return fmt.Errorf("clear index: %w", err)
	}

	for _, source := range sources {
		fill := fmt.Sprintf(`INSERT INTO %s(%s) %s FROM %s s`,
			SearchIndex, indexColumnList, source.selectFor("s"), source.Collection)
		if source.Where != "" {
			fill += " WHERE " + expand(source.Where, "s")
		}
		if _, err := db.NewQuery(fill).Execute(); err != nil {
			return fmt.Errorf("backfill %s: %w", source.Collection, err)
		}
	}

	return nil
}
