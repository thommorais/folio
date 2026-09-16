package services

import (
	"context"
	"strings"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type TodoService struct {
	repo    ports.TodoRepository
	plans   ports.PlanRepository
	tickets ports.TicketRepository
	guard   ports.Guard
	clock   ports.Clock
	ids     ports.IDGenerator
	log     ports.Logger
}

func NewTodoService(repo ports.TodoRepository, plans ports.PlanRepository, tickets ports.TicketRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *TodoService {
	return &TodoService{repo: repo, plans: plans, tickets: tickets, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.TodoUseCase = (*TodoService)(nil)

func (s *TodoService) ListTodos(ctx context.Context, actor ports.Actor, project domain.ProjectID, f domain.TodoFilter) ([]domain.Todo, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	todos, err := s.repo.List(ctx, project, f)
	if err != nil {
		return nil, err
	}
	if err := s.markBlocked(ctx, project, todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func (s *TodoService) markBlocked(ctx context.Context, project domain.ProjectID, todos []domain.Todo) error {
	wanted := false
	for _, t := range todos {
		if len(t.DependsOn) > 0 {
			wanted = true
			break
		}
	}
	if !wanted {
		return nil
	}
	siblings, err := s.repo.List(ctx, project, domain.TodoFilter{Limit: MaxPageSize})
	if err != nil {
		return err
	}
	rules.ApplyBlocked(siblings)
	blocked := make(map[domain.TodoID]bool, len(siblings))
	for _, sib := range siblings {
		blocked[sib.ID] = sib.Blocked
	}
	for i := range todos {
		todos[i].Blocked = blocked[todos[i].ID]
	}
	return nil
}

func (s *TodoService) GetTodo(ctx context.Context, actor ports.Actor, id domain.TodoID) (domain.Todo, error) {
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Todo{}, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, todo.ProjectID); err != nil {
		return domain.Todo{}, err
	}
	return s.withBlocked(ctx, todo)
}

// withBlocked resolves the single-todo Blocked flag. It needs the sibling set
// because a dependency's status is what decides it, and a client reading one
// todo should see the same flag the list would show.
func (s *TodoService) withBlocked(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	if len(todo.DependsOn) == 0 {
		todo.Blocked = false
		return todo, nil
	}
	siblings, err := s.repo.List(ctx, todo.ProjectID, domain.TodoFilter{})
	if err != nil {
		return domain.Todo{}, err
	}
	rules.ApplyBlocked(siblings)
	for _, sib := range siblings {
		if sib.ID == todo.ID {
			todo.Blocked = sib.Blocked
			break
		}
	}
	return todo, nil
}

func (s *TodoService) CreateTodo(ctx context.Context, actor ports.Actor, in ports.CreateTodoInput) (domain.Todo, error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, in.ProjectID); err != nil {
		return domain.Todo{}, err
	}
	return s.create(ctx, actor, in, 0)
}

// create builds and persists one todo. position 0 means "append", which costs
// a listing; batch callers pass an explicit position to avoid re-listing per
// item.
func (s *TodoService) create(ctx context.Context, actor ports.Actor, in ports.CreateTodoInput, position int) (domain.Todo, error) {
	if err := s.checkPlan(ctx, in.ProjectID, in.PlanID); err != nil {
		return domain.Todo{}, err
	}
	if err := ticketScope(ctx, s.tickets, in.TicketID, in.ProjectID); err != nil {
		return domain.Todo{}, err
	}
	due, err := parseDue(in.DueDate)
	if err != nil {
		return domain.Todo{}, err
	}
	if position == 0 {
		existing, err := s.repo.List(ctx, in.ProjectID, domain.TodoFilter{PlanID: in.PlanID})
		if err != nil {
			return domain.Todo{}, err
		}
		position = rules.NextPosition(existing)
	}

	now := s.clock.Now()
	todo := domain.Todo{
		ID:        domain.TodoID(s.ids.NewID()),
		ProjectID: in.ProjectID,
		TicketID:  in.TicketID,
		PlanID:    in.PlanID,
		Title:     strings.TrimSpace(in.Title),
		Details:   in.Details,
		Status:    defaultTodoStatus(in.Status),
		Priority:  defaultPriority(in.Priority),
		Tags:      in.Tags,
		Position:  position,
		DependsOn: in.DependsOn,
		DueDate:   due,
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidateTodo(todo); err != nil {
		return domain.Todo{}, err
	}
	created, err := s.repo.Create(ctx, todo)
	if err != nil {
		return domain.Todo{}, err
	}
	return s.withBlocked(ctx, created)
}

// checkPlan refuses to file a todo under a plan belonging to another project,
// which would make the plan's progress count work nobody can see.
func (s *TodoService) checkPlan(ctx context.Context, project domain.ProjectID, plan domain.PlanID) error {
	if plan == "" {
		return nil
	}
	found, err := s.plans.GetByID(ctx, plan)
	if err != nil {
		if notFound(err) {
			return domain.Invalid("plan", "does not exist")
		}
		return err
	}
	if found.ProjectID != project {
		return domain.Invalid("plan", "belongs to a different project")
	}
	return nil
}

func (s *TodoService) CreateTodos(ctx context.Context, actor ports.Actor, project domain.ProjectID, in []ports.CreateTodoInput) (ports.BatchResult[domain.Todo], error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, project); err != nil {
		return ports.BatchResult[domain.Todo]{}, err
	}
	existing, err := s.repo.List(ctx, project, domain.TodoFilter{})
	if err != nil {
		return ports.BatchResult[domain.Todo]{}, err
	}
	next := rules.NextPosition(existing)

	out := ports.BatchResult[domain.Todo]{Created: make([]domain.Todo, 0, len(in))}
	for i, item := range in {
		item.ProjectID = project
		// One bad item must not cost the caller the good ones: a code agent
		// submitting a triaged batch would have to guess which survived.
		created, err := s.create(ctx, actor, item, next)
		if err != nil {
			out.Errors = append(out.Errors, ports.BatchError{Index: i, Title: item.Title, Reason: err.Error()})
			continue
		}
		out.Created = append(out.Created, created)
		next++
	}
	return out, nil
}

func (s *TodoService) UpdateTodo(ctx context.Context, actor ports.Actor, id domain.TodoID, in ports.UpdateTodoInput) (domain.Todo, error) {
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Todo{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, todo.ProjectID); err != nil {
		return domain.Todo{}, err
	}

	if in.TicketID != nil {
		if err := ticketScope(ctx, s.tickets, *in.TicketID, todo.ProjectID); err != nil {
			return domain.Todo{}, err
		}
		todo.TicketID = *in.TicketID
	}
	if in.PlanID != nil {
		if err := s.checkPlan(ctx, todo.ProjectID, *in.PlanID); err != nil {
			return domain.Todo{}, err
		}
		todo.PlanID = *in.PlanID
	}
	if in.Title != nil {
		todo.Title = strings.TrimSpace(*in.Title)
	}
	if in.Details != nil {
		todo.Details = *in.Details
	}
	if in.Status != nil {
		if err := rules.CanTransitionTodo(todo.Status, *in.Status); err != nil {
			return domain.Todo{}, err
		}
		todo.Status = *in.Status
	}
	if in.Priority != nil {
		todo.Priority = *in.Priority
	}
	if in.Tags != nil {
		todo.Tags = *in.Tags
	}
	if in.Position != nil {
		todo.Position = *in.Position
	}
	if in.DependsOn != nil {
		todo.DependsOn = *in.DependsOn
	}
	if in.DueDate != nil {
		due, err := parseDue(in.DueDate)
		if err != nil {
			return domain.Todo{}, err
		}
		todo.DueDate = due
	}

	if err := rules.ValidateTodo(todo); err != nil {
		return domain.Todo{}, err
	}
	todo.UpdatedAt = s.clock.Now()
	saved, err := s.repo.Update(ctx, todo)
	if err != nil {
		return domain.Todo{}, err
	}
	return s.withBlocked(ctx, saved)
}

func (s *TodoService) SetTodoStatus(ctx context.Context, actor ports.Actor, id domain.TodoID, status domain.TodoStatus) (domain.Todo, error) {
	return s.UpdateTodo(ctx, actor, id, ports.UpdateTodoInput{Status: &status})
}

func (s *TodoService) DeleteTodo(ctx context.Context, actor ports.Actor, id domain.TodoID) error {
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, todo.ProjectID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func defaultTodoStatus(s domain.TodoStatus) domain.TodoStatus {
	if s == "" {
		return domain.TodoPending
	}
	return s
}

func defaultPriority(p domain.Priority) domain.Priority {
	if p == "" {
		return domain.PriorityMedium
	}
	return p
}

// parseDue accepts RFC 3339, treating an empty string as "clear the date" and
// nil as "leave it alone".
func parseDue(raw *string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return nil, domain.Invalid("due_date", "must be an RFC 3339 timestamp, e.g. 2026-12-01T09:00:00Z")
	}
	return &t, nil
}
