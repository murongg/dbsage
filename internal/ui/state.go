package ui

import (
	"dbsage/internal/ai"
	"dbsage/pkg/database"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// ToolConfirmationInfo contains information about a tool that needs confirmation
type ToolConfirmationInfo struct {
	ToolName    string                 `json:"tool_name"`
	ToolCallID  string                 `json:"tool_call_id"`
	Arguments   map[string]interface{} `json:"arguments"`
	Description string                 `json:"description"`
	RiskLevel   string                 `json:"risk_level"` // "low", "medium", "high"
	Options     []ConfirmationOption   `json:"options"`
}

// ConfirmationOption represents an option in the confirmation dialog
type ConfirmationOption struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Action      string `json:"action"` // "execute", "cancel", "edit", "custom"
}

// ToolConfirmationConfig defines which tools need confirmation
type ToolConfirmationConfig struct {
	RequiresConfirmation map[string]bool   `json:"requires_confirmation"`
	RiskLevels           map[string]string `json:"risk_levels"`
	Descriptions         map[string]string `json:"descriptions"`
}

// PendingAIContext stores the context needed to resume AI processing after confirmation
type PendingAIContext struct {
	Messages          []openai.ChatCompletionMessage `json:"messages"`
	CompleteMessage   openai.ChatCompletionMessage   `json:"complete_message"`
	ToolCall          openai.ToolCall                `json:"tool_call"`
	StreamingCallback ai.StreamingCallback           `json:"-"` // Not serializable
}

// GetDefaultToolConfirmationConfig returns the default tool confirmation configuration
func GetDefaultToolConfirmationConfig() *ToolConfirmationConfig {
	return &ToolConfirmationConfig{
		RequiresConfirmation: map[string]bool{
			"execute_sql":            true,
			"find_duplicate_data":    true,
			"get_all_tables":         false,
			"get_table_schema":       false,
			"explain_query":          false,
			"get_table_indexes":      false,
			"get_table_stats":        false,
			"get_slow_queries":       false,
			"get_database_size":      false,
			"get_table_sizes":        false,
			"get_active_connections": false,
		},
		RiskLevels: map[string]string{
			"execute_sql":            "high",
			"find_duplicate_data":    "medium",
			"get_all_tables":         "low",
			"get_table_schema":       "low",
			"explain_query":          "low",
			"get_table_indexes":      "low",
			"get_table_stats":        "low",
			"get_slow_queries":       "low",
			"get_database_size":      "low",
			"get_table_sizes":        "low",
			"get_active_connections": "low",
		},
		Descriptions: map[string]string{
			"execute_sql":            "Execute SQL query on the database",
			"find_duplicate_data":    "Search for duplicate data in table",
			"get_all_tables":         "List all tables in the database",
			"get_table_schema":       "Get table structure information",
			"explain_query":          "Analyze query execution plan",
			"get_table_indexes":      "Get table index information",
			"get_table_stats":        "Get table statistics",
			"get_slow_queries":       "Get slow query information",
			"get_database_size":      "Get database size information",
			"get_table_sizes":        "Get table size information",
			"get_active_connections": "Get active database connections",
		},
	}
}

// AppState represents the application state
type AppState int

const (
	StateWelcome AppState = iota
	StateInput
	StateThinking
	StateResponse
	StateHelp
	StateToolConfirmation
)

// AppStateManager manages the application state
type AppStateManager struct {
	aiClient           *ai.Client
	dbTools            *database.DatabaseTools
	connMgr            *database.ConnectionManager
	cmdHandler         *CommandHandler
	history            []openai.ChatCompletionMessage
	currentState       AppState
	response           string
	error              error
	showHelp           bool
	commandSuggestions []*CommandInfo
	showSuggestions    bool
	parameterHelp      string
	showParameterHelp  bool
	// Tool confirmation fields
	pendingToolConfirmation *ToolConfirmationInfo
	toolConfirmationConfig  *ToolConfirmationConfig
	pendingAIContext        *PendingAIContext // Store AI context for resuming after confirmation
}

