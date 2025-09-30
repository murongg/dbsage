package sqlite

import "dbsage/internal/models"

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
