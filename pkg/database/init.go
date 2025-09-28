package database

import (
	"dbsage/pkg/dbinterfaces"
)

// NewDefaultConnectionService creates a connection service with all supported providers
func NewDefaultConnectionService() dbinterfaces.ConnectionServiceInterface {
	return NewConnectionService()
}

// GetSupportedDatabaseTypes returns all supported database types
func GetSupportedDatabaseTypes() []string {
	pm := NewProviderManager()
	return pm.GetSupportedTypesAsStrings()
}
