package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type DatabaseTools struct {
	db *sql.DB
}

type QueryResult struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	Error   string                   `json:"error,omitempty"`
}

type TableInfo struct {
	TableName string `json:"table_name"`
}

type ColumnInfo struct {
	ColumnName             string  `json:"column_name"`
	DataType               string  `json:"data_type"`
	IsNullable             string  `json:"is_nullable"`
	ColumnDefault          *string `json:"column_default"`
	CharacterMaximumLength *int    `json:"character_maximum_length"`
}

type IndexInfo struct {
	IndexName string `json:"indexname"`
	IndexDef  string `json:"indexdef"`
}

type TableStats struct {
	SchemaName      string      `json:"schemaname"`
	TableName       string      `json:"tablename"`
	AttName         string      `json:"attname"`
	NDistinct       *float64    `json:"n_distinct"`
	MostCommonVals  interface{} `json:"most_common_vals"`
	MostCommonFreqs interface{} `json:"most_common_freqs"`
}

type SlowQuery struct {
	Query     string  `json:"query"`
	Calls     int64   `json:"calls"`
	TotalTime float64 `json:"total_time"`
	MeanTime  float64 `json:"mean_time"`
	Rows      int64   `json:"rows"`
}

type DatabaseSize struct {
	DatName string `json:"datname"`
	Size    string `json:"size"`
}

type TableSize struct {
	SchemaName string `json:"schemaname"`
	TableName  string `json:"tablename"`
	Size       string `json:"size"`
	TableSize  string `json:"table_size"`
	IndexSize  string `json:"index_size"`
}

type ActiveConnection struct {
	PID             int     `json:"pid"`
	Username        string  `json:"usename"`
	ApplicationName string  `json:"application_name"`
	ClientAddr      *string `json:"client_addr"`
	State           string  `json:"state"`
	QueryStart      *string `json:"query_start"`
	QueryPreview    string  `json:"query_preview"`
}

func NewDatabaseTools(databaseURL string) (*DatabaseTools, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseTools{db: db}, nil
}

func (dt *DatabaseTools) Close() error {
	if dt.db != nil {
		return dt.db.Close()
	}
	return nil
}

// CheckConnection verifies that the database connection is healthy
func (dt *DatabaseTools) CheckConnection() error {
	if dt == nil || dt.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Use ping to check if connection is alive
	if err := dt.db.Ping(); err != nil {
		return fmt.Errorf("database connection check failed: %w", err)
	}

	return nil
}

// IsConnectionHealthy returns true if the database connection is healthy
func (dt *DatabaseTools) IsConnectionHealthy() bool {
	return dt.CheckConnection() == nil
}

func (dt *DatabaseTools) ExecuteSQL(query string) (*QueryResult, error) {
	if dt == nil || dt.db == nil {
		return &QueryResult{
			Error: "No database connection available. Please add and switch to a database connection first.",
		}, fmt.Errorf("no database connection available")
	}

	rows, err := dt.db.Query(query)
	if err != nil {
		return &QueryResult{Error: err.Error()}, nil
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return &QueryResult{Error: err.Error()}, nil
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &QueryResult{Error: err.Error()}, nil
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return &QueryResult{
		Columns: columns,
		Rows:    results,
	}, nil
}

func (dt *DatabaseTools) GetAllTables() ([]TableInfo, error) {
	if dt == nil || dt.db == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
	rows, err := dt.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.TableName); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func (dt *DatabaseTools) GetTableSchema(tableName string) ([]ColumnInfo, error) {
	if dt == nil || dt.db == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	query := `
		SELECT 
			column_name, 
			data_type, 
			is_nullable, 
			column_default,
			character_maximum_length
		FROM information_schema.columns 
		WHERE table_name = $1
		ORDER BY ordinal_position
	`
	rows, err := dt.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.ColumnName, &col.DataType, &col.IsNullable,
			&col.ColumnDefault, &col.CharacterMaximumLength); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	return columns, nil
}

