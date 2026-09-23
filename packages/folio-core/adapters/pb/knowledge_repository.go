package pb

import (
	"context"
	"sort"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// KnowledgeRepository stores notes that belong to no domain, so unlike every
// other repository here it never resolves one and never filters by project.
type KnowledgeRepository struct {
	app core.App
}

func NewKnowledgeRepository(app core.App) *KnowledgeRepository {
	return &KnowledgeRepository{app: app}
}

var _ ports.KnowledgeRepository = (*KnowledgeRepository)(nil)

func toKnowledge(rec *core.Record) domain.Knowledge {
	return domain.Knowledge{
		ID:        domain.KnowledgeID(rec.Id),
		ProjectID: domain.ProjectID(rec.GetString("project")),
		Slug:      rec.GetString("slug"),
		Title:     rec.GetString("title"),
		Body:      rec.GetString("body"),
		Tags:      strSlice(rec, "tags"),
		CreatedBy: domain.UserID(rec.GetString("created_by")),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}
}

// applyKnowledge writes the tags column, which the search index reads. Entries
// and issues keep their tags in a join table instead; knowledge has no domain
// to scope one by, so here the column is the record of truth.
func applyKnowledge(rec *core.Record, k domain.Knowledge) {
	rec.Set("project", string(k.ProjectID))
	rec.Set("slug", k.Slug)
	rec.Set("title", k.Title)
	rec.Set("body", k.Body)
	setJSON(rec, "tags", k.Tags)
	if k.CreatedBy != "" {
		rec.Set("created_by", string(k.CreatedBy))
	}
}

func (r *KnowledgeRepository) List(ctx context.Context, f domain.KnowledgeFilter) ([]domain.Knowledge, error) {
	exprs := []dbx.Expression{}
	switch {
	case f.Unattached:
		exprs = append(exprs, dbx.HashExp{"project": ""})
	case f.ProjectID != "":
		exprs = append(exprs, dbx.HashExp{"project": string(f.ProjectID)})
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		exprs = append(exprs, dbx.Or(
			dbx.Like("title", search),
			dbx.Like("body", search),
		))
	}

	records, err := r.app.FindAllRecords(ColKnowledge, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}

	out := make([]domain.Knowledge, 0, len(records))
	for _, rec := range records {
		note := toKnowledge(rec)
		if !hasEveryTag(note.Tags, f.Tags) {
			continue
		}
		out = append(out, note)
	}
	// Newest first: a listing is a browse, and the ranked view is search.
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return applyPaging(out, f.Offset, f.Limit), nil
}

// hasEveryTag reports whether a note carries all the requested tags, compared
// case-insensitively the way the tag join tables do.
func hasEveryTag(have, want []string) bool {
	if len(want) == 0 {
		return true
	}
	set := make(map[string]bool, len(have))
	for _, t := range have {
		set[strings.ToLower(t)] = true
	}
	for _, t := range want {
		if !set[strings.ToLower(strings.TrimSpace(t))] {
			return false
		}
	}
	return true
}

func (r *KnowledgeRepository) GetByID(ctx context.Context, id domain.KnowledgeID) (domain.Knowledge, error) {
	rec, err := r.app.FindRecordById(ColKnowledge, string(id))
	if err != nil {
		return domain.Knowledge{}, mapErr(err)
	}
	return toKnowledge(rec), nil
}

func (r *KnowledgeRepository) GetBySlug(ctx context.Context, slug string) (domain.Knowledge, error) {
	rec, err := r.app.FindFirstRecordByFilter(ColKnowledge, "slug = {:slug}", dbx.Params{"slug": slug})
	if err != nil {
		return domain.Knowledge{}, mapErr(err)
	}
	return toKnowledge(rec), nil
}

func (r *KnowledgeRepository) Create(ctx context.Context, k domain.Knowledge) (domain.Knowledge, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColKnowledge)
	if err != nil {
		return domain.Knowledge{}, mapErr(err)
	}

	rec := core.NewRecord(collection)
	rec.Id = string(k.ID)
	applyKnowledge(rec, k)
	if err := r.app.Save(rec); err != nil {
		return domain.Knowledge{}, mapErr(err)
	}
	return toKnowledge(rec), nil
}

func (r *KnowledgeRepository) Update(ctx context.Context, k domain.Knowledge) (domain.Knowledge, error) {
	rec, err := r.app.FindRecordById(ColKnowledge, string(k.ID))
	if err != nil {
		return domain.Knowledge{}, mapErr(err)
	}
	applyKnowledge(rec, k)
	if err := r.app.Save(rec); err != nil {
		return domain.Knowledge{}, mapErr(err)
	}
	return toKnowledge(rec), nil
}

func (r *KnowledgeRepository) Delete(ctx context.Context, id domain.KnowledgeID) error {
	rec, err := r.app.FindRecordById(ColKnowledge, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
