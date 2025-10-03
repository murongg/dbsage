package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionConfigJSON(t *testing.T) {
	config := ConnectionConfig{
		Name:     "test_db",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
		SSLMode:  "require",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(config)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledConfig ConnectionConfig
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	require.NoError(t, err)

	assert.Equal(t, config, unmarshaledConfig)
}

func TestQueryResultJSON(t *testing.T) {
	result := QueryResult{
		Columns:  []string{"id", "name", "email"},
		Rows:     [][]interface{}{{1, "John", "john@example.com"}, {2, "Jane", "jane@example.com"}},
		RowCount: 2,
		Duration: "15ms",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(result)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledResult QueryResult
	err = json.Unmarshal(jsonData, &unmarshaledResult)
	require.NoError(t, err)

	assert.Equal(t, result.Columns, unmarshaledResult.Columns)
	assert.Equal(t, result.RowCount, unmarshaledResult.RowCount)
	assert.Equal(t, result.Duration, unmarshaledResult.Duration)
	assert.Len(t, unmarshaledResult.Rows, 2)
}

func TestTableInfoValidation(t *testing.T) {
	tableInfo := TableInfo{
		TableName:   "users",
		Schema:      "public",
		TableType:   "table",
		Description: "User information table",
	}

	assert.NotEmpty(t, tableInfo.TableName)
	assert.NotEmpty(t, tableInfo.Schema)
	assert.NotEmpty(t, tableInfo.TableType)
}

func TestColumnInfoValidation(t *testing.T) {
	maxLength := 255
	precision := 10
	scale := 2
	defaultValue := "active"

	columnInfo := ColumnInfo{
		ColumnName:    "status",
		DataType:      "varchar",
		IsNullable:    "NO",
		DefaultValue:  &defaultValue,
		CharMaxLength: &maxLength,
		NumPrecision:  &precision,
		NumScale:      &scale,
		IsPrimaryKey:  false,
		IsForeignKey:  false,
		Description:   "User status column",
	}

	assert.NotEmpty(t, columnInfo.ColumnName)
	assert.NotEmpty(t, columnInfo.DataType)
	assert.NotNil(t, columnInfo.DefaultValue)
	assert.Equal(t, "active", *columnInfo.DefaultValue)
	assert.NotNil(t, columnInfo.CharMaxLength)
	assert.Equal(t, 255, *columnInfo.CharMaxLength)
}

func TestIndexInfoValidation(t *testing.T) {
	indexInfo := IndexInfo{
		IndexName:   "idx_user_email",
		IsUnique:    true,
		IsPrimary:   false,
		Columns:     []string{"email"},
		IndexType:   "btree",
		TableSpace:  "pg_default",
		Description: "Index on user email",
	}

	assert.NotEmpty(t, indexInfo.IndexName)
	assert.True(t, indexInfo.IsUnique)
	assert.False(t, indexInfo.IsPrimary)
	assert.Len(t, indexInfo.Columns, 1)
	assert.Equal(t, "email", indexInfo.Columns[0])
}

func TestTableStatsValidation(t *testing.T) {
	stats := TableStats{
		TableName:   "users",
		RowCount:    1000,
		TableSize:   "64 MB",
		IndexSize:   "16 MB",
		TotalSize:   "80 MB",
		LastVacuum:  "2023-01-01 10:00:00",
		LastAnalyze: "2023-01-01 10:00:00",
		SeqScan:     50,
		SeqTupRead:  5000,
		IdxScan:     950,
		IdxTupFetch: 9500,
	}

	assert.NotEmpty(t, stats.TableName)
	assert.Positive(t, stats.RowCount)
	assert.NotEmpty(t, stats.TableSize)
	assert.Greater(t, stats.IdxScan, stats.SeqScan, "Index scans should be more frequent than sequential scans for this table")
}

func TestSlowQueryValidation(t *testing.T) {
	slowQuery := SlowQuery{
		Query:      "SELECT * FROM users WHERE email LIKE '%@example.com%'",
		Calls:      100,
		TotalTime:  5000.0,
		MeanTime:   50.0,
		MinTime:    10.0,
		MaxTime:    200.0,
		StddevTime: 25.0,
		Rows:       1000,
	}

	assert.NotEmpty(t, slowQuery.Query)
	assert.Positive(t, slowQuery.Calls)
	assert.Positive(t, slowQuery.TotalTime)
	assert.Positive(t, slowQuery.MeanTime)
	assert.LessOrEqual(t, slowQuery.MinTime, slowQuery.MeanTime)
	assert.GreaterOrEqual(t, slowQuery.MaxTime, slowQuery.MeanTime)
}

func TestDatabaseSizeValidation(t *testing.T) {
	dbSize := DatabaseSize{
		DatabaseName: "testdb",
		Size:         "1.5 GB",
		SizeBytes:    1610612736, // 1.5 GB in bytes
	}

	assert.NotEmpty(t, dbSize.DatabaseName)
	assert.NotEmpty(t, dbSize.Size)
	assert.Positive(t, dbSize.SizeBytes)
}

func TestActiveConnectionValidation(t *testing.T) {
	connection := ActiveConnection{
		PID:        12345,
		Username:   "dbuser",
		Database:   "testdb",
		ClientAddr: "192.168.1.100",
		State:      "active",
		Query:      "SELECT * FROM users",
		Duration:   "00:00:05",
	}

	assert.Positive(t, connection.PID)
	assert.NotEmpty(t, connection.Username)
	assert.NotEmpty(t, connection.Database)
	assert.NotEmpty(t, connection.State)
}

func TestQueryResultRowValidation(t *testing.T) {
	result := QueryResult{
		Columns:  []string{"id", "name", "created_at"},
		Rows:     [][]interface{}{{1, "John", "2023-01-01"}, {2, nil, "2023-01-02"}},
		RowCount: 2,
		Duration: "10ms",
	}

	// Test that row length matches column count
	for i, row := range result.Rows {
		assert.Len(t, row, len(result.Columns),
			"Row %d should have same number of values as columns", i)
	}

	// Test that RowCount matches actual row count
	assert.Equal(t, len(result.Rows), result.RowCount)
}

// Benchmark tests
func BenchmarkQueryResultJSON(b *testing.B) {
	result := QueryResult{
		Columns:  []string{"id", "name", "email", "created_at"},
		Rows:     make([][]interface{}, 1000),
		RowCount: 1000,
		Duration: "100ms",
	}

	// Fill with sample data
	for i := 0; i < 1000; i++ {
		result.Rows[i] = []interface{}{i, "User" + string(rune(i)), "user" + string(rune(i)) + "@example.com", "2023-01-01"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(result)
	}
}