func (dt *DatabaseTools) ExplainQuery(query string) (interface{}, error) {
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)
	row := dt.db.QueryRow(explainQuery)

	var result string
	if err := row.Scan(&result); err != nil {
		return nil, err
	}

	var jsonResult interface{}
	if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
		return nil, err
	}

	return jsonResult, nil
}

func (dt *DatabaseTools) GetTableIndexes(tableName string) ([]IndexInfo, error) {
	query := `
		SELECT 
			indexname,
			indexdef
		FROM pg_indexes 
		WHERE tablename = $1
	`
	rows, err := dt.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		if err := rows.Scan(&idx.IndexName, &idx.IndexDef); err != nil {
			return nil, err
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

func (dt *DatabaseTools) GetTableStats(tableName string) ([]TableStats, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			attname,
			n_distinct,
			most_common_vals,
			most_common_freqs
		FROM pg_stats 
		WHERE tablename = $1
	`
	rows, err := dt.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []TableStats
	for rows.Next() {
		var stat TableStats
		if err := rows.Scan(&stat.SchemaName, &stat.TableName, &stat.AttName,
			&stat.NDistinct, &stat.MostCommonVals, &stat.MostCommonFreqs); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (dt *DatabaseTools) FindDuplicateData(tableName string, columns []string) (*QueryResult, error) {
	columnList := strings.Join(columns, ", ")
	query := fmt.Sprintf(`
		SELECT %s, COUNT(*) as duplicate_count
		FROM %s
		GROUP BY %s
		HAVING COUNT(*) > 1
		ORDER BY duplicate_count DESC
		LIMIT 100
	`, columnList, tableName, columnList)

	return dt.ExecuteSQL(query)
}

func (dt *DatabaseTools) GetSlowQueries() ([]SlowQuery, error) {
	query := `
		SELECT 
			query,
			calls,
			total_time,
			mean_time,
			rows
		FROM pg_stat_statements 
		ORDER BY total_time DESC 
		LIMIT 10
	`
	rows, err := dt.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []SlowQuery
	for rows.Next() {
		var q SlowQuery
		if err := rows.Scan(&q.Query, &q.Calls, &q.TotalTime, &q.MeanTime, &q.Rows); err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}

	return queries, nil
}

func (dt *DatabaseTools) GetDatabaseSize() (*DatabaseSize, error) {
	query := `
		SELECT 
			pg_database.datname,
			pg_size_pretty(pg_database_size(pg_database.datname)) AS size
		FROM pg_database
		WHERE datname = current_database()
	`
	row := dt.db.QueryRow(query)

	var size DatabaseSize
	if err := row.Scan(&size.DatName, &size.Size); err != nil {
		return nil, err
	}

	return &size, nil
}

func (dt *DatabaseTools) GetTableSizes() ([]TableSize, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
			pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS index_size
		FROM pg_tables 
		WHERE schemaname = 'public'
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
	`
	rows, err := dt.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sizes []TableSize
	for rows.Next() {
		var size TableSize
		if err := rows.Scan(&size.SchemaName, &size.TableName, &size.Size,
			&size.TableSize, &size.IndexSize); err != nil {
			return nil, err
		}
		sizes = append(sizes, size)
	}

	return sizes, nil
}

func (dt *DatabaseTools) GetActiveConnections() ([]ActiveConnection, error) {
	query := `
		SELECT 
			pid,
			usename,
			application_name,
			client_addr,
			state,
			query_start,
			LEFT(query, 100) as query_preview
		FROM pg_stat_activity 
		WHERE state != 'idle'
		ORDER BY query_start DESC
	`
	rows, err := dt.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []ActiveConnection
	for rows.Next() {
		var conn ActiveConnection
		if err := rows.Scan(&conn.PID, &conn.Username, &conn.ApplicationName,
			&conn.ClientAddr, &conn.State, &conn.QueryStart, &conn.QueryPreview); err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}

	return connections, nil
}
