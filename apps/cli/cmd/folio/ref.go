package main

import (
	"folio/cli/internal/client"
	"folio/cli/internal/config"
)

func bySlugOrID[T any](ref string, bySlug func(project, slug string) (T, error), byID func(id string) (T, error)) (T, error) {
	project := config.Project(flagProject)
	if project == "" {
		return byID(ref)
	}
	found, err := bySlug(project, ref)
	if client.IsNotFound(err) {
		return byID(ref)
	}
	return found, err
}

func ticketID(folio *client.Client, ref string) (string, error) {
	if config.Project(flagProject) == "" {
		return ref, nil
	}
	ticket, err := bySlugOrID(ref, folio.GetTicketBySlug, folio.GetTicket)
	return ticket.ID, err
}
