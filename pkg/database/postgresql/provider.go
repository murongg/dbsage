package postgresql

import (
	"fmt"
	"net/url"

	"dbsage/pkg/dbinterfaces"
)

// PostgreSQLProvider implements the DatabaseProviderInterface for PostgreSQL
type PostgreSQLProvider struct{}

// Ensure PostgreSQLProvider implements DatabaseProviderInterface
var _ dbinterfaces.DatabaseProviderInterface = (*PostgreSQLProvider)(nil)

// NewPostgreSQLProvider creates a new PostgreSQL provider
func NewPostgreSQLProvider() *PostgreSQLProvider {
	return &PostgreSQLProvider{}
}

// CreateConnection creates a new PostgreSQL database connection
func (p *PostgreSQLProvider) CreateConnection(config *dbinterfaces.ConnectionConfig) (dbinterfaces.DatabaseInterface, error) {
	// Convert to local config type
	localConfig := &ConnectionConfig{
		Name:        config.Name,
		Host:        config.Host,
		Port:        config.Port,
		Database:    config.Database,
		Username:    config.Username,
		Password:    config.Password,
		SSLMode:     config.SSLMode,
		Description: config.Description,
		LastUsed:    config.LastUsed,
	}

	if err := p.validateLocalConfig(localConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	connURL := p.buildLocalConnectionURL(localConfig)
	return NewPostgreSQLDatabase(connURL)
}

// GetSupportedDrivers returns the supported PostgreSQL drivers
func (p *PostgreSQLProvider) GetSupportedDrivers() []string {
	return []string{"postgres", "postgresql"}
}

// ValidateConfig validates the PostgreSQL connection configuration (using generic interface)
func (p *PostgreSQLProvider) ValidateConfig(config *dbinterfaces.ConnectionConfig) error {
	return p.validateLocalConfig(&ConnectionConfig{
		Name:        config.Name,
		Host:        config.Host,
		Port:        config.Port,
		Database:    config.Database,
		Username:    config.Username,
		Password:    config.Password,
		SSLMode:     config.SSLMode,
		Description: config.Description,
		LastUsed:    config.LastUsed,
	})
}

// validateLocalConfig validates the local PostgreSQL connection configuration
func (p *PostgreSQLProvider) validateLocalConfig(config *ConnectionConfig) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}
	if config.Port <= 0 {
		return fmt.Errorf("port must be greater than 0")
	}
	if config.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Username == "" {
		return fmt.Errorf("username is required")
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable" // Default SSL mode
	}
	return nil
}

// BuildConnectionURL builds a PostgreSQL connection URL from the configuration (using generic interface)
func (p *PostgreSQLProvider) BuildConnectionURL(config *dbinterfaces.ConnectionConfig) string {
	return p.buildLocalConnectionURL(&ConnectionConfig{
		Name:        config.Name,
		Host:        config.Host,
		Port:        config.Port,
		Database:    config.Database,
		Username:    config.Username,
		Password:    config.Password,
		SSLMode:     config.SSLMode,
		Description: config.Description,
		LastUsed:    config.LastUsed,
	})
}

// buildLocalConnectionURL builds a PostgreSQL connection URL from the local configuration
func (p *PostgreSQLProvider) buildLocalConnectionURL(config *ConnectionConfig) string {
	if config.Password == "" {
		return fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=%s",
			url.QueryEscape(config.Username), config.Host, config.Port,
			url.QueryEscape(config.Database), config.SSLMode)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(config.Username), url.QueryEscape(config.Password),
		config.Host, config.Port, url.QueryEscape(config.Database), config.SSLMode)
}
