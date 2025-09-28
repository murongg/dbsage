package queries

import (
	"database/sql"
	"fmt"
	"time"

	"dbsage/internal/models"
)

// Executor handles SQL query execution
type Executor struct {
	db *sql.DB
}

func NewExecutor(db *sql.DB) *Executor {
	return &Executor{db: db}
}

// ExecuteSQL executes a SQL query and returns structured results
func (e *Executor) ExecuteSQL(query string) (*models.QueryResult, error) {
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
func (e *Executor) ExplainQuery(query string) (*models.QueryResult, error) {
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)
	return e.ExecuteSQL(explainQuery)
}
