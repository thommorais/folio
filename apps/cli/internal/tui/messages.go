package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"folio/cli/internal/client"
)

type projectsLoaded struct {
	projects []client.Project
	err      error
}

type projectLoaded struct {
	project client.Project
	err     error
}

type tabLoaded struct {
	tab  tab
	rows rows
	err  error
}

func loadProjects(api *client.Client) tea.Cmd {
	return func() tea.Msg {
		projects, err := api.ListProjects(false)
		return projectsLoaded{projects: projects, err: err}
	}
}

func loadProject(api *client.Client, ref string) tea.Cmd {
	return func() tea.Msg {
		project, err := api.GetProject(ref)
		return projectLoaded{project: project, err: err}
	}
}

// One tab loads at a time: the other three are only needed once visited, and
// fetching all four on entry makes the first paint wait on the slowest.
func loadTab(api *client.Client, which tab, project string) tea.Cmd {
	return func() tea.Msg {
		switch which {
		case tabTickets:
			tickets, err := api.ListTickets(project, client.TicketFilter{})
			return tabLoaded{tab: which, rows: rows{tickets: tickets}, err: err}
		case tabPlans:
			plans, err := api.ListPlans(project, client.PlanFilter{})
			return tabLoaded{tab: which, rows: rows{plans: plans}, err: err}
		case tabTodos:
			todos, err := api.ListTodos(project, client.TodoFilter{})
			return tabLoaded{tab: which, rows: rows{todos: todos}, err: err}
		case tabLogs:
			logs, err := api.ListJournal(project, client.JournalFilter{})
			return tabLoaded{tab: which, rows: rows{logs: logs}, err: err}
		case tabDocs:
			docs, err := api.ListDocs(project, client.DocFilter{})
			return tabLoaded{tab: which, rows: rows{docs: docs}, err: err}
		}
		return tabLoaded{tab: which}
	}
}
