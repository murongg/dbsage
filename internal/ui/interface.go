package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"dbsage/internal/ai"
	"dbsage/pkg/database"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
)

// Message types
type aiResponseMsg struct {
	response string
	err      error
}

type aiThinkingMsg struct{}

type commandCompletedMsg struct{}

type tickMsg time.Time

type aiStreamChunkMsg struct {
	chunk string
}

type aiStreamCompleteMsg struct {
	fullResponse string
}

type toolConfirmationMsg struct {
	toolInfo *ToolConfirmationInfo
}

type toolConfirmationResponseMsg struct {
	confirmed bool
	action    string // "execute", "cancel", "edit"
}

// ConfirmationItem represents an item in the confirmation list
type ConfirmationItem struct {
	title       string
	description string
	action      string
}

func (i ConfirmationItem) Title() string       { return i.title }
func (i ConfirmationItem) Description() string { return i.description }
func (i ConfirmationItem) FilterValue() string { return i.title }

// Custom styles for confirmation list - minimal padding
var (
	confirmationItemStyle         = lipgloss.NewStyle().PaddingLeft(1)
	confirmationSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
)

// confirmationDelegate implements list.ItemDelegate for confirmation items
type confirmationDelegate struct{}

func newConfirmationDelegate() confirmationDelegate {
	return confirmationDelegate{}
}

func (d confirmationDelegate) Height() int                             { return 1 }
func (d confirmationDelegate) Spacing() int                            { return 0 }
func (d confirmationDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d confirmationDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ConfirmationItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := confirmationItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return confirmationSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// tick generates a tick command for animation updates
func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model is the main Bubble Tea model
type Model struct {
	stateManager      *AppStateManager
	uiRenderer        *UIRenderer
	textInput         textinput.Model
	confirmationList  list.Model // List component for confirmation dialog
	width             int
	height            int
	streamingResponse string       // Accumulates streaming response chunks
	program           *tea.Program // Reference to the Bubble Tea program
}

// NewModel creates a new Bubble Tea model
func NewModel(client *ai.Client, dbTools *database.DatabaseTools, connService *database.ConnectionService) *Model {
	// Initialize text input component
	ti := textinput.New()
	ti.Placeholder = "Enter your question..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 70
	// Keep the default "> " prompt for input

	// Initialize components
	stateManager := NewAppStateManager(client, dbTools, connService)
	uiRenderer := NewUIRenderer()

	// Initialize confirmation list with custom styling
	confirmationList := list.New([]list.Item{}, newConfirmationDelegate(), 0, 0)
	confirmationList.Title = ""
	confirmationList.SetShowStatusBar(false)
	confirmationList.SetFilteringEnabled(false)
	confirmationList.SetShowHelp(false)
	confirmationList.SetShowTitle(true)

	// Set custom styles - minimal spacing
	confirmationList.Styles.Title = lipgloss.NewStyle().MarginLeft(0)
	confirmationList.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(0)
	confirmationList.Styles.HelpStyle = lipgloss.NewStyle().PaddingLeft(0)

	model := &Model{
		stateManager:     stateManager,
		uiRenderer:       uiRenderer,
		textInput:        ti,
		confirmationList: confirmationList,
		width:            80, // Default width
		height:           24, // Default height
	}

	// Set up tool confirmation callback and config
	if client != nil {
		client.SetToolConfirmationCallback(model.handleToolConfirmation)

		// Convert ui config to ai config to avoid circular import
		uiConfig := GetDefaultToolConfirmationConfig()
		if uiConfig != nil {
			aiConfig := ai.NewToolConfirmationConfig(
				uiConfig.RequiresConfirmation,
				uiConfig.RiskLevels,
				uiConfig.Descriptions,
			)
			client.SetToolConfirmationConfig(aiConfig)
		}
	}

	return model
}

// SetProgram sets the program reference for sending messages
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		textinput.Blink,
	)
}

