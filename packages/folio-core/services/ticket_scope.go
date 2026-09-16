package services

import (
	"context"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

// ticketScope resolves the TicketID a child entity is being attached to. The
// project guard alone is not enough: a caller with write access to project A
// could otherwise name a ticket in project B and pull the child across the
// tenancy boundary. An empty id means "no ticket", which is always legal.
func ticketScope(ctx context.Context, tickets ports.TicketRepository, id domain.TicketID, project domain.ProjectID) error {
	if id == "" {
		return nil
	}
	ticket, err := tickets.GetByID(ctx, id)
	if err != nil {
		if notFound(err) {
			return domain.Invalid("ticket", "does not exist")
		}
		return err
	}
	return rules.TicketBelongsTo(ticket, project)
}
