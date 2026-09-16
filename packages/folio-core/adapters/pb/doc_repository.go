package pb

import (
	"context"
	"strconv"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type DocRepository struct {
	app core.App
}

func NewDocRepository(app core.App) *DocRepository {
	return &DocRepository{app: app}
}

var _ ports.DocRepository = (*DocRepository)(nil)

func toDoc(rec *core.Record) domain.Doc {
	return domain.Doc{
		ID:        domain.DocID(rec.Id),
		ProjectID: domain.ProjectID(rec.GetString("project")),
		TicketID:  domain.TicketID(rec.GetString("ticket")),
		Slug:      rec.GetString("slug"),
		Title:     rec.GetString("title"),
		Body:      rec.GetString("body"),
		Tags:      strSlice(rec, "tags"),
		CreatedBy: domain.UserID(rec.GetString("created_by")),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}
}

func (r *DocRepository) List(ctx context.Context, project domain.ProjectID, f domain.DocFilter) ([]domain.Doc, error) {
	filter := []string{"project = {:project}"}
	params := dbx.Params{"project": string(project)}

	if f.TicketID != "" {
		filter = append(filter, "ticket = {:ticket}")
		params["ticket"] = string(f.TicketID)
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		filter = append(filter, "(title ~ {:search} || body ~ {:search})")
		params["search"] = q
	}
	for i, tag := range f.Tags {
		key := "tag" + strconv.Itoa(i)
		filter = append(filter, "tags ~ {:"+key+"}")
		params[key] = `"` + tag + `"`
	}

	records, err := r.app.FindRecordsByFilter(ColDocs, strings.Join(filter, " && "), "title", f.Limit, f.Offset, params)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Doc, 0, len(records))
	for _, rec := range records {
		out = append(out, toDoc(rec))
	}
	return out, nil
}

func (r *DocRepository) GetByID(ctx context.Context, id domain.DocID) (domain.Doc, error) {
	rec, err := r.app.FindRecordById(ColDocs, string(id))
	if err != nil {
		return domain.Doc{}, mapErr(err)
	}
	return toDoc(rec), nil
}

func (r *DocRepository) GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.Doc, error) {
	rec, err := r.app.FindFirstRecordByFilter(
		ColDocs,
		"project = {:project} && slug = {:slug}",
		dbx.Params{"project": string(project), "slug": slug},
	)
	if err != nil {
		return domain.Doc{}, mapErr(err)
	}
	return toDoc(rec), nil
}

func (r *DocRepository) Create(ctx context.Context, d domain.Doc) (domain.Doc, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColDocs)
	if err != nil {
		return domain.Doc{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(d.ID)
	applyDoc(rec, d)
	if err := r.app.Save(rec); err != nil {
		return domain.Doc{}, mapErr(err)
	}
	return toDoc(rec), nil
}

func (r *DocRepository) Update(ctx context.Context, d domain.Doc) (domain.Doc, error) {
	rec, err := r.app.FindRecordById(ColDocs, string(d.ID))
	if err != nil {
		return domain.Doc{}, mapErr(err)
	}
	applyDoc(rec, d)
	if err := r.app.Save(rec); err != nil {
		return domain.Doc{}, mapErr(err)
	}
	return toDoc(rec), nil
}

func applyDoc(rec *core.Record, d domain.Doc) {
	rec.Set("project", string(d.ProjectID))
	rec.Set("ticket", string(d.TicketID))
	rec.Set("slug", d.Slug)
	rec.Set("title", d.Title)
	rec.Set("body", d.Body)
	setJSON(rec, "tags", d.Tags)
	if d.CreatedBy != "" {
		rec.Set("created_by", string(d.CreatedBy))
	}
}

func (r *DocRepository) Delete(ctx context.Context, id domain.DocID) error {
	rec, err := r.app.FindRecordById(ColDocs, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