// Update handles messages and updates state
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.uiRenderer.SetDimensions(m.width, m.height)

		// Update textinput width
		m.textInput.Width = m.width - 12 // Reserve space for borders and labels
		if m.textInput.Width < 20 {
			m.textInput.Width = 20
		}

		// Update confirmation list dimensions
		m.confirmationList.SetWidth(m.width - 4)
		m.confirmationList.SetHeight(10) // Fixed height for confirmation dialog
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Handle escape key differently based on state
			if m.stateManager.GetState() == StateToolConfirmation {
				return m, func() tea.Msg {
					return toolConfirmationResponseMsg{confirmed: false, action: "cancel"}
				}
			}
			return m, tea.Quit

		case "ctrl+h":
			showHelp := !m.stateManager.IsShowHelp()
			m.stateManager.SetShowHelp(showHelp)
			return m, nil

		case "tab":
			// Handle command completion
			if m.stateManager.GetState() == StateInput {
				return m.handleTabCompletion()
			}
			return m, nil

		case "enter":
			if m.stateManager.GetState() == StateInput && strings.TrimSpace(m.textInput.Value()) != "" {
				return m.handleInput()
			} else if m.stateManager.GetState() == StateToolConfirmation {
				// Handle confirmation list selection
				if selectedItem, ok := m.confirmationList.SelectedItem().(ConfirmationItem); ok {
					confirmed := selectedItem.action == "execute"
					return m, func() tea.Msg {
						return toolConfirmationResponseMsg{confirmed: confirmed, action: selectedItem.action}
					}
				}
			}
			return m, nil

		default:
			// Handle different states
			if m.stateManager.GetState() == StateToolConfirmation {
				// Pass keys to confirmation list for navigation
				m.confirmationList, cmd = m.confirmationList.Update(msg)
				return m, cmd
			} else if m.stateManager.GetState() == StateInput || m.stateManager.GetState() == StateResponse {
				// If we were showing a response and user starts typing, clear it and go to input state
				if m.stateManager.GetState() == StateResponse {
					m.stateManager.SetState(StateInput)
					m.stateManager.SetResponse("")
					m.stateManager.SetError(nil)
				}

				m.textInput, cmd = m.textInput.Update(msg)

				// Update command suggestions based on current input
				currentInput := m.textInput.Value()
				m.stateManager.UpdateCommandSuggestions(currentInput)

				return m, cmd
			}
			return m, nil
		}

	case aiThinkingMsg:
		m.stateManager.SetState(StateThinking)
		m.textInput.Blur() // Lose focus because thinking
		return m, tick()   // Start animation ticker

	case aiResponseMsg:
		if msg.err != nil {
			m.stateManager.SetError(msg.err)
			m.stateManager.SetState(StateResponse)
			// Prepare for next input even on error
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, textinput.Blink
		} else {
			m.stateManager.SetResponse(msg.response)
			m.stateManager.SetError(nil)
			m.stateManager.SetState(StateInput) // Switch directly to input state

			// Add to history
			m.stateManager.AddToHistory(openai.ChatMessageRoleAssistant, msg.response)
		}

		// Reset input and prepare for next input
		m.textInput.SetValue("")
		m.textInput.Focus() // Regain focus, prepare for next input
		return m, textinput.Blink

	case commandCompletedMsg:
		// Command completed, keep the response visible but ensure input is focused
		// Don't change state back to StateInput - keep it as StateResponse so response shows

		// Clear any command suggestions or parameter help
		m.stateManager.SetShowSuggestions(false)
		m.stateManager.SetCommandSuggestions(nil)
		m.stateManager.SetShowParameterHelp(false)
		m.stateManager.SetParameterHelp("")

		m.textInput.Focus()
		return m, textinput.Blink

	case tickMsg:
		// Handle animation tick - only continue if still thinking
		if m.stateManager.GetState() == StateThinking {
			return m, tick() // Continue animation
		}
		return m, nil

	case aiStreamChunkMsg:
		// Accumulate streaming response chunks and update UI immediately
		m.streamingResponse += msg.chunk

		// Only update response state if not in tool confirmation
		if m.stateManager.GetState() != StateToolConfirmation {
			m.stateManager.SetState(StateResponse)
			m.stateManager.SetResponse(m.streamingResponse)
			m.stateManager.SetError(nil)
		}
		return m, nil

	case aiStreamCompleteMsg:
		// Streaming is complete, add to history and prepare for next input
		m.stateManager.AddToHistory(openai.ChatMessageRoleAssistant, msg.fullResponse)

		// Only reset to input state if not in tool confirmation
		if m.stateManager.GetState() != StateToolConfirmation {
			m.stateManager.SetState(StateInput)

			// Reset input and prepare for next input
			m.textInput.SetValue("")
			m.textInput.Focus()
		}

		m.streamingResponse = "" // Clear streaming buffer
		return m, textinput.Blink

	case toolConfirmationMsg:
		// Tool confirmation required
		m.stateManager.SetPendingToolConfirmation(msg.toolInfo)
		m.stateManager.SetState(StateToolConfirmation)
		m.textInput.Blur() // Remove focus from input

		// Setup confirmation list items based on tool info
		var items []list.Item
		if msg.toolInfo != nil {
			for _, option := range msg.toolInfo.Options {
				items = append(items, ConfirmationItem{
					title:       option.Label,
					description: option.Description,
					action:      option.Action,
				})
			}
		}

		m.confirmationList.SetItems(items)
		// Set simple title without duplicating SQL
		if msg.toolInfo != nil {
			m.confirmationList.Title = "Choose an action:"
		}

		return m, nil

	case toolConfirmationResponseMsg:
		// Handle tool confirmation response
		if msg.confirmed && msg.action == "execute" {
			// Execute the pending tool
			return m.executePendingTool()
		} else {
			// Cancel or edit - return to input state
			toolInfo := m.stateManager.GetPendingToolConfirmation()
			m.stateManager.ClearPendingToolConfirmation()
			m.stateManager.SetState(StateInput)
			m.textInput.Focus()
			if msg.action == "edit" && toolInfo != nil {
				// Pre-fill input based on tool type
				if toolInfo.ToolName == "execute_sql" {
					if sql, ok := toolInfo.Arguments["sql"].(string); ok {
						m.textInput.SetValue("Execute this SQL: " + sql)
					}
				} else {
					m.textInput.SetValue("Execute " + toolInfo.Description)
				}
			}
			return m, textinput.Blink
		}

	}

	return m, nil
}

