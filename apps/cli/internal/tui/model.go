package tui

import (
	"folio/cli/internal/client"
)

type screen int

const (
	screenProjects screen = iota
	screenProject
	screenDetail
)

type tab int

const (
	tabTickets tab = iota
	tabPlans
	tabTodos
	tabLogs
	tabDocs
)

var tabNames = map[tab]string{
	tabTickets: "tickets",
	tabPlans:   "plans",
	tabTodos:   "todos",
	tabLogs:    "logs",
	tabDocs:    "docs",
}

const tabCount = 5

type rows struct {
	tickets []client.Ticket
	plans   []client.Plan
	todos   []client.Todo
	logs    []client.JournalEntry
	docs    []client.Doc
}

// Each tab keeps its own cursor so switching back lands where you left.
type model struct {
	api    *client.Client
	width  int
	height int

	screen screen
	tab    tab

	projects      []client.Project
	projectCursor int
	project       client.Project

	rows    rows
	cursors [tabCount]int

	// detail holds the body of the selected log or doc, which the list view
	// has no room for.
	detailTitle string
	detailBody  string

	loading bool
	status  string
	err     error
}

func newModel(api *client.Client, project string) model {
	m := model{api: api, loading: true, screen: screenProjects}
	if project != "" {
		m.project = client.Project{Slug: project}
		m.screen = screenProject
	}
	return m
}

func (m model) cursor() int {
	return m.cursors[m.tab]
}

func (m *model) setCursor(at int) {
	if at < 0 {
		at = 0
	}
	if limit := m.rowCount() - 1; at > limit {
		at = limit
	}
	if at < 0 {
		at = 0
	}
	m.cursors[m.tab] = at
}

func (m model) rowCount() int {
	switch m.tab {
	case tabTickets:
		return len(m.rows.tickets)
	case tabPlans:
		return len(m.rows.plans)
	case tabTodos:
		return len(m.rows.todos)
	case tabLogs:
		return len(m.rows.logs)
	case tabDocs:
		return len(m.rows.docs)
	}
	return 0
}
