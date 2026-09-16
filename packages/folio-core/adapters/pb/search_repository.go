package pb

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

// SearchRepository runs one query per requested kind and merges the results
// newest first. Each query is bounded by the caller's limit, so the merge
// never holds more than kinds x limit rows.
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

func (r *SearchRepository) Search(ctx context.Context, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	want := map[domain.SearchKind]bool{}
	for _, k := range q.Kinds {
		want[k] = true
	}
	all := len(want) == 0

	var hits []domain.SearchHit
	if all || want[domain.SearchKindJournal] {
		found, err := r.searchJournal(project, q)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}
	if all || want[domain.SearchKindDoc] {
		found, err := r.searchDocs(project, q)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}
	if all || want[domain.SearchKindTodo] {
		found, err := r.searchTodos(project, q)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}
	if all || want[domain.SearchKindPlan] {
		found, err := r.searchPlans(project, q)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}

	if all || want[domain.SearchKindTicket] {
		found, err := r.searchTickets(project, q)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}

	if all || want[domain.SearchKindWorkLog] {
		found, err := r.searchWorkLogs(project, q)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}

	if all || want[domain.SearchKindCycle] {
		found, err := r.searchResolutions(project, q)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}

	sort.SliceStable(hits, func(i, j int) bool { return hits[i].CreatedAt.After(hits[j].CreatedAt) })
	return applyPaging(hits, q.Offset, q.Limit), nil
}

// textFilter builds the shared "project + text + tags" filter for a
// collection, given which fields hold its searchable text.
func textFilter(project domain.ProjectID, q domain.SearchQuery, fields ...string) (string, dbx.Params) {
	filter := []string{"project = {:project}"}
	params := dbx.Params{"project": string(project)}

	if text := strings.TrimSpace(q.Text); text != "" {
		ors := make([]string, 0, len(fields))
		for _, f := range fields {
			ors = append(ors, f+" ~ {:text}")
		}
		filter = append(filter, "("+strings.Join(ors, " || ")+")")
		params["text"] = text
	}
	for i, tag := range q.Tags {
		key := "tag" + strconv.Itoa(i)
		filter = append(filter, "tags ~ {:"+key+"}")
		params[key] = `"` + tag + `"`
	}
	return strings.Join(filter, " && "), params
}

func (r *SearchRepository) searchJournal(project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	filter, params := textFilter(project, q, "title", "body")
	records, err := r.app.FindRecordsByFilter(ColJournal, filter, "-created", q.Limit, 0, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.SearchHit, 0, len(records))
	for _, rec := range records {
		out = append(out, domain.SearchHit{
			Kind:      domain.SearchKindJournal,
			ID:        rec.Id,
			ProjectID: project,
			Title:     rec.GetString("title"),
			Snippet:   rules.Snippet(rec.GetString("body"), snippetLen),
			Tags:      strSlice(rec, "tags"),
			CreatedAt: rec.GetDateTime("created").Time(),
		})
	}
	return out, nil
}

func (r *SearchRepository) searchDocs(project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	filter, params := textFilter(project, q, "title", "body")
	records, err := r.app.FindRecordsByFilter(ColDocs, filter, "-updated", q.Limit, 0, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.SearchHit, 0, len(records))
	for _, rec := range records {
		out = append(out, domain.SearchHit{
			Kind:      domain.SearchKindDoc,
			ID:        rec.Id,
			ProjectID: project,
			Title:     rec.GetString("title"),
			Snippet:   rules.Snippet(rec.GetString("body"), snippetLen),
			Tags:      strSlice(rec, "tags"),
			CreatedAt: rec.GetDateTime("created").Time(),
		})
	}
	return out, nil
}

