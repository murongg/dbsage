package stats

import (
	"database/sql"
	"fmt"
	"os"

	"dbsage/internal/models"
	"dbsage/pkg/dbinterfaces"
)

// SQLiteTableStatsCollector collects table statistics for SQLite
type SQLiteTableStatsCollector struct {
	db *sql.DB
}

// NewSQLiteTableStatsCollector creates a new SQLite table stats collector
func NewSQLiteTableStatsCollector(db *sql.DB) *SQLiteTableStatsCollector {
	return &SQLiteTableStatsCollector{db: db}
}

// Ensure SQLiteTableStatsCollector implements TableStatsCollectorInterface
var _ dbinterfaces.TableStatsCollectorInterface = (*SQLiteTableStatsCollector)(nil)

// GetTableStats returns detailed statistics for a specific table
func (c *SQLiteTableStatsCollector) GetTableStats(tableName string) (*models.TableStats, error) {
	var stats models.TableStats
	stats.TableName = tableName

	// Get row count
	rowCountQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := c.db.QueryRow(rowCountQuery).Scan(&stats.RowCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get row count: %w", err)
	}

	// Get database file path to calculate size
	var dbPath string
	err = c.db.QueryRow("PRAGMA database_list").Scan(nil, nil, &dbPath)
	if err == nil && dbPath != "" && dbPath != ":memory:" {
		if fileInfo, err := os.Stat(dbPath); err == nil {
			sizeBytes := fileInfo.Size()
			sizeMB := float64(sizeBytes) / (1024 * 1024)
			stats.TotalSize = fmt.Sprintf("%.2f MB", sizeMB)
			stats.TableSize = fmt.Sprintf("%.2f MB", sizeMB) // SQLite doesn't separate table/index sizes easily
		}
	}

	// SQLite doesn't have direct equivalents for PostgreSQL-specific stats
	// We'll set them to 0 or empty
	stats.SeqScan = 0
	stats.SeqTupRead = 0
	stats.IdxScan = 0
	stats.IdxTupFetch = 0
	stats.IndexSize = "0 MB"
	stats.LastVacuum = ""
	stats.LastAnalyze = ""

	return &stats, nil
}

// GetTableSizes returns size information for all tables
func (c *SQLiteTableStatsCollector) GetTableSizes() ([]map[string]interface{}, error) {
	// Get all table names
	query := `
		SELECT name 
		FROM sqlite_master 
		WHERE type = 'table' 
		AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}

	// Get database file size once
	var totalSizeBytes int64
	var dbPath string
	err = c.db.QueryRow("PRAGMA database_list").Scan(nil, nil, &dbPath)
	if err == nil && dbPath != "" && dbPath != ":memory:" {
		if fileInfo, err := os.Stat(dbPath); err == nil {
			totalSizeBytes = fileInfo.Size()
		}
	}

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		// Get row count for each table
		var rowCount int64
		rowCountQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		err = c.db.QueryRow(rowCountQuery).Scan(&rowCount)
		if err != nil {
			// If we can't get row count, skip this table
			continue
		}

		// Estimate table size based on row count proportion
		// This is a rough estimation since SQLite doesn't provide per-table sizes
		estimatedSizeBytes := totalSizeBytes / 10 // Rough estimation
		sizeMB := float64(estimatedSizeBytes) / (1024 * 1024)

		result := map[string]interface{}{
			"table_name": tableName,
			"size_mb":    fmt.Sprintf("%.2f", sizeMB),
			"size_bytes": estimatedSizeBytes,
			"row_count":  rowCount,
		}
		results = append(results, result)
	}

	return results, nil
}

// GetSlowQueries returns slow query information (limited in SQLite)
func (c *SQLiteTableStatsCollector) GetSlowQueries() ([]models.SlowQuery, error) {
	// SQLite doesn't have a built-in slow query log like MySQL/PostgreSQL
	// We'll return an empty slice with a note
	return []models.SlowQuery{}, nil
}
