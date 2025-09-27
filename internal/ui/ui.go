package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
)

// UIRenderer handles all UI rendering logic
type UIRenderer struct {
	width            int
	height           int
	styles           *UIStyles
	markdownRenderer *MarkdownRenderer
}

// UIStyles contains all the styling information
type UIStyles struct {
	TitleStyle       lipgloss.Style
	WelcomeStyle     lipgloss.Style
	InputStyle       lipgloss.Style
	ResponseStyle    lipgloss.Style
	ConnectionStyle  lipgloss.Style // New style for connection lists
	ErrorStyle       lipgloss.Style
	HelpStyle        lipgloss.Style
	DividerStyle     lipgloss.Style
	PromptStyle      lipgloss.Style
	ThinkingStyle    lipgloss.Style
	UserMessageStyle lipgloss.Style // Style for user messages
}

// NewUIRenderer creates a new UI renderer
func NewUIRenderer() *UIRenderer {
	styles := &UIStyles{
		TitleStyle: lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center).
			MarginBottom(1),

		WelcomeStyle: lipgloss.NewStyle().
			MarginBottom(1),

		InputStyle: lipgloss.NewStyle().
			Padding(0, 1),

		ResponseStyle: lipgloss.NewStyle().
			MarginBottom(1),

		ConnectionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D9FF")). // Bright cyan for connection info
			MarginBottom(1),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true),

		HelpStyle: lipgloss.NewStyle().
			MarginBottom(1),

		DividerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),

		PromptStyle: lipgloss.NewStyle().
			Bold(true),

		ThinkingStyle: lipgloss.NewStyle().
			Bold(true),

		UserMessageStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")). // 浅灰色
			Faint(true),                           // 使字体更淡
	}

	// Create markdown renderer
	markdownRenderer := NewMarkdownRenderer(80)

	return &UIRenderer{
		width:            80,
		height:           24,
		styles:           styles,
		markdownRenderer: markdownRenderer,
	}
}

// SetDimensions updates the UI dimensions
func (ui *UIRenderer) SetDimensions(width, height int) {
	ui.width = width
	ui.height = height

	// Update markdown renderer width
	if ui.markdownRenderer != nil {
		ui.markdownRenderer.SetWidth(width)
	}
}

// RenderWelcomeBox renders the fixed welcome message box
func (ui *UIRenderer) RenderWelcomeBox() string {
	// Welcome message content
	content := "DBSage - Database Sage\n" +
		"Execute SQL Queries • Analyze Performance • Optimize Statements • Manage Indexes"

	// Calculate box width
	boxWidth := 60
	if ui.width > 0 && ui.width-4 < boxWidth {
		boxWidth = ui.width - 4
	}

	// Create bordered style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Center)

	return boxStyle.Render(content)
}

// RenderWelcome renders welcome message (kept for compatibility)
func (ui *UIRenderer) RenderWelcome() string {
	welcome := "Welcome to DBSage - Database Sage!\n\n" +
		"I am your database management expert, I can help you with:\n" +
		"• Execute SQL queries\n" +
		"• Analyze database performance\n" +
		"• Optimize query statements\n" +
		"• Manage indexes and table structures\n\n" +
		"Enter your question to start the conversation, or type 'help' to see more features"

	return ui.styles.WelcomeStyle.Render(welcome)
}

// RenderHelp renders help information
func (ui *UIRenderer) RenderHelp() string {
	help := "DBSage Feature List:\n\n" +
		"Basic Queries:\n" +
		"  • Execute SQL queries\n" +
		"  • View all tables\n" +
		"  • Get table structure\n\n" +
		"Performance Analysis:\n" +
		"  • Analyze query execution plans (EXPLAIN)\n" +
		"  • View slow queries\n" +
		"  • Get table statistics\n\n" +
		"Data Management:\n" +
		"  • Find duplicate data\n" +
		"  • Check database size\n" +
		"  • Analyze table sizes\n\n" +
		"System Monitoring:\n" +
		"  • View active connections\n" +
		"  • Monitor system status\n\n" +
		"Connection Management:\n" +
		"  • /add <name> [host] [port] [db] [user] [pass] [ssl] - Add connection\n" +
		"  • /list - List all connections\n" +
		"  • /switch <name> - Switch to connection\n" +
		"  • /remove <name> - Remove connection\n" +
		"  • /status - Show connection status\n\n" +
		"Tip: Simply describe your needs, AI will automatically select the right tools to help you complete the task!"

	return ui.styles.HelpStyle.Render(help)
}

