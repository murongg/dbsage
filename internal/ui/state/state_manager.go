package state

import (
	"strings"

	"dbsage/internal/ai"
	"dbsage/internal/models"
	"dbsage/internal/ui/handlers"
	"dbsage/pkg/dbinterfaces"

	"github.com/sashabaranov/go-openai"
)

// StateManager manages the application state
type StateManager struct {
	aiClient           *ai.Client
	dbTools            dbinterfaces.DatabaseInterface
	connMgr            dbinterfaces.ConnectionManagerInterface
	cmdHandler         *handlers.CommandHandler
	history            []openai.ChatCompletionMessage
	currentState       models.AppState
	response           string
	error              error
	showHelp           bool
	commandSuggestions []*models.CommandInfo
	showSuggestions    bool
	selectedSuggestion int // Index of selected suggestion
	parameterHelp      string
	showParameterHelp  bool
	// Tool confirmation fields
	pendingToolConfirmation *models.ToolConfirmationInfo
	toolConfirmationConfig  *models.ToolConfirmationConfig
	pendingAIContext        *models.PendingAIContext // Store AI context for resuming after confirmation
	// Guidance fields
	currentGuidance *models.GuidanceInfo
	hasApiKey       bool
}

// NewStateManager creates a new state manager
func NewStateManager(aiClient *ai.Client, dbTools dbinterfaces.DatabaseInterface, connService dbinterfaces.ConnectionServiceInterface) *StateManager {
	var connMgr dbinterfaces.ConnectionManagerInterface
	if connService != nil {
		connMgr = connService.GetConnectionManager()
	}

	cmdHandler := handlers.NewCommandHandler(connService)

	hasApiKey := aiClient != nil

	sm := &StateManager{
		aiClient:               aiClient,
		dbTools:                dbTools,
		connMgr:                connMgr,
		cmdHandler:             cmdHandler,
		currentState:           models.StateInput,
		history:                make([]openai.ChatCompletionMessage, 0),
		toolConfirmationConfig: GetDefaultToolConfirmationConfig(),
		hasApiKey:              hasApiKey,
	}

	// Check if we need to show guidance
	sm.checkAndSetInitialGuidance()

	return sm
}

// State getters and setters
func (sm *StateManager) GetState() models.AppState {
	return sm.currentState
}

func (sm *StateManager) SetState(state models.AppState) {
	sm.currentState = state
}

func (sm *StateManager) GetResponse() string {
	return sm.response
}

func (sm *StateManager) SetResponse(response string) {
	sm.response = response
}

func (sm *StateManager) GetError() error {
	return sm.error
}

func (sm *StateManager) SetError(err error) {
	sm.error = err
}

// Help state management
func (sm *StateManager) IsShowHelp() bool {
	return sm.showHelp
}

func (sm *StateManager) SetShowHelp(show bool) {
	sm.showHelp = show
}

// Command suggestions management
func (sm *StateManager) GetCommandSuggestions() []*models.CommandInfo {
	return sm.commandSuggestions
}

func (sm *StateManager) SetCommandSuggestions(suggestions []*models.CommandInfo) {
	sm.commandSuggestions = suggestions
	sm.selectedSuggestion = 0 // Reset selection when suggestions change
}

func (sm *StateManager) IsShowSuggestions() bool {
	return sm.showSuggestions
}

func (sm *StateManager) SetShowSuggestions(show bool) {
	sm.showSuggestions = show
	if !show {
		sm.selectedSuggestion = 0
	}
}

// Suggestion navigation methods
func (sm *StateManager) GetSelectedSuggestion() int {
	return sm.selectedSuggestion
}

func (sm *StateManager) MoveSuggestionUp() {
	if len(sm.commandSuggestions) == 0 {
		return
	}
	sm.selectedSuggestion--
	if sm.selectedSuggestion < 0 {
		sm.selectedSuggestion = len(sm.commandSuggestions) - 1
	}
}

func (sm *StateManager) MoveSuggestionDown() {
	if len(sm.commandSuggestions) == 0 {
		return
	}
	sm.selectedSuggestion++
	if sm.selectedSuggestion >= len(sm.commandSuggestions) {
		sm.selectedSuggestion = 0
	}
}

func (sm *StateManager) GetSelectedSuggestionCommand() string {
	if len(sm.commandSuggestions) == 0 || sm.selectedSuggestion < 0 || sm.selectedSuggestion >= len(sm.commandSuggestions) {
		return ""
	}
	return sm.commandSuggestions[sm.selectedSuggestion].Name
}

// Parameter help management
func (sm *StateManager) GetParameterHelp() string {
	return sm.parameterHelp
}

func (sm *StateManager) SetParameterHelp(help string) {
	sm.parameterHelp = help
}

func (sm *StateManager) IsShowParameterHelp() bool {
	return sm.showParameterHelp
}

func (sm *StateManager) SetShowParameterHelp(show bool) {
	sm.showParameterHelp = show
}

// AI client management
func (sm *StateManager) GetAIClient() *ai.Client {
	return sm.aiClient
}

func (sm *StateManager) GetDatabaseTools() dbinterfaces.DatabaseInterface {
	return sm.dbTools
}

func (sm *StateManager) GetConnectionManager() dbinterfaces.ConnectionManagerInterface {
	return sm.connMgr
}

