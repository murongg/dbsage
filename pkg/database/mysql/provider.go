package mysql

import (
	"fmt"
	"net/url"
	"strings"

	"dbsage/pkg/dbinterfaces"
)

// MySQLProvider implements the DatabaseProviderInterface for MySQL
type MySQLProvider struct{}

// Ensure MySQLProvider implements DatabaseProviderInterface
var _ dbinterfaces.DatabaseProviderInterface = (*MySQLProvider)(nil)

// NewMySQLProvider creates a new MySQL provider
func NewMySQLProvider() *MySQLProvider {
	return &MySQLProvider{}
}

// CreateConnection creates a new MySQL database connection
func (p *MySQLProvider) CreateConnection(config *dbinterfaces.ConnectionConfig) (dbinterfaces.DatabaseInterface, error) {
	// Convert to local config type
	localConfig := &ConnectionConfig{
		Name:        config.Name,
		Host:        config.Host,
		Port:        config.Port,
		Database:    config.Database,
		Username:    config.Username,
		Password:    config.Password,
		Charset:     "utf8mb4",            // Default charset
		Collation:   "utf8mb4_unicode_ci", // Default collation
		Timeout:     "30s",                // Default timeout
		Description: config.Description,
		LastUsed:    config.LastUsed,
	}

	if err := p.validateLocalConfig(localConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	connURL := p.buildLocalConnectionURL(localConfig)
	return NewMySQLDatabase(connURL)
}

// GetSupportedDrivers returns the supported MySQL drivers
func (p *MySQLProvider) GetSupportedDrivers() []string {
	return []string{"mysql"}
}

// ValidateConfig validates the MySQL connection configuration (using generic interface)
func (p *MySQLProvider) ValidateConfig(config *dbinterfaces.ConnectionConfig) error {
	return p.validateLocalConfig(&ConnectionConfig{
		Name:        config.Name,
		Host:        config.Host,
		Port:        config.Port,
		Database:    config.Database,
		Username:    config.Username,
		Password:    config.Password,
		Charset:     "utf8mb4",
		Collation:   "utf8mb4_unicode_ci",
		Timeout:     "30s",
		Description: config.Description,
		LastUsed:    config.LastUsed,
	})
}

// validateLocalConfig validates the local MySQL connection configuration
func (p *MySQLProvider) validateLocalConfig(config *ConnectionConfig) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}
	if config.Port <= 0 {
		config.Port = 3306 // Default MySQL port
	}
	if config.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Username == "" {
		return fmt.Errorf("username is required")
	}
	if config.Charset == "" {
		config.Charset = "utf8mb4" // Default charset
	}
	if config.Collation == "" {
		config.Collation = "utf8mb4_unicode_ci" // Default collation
	}
	if config.Timeout == "" {
		config.Timeout = "30s" // Default timeout
	}
	return nil
}

// BuildConnectionURL builds a MySQL connection URL from the configuration (using generic interface)
func (p *MySQLProvider) BuildConnectionURL(config *dbinterfaces.ConnectionConfig) string {
	return p.buildLocalConnectionURL(&ConnectionConfig{
		Name:        config.Name,
		Host:        config.Host,
		Port:        config.Port,
		Database:    config.Database,
		Username:    config.Username,
		Password:    config.Password,
		Charset:     "utf8mb4",
		Collation:   "utf8mb4_unicode_ci",
		Timeout:     "30s",
		Description: config.Description,
		LastUsed:    config.LastUsed,
	})
}

// buildLocalConnectionURL builds a MySQL connection URL from the local configuration
func (p *MySQLProvider) buildLocalConnectionURL(config *ConnectionConfig) string {
	// MySQL DSN format: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		url.QueryEscape(config.Username),
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		url.QueryEscape(config.Database))

	// Add MySQL specific parameters
	params := make(map[string]string)
	params["charset"] = config.Charset
	params["collation"] = config.Collation
	params["timeout"] = config.Timeout
	params["readTimeout"] = config.Timeout
	params["writeTimeout"] = config.Timeout
	params["parseTime"] = "true"
	params["loc"] = "Local"

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