// NewAppStateManager creates a new state manager
func NewAppStateManager(aiClient *ai.Client, dbTools *database.DatabaseTools, connService *database.ConnectionService) *AppStateManager {
	cmdHandler := NewCommandHandler(connService)

	return &AppStateManager{
		aiClient:               aiClient,
		dbTools:                dbTools,
		connMgr:                connService.GetConnectionManager(),
		cmdHandler:             cmdHandler,
		history:                make([]openai.ChatCompletionMessage, 0),
		currentState:           StateInput, // Start in input state
		toolConfirmationConfig: GetDefaultToolConfirmationConfig(),
	}
}

// GetState returns the current state
func (sm *AppStateManager) GetState() AppState {
	return sm.currentState
}

// SetState sets the current state
func (sm *AppStateManager) SetState(state AppState) {
	sm.currentState = state
}

// GetHistory returns the conversation history
func (sm *AppStateManager) GetHistory() []openai.ChatCompletionMessage {
	return sm.history
}

// AddToHistory adds a message to the conversation history
func (sm *AppStateManager) AddToHistory(role string, content string) {
	sm.history = append(sm.history, openai.ChatCompletionMessage{
		Role:    role,
		Content: content,
	})
}

// ClearHistory clears the conversation history
func (sm *AppStateManager) ClearHistory() {
	sm.history = make([]openai.ChatCompletionMessage, 0)
	sm.response = ""
	sm.error = nil
}

// GetResponse returns the current response
func (sm *AppStateManager) GetResponse() string {
	return sm.response
}

// SetResponse sets the current response
func (sm *AppStateManager) SetResponse(response string) {
	sm.response = response
}

// GetError returns the current error
func (sm *AppStateManager) GetError() error {
	return sm.error
}

// SetError sets the current error
func (sm *AppStateManager) SetError(err error) {
	sm.error = err
}

// IsShowHelp returns whether help should be shown
func (sm *AppStateManager) IsShowHelp() bool {
	return sm.showHelp
}

// SetShowHelp sets whether help should be shown
func (sm *AppStateManager) SetShowHelp(show bool) {
	sm.showHelp = show
}

// GetCommandSuggestions returns the current command suggestions
func (sm *AppStateManager) GetCommandSuggestions() []*CommandInfo {
	return sm.commandSuggestions
}

// SetCommandSuggestions sets the command suggestions
func (sm *AppStateManager) SetCommandSuggestions(suggestions []*CommandInfo) {
	sm.commandSuggestions = suggestions
}

// IsShowSuggestions returns whether command suggestions should be shown
func (sm *AppStateManager) IsShowSuggestions() bool {
	return sm.showSuggestions
}

// SetShowSuggestions sets whether command suggestions should be shown
func (sm *AppStateManager) SetShowSuggestions(show bool) {
	sm.showSuggestions = show
}

// GetParameterHelp returns the current parameter help
func (sm *AppStateManager) GetParameterHelp() string {
	return sm.parameterHelp
}

// SetParameterHelp sets the parameter help
func (sm *AppStateManager) SetParameterHelp(help string) {
	sm.parameterHelp = help
}

// IsShowParameterHelp returns whether parameter help should be shown
func (sm *AppStateManager) IsShowParameterHelp() bool {
	return sm.showParameterHelp
}

// SetShowParameterHelp sets whether parameter help should be shown
func (sm *AppStateManager) SetShowParameterHelp(show bool) {
	sm.showParameterHelp = show
}

// UpdateCommandSuggestions updates command suggestions based on current input
func (sm *AppStateManager) UpdateCommandSuggestions(input string) {
	// Check for @ symbol to show connection selector
	if strings.HasPrefix(strings.TrimSpace(input), "@") {
		sm.updateConnectionSuggestions(input)
		return
	}

	if sm.cmdHandler.IsCommandInput(input) {
		// Check if it's a complete command that needs parameter help
		if sm.cmdHandler.IsCompleteCommand(input) {
			paramHelp := sm.cmdHandler.GetParameterHelp(input)
			if paramHelp != "" {
				sm.parameterHelp = paramHelp
				sm.showParameterHelp = true
				sm.commandSuggestions = nil
				sm.showSuggestions = false
				return
			}
		}

		// Check if it's a command with partial parameters (e.g., "/add test")
		if sm.cmdHandler.IsCommandWithPartialParams(input) {
			// Use contextual parameter help that shows progress
			paramHelp := sm.cmdHandler.GetContextualParameterHelp(input)
			if paramHelp != "" {
				sm.parameterHelp = paramHelp
				sm.showParameterHelp = true
				sm.commandSuggestions = nil
				sm.showSuggestions = false
				return
			}
		}

		// Otherwise show command suggestions
		suggestions := sm.cmdHandler.GetCommandSuggestions(input)
		sm.commandSuggestions = suggestions
		sm.showSuggestions = len(suggestions) > 0
		sm.parameterHelp = ""
		sm.showParameterHelp = false
	} else {
		sm.commandSuggestions = nil
		sm.showSuggestions = false
		sm.parameterHelp = ""
		sm.showParameterHelp = false
	}
}

