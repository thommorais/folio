package services

import (
	"context"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type CycleService struct {
	repo    ports.CycleRepository
	tickets ports.TicketRepository
	guard   ports.Guard
	clock   ports.Clock
	ids     ports.IDGenerator
	log     ports.Logger
}

func NewCycleService(repo ports.CycleRepository, tickets ports.TicketRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *CycleService {
	return &CycleService{repo: repo, tickets: tickets, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.CycleUseCase = (*CycleService)(nil)

func (s *CycleService) ListCycles(ctx context.Context, actor ports.Actor, ticket domain.TicketID) ([]domain.Cycle, error) {
	owner, err := s.tickets.GetByID(ctx, ticket)
	if err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, owner.ProjectID); err != nil {
		return nil, err
	}
	cycles, err := s.repo.ListByTicket(ctx, ticket)
	if err != nil {
		return nil, err
	}
	if cycles == nil {
		cycles = []domain.Cycle{}
	}
	return cycles, nil
}

func (s *CycleService) OpenCycle(ctx context.Context, actor ports.Actor, ticket domain.TicketID) (domain.Cycle, error) {
	owner, err := s.tickets.GetByID(ctx, ticket)
	if err != nil {
		return domain.Cycle{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, owner.ProjectID); err != nil {
		return domain.Cycle{}, err
	}
	existing, err := s.repo.ListByTicket(ctx, ticket)
	if err != nil {
		return domain.Cycle{}, err
	}
	if current, ok := rules.CurrentCycle(existing); ok && !current.IsClosed() {
		return domain.Cycle{}, domain.ErrConflict
	}

	now := s.clock.Now()
	cycle := domain.Cycle{
		ID:        domain.CycleID(s.ids.NewID()),
		ProjectID: owner.ProjectID,
		TicketID:  ticket,
		Ordinal:   rules.NextOrdinal(existing),
		Phase:     domain.PhasePlan,
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidateCycle(cycle); err != nil {
		return domain.Cycle{}, err
	}
	return s.repo.Create(ctx, cycle)
}

func (s *CycleService) AdvancePhase(ctx context.Context, actor ports.Actor, id domain.CycleID, phase domain.Phase) (domain.Cycle, error) {
	cycle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Cycle{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, cycle.ProjectID); err != nil {
		return domain.Cycle{}, err
	}
	if cycle.IsClosed() {
		return domain.Cycle{}, domain.Invalid("phase", "a resolved cycle cannot advance; open the next one")
	}
	if err := rules.CanTransitionPhase(cycle.Phase, phase); err != nil {
		return domain.Cycle{}, err
	}
	cycle.Phase = phase
	cycle.UpdatedAt = s.clock.Now()
	if err := rules.ValidateCycle(cycle); err != nil {
		return domain.Cycle{}, err
	}
	return s.repo.Update(ctx, cycle)
}

func (s *CycleService) ResolveCycle(ctx context.Context, actor ports.Actor, id domain.CycleID, resolution string) (domain.Cycle, error) {
	cycle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Cycle{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, cycle.ProjectID); err != nil {
		return domain.Cycle{}, err
	}
	if strings.TrimSpace(resolution) == "" {
		return domain.Cycle{}, domain.Invalid("resolution", "is required")
	}

	now := s.clock.Now()
	cycle.Resolution = strings.TrimSpace(resolution)
	cycle.UpdatedAt = now
	if !cycle.IsClosed() {
		cycle.ClosedAt = &now
	}
	if err := rules.ValidateCycle(cycle); err != nil {
		return domain.Cycle{}, err
	}
	return s.repo.Update(ctx, cycle)
}
