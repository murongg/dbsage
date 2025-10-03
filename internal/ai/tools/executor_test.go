package tools

import (
	"encoding/json"
	"errors"
	"testing"

	"dbsage/internal/models"
	"dbsage/pkg/dbinterfaces"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDatabaseInterface for testing
type MockDatabaseInterface struct {
	mock.Mock
}

func (m *MockDatabaseInterface) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabaseInterface) IsConnectionHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDatabaseInterface) CheckConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabaseInterface) ExecuteSQL(query string) (*models.QueryResult, error) {
	args := m.Called(query)
	return args.Get(0).(*models.QueryResult), args.Error(1)
}

func (m *MockDatabaseInterface) ExplainQuery(query string) (*models.QueryResult, error) {
	args := m.Called(query)
	return args.Get(0).(*models.QueryResult), args.Error(1)
}

func (m *MockDatabaseInterface) GetAllTables() ([]models.TableInfo, error) {
	args := m.Called()
	return args.Get(0).([]models.TableInfo), args.Error(1)
}

func (m *MockDatabaseInterface) GetTableSchema(tableName string) ([]models.ColumnInfo, error) {
	args := m.Called(tableName)
	return args.Get(0).([]models.ColumnInfo), args.Error(1)
}

func (m *MockDatabaseInterface) GetTableIndexes(tableName string) ([]models.IndexInfo, error) {
	args := m.Called(tableName)
	return args.Get(0).([]models.IndexInfo), args.Error(1)
}

func (m *MockDatabaseInterface) FindDuplicateData(tableName string, columns []string) (*models.QueryResult, error) {
	args := m.Called(tableName, columns)
	return args.Get(0).(*models.QueryResult), args.Error(1)
}

func TestNewExecutor(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	assert.NotNil(t, executor)
	assert.Equal(t, mockDB, executor.dbTools)
	assert.Nil(t, executor.getDbTools)
}

func TestNewExecutorWithDynamicTools(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	getDbTools := func() dbinterfaces.DatabaseInterface {
		return mockDB
	}

	executor := NewExecutorWithDynamicTools(getDbTools)

	assert.NotNil(t, executor)
	assert.Nil(t, executor.dbTools)
	assert.NotNil(t, executor.getDbTools)
	assert.Equal(t, mockDB, executor.getDbTools())
}

func TestExecutor_Execute_NoDatabaseConnection(t *testing.T) {
	// Test with nil database tools
	executor := NewExecutor(nil)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "execute_sql",
			Arguments: `{"sql": "SELECT 1"}`,
		},
	}

	result, err := executor.Execute(toolCall)
	require.NoError(t, err)
	assert.Contains(t, result, "No database connection available")
}

func TestExecutor_Execute_DynamicToolsNil(t *testing.T) {
	// Test with dynamic tools function returning nil
	getDbTools := func() dbinterfaces.DatabaseInterface {
		return nil
	}
	executor := NewExecutorWithDynamicTools(getDbTools)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "execute_sql",
			Arguments: `{"sql": "SELECT 1"}`,
		},
	}

	result, err := executor.Execute(toolCall)
	require.NoError(t, err)
	assert.Contains(t, result, "No database connection available")
}

func TestExecutor_Execute_InvalidJSON(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "execute_sql",
			Arguments: `{invalid json}`,
		},
	}

	result, err := executor.Execute(toolCall)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse tool arguments")
	assert.Empty(t, result)
}

func TestExecutor_Execute_UnknownTool(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "unknown_tool",
			Arguments: `{}`,
		},
	}

	result, err := executor.Execute(toolCall)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool: unknown_tool")
	assert.Empty(t, result)
}

func TestExecutor_ExecuteSQL(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	expectedResult := &models.QueryResult{
		Columns:  []string{"id", "name"},
		Rows:     [][]interface{}{{1, "John"}, {2, "Jane"}},
		RowCount: 2,
		Duration: "10ms",
	}

	mockDB.On("ExecuteSQL", "SELECT * FROM users").Return(expectedResult, nil)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "execute_sql",
			Arguments: `{"sql": "SELECT * FROM users"}`,
		},
	}

	result, err := executor.Execute(toolCall)
	require.NoError(t, err)

	var queryResult models.QueryResult
	err = json.Unmarshal([]byte(result), &queryResult)
	require.NoError(t, err)
	assert.Equal(t, expectedResult.Columns, queryResult.Columns)
	assert.Equal(t, expectedResult.RowCount, queryResult.RowCount)

	mockDB.AssertExpectations(t)
}

