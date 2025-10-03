package postgresql

import (
	"database/sql"
	"fmt"
	"strings"

	"dbsage/internal/models"
	"dbsage/pkg/database/postgresql/queries"
	"dbsage/pkg/dbinterfaces"

	_ "github.com/lib/pq"
)

// PostgreSQLDatabase implements the DatabaseInterface for PostgreSQL
type PostgreSQLDatabase struct {
	db            *sql.DB
	queryExecutor dbinterfaces.QueryExecutorInterface
}

// Ensure PostgreSQLDatabase implements DatabaseInterface
var _ dbinterfaces.DatabaseInterface = (*PostgreSQLDatabase)(nil)

// NewPostgreSQLDatabase creates a new PostgreSQL database instance
func NewPostgreSQLDatabase(connectionURL string) (*PostgreSQLDatabase, error) {
	db, err := sql.Open("postgres", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQLDatabase{
		db:            db,
		queryExecutor: queries.NewPostgreSQLExecutor(db),
	}, nil
}

// Close closes the database connection
func (pg *PostgreSQLDatabase) Close() error {
	if pg.db != nil {
		return pg.db.Close()
	}
	return nil
}

// IsConnectionHealthy checks if the database connection is healthy
func (pg *PostgreSQLDatabase) IsConnectionHealthy() bool {
	if pg.db == nil {
		return false
	}
	return pg.db.Ping() == nil
}

// CheckConnection checks if the database connection is working
func (pg *PostgreSQLDatabase) CheckConnection() error {
	if pg.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return pg.db.Ping()
}

// ExecuteSQL executes a SQL query
func (pg *PostgreSQLDatabase) ExecuteSQL(query string) (*models.QueryResult, error) {
	return pg.queryExecutor.ExecuteSQL(query)
}

// ExplainQuery analyzes a query's execution plan
func (pg *PostgreSQLDatabase) ExplainQuery(query string) (*models.QueryResult, error) {
	return pg.queryExecutor.ExplainQuery(query)
}

// GetAllTables returns a list of all tables
func (pg *PostgreSQLDatabase) GetAllTables() ([]models.TableInfo, error) {
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

	rows, err := pg.db.Query(query)
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
func (pg *PostgreSQLDatabase) GetTableSchema(tableName string) ([]models.ColumnInfo, error) {
	query := `
		SELECT 
			isc."column_name",
			isc.data_type,
			isc.is_nullable,
			isc.column_default,
			isc.character_maximum_length,
			isc.numeric_precision,
			isc.numeric_scale,
			CASE WHEN pk."column_name" IS NOT NULL THEN true ELSE false END as is_primary_key,
			CASE WHEN fk."column_name" IS NOT NULL THEN true ELSE false END as is_foreign_key,
			COALESCE(col_description(c.oid, a.attnum), '') as description
		FROM information_schema.columns isc
		LEFT JOIN pg_class c ON c.relname = isc.table_name
		LEFT JOIN pg_attribute a ON a.attrelid = c.oid AND a.attname = isc."column_name"
		LEFT JOIN (
			SELECT ku."column_name"
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_name = $1
		) pk ON pk."column_name" = isc."column_name"
		LEFT JOIN (
			SELECT ku."column_name"
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name = $1
		) fk ON fk."column_name" = isc."column_name"
		WHERE isc.table_name = $1
		ORDER BY isc.ordinal_position
	`

	rows, err := pg.db.Query(query, tableName)
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
func (pg *PostgreSQLDatabase) GetTableIndexes(tableName string) ([]models.IndexInfo, error) {
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

	rows, err := pg.db.Query(query, tableName)
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
		columnsArray = strings.Trim(columnsArray, "{}")
		if columnsArray != "" {
			idx.Columns = strings.Split(columnsArray, ",")
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// FindDuplicateData finds duplicate records in a table
func (pg *PostgreSQLDatabase) FindDuplicateData(tableName string, columns []string) (*models.QueryResult, error) {
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

	return pg.queryExecutor.ExecuteSQL(query)
}