// handleTabCompletion handles tab key for command completion
func (m *Model) handleTabCompletion() (tea.Model, tea.Cmd) {
	currentInput := strings.TrimSpace(m.textInput.Value())

	// Complete if input starts with "/" or "@"
	if !strings.HasPrefix(currentInput, "/") && !strings.HasPrefix(currentInput, "@") {
		return m, nil
	}

	// Get command suggestions
	suggestions := m.stateManager.GetCommandSuggestions()
	if len(suggestions) == 1 {
		// If there's only one suggestion, complete it
		completed := suggestions[0].Name
		if !strings.HasSuffix(completed, " ") {
			completed += " "
		}
		m.textInput.SetValue(completed)

		// Clear suggestions and parameter help after completion
		m.stateManager.SetShowSuggestions(false)
		m.stateManager.SetCommandSuggestions(nil)
		m.stateManager.SetShowParameterHelp(false)
		m.stateManager.SetParameterHelp("")

		return m, nil
	} else if len(suggestions) > 1 {
		// Find common prefix for partial completion
		commonPrefix := findCommonPrefix(suggestions)
		if len(commonPrefix) > len(currentInput) {
			m.textInput.SetValue(commonPrefix)
		}
		return m, nil
	}

	return m, nil
}

// findCommonPrefix finds the common prefix among command suggestions
func findCommonPrefix(suggestions []*CommandInfo) string {
	if len(suggestions) == 0 {
		return ""
	}

	prefix := suggestions[0].Name
	for _, cmd := range suggestions[1:] {
		// Find common prefix between current prefix and this command
		for i := 0; i < len(prefix) && i < len(cmd.Name); i++ {
			if prefix[i] != cmd.Name[i] {
				prefix = prefix[:i]
				break
			}
		}
		if len(prefix) == 0 {
			break
		}
	}

	return prefix
}

// handleInput handles user input
func (m *Model) handleInput() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.textInput.Value())

	// Clear command suggestions and parameter help when processing input
	m.stateManager.SetShowSuggestions(false)
	m.stateManager.SetCommandSuggestions(nil)
	m.stateManager.SetShowParameterHelp(false)
	m.stateManager.SetParameterHelp("")

	// Process input through state manager
	if !m.stateManager.ProcessInput(input) {
		// Exit was requested
		return m, tea.Quit
	}

	// Check if it was handled as a command
	if m.stateManager.GetState() != StateThinking {
		// Command was handled, reset input and ensure focus
		m.textInput.SetValue("")
		m.textInput.Focus() // Ensure input box keeps focus

		// If the command resulted in a response, we need to trigger a view update
		if m.stateManager.GetState() == StateResponse {
			return m, func() tea.Msg { return commandCompletedMsg{} }
		}

		return m, nil
	}

	// Send thinking state message and start AI query
	return m, tea.Batch(
		func() tea.Msg { return aiThinkingMsg{} },
		m.queryAI(),
	)
}

