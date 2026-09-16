package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Init() tea.Cmd {
	if m.screen == screenProject {
		return tea.Batch(
			loadProject(m.api, m.project.Slug),
			loadTab(m.api, m.tab, m.project.Slug),
		)
	}
	return loadProjects(m.api)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case projectsLoaded:
		m.loading = false
		m.err = msg.err
		m.projects = msg.projects
		return m, nil

	case projectLoaded:
		m.err = msg.err
		if msg.err == nil {
			m.project = msg.project
		}
		return m, nil

	case tabLoaded:
		// A stale response from a tab the user has already left would
		// overwrite the visible rows.
		if msg.tab != m.tab {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		m.rows = msg.rows
		m.setCursor(m.cursor())
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc":
		return m.back()

	case "up", "k":
		return m.move(-1)

	case "down", "j":
		return m.move(1)

	case "enter", "l", "right":
		return m.forward()

	case "tab", "]":
		return m.switchTab(1)

	case "shift+tab", "[":
		return m.switchTab(-1)

	case "1", "2", "3", "4", "5":
		if m.screen == screenProjects {
			return m, nil
		}
		m.tab = tab(msg.String()[0] - '1')
		return m.reloadTab()

	case "r":
		return m.reloadTab()
	}

	return m, nil
}

func (m model) move(delta int) (tea.Model, tea.Cmd) {
	if m.screen == screenProjects {
		next := m.projectCursor + delta
		if next >= 0 && next < len(m.projects) {
			m.projectCursor = next
		}
		return m, nil
	}
	if m.screen == screenProject {
		m.setCursor(m.cursor() + delta)
	}
	return m, nil
}

func (m model) forward() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenProjects:
		if len(m.projects) == 0 {
			return m, nil
		}
		m.project = m.projects[m.projectCursor]
		m.screen = screenProject
		m.rows = rows{}
		m.cursors = [tabCount]int{}
		m.loading = true
		return m, loadTab(m.api, m.tab, m.project.Slug)

	case screenProject:
		title, body, ok := m.selectedDetail()
		if !ok {
			return m, nil
		}
		m.detailTitle, m.detailBody = title, body
		m.screen = screenDetail
		return m, nil
	}
	return m, nil
}

func (m model) back() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenDetail:
		m.screen = screenProject
		return m, nil
	case screenProject:
		m.screen = screenProjects
		if m.projects == nil {
			m.loading = true
			return m, loadProjects(m.api)
		}
		return m, nil
	}
	return m, tea.Quit
}

func (m model) switchTab(delta int) (tea.Model, tea.Cmd) {
	if m.screen != screenProject {
		return m, nil
	}
	m.tab = tab((int(m.tab) + delta + tabCount) % tabCount)
	return m.reloadTab()
}

func (m model) reloadTab() (tea.Model, tea.Cmd) {
	if m.screen == screenProjects {
		m.loading = true
		return m, loadProjects(m.api)
	}
	m.loading = true
	m.rows = rows{}
	return m, loadTab(m.api, m.tab, m.project.Slug)
}

func (m model) selectedDetail() (title, body string, ok bool) {
	at := m.cursor()
	switch m.tab {
	case tabTickets:
		if at < len(m.rows.tickets) {
			ticket := m.rows.tickets[at]
			return ticket.Title, ticket.Body, true
		}
	case tabLogs:
		if at < len(m.rows.logs) {
			entry := m.rows.logs[at]
			return entry.Title, entry.Body, true
		}
	case tabDocs:
		if at < len(m.rows.docs) {
			doc := m.rows.docs[at]
			return doc.Title, doc.Body, true
		}
	case tabPlans:
		if at < len(m.rows.plans) {
			plan := m.rows.plans[at]
			return plan.Title, plan.Goal, true
		}
	case tabTodos:
		if at < len(m.rows.todos) {
			todo := m.rows.todos[at]
			return todo.Title, todo.Details, true
		}
	}
	return "", "", false
}
