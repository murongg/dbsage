package handlers

import (
	"fmt"
	"strings"

	"dbsage/internal/models"
	"dbsage/pkg/database"
	"dbsage/pkg/dbinterfaces"
)

// CommandHandler handles slash commands and @ database commands
type CommandHandler struct {
	connService dbinterfaces.ConnectionServiceInterface
}

func NewCommandHandler(connService dbinterfaces.ConnectionServiceInterface) *CommandHandler {
	return &CommandHandler{
		connService: connService,
	}
}

// ProcessCommand processes slash commands and @ database commands
func (h *CommandHandler) ProcessCommand(input string) (bool, string, error) {
	input = strings.TrimSpace(input)

	if strings.HasPrefix(input, "/") {
		return h.processSlashCommand(input)
	}

	if strings.HasPrefix(input, "@") {
		return h.processDatabaseCommand(input)
	}

	return false, "", nil // Not a command
}

// processSlashCommand handles slash commands like /help, /add, etc.
func (h *CommandHandler) processSlashCommand(input string) (bool, string, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return true, "Invalid command", nil
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "/help":
		return true, h.getHelpMessage(), nil

	case "/add":
		if len(args) < 1 {
			return true, "Usage: /add <connection_name> [database_url]\nExample: /add mydb postgres://user:pass@host:5432/db", nil
		}
		if len(args) >= 2 {
			return h.addConnectionWithURL(args[0], args[1])
		}
		return h.addConnection(args[0])

	case "/switch":
		if len(args) < 1 {
			return true, "Usage: /switch <connection_name>\nExample: /switch mydb", nil
		}
		return h.switchConnection(args[0])

	case "/list":
		return h.listConnections()

	case "/remove":
		if len(args) < 1 {
			return true, "Usage: /remove <connection_name>\nExample: /remove mydb", nil
		}
		return h.removeConnection(args[0])

	case "/clear":
		return true, "CLEAR_SCREEN", nil

	case "/exit", "/quit":
		return true, "EXIT", nil

	default:
		return true, fmt.Sprintf("Unknown command: %s\nType /help for available commands", command), nil
	}
}

// processDatabaseCommand handles @ database commands
func (h *CommandHandler) processDatabaseCommand(input string) (bool, string, error) {
	// Remove @ prefix
	content := strings.TrimPrefix(input, "@")
	content = strings.TrimSpace(content)

	if content == "" {
		// Show available connections for selection
		return h.showConnectionsForSelection()
	}

	// Check if it's a connection name (no spaces, looks like identifier)
	if !strings.Contains(content, " ") && h.isValidConnectionName(content) {
		// Try to switch to this connection
		return h.switchConnection(content)
	}

	// Otherwise, treat it as a database query to be processed by AI
	// Return false to let AI handle it
	return false, "", nil
}

// showConnectionsForSelection shows available connections for @ selection
func (h *CommandHandler) showConnectionsForSelection() (bool, string, error) {
	if h.connService == nil {
		return true, "Connection service not available", nil
	}

	connections, status, current := h.connService.GetConnectionInfo()

	if len(connections) == 0 {
		return true, "No database connections configured.\nUse /add <name> to add a connection.", nil
	}

	var result strings.Builder
	result.WriteString("Select a database connection:\n\n")

	for name, config := range connections {
		statusStr := status[name]
		if statusStr == "" {
			statusStr = "unknown"
		}

		marker := " "
		if name == current {
			marker = "*"
		}

		result.WriteString(fmt.Sprintf("%s @%s (%s:%d/%s) - %s\n",
			marker, name, config.Host, config.Port, config.Database, statusStr))
	}

	result.WriteString("\nUsage: @<connection_name> to switch")
	if current != "" {
		result.WriteString(fmt.Sprintf("\nCurrent: %s", current))
	}

	return true, result.String(), nil
}

// isValidConnectionName checks if a string could be a valid connection name
func (h *CommandHandler) isValidConnectionName(name string) bool {
	if h.connService == nil {
		return false
	}

	connections, _, _ := h.connService.GetConnectionInfo()
	_, exists := connections[name]
	return exists
}

// getHelpMessage returns the help message
func (h *CommandHandler) getHelpMessage() string {
	return `Available commands:

Database Commands:
- /add <name> [url]: Add database connection (supports postgresql, mysql)
  Examples: 
    /add mydb postgres://user:pass@host:5432/db
    /add mydb (interactive setup)
- /switch <name>: Switch to connection  
- /list: List all connections with types
- /remove <name>: Remove connection

General Commands:
- /help: Show this help
- /clear: Clear screen
- /exit or /quit: Exit application

Database Selection & Queries:
- @: Show available database connections
- @<connection_name>: Switch to database connection
- @<query>: Execute database query directly
- Examples: @mydb, @show tables`
}