func (r *SearchRepository) searchTodos(project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	filter, params := textFilter(project, q, "title", "details")
	records, err := r.app.FindRecordsByFilter(ColTodos, filter, "-created", q.Limit, 0, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.SearchHit, 0, len(records))
	for _, rec := range records {
		out = append(out, domain.SearchHit{
			Kind:      domain.SearchKindTodo,
			ID:        rec.Id,
			ProjectID: project,
			Title:     rec.GetString("title"),
			Snippet:   rules.Snippet(rec.GetString("details"), snippetLen),
			Tags:      strSlice(rec, "tags"),
			CreatedAt: rec.GetDateTime("created").Time(),
		})
	}
	return out, nil
}

func (r *SearchRepository) searchPlans(project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	filter, params := textFilter(project, q, "title", "goal")
	records, err := r.app.FindRecordsByFilter(ColPlans, filter, "-created", q.Limit, 0, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.SearchHit, 0, len(records))
	for _, rec := range records {
		out = append(out, domain.SearchHit{
			Kind:      domain.SearchKindPlan,
			ID:        rec.Id,
			ProjectID: project,
			Title:     rec.GetString("title"),
			Snippet:   rules.Snippet(rec.GetString("goal"), snippetLen),
			Tags:      strSlice(rec, "tags"),
			CreatedAt: rec.GetDateTime("created").Time(),
		})
	}
	return out, nil
}

func (r *SearchRepository) searchTickets(project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	filter, params := textFilter(project, q, "title", "body")
	records, err := r.app.FindRecordsByFilter(ColTickets, filter, "-created", q.Limit, 0, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.SearchHit, 0, len(records))
	for _, rec := range records {
		out = append(out, domain.SearchHit{
			Kind:      domain.SearchKindTicket,
			ID:        rec.Id,
			ProjectID: project,
			Title:     rec.GetString("title"),
			Snippet:   rules.Snippet(rec.GetString("body"), snippetLen),
			Tags:      strSlice(rec, "tags"),
			CreatedAt: rec.GetDateTime("created").Time(),
		})
	}
	return out, nil
}

// searchWorkLogs spans the three work log collections. They carry no title of
// their own, so the hit is titled by the parent kind and the body does the
// work of both title and snippet.
func (r *SearchRepository) searchWorkLogs(project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	if len(q.Tags) > 0 {
		return nil, nil
	}
	sources := []struct {
		collection string
		title      string
	}{
		{ColTicketLogs, "Ticket work log"},
		{ColPlanLogs, "Plan work log"},
		{ColTodoLogs, "Todo work log"},
	}

	var out []domain.SearchHit
	for _, source := range sources {
		filter, params := textFilter(project, q, "body")
		records, err := r.app.FindRecordsByFilter(source.collection, filter, "-created", q.Limit, 0, params)
		if err != nil {
			return nil, mapErr(err)
		}
		for _, rec := range records {
			out = append(out, domain.SearchHit{
				Kind:      domain.SearchKindWorkLog,
				ID:        rec.Id,
				ProjectID: project,
				Title:     source.title,
				Snippet:   rules.Snippet(rec.GetString("body"), snippetLen),
				CreatedAt: rec.GetDateTime("created").Time(),
			})
		}
	}
	return out, nil
}

// searchResolutions finds the one-line answer a cycle closed with, which is
// the record of why a ticket was done and the reason it can be reviewed.
func (r *SearchRepository) searchResolutions(project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	if len(q.Tags) > 0 {
		return nil, nil
	}
	filter, params := textFilter(project, q, "resolution")
	filter += " && resolution != ''"
	records, err := r.app.FindRecordsByFilter(ColCycles, filter, "-created", q.Limit, 0, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.SearchHit, 0, len(records))
	for _, rec := range records {
		out = append(out, domain.SearchHit{
			Kind:      domain.SearchKindCycle,
			ID:        rec.Id,
			ProjectID: project,
			Title:     "Cycle " + strconv.Itoa(rec.GetInt("ordinal")) + " resolution",
			Snippet:   rules.Snippet(rec.GetString("resolution"), snippetLen),
			CreatedAt: rec.GetDateTime("created").Time(),
		})
	}
	return out, nil
}
