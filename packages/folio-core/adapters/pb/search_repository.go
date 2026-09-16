package pb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

// SearchRepository answers a query from the FTS5 index rather than the source
// collections, so one statement covers every kind and the rows come back
// ranked by relevance instead of recency.
type SearchRepository struct {
	app core.App
}

func NewSearchRepository(app core.App) *SearchRepository {
	return &SearchRepository{app: app}
}

var _ ports.SearchRepository = (*SearchRepository)(nil)

// snippetLen is how much surrounding text a hit carries: enough for an agent
// to judge relevance without fetching the record.
const snippetLen = 200

// Column weights for bm25. Title is worth more than body because a query that
// names a record is almost always looking for that record, and the columns
// before title are UNINDEXED so they take a zero.
const bm25Weights = "0.0, 0.0, 0.0, 0.0, 0.0, 10.0, 1.0"

type searchRow struct {
	Kind    string `db:"kind"`
	RecID   string `db:"rec_id"`
	Project string `db:"project"`
	Tags    string `db:"tags"`
	Created string `db:"created"`
	Title   string `db:"title"`
	Body    string `db:"body"`
}

func (r *SearchRepository) Search(ctx context.Context, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	where := []string{"project = {:project}"}
	params := dbx.Params{"project": string(project)}

	if match := rules.FTSQuery(q.Text); match != "" {
		where = append(where, SearchIndex+" MATCH {:match}")
		params["match"] = match
	}

	if len(q.Kinds) > 0 {
		placeholders := make([]string, 0, len(q.Kinds))
		for i, kind := range q.Kinds {
			key := "kind" + strconv.Itoa(i)
			placeholders = append(placeholders, "{:"+key+"}")
			params[key] = string(kind)
		}
		where = append(where, "kind IN ("+strings.Join(placeholders, ", ")+")")
	}

	// Tags are stored as a JSON array, so a containment test is a LIKE over
	// the quoted value. Every kind is filtered the same way; kinds that carry
	// no tags simply stop matching, which is what a tag filter should mean.
	for i, tag := range q.Tags {
		key := "tag" + strconv.Itoa(i)
		where = append(where, "tags LIKE {:"+key+"}")
		params[key] = `%"` + tag + `"%`
	}

	// Ranking only means something once a MATCH has scored the rows. Without
	// one, newest first is the honest order and matches the old behaviour.
	order := "created DESC"
	if _, matched := params["match"]; matched {
		order = fmt.Sprintf("bm25(%s, %s) ASC, created DESC", SearchIndex, bm25Weights)
	}

	query := fmt.Sprintf(
		`SELECT kind, rec_id, project, tags, created, title, body FROM %s WHERE %s ORDER BY %s LIMIT {:limit} OFFSET {:offset}`,
		SearchIndex, strings.Join(where, " AND "), order,
	)

	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	params["limit"] = limit
	params["offset"] = max(q.Offset, 0)

	var rows []searchRow
	if err := r.app.DB().NewQuery(query).Bind(params).WithContext(ctx).All(&rows); err != nil {
		return nil, mapErr(err)
	}

	hits := make([]domain.SearchHit, 0, len(rows))
	for _, row := range rows {
		hits = append(hits, domain.SearchHit{
			Kind:      domain.SearchKind(row.Kind),
			ID:        row.RecID,
			ProjectID: domain.ProjectID(row.Project),
			Title:     title(row),
			Snippet:   rules.Snippet(row.Body, snippetLen),
			Tags:      parseTags(row.Tags),
			CreatedAt: parseCreated(row.Created),
		})
	}

	return hits, nil
}

// title names a cycle hit by its ordinal, which the index cannot know because
// the resolution is all it stores.
func title(row searchRow) string {
	if domain.SearchKind(row.Kind) == domain.SearchKindCycle {
		return "Cycle resolution"
	}
	return row.Title
}

// parseTags reads the JSON array PocketBase stores tags in. A malformed or
// absent value yields no tags rather than an error: a hit is still useful
// without them.
func parseTags(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "[]" || trimmed == "null" {
		return nil
	}

	var tags []string
	if err := json.Unmarshal([]byte(trimmed), &tags); err != nil {
		return nil
	}
	return tags
}

// parseCreated reads PocketBase's stored timestamp. The index keeps it as the
// raw string, so the layout has to match what PocketBase writes.
func parseCreated(raw string) time.Time {
	for _, layout := range []string{"2006-01-02 15:04:05.000Z", "2006-01-02 15:04:05Z", time.RFC3339} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t
		}
	}
	return time.Time{}
}
