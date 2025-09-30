package sqlite

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"dbsage/pkg/dbinterfaces"
)

// SQLiteProvider implements the DatabaseProviderInterface for SQLite
type SQLiteProvider struct{}

// Ensure SQLiteProvider implements DatabaseProviderInterface
var _ dbinterfaces.DatabaseProviderInterface = (*SQLiteProvider)(nil)

// NewSQLiteProvider creates a new SQLite provider
func NewSQLiteProvider() *SQLiteProvider {
	return &SQLiteProvider{}
}

// CreateConnection creates a new SQLite database connection
func (p *SQLiteProvider) CreateConnection(config *dbinterfaces.ConnectionConfig) (dbinterfaces.DatabaseInterface, error) {
	// Convert to local config type
	localConfig := &ConnectionConfig{
		Name:        config.Name,
		Database:    config.Database,
		Mode:        "rwc",    // Default mode: read-write-create
		Cache:       "shared", // Default cache mode
		Timeout:     "30s",    // Default timeout
		Description: config.Description,
		LastUsed:    config.LastUsed,
	}

	if err := p.validateLocalConfig(localConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	connURL := p.buildLocalConnectionURL(localConfig)
	return NewSQLiteDatabase(connURL)
}

// GetSupportedDrivers returns the supported SQLite drivers
func (p *SQLiteProvider) GetSupportedDrivers() []string {
	return []string{"sqlite3"}
}

// ValidateConfig validates the SQLite connection configuration (using generic interface)
func (p *SQLiteProvider) ValidateConfig(config *dbinterfaces.ConnectionConfig) error {
	return p.validateLocalConfig(&ConnectionConfig{
		Name:        config.Name,
		Database:    config.Database,
		Mode:        "rwc",
		Cache:       "shared",
		Timeout:     "30s",
		Description: config.Description,
		LastUsed:    config.LastUsed,
	})
}

// validateLocalConfig validates the local SQLite connection configuration
func (p *SQLiteProvider) validateLocalConfig(config *ConnectionConfig) error {
	if config.Database == "" {
		return fmt.Errorf("database file path is required")
	}

	// Validate mode
	if config.Mode == "" {
		config.Mode = "rwc" // Default mode
	}
	validModes := []string{"ro", "rw", "rwc", "memory"}
	isValidMode := false
	for _, mode := range validModes {
		if config.Mode == mode {
			isValidMode = true
			break
		}
	}
	if !isValidMode {
		return fmt.Errorf("invalid mode '%s', must be one of: %s", config.Mode, strings.Join(validModes, ", "))
	}

	// Validate cache
	if config.Cache == "" {
		config.Cache = "shared" // Default cache
	}
	validCaches := []string{"shared", "private"}
	isValidCache := false
	for _, cache := range validCaches {
		if config.Cache == cache {
			isValidCache = true
			break
		}
	}
	if !isValidCache {
		return fmt.Errorf("invalid cache '%s', must be one of: %s", config.Cache, strings.Join(validCaches, ", "))
	}

	if config.Timeout == "" {
		config.Timeout = "30s" // Default timeout
	}

	return nil
}

// BuildConnectionURL builds a SQLite connection URL from the configuration (using generic interface)
func (p *SQLiteProvider) BuildConnectionURL(config *dbinterfaces.ConnectionConfig) string {
	return p.buildLocalConnectionURL(&ConnectionConfig{
		Name:        config.Name,
		Database:    config.Database,
		Mode:        "rwc",
		Cache:       "shared",
		Timeout:     "30s",
		Description: config.Description,
		LastUsed:    config.LastUsed,
	})
}

// buildLocalConnectionURL builds a SQLite connection URL from the local configuration
func (p *SQLiteProvider) buildLocalConnectionURL(config *ConnectionConfig) string {
	// Handle special cases
	if config.Database == ":memory:" {
		return "file::memory:?cache=shared"
	}

	// Clean the file path
	dbPath := filepath.Clean(config.Database)

	// SQLite DSN format: file:path?param1=value1&param2=value2
	dsn := fmt.Sprintf("file:%s", url.QueryEscape(dbPath))

	// Add SQLite specific parameters
	params := make(map[string]string)
	params["mode"] = config.Mode
	params["cache"] = config.Cache
	params["_timeout"] = config.Timeout
	params["_journal_mode"] = "WAL"   // Use WAL mode for better concurrency
	params["_synchronous"] = "NORMAL" // Balance between safety and performance
	params["_foreign_keys"] = "on"    // Enable foreign key constraints

	// Build query string
	var paramPairs []string
	for key, value := range params {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", key, url.QueryEscape(value)))
	}

	if len(paramPairs) > 0 {
		dsn += "?" + strings.Join(paramPairs, "&")
	}

	return dsn
}
