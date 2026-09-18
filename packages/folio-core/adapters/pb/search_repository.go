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

type SearchRepository struct {
	app core.App
}

func NewSearchRepository(app core.App) *SearchRepository {
	return &SearchRepository{app: app}
}

var _ ports.SearchRepository = (*SearchRepository)(nil)

const snippetLen = 200

func (r *SearchRepository) Search(ctx context.Context, project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	want := map[domain.SearchKind]bool{}
	for _, k := range q.Kinds {
		want[k] = true
	}
	all := len(want) == 0

	var hits []domain.SearchHit

	issueKinds := make([]string, 0, 2)
	if all || want[domain.SearchKindTicket] {
		issueKinds = append(issueKinds, string(domain.IssueTicket))
	}
	if all || want[domain.SearchKindTodo] {
		issueKinds = append(issueKinds, string(domain.IssueTodo))
	}
	if len(issueKinds) > 0 {
		found, err := r.searchIssues(project, q, issueKinds)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}

	entryKinds := make([]string, 0, 3)
	if all || want[domain.SearchKindJournal] {
		entryKinds = append(entryKinds, string(domain.EntryJournal))
	}
	if all || want[domain.SearchKindDoc] {
		entryKinds = append(entryKinds, string(domain.EntryDoc))
	}
	if all || want[domain.SearchKindWorkLog] {
		entryKinds = append(entryKinds, string(domain.EntryLog))
	}
	if len(entryKinds) > 0 {
		found, err := r.searchEntries(project, q, entryKinds)
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
	if all || want[domain.SearchKindCycle] {
		found, err := r.searchCycles(project, q)
		if err != nil {
			return nil, err
		}
		hits = append(hits, found...)
	}

	sort.SliceStable(hits, func(i, j int) bool { return hits[i].CreatedAt.After(hits[j].CreatedAt) })
	return applyPaging(hits, q.Offset, q.Limit), nil
}

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

func kindClause(filter string, params dbx.Params, kinds []string) (string, dbx.Params) {
	ors := make([]string, 0, len(kinds))
	for i, k := range kinds {
		key := "kind" + strconv.Itoa(i)
		ors = append(ors, "kind = {:"+key+"}")
		params[key] = k
	}
	return filter + " && (" + strings.Join(ors, " || ") + ")", params
}

func (r *SearchRepository) searchIssues(project domain.ProjectID, q domain.SearchQuery, kinds []string) ([]domain.SearchHit, error) {
	filter, params := textFilter(project, q, "title", "body")
	filter, params = kindClause(filter, params, kinds)

	records, err := r.app.FindRecordsByFilter(ColIssues, filter, "-created", q.Limit, 0, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.SearchHit, 0, len(records))
	for _, rec := range records {
		kind := domain.SearchKindTicket
		if rec.GetString("kind") == string(domain.IssueTodo) {
			kind = domain.SearchKindTodo
		}
		out = append(out, domain.SearchHit{
			Kind:      kind,
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

var entrySearchKind = map[string]domain.SearchKind{
	string(domain.EntryJournal): domain.SearchKindJournal,
	string(domain.EntryDoc):     domain.SearchKindDoc,
	string(domain.EntryLog):     domain.SearchKindWorkLog,
}

func (r *SearchRepository) searchEntries(project domain.ProjectID, q domain.SearchQuery, kinds []string) ([]domain.SearchHit, error) {
	filter, params := textFilter(project, q, "title", "body")
	filter, params = kindClause(filter, params, kinds)

	records, err := r.app.FindRecordsByFilter(ColEntries, filter, "-created", q.Limit, 0, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.SearchHit, 0, len(records))
	for _, rec := range records {
		title := rec.GetString("title")
		if title == "" {
			title = "Work log"
		}
		out = append(out, domain.SearchHit{
			Kind:      entrySearchKind[rec.GetString("kind")],
			ID:        rec.Id,
			ProjectID: project,
			Title:     title,
			Snippet:   rules.Snippet(rec.GetString("body"), snippetLen),
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

func (r *SearchRepository) searchCycles(project domain.ProjectID, q domain.SearchQuery) ([]domain.SearchHit, error) {
	filter := []string{"project = {:project}", "resolution != ''"}
	params := dbx.Params{"project": string(project)}
	if text := strings.TrimSpace(q.Text); text != "" {
		filter = append(filter, "resolution ~ {:text}")
		params["text"] = text
	}
	if len(q.Tags) > 0 {
		return []domain.SearchHit{}, nil
	}

	records, err := r.app.FindRecordsByFilter(ColCycles, strings.Join(filter, " && "), "-created", q.Limit, 0, params)
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
