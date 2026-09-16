package pb

import (
	"context"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type TodoRepository struct {
	app core.App
}

func NewTodoRepository(app core.App) *TodoRepository {
	return &TodoRepository{app: app}
}

var _ ports.TodoRepository = (*TodoRepository)(nil)

func toTodo(rec *core.Record) domain.Todo {
	return domain.Todo{
		ID:        domain.TodoID(rec.Id),
		ProjectID: domain.ProjectID(rec.GetString("project")),
		TicketID:  domain.TicketID(rec.GetString("ticket")),
		PlanID:    domain.PlanID(rec.GetString("plan")),
		Title:     rec.GetString("title"),
		Details:   rec.GetString("details"),
		Status:    domain.TodoStatus(rec.GetString("status")),
		Priority:  domain.Priority(rec.GetString("priority")),
		Tags:      strSlice(rec, "tags"),
		Position:  rec.GetInt("position"),
		DependsOn: toTodoIDs(strSlice(rec, "depends_on")),
		DueDate:   timePtr(rec.GetDateTime("due_date")),
		CreatedBy: domain.UserID(rec.GetString("created_by")),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}
}

func (r *TodoRepository) List(ctx context.Context, project domain.ProjectID, f domain.TodoFilter) ([]domain.Todo, error) {
	exprs := []dbx.Expression{dbx.HashExp{"project": string(project)}}
	if f.PlanID != "" {
		exprs = append(exprs, dbx.HashExp{"plan": string(f.PlanID)})
	}
	if f.TicketID != "" {
		exprs = append(exprs, dbx.HashExp{"ticket": string(f.TicketID)})
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
	if q := strings.TrimSpace(f.Search); q != "" {
		exprs = append(exprs, dbx.Or(
			dbx.Like("title", q),
			dbx.Like("details", q),
		))
	}
	for _, tag := range f.Tags {
		// Tags are a JSON array; a LIKE on the encoded form is enough to
		// filter, and the service re-checks nothing because storage is the
		// only place that knows the encoding.
		exprs = append(exprs, dbx.Like("tags", `"`+tag+`"`))
	}

	records, err := r.app.FindAllRecords(ColTodos, exprs...)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Todo, 0, len(records))
	for _, rec := range records {
		out = append(out, toTodo(rec))
	}
	return applyPaging(out, f.Offset, f.Limit), nil
}

// applyPaging trims an already-loaded slice. PocketBase's FindAllRecords has
// no limit argument, so paging happens here; the service clamps the limit so
// the slice stays bounded.
func applyPaging[T any](items []T, offset, limit int) []T {
	if offset > 0 {
		if offset >= len(items) {
			return items[:0]
		}
		items = items[offset:]
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items
}

func (r *TodoRepository) ListByPlan(ctx context.Context, plan domain.PlanID) ([]domain.Todo, error) {
	records, err := r.app.FindAllRecords(ColTodos, dbx.HashExp{"plan": string(plan)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Todo, 0, len(records))
	for _, rec := range records {
		out = append(out, toTodo(rec))
	}
	return out, nil
}

// ListByTicket returns every todo filed under a ticket, including those that
// also sit under one of its plans.
func (r *TodoRepository) ListByTicket(ctx context.Context, ticket domain.TicketID) ([]domain.Todo, error) {
	records, err := r.app.FindAllRecords(ColTodos, dbx.HashExp{"ticket": string(ticket)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Todo, 0, len(records))
	for _, rec := range records {
		out = append(out, toTodo(rec))
	}
	return out, nil
}

func (r *TodoRepository) GetByID(ctx context.Context, id domain.TodoID) (domain.Todo, error) {
	rec, err := r.app.FindRecordById(ColTodos, string(id))
	if err != nil {
		return domain.Todo{}, mapErr(err)
	}
	return toTodo(rec), nil
}

func (r *TodoRepository) Create(ctx context.Context, t domain.Todo) (domain.Todo, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColTodos)
	if err != nil {
		return domain.Todo{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	rec.Id = string(t.ID)
	applyTodo(rec, t)
	if err := r.app.Save(rec); err != nil {
		return domain.Todo{}, mapErr(err)
	}
	return toTodo(rec), nil
}

func (r *TodoRepository) Update(ctx context.Context, t domain.Todo) (domain.Todo, error) {
	rec, err := r.app.FindRecordById(ColTodos, string(t.ID))
	if err != nil {
		return domain.Todo{}, mapErr(err)
	}
	applyTodo(rec, t)
	if err := r.app.Save(rec); err != nil {
		return domain.Todo{}, mapErr(err)
	}
	return toTodo(rec), nil
}

func applyTodo(rec *core.Record, t domain.Todo) {
	rec.Set("project", string(t.ProjectID))
	rec.Set("plan", string(t.PlanID))
	rec.Set("ticket", string(t.TicketID))
	rec.Set("title", t.Title)
	rec.Set("details", t.Details)
	rec.Set("status", string(t.Status))
	rec.Set("priority", string(t.Priority))
	rec.Set("position", t.Position)
	setJSON(rec, "tags", t.Tags)
	setJSON(rec, "depends_on", fromTodoIDs(t.DependsOn))
	setDate(rec, "due_date", t.DueDate)
	if t.CreatedBy != "" {
		rec.Set("created_by", string(t.CreatedBy))
	}
}

func (r *TodoRepository) Delete(ctx context.Context, id domain.TodoID) error {
	rec, err := r.app.FindRecordById(ColTodos, string(id))
	if err != nil {
		return mapErr(err)
	}
	return mapErr(r.app.Delete(rec))
}
