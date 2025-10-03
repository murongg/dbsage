package sqlite

import "dbsage/internal/models"

// QueryExecutorInterface defines the interface for query execution
type QueryExecutorInterface interface {
	ExecuteSQL(query string) (*models.QueryResult, error)
	ExplainQuery(query string) (*models.QueryResult, error)
}
