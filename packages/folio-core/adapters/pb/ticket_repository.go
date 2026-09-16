package pb

import (
	"context"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type TicketRepository struct {
	app core.App
}

func NewTicketRepository(app core.App) *TicketRepository {
	return &TicketRepository{app: app}
}

var _ ports.TicketRepository = (*TicketRepository)(nil)

func toTicket(rec *core.Record) domain.Ticket {
	return domain.Ticket{
		ID:          domain.TicketID(rec.Id),
		ProjectID:   domain.ProjectID(rec.GetString("project")),
		ParentID:    domain.TicketID(rec.GetString("parent")),
		Slug:        rec.GetString("slug"),
		Title:       rec.GetString("title"),
		Body:        rec.GetString("body"),
		Status:      domain.TicketStatus(rec.GetString("status")),
		Priority:    domain.Priority(rec.GetString("priority")),
		Assignee:    domain.UserID(rec.GetString("assignee")),
		Tags:        strSlice(rec, "tags"),
		ExternalRef: rec.GetString("external_ref"),
		DependsOn:   toTicketIDs(strSlice(rec, "depends_on")),
		Wayfinder:   domain.WayfinderType(rec.GetString("wayfinder")),
		CreatedBy:   domain.UserID(rec.GetString("created_by")),
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}
}

func (r *TicketRepository) List(ctx context.Context, project domain.ProjectID, f domain.TicketFilter) ([]domain.Ticket, error) {
	exprs := []dbx.Expression{dbx.HashExp{"project": string(project)}}
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
	if f.ParentID != "" {
		exprs = append(exprs, dbx.HashExp{"parent": string(f.ParentID)})
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		exprs = append(exprs, dbx.Or(
			dbx.Like("title", q),
			dbx.Like("body", q),
		))
	}
	for _, tag := range f.Tags {
		exprs = append(exprs, dbx.Like("tags", `"`+tag+`"`))
	}

	records, err := r.app.FindAllRecords(ColTickets, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Ticket, 0, len(records))
	for _, rec := range records {
		out = append(out, toTicket(rec))
	}
	return applyPaging(out, f.Offset, f.Limit), nil
}

func (r *TicketRepository) ListByParent(ctx context.Context, parent domain.TicketID) ([]domain.Ticket, error) {
	records, err := r.app.FindAllRecords(ColTickets, dbx.HashExp{"parent": string(parent)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Ticket, 0, len(records))
	for _, rec := range records {
		out = append(out, toTicket(rec))
	}
	return out, nil
}

func (r *TicketRepository) GetByID(ctx context.Context, id domain.TicketID) (domain.Ticket, error) {
	rec, err := r.app.FindRecordById(ColTickets, string(id))
	if err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	return toTicket(rec), nil
}

func (r *TicketRepository) GetBySlug(ctx context.Context, project domain.ProjectID, slug string) (domain.Ticket, error) {
	rec, err := r.app.FindFirstRecordByFilter(
		ColTickets,
		"project = {:project} && slug = {:slug}",
		dbx.Params{"project": string(project), "slug": slug},
	)
	if err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	return toTicket(rec), nil
}

func (r *TicketRepository) Create(ctx context.Context, t domain.Ticket) (domain.Ticket, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColTickets)
	if err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(t.ID)
	applyTicket(rec, t)
	if err := r.app.Save(rec); err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	return toTicket(rec), nil
}

func (r *TicketRepository) Update(ctx context.Context, t domain.Ticket) (domain.Ticket, error) {
	rec, err := r.app.FindRecordById(ColTickets, string(t.ID))
	if err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	applyTicket(rec, t)
	if err := r.app.Save(rec); err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	return toTicket(rec), nil
}

func applyTicket(rec *core.Record, t domain.Ticket) {
	rec.Set("project", string(t.ProjectID))
	rec.Set("parent", string(t.ParentID))
	rec.Set("slug", t.Slug)
	rec.Set("title", t.Title)
	rec.Set("body", t.Body)
	rec.Set("status", string(t.Status))
	rec.Set("priority", string(t.Priority))
	rec.Set("assignee", string(t.Assignee))
	setJSON(rec, "tags", t.Tags)
	rec.Set("external_ref", t.ExternalRef)
	setJSON(rec, "depends_on", fromTicketIDs(t.DependsOn))
	rec.Set("wayfinder", string(t.Wayfinder))
	if t.CreatedBy != "" {
		rec.Set("created_by", string(t.CreatedBy))
	}
}

func (r *TicketRepository) Delete(ctx context.Context, id domain.TicketID) error {
	rec, err := r.app.FindRecordById(ColTickets, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
