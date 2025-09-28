package database

import (
	"fmt"
	"strings"

	"dbsage/pkg/database/mysql"
	"dbsage/pkg/database/postgresql"
	"dbsage/pkg/dbinterfaces"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
	SQLite     DatabaseType = "sqlite"
	MongoDB    DatabaseType = "mongodb"
)

// ProviderManager manages database providers for different database types
type ProviderManager struct {
	providers map[DatabaseType]dbinterfaces.DatabaseProviderInterface
}

// NewProviderManager creates a new provider manager with all supported providers
func NewProviderManager() *ProviderManager {
	pm := &ProviderManager{
		providers: make(map[DatabaseType]dbinterfaces.DatabaseProviderInterface),
	}

	// Register built-in providers
	pm.RegisterProvider(PostgreSQL, postgresql.NewPostgreSQLProvider())
	pm.RegisterProvider(MySQL, mysql.NewMySQLProvider())

	// TODO: Add other providers when implemented
	// pm.RegisterProvider(SQLite, sqlite.NewSQLiteProvider())
	// pm.RegisterProvider(MongoDB, mongodb.NewMongoDBProvider())

	return pm
}

// RegisterProvider registers a provider for a specific database type
func (pm *ProviderManager) RegisterProvider(dbType DatabaseType, provider dbinterfaces.DatabaseProviderInterface) {
	pm.providers[dbType] = provider
}

// GetProvider returns the provider for the specified database type
func (pm *ProviderManager) GetProvider(dbType DatabaseType) (dbinterfaces.DatabaseProviderInterface, error) {
	provider, exists := pm.providers[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	return provider, nil
}

// GetProviderByString returns the provider for the specified database type string
func (pm *ProviderManager) GetProviderByString(dbTypeStr string) (dbinterfaces.DatabaseProviderInterface, error) {
	dbType, err := ParseDatabaseType(dbTypeStr)
	if err != nil {
		return nil, err
	}
	return pm.GetProvider(dbType)
}

// CreateConnection creates a database connection using the appropriate provider based on config type
func (pm *ProviderManager) CreateConnection(config *dbinterfaces.ConnectionConfig) (dbinterfaces.DatabaseInterface, error) {
	if config.Type == "" {
		return nil, fmt.Errorf("database type is required in connection configuration")
	}

	provider, err := pm.GetProviderByString(config.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider for type %s: %w", config.Type, err)
	}

	return provider.CreateConnection(config)
}

// GetSupportedTypes returns all supported database types
func (pm *ProviderManager) GetSupportedTypes() []DatabaseType {
	types := make([]DatabaseType, 0, len(pm.providers))
	for dbType := range pm.providers {
		types = append(types, dbType)
	}
	return types
}

// GetSupportedTypesAsStrings returns all supported database types as strings
func (pm *ProviderManager) GetSupportedTypesAsStrings() []string {
	types := pm.GetSupportedTypes()
	result := make([]string, len(types))
	for i, dbType := range types {
		result[i] = string(dbType)
	}
	return result
}

// ValidateConfig validates a connection configuration
func (pm *ProviderManager) ValidateConfig(config *dbinterfaces.ConnectionConfig) error {
	if config.Type == "" {
		return fmt.Errorf("database type is required")
	}

	provider, err := pm.GetProviderByString(config.Type)
	if err != nil {
		return err
	}

	return provider.ValidateConfig(config)
}

// ParseDatabaseType parses a string to DatabaseType
func ParseDatabaseType(s string) (DatabaseType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "postgresql", "postgres", "pg":
		return PostgreSQL, nil
	case "mysql":
		return MySQL, nil
	case "sqlite", "sqlite3":
		return SQLite, nil
	case "mongodb", "mongo":
		return MongoDB, nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", s)
	}
}

// GetDatabaseTypeString returns the string representation of a database type
func GetDatabaseTypeString(dbType DatabaseType) string {
	return string(dbType)
}

// GetDefaultPort returns the default port for a database type
func GetDefaultPort(dbType DatabaseType) int {
	switch dbType {
	case PostgreSQL:
		return 5432
	case MySQL:
		return 3306
	case SQLite:
		return 0 // SQLite doesn't use ports
	case MongoDB:
		return 27017
	default:
		return 0
	}
}
