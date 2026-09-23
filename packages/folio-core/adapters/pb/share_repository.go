package pb

import (
	"context"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type ShareRepository struct {
	app core.App
}

func NewShareRepository(app core.App) *ShareRepository {
	return &ShareRepository{app: app}
}

var _ ports.ShareRepository = (*ShareRepository)(nil)

func toShare(rec *core.Record) domain.Share {
	return domain.Share{
		ID:             domain.ShareID(rec.Id),
		ProjectID:      domain.ProjectID(rec.GetString("project")),
		IssueID:        domain.IssueID(rec.GetString("issue")),
		PlanID:         domain.PlanID(rec.GetString("plan")),
		Label:          rec.GetString("label"),
		Token:          rec.GetString("token"),
		CreatedBy:      domain.UserID(rec.GetString("created_by")),
		CreatedAt:      rec.GetDateTime("created").Time(),
		LastAccessedAt: timePtr(rec.GetDateTime("last_accessed_at")),
	}
}

func (r *ShareRepository) List(ctx context.Context, project domain.ProjectID) ([]domain.Share, error) {
	records, err := r.app.FindRecordsByFilter(ColShares, "project = {:project}", "-created", 0, 0, dbx.Params{"project": string(project)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Share, 0, len(records))
	for _, rec := range records {
		out = append(out, toShare(rec))
	}
	return out, nil
}

func (r *ShareRepository) GetByID(ctx context.Context, id domain.ShareID) (domain.Share, error) {
	rec, err := r.app.FindRecordById(ColShares, string(id))
	if err != nil {
		return domain.Share{}, mapErr(err)
	}
	return toShare(rec), nil
}

func (r *ShareRepository) GetByToken(ctx context.Context, token string) (domain.Share, error) {
	rec, err := r.app.FindFirstRecordByData(ColShares, "token", token)
	if err != nil {
		return domain.Share{}, mapErr(err)
	}
	return toShare(rec), nil
}

func (r *ShareRepository) Create(ctx context.Context, s domain.Share) (domain.Share, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColShares)
	if err != nil {
		return domain.Share{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(s.ID)
	rec.Set("project", string(s.ProjectID))
	rec.Set("issue", string(s.IssueID))
	rec.Set("plan", string(s.PlanID))
	rec.Set("label", s.Label)
	rec.Set("token", s.Token)
	rec.Set("created_by", string(s.CreatedBy))
	if err := r.app.Save(rec); err != nil {
		return domain.Share{}, mapErr(err)
	}
	return toShare(rec), nil
}

func (r *ShareRepository) Delete(ctx context.Context, id domain.ShareID) error {
	rec, err := r.app.FindRecordById(ColShares, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}

func (r *ShareRepository) Touch(ctx context.Context, id domain.ShareID, at time.Time) error {
	rec, err := r.app.FindRecordById(ColShares, string(id))
	if err != nil {
		return mapErr(err)
	}
	setDate(rec, "last_accessed_at", &at)
	return mapErr(r.app.Save(rec))
}
