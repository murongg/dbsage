package queries

import (
	"database/sql"
	"fmt"
	"time"

	"dbsage/internal/models"
	"dbsage/pkg/dbinterfaces"
)

// MySQLExecutor handles SQL query execution for MySQL
type MySQLExecutor struct {
	db *sql.DB
}

// NewMySQLExecutor creates a new MySQL query executor
func NewMySQLExecutor(db *sql.DB) *MySQLExecutor {
	return &MySQLExecutor{db: db}
}

// Ensure MySQLExecutor implements QueryExecutorInterface
var _ dbinterfaces.QueryExecutorInterface = (*MySQLExecutor)(nil)

// ExecuteSQL executes a SQL query and returns structured results
func (e *MySQLExecutor) ExecuteSQL(query string) (*models.QueryResult, error) {
	start := time.Now()

	rows, err := e.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}

	// Prepare slice for row data
	var resultRows [][]interface{}

	// Create a slice to hold the values for each row
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Iterate through rows
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to proper types
		row := make([]interface{}, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = nil
			} else {
				switch v := val.(type) {
				case []byte:
					row[i] = string(v)
				default:
					row[i] = v
				}
			}
		}
		resultRows = append(resultRows, row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	duration := time.Since(start)

	return &models.QueryResult{
		Columns:  columns,
		Rows:     resultRows,
		RowCount: len(resultRows),
		Duration: duration.String(),
	}, nil
}

// ExplainQuery analyzes a query's execution plan
func (e *MySQLExecutor) ExplainQuery(query string) (*models.QueryResult, error) {
	explainQuery := fmt.Sprintf("EXPLAIN FORMAT=JSON %s", query)
	return e.ExecuteSQL(explainQuery)
}
