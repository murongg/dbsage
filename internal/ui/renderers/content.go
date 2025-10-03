package renderers

import (
	"fmt"
	"regexp"
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

// RenderWelcomeBoxWithStatus renders the welcome message with status indicators
func (r *ContentRenderer) RenderWelcomeBoxWithStatus(hasApiKey bool, hasDatabase bool) string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true).
		Render("DBSage - Database AI Assistant")

	description := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Ask me anything about your database! You can use natural language or commands.")

	var statusIndicators []string

	// API Key status
	if hasApiKey {
		apiStatus := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Render("✓ OpenAI API Key configured")
		statusIndicators = append(statusIndicators, apiStatus)
	} else {
		apiStatus := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("✗ OpenAI API Key missing")
		statusIndicators = append(statusIndicators, apiStatus)
	}

	// Database status
	if hasDatabase {
		dbStatus := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Render("✓ Database connected")
		statusIndicators = append(statusIndicators, dbStatus)
	} else {
		dbStatus := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("✗ No database connected")
		statusIndicators = append(statusIndicators, dbStatus)
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("? for help")

	var content []string
	content = append(content, title, "", description, "")
	content = append(content, statusIndicators...)
	content = append(content, "", help)

	return strings.Join(content, "\n")
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

// RenderGuidance renders user guidance information
func (r *ContentRenderer) RenderGuidance(guidance *models.GuidanceInfo) string {
	if guidance == nil {
		return ""
	}

	// Title with icon
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true).
		Width(r.width - 4).
		Render(guidance.Title)

	// Main message
	message := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(r.width - 4).
		Render(guidance.Message)

	var sections []string
	sections = append(sections, title, "", message)

	// Instructions section
	if len(guidance.Instructions) > 0 {
		instructionsTitle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true).
			Width(r.width - 4).
			Render("Instructions:")
		sections = append(sections, "", instructionsTitle)

		for _, instruction := range guidance.Instructions {
			// Highlight commands in instruction text
			highlightedInstruction := r.highlightCommands(instruction)
			instructionText := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Width(r.width - 6). // Extra indent
				Render("  " + highlightedInstruction)
			sections = append(sections, instructionText)
		}
	}

	// Actions section
	if len(guidance.Actions) > 0 {
		actionsTitle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Width(r.width - 4).
			Render("Quick Actions:")
		sections = append(sections, "", actionsTitle)

		for _, action := range guidance.Actions {
			// Highlight commands in action text
			highlightedAction := r.highlightCommands(action)
			actionText := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Width(r.width - 6). // Extra indent
				Render("  " + highlightedAction)
			sections = append(sections, actionText)
		}
	}

	// Add border around the guidance box
	content := strings.Join(sections, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(r.width - 2)

	return boxStyle.Render(content)
}

// highlightCommands highlights guidance-related commands and text
func (r *ContentRenderer) highlightCommands(text string) string {
	// Define patterns for guidance-specific content
	patterns := []struct {
		regex *regexp.Regexp
		color string
	}{
		// DBSage commands (slash commands)
		{regexp.MustCompile(`'/[a-zA-Z]+'`), "39"}, // '/add', '/help', etc. - cyan
		// Environment variables
		{regexp.MustCompile(`'export [A-Z_]+=[^']+'`), "226"}, // 'export OPENAI_API_KEY=...' - yellow
		{regexp.MustCompile(`'OPENAI_[A-Z_]+=[^']+'`), "226"}, // Direct env vars - yellow
		// Keyboard shortcuts
		{regexp.MustCompile(`'q'`), "208"},           // 'q' key - orange
		{regexp.MustCompile(`'\?'`), "208"},          // '?' key - orange
		{regexp.MustCompile(`'Ctrl\+[A-Z]'`), "208"}, // Ctrl+C - orange
		// Natural language queries (examples)
		{regexp.MustCompile(`'[^']*\?'`), "46"}, // Questions in quotes - green
		{regexp.MustCompile(`"[^"]*"`), "46"},   // Double quoted examples - green
	}

	result := text
	for _, pattern := range patterns {
		result = pattern.regex.ReplaceAllStringFunc(result, func(match string) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(pattern.color)).
				Bold(true).
				Render(match)
		})
	}

	return result
}

// RenderVersionUpdate renders version update notification
func (r *ContentRenderer) RenderVersionUpdate(updateInfo *models.VersionUpdateInfo) string {
	if updateInfo == nil || !updateInfo.HasUpdate {
		return ""
	}

	// Title with icon
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Width(r.width - 4).
		Render("🚀 New Version Available!")

	// Version information
	versionInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(r.width - 4).
		Render(fmt.Sprintf("Current: %s → Latest: %s", updateInfo.CurrentVersion, updateInfo.LatestVersion))

	// Release URL
	releaseURL := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Underline(true).
		Width(r.width - 4).
		Render(updateInfo.ReleaseURL)

	var sections []string
	sections = append(sections, title, "", versionInfo, "", releaseURL)

	// Release notes if available
	if updateInfo.ReleaseNotes != "" {
		notesTitle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Width(r.width - 4).
			Render("Release Notes:")

		// Truncate release notes if too long
		notes := updateInfo.ReleaseNotes
		if len(notes) > 200 {
			notes = notes[:200] + "..."
		}

		notesContent := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(r.width - 4).
			Render(notes)

		sections = append(sections, "", notesTitle, notesContent)
	}

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(r.width - 4).
		Render("Visit the link above to download the latest version.")

	sections = append(sections, "", instructions)

	// Add border around the notification
	content := strings.Join(sections, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("46")).
		Padding(1, 2).
		Width(r.width - 2)

	return boxStyle.Render(content)
}
