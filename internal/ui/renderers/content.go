package renderers

import (
	"fmt"
	"strings"

	"dbsage/internal/models"

	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
)

// ContentRenderer handles content rendering
type ContentRenderer struct {
	width int
}

func NewContentRenderer() *ContentRenderer {
	return &ContentRenderer{width: 80}
}

// SetWidth updates the renderer width
func (r *ContentRenderer) SetWidth(width int) {
	r.width = width
}

// RenderWelcomeBox renders the welcome message
func (r *ContentRenderer) RenderWelcomeBox() string {
	welcome := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Render("DBSage - Database AI Assistant") +
		"\n\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Ask me anything about your database! You can use natural language or commands.") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("? for help")

	return welcome
}

// RenderHelp renders the help information
func (r *ContentRenderer) RenderHelp() string {
	helpContent := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Render("Available commands:") +
		"\n\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Database Commands:") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("- /add <name>: Add database connection") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("- /switch <name>: Switch to connection") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("- /list: List all connections") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("- /remove <name>: Remove connection") +
		"\n\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("General Commands:") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("- /help: Show this help") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("- /clear: Clear screen") +
		"\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("- /exit or /quit: Exit application")

	return helpContent
}

// RenderHistory renders conversation history
func (r *ContentRenderer) RenderHistory(history []openai.ChatCompletionMessage) string {
	if len(history) == 0 {
		return ""
	}

	var sections []string

	for _, msg := range history {
		if msg.Role == openai.ChatMessageRoleUser {
			sections = append(sections, r.renderUserMessage(msg.Content))
		} else if msg.Role == openai.ChatMessageRoleAssistant && msg.Content != "" {
			sections = append(sections, r.renderAssistantMessage(msg.Content))
		}
	}

	return strings.Join(sections, "\n\n")
}

// renderUserMessage renders a user message
func (r *ContentRenderer) renderUserMessage(content string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(r.width - 4). // Leave some margin
		Render("> " + content)
}

// renderAssistantMessage renders an assistant message
func (r *ContentRenderer) renderAssistantMessage(content string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(r.width - 4). // Leave some margin
		Render("• " + content)
}

// RenderResponse renders a response message
func (r *ContentRenderer) RenderResponse(response string) string {
	return r.renderAssistantMessage(response)
}

// RenderError renders an error message
func (r *ContentRenderer) RenderError(err error) string {
	errorContent := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Width(r.width - 4). // Leave some margin
		Render("Error: " + err.Error())

	return errorContent
}

// RenderThinking renders the thinking animation
func (r *ContentRenderer) RenderThinking() string {
	thinkingContent := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(r.width - 4). // Leave some margin
		Render("Processing...")

	return thinkingContent
}

// RenderToolConfirmationBox renders tool confirmation dialog
func (r *ContentRenderer) RenderToolConfirmationBox(toolInfo *models.ToolConfirmationInfo, listView string) string {
	if toolInfo == nil {
		return ""
	}

	// Get risk level color
	riskColor := "214"
	switch toolInfo.RiskLevel {
	case "high":
		riskColor = "196"
	case "medium":
		riskColor = "214"
	case "low":
		riskColor = "46"
	}

	// Build content with proper width constraints
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(r.width - 4).
		Render("Tool Confirmation Required")

	toolName := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(r.width - 4).
		Render(fmt.Sprintf("Tool: %s", toolInfo.ToolName))

	description := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(r.width - 4).
		Render(fmt.Sprintf("Description: %s", toolInfo.Description))

	riskLevel := lipgloss.NewStyle().
		Foreground(lipgloss.Color(riskColor)).
		Width(r.width - 4).
		Render(fmt.Sprintf("Risk Level: %s", strings.ToUpper(toolInfo.RiskLevel)))

	content := title + "\n\n" + toolName + "\n" + description + "\n" + riskLevel + "\n\n" + listView
	return content
}
