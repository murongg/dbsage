package database

import (
	"errors"
	"testing"

	"dbsage/internal/models"
	"dbsage/pkg/dbinterfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDatabaseInterface is a mock implementation of DatabaseInterface
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

func (m *MockDatabaseInterface) GetTableStats(tableName string) (*models.TableStats, error) {
	args := m.Called(tableName)
	return args.Get(0).(*models.TableStats), args.Error(1)
}

func (m *MockDatabaseInterface) GetTableSizes() ([]map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDatabaseInterface) GetSlowQueries() ([]models.SlowQuery, error) {
	args := m.Called()
	return args.Get(0).([]models.SlowQuery), args.Error(1)
}

func (m *MockDatabaseInterface) GetDatabaseSize() (*models.DatabaseSize, error) {
	args := m.Called()
	return args.Get(0).(*models.DatabaseSize), args.Error(1)
}

func (m *MockDatabaseInterface) GetActiveConnections() ([]models.ActiveConnection, error) {
	args := m.Called()
	return args.Get(0).([]models.ActiveConnection), args.Error(1)
}

func (m *MockDatabaseInterface) AnalyzeQueryPerformance(query string) (*models.PerformanceAnalysis, error) {
	args := m.Called(query)
	return args.Get(0).(*models.PerformanceAnalysis), args.Error(1)
}

func (m *MockDatabaseInterface) SuggestIndexes(tableName string) ([]models.IndexSuggestion, error) {
	args := m.Called(tableName)
	return args.Get(0).([]models.IndexSuggestion), args.Error(1)
}

func (m *MockDatabaseInterface) GetQueryPatterns() ([]models.QueryPattern, error) {
	args := m.Called()
	return args.Get(0).([]models.QueryPattern), args.Error(1)
}

func (m *MockDatabaseInterface) OptimizeQuery(query string) ([]models.QueryOptimizationSuggestion, error) {
	args := m.Called(query)
	return args.Get(0).([]models.QueryOptimizationSuggestion), args.Error(1)
}

func (m *MockDatabaseInterface) AnalyzeTablePerformance(tableName string) (*models.PerformanceAnalysis, error) {
	args := m.Called(tableName)
	return args.Get(0).(*models.PerformanceAnalysis), args.Error(1)
}

// MockConnectionManager is a mock implementation of ConnectionManagerInterface
type MockConnectionManager struct {
	mock.Mock
}

func (m *MockConnectionManager) AddConnection(config *dbinterfaces.ConnectionConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockConnectionManager) RemoveConnection(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockConnectionManager) ListConnections() map[string]*dbinterfaces.ConnectionConfig {
	args := m.Called()
	return args.Get(0).(map[string]*dbinterfaces.ConnectionConfig)
}

func (m *MockConnectionManager) GetCurrentConnection() (dbinterfaces.DatabaseInterface, string, error) {
	args := m.Called()
	return args.Get(0).(dbinterfaces.DatabaseInterface), args.String(1), args.Error(2)
}

func (m *MockConnectionManager) SwitchConnection(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockConnectionManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnectionManager) GetConnectionStatus() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func (m *MockConnectionManager) GetLastUsedConnection() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConnectionManager) GetConnectionsSortedByLastUsed() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestConnectionService_GetCurrentTools(t *testing.T) {
	mockManager := &MockConnectionManager{}
	mockDB := &MockDatabaseInterface{}

	service := &ConnectionService{
		manager: mockManager,
		current: mockDB,
	}

	// Test healthy connection
	mockDB.On("IsConnectionHealthy").Return(true)
	result := service.GetCurrentTools()
	assert.Equal(t, mockDB, result)

	// Test unhealthy connection with successful refresh
	mockDB.On("IsConnectionHealthy").Return(false)
	newMockDB := &MockDatabaseInterface{}
	mockManager.On("GetCurrentConnection").Return(newMockDB, "test_conn", nil)

	result = service.GetCurrentTools()
	assert.Equal(t, newMockDB, result)
	assert.Equal(t, newMockDB, service.current)

	mockDB.AssertExpectations(t)
	mockManager.AssertExpectations(t)
}

func TestConnectionService_GetCurrentTools_UnhealthyConnectionFailedRefresh(t *testing.T) {
	mockManager := &MockConnectionManager{}
	mockDB := &MockDatabaseInterface{}

	service := &ConnectionService{
		manager: mockManager,
		current: mockDB,
	}

	// Test unhealthy connection with failed refresh
	mockDB.On("IsConnectionHealthy").Return(false)
	mockManager.On("GetCurrentConnection").Return(nil, "", errors.New("connection failed"))

	result := service.GetCurrentTools()
	assert.Nil(t, result)
	assert.Nil(t, service.current)

	mockDB.AssertExpectations(t)
	mockManager.AssertExpectations(t)
}

