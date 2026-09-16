package services

import (
	"context"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type PlanService struct {
	repo    ports.PlanRepository
	todos   ports.TodoRepository
	tickets ports.TicketRepository
	// todoUC creates the todos nested in a CreatePlan call, so their defaults
	// and validation stay in one place.
	todoUC ports.TodoUseCase
	guard  ports.Guard
	clock  ports.Clock
	ids    ports.IDGenerator
	log    ports.Logger
}

func NewPlanService(repo ports.PlanRepository, todos ports.TodoRepository, tickets ports.TicketRepository, todoUC ports.TodoUseCase, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *PlanService {
	return &PlanService{repo: repo, todos: todos, tickets: tickets, todoUC: todoUC, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.PlanUseCase = (*PlanService)(nil)

func (s *PlanService) ListPlans(ctx context.Context, actor ports.Actor, project domain.ProjectID, statuses []domain.PlanStatus) ([]domain.Plan, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	plans, err := s.repo.List(ctx, project, statuses)
	if err != nil {
		return nil, err
	}
	for i := range plans {
		if plans[i].Progress, err = s.progress(ctx, plans[i].ID); err != nil {
			return nil, err
		}
	}
	return plans, nil
}

func (s *PlanService) GetPlan(ctx context.Context, actor ports.Actor, id domain.PlanID) (domain.Plan, error) {
	plan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Plan{}, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, plan.ProjectID); err != nil {
		return domain.Plan{}, err
	}
	if plan.Progress, err = s.progress(ctx, plan.ID); err != nil {
		return domain.Plan{}, err
	}
	return plan, nil
}

func (s *PlanService) progress(ctx context.Context, id domain.PlanID) (domain.Progress, error) {
	todos, err := s.todos.ListByPlan(ctx, id)
	if err != nil {
		return domain.Progress{}, err
	}
	return rules.ProgressOf(todos), nil
}

// CreatePlan writes the plan and any todos supplied with it. The plan is
// persisted first so the todos have something to attach to; a failure part-way
// leaves a plan with fewer todos than asked for, which the returned progress
// makes visible.
func (s *PlanService) CreatePlan(ctx context.Context, actor ports.Actor, in ports.CreatePlanInput) (domain.Plan, error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, in.ProjectID); err != nil {
		return domain.Plan{}, err
	}

	if err := ticketScope(ctx, s.tickets, in.TicketID, in.ProjectID); err != nil {
		return domain.Plan{}, err
	}

	now := s.clock.Now()
	plan := domain.Plan{
		ID:        domain.PlanID(s.ids.NewID()),
		ProjectID: in.ProjectID,
		TicketID:  in.TicketID,
		Title:     strings.TrimSpace(in.Title),
		Goal:      in.Goal,
		Status:    defaultPlanStatus(in.Status),
		Tags:      in.Tags,
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidatePlan(plan); err != nil {
		return domain.Plan{}, err
	}
	created, err := s.repo.Create(ctx, plan)
	if err != nil {
		return domain.Plan{}, err
	}

	if len(in.Todos) > 0 {
		nested := make([]ports.CreateTodoInput, len(in.Todos))
		for i, t := range in.Todos {
			t.ProjectID = created.ProjectID
			t.PlanID = created.ID
			// A plan's todos inherit its ticket, so a ticket's progress
			// counts work planned under it.
			t.TicketID = created.TicketID
			nested[i] = t
		}
		batch, err := s.todoUC.CreateTodos(ctx, actor, created.ProjectID, nested)
		if err != nil {
			return domain.Plan{}, err
		}
		for _, e := range batch.Errors {
			s.log.Warn("plan todo rejected", map[string]any{"plan": string(created.ID), "title": e.Title, "reason": e.Reason})
		}
	}

	if created.Progress, err = s.progress(ctx, created.ID); err != nil {
		return domain.Plan{}, err
	}
	return created, nil
}

func (s *PlanService) UpdatePlan(ctx context.Context, actor ports.Actor, id domain.PlanID, in ports.UpdatePlanInput) (domain.Plan, error) {
	plan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Plan{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, plan.ProjectID); err != nil {
		return domain.Plan{}, err
	}

	if in.TicketID != nil {
		if err := ticketScope(ctx, s.tickets, *in.TicketID, plan.ProjectID); err != nil {
			return domain.Plan{}, err
		}
		plan.TicketID = *in.TicketID
	}
	if in.Title != nil {
		plan.Title = strings.TrimSpace(*in.Title)
	}
	if in.Goal != nil {
		plan.Goal = *in.Goal
	}
	if in.Status != nil {
		plan.Status = *in.Status
	}
	if in.Tags != nil {
		plan.Tags = *in.Tags
	}
	if err := rules.ValidatePlan(plan); err != nil {
		return domain.Plan{}, err
	}
	plan.UpdatedAt = s.clock.Now()
	saved, err := s.repo.Update(ctx, plan)
	if err != nil {
		return domain.Plan{}, err
	}
	if saved.Progress, err = s.progress(ctx, saved.ID); err != nil {
		return domain.Plan{}, err
	}
	return saved, nil
}

// DeletePlan detaches its todos rather than deleting them: the plan is the
// agent's framing of the work, and dropping it should not silently destroy
// the work itself.
func (s *PlanService) DeletePlan(ctx context.Context, actor ports.Actor, id domain.PlanID) error {
	plan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, plan.ProjectID); err != nil {
		return err
	}

	todos, err := s.todos.ListByPlan(ctx, id)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	for _, t := range todos {
		t.PlanID = ""
		t.UpdatedAt = now
		if _, err := s.todos.Update(ctx, t); err != nil {
			return err
		}
	}
	return s.repo.Delete(ctx, id)
}

func defaultPlanStatus(s domain.PlanStatus) domain.PlanStatus {
	if s == "" {
		return domain.PlanDraft
	}
	return s
}