// updateConnectionSuggestions updates connection suggestions when @ symbol is used
func (sm *AppStateManager) updateConnectionSuggestions(input string) {
	// Get all available connections
	connections := sm.connMgr.ListConnections()
	if len(connections) == 0 {
		// No connections available
		sm.parameterHelp = "No database connections configured.\nUse '/add <name>' to add connections first."
		sm.showParameterHelp = true
		sm.commandSuggestions = nil
		sm.showSuggestions = false
		return
	}

	// Filter connections based on input after @
	inputText := strings.TrimSpace(input)
	searchTerm := ""
	if len(inputText) > 1 {
		searchTerm = strings.ToLower(inputText[1:]) // Remove @ symbol
	}

	var connectionSuggestions []*CommandInfo
	for name, config := range connections {
		// Filter by search term if provided
		if searchTerm == "" || strings.Contains(strings.ToLower(name), searchTerm) {
			// Get connection status
			status := sm.connMgr.GetConnectionStatus()
			statusText := "disconnected"
			if s, exists := status[name]; exists {
				switch s {
				case "active":
					statusText = "active"
				case "connected":
					statusText = "connected"
				default:
					statusText = "disconnected"
				}
			}

			suggestion := &CommandInfo{
				Name:        "@" + name,
				Usage:       "@" + name,
				Description: fmt.Sprintf("%s:%d/%s [%s]", config.Host, config.Port, config.Database, statusText),
				Example:     "@" + name,
			}
			connectionSuggestions = append(connectionSuggestions, suggestion)
		}
	}

	if len(connectionSuggestions) > 0 {
		sm.commandSuggestions = connectionSuggestions
		sm.showSuggestions = true
		sm.parameterHelp = "Select a connection to switch to it"
		sm.showParameterHelp = true
	} else {
		sm.commandSuggestions = nil
		sm.showSuggestions = false
		sm.parameterHelp = "No connections found matching: " + searchTerm
		sm.showParameterHelp = true
	}
}

// handleConnectionSelection handles connection selection with @ symbol
func (sm *AppStateManager) handleConnectionSelection(input string) bool {
	inputText := strings.TrimSpace(input)
	if len(inputText) <= 1 {
		// Just @ symbol, show help
		sm.response = "Type @<connection_name> to switch connections\nUse '/list' to see available connections"
		sm.error = nil
		sm.currentState = StateResponse
		return true
	}

	// Extract connection name after @
	connectionName := inputText[1:] // Remove @ symbol

	// Check if connection exists
	connections := sm.connMgr.ListConnections()
	if _, exists := connections[connectionName]; !exists {
		sm.response = fmt.Sprintf("Connection '%s' not found.\nAvailable connections:\n", connectionName)
		for name := range connections {
			sm.response += fmt.Sprintf("  • @%s\n", name)
		}
		sm.response += "\nUse '/add <name>' to add new connections."
		sm.error = nil
		sm.currentState = StateResponse
		return true
	}

	// Try to switch to the connection
	result := sm.cmdHandler.HandleCommand("/switch " + connectionName)
	if result != nil {
		sm.response = result.Message
		sm.error = result.Error
		sm.currentState = StateResponse

		// Update database tools if switch was successful
		if result.Success {
			if dbTools, _, err := sm.cmdHandler.GetCurrentDatabaseTools(); err == nil {
				sm.dbTools = dbTools
			}
		}
		return true
	}

	// Fallback error
	sm.response = fmt.Sprintf("Failed to switch to connection '%s'", connectionName)
	sm.error = nil
	sm.currentState = StateResponse
	return true
}

