package services

import (
	"context"
	"sort"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

const DefaultRecentJournal = 10

func (s *TicketService) GetTicketBrief(ctx context.Context, actor ports.Actor, id domain.TicketID, in ports.BriefOptions) (domain.TicketBrief, error) {
	ticket, err := s.GetTicket(ctx, actor, id)
	if err != nil {
		return domain.TicketBrief{}, err
	}
	return s.brief(ctx, ticket, in)
}

func (s *TicketService) GetTicketBriefBySlug(ctx context.Context, actor ports.Actor, project domain.ProjectID, slug string, in ports.BriefOptions) (domain.TicketBrief, error) {
	ticket, err := s.GetTicketBySlug(ctx, actor, project, slug)
	if err != nil {
		return domain.TicketBrief{}, err
	}
	return s.brief(ctx, ticket, in)
}

func (s *TicketService) brief(ctx context.Context, ticket domain.Ticket, in ports.BriefOptions) (domain.TicketBrief, error) {
	recent := in.RecentJournal
	if recent <= 0 {
		recent = DefaultRecentJournal
	}

	plans, err := s.plans.ListByTicket(ctx, ticket.ID)
	if err != nil {
		return domain.TicketBrief{}, err
	}

	todos, err := s.todos.ListByTicket(ctx, ticket.ID)
	if err != nil {
		return domain.TicketBrief{}, err
	}
	sortOpenFirst(todos)

	journal, err := s.journal.List(ctx, ticket.ProjectID, domain.JournalFilter{TicketID: ticket.ID, Limit: recent})
	if err != nil {
		return domain.TicketBrief{}, err
	}

	docs, err := s.docs.List(ctx, ticket.ProjectID, domain.DocFilter{TicketID: ticket.ID, Limit: MaxPageSize})
	if err != nil {
		return domain.TicketBrief{}, err
	}

	cycles, err := s.cycles.ListByTicket(ctx, ticket.ID)
	if err != nil {
		return domain.TicketBrief{}, err
	}
	if cycles == nil {
		cycles = []domain.Cycle{}
	}

	return domain.TicketBrief{
		Ticket:  ticket,
		Plans:   plans,
		Todos:   todos,
		Journal: journal,
		Docs:    docs,
		Cycles:  cycles,
	}, nil
}

func sortOpenFirst(todos []domain.Todo) {
	sort.SliceStable(todos, func(i, j int) bool {
		return !todos[i].Status.IsTerminal() && todos[j].Status.IsTerminal()
	})
}
