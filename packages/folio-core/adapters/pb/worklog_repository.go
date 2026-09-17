package pb

import (
	"context"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type TicketLogRepository struct{ app core.App }
type PlanLogRepository struct{ app core.App }
type TodoLogRepository struct{ app core.App }

func NewTicketLogRepository(app core.App) *TicketLogRepository { return &TicketLogRepository{app: app} }
func NewPlanLogRepository(app core.App) *PlanLogRepository     { return &PlanLogRepository{app: app} }
func NewTodoLogRepository(app core.App) *TodoLogRepository     { return &TodoLogRepository{app: app} }

var (
	_ ports.TicketLogRepository = (*TicketLogRepository)(nil)
	_ ports.PlanLogRepository   = (*PlanLogRepository)(nil)
	_ ports.TodoLogRepository   = (*TodoLogRepository)(nil)
)

func toTicketLog(rec *core.Record) domain.TicketLog {
	return domain.TicketLog{
		ID:        domain.TicketLogID(rec.Id),
		ProjectID: domain.ProjectID(rec.GetString("project")),
		IssueID:  domain.IssueID(rec.GetString("ticket")),
		CycleID:   domain.CycleID(rec.GetString("cycle")),
		Body:      rec.GetString("body"),
		CreatedBy: domain.UserID(rec.GetString("created_by")),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}
}

func applyTicketLog(rec *core.Record, l domain.TicketLog) {
	rec.Set("project", string(l.ProjectID))
	rec.Set("ticket", string(l.IssueID))
	rec.Set("cycle", string(l.CycleID))
	rec.Set("body", l.Body)
	if l.CreatedBy != "" {
		rec.Set("created_by", string(l.CreatedBy))
	}
}

func (r *TicketLogRepository) List(ctx context.Context, project domain.ProjectID, f domain.TicketLogFilter) ([]domain.TicketLog, error) {
	exprs := []dbx.Expression{dbx.HashExp{"project": string(project)}}
	if f.IssueID != "" {
		exprs = append(exprs, dbx.HashExp{"ticket": string(f.IssueID)})
	}
	if f.CycleID != "" {
		exprs = append(exprs, dbx.HashExp{"cycle": string(f.CycleID)})
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		exprs = append(exprs, dbx.Like("body", q))
	}
	records, err := r.app.FindAllRecords(ColTicketLogs, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.TicketLog, 0, len(records))
	for _, rec := range records {
		out = append(out, toTicketLog(rec))
	}
	return applyPaging(out, f.Offset, f.Limit), nil
}

func (r *TicketLogRepository) GetByID(ctx context.Context, id domain.TicketLogID) (domain.TicketLog, error) {
	rec, err := r.app.FindRecordById(ColTicketLogs, string(id))
	if err != nil {
		return domain.TicketLog{}, mapErr(err)
	}
	return toTicketLog(rec), nil
}

func (r *TicketLogRepository) Create(ctx context.Context, l domain.TicketLog) (domain.TicketLog, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColTicketLogs)
	if err != nil {
		return domain.TicketLog{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(l.ID)
	applyTicketLog(rec, l)
	if err := r.app.Save(rec); err != nil {
		return domain.TicketLog{}, mapErr(err)
	}
	return toTicketLog(rec), nil
}

func (r *TicketLogRepository) Update(ctx context.Context, l domain.TicketLog) (domain.TicketLog, error) {
	rec, err := r.app.FindRecordById(ColTicketLogs, string(l.ID))
	if err != nil {
		return domain.TicketLog{}, mapErr(err)
	}
	applyTicketLog(rec, l)
	if err := r.app.Save(rec); err != nil {
		return domain.TicketLog{}, mapErr(err)
	}
	return toTicketLog(rec), nil
}

func (r *TicketLogRepository) Delete(ctx context.Context, id domain.TicketLogID) error {
	rec, err := r.app.FindRecordById(ColTicketLogs, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}

func toPlanLog(rec *core.Record) domain.PlanLog {
	return domain.PlanLog{
		ID:        domain.PlanLogID(rec.Id),
		ProjectID: domain.ProjectID(rec.GetString("project")),
		PlanID:    domain.PlanID(rec.GetString("plan")),
		Body:      rec.GetString("body"),
		CreatedBy: domain.UserID(rec.GetString("created_by")),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}
}

func applyPlanLog(rec *core.Record, l domain.PlanLog) {
	rec.Set("project", string(l.ProjectID))
	rec.Set("plan", string(l.PlanID))
	rec.Set("body", l.Body)
	if l.CreatedBy != "" {
		rec.Set("created_by", string(l.CreatedBy))
	}
}

func (r *PlanLogRepository) List(ctx context.Context, project domain.ProjectID, f domain.PlanLogFilter) ([]domain.PlanLog, error) {
	exprs := []dbx.Expression{dbx.HashExp{"project": string(project)}}
	if f.PlanID != "" {
		exprs = append(exprs, dbx.HashExp{"plan": string(f.PlanID)})
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		exprs = append(exprs, dbx.Like("body", q))
	}
	records, err := r.app.FindAllRecords(ColPlanLogs, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.PlanLog, 0, len(records))
	for _, rec := range records {
		out = append(out, toPlanLog(rec))
	}
	return applyPaging(out, f.Offset, f.Limit), nil
}

func (r *PlanLogRepository) GetByID(ctx context.Context, id domain.PlanLogID) (domain.PlanLog, error) {
	rec, err := r.app.FindRecordById(ColPlanLogs, string(id))
	if err != nil {
		return domain.PlanLog{}, mapErr(err)
	}
	return toPlanLog(rec), nil
}

func (r *PlanLogRepository) Create(ctx context.Context, l domain.PlanLog) (domain.PlanLog, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColPlanLogs)
	if err != nil {
		return domain.PlanLog{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(l.ID)
	applyPlanLog(rec, l)
	if err := r.app.Save(rec); err != nil {
		return domain.PlanLog{}, mapErr(err)
	}
	return toPlanLog(rec), nil
}

func (r *PlanLogRepository) Update(ctx context.Context, l domain.PlanLog) (domain.PlanLog, error) {
	rec, err := r.app.FindRecordById(ColPlanLogs, string(l.ID))
	if err != nil {
		return domain.PlanLog{}, mapErr(err)
	}
	applyPlanLog(rec, l)
	if err := r.app.Save(rec); err != nil {
		return domain.PlanLog{}, mapErr(err)
	}
	return toPlanLog(rec), nil
}

func (r *PlanLogRepository) Delete(ctx context.Context, id domain.PlanLogID) error {
	rec, err := r.app.FindRecordById(ColPlanLogs, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}

func toTodoLog(rec *core.Record) domain.TodoLog {
	return domain.TodoLog{
		ID:        domain.TodoLogID(rec.Id),
		ProjectID: domain.ProjectID(rec.GetString("project")),
		IssueID:    domain.IssueID(rec.GetString("todo")),
		Body:      rec.GetString("body"),
		CreatedBy: domain.UserID(rec.GetString("created_by")),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}
}

func applyTodoLog(rec *core.Record, l domain.TodoLog) {
	rec.Set("project", string(l.ProjectID))
	rec.Set("todo", string(l.IssueID))
	rec.Set("body", l.Body)
	if l.CreatedBy != "" {
		rec.Set("created_by", string(l.CreatedBy))
	}
}

func (r *TodoLogRepository) List(ctx context.Context, project domain.ProjectID, f domain.TodoLogFilter) ([]domain.TodoLog, error) {
	exprs := []dbx.Expression{dbx.HashExp{"project": string(project)}}
	if f.IssueID != "" {
		exprs = append(exprs, dbx.HashExp{"todo": string(f.IssueID)})
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		exprs = append(exprs, dbx.Like("body", q))
	}
	records, err := r.app.FindAllRecords(ColTodoLogs, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.TodoLog, 0, len(records))
	for _, rec := range records {
		out = append(out, toTodoLog(rec))
	}
	return applyPaging(out, f.Offset, f.Limit), nil
}

func (r *TodoLogRepository) GetByID(ctx context.Context, id domain.TodoLogID) (domain.TodoLog, error) {
	rec, err := r.app.FindRecordById(ColTodoLogs, string(id))
	if err != nil {
		return domain.TodoLog{}, mapErr(err)
	}
	return toTodoLog(rec), nil
}

func (r *TodoLogRepository) Create(ctx context.Context, l domain.TodoLog) (domain.TodoLog, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColTodoLogs)
	if err != nil {
		return domain.TodoLog{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(l.ID)
	applyTodoLog(rec, l)
	if err := r.app.Save(rec); err != nil {
		return domain.TodoLog{}, mapErr(err)
	}
	return toTodoLog(rec), nil
}

func (r *TodoLogRepository) Update(ctx context.Context, l domain.TodoLog) (domain.TodoLog, error) {
	rec, err := r.app.FindRecordById(ColTodoLogs, string(l.ID))
	if err != nil {
		return domain.TodoLog{}, mapErr(err)
	}
	applyTodoLog(rec, l)
	if err := r.app.Save(rec); err != nil {
		return domain.TodoLog{}, mapErr(err)
	}
	return toTodoLog(rec), nil
}

func (r *TodoLogRepository) Delete(ctx context.Context, id domain.TodoLogID) error {
	rec, err := r.app.FindRecordById(ColTodoLogs, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