// queryAI queries AI with real streaming support
func (m *Model) queryAI() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		history := m.stateManager.GetHistory()
		aiClient := m.stateManager.GetAIClient()

		// Start streaming in a goroutine
		go func() {
			var fullResponse strings.Builder

			err := aiClient.QueryWithToolsStreaming(ctx, history, func(chunk string) error {
				fullResponse.WriteString(chunk)

				// Send chunk message to the program if we have a reference
				if m.program != nil {
					m.program.Send(aiStreamChunkMsg{chunk: chunk})
				}
				return nil
			})

			// Send completion or error message
			if m.program != nil {
				if err != nil {
					m.program.Send(aiResponseMsg{response: "", err: err})
				} else {
					m.program.Send(aiStreamCompleteMsg{fullResponse: fullResponse.String()})
				}
			}
		}()

		// Return immediately - the streaming will happen asynchronously
		return nil
	}
}

// View renders the interface
func (m *Model) View() string {
	var contentSections []string

	// Fixed welcome message box
	welcomeBox := m.uiRenderer.RenderWelcomeBox()
	contentSections = append(contentSections, welcomeBox)

	// Help information (if needed)
	if m.stateManager.IsShowHelp() {
		help := m.uiRenderer.RenderHelp()
		contentSections = append(contentSections, help)
	}

	// Conversation history
	history := m.stateManager.GetHistory()
	if len(history) > 0 {
		historyContent := m.uiRenderer.RenderHistory(history)
		contentSections = append(contentSections, historyContent)
	}

	// Current response - show both command responses and errors
	if m.stateManager.GetState() == StateResponse {
		if m.stateManager.GetError() != nil {
			errorContent := m.uiRenderer.RenderError(m.stateManager.GetError())
			contentSections = append(contentSections, errorContent)
		} else if response := m.stateManager.GetResponse(); response != "" {
			responseContent := m.uiRenderer.RenderResponse(response)
			contentSections = append(contentSections, responseContent)
		}
	}

	// Thinking state
	if m.stateManager.GetState() == StateThinking {
		thinking := m.uiRenderer.RenderThinking()
		contentSections = append(contentSections, thinking)
	}

	// Tool confirmation state - show combined tool info and list in one box
	if m.stateManager.GetState() == StateToolConfirmation {
		toolInfo := m.stateManager.GetPendingToolConfirmation()
		if toolInfo != nil {
			// Render combined tool info and confirmation options
			combinedContent := m.uiRenderer.RenderToolConfirmationBox(toolInfo, m.confirmationList.View())
			contentSections = append(contentSections, combinedContent)
		}
	}

	// Input box area - hide when in tool confirmation state
	var inputBox string
	var commandList string
	var parameterHelp string

	if m.stateManager.GetState() != StateToolConfirmation {
		// Show input box and related components only when not in confirmation state
		inputBox = m.renderInputBoxWithConnection()

		// Command list (show below input box, like Claude Code CLI)
		if m.stateManager.IsShowSuggestions() {
			suggestions := m.stateManager.GetCommandSuggestions()
			commandList = m.uiRenderer.RenderCommandList(suggestions)
		}

		// Parameter help (show when specific command is selected)
		if m.stateManager.IsShowParameterHelp() {
			help := m.stateManager.GetParameterHelp()
			parameterHelp = m.uiRenderer.RenderParameterHelp(help)
		}
	}

	// Show divider if there's conversation history and not showing help, suggestions, or parameter help
	showDivider := len(history) > 0 && !m.stateManager.IsShowHelp() &&
		!m.stateManager.IsShowSuggestions() && !m.stateManager.IsShowParameterHelp()

	return m.uiRenderer.BuildLayout(contentSections, inputBox, commandList, parameterHelp, showDivider)
}