// RenderHistory renders conversation history
func (ui *UIRenderer) RenderHistory(history []openai.ChatCompletionMessage) string {
	if len(history) == 0 {
		return ""
	}

	var lines []string

	// Calculate available width for content
	maxWidth := ui.width - 4 // Leave some margin
	if maxWidth < 40 {
		maxWidth = 40
	}

	// Show all history messages, referencing Claude Code's simple style
	for i, msg := range history {
		if msg.Role == openai.ChatMessageRoleUser {
			// User message - apply special styling for lighter color and smaller appearance with ">" prefix
			wrappedContent := WrapText(msg.Content, maxWidth-2) // Account for "> " prefix
			styledContent := ui.styles.UserMessageStyle.Render("> " + wrappedContent)
			lines = append(lines, styledContent)

		} else if msg.Role == openai.ChatMessageRoleAssistant {
			// AI reply - render as markdown with enhanced styling
			renderedContent := ui.markdownRenderer.RenderMarkdown(msg.Content)
			if renderedContent != "" {
				lines = append(lines, renderedContent)
			} else {
				// Fallback to plain text if markdown rendering fails
				wrappedContent := WrapText(msg.Content, maxWidth-2) // Account for "• " prefix
				lines = append(lines, "• "+wrappedContent)
			}
		}

		// Add empty line separator between messages
		if i < len(history)-1 {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

// RenderThinking renders thinking state
func (ui *UIRenderer) RenderThinking() string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frame := frames[int(time.Now().UnixNano()/100000000)%len(frames)]

	// Reference Claude Code's simple style
	thinking := fmt.Sprintf("• %s Thinking...", frame)
	return thinking
}

// RenderError renders error message
func (ui *UIRenderer) RenderError(err error) string {
	errorMsg := fmt.Sprintf("Error: %v", err)

	// Wrap error message if it's too long
	maxWidth := ui.width - 4
	if maxWidth < 40 {
		maxWidth = 40
	}
	wrappedError := WrapText(errorMsg, maxWidth)

	return ui.styles.ErrorStyle.Render(wrappedError)
}

// RenderResponse renders command response content
func (ui *UIRenderer) RenderResponse(response string) string {
	// Check if this is a connection list response (special handling)
	if strings.Contains(response, "Database Connections:") {
		maxWidth := ui.width - 4
		if maxWidth < 40 {
			maxWidth = 40
		}
		wrappedResponse := WrapText(response, maxWidth)
		return ui.renderConnectionList(wrappedResponse)
	}

	// Try to render as markdown first
	renderedMarkdown := ui.markdownRenderer.RenderMarkdown(response)
	if renderedMarkdown != "" {
		return renderedMarkdown
	}

	// Fallback to plain text rendering
	maxWidth := ui.width - 4
	if maxWidth < 40 {
		maxWidth = 40
	}
	wrappedResponse := WrapText(response, maxWidth)
	return ui.styles.ResponseStyle.Render(wrappedResponse)
}

// renderConnectionList renders connection list with enhanced colors
func (ui *UIRenderer) renderConnectionList(response string) string {
	lines := strings.Split(response, "\n")
	var styledLines []string

	for _, line := range lines {
		if strings.Contains(line, "Database Connections:") {
			// Header in bright cyan with bold
			headerStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D9FF")).
				Bold(true)
			styledLines = append(styledLines, headerStyle.Render(line))
		} else if strings.Contains(line, "Commands:") {
			// Commands hint in dim gray
			commandStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Italic(true)
			styledLines = append(styledLines, commandStyle.Render(line))
		} else if strings.TrimSpace(line) != "" && (strings.Contains(line, "[active]") || strings.Contains(line, "[connected]") || strings.Contains(line, "[disconnected]")) {
			// Connection entries with different colors based on status
			if strings.Contains(line, "[active]") {
				// Active connections in bright green
				connStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88"))
				styledLines = append(styledLines, connStyle.Render(line))
			} else if strings.Contains(line, "[connected]") {
				// Connected connections in yellow
				connStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
				styledLines = append(styledLines, connStyle.Render(line))
			} else {
				// Disconnected connections in light blue
				connStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB"))
				styledLines = append(styledLines, connStyle.Render(line))
			}
		} else {
			// Empty lines or other content - no special styling
			styledLines = append(styledLines, line)
		}
	}

	return strings.Join(styledLines, "\n")
}

// RenderConnectionIndicator renders the connection indicator for the input box
func (ui *UIRenderer) RenderConnectionIndicator(connectionName string) string {
	// Create connection indicator with styling
	indicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF88")). // Bright green for active connection
		Background(lipgloss.Color("#1a1a1a")). // Dark background
		Bold(true).
		Padding(0, 1).
		MarginRight(1)

	indicator := fmt.Sprintf("[%s]", connectionName)
	return indicatorStyle.Render(indicator)
}

// RenderCommandList renders command list below input (like Claude Code CLI)
func (ui *UIRenderer) RenderCommandList(suggestions []*CommandInfo) string {
	if len(suggestions) == 0 {
		return ""
	}

	var lines []string

	// Calculate the maximum command name width for alignment (similar to Claude Code CLI)
	maxNameWidth := 0
	for _, cmd := range suggestions {
		if len(cmd.Name) > maxNameWidth {
			maxNameWidth = len(cmd.Name)
		}
	}

	for _, cmd := range suggestions {
		// Format exactly like Claude Code CLI: /command-name    Description
		padding := strings.Repeat(" ", maxNameWidth-len(cmd.Name)+4)
		line := fmt.Sprintf("%s%s%s", cmd.Name, padding, cmd.Description)

		// Apply special styling for connection suggestions (starting with @)
		if strings.HasPrefix(cmd.Name, "@") {
			// Connection suggestions in cyan with special formatting
			connStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D9FF")).
				Bold(true)
			line = connStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Style to match Claude Code CLI - muted gray color, clean formatting
	listStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		MarginTop(1).
		MarginLeft(0)

	return listStyle.Render(strings.Join(lines, "\n"))
}

// RenderParameterHelp renders parameter help for a specific command
func (ui *UIRenderer) RenderParameterHelp(help string) string {
	if help == "" {
		return ""
	}

	// Style for parameter help - more prominent than command list
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Background(lipgloss.Color("#1a1a1a")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(1, 2).
		MarginTop(1)

	return helpStyle.Render(help)
}

// RenderFooter renders bottom hints
func (ui *UIRenderer) RenderFooter() string {
	footer := "? Type 'help' for assistance | Start with '/' for commands | Tab to complete"

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	return style.Render(footer)
}

// RenderToolInfo renders tool information with essential details
func (ui *UIRenderer) RenderToolInfo(toolInfo *ToolConfirmationInfo) string {
	if toolInfo == nil {
		return ""
	}

	// Risk indicator with color
	var riskColor string
	var riskText string
	switch toolInfo.RiskLevel {
	case "high":
		riskColor = "#FF4444" // Red
		riskText = "HIGH RISK"
	case "medium":
		riskColor = "#FFD700" // Gold
		riskText = "MEDIUM RISK"
	default:
		riskColor = "#00D9FF" // Cyan
		riskText = "LOW RISK"
	}

	var content []string

	// Show risk level and operation type
	content = append(content, fmt.Sprintf("OPERATION: %s (%s)", toolInfo.Description, riskText))
	content = append(content, "")

	// Show SQL query prominently if it's an execute_sql operation
	if toolInfo.ToolName == "execute_sql" {
		if query, ok := toolInfo.Arguments["sql"].(string); ok && query != "" {
			content = append(content, "SQL TO EXECUTE:")
			content = append(content, "")

			// Format SQL with syntax highlighting colors
			sqlStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D9FF")). // Cyan for SQL
				Background(lipgloss.Color("#1a1a1a")).
				Padding(0, 1)

			// Split query into lines and format each line
			lines := strings.Split(strings.TrimSpace(query), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					content = append(content, sqlStyle.Render("  "+strings.TrimSpace(line)))
				}
			}
		}
	} else {
		// Show other tool parameters
		if len(toolInfo.Arguments) > 0 {
			content = append(content, "PARAMETERS:")
			for key, value := range toolInfo.Arguments {
				content = append(content, fmt.Sprintf("  %s: %v", key, value))
			}
		}
	}

	// Clean style with colored border
	infoStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(riskColor)).
		Padding(1, 2).
		MarginBottom(1).
		Width(60)

	return infoStyle.Render(strings.Join(content, "\n"))
}

// RenderToolConfirmationBox renders tool information and confirmation options in one unified box
func (ui *UIRenderer) RenderToolConfirmationBox(toolInfo *ToolConfirmationInfo, listContent string) string {
	if toolInfo == nil {
		return ""
	}

	// Risk indicator with color
	var riskColor string
	var riskText string
	switch toolInfo.RiskLevel {
	case "high":
		riskColor = "#FF4444" // Red
		riskText = "HIGH RISK"
	case "medium":
		riskColor = "#FFD700" // Gold
		riskText = "MEDIUM RISK"
	default:
		riskColor = "#00D9FF" // Cyan
		riskText = "LOW RISK"
	}

	var content []string

	// Show risk level and operation type (compact)
	content = append(content, fmt.Sprintf("%s (%s)", toolInfo.Description, riskText))

	// Show SQL query prominently if it's an execute_sql operation
	if toolInfo.ToolName == "execute_sql" {
		if query, ok := toolInfo.Arguments["sql"].(string); ok && query != "" {
			// Format SQL with simple style - single line if possible
			sqlText := strings.TrimSpace(query)
			sqlText = strings.ReplaceAll(sqlText, "\n", " ")
			sqlText = strings.ReplaceAll(sqlText, "\t", " ")
			content = append(content, sqlText)
		}
	} else {
		// Show other tool parameters (compact)
		if len(toolInfo.Arguments) > 0 {
			for key, value := range toolInfo.Arguments {
				content = append(content, fmt.Sprintf("%s: %v", key, value))
			}
		}
	}

	// Add separator and confirmation list (no extra spacing)
	content = append(content, strings.Repeat("─", 30))
	content = append(content, listContent)

	// Create unified style with appropriate border color
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(riskColor)).
		Padding(1, 1).
		MarginBottom(1)
		// Remove fixed width to auto-fit content

	return boxStyle.Render(strings.Join(content, "\n"))
}

// RenderToolConfirmation renders tool confirmation dialog (kept for backward compatibility)
func (ui *UIRenderer) RenderToolConfirmation(toolInfo *ToolConfirmationInfo) string {
	if toolInfo == nil {
		return ""
	}

	// Create the confirmation dialog content based on risk level
	var title string
	var borderColor string

	switch toolInfo.RiskLevel {
	case "high":
		title = "[!] High Risk Operation Confirmation"
		borderColor = "#FF4444" // Red for high risk
	case "medium":
		title = "[*] Medium Risk Operation Confirmation"
		borderColor = "#FFD700" // Gold for medium risk
	default:
		title = "[i] Operation Confirmation"
		borderColor = "#00D9FF" // Cyan for low risk
	}

	question := fmt.Sprintf("Do you want to execute this %s operation?", toolInfo.ToolName)

	// Format the tool arguments
	formattedArgs := ui.formatToolArguments(toolInfo.ToolName, toolInfo.Arguments)

	// Build the dialog content
	var content []string
	content = append(content, title)
	content = append(content, "")
	content = append(content, question)
	content = append(content, "")
	content = append(content, fmt.Sprintf("Operation: %s", toolInfo.Description))
	if formattedArgs != "" {
		content = append(content, "")
		content = append(content, "Parameters:")
		content = append(content, formattedArgs)
	}
	content = append(content, "")
	content = append(content, "Options:")
	for _, option := range toolInfo.Options {
		content = append(content, fmt.Sprintf("%s. %s", option.Key, option.Label))
	}

	// Calculate dialog width
	dialogWidth := 80
	if ui.width > 0 && ui.width-8 < dialogWidth {
		dialogWidth = ui.width - 8
	}

	// Create dialog style with appropriate border color
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(2, 3).
		Width(dialogWidth).
		Align(lipgloss.Left).
		Background(lipgloss.Color("#1a1a1a"))

	return dialogStyle.Render(strings.Join(content, "\n"))
}

// formatToolArguments formats tool arguments based on tool type
func (ui *UIRenderer) formatToolArguments(toolName string, arguments map[string]interface{}) string {
	if len(arguments) == 0 {
		return ""
	}

	argStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")). // Cyan for parameters
		Background(lipgloss.Color("#2a2a2a")).
		Padding(1, 2).
		MarginLeft(2)

	var formatted []string

	// Special formatting for SQL queries
	if toolName == "execute_sql" {
		if sql, ok := arguments["sql"].(string); ok {
			lines := strings.Split(strings.TrimSpace(sql), "\n")
			for _, line := range lines {
				formatted = append(formatted, argStyle.Render("  "+strings.TrimSpace(line)))
			}
			return strings.Join(formatted, "\n")
		}
	}

	// General parameter formatting
	for key, value := range arguments {
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = fmt.Sprintf("\"%s\"", v)
		case []interface{}:
			// For arrays, format as comma-separated list
			var items []string
			for _, item := range v {
				items = append(items, fmt.Sprintf("%v", item))
			}
			valueStr = fmt.Sprintf("[%s]", strings.Join(items, ", "))
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		formatted = append(formatted, argStyle.Render(fmt.Sprintf("  %s: %s", key, valueStr)))
	}

	return strings.Join(formatted, "\n")
}

// formatSQLQuery formats SQL query with basic highlighting (kept for backward compatibility)
func (ui *UIRenderer) formatSQLQuery(query string) string {
	// Basic SQL formatting - add indentation and highlight keywords
	lines := strings.Split(strings.TrimSpace(query), "\n")
	var formatted []string

	sqlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")). // Cyan for SQL
		Background(lipgloss.Color("#2a2a2a")).
		Padding(1, 2).
		MarginLeft(2)

	for _, line := range lines {
		formatted = append(formatted, sqlStyle.Render("  "+strings.TrimSpace(line)))
	}

	return strings.Join(formatted, "\n")
}

// RenderDivider renders a simple divider
func (ui *UIRenderer) RenderDivider() string {
	divider := strings.Repeat("─", 40)
	return ui.styles.DividerStyle.Render(divider)
}

// BuildLayout builds the complete UI layout
func (ui *UIRenderer) BuildLayout(content []string, inputBox string, commandList string, parameterHelp string, showDivider bool) string {
	// Calculate available height, reserve space for input box, command list, parameter help and bottom hints
	reservedHeight := 6 // Base reserved height
	if commandList != "" {
		// Add extra space for command list
		commandLines := strings.Split(commandList, "\n")
		reservedHeight += len(commandLines) + 2 // +2 for spacing
	}
	if parameterHelp != "" {
		// Add extra space for parameter help
		paramLines := strings.Split(parameterHelp, "\n")
		reservedHeight += len(paramLines) + 4 // +4 for border and padding
	}

	contentHeight := ui.height - reservedHeight
	if contentHeight < 10 { // Ensure minimum height
		contentHeight = 10
	}

	// Merge all content
	fullContent := strings.Join(content, "\n\n")

	// If content is too long, truncate or scroll
	contentLines := strings.Split(fullContent, "\n")
	if len(contentLines) > contentHeight && contentHeight > 0 {
		// Show latest content
		startLine := len(contentLines) - contentHeight
		if startLine < 0 {
			startLine = 0
		}
		if startLine < len(contentLines) {
			contentLines = contentLines[startLine:]
			fullContent = strings.Join(contentLines, "\n")
		}
	}

	// Build final interface
	var finalSections []string

	// Add content area
	finalSections = append(finalSections, fullContent)

	// Add simple divider if needed
	if showDivider {
		finalSections = append(finalSections, ui.RenderDivider())
	}

	// Input box area - only show if not empty
	if inputBox != "" {
		finalSections = append(finalSections, inputBox)
	}

	// Command list below input box (like Claude Code CLI)
	if commandList != "" {
		finalSections = append(finalSections, commandList)
	}

	// Parameter help below command list (when specific command is selected)
	if parameterHelp != "" {
		finalSections = append(finalSections, parameterHelp)
	}

	// Bottom hints
	finalSections = append(finalSections, ui.RenderFooter())

	return strings.Join(finalSections, "\n")
}
