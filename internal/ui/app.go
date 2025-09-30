package ui

import (
	"dbsage/internal/ai"
	"dbsage/internal/models"
	"dbsage/internal/ui/components"
	"dbsage/internal/ui/handlers"
	"dbsage/internal/ui/renderers"
	"dbsage/internal/ui/state"
	"dbsage/pkg/dbinterfaces"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the main Bubble Tea model
type Model struct {
	stateManager      *state.StateManager
	contentRenderer   *renderers.ContentRenderer
	commandRenderer   *renderers.CommandRenderer
	layoutRenderer    *renderers.LayoutRenderer
	inputHandler      *handlers.InputHandler
	toolHandler       *handlers.ToolHandler
	textInput         textinput.Model
	confirmationList  list.Model
	width             int
	height            int
	streamingResponse string
	program           *tea.Program
}

// NewModel creates a new Bubble Tea model
func NewModel(client *ai.Client, dbTools dbinterfaces.DatabaseInterface, connService dbinterfaces.ConnectionServiceInterface) *Model {
	// Initialize components
	stateManager := state.NewStateManager(client, dbTools, connService)
	contentRenderer := renderers.NewContentRenderer()
	commandRenderer := renderers.NewCommandRenderer()
	layoutRenderer := renderers.NewLayoutRenderer()
	inputHandler := handlers.NewInputHandler()
	toolHandler := handlers.NewToolHandler()

	// Initialize UI components
	textInput := components.CreateTextInput()
	confirmationList := components.CreateConfirmationList()

	model := &Model{
		stateManager:     stateManager,
		contentRenderer:  contentRenderer,
		commandRenderer:  commandRenderer,
		layoutRenderer:   layoutRenderer,
		inputHandler:     inputHandler,
		toolHandler:      toolHandler,
		textInput:        textInput,
		confirmationList: confirmationList,
		width:            80,
		height:           24,
	}

	// Set up tool confirmation callback
	if client != nil {
		client.SetToolConfirmationCallback(model.handleToolConfirmationFromAI)

		// Set tool confirmation config
		uiConfig := state.GetDefaultToolConfirmationConfig()
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case models.AIThinkingMsg:
		return m.handleAIThinking()

	case models.AIResponseMsg:
		return m.handleAIResponse(msg)

	case models.CommandCompletedMsg:
		return m.handleCommandCompleted()

	case models.TickMsg:
		return m.handleTick()

	case models.AIStreamChunkMsg:
		return m.handleStreamChunk(msg)

	case models.AIStreamCompleteMsg:
		return m.handleStreamComplete(msg)

	case models.ToolConfirmationMsg:
		return m.handleToolConfirmation(msg)

	case models.ToolConfirmationResponseMsg:
		return m.handleToolConfirmationResponse(msg)
	}

	return m, nil
}

// View renders the interface
func (m *Model) View() string {
	var contentSections []string

	// Fixed welcome message box with status
	hasApiKey := m.stateManager.HasApiKey()
	hasDatabase := m.stateManager.GetDatabaseTools() != nil
	welcomeBox := m.contentRenderer.RenderWelcomeBoxWithStatus(hasApiKey, hasDatabase)
	contentSections = append(contentSections, welcomeBox)

	// Guidance information (if needed)
	if guidance := m.stateManager.GetCurrentGuidance(); guidance != nil {
		guidanceContent := m.contentRenderer.RenderGuidance(guidance)
		contentSections = append(contentSections, guidanceContent)
	}

	// Help information (if needed)
	if m.stateManager.IsShowHelp() {
		help := m.contentRenderer.RenderHelp()
		contentSections = append(contentSections, help)
	}

	// Conversation history
	history := m.stateManager.GetHistory()
	if len(history) > 0 {
		historyContent := m.contentRenderer.RenderHistory(history)
		contentSections = append(contentSections, historyContent)
	}

	// Current response
	if m.stateManager.GetState() == models.StateResponse {
		if m.stateManager.GetError() != nil {
			errorContent := m.contentRenderer.RenderError(m.stateManager.GetError())
			contentSections = append(contentSections, errorContent)
		} else if response := m.stateManager.GetResponse(); response != "" {
			responseContent := m.contentRenderer.RenderResponse(response)
			contentSections = append(contentSections, responseContent)
		}
	}

	// Thinking state
	if m.stateManager.GetState() == models.StateThinking {
		thinking := m.contentRenderer.RenderThinking()
		contentSections = append(contentSections, thinking)
	}

	// Tool confirmation state
	if m.stateManager.GetState() == models.StateToolConfirmation {
		toolInfo := m.stateManager.GetPendingToolConfirmation()
		if toolInfo != nil {
			combinedContent := m.contentRenderer.RenderToolConfirmationBox(toolInfo, m.confirmationList.View())
			contentSections = append(contentSections, combinedContent)
		}
	}

	// Input area
	var inputBox string
	var commandList string
	var parameterHelp string

	if m.stateManager.GetState() != models.StateToolConfirmation {
		inputBox = m.renderInputBoxWithConnection()

		if m.stateManager.IsShowSuggestions() {
			suggestions := m.stateManager.GetCommandSuggestions()
			selectedIndex := m.stateManager.GetSelectedSuggestion()
			commandList = m.commandRenderer.RenderCommandList(suggestions, selectedIndex)
		}

		if m.stateManager.IsShowParameterHelp() {
			help := m.stateManager.GetParameterHelp()
			parameterHelp = m.commandRenderer.RenderParameterHelp(help)
		}
	}

	// Show divider if needed
	showDivider := len(history) > 0 && !m.stateManager.IsShowHelp() &&
		!m.stateManager.IsShowSuggestions() && !m.stateManager.IsShowParameterHelp()

	return m.layoutRenderer.BuildLayout(contentSections, inputBox, commandList, parameterHelp, showDivider)
}

// renderInputBoxWithConnection renders the input box with connection indicator
func (m *Model) renderInputBoxWithConnection() string {
	currentConn := ""
	if connMgr := m.stateManager.GetConnectionManager(); connMgr != nil {
		if _, name, err := connMgr.GetCurrentConnection(); err == nil && name != "" {
			currentConn = name
		}
	}

	if currentConn == "" {
		m.textInput.Prompt = "> "
		return m.textInput.View()
	}

	// Use the styled connection indicator
	connectionIndicator := m.layoutRenderer.RenderConnectionIndicator(currentConn)
	m.textInput.Prompt = ""
	inputBox := m.textInput.View()

	// Combine connection indicator with input box
	return connectionIndicator + inputBox
}

// Run runs the Bubble Tea program
func Run(aiClient *ai.Client, dbTools dbinterfaces.DatabaseInterface, connService dbinterfaces.ConnectionServiceInterface) error {
	model := NewModel(aiClient, dbTools, connService)

	p := tea.NewProgram(
		model,
	)

	model.SetProgram(p)

	_, err := p.Run()
	return err
}
