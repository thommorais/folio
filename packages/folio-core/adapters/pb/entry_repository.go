package pb

import (
	"context"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type EntryRepository struct {
	app  core.App
	tags *TagRepository
}

func NewEntryRepository(app core.App) *EntryRepository {
	return &EntryRepository{app: app, tags: NewTagRepository(app)}
}

var _ ports.EntryRepository = (*EntryRepository)(nil)

func toEntry(rec *core.Record) domain.Entry {
	return domain.Entry{
		ID:          domain.EntryID(rec.Id),
		Kind:        domain.EntryKind(rec.GetString("kind")),
		ProjectID:   domain.ProjectID(rec.GetString("project")),
		IssueID:     domain.IssueID(rec.GetString("issue")),
		PlanID:      domain.PlanID(rec.GetString("plan")),
		CycleID:     domain.CycleID(rec.GetString("cycle")),
		Slug:        rec.GetString("slug"),
		Title:       rec.GetString("title"),
		Body:        rec.GetString("body"),
		Branch:      rec.GetString("branch"),
		PR:          rec.GetString("pr"),
		ExternalRef: rec.GetString("external_ref"),
		Meta:        jsonMap(rec, "meta"),
		CreatedBy:   domain.UserID(rec.GetString("created_by")),
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}
}

func applyEntry(rec *core.Record, e domain.Entry) {
	rec.Set("kind", string(e.Kind))
	rec.Set("project", string(e.ProjectID))
	rec.Set("issue", string(e.IssueID))
	rec.Set("plan", string(e.PlanID))
	rec.Set("cycle", string(e.CycleID))
	rec.Set("slug", e.Slug)
	rec.Set("title", e.Title)
	rec.Set("body", e.Body)
	rec.Set("branch", e.Branch)
	rec.Set("pr", e.PR)
	rec.Set("external_ref", e.ExternalRef)
	setJSON(rec, "meta", e.Meta)
	if e.CreatedBy != "" {
		rec.Set("created_by", string(e.CreatedBy))
	}
}

func (r *EntryRepository) List(ctx context.Context, project domain.ProjectID, f domain.EntryFilter) ([]domain.Entry, error) {
	exprs := []dbx.Expression{dbx.HashExp{"project": string(project)}}
	if f.Kind != "" {
		exprs = append(exprs, dbx.HashExp{"kind": string(f.Kind)})
	}
	if f.IssueID != "" {
		exprs = append(exprs, dbx.HashExp{"issue": string(f.IssueID)})
	}
	if f.PlanID != "" {
		exprs = append(exprs, dbx.HashExp{"plan": string(f.PlanID)})
	}
	if f.CycleID != "" {
		exprs = append(exprs, dbx.HashExp{"cycle": string(f.CycleID)})
	}
	if f.Branch != "" {
		exprs = append(exprs, dbx.HashExp{"branch": f.Branch})
	}
	if f.ExternalRef != "" {
		exprs = append(exprs, dbx.HashExp{"external_ref": f.ExternalRef})
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		exprs = append(exprs, dbx.Or(dbx.Like("title", q), dbx.Like("body", q)))
	}
	if f.Since != nil {
		exprs = append(exprs, dbx.NewExp("created >= {:since}", dbx.Params{"since": f.Since.UTC().Format("2006-01-02 15:04:05.000Z")}))
	}
	if f.Until != nil {
		exprs = append(exprs, dbx.NewExp("created <= {:until}", dbx.Params{"until": f.Until.UTC().Format("2006-01-02 15:04:05.000Z")}))
	}

	records, err := r.app.FindAllRecords(ColEntries, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Entry, 0, len(records))
	for _, rec := range records {
		out = append(out, toEntry(rec))
	}
	if err := r.attachTags(ctx, out); err != nil {
		return nil, err
	}
	if len(f.Tags) > 0 {
		out = filterEntriesByTags(out, f.Tags)
	}
	return applyPaging(out, f.Offset, f.Limit), nil
}

func (r *EntryRepository) attachTags(ctx context.Context, entries []domain.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, string(e.ID))
	}
	byID, err := r.tags.TagsOf(ctx, ports.TagEntry, ids)
	if err != nil {
		return err
	}
	for i := range entries {
		entries[i].Tags = byID[string(entries[i].ID)]
	}
	return nil
}

func filterEntriesByTags(entries []domain.Entry, want []string) []domain.Entry {
	out := make([]domain.Entry, 0, len(entries))
	for _, entry := range entries {
		have := make(map[string]bool, len(entry.Tags))
		for _, t := range entry.Tags {
			have[strings.ToLower(t)] = true
		}
		match := true
		for _, w := range want {
			if !have[strings.ToLower(w)] {
				match = false
				break
			}
		}
		if match {
			out = append(out, entry)
		}
	}
	return out
}

func (r *EntryRepository) GetByID(ctx context.Context, id domain.EntryID) (domain.Entry, error) {
	rec, err := r.app.FindRecordById(ColEntries, string(id))
	if err != nil {
		return domain.Entry{}, mapErr(err)
	}
	one := []domain.Entry{toEntry(rec)}
	if err := r.attachTags(ctx, one); err != nil {
		return domain.Entry{}, err
	}
	return one[0], nil
}

func (r *EntryRepository) GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.Entry, error) {
	rec, err := r.app.FindFirstRecordByFilter(
		ColEntries,
		"project = {:project} && slug = {:slug}",
		dbx.Params{"project": string(project), "slug": slug},
	)
	if err != nil {
		return domain.Entry{}, mapErr(err)
	}
	one := []domain.Entry{toEntry(rec)}
	if err := r.attachTags(ctx, one); err != nil {
		return domain.Entry{}, err
	}
	return one[0], nil
}

func (r *EntryRepository) Create(ctx context.Context, e domain.Entry) (domain.Entry, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColEntries)
	if err != nil {
		return domain.Entry{}, mapErr(err)
	}
	domainID, err := r.domainOf(e.ProjectID)
	if err != nil {
		return domain.Entry{}, err
	}

	rec := core.NewRecord(collection)
	rec.Id = string(e.ID)
	rec.Set("domain", domainID)
	applyEntry(rec, e)
	if err := r.app.Save(rec); err != nil {
		return domain.Entry{}, mapErr(err)
	}
	if err := r.tags.SetTags(ctx, ports.TagEntry, rec.Id, domain.DomainID(domainID), e.Tags); err != nil {
		return domain.Entry{}, err
	}
	out := toEntry(rec)
	out.Tags = e.Tags
	return out, nil
}

func (r *EntryRepository) domainOf(project domain.ProjectID) (string, error) {
	rec, err := r.app.FindRecordById(ColProjects, string(project))
	if err != nil {
		return "", mapErr(err)
	}
	return rec.GetString("domain"), nil
}

func (r *EntryRepository) Update(ctx context.Context, e domain.Entry) (domain.Entry, error) {
	rec, err := r.app.FindRecordById(ColEntries, string(e.ID))
	if err != nil {
		return domain.Entry{}, mapErr(err)
	}
	applyEntry(rec, e)
	if err := r.app.Save(rec); err != nil {
		return domain.Entry{}, mapErr(err)
	}
	if err := r.tags.SetTags(ctx, ports.TagEntry, rec.Id, domain.DomainID(rec.GetString("domain")), e.Tags); err != nil {
		return domain.Entry{}, err
	}
	out := toEntry(rec)
	out.Tags = e.Tags
	return out, nil
}

func (r *EntryRepository) Delete(ctx context.Context, id domain.EntryID) error {
	rec, err := r.app.FindRecordById(ColEntries, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