// renderInputBoxWithConnection renders the input box with connection indicator
func (m *Model) renderInputBoxWithConnection() string {
	// Get current connection info
	currentConn := ""
	if connMgr := m.stateManager.GetConnectionManager(); connMgr != nil {
		if _, name, err := connMgr.GetCurrentConnection(); err == nil && name != "" {
			currentConn = name
		}
	}

	// Render the basic input box
	inputBox := m.textInput.View()

	// If there's no current connection, return the input box as-is
	if currentConn == "" {
		return inputBox
	}

	// Create connection indicator
	connIndicator := m.uiRenderer.RenderConnectionIndicator(currentConn)

	// Combine connection indicator with input box
	return connIndicator + inputBox
}

// Run runs the Bubble Tea program
func Run(aiClient *ai.Client, dbTools *database.DatabaseTools, connService *database.ConnectionService) error {
	model := NewModel(aiClient, dbTools, connService)

	p := tea.NewProgram(
		model,
		// Remove tea.WithAltScreen() to allow text selection and copying
		tea.WithMouseCellMotion(),
	)

	// Set the program reference in the model for streaming
	model.SetProgram(p)

	_, err := p.Run()
	return err
}

// executePendingTool executes the pending tool using the AI client
func (m *Model) executePendingTool() (tea.Model, tea.Cmd) {
	// Get the pending AI context
	aiContext := m.stateManager.GetPendingAIContext()
	if aiContext == nil {
		m.stateManager.SetError(fmt.Errorf("no pending AI context to resume"))
		m.stateManager.SetState(StateResponse)
		m.textInput.Focus()
		return m, textinput.Blink
	}

	// Get AI client
	aiClient := m.stateManager.GetAIClient()
	if aiClient == nil {
		m.stateManager.SetError(fmt.Errorf("no AI client available"))
		m.stateManager.SetState(StateResponse)
		m.stateManager.ClearPendingToolConfirmation()
		m.textInput.Focus()
		return m, textinput.Blink
	}

	// Clear pending confirmation
	m.stateManager.ClearPendingToolConfirmation()

	// Set state to thinking before continuing
	m.stateManager.SetState(StateThinking)

	// Continue AI processing with the confirmed tool
	return m, func() tea.Msg {
		ctx := context.Background()

		// Create streaming callback with completion handling
		var fullResponse strings.Builder
		streamingCallback := func(chunk string) error {
			if m.program != nil {
				m.program.Send(aiStreamChunkMsg{chunk: chunk})
			}
			// Accumulate response for completion
			fullResponse.WriteString(chunk)
			return nil
		}

		err := aiClient.ContinueWithConfirmedTool(
			ctx,
			aiContext.Messages,
			aiContext.CompleteMessage,
			aiContext.ToolCall,
			streamingCallback,
		)

		// Send completion or error message
		if m.program != nil {
			if err != nil {
				m.program.Send(aiResponseMsg{response: "", err: err})
			} else {
				m.program.Send(aiStreamCompleteMsg{fullResponse: fullResponse.String()})
			}
		}

		return nil
	}
}

// handleToolConfirmation handles tool confirmation requests from the AI client
func (m *Model) handleToolConfirmation(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, streamingCallback ai.StreamingCallback) (bool, error) {
	// Check if this tool requires confirmation
	if !m.stateManager.RequiresConfirmation(toolCall.Function.Name) {
		return true, nil // No confirmation needed, proceed
	}

	// Parse arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return false, fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	// Create tool confirmation info
	toolInfo := m.stateManager.CreateToolConfirmationInfo(toolCall.Function.Name, toolCall.ID, args)
	if toolInfo == nil {
		return true, nil // Fallback to allowing execution
	}

	// Store the AI context for later resumption
	aiContext := &PendingAIContext{
		Messages:          messages,
		CompleteMessage:   completeMessage,
		ToolCall:          toolCall,
		StreamingCallback: streamingCallback,
	}
	m.stateManager.SetPendingAIContext(aiContext)

	// Send confirmation message to the UI if we have a program reference
	if m.program != nil {
		m.program.Send(toolConfirmationMsg{toolInfo: toolInfo})

		// Return false to indicate that confirmation is pending
		// The actual execution will happen when the user responds
		return false, nil
	}

	// Fallback: if no program reference, allow execution
	return true, nil
}