// ProcessInput processes user input and returns whether it was handled as a command
func (sm *AppStateManager) ProcessInput(input string) bool {
	// Check if this is a connection selection with @
	if strings.HasPrefix(strings.TrimSpace(input), "@") {
		return sm.handleConnectionSelection(input)
	}

	// Try to handle as command first
	if result := sm.cmdHandler.HandleCommand(input); result != nil {
		switch result.Message {
		case "exit":
			return false // Signal to exit
		case "help":
			sm.showHelp = true
			return true
		case "clear":
			sm.ClearHistory()
			sm.currentState = StateWelcome
			return true
		default:
			// Command result
			sm.response = result.Message
			sm.error = result.Error
			sm.currentState = StateResponse

			// Update database tools if connection changed
			if result.Success && (result.Message != "" &&
				(strings.Contains(result.Message, "Successfully added") ||
					strings.Contains(result.Message, "Switched to"))) {
				if dbTools, _, err := sm.cmdHandler.GetCurrentDatabaseTools(); err == nil {
					sm.dbTools = dbTools
				}
			}
			return true
		}
	}

	// Not a command, add to history for AI processing
	sm.AddToHistory(openai.ChatMessageRoleUser, input)
	sm.currentState = StateThinking
	return true
}

// GetAIClient returns the AI client
func (sm *AppStateManager) GetAIClient() *ai.Client {
	return sm.aiClient
}

// GetDatabaseTools returns the current database tools
func (sm *AppStateManager) GetDatabaseTools() *database.DatabaseTools {
	return sm.dbTools
}

// UpdateDatabaseTools updates the database tools (e.g., after connection switch)
func (sm *AppStateManager) UpdateDatabaseTools() {
	if dbTools, _, err := sm.cmdHandler.GetCurrentDatabaseTools(); err == nil {
		sm.dbTools = dbTools
	}
}

// GetConnectionManager returns the connection manager
func (sm *AppStateManager) GetConnectionManager() *database.ConnectionManager {
	return sm.connMgr
}

// SetPendingToolConfirmation sets the pending tool confirmation
func (sm *AppStateManager) SetPendingToolConfirmation(toolInfo *ToolConfirmationInfo) {
	sm.pendingToolConfirmation = toolInfo
}

// GetPendingToolConfirmation returns the pending tool confirmation
func (sm *AppStateManager) GetPendingToolConfirmation() *ToolConfirmationInfo {
	return sm.pendingToolConfirmation
}

// ClearPendingToolConfirmation clears the pending tool confirmation
func (sm *AppStateManager) ClearPendingToolConfirmation() {
	sm.pendingToolConfirmation = nil
	sm.pendingAIContext = nil
}

// SetPendingAIContext sets the pending AI context for resuming after confirmation
func (sm *AppStateManager) SetPendingAIContext(context *PendingAIContext) {
	sm.pendingAIContext = context
}

// GetPendingAIContext returns the pending AI context
func (sm *AppStateManager) GetPendingAIContext() *PendingAIContext {
	return sm.pendingAIContext
}

// RequiresConfirmation checks if a tool requires confirmation
func (sm *AppStateManager) RequiresConfirmation(toolName string) bool {
	if sm.toolConfirmationConfig == nil {
		return false
	}
	return sm.toolConfirmationConfig.RequiresConfirmation[toolName]
}

// CreateToolConfirmationInfo creates tool confirmation info for a given tool
func (sm *AppStateManager) CreateToolConfirmationInfo(toolName, toolCallID string, arguments map[string]interface{}) *ToolConfirmationInfo {
	config := sm.toolConfirmationConfig
	if config == nil {
		return nil
	}

	riskLevel := config.RiskLevels[toolName]
	description := config.Descriptions[toolName]

	// Minimal options
	options := []ConfirmationOption{
		{Key: "y", Label: "Execute", Description: "", Action: "execute"},
		{Key: "n", Label: "Cancel", Description: "", Action: "cancel"},
	}

	return &ToolConfirmationInfo{
		ToolName:    toolName,
		ToolCallID:  toolCallID,
		Arguments:   arguments,
		Description: description,
		RiskLevel:   riskLevel,
		Options:     options,
	}
}
