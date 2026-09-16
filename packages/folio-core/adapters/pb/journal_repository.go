package pb

import (
	"context"
	"strconv"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type JournalRepository struct {
	app core.App
}

func NewJournalRepository(app core.App) *JournalRepository {
	return &JournalRepository{app: app}
}

var _ ports.JournalRepository = (*JournalRepository)(nil)

func toJournalEntry(rec *core.Record) domain.JournalEntry {
	return domain.JournalEntry{
		ID:          domain.JournalID(rec.Id),
		ProjectID:   domain.ProjectID(rec.GetString("project")),
		Slug:        rec.GetString("slug"),
		PlanID:      domain.PlanID(rec.GetString("plan")),
		TodoID:      domain.TodoID(rec.GetString("todo")),
		Title:       rec.GetString("title"),
		Body:        rec.GetString("body"),
		Branch:      rec.GetString("branch"),
		PR:          rec.GetString("pr"),
		TicketID:    domain.TicketID(rec.GetString("ticket")),
		ExternalRef: rec.GetString("external_ref"),
		Meta:        jsonMap(rec, "meta"),
		Tags:        strSlice(rec, "tags"),
		CreatedBy:   domain.UserID(rec.GetString("created_by")),
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}
}

// List returns entries newest first: the recent work is what a reader
// catching up on a project needs first.
func (r *JournalRepository) List(ctx context.Context, project domain.ProjectID, f domain.JournalFilter) ([]domain.JournalEntry, error) {
	filter := []string{"project = {:project}"}
	params := dbx.Params{"project": string(project)}

	if f.PlanID != "" {
		filter = append(filter, "plan = {:plan}")
		params["plan"] = string(f.PlanID)
	}
	if f.TodoID != "" {
		filter = append(filter, "todo = {:todo}")
		params["todo"] = string(f.TodoID)
	}
	if f.Branch != "" {
		filter = append(filter, "branch = {:branch}")
		params["branch"] = f.Branch
	}
	if f.TicketID != "" {
		filter = append(filter, "ticket = {:ticket}")
		params["ticket"] = string(f.TicketID)
	}
	if f.ExternalRef != "" {
		filter = append(filter, "external_ref = {:external_ref}")
		params["external_ref"] = f.ExternalRef
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
	if f.Since != nil {
		filter = append(filter, "created >= {:since}")
		params["since"] = f.Since.UTC().Format(types.DefaultDateLayout)
	}
	if f.Until != nil {
		filter = append(filter, "created <= {:until}")
		params["until"] = f.Until.UTC().Format(types.DefaultDateLayout)
	}

	records, err := r.app.FindRecordsByFilter(
		ColJournal,
		strings.Join(filter, " && "),
		"-created",
		f.Limit,
		f.Offset,
		params,
	)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.JournalEntry, 0, len(records))
	for _, rec := range records {
		out = append(out, toJournalEntry(rec))
	}
	return out, nil
}

func (r *JournalRepository) GetByID(ctx context.Context, id domain.JournalID) (domain.JournalEntry, error) {
	rec, err := r.app.FindRecordById(ColJournal, string(id))
	if err != nil {
		return domain.JournalEntry{}, mapErr(err)
	}
	return toJournalEntry(rec), nil
}

func (r *JournalRepository) GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.JournalEntry, error) {
	rec, err := r.app.FindFirstRecordByFilter(
		ColJournal,
		"project = {:project} && slug = {:slug}",
		dbx.Params{"project": string(project), "slug": slug},
	)
	if err != nil {
		return domain.JournalEntry{}, mapErr(err)
	}
	return toJournalEntry(rec), nil
}

func (r *JournalRepository) Create(ctx context.Context, e domain.JournalEntry) (domain.JournalEntry, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColJournal)
	if err != nil {
		return domain.JournalEntry{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(e.ID)
	applyJournalEntry(rec, e)
	if err := r.app.Save(rec); err != nil {
		return domain.JournalEntry{}, mapErr(err)
	}
	return toJournalEntry(rec), nil
}

func (r *JournalRepository) Update(ctx context.Context, e domain.JournalEntry) (domain.JournalEntry, error) {
	rec, err := r.app.FindRecordById(ColJournal, string(e.ID))
	if err != nil {
		return domain.JournalEntry{}, mapErr(err)
	}
	applyJournalEntry(rec, e)
	if err := r.app.Save(rec); err != nil {
		return domain.JournalEntry{}, mapErr(err)
	}
	return toJournalEntry(rec), nil
}

func applyJournalEntry(rec *core.Record, e domain.JournalEntry) {
	rec.Set("project", string(e.ProjectID))
	rec.Set("slug", e.Slug)
	rec.Set("plan", string(e.PlanID))
	rec.Set("todo", string(e.TodoID))
	rec.Set("title", e.Title)
	rec.Set("body", e.Body)
	rec.Set("branch", e.Branch)
	rec.Set("pr", e.PR)
	rec.Set("ticket", string(e.TicketID))
	rec.Set("external_ref", e.ExternalRef)
	setJSON(rec, "meta", e.Meta)
	setJSON(rec, "tags", e.Tags)
	if e.CreatedBy != "" {
		rec.Set("created_by", string(e.CreatedBy))
	}
}

func (r *JournalRepository) Delete(ctx context.Context, id domain.JournalID) error {
	rec, err := r.app.FindRecordById(ColJournal, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