// addConnection adds a new database connection
func (h *CommandHandler) addConnection(name string) (bool, string, error) {
	if h.connService == nil {
		return true, "Connection service not available", nil
	}

	// This is a simplified version - in reality you'd need to collect connection details
	return true, fmt.Sprintf("To add connection '%s', you need to provide:\n- Database type (postgresql, mysql)\n- Host\n- Port\n- Database name\n- Username\n- Password\n\nSupported database types: postgresql, mysql\nUse the interactive setup or configuration file.\n\nAlternatively, use: /add %s <database_url>\nExample: /add %s postgres://user:pass@host:5432/db", name, name, name), nil
}

// addConnectionWithURL adds a new database connection using a database URL
func (h *CommandHandler) addConnectionWithURL(name, databaseURL string) (bool, string, error) {
	if h.connService == nil {
		return true, "Connection service not available", nil
	}

	// Validate URL format
	if !database.ValidateDatabaseURL(databaseURL) {
		return true, "Invalid database URL format. Supported formats:\n- postgres://user:password@host:port/database\n- postgresql://user:password@host:port/database", nil
	}

	// Parse the database URL
	config, err := database.ParseDatabaseURL(databaseURL)
	if err != nil {
		return true, fmt.Sprintf("Failed to parse database URL: %v", err), nil
	}

	// Set the connection name and type
	config.Name = name
	config.Type = "postgresql"

	// Add the connection
	err = h.connService.AddConnection(config)
	if err != nil {
		return true, fmt.Sprintf("Failed to add connection '%s': %v", name, err), nil
	}

	return true, fmt.Sprintf("Successfully added connection '%s'\nHost: %s:%d\nDatabase: %s\nUsername: %s",
		name, config.Host, config.Port, config.Database, config.Username), nil
}

// switchConnection switches to a different connection
func (h *CommandHandler) switchConnection(name string) (bool, string, error) {
	if h.connService == nil {
		return true, "Connection service not available", nil
	}

	err := h.connService.SwitchConnection(name)
	if err != nil {
		return true, fmt.Sprintf("Failed to switch to connection '%s': %v", name, err), nil
	}

	return true, fmt.Sprintf("Switched to connection: %s", name), nil
}

// listConnections lists all available connections
func (h *CommandHandler) listConnections() (bool, string, error) {
	if h.connService == nil {
		return true, "Connection service not available", nil
	}

	connections, status, current := h.connService.GetConnectionInfo()

	if len(connections) == 0 {
		return true, "No database connections configured.\nUse /add <name> to add a connection.", nil
	}

	var result strings.Builder
	result.WriteString("Database Connections:\n\n")

	for name, config := range connections {
		statusStr := status[name]
		if statusStr == "" {
			statusStr = "unknown"
		}

		marker := " "
		if name == current {
			marker = "*"
		}

		dbType := config.Type
		if dbType == "" {
			dbType = "unknown"
		}

		result.WriteString(fmt.Sprintf("%s %s [%s] (%s:%d/%s) - %s\n",
			marker, name, dbType, config.Host, config.Port, config.Database, statusStr))
	}

	if current != "" {
		result.WriteString(fmt.Sprintf("\nCurrent: %s", current))
	}

	return true, result.String(), nil
}

// removeConnection removes a database connection
func (h *CommandHandler) removeConnection(name string) (bool, string, error) {
	if h.connService == nil {
		return true, "Connection service not available", nil
	}

	err := h.connService.RemoveConnection(name)
	if err != nil {
		return true, fmt.Sprintf("Failed to remove connection '%s': %v", name, err), nil
	}

	return true, fmt.Sprintf("Removed connection: %s", name), nil
}

// GetCommandSuggestions returns command suggestions based on input
func (h *CommandHandler) GetCommandSuggestions(input string) []*models.CommandInfo {
	var suggestions []*models.CommandInfo

	if strings.HasPrefix(input, "/") {
		commands := []*models.CommandInfo{
			{Name: "/help", Description: "Show available commands", Category: "general"},
			{Name: "/add", Description: "Add database connection", Category: "database"},
			{Name: "/switch", Description: "Switch to connection", Category: "database"},
			{Name: "/list", Description: "List all connections", Category: "database"},
			{Name: "/remove", Description: "Remove connection", Category: "database"},
			{Name: "/clear", Description: "Clear screen", Category: "general"},
			{Name: "/exit", Description: "Exit application", Category: "general"},
			{Name: "/quit", Description: "Exit application", Category: "general"},
		}

		for _, cmd := range commands {
			if strings.HasPrefix(cmd.Name, input) {
				suggestions = append(suggestions, cmd)
			}
		}
	} else if strings.HasPrefix(input, "@") {
		// Add @ command suggestions
		suggestions = append(suggestions, &models.CommandInfo{
			Name: "@", Description: "Show available database connections", Category: "database",
		})

		// Add connection names as suggestions
		if h.connService != nil {
			connections, _, _ := h.connService.GetConnectionInfo()
			for name := range connections {
				connSuggestion := "@" + name
				if strings.HasPrefix(connSuggestion, input) {
					suggestions = append(suggestions, &models.CommandInfo{
						Name: connSuggestion, Description: "Switch to " + name, Category: "database",
					})
				}
			}
		}
	}

	return suggestions
}
