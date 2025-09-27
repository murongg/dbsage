package ui

import (
	"fmt"
	"strings"
	"time"

	"dbsage/pkg/database"
)

// CommandResult represents the result of a command execution
type CommandResult struct {
	Message string
	Error   error
	Success bool
}

// CommandInfo represents information about a command
type CommandInfo struct {
	Name          string
	Usage         string
	Description   string
	Example       string
	ParameterHelp string // Detailed parameter help for when command is selected
}

// CommandHandler handles various CLI commands
type CommandHandler struct {
	connService *database.ConnectionService
	commands    map[string]*CommandInfo
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(connService *database.ConnectionService) *CommandHandler {
	ch := &CommandHandler{
		connService: connService,
		commands:    make(map[string]*CommandInfo),
	}

	// Initialize command information
	ch.initializeCommands()
	return ch
}

// initializeCommands initializes the command information
func (ch *CommandHandler) initializeCommands() {
	ch.commands = map[string]*CommandInfo{
		"/add": {
			Name:        "/add",
			Usage:       "/add <name> [url|host port db user pass ssl]",
			Description: "Add a new database connection",
			Example:     "/add mydb postgres://user:pass@host:5432/db",
			ParameterHelp: "Parameters:\n" +
				"  <name>     Connection name (required)\n" +
				"  [url]      Database URL: postgres://user:pass@host:port/db?sslmode=disable\n" +
				"  OR individual parameters:\n" +
				"  [host]     Database host (default: localhost)\n" +
				"  [port]     Database port (default: 5432)\n" +
				"  [database] Database name (default: postgres)\n" +
				"  [username] Username for authentication\n" +
				"  [password] Password for authentication\n" +
				"  [ssl_mode] SSL mode: disable, require, verify-ca, verify-full\n\n" +
				"Examples:\n" +
				"  /add mydb postgres://user:pass@localhost:5432/mydb?sslmode=disable\n" +
				"  /add prod localhost 5432 production admin secret require\n" +
				"  /add local (interactive mode)",
		},
		"/list": {
			Name:        "/list",
			Usage:       "/list",
			Description: "List all database connections",
			Example:     "/list",
			ParameterHelp: "Shows all configured database connections with their status.\n" +
				"No parameters required.\n\n" +
				"Output includes:\n" +
				"  • Connection name and status (active/connected/disconnected)\n" +
				"  • Host, port, and database information\n" +
				"  • Username (password hidden for security)",
		},
		"/switch": {
			Name:        "/switch",
			Usage:       "/switch <name>",
			Description: "Switch to a different database connection",
			Example:     "/switch mydb",
			ParameterHelp: "Parameters:\n" +
				"  <name>     Name of the connection to switch to (required)\n\n" +
				"The connection must already exist. Use '/list' to see available connections.\n" +
				"After switching, all database operations will use the selected connection.",
		},
		"/remove": {
			Name:        "/remove",
			Usage:       "/remove <name>",
			Description: "Remove a database connection",
			Example:     "/remove mydb",
			ParameterHelp: "Parameters:\n" +
				"  <name>     Name of the connection to remove (required)\n\n" +
				"This will permanently delete the connection configuration.\n" +
				"If you're currently using this connection, DBSage will switch to another\n" +
				"available connection automatically.",
		},
		"/status": {
			Name:        "/status",
			Usage:       "/status",
			Description: "Show connection status information",
			Example:     "/status",
			ParameterHelp: "Shows detailed status information about all connections.\n" +
				"No parameters required.\n\n" +
				"Output includes:\n" +
				"  • Current active connection\n" +
				"  • Status of all configured connections\n" +
				"  • Connection statistics and health information",
		},
	}
}

// HandleCommand processes a command and returns the result
func (ch *CommandHandler) HandleCommand(input string) *CommandResult {
	input = strings.TrimSpace(input)
	lowerInput := strings.ToLower(input)

	switch {
	case lowerInput == "exit" || lowerInput == "quit":
		return &CommandResult{Message: "exit", Success: true}
	case lowerInput == "help":
		return &CommandResult{Message: "help", Success: true}
	case lowerInput == "clear":
		return &CommandResult{Message: "clear", Success: true}
	case strings.HasPrefix(lowerInput, "/add"):
		return ch.handleAddConnection(input)
	case strings.HasPrefix(lowerInput, "/list"):
		return ch.handleListConnections()
	case strings.HasPrefix(lowerInput, "/switch"):
		return ch.handleSwitchConnection(input)
	case strings.HasPrefix(lowerInput, "/remove"):
		return ch.handleRemoveConnection(input)
	case strings.HasPrefix(lowerInput, "/status"):
		return ch.handleConnectionStatus()
	default:
		// Not a command, return nil to indicate normal AI processing
		return nil
	}
}

// GetCurrentDatabaseTools returns the current database tools
func (ch *CommandHandler) GetCurrentDatabaseTools() (*database.DatabaseTools, string, error) {
	connMgr := ch.connService.GetConnectionManager()
	return connMgr.GetCurrentConnection()
}

// GetCommandSuggestions returns command suggestions based on partial input
func (ch *CommandHandler) GetCommandSuggestions(partial string) []*CommandInfo {
	partial = strings.ToLower(strings.TrimSpace(partial))

	// If input is just "/", show all commands
	if partial == "/" {
		var suggestions []*CommandInfo
		for _, cmd := range ch.commands {
			suggestions = append(suggestions, cmd)
		}
		return suggestions
	}

	// If input starts with "/", find matching commands
	if strings.HasPrefix(partial, "/") {
		var suggestions []*CommandInfo
		for cmdName, cmd := range ch.commands {
			if strings.HasPrefix(strings.ToLower(cmdName), partial) {
				suggestions = append(suggestions, cmd)
			}
		}
		return suggestions
	}

	return nil
}

// GetAllCommands returns all available commands
func (ch *CommandHandler) GetAllCommands() map[string]*CommandInfo {
	return ch.commands
}

// IsCommandInput checks if the input looks like a command
func (ch *CommandHandler) IsCommandInput(input string) bool {
	return strings.HasPrefix(strings.TrimSpace(input), "/")
}

// GetParameterHelp returns parameter help for a specific command
func (ch *CommandHandler) GetParameterHelp(input string) string {
	input = strings.TrimSpace(input)

	// Check if input is exactly a command (with optional trailing space)
	normalizedInput := strings.TrimSpace(input)
	if strings.HasSuffix(normalizedInput, " ") {
		normalizedInput = strings.TrimSpace(normalizedInput)
	}

	// Look for exact command match
	if cmd, exists := ch.commands[normalizedInput]; exists {
		return cmd.ParameterHelp
	}

	// Check if input ends with a space (indicating command is complete and waiting for parameters)
	if strings.HasSuffix(input, " ") {
		parts := strings.Fields(input)
		if len(parts) > 0 {
			if cmd, exists := ch.commands[parts[0]]; exists {
				return cmd.ParameterHelp
			}
		}
	}

	return ""
}

// GetContextualParameterHelp returns parameter help with context about current progress
func (ch *CommandHandler) GetContextualParameterHelp(input string) string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ""
	}

	commandName := parts[0]
	cmd, exists := ch.commands[commandName]
	if !exists {
		return ""
	}

	// For /add command, provide contextual help based on current parameters
	if commandName == "/add" {
		paramCount := len(parts) - 1 // Exclude command name

		baseHelp := cmd.ParameterHelp

		if paramCount == 1 {
			// User has entered connection name, show what's next
			contextHelp := "\n🎯 Next: Enter database connection details\n" +
				"You can provide either:\n" +
				"• A complete database URL: postgres://user:pass@host:port/db\n" +
				"• Individual parameters: host port database username password ssl_mode\n" +
				"• Or press Enter for interactive mode"
			return baseHelp + contextHelp
		} else if paramCount >= 2 {
			// User is entering individual parameters
			contextHelp := "\n🎯 Current progress: "
			switch paramCount {
			case 2:
				contextHelp += "Connection name ✓, checking if parameter 2 is URL or host..."
			case 3:
				contextHelp += "Name ✓, Host ✓ - Next: port number"
			case 4:
				contextHelp += "Name ✓, Host ✓, Port ✓ - Next: database name"
			case 5:
				contextHelp += "Name ✓, Host ✓, Port ✓, Database ✓ - Next: username"
			case 6:
				contextHelp += "Name ✓, Host ✓, Port ✓, Database ✓, Username ✓ - Next: password"
			case 7:
				contextHelp += "Almost done! - Next: SSL mode (disable/require/verify-ca/verify-full)"
			default:
				contextHelp += "All parameters provided - press Enter to create connection"
			}
			return baseHelp + contextHelp
		}
	}

	return cmd.ParameterHelp
}