func TestConnectionService_AddConnection(t *testing.T) {
	mockManager := &MockConnectionManager{}
	service := &ConnectionService{
		manager: mockManager,
		current: nil,
	}

	config := &dbinterfaces.ConnectionConfig{
		Name:     "test_db",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "password",
	}

	// Test successful add connection (first connection)
	mockDB := &MockDatabaseInterface{}
	mockManager.On("AddConnection", config).Return(nil)
	mockManager.On("GetCurrentConnection").Return(mockDB, "test_db", nil)

	err := service.AddConnection(config)
	require.NoError(t, err)
	assert.Equal(t, mockDB, service.current)

	mockManager.AssertExpectations(t)
}

func TestConnectionService_AddConnection_Error(t *testing.T) {
	mockManager := &MockConnectionManager{}
	service := &ConnectionService{
		manager: mockManager,
		current: nil,
	}

	config := &dbinterfaces.ConnectionConfig{
		Name: "test_db",
	}

	// Test failed add connection
	expectedError := errors.New("connection failed")
	mockManager.On("AddConnection", config).Return(expectedError)

	err := service.AddConnection(config)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	mockManager.AssertExpectations(t)
}

func TestConnectionService_SwitchConnection(t *testing.T) {
	mockManager := &MockConnectionManager{}
	service := &ConnectionService{
		manager: mockManager,
	}

	// Test successful switch
	mockDB := &MockDatabaseInterface{}
	mockManager.On("SwitchConnection", "new_db").Return(nil)
	mockManager.On("GetCurrentConnection").Return(mockDB, "new_db", nil)

	err := service.SwitchConnection("new_db")
	require.NoError(t, err)
	assert.Equal(t, mockDB, service.current)

	mockManager.AssertExpectations(t)
}

func TestConnectionService_SwitchConnection_ManagerError(t *testing.T) {
	mockManager := &MockConnectionManager{}
	service := &ConnectionService{
		manager: mockManager,
	}

	// Test switch error from manager
	expectedError := errors.New("switch failed")
	mockManager.On("SwitchConnection", "nonexistent_db").Return(expectedError)

	err := service.SwitchConnection("nonexistent_db")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	mockManager.AssertExpectations(t)
}

func TestConnectionService_SwitchConnection_GetCurrentError(t *testing.T) {
	mockManager := &MockConnectionManager{}
	service := &ConnectionService{
		manager: mockManager,
	}

	// Test successful switch but failed to get current connection
	mockManager.On("SwitchConnection", "new_db").Return(nil)
	mockManager.On("GetCurrentConnection").Return(nil, "", errors.New("get current failed"))

	err := service.SwitchConnection("new_db")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update current connection after switch")

	mockManager.AssertExpectations(t)
}

func TestConnectionService_RemoveConnection(t *testing.T) {
	mockManager := &MockConnectionManager{}
	mockDB := &MockDatabaseInterface{}
	service := &ConnectionService{
		manager: mockManager,
		current: mockDB,
	}

	// Test successful remove with new current connection
	newMockDB := &MockDatabaseInterface{}
	mockManager.On("RemoveConnection", "old_db").Return(nil)
	mockManager.On("GetCurrentConnection").Return(newMockDB, "remaining_db", nil)

	err := service.RemoveConnection("old_db")
	require.NoError(t, err)
	assert.Equal(t, newMockDB, service.current)

	mockManager.AssertExpectations(t)
}

func TestConnectionService_RemoveConnection_NoCurrentLeft(t *testing.T) {
	mockManager := &MockConnectionManager{}
	mockDB := &MockDatabaseInterface{}
	service := &ConnectionService{
		manager: mockManager,
		current: mockDB,
	}

	// Test successful remove but no connections left
	mockManager.On("RemoveConnection", "last_db").Return(nil)
	mockManager.On("GetCurrentConnection").Return(nil, "", errors.New("no connections"))

	err := service.RemoveConnection("last_db")
	require.NoError(t, err)
	assert.Nil(t, service.current)

	mockManager.AssertExpectations(t)
}

func TestConnectionService_GetConnectionInfo(t *testing.T) {
	mockManager := &MockConnectionManager{}
	service := &ConnectionService{
		manager: mockManager,
	}

	expectedConnections := map[string]*dbinterfaces.ConnectionConfig{
		"db1": {Name: "db1"},
		"db2": {Name: "db2"},
	}
	expectedStatus := map[string]string{
		"db1": "connected",
		"db2": "disconnected",
	}

	mockManager.On("ListConnections").Return(expectedConnections)
	mockManager.On("GetConnectionStatus").Return(expectedStatus)
	mockManager.On("GetCurrentConnection").Return(nil, "db1", nil)

	connections, status, current := service.GetConnectionInfo()

	assert.Equal(t, expectedConnections, connections)
	assert.Equal(t, expectedStatus, status)
	assert.Equal(t, "db1", current)

	mockManager.AssertExpectations(t)
}

