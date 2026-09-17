package pb

import (
	"context"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type IssueRepository struct {
	app core.App
}

func NewIssueRepository(app core.App) *IssueRepository {
	return &IssueRepository{app: app}
}

var _ ports.IssueRepository = (*IssueRepository)(nil)

func toIssue(rec *core.Record) domain.Issue {
	return domain.Issue{
		ID:          domain.IssueID(rec.Id),
		Kind:        domain.IssueKind(rec.GetString("kind")),
		ProjectID:   domain.ProjectID(rec.GetString("project")),
		PlanID:      domain.PlanID(rec.GetString("plan")),
		Slug:        rec.GetString("slug"),
		Title:       rec.GetString("title"),
		Body:        rec.GetString("body"),
		Status:      domain.IssueStatus(rec.GetString("status")),
		Priority:    domain.Priority(rec.GetString("priority")),
		Size:        domain.Size(rec.GetInt("size")),
		Assignee:    domain.UserID(rec.GetString("assignee")),
		Tags:        strSlice(rec, "tags"),
		Position:    rec.GetInt("position"),
		DueDate:     timePtr(rec.GetDateTime("due_date")),
		Wayfinder:   domain.WayfinderType(rec.GetString("wayfinder")),
		ExternalRef: rec.GetString("external_ref"),
		CreatedBy:   domain.UserID(rec.GetString("created_by")),
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}
}

func applyIssue(rec *core.Record, i domain.Issue) {
	rec.Set("kind", string(i.Kind))
	rec.Set("project", string(i.ProjectID))
	rec.Set("plan", string(i.PlanID))
	rec.Set("slug", i.Slug)
	rec.Set("title", i.Title)
	rec.Set("body", i.Body)
	rec.Set("status", string(i.Status))
	rec.Set("priority", string(i.Priority))
	rec.Set("size", int(i.Size))
	rec.Set("assignee", string(i.Assignee))
	setJSON(rec, "tags", i.Tags)
	rec.Set("position", i.Position)
	setDate(rec, "due_date", i.DueDate)
	rec.Set("wayfinder", string(i.Wayfinder))
	rec.Set("external_ref", i.ExternalRef)
	if i.CreatedBy != "" {
		rec.Set("created_by", string(i.CreatedBy))
	}
}

func (r *IssueRepository) List(ctx context.Context, project domain.ProjectID, f domain.IssueFilter) ([]domain.Issue, error) {
	exprs := []dbx.Expression{dbx.HashExp{"project": string(project)}}
	if f.Kind != "" {
		exprs = append(exprs, dbx.HashExp{"kind": string(f.Kind)})
	}
	if f.PlanID != "" {
		exprs = append(exprs, dbx.HashExp{"plan": string(f.PlanID)})
	}
	if len(f.Status) > 0 {
		values := make([]any, 0, len(f.Status))
		for _, s := range f.Status {
			values = append(values, string(s))
		}
		exprs = append(exprs, dbx.In("status", values...))
	}
	if f.Priority != "" {
		exprs = append(exprs, dbx.HashExp{"priority": string(f.Priority)})
	}
	if f.Assignee != "" {
		exprs = append(exprs, dbx.HashExp{"assignee": string(f.Assignee)})
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		exprs = append(exprs, dbx.Or(dbx.Like("title", q), dbx.Like("body", q)))
	}
	for _, tag := range f.Tags {
		exprs = append(exprs, dbx.Like("tags", `"`+tag+`"`))
	}

	records, err := r.app.FindAllRecords(ColIssues, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Issue, 0, len(records))
	for _, rec := range records {
		out = append(out, toIssue(rec))
	}

	if f.ParentID != "" {
		children, err := r.childrenOf(f.ParentID)
		if err != nil {
			return nil, err
		}
		out = filterByID(out, children)
	}
	return applyPaging(out, f.Offset, f.Limit), nil
}

func filterByID(issues []domain.Issue, keep map[domain.IssueID]bool) []domain.Issue {
	out := make([]domain.Issue, 0, len(keep))
	for _, i := range issues {
		if keep[i.ID] {
			out = append(out, i)
		}
	}
	return out
}

func (r *IssueRepository) childrenOf(parent domain.IssueID) (map[domain.IssueID]bool, error) {
	rows, err := r.app.FindAllRecords(ColLinks, dbx.HashExp{"to": string(parent), "kind": string(domain.LinkParent)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make(map[domain.IssueID]bool, len(rows))
	for _, row := range rows {
		out[domain.IssueID(row.GetString("from"))] = true
	}
	return out, nil
}

func (r *IssueRepository) ListByParent(ctx context.Context, parent domain.IssueID) ([]domain.Issue, error) {
	children, err := r.childrenOf(parent)
	if err != nil {
		return nil, err
	}
	return r.byIDs(children)
}

func (r *IssueRepository) byIDs(ids map[domain.IssueID]bool) ([]domain.Issue, error) {
	if len(ids) == 0 {
		return []domain.Issue{}, nil
	}
	values := make([]any, 0, len(ids))
	for id := range ids {
		values = append(values, string(id))
	}
	records, err := r.app.FindAllRecords(ColIssues, dbx.In("id", values...))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Issue, 0, len(records))
	for _, rec := range records {
		out = append(out, toIssue(rec))
	}
	return out, nil
}

func (r *IssueRepository) ListByPlan(ctx context.Context, plan domain.PlanID) ([]domain.Issue, error) {
	records, err := r.app.FindAllRecords(ColIssues, dbx.HashExp{"plan": string(plan)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Issue, 0, len(records))
	for _, rec := range records {
		out = append(out, toIssue(rec))
	}
	return out, nil
}

func (r *IssueRepository) GetByID(ctx context.Context, id domain.IssueID) (domain.Issue, error) {
	rec, err := r.app.FindRecordById(ColIssues, string(id))
	if err != nil {
		return domain.Issue{}, mapErr(err)
	}
	return toIssue(rec), nil
}

func (r *IssueRepository) GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.Issue, error) {
	rec, err := r.app.FindFirstRecordByFilter(
		ColIssues,
		"project = {:project} && slug = {:slug}",
		dbx.Params{"project": string(project), "slug": slug},
	)
	if err != nil {
		return domain.Issue{}, mapErr(err)
	}
	return toIssue(rec), nil
}

func (r *IssueRepository) Create(ctx context.Context, i domain.Issue) (domain.Issue, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColIssues)
	if err != nil {
		return domain.Issue{}, mapErr(err)
	}
	domainID, err := r.domainOf(i.ProjectID)
	if err != nil {
		return domain.Issue{}, err
	}

	rec := core.NewRecord(collection)
	rec.Id = string(i.ID)
	rec.Set("domain", domainID)
	applyIssue(rec, i)
	if err := r.app.Save(rec); err != nil {
		return domain.Issue{}, mapErr(err)
	}
	return toIssue(rec), nil
}

func (r *IssueRepository) domainOf(project domain.ProjectID) (string, error) {
	rec, err := r.app.FindRecordById(ColProjects, string(project))
	if err != nil {
		return "", mapErr(err)
	}
	return rec.GetString("domain"), nil
}

func (r *IssueRepository) Update(ctx context.Context, i domain.Issue) (domain.Issue, error) {
	rec, err := r.app.FindRecordById(ColIssues, string(i.ID))
	if err != nil {
		return domain.Issue{}, mapErr(err)
	}
	applyIssue(rec, i)
	if err := r.app.Save(rec); err != nil {
		return domain.Issue{}, mapErr(err)
	}
	return toIssue(rec), nil
}

func (r *IssueRepository) Delete(ctx context.Context, id domain.IssueID) error {
	rec, err := r.app.FindRecordById(ColIssues, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}

func (r *IssueRepository) Links(ctx context.Context, id domain.IssueID) ([]domain.IssueLink, error) {
	rows, err := r.app.FindAllRecords(ColLinks, dbx.Or(
		dbx.HashExp{"from": string(id)},
		dbx.HashExp{"to": string(id)},
	))
	if err != nil {
		return nil, mapErr(err)
	}
	return toLinks(rows), nil
}

func (r *IssueRepository) LinksOfProject(ctx context.Context, project domain.ProjectID) ([]domain.IssueLink, error) {
	issues, err := r.app.FindAllRecords(ColIssues, dbx.HashExp{"project": string(project)})
	if err != nil {
		return nil, mapErr(err)
	}
	if len(issues) == 0 {
		return []domain.IssueLink{}, nil
	}
	ids := make([]any, 0, len(issues))
	for _, rec := range issues {
		ids = append(ids, rec.Id)
	}
	rows, err := r.app.FindAllRecords(ColLinks, dbx.Or(
		dbx.In("from", ids...),
		dbx.In("to", ids...),
	))
	if err != nil {
		return nil, mapErr(err)
	}
	return toLinks(rows), nil
}

func toLinks(rows []*core.Record) []domain.IssueLink {
	out := make([]domain.IssueLink, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.IssueLink{
			ID:       row.Id,
			DomainID: domain.DomainID(row.GetString("domain")),
			From:     domain.IssueID(row.GetString("from")),
			To:       domain.IssueID(row.GetString("to")),
			Kind:     domain.LinkKind(row.GetString("kind")),
		})
	}
	return out
}

func (r *IssueRepository) Link(ctx context.Context, from, to domain.IssueID, kind domain.LinkKind) error {
	existing, err := r.app.FindFirstRecordByFilter(
		ColLinks,
		"from = {:from} && to = {:to} && kind = {:kind}",
		dbx.Params{"from": string(from), "to": string(to), "kind": string(kind)},
	)
	if err == nil && existing != nil {
		return nil
	}

	src, err := r.app.FindRecordById(ColIssues, string(from))
	if err != nil {
		return mapErr(err)
	}
	collection, err := r.app.FindCollectionByNameOrId(ColLinks)
	if err != nil {
		return mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Set("domain", src.GetString("domain"))
	rec.Set("from", string(from))
	rec.Set("to", string(to))
	rec.Set("kind", string(kind))
	return mapErr(r.app.Save(rec))
}

func (r *IssueRepository) Unlink(ctx context.Context, from, to domain.IssueID, kind domain.LinkKind) error {
	rec, err := r.app.FindFirstRecordByFilter(
		ColLinks,
		"from = {:from} && to = {:to} && kind = {:kind}",
		dbx.Params{"from": string(from), "to": string(to), "kind": string(kind)},
	)
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}

func (r *IssueRepository) SetLinks(ctx context.Context, from domain.IssueID, kind domain.LinkKind, to []domain.IssueID) error {
	rows, err := r.app.FindAllRecords(ColLinks, dbx.HashExp{"from": string(from), "kind": string(kind)})
	if err != nil {
		return mapErr(err)
	}
	want := make(map[domain.IssueID]bool, len(to))
	for _, id := range to {
		want[id] = true
	}

	for _, row := range rows {
		target := domain.IssueID(row.GetString("to"))
		if want[target] {
			delete(want, target)
			continue
		}
		if err := r.app.Delete(row); err != nil {
			return mapErr(err)
		}
	}
	for id := range want {
		if err := r.Link(ctx, from, id, kind); err != nil {
			return err
		}
	}
	return nil
}