// IsCompleteCommand checks if the input is a complete command waiting for parameters
func (ch *CommandHandler) IsCompleteCommand(input string) bool {
	// Check if input is exactly a command name or command name with trailing space
	trimmed := strings.TrimSpace(input)

	// Exact command match
	if _, exists := ch.commands[trimmed]; exists {
		return true
	}

	// Command with trailing space (ready for parameters)
	if strings.HasSuffix(input, " ") {
		parts := strings.Fields(input)
		if len(parts) > 0 {
			if _, exists := ch.commands[parts[0]]; exists {
				return true
			}
		}
	}

	return false
}

// IsCommandWithPartialParams checks if input is a command with some parameters already entered
func (ch *CommandHandler) IsCommandWithPartialParams(input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	// Check if first part is a valid command
	if _, exists := ch.commands[parts[0]]; exists {
		// If we have more than just the command, it's a command with partial parameters
		return len(parts) > 1
	}

	return false
}

// handleAddConnection handles the /add command for adding database connections
func (ch *CommandHandler) handleAddConnection(input string) *CommandResult {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		message := "Usage: /add <connection_name> [connection_url|host port database username password ssl_mode]\n" +
			"Examples:\n" +
			"  /add mydb postgres://user:pass@localhost:5432/mydb?sslmode=disable\n" +
			"  /add mydb localhost 5432 postgres user pass disable\n" +
			"  /add mydb (interactive mode)"
		return &CommandResult{Message: message, Success: true}
	}

	name := parts[1]

	// Check if connection already exists
	connections := ch.connService.GetConnectionManager().ListConnections()
	if _, exists := connections[name]; exists {
		message := fmt.Sprintf("Error: Connection '%s' already exists. Use /remove first or choose a different name.", name)
		return &CommandResult{Message: message, Success: false}
	}

	var config *database.ConnectionConfig

	if len(parts) == 2 {
		// Only connection name provided - ask for more information
		message := fmt.Sprintf("Connection name '%s' is available.\n"+
			"Please provide connection details:\n\n"+
			"Option 1 - Database URL:\n"+
			"  /add %s postgres://user:pass@host:port/database?sslmode=disable\n\n"+
			"Option 2 - Individual parameters:\n"+
			"  /add %s host port database username password ssl_mode\n\n"+
			"Option 3 - Use defaults (localhost:5432/postgres):\n"+
			"  /add %s localhost 5432 postgres postgres \"\" disable", name, name, name, name)
		return &CommandResult{Message: message, Success: true}
	} else if len(parts) == 3 {
		// Check if the second parameter is a database URL
		potentialURL := parts[2]
		if database.ValidateDatabaseURL(potentialURL) {
			// Parse database URL
			parsedConfig, err := database.ParseDatabaseURL(potentialURL)
			if err != nil {
				message := fmt.Sprintf("Error: Failed to parse database URL: %v", err)
				return &CommandResult{Message: message, Error: err, Success: false}
			}

			config = parsedConfig
			config.Name = name
			config.Description = fmt.Sprintf("Added via CLI URL on %s", time.Now().Format("2006-01-02 15:04:05"))
		} else {
			message := "Error: Invalid database URL format. Expected: postgres://user:pass@host:port/database?sslmode=disable"
			return &CommandResult{Message: message, Success: false}
		}
	} else if len(parts) >= 7 {
		// Parse individual parameters
		host := parts[2]
		port := 5432
		if p, err := fmt.Sscanf(parts[3], "%d", &port); err != nil || p != 1 {
			port = 5432
		}
		dbName := parts[4]
		username := parts[5]
		password := parts[6]

		// Handle empty password strings
		if password == "\"\"" || password == "''" {
			password = ""
		}
		sslMode := "disable"
		if len(parts) > 7 {
			sslMode = parts[7]
		}

		config = &database.ConnectionConfig{
			Name:        name,
			Host:        host,
			Port:        port,
			Database:    dbName,
			Username:    username,
			Password:    password,
			SSLMode:     sslMode,
			Description: fmt.Sprintf("Added via CLI parameters on %s", time.Now().Format("2006-01-02 15:04:05")),
		}
	} else {
		// Invalid parameter count
		message := "Error: Invalid parameters. Usage:\n" +
			"  /add <name> postgres://user:pass@host:port/db\n" +
			"  /add <name> host port database username password ssl_mode\n" +
			"  /add <name> (for defaults: localhost:5432/postgres)"
		return &CommandResult{Message: message, Success: false}
	}

	// Try to add the connection
	if err := ch.connService.AddConnection(config); err != nil {
		message := fmt.Sprintf("Error: Failed to add connection '%s': %v", name, err)
		return &CommandResult{Message: message, Error: err, Success: false}
	}

	message := fmt.Sprintf("Successfully added connection '%s' (%s:%d/%s)",
		name, config.Host, config.Port, config.Database)
	return &CommandResult{Message: message, Success: true}
}

