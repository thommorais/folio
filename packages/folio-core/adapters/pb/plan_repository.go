package pb

import (
	"context"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type PlanRepository struct {
	app core.App
}

func NewPlanRepository(app core.App) *PlanRepository {
	return &PlanRepository{app: app}
}

var _ ports.PlanRepository = (*PlanRepository)(nil)

func toPlan(rec *core.Record) domain.Plan {
	return domain.Plan{
		ID:        domain.PlanID(rec.Id),
		ProjectID: domain.ProjectID(rec.GetString("project")),
		TicketID:  domain.TicketID(rec.GetString("ticket")),
		Title:     rec.GetString("title"),
		Goal:      rec.GetString("goal"),
		Status:    domain.PlanStatus(rec.GetString("status")),
		Tags:      strSlice(rec, "tags"),
		CreatedBy: domain.UserID(rec.GetString("created_by")),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}
}

func (r *PlanRepository) List(ctx context.Context, project domain.ProjectID, statuses []domain.PlanStatus) ([]domain.Plan, error) {
	exprs := []dbx.Expression{dbx.HashExp{"project": string(project)}}
	if len(statuses) > 0 {
		values := make([]any, 0, len(statuses))
		for _, s := range statuses {
			values = append(values, string(s))
		}
		exprs = append(exprs, dbx.In("status", values...))
	}
	records, err := r.app.FindAllRecords(ColPlans, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Plan, 0, len(records))
	for _, rec := range records {
		out = append(out, toPlan(rec))
	}
	return out, nil
}

// ListByTicket returns every plan filed under a ticket, so deleting the
// ticket can detach them.
func (r *PlanRepository) ListByTicket(ctx context.Context, ticket domain.TicketID) ([]domain.Plan, error) {
	records, err := r.app.FindAllRecords(ColPlans, dbx.HashExp{"ticket": string(ticket)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Plan, 0, len(records))
	for _, rec := range records {
		out = append(out, toPlan(rec))
	}
	return out, nil
}

func (r *PlanRepository) GetByID(ctx context.Context, id domain.PlanID) (domain.Plan, error) {
	rec, err := r.app.FindRecordById(ColPlans, string(id))
	if err != nil {
		return domain.Plan{}, mapErr(err)
	}
	return toPlan(rec), nil
}

func (r *PlanRepository) Create(ctx context.Context, p domain.Plan) (domain.Plan, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColPlans)
	if err != nil {
		return domain.Plan{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(p.ID)
	applyPlan(rec, p)
	if err := r.app.Save(rec); err != nil {
		return domain.Plan{}, mapErr(err)
	}
	return toPlan(rec), nil
}

func (r *PlanRepository) Update(ctx context.Context, p domain.Plan) (domain.Plan, error) {
	rec, err := r.app.FindRecordById(ColPlans, string(p.ID))
	if err != nil {
		return domain.Plan{}, mapErr(err)
	}
	applyPlan(rec, p)
	if err := r.app.Save(rec); err != nil {
		return domain.Plan{}, mapErr(err)
	}
	return toPlan(rec), nil
}

func applyPlan(rec *core.Record, p domain.Plan) {
	rec.Set("project", string(p.ProjectID))
	rec.Set("ticket", string(p.TicketID))
	rec.Set("title", p.Title)
	rec.Set("goal", p.Goal)
	rec.Set("status", string(p.Status))
	setJSON(rec, "tags", p.Tags)
	if p.CreatedBy != "" {
		rec.Set("created_by", string(p.CreatedBy))
	}
}

func (r *PlanRepository) Delete(ctx context.Context, id domain.PlanID) error {
	rec, err := r.app.FindRecordById(ColPlans, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
