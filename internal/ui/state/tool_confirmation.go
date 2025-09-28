package state

import (
	"dbsage/internal/models"
	"dbsage/internal/ui/handlers"
)

// Tool confirmation management
func (sm *StateManager) GetPendingToolConfirmation() *models.ToolConfirmationInfo {
	return sm.pendingToolConfirmation
}

func (sm *StateManager) SetPendingToolConfirmation(info *models.ToolConfirmationInfo) {
	sm.pendingToolConfirmation = info
}

func (sm *StateManager) ClearPendingToolConfirmation() {
	sm.pendingToolConfirmation = nil
}

func (sm *StateManager) GetPendingAIContext() *models.PendingAIContext {
	return sm.pendingAIContext
}

func (sm *StateManager) SetPendingAIContext(context *models.PendingAIContext) {
	sm.pendingAIContext = context
}

func (sm *StateManager) ClearPendingAIContext() {
	sm.pendingAIContext = nil
}

// RequiresConfirmation checks if a tool requires confirmation
func (sm *StateManager) RequiresConfirmation(toolName string) bool {
	toolHandler := handlers.NewToolHandler()
	return toolHandler.CheckToolConfirmation(toolName, sm.toolConfirmationConfig)
}

// CreateToolConfirmationInfo creates tool confirmation info
func (sm *StateManager) CreateToolConfirmationInfo(toolName, toolCallID string, args map[string]interface{}) *models.ToolConfirmationInfo {
	toolHandler := handlers.NewToolHandler()
	return toolHandler.CreateToolConfirmationInfo(toolName, toolCallID, args, sm.toolConfirmationConfig)
}

// GetDefaultToolConfirmationConfig returns the default tool confirmation configuration
func GetDefaultToolConfirmationConfig() *models.ToolConfirmationConfig {
	return &models.ToolConfirmationConfig{
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
			"find_duplicate_data":    "Find duplicate data in table",
			"get_all_tables":         "Get list of all tables",
			"get_table_schema":       "Get table schema information",
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