func TestExecutor_ExecuteSQL_MissingArgument(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "execute_sql",
			Arguments: `{}`,
		},
	}

	result, err := executor.Execute(toolCall)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql argument is required")
	assert.Empty(t, result)
}

func TestExecutor_ExecuteSQL_DatabaseError(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	mockDB.On("ExecuteSQL", "INVALID SQL").Return((*models.QueryResult)(nil), errors.New("syntax error"))

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "execute_sql",
			Arguments: `{"sql": "INVALID SQL"}`,
		},
	}

	result, err := executor.Execute(toolCall)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "syntax error")
	assert.Empty(t, result)

	mockDB.AssertExpectations(t)
}

func TestExecutor_GetAllTables(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	expectedTables := []models.TableInfo{
		{TableName: "users", Schema: "public", TableType: "table"},
		{TableName: "orders", Schema: "public", TableType: "table"},
	}

	mockDB.On("GetAllTables").Return(expectedTables, nil)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "get_all_tables",
			Arguments: `{}`,
		},
	}

	result, err := executor.Execute(toolCall)
	require.NoError(t, err)

	var tables []models.TableInfo
	err = json.Unmarshal([]byte(result), &tables)
	require.NoError(t, err)
	assert.Equal(t, expectedTables, tables)

	mockDB.AssertExpectations(t)
}

func TestExecutor_GetTableSchema(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	expectedSchema := []models.ColumnInfo{
		{ColumnName: "id", DataType: "integer", IsPrimaryKey: true},
		{ColumnName: "name", DataType: "varchar", IsNullable: "NO"},
	}

	mockDB.On("GetTableSchema", "users").Return(expectedSchema, nil)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "get_table_schema",
			Arguments: `{"tableName": "users"}`,
		},
	}

	result, err := executor.Execute(toolCall)
	require.NoError(t, err)

	var schema []models.ColumnInfo
	err = json.Unmarshal([]byte(result), &schema)
	require.NoError(t, err)
	assert.Equal(t, expectedSchema, schema)

	mockDB.AssertExpectations(t)
}

func TestExecutor_GetTableSchema_MissingTableName(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "get_table_schema",
			Arguments: `{}`,
		},
	}

	result, err := executor.Execute(toolCall)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tableName argument is required")
	assert.Empty(t, result)
}

func TestExecutor_FindDuplicateData(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	expectedResult := &models.QueryResult{
		Columns:  []string{"email", "count"},
		Rows:     [][]interface{}{{"john@example.com", 2}},
		RowCount: 1,
		Duration: "15ms",
	}

	mockDB.On("FindDuplicateData", "users", []string{"email"}).Return(expectedResult, nil)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "find_duplicate_data",
			Arguments: `{"tableName": "users", "columns": ["email"]}`,
		},
	}

	result, err := executor.Execute(toolCall)
	require.NoError(t, err)

	var queryResult models.QueryResult
	err = json.Unmarshal([]byte(result), &queryResult)
	require.NoError(t, err)
	assert.Equal(t, expectedResult.Columns, queryResult.Columns)
	assert.Equal(t, expectedResult.RowCount, queryResult.RowCount)

	mockDB.AssertExpectations(t)
}

func TestExecutor_FindDuplicateData_MissingColumns(t *testing.T) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "find_duplicate_data",
			Arguments: `{"tableName": "users"}`,
		},
	}

	result, err := executor.Execute(toolCall)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "columns argument is required")
	assert.Empty(t, result)
}

// Benchmark tests
func BenchmarkExecutor_ExecuteSQL(b *testing.B) {
	mockDB := &MockDatabaseInterface{}
	executor := NewExecutor(mockDB)

	expectedResult := &models.QueryResult{
		Columns:  []string{"id", "name"},
		Rows:     [][]interface{}{{1, "John"}},
		RowCount: 1,
		Duration: "5ms",
	}

	mockDB.On("ExecuteSQL", "SELECT * FROM users").Return(expectedResult, nil)

	toolCall := openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      "execute_sql",
			Arguments: `{"sql": "SELECT * FROM users"}`,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.Execute(toolCall)
	}
}
