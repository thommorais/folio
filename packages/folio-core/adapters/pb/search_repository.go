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
	Kind        string `db:"kind"`
	RecID       string `db:"rec_id"`
	Slug        string `db:"slug"`
	Project     string `db:"project"`
	ProjectSlug string `db:"project_slug"`
	Tags        string `db:"tags"`
	Created     string `db:"created"`
	Title       string `db:"title"`
	Body        string `db:"body"`
}

func (r *SearchRepository) Search(ctx context.Context, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	return r.search(ctx, []domain.ProjectID{project}, q)
}

// SearchAcross ranks one result set spanning several projects. The index holds
// every project already, so this is the same single query with a wider scope
// rather than a fan-out whose scores could not be compared.
func (r *SearchRepository) SearchAcross(ctx context.Context, projects []domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	if len(projects) == 0 {
		return []domain.SearchHit{}, nil
	}
	return r.search(ctx, projects, q)
}

func (r *SearchRepository) search(ctx context.Context, projects []domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	placeholders := make([]string, 0, len(projects))
	params := dbx.Params{}
	for i, id := range projects {
		key := "project" + strconv.Itoa(i)
		placeholders = append(placeholders, "{:"+key+"}")
		params[key] = string(id)
	}
	where := []string{SearchIndex + ".project IN (" + strings.Join(placeholders, ", ") + ")"}

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
		where = append(where, SearchIndex+".kind IN ("+strings.Join(placeholders, ", ")+")")
	}

	// Tags are stored as a JSON array, so a containment test is a LIKE over
	// the quoted value. Every kind is filtered the same way; kinds that carry
	// no tags simply stop matching, which is what a tag filter should mean.
	for i, tag := range q.Tags {
		key := "tag" + strconv.Itoa(i)
		where = append(where, SearchIndex+".tags LIKE {:"+key+"}")
		params[key] = `%"` + tag + `"%`
	}

	// Ranking only means something once a MATCH has scored the rows. Without
	// one, newest first is the honest order and matches the old behaviour.
	order := SearchIndex + ".created DESC"
	if _, matched := params["match"]; matched {
		order = fmt.Sprintf("bm25(%s, %s) ASC, %s.created DESC", SearchIndex, bm25Weights, SearchIndex)
	}

	// The index stores the project id; the slug is what names a hit in a
	// global result, so it is joined in rather than resolved per row later.
	query := fmt.Sprintf(
		`SELECT %[1]s.kind, %[1]s.rec_id, %[1]s.slug, %[1]s.project, COALESCE(p.slug, '') AS project_slug,
		        %[1]s.tags, %[1]s.created, %[1]s.title, %[1]s.body
		 FROM %[1]s LEFT JOIN %[2]s p ON p.id = %[1]s.project
		 WHERE %[3]s ORDER BY %[4]s LIMIT {:limit} OFFSET {:offset}`,
		SearchIndex, ColProjects, strings.Join(where, " AND "), order,
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
			Kind:        domain.SearchKind(row.Kind),
			ID:          row.RecID,
			ProjectID:   domain.ProjectID(row.Project),
			ProjectSlug: row.ProjectSlug,
			Slug:        row.Slug,
			Title:       title(row),
			Snippet:     rules.Snippet(row.Body, snippetLen),
			Tags:        parseTags(row.Tags),
			CreatedAt:   parseCreated(row.Created),
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
