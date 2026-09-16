package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"folio/cli/internal/client"
)

// Run starts the browser. project preselects one, skipping the project list.
func Run(api *client.Client, project string) error {
	program := tea.NewProgram(newModel(api, project), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
