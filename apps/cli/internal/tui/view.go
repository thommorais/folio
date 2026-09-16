package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	bold     = lipgloss.NewStyle().Bold(true)
	selected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	activeTb = lipgloss.NewStyle().Bold(true).Underline(true)
	warn     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	ok       = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

func (m model) View() string {
	var b strings.Builder

	switch m.screen {
	case screenProjects:
		b.WriteString(m.viewProjects())
	case screenProject:
		b.WriteString(m.viewProject())
	case screenDetail:
		b.WriteString(m.viewDetail())
	}

	b.WriteString("\n")
	b.WriteString(dim.Render(m.help()))
	b.WriteString("\n")

	return b.String()
}

func (m model) header(text string) string {
	return bold.Render(text) + "\n" + dim.Render(strings.Repeat("─", lipgloss.Width(text))) + "\n\n"
}

func (m model) viewProjects() string {
	var b strings.Builder
	b.WriteString(m.header("folio"))

	if m.err != nil {
		return b.String() + warn.Render(m.err.Error()) + "\n"
	}
	if m.loading {
		return b.String() + dim.Render("loading…") + "\n"
	}
	if len(m.projects) == 0 {
		return b.String() + dim.Render("no projects") + "\n"
	}

	for i, project := range m.projects {
		line := fmt.Sprintf("%-22s %s", project.Slug, dim.Render(project.Name))
		if i == m.projectCursor {
			b.WriteString(selected.Render("› " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) viewProject() string {
	var b strings.Builder

	name := m.project.Name
	if name == "" {
		name = m.project.Slug
	}
	b.WriteString(m.header(name))
	b.WriteString(m.tabBar() + "\n\n")

	if m.err != nil {
		return b.String() + warn.Render(m.err.Error()) + "\n"
	}
	if m.loading {
		return b.String() + dim.Render("loading…") + "\n"
	}

	lines := m.rowLines()
	if len(lines) == 0 {
		return b.String() + dim.Render("no "+tabNames[m.tab]) + "\n"
	}

	for i, line := range lines {
		if i == m.cursor() {
			b.WriteString(selected.Render("› " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) tabBar() string {
	parts := make([]string, 0, tabCount)
	for i := tab(0); i < tabCount; i++ {
		label := fmt.Sprintf("%d %s", i+1, tabNames[i])
		if i == m.tab {
			parts = append(parts, activeTb.Render(label))
		} else {
			parts = append(parts, dim.Render(label))
		}
	}
	return strings.Join(parts, "   ")
}

func (m model) rowLines() []string {
	switch m.tab {
	case tabTickets:
		lines := make([]string, 0, len(m.rows.tickets))
		for _, ticket := range m.rows.tickets {
			progress := fmt.Sprintf("%d/%d", ticket.Progress.Done, ticket.Progress.Total)
			lines = append(lines, fmt.Sprintf("%-42s %-12s %s", truncate(ticket.Title, 42), ticket.Status, dim.Render(progress)))
		}
		return lines

	case tabPlans:
		lines := make([]string, 0, len(m.rows.plans))
		for _, plan := range m.rows.plans {
			progress := fmt.Sprintf("%d/%d", plan.Progress.Done, plan.Progress.Total)
			lines = append(lines, fmt.Sprintf("%-42s %-10s %s", truncate(plan.Title, 42), plan.Status, dim.Render(progress)))
		}
		return lines

	case tabTodos:
		lines := make([]string, 0, len(m.rows.todos))
		for _, todo := range m.rows.todos {
			mark := "○"
			style := lipgloss.NewStyle()
			switch todo.Status {
			case "done":
				mark, style = "●", ok
			case "in_progress":
				mark = "◐"
			case "blocked", "cancelled":
				mark, style = "◌", warn
			}
			lines = append(lines, fmt.Sprintf("%s %-42s %s", style.Render(mark), truncate(todo.Title, 42), dim.Render(todo.Priority)))
		}
		return lines

	case tabLogs:
		lines := make([]string, 0, len(m.rows.logs))
		for _, entry := range m.rows.logs {
			date := entry.CreatedAt
			if len(date) >= 10 {
				date = date[:10]
			}
			lines = append(lines, fmt.Sprintf("%-46s %s", truncate(entry.Title, 46), dim.Render(date)))
		}
		return lines

	case tabDocs:
		lines := make([]string, 0, len(m.rows.docs))
		for _, doc := range m.rows.docs {
			lines = append(lines, fmt.Sprintf("%-42s %s", truncate(doc.Title, 42), dim.Render(truncate(doc.Slug, 28))))
		}
		return lines
	}
	return nil
}

func (m model) viewDetail() string {
	var b strings.Builder
	b.WriteString(m.header(m.detailTitle))

	body := strings.TrimSpace(m.detailBody)
	if body == "" {
		body = dim.Render("(empty)")
	}
	b.WriteString(body + "\n")

	return b.String()
}

func (m model) help() string {
	switch m.screen {
	case screenProjects:
		return "↑↓ move · enter open · r reload · q quit"
	case screenDetail:
		return "esc back · q quit"
	default:
		return "↑↓ move · tab switch · 1-5 jump · enter detail · esc back · r reload · q quit"
	}
}

// Titles come from the API and may carry newlines or tabs, which would break
// row alignment, so every cell is flattened to a single line first.
func truncate(text string, width int) string {
	text = strings.Join(strings.Fields(text), " ")

	if lipgloss.Width(text) <= width {
		return text
	}
	runes := []rune(text)
	if width < 2 || len(runes) <= width {
		return text
	}
	return string(runes[:width-1]) + "…"
}
