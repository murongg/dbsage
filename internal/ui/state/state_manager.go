package state

import (
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
}

// NewStateManager creates a new state manager
func NewStateManager(aiClient *ai.Client, dbTools dbinterfaces.DatabaseInterface, connService dbinterfaces.ConnectionServiceInterface) *StateManager {
	var connMgr dbinterfaces.ConnectionManagerInterface
	if connService != nil {
		connMgr = connService.GetConnectionManager()
	}

	cmdHandler := handlers.NewCommandHandler(connService)

	return &StateManager{
		aiClient:               aiClient,
		dbTools:                dbTools,
		connMgr:                connMgr,
		cmdHandler:             cmdHandler,
		currentState:           models.StateInput,
		history:                make([]openai.ChatCompletionMessage, 0),
		toolConfirmationConfig: GetDefaultToolConfirmationConfig(),
	}
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
