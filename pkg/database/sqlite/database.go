package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"dbsage/internal/models"
	"dbsage/pkg/database/sqlite/optimization"
	"dbsage/pkg/database/sqlite/queries"
	"dbsage/pkg/database/sqlite/stats"
	"dbsage/pkg/dbinterfaces"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDatabase implements the DatabaseInterface for SQLite
type SQLiteDatabase struct {
	db                     *sql.DB
	queryExecutor          dbinterfaces.QueryExecutorInterface
	tableStatsCollector    dbinterfaces.TableStatsCollectorInterface
	databaseStatsCollector dbinterfaces.DatabaseStatsCollectorInterface
	queryOptimizer         *optimization.SQLiteQueryOptimizer
}

// Ensure SQLiteDatabase implements DatabaseInterface
var _ dbinterfaces.DatabaseInterface = (*SQLiteDatabase)(nil)

// NewSQLiteDatabase creates a new SQLite database instance
func NewSQLiteDatabase(connectionURL string) (*SQLiteDatabase, error) {
	db, err := sql.Open("sqlite3", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SQLiteDatabase{
		db:                     db,
		queryExecutor:          queries.NewSQLiteExecutor(db),
		tableStatsCollector:    stats.NewSQLiteTableStatsCollector(db),
		databaseStatsCollector: stats.NewSQLiteDatabaseStatsCollector(db),
		queryOptimizer:         optimization.NewSQLiteQueryOptimizer(db),
	}, nil
}

// Close closes the database connection
func (s *SQLiteDatabase) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// IsConnectionHealthy checks if the database connection is healthy
func (s *SQLiteDatabase) IsConnectionHealthy() bool {
	if s.db == nil {
		return false
	}
	return s.db.Ping() == nil
}

// CheckConnection checks if the database connection is working
func (s *SQLiteDatabase) CheckConnection() error {
	if s.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return s.db.Ping()
}

// ExecuteSQL executes a SQL query
func (s *SQLiteDatabase) ExecuteSQL(query string) (*models.QueryResult, error) {
	return s.queryExecutor.ExecuteSQL(query)
}

// ExplainQuery analyzes a query's execution plan
func (s *SQLiteDatabase) ExplainQuery(query string) (*models.QueryResult, error) {
	return s.queryExecutor.ExplainQuery(query)
}

// GetAllTables returns a list of all tables
func (s *SQLiteDatabase) GetAllTables() ([]models.TableInfo, error) {
	query := `
		SELECT 
			name as table_name,
			'main' as table_schema,
			type as table_type,
			'' as table_comment
		FROM sqlite_master 
		WHERE type IN ('table', 'view')
		AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []models.TableInfo
	for rows.Next() {
		var table models.TableInfo
		err := rows.Scan(&table.TableName, &table.Schema, &table.TableType, &table.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// GetTableSchema returns detailed schema information for a table
func (s *SQLiteDatabase) GetTableSchema(tableName string) ([]models.ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var cid int
		var col models.ColumnInfo
		var notNull int
		var defaultValue sql.NullString

		err := rows.Scan(
			&cid,
			&col.ColumnName,
			&col.DataType,
			&notNull,
			&defaultValue,
			&col.IsPrimaryKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}

		// Convert SQLite boolean representation
		if notNull == 0 {
			col.IsNullable = "YES"
		} else {
			col.IsNullable = "NO"
		}
		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}

		// SQLite doesn't have explicit foreign key info in PRAGMA table_info
		// We'll need to check PRAGMA foreign_key_list separately
		col.IsForeignKey = false

		columns = append(columns, col)
	}

	// Check for foreign keys
	fkQuery := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	fkRows, err := s.db.Query(fkQuery)
	if err == nil {
		defer fkRows.Close()
		fkColumns := make(map[string]bool)
		for fkRows.Next() {
			var id, seq int
			var table, from, to, onUpdate, onDelete, match string
			err := fkRows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match)
			if err == nil {
				fkColumns[from] = true
			}
		}

		// Update foreign key status
		for i := range columns {
			if fkColumns[columns[i].ColumnName] {
				columns[i].IsForeignKey = true
			}
		}
	}

	return columns, nil
}

// GetTableIndexes returns index information for a table
func (s *SQLiteDatabase) GetTableIndexes(tableName string) ([]models.IndexInfo, error) {
	query := fmt.Sprintf("PRAGMA index_list(%s)", tableName)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table indexes: %w", err)
	}
	defer rows.Close()

	var indexes []models.IndexInfo
	for rows.Next() {
		var seq int
		var idx models.IndexInfo
		var unique, partial int
		var origin string

		err := rows.Scan(
			&seq,
			&idx.IndexName,
			&unique,
			&origin,
			&partial,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index row: %w", err)
		}

		idx.IsUnique = unique == 1
		idx.IsPrimary = origin == "pk"
		idx.IndexType = "BTREE" // SQLite primarily uses B-tree indexes

		// Get index columns
		colQuery := fmt.Sprintf("PRAGMA index_info(%s)", idx.IndexName)
		colRows, err := s.db.Query(colQuery)
		if err == nil {
			var columns []string
			for colRows.Next() {
				var seqno, cid int
				var name string
				err := colRows.Scan(&seqno, &cid, &name)
				if err == nil {
					columns = append(columns, name)
				}
			}
			colRows.Close()
			idx.Columns = columns
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// FindDuplicateData finds duplicate records in a table
func (s *SQLiteDatabase) FindDuplicateData(tableName string, columns []string) (*models.QueryResult, error) {
	if len(columns) == 0 {
		return nil, fmt.Errorf("at least one column must be specified")
	}

	columnList := strings.Join(columns, ", ")
	query := fmt.Sprintf(`
		SELECT %s, COUNT(*) as duplicate_count
		FROM %s
		GROUP BY %s
		HAVING COUNT(*) > 1
		ORDER BY COUNT(*) DESC
		LIMIT 100
	`, columnList, tableName, columnList)

	return s.queryExecutor.ExecuteSQL(query)
}

// GetTableStats returns statistics for a table
func (s *SQLiteDatabase) GetTableStats(tableName string) (*models.TableStats, error) {
	return s.tableStatsCollector.GetTableStats(tableName)
}

// GetTableSizes returns size information for all tables
func (s *SQLiteDatabase) GetTableSizes() ([]map[string]interface{}, error) {
	return s.tableStatsCollector.GetTableSizes()
}

// GetSlowQueries returns slow query information (limited in SQLite)
func (s *SQLiteDatabase) GetSlowQueries() ([]models.SlowQuery, error) {
	return s.tableStatsCollector.GetSlowQueries()
}

// GetDatabaseSize returns database size information
func (s *SQLiteDatabase) GetDatabaseSize() (*models.DatabaseSize, error) {
	return s.databaseStatsCollector.GetDatabaseSize()
}

// GetActiveConnections returns active connection information
func (s *SQLiteDatabase) GetActiveConnections() ([]models.ActiveConnection, error) {
	return s.databaseStatsCollector.GetActiveConnections()
}

// Query Optimization Methods

// AnalyzeQueryPerformance analyzes the performance of a given query
func (s *SQLiteDatabase) AnalyzeQueryPerformance(query string) (*models.PerformanceAnalysis, error) {
	return s.queryOptimizer.AnalyzeQueryPerformance(query)
}

// SuggestIndexes suggests indexes for a specific table
func (s *SQLiteDatabase) SuggestIndexes(tableName string) ([]models.IndexSuggestion, error) {
	return s.queryOptimizer.SuggestIndexes(tableName)
}

// GetQueryPatterns analyzes query patterns (limited in SQLite)
func (s *SQLiteDatabase) GetQueryPatterns() ([]models.QueryPattern, error) {
	return s.queryOptimizer.GetQueryPatterns()
}

// OptimizeQuery provides optimization suggestions for a specific query
func (s *SQLiteDatabase) OptimizeQuery(query string) ([]models.QueryOptimizationSuggestion, error) {
	return s.queryOptimizer.OptimizeQuery(query)
}

// AnalyzeTablePerformance analyzes performance issues specific to a table
func (s *SQLiteDatabase) AnalyzeTablePerformance(tableName string) (*models.PerformanceAnalysis, error) {
	return s.queryOptimizer.AnalyzeTablePerformance(tableName)
}
