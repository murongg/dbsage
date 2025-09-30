package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"dbsage/internal/models"
	"dbsage/pkg/database/mysql/queries"
	"dbsage/pkg/database/mysql/stats"
	"dbsage/pkg/dbinterfaces"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDatabase implements the DatabaseInterface for MySQL
type MySQLDatabase struct {
	db                     *sql.DB
	queryExecutor          dbinterfaces.QueryExecutorInterface
	tableStatsCollector    dbinterfaces.TableStatsCollectorInterface
	databaseStatsCollector dbinterfaces.DatabaseStatsCollectorInterface
}

// Ensure MySQLDatabase implements DatabaseInterface
var _ dbinterfaces.DatabaseInterface = (*MySQLDatabase)(nil)

// NewMySQLDatabase creates a new MySQL database instance
func NewMySQLDatabase(connectionURL string) (*MySQLDatabase, error) {
	db, err := sql.Open("mysql", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MySQLDatabase{
		db:                     db,
		queryExecutor:          queries.NewMySQLExecutor(db),
		tableStatsCollector:    stats.NewMySQLTableStatsCollector(db),
		databaseStatsCollector: stats.NewMySQLDatabaseStatsCollector(db),
	}, nil
}

// Close closes the database connection
func (m *MySQLDatabase) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// IsConnectionHealthy checks if the database connection is healthy
func (m *MySQLDatabase) IsConnectionHealthy() bool {
	if m.db == nil {
		return false
	}
	return m.db.Ping() == nil
}

// CheckConnection checks if the database connection is working
func (m *MySQLDatabase) CheckConnection() error {
	if m.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return m.db.Ping()
}

// ExecuteSQL executes a SQL query
func (m *MySQLDatabase) ExecuteSQL(query string) (*models.QueryResult, error) {
	return m.queryExecutor.ExecuteSQL(query)
}

// ExplainQuery analyzes a query's execution plan
func (m *MySQLDatabase) ExplainQuery(query string) (*models.QueryResult, error) {
	return m.queryExecutor.ExplainQuery(query)
}

// GetAllTables returns a list of all tables
func (m *MySQLDatabase) GetAllTables() ([]models.TableInfo, error) {
	query := `
		SELECT 
			table_name,
			table_schema,
			table_type,
			COALESCE(table_comment, '') as table_comment
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
		ORDER BY table_schema, table_name
	`

	rows, err := m.db.Query(query)
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
func (m *MySQLDatabase) GetTableSchema(tableName string) ([]models.ColumnInfo, error) {
	query := "SELECT " +
		"`column_name`, " +
		"data_type, " +
		"is_nullable, " +
		"column_default, " +
		"character_maximum_length, " +
		"numeric_precision, " +
		"numeric_scale, " +
		"CASE WHEN column_key = 'PRI' THEN true ELSE false END as is_primary_key, " +
		"CASE WHEN column_key = 'MUL' THEN true ELSE false END as is_foreign_key, " +
		"COALESCE(column_comment, '') as description " +
		"FROM information_schema.columns " +
		"WHERE table_schema = DATABASE() AND table_name = ? " +
		"ORDER BY ordinal_position"

	rows, err := m.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var col models.ColumnInfo
		var defaultValue sql.NullString
		var description sql.NullString
		err := rows.Scan(
			&col.ColumnName,
			&col.DataType,
			&col.IsNullable,
			&defaultValue,
			&col.CharMaxLength,
			&col.NumPrecision,
			&col.NumScale,
			&col.IsPrimaryKey,
			&col.IsForeignKey,
			&description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}

		// Handle nullable fields
		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}
		if description.Valid {
			col.Description = description.String
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// GetTableIndexes returns index information for a table
func (m *MySQLDatabase) GetTableIndexes(tableName string) ([]models.IndexInfo, error) {
	query := "SELECT " +
		"index_name, " +
		"CASE WHEN non_unique = 0 THEN true ELSE false END as is_unique, " +
		"CASE WHEN index_name = 'PRIMARY' THEN true ELSE false END as is_primary, " +
		"GROUP_CONCAT(`column_name` ORDER BY seq_in_index) as columns, " +
		"index_type, " +
		"'' as tablespace, " +
		"COALESCE(index_comment, '') as description " +
		"FROM information_schema.statistics " +
		"WHERE table_schema = DATABASE() AND table_name = ? " +
		"GROUP BY index_name, non_unique, index_type, index_comment " +
		"ORDER BY index_name"

	rows, err := m.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query table indexes: %w", err)
	}
	defer rows.Close()

	var indexes []models.IndexInfo
	for rows.Next() {
		var idx models.IndexInfo
		var columnsStr string
		err := rows.Scan(
			&idx.IndexName,
			&idx.IsUnique,
			&idx.IsPrimary,
			&columnsStr,
			&idx.IndexType,
			&idx.TableSpace,
			&idx.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index row: %w", err)
		}

		// Parse the columns
		if columnsStr != "" {
			idx.Columns = strings.Split(columnsStr, ",")
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// FindDuplicateData finds duplicate records in a table
func (m *MySQLDatabase) FindDuplicateData(tableName string, columns []string) (*models.QueryResult, error) {
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

	return m.queryExecutor.ExecuteSQL(query)
}

// GetTableStats returns statistics for a table
func (m *MySQLDatabase) GetTableStats(tableName string) (*models.TableStats, error) {
	return m.tableStatsCollector.GetTableStats(tableName)
}

// GetTableSizes returns size information for all tables
func (m *MySQLDatabase) GetTableSizes() ([]map[string]interface{}, error) {
	return m.tableStatsCollector.GetTableSizes()
}

// GetSlowQueries returns slow query information
func (m *MySQLDatabase) GetSlowQueries() ([]models.SlowQuery, error) {
	return m.tableStatsCollector.GetSlowQueries()
}

// GetDatabaseSize returns database size information
func (m *MySQLDatabase) GetDatabaseSize() (*models.DatabaseSize, error) {
	return m.databaseStatsCollector.GetDatabaseSize()
}

// GetActiveConnections returns active connection information
func (m *MySQLDatabase) GetActiveConnections() ([]models.ActiveConnection, error) {
	return m.databaseStatsCollector.GetActiveConnections()
}
