package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"dbsage/pkg/dbinterfaces"
)

// ConnectionManager manages multiple database connections
type ConnectionManager struct {
	connections     map[string]dbinterfaces.DatabaseInterface
	configs         map[string]*dbinterfaces.ConnectionConfig
	providerManager *ProviderManager
	current         string
	mu              sync.RWMutex
	configFile      string
}

// Ensure ConnectionManager implements ConnectionManagerInterface
var _ dbinterfaces.ConnectionManagerInterface = (*ConnectionManager)(nil)

// NewConnectionManager creates a new connection manager
func NewConnectionManager() dbinterfaces.ConnectionManagerInterface {
	homeDir, _ := os.UserHomeDir()
	configFile := filepath.Join(homeDir, ".dbsage", "connections.json")

	cm := &ConnectionManager{
		connections:     make(map[string]dbinterfaces.DatabaseInterface),
		configs:         make(map[string]*dbinterfaces.ConnectionConfig),
		providerManager: NewProviderManager(),
		configFile:      configFile,
	}

	// Load existing connections
	if err := cm.loadConnections(); err != nil {
		// Log error but don't fail initialization
		// This allows the manager to work even if the config file is corrupted
		log.Printf("Warning: failed to load connections: %v", err)
	}
	return cm
}

// AddConnection adds a new database connection
func (cm *ConnectionManager) AddConnection(config *dbinterfaces.ConnectionConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create connection using the provider manager
	dbInterface, err := cm.providerManager.CreateConnection(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Store connection
	cm.connections[config.Name] = dbInterface
	cm.configs[config.Name] = config

	// Set as current if it's the first connection
	if cm.current == "" {
		cm.current = config.Name
		// Update last used time for the new current connection
		config.LastUsed = time.Now().Format(time.RFC3339)
	}

	// Save to file
	return cm.saveConnections()
}

// RemoveConnection removes a database connection
func (cm *ConnectionManager) RemoveConnection(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Close connection if exists
	if conn, exists := cm.connections[name]; exists {
		conn.Close()
		delete(cm.connections, name)
	}

	delete(cm.configs, name)

	// Update current if needed
	if cm.current == name {
		cm.current = ""
		// Set to first available connection
		for n := range cm.configs {
			cm.current = n
			break
		}
	}

	return cm.saveConnections()
}

// ListConnections returns all connection configurations
func (cm *ConnectionManager) ListConnections() map[string]*dbinterfaces.ConnectionConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]*dbinterfaces.ConnectionConfig)
	for name, config := range cm.configs {
		result[name] = config
	}
	return result
}

// GetCurrentConnection returns the current active connection
func (cm *ConnectionManager) GetCurrentConnection() (dbinterfaces.DatabaseInterface, string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.current == "" {
		return nil, "", fmt.Errorf("no active database connection")
	}

	conn, exists := cm.connections[cm.current]
	if !exists {
		return nil, "", fmt.Errorf("current connection '%s' not found", cm.current)
	}

	// Check connection health before returning
	if err := conn.CheckConnection(); err != nil {
		return nil, "", fmt.Errorf("current connection '%s' is unhealthy: %w", cm.current, err)
	}

	return conn, cm.current, nil
}

// SwitchConnection switches to a different connection
func (cm *ConnectionManager) SwitchConnection(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.configs[name]; !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}

	// Reconnect if connection doesn't exist or check health if it exists
	if conn, exists := cm.connections[name]; !exists {
		config := cm.configs[name]
		dbInterface, err := cm.providerManager.CreateConnection(config)
		if err != nil {
			return fmt.Errorf("failed to reconnect to database '%s': %w", name, err)
		}
		cm.connections[name] = dbInterface
	} else {
		// Check existing connection health
		if err := conn.CheckConnection(); err != nil {
			// Connection is unhealthy, try to reconnect
			config := cm.configs[name]

			// Close the unhealthy connection
			conn.Close()
			delete(cm.connections, name)

			// Create new connection
			dbInterface, err := cm.providerManager.CreateConnection(config)
			if err != nil {
				return fmt.Errorf("failed to reconnect to database '%s' after health check failure: %w", name, err)
			}
			cm.connections[name] = dbInterface
		}
	}

	cm.current = name

	// Update last used time
	if config, exists := cm.configs[name]; exists {
		config.LastUsed = time.Now().Format(time.RFC3339)
		// Save the updated configuration
		if err := cm.saveConnections(); err != nil {
			log.Printf("Warning: failed to save connections after switching: %v", err)
		}
	}

	return nil
}

// Close closes all connections
func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, conn := range cm.connections {
		conn.Close()
	}
	cm.connections = make(map[string]dbinterfaces.DatabaseInterface)
	return nil
}

// loadConnections loads connections from config file
func (cm *ConnectionManager) loadConnections() error {
	// Ensure config directory exists
	configDir := filepath.Dir(cm.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Read config file
	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's OK
		}
		return err
	}

	// Parse JSON
	var configs map[string]*dbinterfaces.ConnectionConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return err
	}

	// Store configs (don't auto-connect yet)
	cm.configs = configs

	// Set current to the most recently used connection if available
	if len(configs) > 0 && cm.current == "" {
		lastUsedName := cm.GetLastUsedConnection()
		if lastUsedName != "" {
			cm.current = lastUsedName
		} else {
			// If no last used connection, pick the first one
			for name := range configs {
				cm.current = name
				break
			}
		}
	}

	return nil
}

// saveConnections saves connections to config file
func (cm *ConnectionManager) saveConnections() error {
	// Ensure config directory exists
	configDir := filepath.Dir(cm.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cm.configs, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(cm.configFile, data, 0644)
}

// GetConnectionStatus returns status information about connections
func (cm *ConnectionManager) GetConnectionStatus() map[string]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := make(map[string]string)
	for name := range cm.configs {
		if conn, exists := cm.connections[name]; exists {
			// Check connection health
			if err := conn.CheckConnection(); err != nil {
				status[name] = "unhealthy"
			} else if name == cm.current {
				status[name] = "active"
			} else {
				status[name] = "connected"
			}
		} else {
			status[name] = "disconnected"
		}
	}
	return status
}

// GetLastUsedConnection returns the name of the most recently used connection
func (cm *ConnectionManager) GetLastUsedConnection() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var lastUsedName string
	var lastUsedTime time.Time

	for name, config := range cm.configs {
		if config.LastUsed != "" {
			if parsedTime, err := time.Parse(time.RFC3339, config.LastUsed); err == nil {
				if parsedTime.After(lastUsedTime) {
					lastUsedTime = parsedTime
					lastUsedName = name
				}
			}
		}
	}

	return lastUsedName
}

// GetConnectionsSortedByLastUsed returns connections sorted by last used time (most recent first)
func (cm *ConnectionManager) GetConnectionsSortedByLastUsed() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	type connectionWithTime struct {
		name     string
		lastUsed time.Time
	}

	var connections []connectionWithTime

	for name, config := range cm.configs {
		conn := connectionWithTime{name: name}
		if config.LastUsed != "" {
			if parsedTime, err := time.Parse(time.RFC3339, config.LastUsed); err == nil {
				conn.lastUsed = parsedTime
			}
		}
		connections = append(connections, conn)
	}

	// Sort by last used time (most recent first)
	sort.Slice(connections, func(i, j int) bool {
		return connections[i].lastUsed.After(connections[j].lastUsed)
	})

	var result []string
	for _, conn := range connections {
		result = append(result, conn.name)
	}

	return result
}
