package stats

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"dbsage/internal/models"
	"dbsage/pkg/dbinterfaces"
)

// SQLiteDatabaseStatsCollector collects database-level statistics for SQLite
type SQLiteDatabaseStatsCollector struct {
	db *sql.DB
}

// NewSQLiteDatabaseStatsCollector creates a new SQLite database stats collector
func NewSQLiteDatabaseStatsCollector(db *sql.DB) *SQLiteDatabaseStatsCollector {
	return &SQLiteDatabaseStatsCollector{db: db}
}

// Ensure SQLiteDatabaseStatsCollector implements DatabaseStatsCollectorInterface
var _ dbinterfaces.DatabaseStatsCollectorInterface = (*SQLiteDatabaseStatsCollector)(nil)

// GetDatabaseSize returns database size information
func (c *SQLiteDatabaseStatsCollector) GetDatabaseSize() (*models.DatabaseSize, error) {
	var result models.DatabaseSize

	// Get database file path
	var dbPath string
	err := c.db.QueryRow("PRAGMA database_list").Scan(nil, nil, &dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	if dbPath == ":memory:" {
		result.DatabaseName = "In-Memory Database"
		result.Size = "N/A (In-Memory)"
		result.SizeBytes = 0
		return &result, nil
	}

	// Get database name from file path
	result.DatabaseName = filepath.Base(dbPath)

	// Get file size
	if fileInfo, err := os.Stat(dbPath); err == nil {
		result.SizeBytes = fileInfo.Size()
		sizeMB := float64(result.SizeBytes) / (1024 * 1024)
		result.Size = fmt.Sprintf("%.2f MB", sizeMB)
	} else {
		result.Size = "Unknown"
		result.SizeBytes = 0
	}

	return &result, nil
}

// GetActiveConnections returns information about active connections
func (c *SQLiteDatabaseStatsCollector) GetActiveConnections() ([]models.ActiveConnection, error) {
	// SQLite is a file-based database and doesn't have the concept of multiple
	// active connections in the same way as client-server databases
	// We'll return information about the current connection

	var results []models.ActiveConnection

	// Create a single connection entry representing the current connection
	conn := models.ActiveConnection{
		PID:        1, // SQLite doesn't have PIDs, use 1 as placeholder
		Username:   "local",
		Database:   "main",
		ClientAddr: "local",
		State:      "active",
		Query:      "Connected to SQLite database",
		Duration:   "N/A",
	}

	results = append(results, conn)

	return results, nil
}
