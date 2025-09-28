package stats

import (
	"database/sql"
	"fmt"

	"dbsage/internal/models"
	"dbsage/pkg/dbinterfaces"
)

// MySQLTableStatsCollector collects table statistics for MySQL
type MySQLTableStatsCollector struct {
	db *sql.DB
}

// NewMySQLTableStatsCollector creates a new MySQL table stats collector
func NewMySQLTableStatsCollector(db *sql.DB) *MySQLTableStatsCollector {
	return &MySQLTableStatsCollector{db: db}
}

// Ensure MySQLTableStatsCollector implements TableStatsCollectorInterface
var _ dbinterfaces.TableStatsCollectorInterface = (*MySQLTableStatsCollector)(nil)

// GetTableStats returns detailed statistics for a specific table
func (c *MySQLTableStatsCollector) GetTableStats(tableName string) (*models.TableStats, error) {
	// MySQL doesn't have the same detailed stats as PostgreSQL, so we'll use information_schema
	query := `
		SELECT 
			CONCAT(table_schema, '.', table_name) as table_name,
			table_rows as row_count,
			ROUND(((data_length + index_length) / 1024 / 1024), 2) as total_size_mb,
			ROUND((data_length / 1024 / 1024), 2) as data_size_mb,
			ROUND((index_length / 1024 / 1024), 2) as index_size_mb,
			create_time,
			update_time
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() AND table_name = ?
	`

	var stats models.TableStats
	var createTime, updateTime sql.NullTime
	var totalSizeMB, dataSizeMB, indexSizeMB sql.NullFloat64

	err := c.db.QueryRow(query, tableName).Scan(
		&stats.TableName,
		&stats.RowCount,
		&totalSizeMB,
		&dataSizeMB,
		&indexSizeMB,
		&createTime,
		&updateTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get table stats: %w", err)
	}

	// Convert sizes to human readable format
	if totalSizeMB.Valid {
		stats.TotalSize = fmt.Sprintf("%.2f MB", totalSizeMB.Float64)
	}
	if dataSizeMB.Valid {
		stats.TableSize = fmt.Sprintf("%.2f MB", dataSizeMB.Float64)
	}
	if indexSizeMB.Valid {
		stats.IndexSize = fmt.Sprintf("%.2f MB", indexSizeMB.Float64)
	}

	// Convert times to strings
	if updateTime.Valid {
		stats.LastAnalyze = updateTime.Time.Format("2006-01-02 15:04:05")
	}

	// MySQL doesn't have direct equivalents for these PostgreSQL-specific stats
	// We'll set them to 0 or empty
	stats.SeqScan = 0
	stats.SeqTupRead = 0
	stats.IdxScan = 0
	stats.IdxTupFetch = 0
	stats.LastVacuum = ""

	return &stats, nil
}

// GetTableSizes returns size information for all tables
func (c *MySQLTableStatsCollector) GetTableSizes() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			table_schema,
			table_name,
			ROUND(((data_length + index_length) / 1024 / 1024), 2) as size_mb,
			(data_length + index_length) as size_bytes
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
		ORDER BY (data_length + index_length) DESC
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table sizes: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var schema, table string
		var sizeMB float64
		var sizeBytes int64

		err := rows.Scan(&schema, &table, &sizeMB, &sizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table size row: %w", err)
		}

		results = append(results, map[string]interface{}{
			"schema":     schema,
			"table_name": table,
			"size":       fmt.Sprintf("%.2f MB", sizeMB),
			"size_bytes": sizeBytes,
		})
	}

	return results, nil
}

// GetSlowQueries returns slow query information
func (c *MySQLTableStatsCollector) GetSlowQueries() ([]models.SlowQuery, error) {
	// MySQL 5.7+ has performance_schema for slow queries
	query := `
		SELECT 
			digest_text as query_text,
			count_star as calls,
			sum_timer_wait / 1000000000 as total_time_seconds,
			avg_timer_wait / 1000000000 as avg_time_seconds,
			min_timer_wait / 1000000000 as min_time_seconds,
			max_timer_wait / 1000000000 as max_time_seconds,
			stddev_timer_wait / 1000000000 as stddev_time_seconds,
			sum_rows_examined as rows_examined
		FROM performance_schema.events_statements_summary_by_digest 
		WHERE digest_text IS NOT NULL
		ORDER BY sum_timer_wait DESC 
		LIMIT 20
	`

	rows, err := c.db.Query(query)
	if err != nil {
		// If performance_schema is not available, return empty results
		return []models.SlowQuery{}, nil
	}
	defer rows.Close()

	var results []models.SlowQuery
	for rows.Next() {
		var query models.SlowQuery

		err := rows.Scan(
			&query.Query,
			&query.Calls,
			&query.TotalTime,
			&query.MeanTime,
			&query.MinTime,
			&query.MaxTime,
			&query.StddevTime,
			&query.Rows,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slow query row: %w", err)
		}

		results = append(results, query)
	}

	return results, nil
}
