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

const assigneeHelp = "user id, or me"

func resolveMe(folio *client.Client, assignee *string) error {
	if assignee == nil || *assignee != "me" {
		return nil
	}
	id, err := folio.Me()
	if err != nil {
		return err
	}
	*assignee = id
	return nil
}

// onTicket tries the reference as an id first, so an id costs one call, and
// resolves it as a slug only when that 404s and a project is selected. The 404
// of a write has no side effect, so the retry is safe.
func onTicket[T any](folio *client.Client, ref string, call func(id string) (T, error)) (T, error) {
	result, err := call(ref)
	project := config.Project(flagProject)
	if project == "" || !client.IsNotFound(err) {
		return result, err
	}

	ticket, lookupErr := folio.GetTicketBySlug(project, ref)
	if lookupErr != nil {
		return result, err
	}
	return call(ticket.ID)
}
