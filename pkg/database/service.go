package database

import (
	"fmt"
	"log"
)

// ConnectionService provides a high-level interface for database connection management
type ConnectionService struct {
	manager *ConnectionManager
	current *DatabaseTools
}

// NewConnectionService creates a new connection service
func NewConnectionService() *ConnectionService {
	manager := NewConnectionManager()
	service := &ConnectionService{
		manager: manager,
	}

	// Try to establish initial connection
	service.initializeConnection()
	return service
}

// initializeConnection tries to establish an initial database connection
func (cs *ConnectionService) initializeConnection() {
	// Try to get the most recently used connection from manager
	lastUsedName := cs.manager.GetLastUsedConnection()

	if lastUsedName != "" {
		// Try to connect to the last used connection
		if err := cs.manager.SwitchConnection(lastUsedName); err == nil {
			if dbTools, name, err := cs.manager.GetCurrentConnection(); err == nil {
				cs.current = dbTools
				log.Printf("Reconnected to last used database: %s", name)
				return
			}
		}
		log.Printf("Failed to reconnect to last used database: %s, trying current connection", lastUsedName)
	}

	// Fallback to current connection from manager
	if dbTools, name, err := cs.manager.GetCurrentConnection(); err == nil {
		cs.current = dbTools
		log.Printf("Connected to database: %s", name)
	} else {
		log.Println("No database connections configured. Use '/add' command to add connections.")
	}
}

// GetCurrentTools returns the current database tools
func (cs *ConnectionService) GetCurrentTools() *DatabaseTools {
	// Check if current connection is healthy
	if cs.current != nil && !cs.current.IsConnectionHealthy() {
		// Try to refresh the current connection
		if dbTools, _, err := cs.manager.GetCurrentConnection(); err == nil {
			cs.current = dbTools
		} else {
			cs.current = nil
		}
	}
	return cs.current
}

// GetConnectionManager returns the connection manager
func (cs *ConnectionService) GetConnectionManager() *ConnectionManager {
	return cs.manager
}

// AddConnection adds a new database connection
func (cs *ConnectionService) AddConnection(config *ConnectionConfig) error {
	err := cs.manager.AddConnection(config)
	if err != nil {
		return err
	}

	// Update current connection if this is the first one or if requested
	if cs.current == nil {
		if dbTools, _, err := cs.manager.GetCurrentConnection(); err == nil {
			cs.current = dbTools
		}
	}

	return nil
}

// SwitchConnection switches to a different connection
func (cs *ConnectionService) SwitchConnection(name string) error {
	err := cs.manager.SwitchConnection(name)
	if err != nil {
		return err
	}

	// Update current tools
	if dbTools, _, err := cs.manager.GetCurrentConnection(); err == nil {
		cs.current = dbTools
		return nil
	}

	return fmt.Errorf("failed to update current connection after switch")
}

// RemoveConnection removes a database connection
func (cs *ConnectionService) RemoveConnection(name string) error {
	err := cs.manager.RemoveConnection(name)
	if err != nil {
		return err
	}

	// Update current tools if the removed connection was current
	if dbTools, _, err := cs.manager.GetCurrentConnection(); err == nil {
		cs.current = dbTools
	} else {
		cs.current = nil
	}

	return nil
}

// GetConnectionInfo returns information about all connections
func (cs *ConnectionService) GetConnectionInfo() (map[string]*ConnectionConfig, map[string]string, string) {
	connections := cs.manager.ListConnections()
	status := cs.manager.GetConnectionStatus()

	current := ""
	if _, name, err := cs.manager.GetCurrentConnection(); err == nil {
		current = name
	}

	return connections, status, current
}

// TestConnection tests a connection configuration without adding it
func (cs *ConnectionService) TestConnection(config *ConnectionConfig) error {
	// Build connection URL
	connURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.Username, config.Password, config.Host, config.Port,
		config.Database, config.SSLMode)

	// Test connection
	dbTools, err := NewDatabaseTools(connURL)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer dbTools.Close()

	return nil
}

// Close closes all connections
func (cs *ConnectionService) Close() error {
	if cs.current != nil {
		cs.current = nil
	}
	return cs.manager.Close()
}

// IsConnected returns whether there's an active database connection
func (cs *ConnectionService) IsConnected() bool {
	return cs.current != nil
}

// IsConnectionHealthy returns whether the current connection is healthy
func (cs *ConnectionService) IsConnectionHealthy() bool {
	if cs.current == nil {
		return false
	}
	return cs.current.IsConnectionHealthy()
}

// EnsureHealthyConnection ensures the current connection is healthy, attempts to reconnect if not
func (cs *ConnectionService) EnsureHealthyConnection() error {
	if cs.current == nil {
		return fmt.Errorf("no active database connection")
	}

	if !cs.current.IsConnectionHealthy() {
		// Try to get a healthy connection from the manager
		if dbTools, name, err := cs.manager.GetCurrentConnection(); err == nil {
			cs.current = dbTools
			log.Printf("Reconnected to database: %s", name)
			return nil
		} else {
			cs.current = nil
			return fmt.Errorf("failed to restore healthy connection: %w", err)
		}
	}

	return nil
}

// GetConnectionStats returns statistics about connections
func (cs *ConnectionService) GetConnectionStats() map[string]interface{} {
	connections := cs.manager.ListConnections()
	status := cs.manager.GetConnectionStatus()

	stats := map[string]interface{}{
		"total_connections":        len(connections),
		"active_connections":       0,
		"connected_connections":    0,
		"unhealthy_connections":    0,
		"disconnected_connections": 0,
		"has_current":              cs.current != nil,
		"current_is_healthy":       cs.IsConnectionHealthy(),
	}

	for _, s := range status {
		switch s {
		case "active":
			stats["active_connections"] = stats["active_connections"].(int) + 1
		case "connected":
			stats["connected_connections"] = stats["connected_connections"].(int) + 1
		case "unhealthy":
			stats["unhealthy_connections"] = stats["unhealthy_connections"].(int) + 1
		default:
			stats["disconnected_connections"] = stats["disconnected_connections"].(int) + 1
		}
	}

	return stats
}
