package stats

import (
	"database/sql"
	"fmt"

	"dbsage/internal/models"
)

// DatabaseStatsCollector collects database-level statistics
type DatabaseStatsCollector struct {
	db *sql.DB
}

func NewDatabaseStatsCollector(db *sql.DB) *DatabaseStatsCollector {
	return &DatabaseStatsCollector{db: db}
}

// GetDatabaseSize returns database size information
func (c *DatabaseStatsCollector) GetDatabaseSize() (*models.DatabaseSize, error) {
	query := `
		SELECT 
			current_database() as database_name,
			pg_size_pretty(pg_database_size(current_database())) as size,
			pg_database_size(current_database()) as size_bytes
	`

	var result models.DatabaseSize
	err := c.db.QueryRow(query).Scan(
		&result.DatabaseName,
		&result.Size,
		&result.SizeBytes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}

	return &result, nil
}

// GetActiveConnections returns information about active connections
func (c *DatabaseStatsCollector) GetActiveConnections() ([]models.ActiveConnection, error) {
	query := `
		SELECT 
			pid,
			usename,
			datname,
			client_addr,
			state,
			query,
			now() - query_start as duration
		FROM pg_stat_activity 
		WHERE state = 'active' AND pid != pg_backend_pid()
		ORDER BY query_start DESC
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active connections: %w", err)
	}
	defer rows.Close()

	var results []models.ActiveConnection
	for rows.Next() {
		var conn models.ActiveConnection
		var clientAddr sql.NullString
		var duration sql.NullString

		err := rows.Scan(
			&conn.PID,
			&conn.Username,
			&conn.Database,
			&clientAddr,
			&conn.State,
			&conn.Query,
			&duration,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection row: %w", err)
		}

		if clientAddr.Valid {
			conn.ClientAddr = clientAddr.String
		}
		if duration.Valid {
			conn.Duration = duration.String
		}

		results = append(results, conn)
	}

	return results, nil
}
