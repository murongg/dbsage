package renderers

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LayoutRenderer handles the main layout rendering
type LayoutRenderer struct {
	width  int
	height int
}

func NewLayoutRenderer() *LayoutRenderer {
	return &LayoutRenderer{
		width:  80,
		height: 24,
	}
}

// SetDimensions updates the renderer dimensions
func (r *LayoutRenderer) SetDimensions(width, height int) {
	r.width = width
	r.height = height
}

// BuildLayout builds the complete layout
func (r *LayoutRenderer) BuildLayout(
	contentSections []string,
	inputBox string,
	commandList string,
	parameterHelp string,
	showDivider bool,
) string {
	var sections []string

	// Add content sections
	sections = append(sections, contentSections...)

	// Add divider if needed
	if showDivider {
		divider := r.renderDivider()
		sections = append(sections, divider)
	}

	// Add input area components
	if inputBox != "" {
		sections = append(sections, inputBox)
	}

	if commandList != "" {
		sections = append(sections, commandList)
	}

	if parameterHelp != "" {
		sections = append(sections, parameterHelp)
	}

	return strings.Join(sections, "\n\n")
}

// renderDivider renders a visual divider
func (r *LayoutRenderer) renderDivider() string {
	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(r.width - 4)

	return dividerStyle.Render(strings.Repeat("─", r.width-4))
}

// RenderConnectionIndicator renders the database connection indicator
func (r *LayoutRenderer) RenderConnectionIndicator(connectionName string) string {
	if connectionName == "" {
		return ""
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("69")). // 使用蓝色高亮连接名
		Bold(true).                       // 加粗显示
		MarginRight(1)

	return style.Render(connectionName + " >")
}
