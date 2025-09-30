package database

import (
	"fmt"

	"dbsage/pkg/dbinterfaces"
)

// ProviderFactory creates database providers based on type
type ProviderFactory struct{}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateProvider creates a database provider based on the specified type
func (f *ProviderFactory) CreateProvider(dbType DatabaseType) (dbinterfaces.DatabaseProviderInterface, error) {
	switch dbType {
	case PostgreSQL:
		// We'll need to create the PostgreSQL provider in a separate package or function
		// to avoid import cycles
		return nil, fmt.Errorf("PostgreSQL provider creation moved to avoid import cycle")
	case MySQL:
		return nil, fmt.Errorf("MySQL provider not yet implemented")
	case SQLite:
		return nil, fmt.Errorf("SQLite provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// DetectDatabaseType attempts to detect the database type from configuration
func (f *ProviderFactory) DetectDatabaseType(config *dbinterfaces.ConnectionConfig) DatabaseType {
	// For now, we'll default to PostgreSQL
	// In the future, this could be enhanced to detect based on:
	// - Connection string format
	// - Port number (5432 for PostgreSQL, 3306 for MySQL, etc.)
	// - Explicit database type field in config
	return PostgreSQL
}

// GetSupportedTypes returns a list of supported database types
func (f *ProviderFactory) GetSupportedTypes() []DatabaseType {
	return []DatabaseType{PostgreSQL, MySQL, SQLite}
}

// NewDefaultConnectionService will be implemented in a separate file to avoid import cycles
// This function should be created in the main package or a separate initialization package
