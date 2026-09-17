package services

import (
	"context"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type WorkLogService struct {
	ticketLogs ports.TicketLogRepository
	planLogs   ports.PlanLogRepository
	todoLogs   ports.TodoLogRepository
	issues    ports.IssueRepository
	plans      ports.PlanRepository
	issueRepo      ports.IssueRepository
	cycles     ports.CycleRepository
	guard      ports.Guard
	clock      ports.Clock
	ids        ports.IDGenerator
	log        ports.Logger
}

func NewWorkLogService(ticketLogs ports.TicketLogRepository, planLogs ports.PlanLogRepository, todoLogs ports.TodoLogRepository, issues ports.IssueRepository, plans ports.PlanRepository, cycles ports.CycleRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *WorkLogService {
	return &WorkLogService{
		ticketLogs: ticketLogs, planLogs: planLogs, todoLogs: todoLogs,
		issues: issues, plans: plans, cycles: cycles,
		guard: guard, clock: clock, ids: ids, log: log,
	}
}

var _ ports.WorkLogUseCase = (*WorkLogService)(nil)

func (s *WorkLogService) ListTicketLogs(ctx context.Context, actor ports.Actor, ticket domain.IssueID, f domain.TicketLogFilter) ([]domain.TicketLog, error) {
	owner, err := s.issues.GetByID(ctx, ticket)
	if err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, owner.ProjectID); err != nil {
		return nil, err
	}
	f.IssueID = ticket
	f.Limit = clampLimit(f.Limit)
	entries, err := s.ticketLogs.List(ctx, owner.ProjectID, f)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []domain.TicketLog{}
	}
	return entries, nil
}

func (s *WorkLogService) WriteTicketLog(ctx context.Context, actor ports.Actor, ticket domain.IssueID, body string) (domain.TicketLog, error) {
	owner, err := s.issues.GetByID(ctx, ticket)
	if err != nil {
		return domain.TicketLog{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, owner.ProjectID); err != nil {
		return domain.TicketLog{}, err
	}

	cycles, err := s.cycles.ListByIssue(ctx, ticket)
	if err != nil {
		return domain.TicketLog{}, err
	}
	var cycleID domain.CycleID
	if current, ok := rules.CurrentCycle(cycles); ok && !current.IsClosed() {
		cycleID = current.ID
	}

	now := s.clock.Now()
	entry := domain.TicketLog{
		ID:        domain.TicketLogID(s.ids.NewID()),
		ProjectID: owner.ProjectID,
		IssueID:  ticket,
		CycleID:   cycleID,
		Body:      strings.TrimSpace(body),
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidateTicketLog(entry); err != nil {
		return domain.TicketLog{}, err
	}
	return s.ticketLogs.Create(ctx, entry)
}

func (s *WorkLogService) DeleteTicketLog(ctx context.Context, actor ports.Actor, id domain.TicketLogID) error {
	entry, err := s.ticketLogs.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, entry.ProjectID); err != nil {
		return err
	}
	return s.ticketLogs.Delete(ctx, id)
}

func (s *WorkLogService) ListPlanLogs(ctx context.Context, actor ports.Actor, plan domain.PlanID, f domain.PlanLogFilter) ([]domain.PlanLog, error) {
	owner, err := s.plans.GetByID(ctx, plan)
	if err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, owner.ProjectID); err != nil {
		return nil, err
	}
	f.PlanID = plan
	f.Limit = clampLimit(f.Limit)
	entries, err := s.planLogs.List(ctx, owner.ProjectID, f)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []domain.PlanLog{}
	}
	return entries, nil
}

func (s *WorkLogService) WritePlanLog(ctx context.Context, actor ports.Actor, plan domain.PlanID, body string) (domain.PlanLog, error) {
	owner, err := s.plans.GetByID(ctx, plan)
	if err != nil {
		return domain.PlanLog{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, owner.ProjectID); err != nil {
		return domain.PlanLog{}, err
	}

	now := s.clock.Now()
	entry := domain.PlanLog{
		ID:        domain.PlanLogID(s.ids.NewID()),
		ProjectID: owner.ProjectID,
		PlanID:    plan,
		Body:      strings.TrimSpace(body),
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidatePlanLog(entry); err != nil {
		return domain.PlanLog{}, err
	}
	return s.planLogs.Create(ctx, entry)
}

func (s *WorkLogService) DeletePlanLog(ctx context.Context, actor ports.Actor, id domain.PlanLogID) error {
	entry, err := s.planLogs.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, entry.ProjectID); err != nil {
		return err
	}
	return s.planLogs.Delete(ctx, id)
}

func (s *WorkLogService) ListTodoLogs(ctx context.Context, actor ports.Actor, todo domain.IssueID, f domain.TodoLogFilter) ([]domain.TodoLog, error) {
	owner, err := s.issues.GetByID(ctx, todo)
	if err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, owner.ProjectID); err != nil {
		return nil, err
	}
	f.IssueID = todo
	f.Limit = clampLimit(f.Limit)
	entries, err := s.todoLogs.List(ctx, owner.ProjectID, f)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []domain.TodoLog{}
	}
	return entries, nil
}

func (s *WorkLogService) WriteTodoLog(ctx context.Context, actor ports.Actor, todo domain.IssueID, body string) (domain.TodoLog, error) {
	owner, err := s.issues.GetByID(ctx, todo)
	if err != nil {
		return domain.TodoLog{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, owner.ProjectID); err != nil {
		return domain.TodoLog{}, err
	}

	now := s.clock.Now()
	entry := domain.TodoLog{
		ID:        domain.TodoLogID(s.ids.NewID()),
		ProjectID: owner.ProjectID,
		IssueID:    todo,
		Body:      strings.TrimSpace(body),
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidateTodoLog(entry); err != nil {
		return domain.TodoLog{}, err
	}
	return s.todoLogs.Create(ctx, entry)
}

func (s *WorkLogService) DeleteTodoLog(ctx context.Context, actor ports.Actor, id domain.TodoLogID) error {
	entry, err := s.todoLogs.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, entry.ProjectID); err != nil {
		return err
	}
	return s.todoLogs.Delete(ctx, id)
}
