package database

import (
	"database/sql"
	"fmt"
	"strings"

	"dbsage/internal/models"
	"dbsage/pkg/database/queries"
	"dbsage/pkg/database/stats"

	_ "github.com/lib/pq"
)

// DatabaseTools provides high-level database operations
type DatabaseTools struct {
	db                     *sql.DB
	queryExecutor          *queries.Executor
	tableStatsCollector    *stats.TableStatsCollector
	databaseStatsCollector *stats.DatabaseStatsCollector
}

// NewDatabaseTools creates a new database tools instance
func NewDatabaseTools(connectionURL string) (*DatabaseTools, error) {
	db, err := sql.Open("postgres", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseTools{
		db:                     db,
		queryExecutor:          queries.NewExecutor(db),
		tableStatsCollector:    stats.NewTableStatsCollector(db),
		databaseStatsCollector: stats.NewDatabaseStatsCollector(db),
	}, nil
}

// Close closes the database connection
func (dt *DatabaseTools) Close() error {
	if dt.db != nil {
		return dt.db.Close()
	}
	return nil
}

// IsConnectionHealthy checks if the database connection is healthy
func (dt *DatabaseTools) IsConnectionHealthy() bool {
	if dt.db == nil {
		return false
	}
	return dt.db.Ping() == nil
}

// CheckConnection checks if the database connection is working
func (dt *DatabaseTools) CheckConnection() error {
	if dt.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return dt.db.Ping()
}

// ExecuteSQL executes a SQL query
func (dt *DatabaseTools) ExecuteSQL(query string) (*models.QueryResult, error) {
	return dt.queryExecutor.ExecuteSQL(query)
}

// ExplainQuery analyzes a query's execution plan
func (dt *DatabaseTools) ExplainQuery(query string) (*models.QueryResult, error) {
	return dt.queryExecutor.ExplainQuery(query)
}

// GetAllTables returns a list of all tables
func (dt *DatabaseTools) GetAllTables() ([]models.TableInfo, error) {
	query := `
		SELECT 
			table_name,
			table_schema,
			table_type,
			COALESCE(obj_description(c.oid), '') as table_comment
		FROM information_schema.tables t
		LEFT JOIN pg_class c ON c.relname = t.table_name
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY table_schema, table_name
	`

	rows, err := dt.db.Query(query)
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
func (dt *DatabaseTools) GetTableSchema(tableName string) ([]models.ColumnInfo, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
			CASE WHEN fk.column_name IS NOT NULL THEN true ELSE false END as is_foreign_key,
			COALESCE(col_description(c.oid, a.attnum), '') as description
		FROM information_schema.columns isc
		LEFT JOIN pg_class c ON c.relname = isc.table_name
		LEFT JOIN pg_attribute a ON a.attrelid = c.oid AND a.attname = isc.column_name
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_name = $1
		) pk ON pk.column_name = isc.column_name
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name = $1
		) fk ON fk.column_name = isc.column_name
		WHERE isc.table_name = $1
		ORDER BY isc.ordinal_position
	`

	rows, err := dt.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var col models.ColumnInfo
		err := rows.Scan(
			&col.ColumnName,
			&col.DataType,
			&col.IsNullable,
			&col.DefaultValue,
			&col.CharMaxLength,
			&col.NumPrecision,
			&col.NumScale,
			&col.IsPrimaryKey,
			&col.IsForeignKey,
			&col.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}
		columns = append(columns, col)
	}

	return columns, nil
}

// GetTableIndexes returns index information for a table
func (dt *DatabaseTools) GetTableIndexes(tableName string) ([]models.IndexInfo, error) {
	query := `
		SELECT 
			i.relname as index_name,
			idx.indisunique as is_unique,
			idx.indisprimary as is_primary,
			array_agg(a.attname ORDER BY a.attnum) as columns,
			am.amname as index_type,
			COALESCE(ts.spcname, 'default') as tablespace,
			COALESCE(obj_description(i.oid), '') as description
		FROM pg_index idx
		JOIN pg_class i ON i.oid = idx.indexrelid
		JOIN pg_class t ON t.oid = idx.indrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(idx.indkey)
		JOIN pg_am am ON am.oid = i.relam
		LEFT JOIN pg_tablespace ts ON ts.oid = i.reltablespace
		WHERE t.relname = $1
		GROUP BY i.relname, idx.indisunique, idx.indisprimary, am.amname, ts.spcname, i.oid
		ORDER BY i.relname
	`

	rows, err := dt.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query table indexes: %w", err)
	}
	defer rows.Close()

	var indexes []models.IndexInfo
	for rows.Next() {
		var idx models.IndexInfo
		var columnsArray string
		err := rows.Scan(
			&idx.IndexName,
			&idx.IsUnique,
			&idx.IsPrimary,
			&columnsArray,
			&idx.IndexType,
			&idx.TableSpace,
			&idx.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index row: %w", err)
		}

		// Parse the columns array (PostgreSQL array format)
		// This is a simplified parser - you might want to use a proper array parser
		columnsArray = strings.Trim(columnsArray, "{}")
		if columnsArray != "" {
			idx.Columns = strings.Split(columnsArray, ",")
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// GetTableStats returns statistics for a table
func (dt *DatabaseTools) GetTableStats(tableName string) (*models.TableStats, error) {
	return dt.tableStatsCollector.GetTableStats(tableName)
}

// GetTableSizes returns size information for all tables
func (dt *DatabaseTools) GetTableSizes() ([]map[string]interface{}, error) {
	return dt.tableStatsCollector.GetTableSizes()
}

// GetSlowQueries returns slow query information
func (dt *DatabaseTools) GetSlowQueries() ([]models.SlowQuery, error) {
	return dt.tableStatsCollector.GetSlowQueries()
}

// GetDatabaseSize returns database size information
func (dt *DatabaseTools) GetDatabaseSize() (*models.DatabaseSize, error) {
	return dt.databaseStatsCollector.GetDatabaseSize()
}

// GetActiveConnections returns active connection information
func (dt *DatabaseTools) GetActiveConnections() ([]models.ActiveConnection, error) {
	return dt.databaseStatsCollector.GetActiveConnections()
}

// FindDuplicateData finds duplicate records in a table
func (dt *DatabaseTools) FindDuplicateData(tableName string, columns []string) (*models.QueryResult, error) {
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

	return dt.queryExecutor.ExecuteSQL(query)
}
