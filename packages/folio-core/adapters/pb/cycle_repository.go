package pb

import (
	"context"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type CycleRepository struct {
	app core.App
}

func NewCycleRepository(app core.App) *CycleRepository {
	return &CycleRepository{app: app}
}

var _ ports.CycleRepository = (*CycleRepository)(nil)

func toCycle(rec *core.Record) domain.Cycle {
	return domain.Cycle{
		ID:         domain.CycleID(rec.Id),
		ProjectID:  domain.ProjectID(rec.GetString("project")),
		IssueID:   domain.IssueID(rec.GetString("issue")),
		Ordinal:    rec.GetInt("ordinal"),
		Phase:      domain.Phase(rec.GetString("phase")),
		Resolution: rec.GetString("resolution"),
		MapID:      domain.IssueID(rec.GetString("map")),
		CreatedBy:  domain.UserID(rec.GetString("created_by")),
		CreatedAt:  rec.GetDateTime("created").Time(),
		UpdatedAt:  rec.GetDateTime("updated").Time(),
		ClosedAt:   timePtr(rec.GetDateTime("closed_at")),
	}
}

func applyCycle(rec *core.Record, c domain.Cycle) {
	rec.Set("project", string(c.ProjectID))
	rec.Set("issue", string(c.IssueID))
	rec.Set("ordinal", c.Ordinal)
	rec.Set("phase", string(c.Phase))
	rec.Set("resolution", c.Resolution)
	rec.Set("map", string(c.MapID))
	setDate(rec, "closed_at", c.ClosedAt)
	if c.CreatedBy != "" {
		rec.Set("created_by", string(c.CreatedBy))
	}
}

func (r *CycleRepository) ListByIssue(ctx context.Context, issue domain.IssueID) ([]domain.Cycle, error) {
	records, err := r.app.FindAllRecords(ColCycles, dbx.HashExp{"issue": string(issue)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Cycle, 0, len(records))
	for _, rec := range records {
		out = append(out, toCycle(rec))
	}
	return out, nil
}

func (r *CycleRepository) GetByID(ctx context.Context, id domain.CycleID) (domain.Cycle, error) {
	rec, err := r.app.FindRecordById(ColCycles, string(id))
	if err != nil {
		return domain.Cycle{}, mapErr(err)
	}
	return toCycle(rec), nil
}

func (r *CycleRepository) Create(ctx context.Context, c domain.Cycle) (domain.Cycle, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColCycles)
	if err != nil {
		return domain.Cycle{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(c.ID)
	applyCycle(rec, c)
	if err := r.app.Save(rec); err != nil {
		return domain.Cycle{}, mapErr(err)
	}
	return toCycle(rec), nil
}

func (r *CycleRepository) Update(ctx context.Context, c domain.Cycle) (domain.Cycle, error) {
	rec, err := r.app.FindRecordById(ColCycles, string(c.ID))
	if err != nil {
		return domain.Cycle{}, mapErr(err)
	}
	applyCycle(rec, c)
	if err := r.app.Save(rec); err != nil {
		return domain.Cycle{}, mapErr(err)
	}
	return toCycle(rec), nil
}

func (r *CycleRepository) Delete(ctx context.Context, id domain.CycleID) error {
	rec, err := r.app.FindRecordById(ColCycles, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