func TestConnectionService_IsConnected(t *testing.T) {
	service := &ConnectionService{}

	// Test with no current connection
	assert.False(t, service.IsConnected())

	// Test with current connection
	mockDB := &MockDatabaseInterface{}
	service.current = mockDB
	assert.True(t, service.IsConnected())
}

func TestConnectionService_IsConnectionHealthy(t *testing.T) {
	service := &ConnectionService{}

	// Test with no current connection
	assert.False(t, service.IsConnectionHealthy())

	// Test with healthy connection
	mockDB := &MockDatabaseInterface{}
	mockDB.On("IsConnectionHealthy").Return(true)
	service.current = mockDB
	assert.True(t, service.IsConnectionHealthy())

	// Test with unhealthy connection
	mockDB.On("IsConnectionHealthy").Return(false)
	assert.False(t, service.IsConnectionHealthy())

	mockDB.AssertExpectations(t)
}

func TestConnectionService_EnsureHealthyConnection(t *testing.T) {
	mockManager := &MockConnectionManager{}
	mockDB := &MockDatabaseInterface{}
	service := &ConnectionService{
		manager: mockManager,
		current: mockDB,
	}

	// Test with healthy connection
	mockDB.On("IsConnectionHealthy").Return(true)
	err := service.EnsureHealthyConnection()
	require.NoError(t, err)

	// Test with unhealthy connection that can be restored
	mockDB.On("IsConnectionHealthy").Return(false)
	newMockDB := &MockDatabaseInterface{}
	mockManager.On("GetCurrentConnection").Return(newMockDB, "restored_db", nil)

	err = service.EnsureHealthyConnection()
	require.NoError(t, err)
	assert.Equal(t, newMockDB, service.current)

	mockDB.AssertExpectations(t)
	mockManager.AssertExpectations(t)
}

func TestConnectionService_EnsureHealthyConnection_NoConnection(t *testing.T) {
	service := &ConnectionService{
		current: nil,
	}

	err := service.EnsureHealthyConnection()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active database connection")
}

func TestConnectionService_EnsureHealthyConnection_RestoreFailed(t *testing.T) {
	mockManager := &MockConnectionManager{}
	mockDB := &MockDatabaseInterface{}
	service := &ConnectionService{
		manager: mockManager,
		current: mockDB,
	}

	// Test with unhealthy connection that cannot be restored
	mockDB.On("IsConnectionHealthy").Return(false)
	mockManager.On("GetCurrentConnection").Return(nil, "", errors.New("restore failed"))

	err := service.EnsureHealthyConnection()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to restore healthy connection")
	assert.Nil(t, service.current)

	mockDB.AssertExpectations(t)
	mockManager.AssertExpectations(t)
}

func TestConnectionService_GetConnectionStats(t *testing.T) {
	mockManager := &MockConnectionManager{}
	mockDB := &MockDatabaseInterface{}
	service := &ConnectionService{
		manager: mockManager,
		current: mockDB,
	}

	connections := map[string]*dbinterfaces.ConnectionConfig{
		"db1": {Name: "db1"},
		"db2": {Name: "db2"},
		"db3": {Name: "db3"},
	}
	status := map[string]string{
		"db1": "active",
		"db2": "connected",
		"db3": "disconnected",
	}

	mockManager.On("ListConnections").Return(connections)
	mockManager.On("GetConnectionStatus").Return(status)
	mockDB.On("IsConnectionHealthy").Return(true)

	stats := service.GetConnectionStats()

	assert.Equal(t, 3, stats["total_connections"])
	assert.Equal(t, 1, stats["active_connections"])
	assert.Equal(t, 1, stats["connected_connections"])
	assert.Equal(t, 0, stats["unhealthy_connections"])
	assert.Equal(t, 1, stats["disconnected_connections"])
	assert.Equal(t, true, stats["has_current"])
	assert.Equal(t, true, stats["current_is_healthy"])

	mockManager.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestConnectionService_Close(t *testing.T) {
	mockManager := &MockConnectionManager{}
	mockDB := &MockDatabaseInterface{}
	service := &ConnectionService{
		manager: mockManager,
		current: mockDB,
	}

	mockManager.On("Close").Return(nil)

	err := service.Close()
	require.NoError(t, err)
	assert.Nil(t, service.current)

	mockManager.AssertExpectations(t)
}