// handleListConnections handles the /list command
func (ch *CommandHandler) handleListConnections() *CommandResult {
	connections := ch.connService.GetConnectionManager().ListConnections()
	status := ch.connService.GetConnectionManager().GetConnectionStatus()

	if len(connections) == 0 {
		message := "No database connections configured.\nUse '/add <name>' to add a connection."
		return &CommandResult{Message: message, Success: true}
	}

	var lines []string
	lines = append(lines, "Database Connections:")
	lines = append(lines, "")

	// Get connections sorted by last used time for better display order
	sortedNames := ch.connService.GetConnectionManager().GetConnectionsSortedByLastUsed()

	for _, name := range sortedNames {
		config, exists := connections[name]
		if !exists {
			continue
		}

		statusText := "disconnected"
		if s, exists := status[name]; exists {
			switch s {
			case "active":
				statusText = "active"
			case "connected":
				statusText = "connected"
			case "unhealthy":
				statusText = "unhealthy"
			}
		}

		// Format last used time
		lastUsedText := ""
		if config.LastUsed != "" {
			if parsedTime, err := time.Parse(time.RFC3339, config.LastUsed); err == nil {
				lastUsedText = fmt.Sprintf(" (last used: %s)", parsedTime.Format("2006-01-02 15:04"))
			}
		}

		line := fmt.Sprintf("  %s - %s:%d/%s (%s) [%s]%s",
			name, config.Host, config.Port, config.Database, config.Username, statusText, lastUsedText)
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, "Commands: /switch <name> | /remove <name> | /status")
	message := strings.Join(lines, "\n")
	return &CommandResult{Message: message, Success: true}
}