// History management
func (sm *StateManager) GetHistory() []openai.ChatCompletionMessage {
	return sm.history
}

func (sm *StateManager) AddToHistory(role string, content string) {
	sm.history = append(sm.history, openai.ChatCompletionMessage{
		Role:    role,
		Content: content,
	})
}

func (sm *StateManager) ClearHistory() {
	sm.history = make([]openai.ChatCompletionMessage, 0)
}

// Command processing
func (sm *StateManager) ProcessInput(input string) (bool, string) {
	if sm.cmdHandler == nil {
		return false, ""
	}

	handled, response, err := sm.cmdHandler.ProcessCommand(input)
	if err != nil {
		sm.SetError(err)
		sm.SetState(models.StateResponse)
		return true, ""
	}

	if handled {
		if response == "CLEAR_SCREEN" {
			sm.ClearHistory()
			sm.SetResponse("")
			sm.SetError(nil)
			return true, ""
		}

		if response == "EXIT" {
			return false, "" // Signal to exit
		}

		sm.SetResponse(response)
		sm.SetError(nil)
		sm.SetState(models.StateResponse)

		// Check if command may have affected database connections and refresh guidance
		if strings.HasPrefix(input, "/add") || strings.HasPrefix(input, "/switch") || strings.HasPrefix(input, "/remove") {
			// Get updated database tools from connection service
			if sm.connMgr != nil {
				if connService, ok := sm.connMgr.(interface {
					GetCurrentTools() dbinterfaces.DatabaseInterface
				}); ok {
					updatedTools := connService.GetCurrentTools()
					sm.UpdateDatabaseTools(updatedTools)
				}
			}
		}

		return true, ""
	}

	return true, "" // Not handled as command, continue with AI processing
}

// UpdateCommandSuggestions updates command suggestions based on input
func (sm *StateManager) UpdateCommandSuggestions(input string) {
	if sm.cmdHandler == nil {
		return
	}

	suggestions := sm.cmdHandler.GetCommandSuggestions(input)
	sm.SetCommandSuggestions(suggestions)
	sm.SetShowSuggestions(len(suggestions) > 0)
}

// Guidance management
func (sm *StateManager) GetCurrentGuidance() *models.GuidanceInfo {
	return sm.currentGuidance
}

func (sm *StateManager) SetCurrentGuidance(guidance *models.GuidanceInfo) {
	sm.currentGuidance = guidance
}

func (sm *StateManager) HasApiKey() bool {
	return sm.hasApiKey
}

func (sm *StateManager) checkAndSetInitialGuidance() {
	if !sm.hasApiKey {
		sm.currentGuidance = &models.GuidanceInfo{
			Type:    "api_key_missing",
			Title:   "🔑 API Key Required",
			Message: "To use DBSage AI features, you need to configure your OpenAI API key.",
			Instructions: []string{
				"1. Get your API key from OpenAI (https://platform.openai.com/api-keys)",
				"2. Set the environment variable: 'export OPENAI_API_KEY=your_api_key_here'",
				"3. Optionally set: 'export OPENAI_BASE_URL=https://api.openai.com/v1'",
				"4. Restart DBSage to use AI features",
			},
			Actions: []string{
				"You can still use database commands like '/add', '/list', '/switch' without API key",
				"Press 'q' to dismiss this message",
			},
		}
		return
	}

	if sm.dbTools == nil {
		sm.currentGuidance = &models.GuidanceInfo{
			Type:    "no_database",
			Title:   "🗄️ No Database Connected",
			Message: "Welcome to DBSage! You need to connect to a database to get started.",
			Instructions: []string{
				"1. Use '/add' <name> to add a new database connection",
				"2. Follow the prompts to enter connection details",
				"3. Use '/switch' <name> to switch between databases",
				"4. Type '/help' for more commands",
			},
			Actions: []string{
				"Example: '/add' mydb",
				"Press 'q' to dismiss this message",
			},
		}
		return
	}

	// Check if this is first time use (no history)
	if len(sm.history) == 0 {
		sm.currentGuidance = &models.GuidanceInfo{
			Type:    "first_time",
			Title:   "👋 Welcome to DBSage!",
			Message: "Your AI-powered database assistant is ready to help.",
			Instructions: []string{
				"• Ask questions in natural language: \"Show me all users\"",
				"• Use commands: '/help', '/list', '/add', '/switch'",
				"• Get table information: \"What tables do I have?\"",
				"• Analyze data: \"Find the top 10 customers by revenue\"",
				"• Optimize queries: \"How can I improve this query?\"",
			},
			Actions: []string{
				"Try asking: \"What tables are in my database?\"",
				"Press 'q' to dismiss this message",
			},
		}
	}
}

func (sm *StateManager) DismissGuidance() {
	sm.currentGuidance = nil
}

func (sm *StateManager) UpdateDatabaseTools(dbTools dbinterfaces.DatabaseInterface) {
	sm.dbTools = dbTools
	// Re-check guidance after database tools update
	sm.checkAndSetInitialGuidance()
}

func (sm *StateManager) RefreshGuidance() {
	// Re-check guidance state (useful after connection changes)
	sm.checkAndSetInitialGuidance()
}
