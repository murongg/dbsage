package renderers

import (
	"strings"

	"dbsage/internal/models"

	"github.com/charmbracelet/lipgloss"
)

// CommandRenderer handles command list and parameter help rendering
type CommandRenderer struct {
	width int
}

func NewCommandRenderer() *CommandRenderer {
	return &CommandRenderer{width: 80}
}

// SetWidth updates the renderer width
func (r *CommandRenderer) SetWidth(width int) {
	r.width = width
}

// RenderCommandList renders the command suggestions list
func (r *CommandRenderer) RenderCommandList(suggestions []*models.CommandInfo, selectedIndex int) string {
	if len(suggestions) == 0 {
		return ""
	}

	var items []string
	for i, cmd := range suggestions {
		var item string
		if i == selectedIndex {
			// Highlight selected item
			item = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("69")).
				Render("> " + cmd.Name + ": " + cmd.Description)
		} else {
			item = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render("  " + cmd.Name + ": " + cmd.Description)
		}
		items = append(items, item)
	}

	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Render("Available Commands:") +
		"\n" +
		strings.Join(items, "\n") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Use ↑↓ to navigate, Tab/Enter to select")

	return content
}

// RenderParameterHelp renders parameter help information
func (r *CommandRenderer) RenderParameterHelp(help string) string {
	if help == "" {
		return ""
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(r.width - 4).
		Render("Parameter Help:")

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(r.width - 4).
		Render(help)

	return title + "\n" + helpText
}
