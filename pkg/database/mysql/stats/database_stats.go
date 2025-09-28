package stats

import (
	"database/sql"
	"fmt"

	"dbsage/internal/models"
	"dbsage/pkg/dbinterfaces"
)

// MySQLDatabaseStatsCollector collects database-level statistics for MySQL
type MySQLDatabaseStatsCollector struct {
	db *sql.DB
}

// NewMySQLDatabaseStatsCollector creates a new MySQL database stats collector
func NewMySQLDatabaseStatsCollector(db *sql.DB) *MySQLDatabaseStatsCollector {
	return &MySQLDatabaseStatsCollector{db: db}
}

// Ensure MySQLDatabaseStatsCollector implements DatabaseStatsCollectorInterface
var _ dbinterfaces.DatabaseStatsCollectorInterface = (*MySQLDatabaseStatsCollector)(nil)

// GetDatabaseSize returns database size information
func (c *MySQLDatabaseStatsCollector) GetDatabaseSize() (*models.DatabaseSize, error) {
	query := `
		SELECT 
			schema_name as database_name,
			ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) as size_mb,
			SUM(data_length + index_length) as size_bytes
		FROM information_schema.tables 
		WHERE schema_name = DATABASE()
		GROUP BY schema_name
	`

	var result models.DatabaseSize
	var sizeMB float64

	err := c.db.QueryRow(query).Scan(
		&result.DatabaseName,
		&sizeMB,
		&result.SizeBytes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}

	result.Size = fmt.Sprintf("%.2f MB", sizeMB)

	return &result, nil
}

// GetActiveConnections returns information about active connections
func (c *MySQLDatabaseStatsCollector) GetActiveConnections() ([]models.ActiveConnection, error) {
	query := `
		SELECT 
			id as connection_id,
			user,
			db as database_name,
			host,
			command as state,
			COALESCE(info, '') as query,
			time as duration_seconds
		FROM information_schema.processlist 
		WHERE command != 'Sleep' AND id != CONNECTION_ID()
		ORDER BY time DESC
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active connections: %w", err)
	}
	defer rows.Close()

	var results []models.ActiveConnection
	for rows.Next() {
		var conn models.ActiveConnection
		var durationSeconds sql.NullInt64
		var database sql.NullString

		err := rows.Scan(
			&conn.PID,
			&conn.Username,
			&database,
			&conn.ClientAddr,
			&conn.State,
			&conn.Query,
			&durationSeconds,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection row: %w", err)
		}

		if database.Valid {
			conn.Database = database.String
		}
		if durationSeconds.Valid {
			conn.Duration = fmt.Sprintf("%d seconds", durationSeconds.Int64)
		}

		results = append(results, conn)
	}

	return results, nil
}
