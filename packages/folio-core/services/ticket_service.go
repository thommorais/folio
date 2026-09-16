package services

import (
	"context"
	"sort"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

// TicketService manages tickets: units of work under a project that carry
// their own plans, todos, logs and docs.
type TicketService struct {
	repo    ports.TicketRepository
	todos   ports.TodoRepository
	plans   ports.PlanRepository
	journal ports.JournalRepository
	docs    ports.DocRepository
	cycles  ports.CycleRepository
	guard   ports.Guard
	clock   ports.Clock
	ids     ports.IDGenerator
	log     ports.Logger
}

func NewTicketService(repo ports.TicketRepository, todos ports.TodoRepository, plans ports.PlanRepository, journal ports.JournalRepository, docs ports.DocRepository, cycles ports.CycleRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *TicketService {
	return &TicketService{repo: repo, todos: todos, plans: plans, journal: journal, docs: docs, cycles: cycles, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.TicketUseCase = (*TicketService)(nil)

func (s *TicketService) ListTickets(ctx context.Context, actor ports.Actor, project domain.ProjectID, f domain.TicketFilter) ([]domain.Ticket, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	f.Limit = clampLimit(f.Limit)
	tickets, err := s.repo.List(ctx, project, f)
	if err != nil {
		return nil, err
	}
	for i := range tickets {
		if tickets[i].Progress, err = s.progress(ctx, tickets[i].ID); err != nil {
			return nil, err
		}
	}
	if tickets == nil {
		tickets = []domain.Ticket{}
	}
	if err := s.markBlocked(ctx, project, tickets); err != nil {
		return nil, err
	}
	return tickets, nil
}

func (s *TicketService) withCurrentCycle(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	cycles, err := s.cycles.ListByTicket(ctx, ticket.ID)
	if err != nil {
		return domain.Ticket{}, err
	}
	if current, ok := rules.CurrentCycle(cycles); ok {
		ticket.Cycle, ticket.Phase = current.Ordinal, current.Phase
	}
	return ticket, nil
}

func (s *TicketService) markBlocked(ctx context.Context, project domain.ProjectID, tickets []domain.Ticket) error {
	wanted := false
	for _, t := range tickets {
		if len(t.DependsOn) > 0 {
			wanted = true
			break
		}
	}
	if !wanted {
		return nil
	}
	siblings, err := s.repo.List(ctx, project, domain.TicketFilter{Limit: MaxPageSize})
	if err != nil {
		return err
	}
	rules.ApplyTicketBlocked(siblings)
	blocked := make(map[domain.TicketID]bool, len(siblings))
	for _, sib := range siblings {
		blocked[sib.ID] = sib.Blocked
	}
	for i := range tickets {
		tickets[i].Blocked = blocked[tickets[i].ID]
	}
	return nil
}

func (s *TicketService) GetTicket(ctx context.Context, actor ports.Actor, id domain.TicketID) (domain.Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, ticket.ProjectID); err != nil {
		return domain.Ticket{}, err
	}
	if ticket.Progress, err = s.progress(ctx, ticket.ID); err != nil {
		return domain.Ticket{}, err
	}
	ticket, err = s.withTicketBlocked(ctx, ticket)
	if err != nil {
		return domain.Ticket{}, err
	}
	return s.withCurrentCycle(ctx, ticket)
}

func (s *TicketService) GetTicketBySlug(ctx context.Context, actor ports.Actor, project domain.ProjectID, slug string) (domain.Ticket, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return domain.Ticket{}, err
	}
	ticket, err := s.repo.GetBySlug(ctx, project, slug)
	if err != nil {
		return domain.Ticket{}, err
	}
	if ticket.Progress, err = s.progress(ctx, ticket.ID); err != nil {
		return domain.Ticket{}, err
	}
	ticket, err = s.withTicketBlocked(ctx, ticket)
	if err != nil {
		return domain.Ticket{}, err
	}
	return s.withCurrentCycle(ctx, ticket)
}

// progress counts the ticket's todos, including those nested under its plans,
// because both hang off the ticket by TicketID.
func (s *TicketService) progress(ctx context.Context, id domain.TicketID) (domain.Progress, error) {
	todos, err := s.todos.ListByTicket(ctx, id)
	if err != nil {
		return domain.Progress{}, err
	}
	return rules.ProgressOf(todos), nil
}

func (s *TicketService) CreateTicket(ctx context.Context, actor ports.Actor, in ports.CreateTicketInput) (domain.Ticket, error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, in.ProjectID); err != nil {
		return domain.Ticket{}, err
	}

	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = rules.Slugify(in.Title)
	}
	now := s.clock.Now()
	ticket := domain.Ticket{
		ID:          domain.TicketID(s.ids.NewID()),
		ProjectID:   in.ProjectID,
		ParentID:    in.ParentID,
		Slug:        slug,
		Title:       strings.TrimSpace(in.Title),
		Body:        in.Body,
		Status:      defaultTicketStatus(in.Status),
		Priority:    defaultPriority(in.Priority),
		Assignee:    in.Assignee,
		Tags:        in.Tags,
		ExternalRef: in.ExternalRef,
		DependsOn:   in.DependsOn,
		Wayfinder:   in.Wayfinder,
		CreatedBy:   actor.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := rules.ValidateTicket(ticket); err != nil {
		return domain.Ticket{}, err
	}
	if err := s.checkGraph(ctx, ticket); err != nil {
		return domain.Ticket{}, err
	}
	if err := s.slugFree(ctx, ticket.ProjectID, ticket.Slug, ""); err != nil {
		return domain.Ticket{}, err
	}
	created, err := s.repo.Create(ctx, ticket)
	if err != nil {
		return domain.Ticket{}, err
	}
	return s.withTicketBlocked(ctx, created)
}

func (s *TicketService) checkGraph(ctx context.Context, ticket domain.Ticket) error {
	if ticket.ParentID != "" {
		if err := ticketScope(ctx, s.repo, ticket.ParentID, ticket.ProjectID); err != nil {
			return err
		}
	}
	for _, dep := range ticket.DependsOn {
		if err := ticketScope(ctx, s.repo, dep, ticket.ProjectID); err != nil {
			return err
		}
	}
	if ticket.ParentID == "" && len(ticket.DependsOn) == 0 {
		return nil
	}
	siblings, err := s.repo.List(ctx, ticket.ProjectID, domain.TicketFilter{Limit: MaxPageSize})
	if err != nil {
		return err
	}
	if err := rules.CheckNoTicketAncestry(ticket, siblings); err != nil {
		return err
	}
	return rules.CheckNoTicketCycle(ticket, siblings)
}

func (s *TicketService) withTicketBlocked(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	if len(ticket.DependsOn) == 0 {
		ticket.Blocked = false
		return ticket, nil
	}
	siblings, err := s.repo.List(ctx, ticket.ProjectID, domain.TicketFilter{Limit: MaxPageSize})
	if err != nil {
		return domain.Ticket{}, err
	}
	rules.ApplyTicketBlocked(siblings)
	for _, sib := range siblings {
		if sib.ID == ticket.ID {
			ticket.Blocked = sib.Blocked
			break
		}
	}
	return ticket, nil
}

func (s *TicketService) Frontier(ctx context.Context, actor ports.Actor, mapID domain.TicketID) ([]domain.Ticket, error) {
	parent, err := s.repo.GetByID(ctx, mapID)
	if err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, parent.ProjectID); err != nil {
		return nil, err
	}
	children, err := s.repo.ListByParent(ctx, mapID)
	if err != nil {
		return nil, err
	}
	siblings, err := s.repo.List(ctx, parent.ProjectID, domain.TicketFilter{Limit: MaxPageSize})
	if err != nil {
		return nil, err
	}
	rules.ApplyTicketBlocked(siblings)
	blocked := make(map[domain.TicketID]bool, len(siblings))
	for _, sib := range siblings {
		blocked[sib.ID] = sib.Blocked
	}

	out := make([]domain.Ticket, 0, len(children))
	for _, child := range children {
		if child.Status.IsTerminal() || child.Assignee != "" || blocked[child.ID] {
			continue
		}
		child.Blocked = false
		out = append(out, child)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

// slugFree rejects a slug already used in the project. except is the ID of the
// ticket being updated, so keeping its own slug is not a collision with
// itself.
func (s *TicketService) slugFree(ctx context.Context, project domain.ProjectID, slug string, except domain.TicketID) error {
	existing, err := s.repo.GetBySlug(ctx, project, slug)
	if err != nil {
		if notFound(err) {
			return nil
		}
		return err
	}
	if existing.ID == except {
		return nil
	}
	return domain.ErrConflict
}

func (s *TicketService) UpdateTicket(ctx context.Context, actor ports.Actor, id domain.TicketID, in ports.UpdateTicketInput) (domain.Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, ticket.ProjectID); err != nil {
		return domain.Ticket{}, err
	}

	if in.Status != nil {
		if err := rules.CanTransitionTicket(ticket.Status, *in.Status); err != nil {
			return domain.Ticket{}, err
		}
		if in.Status.IsTerminal() {
			cycles, err := s.cycles.ListByTicket(ctx, ticket.ID)
			if err != nil {
				return domain.Ticket{}, err
			}
			if err := rules.CheckClosable(*in.Status, cycles); err != nil {
				return domain.Ticket{}, err
			}
		}
		ticket.Status = *in.Status
	}
	if in.ParentID != nil {
		ticket.ParentID = *in.ParentID
	}
	if in.DependsOn != nil {
		ticket.DependsOn = *in.DependsOn
	}
	if in.Wayfinder != nil {
		ticket.Wayfinder = *in.Wayfinder
	}
	if in.Slug != nil {
		ticket.Slug = strings.TrimSpace(*in.Slug)
	}
	if in.Title != nil {
		ticket.Title = strings.TrimSpace(*in.Title)
	}
	if in.Body != nil {
		ticket.Body = *in.Body
	}
	if in.Priority != nil {
		ticket.Priority = *in.Priority
	}
	if in.Assignee != nil {
		ticket.Assignee = *in.Assignee
	}
	if in.Tags != nil {
		ticket.Tags = *in.Tags
	}
	if in.ExternalRef != nil {
		ticket.ExternalRef = *in.ExternalRef
	}
	if err := rules.ValidateTicket(ticket); err != nil {
		return domain.Ticket{}, err
	}
	if in.ParentID != nil || in.DependsOn != nil {
		if err := s.checkGraph(ctx, ticket); err != nil {
			return domain.Ticket{}, err
		}
	}
	if in.Slug != nil {
		if err := s.slugFree(ctx, ticket.ProjectID, ticket.Slug, ticket.ID); err != nil {
			return domain.Ticket{}, err
		}
	}
	ticket.UpdatedAt = s.clock.Now()
	saved, err := s.repo.Update(ctx, ticket)
	if err != nil {
		return domain.Ticket{}, err
	}
	if saved.Progress, err = s.progress(ctx, saved.ID); err != nil {
		return domain.Ticket{}, err
	}
	saved, err = s.withTicketBlocked(ctx, saved)
	if err != nil {
		return domain.Ticket{}, err
	}
	return s.withCurrentCycle(ctx, saved)
}

func (s *TicketService) SetTicketStatus(ctx context.Context, actor ports.Actor, id domain.TicketID, status domain.TicketStatus) (domain.Ticket, error) {
	return s.UpdateTicket(ctx, actor, id, ports.UpdateTicketInput{Status: &status})
}

// DeleteTicket detaches its plans, todos, logs and docs rather than deleting
// them, for the same reason DeletePlan does: the ticket is a framing of the
// work, and dropping it should not destroy the work itself. The children fall
// back to the project they already belong to.
func (s *TicketService) DeleteTicket(ctx context.Context, actor ports.Actor, id domain.TicketID) error {
	ticket, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, ticket.ProjectID); err != nil {
		return err
	}
	if err := s.detachChildren(ctx, ticket); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *TicketService) detachChildren(ctx context.Context, ticket domain.Ticket) error {
	now := s.clock.Now()

	todos, err := s.todos.ListByTicket(ctx, ticket.ID)
	if err != nil {
		return err
	}
	for _, t := range todos {
		t.TicketID = ""
		t.UpdatedAt = now
		if _, err := s.todos.Update(ctx, t); err != nil {
			return err
		}
	}

	plans, err := s.plans.ListByTicket(ctx, ticket.ID)
	if err != nil {
		return err
	}
	for _, p := range plans {
		p.TicketID = ""
		p.UpdatedAt = now
		if _, err := s.plans.Update(ctx, p); err != nil {
			return err
		}
	}

	entries, err := s.journal.List(ctx, ticket.ProjectID, domain.JournalFilter{TicketID: ticket.ID, Limit: MaxPageSize})
	if err != nil {
		return err
	}
	for _, e := range entries {
		e.TicketID = ""
		e.UpdatedAt = now
		if _, err := s.journal.Update(ctx, e); err != nil {
			return err
		}
	}

	docs, err := s.docs.List(ctx, ticket.ProjectID, domain.DocFilter{TicketID: ticket.ID, Limit: MaxPageSize})
	if err != nil {
		return err
	}
	for _, d := range docs {
		d.TicketID = ""
		d.UpdatedAt = now
		if _, err := s.docs.Update(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

func defaultTicketStatus(s domain.TicketStatus) domain.TicketStatus {
	if s == "" {
		return domain.TicketOpen
	}
	return s
}