// handleSwitchConnection handles the /switch command
func (ch *CommandHandler) handleSwitchConnection(input string) *CommandResult {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		message := "Usage: /switch <connection_name>\nUse '/list' to see available connections."
		return &CommandResult{Message: message, Success: true}
	}

	name := parts[1]
	if err := ch.connService.SwitchConnection(name); err != nil {
		message := fmt.Sprintf("Error: Failed to switch to connection '%s': %v", name, err)
		return &CommandResult{Message: message, Error: err, Success: false}
	}

	message := fmt.Sprintf("Switched to connection '%s'", name)
	return &CommandResult{Message: message, Success: true}
}

// handleRemoveConnection handles the /remove command
func (ch *CommandHandler) handleRemoveConnection(input string) *CommandResult {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		message := "Usage: /remove <connection_name>\nUse '/list' to see available connections."
		return &CommandResult{Message: message, Success: true}
	}

	name := parts[1]
	connections := ch.connService.GetConnectionManager().ListConnections()
	if _, exists := connections[name]; !exists {
		message := fmt.Sprintf("Error: Connection '%s' not found.", name)
		return &CommandResult{Message: message, Success: false}
	}

	if err := ch.connService.GetConnectionManager().RemoveConnection(name); err != nil {
		message := fmt.Sprintf("Error: Failed to remove connection '%s': %v", name, err)
		return &CommandResult{Message: message, Error: err, Success: false}
	}

	message := fmt.Sprintf("Successfully removed connection '%s'", name)
	return &CommandResult{Message: message, Success: true}
}

// handleConnectionStatus handles the /status command
func (ch *CommandHandler) handleConnectionStatus() *CommandResult {
	_, current, err := ch.connService.GetConnectionManager().GetCurrentConnection()
	status := ch.connService.GetConnectionManager().GetConnectionStatus()
	connections := ch.connService.GetConnectionManager().ListConnections()

	var lines []string
	lines = append(lines, "Connection Status:")
	lines = append(lines, "")

	if err != nil {
		lines = append(lines, "No active connection")
	} else {
		lines = append(lines, fmt.Sprintf("Current: %s", current))
	}

	// Show last used connection info
	lastUsedName := ch.connService.GetConnectionManager().GetLastUsedConnection()
	if lastUsedName != "" && lastUsedName != current {
		lines = append(lines, fmt.Sprintf("Last used: %s", lastUsedName))
	}

	lines = append(lines, "")
	lines = append(lines, "All connections (sorted by last used):")

	// Get connections sorted by last used time
	sortedNames := ch.connService.GetConnectionManager().GetConnectionsSortedByLastUsed()

	for _, name := range sortedNames {
		if _, exists := connections[name]; !exists {
			continue
		}

		statusText := "disconnected"
		if s, exists := status[name]; exists {
			switch s {
			case "active":
				statusText = "active (current)"
			case "connected":
				statusText = "connected"
			case "unhealthy":
				statusText = "unhealthy"
			}
		}

		// Add last used time info
		config := connections[name]
		lastUsedText := ""
		if config.LastUsed != "" {
			if parsedTime, err := time.Parse(time.RFC3339, config.LastUsed); err == nil {
				lastUsedText = fmt.Sprintf(" - last used: %s", parsedTime.Format("2006-01-02 15:04"))
			}
		}

		lines = append(lines, fmt.Sprintf("  %s [%s]%s", name, statusText, lastUsedText))
	}

	message := strings.Join(lines, "\n")
	return &CommandResult{Message: message, Success: true}
}
