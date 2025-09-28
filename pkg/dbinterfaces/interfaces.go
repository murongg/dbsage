package dbinterfaces

import (
	"dbsage/internal/models"
)

// DatabaseInterface defines the core interface that all database implementations must satisfy
type DatabaseInterface interface {
	// Connection management
	Close() error
	IsConnectionHealthy() bool
	CheckConnection() error

	// Query execution
	ExecuteSQL(query string) (*models.QueryResult, error)
	ExplainQuery(query string) (*models.QueryResult, error)

	// Schema information
	GetAllTables() ([]models.TableInfo, error)
	GetTableSchema(tableName string) ([]models.ColumnInfo, error)
	GetTableIndexes(tableName string) ([]models.IndexInfo, error)

	// Table operations
	FindDuplicateData(tableName string, columns []string) (*models.QueryResult, error)

	// Statistics
	GetTableStats(tableName string) (*models.TableStats, error)
	GetTableSizes() ([]map[string]interface{}, error)
	GetSlowQueries() ([]models.SlowQuery, error)
	GetDatabaseSize() (*models.DatabaseSize, error)
	GetActiveConnections() ([]models.ActiveConnection, error)
}

// QueryExecutorInterface defines the interface for query execution
type QueryExecutorInterface interface {
	ExecuteSQL(query string) (*models.QueryResult, error)
	ExplainQuery(query string) (*models.QueryResult, error)
}

// TableStatsCollectorInterface defines the interface for table statistics collection
type TableStatsCollectorInterface interface {
	GetTableStats(tableName string) (*models.TableStats, error)
	GetTableSizes() ([]map[string]interface{}, error)
	GetSlowQueries() ([]models.SlowQuery, error)
}

// DatabaseStatsCollectorInterface defines the interface for database-level statistics collection
type DatabaseStatsCollectorInterface interface {
	GetDatabaseSize() (*models.DatabaseSize, error)
	GetActiveConnections() ([]models.ActiveConnection, error)
}

// ConnectionConfig represents a database connection configuration
type ConnectionConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // Database type: postgresql, mysql, sqlite, etc.
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	SSLMode     string `json:"ssl_mode"`
	Description string `json:"description"`
	LastUsed    string `json:"last_used,omitempty"` // ISO 8601 timestamp
}

// DatabaseProviderInterface defines the interface for database providers
type DatabaseProviderInterface interface {
	// CreateConnection creates a new database connection using the provided configuration
	CreateConnection(config *ConnectionConfig) (DatabaseInterface, error)

	// GetSupportedDrivers returns a list of supported database drivers
	GetSupportedDrivers() []string

	// ValidateConfig validates the connection configuration
	ValidateConfig(config *ConnectionConfig) error

	// BuildConnectionURL builds a connection URL from the configuration
	BuildConnectionURL(config *ConnectionConfig) string
}

// ConnectionManagerInterface defines the interface for connection management
type ConnectionManagerInterface interface {
	AddConnection(config *ConnectionConfig) error
	RemoveConnection(name string) error
	ListConnections() map[string]*ConnectionConfig
	GetCurrentConnection() (DatabaseInterface, string, error)
	SwitchConnection(name string) error
	Close() error
	GetConnectionStatus() map[string]string
	GetLastUsedConnection() string
	GetConnectionsSortedByLastUsed() []string
}

// ConnectionServiceInterface defines the interface for connection service
type ConnectionServiceInterface interface {
	GetCurrentTools() DatabaseInterface
	GetConnectionManager() ConnectionManagerInterface
	AddConnection(config *ConnectionConfig) error
	SwitchConnection(name string) error
	RemoveConnection(name string) error
	GetConnectionInfo() (map[string]*ConnectionConfig, map[string]string, string)
	TestConnection(config *ConnectionConfig) error
	Close() error
	IsConnected() bool
	IsConnectionHealthy() bool
	EnsureHealthyConnection() error
	GetConnectionStats() map[string]interface{}
}
