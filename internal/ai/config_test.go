package ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToolConfirmationConfig(t *testing.T) {
	requiresConfirmation := map[string]bool{
		"execute_sql":    true,
		"get_all_tables": false,
	}
	riskLevels := map[string]string{
		"execute_sql":    "high",
		"get_all_tables": "low",
	}
	descriptions := map[string]string{
		"execute_sql":    "Executes arbitrary SQL queries",
		"get_all_tables": "Lists all database tables",
	}

	config := NewToolConfirmationConfig(requiresConfirmation, riskLevels, descriptions)

	assert.NotNil(t, config)
	assert.Equal(t, requiresConfirmation, config.RequiresConfirmation)
	assert.Equal(t, riskLevels, config.RiskLevels)
	assert.Equal(t, descriptions, config.Descriptions)
}

func TestToolConfirmationConfig_JSON(t *testing.T) {
	config := &ToolConfirmationConfig{
		RequiresConfirmation: map[string]bool{
			"execute_sql":    true,
			"get_all_tables": false,
		},
		RiskLevels: map[string]string{
			"execute_sql":    "high",
			"get_all_tables": "low",
		},
		Descriptions: map[string]string{
			"execute_sql":    "Executes arbitrary SQL queries",
			"get_all_tables": "Lists all database tables",
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(config)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledConfig ToolConfirmationConfig
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	require.NoError(t, err)

	assert.Equal(t, config.RequiresConfirmation, unmarshaledConfig.RequiresConfirmation)
	assert.Equal(t, config.RiskLevels, unmarshaledConfig.RiskLevels)
	assert.Equal(t, config.Descriptions, unmarshaledConfig.Descriptions)
}

func TestToolConfirmationConfig_EmptyValues(t *testing.T) {
	// Test with empty/nil values
	config := NewToolConfirmationConfig(nil, nil, nil)

	assert.NotNil(t, config)
	assert.Nil(t, config.RequiresConfirmation)
	assert.Nil(t, config.RiskLevels)
	assert.Nil(t, config.Descriptions)
}

func TestToolConfirmationConfig_CheckRequiresConfirmation(t *testing.T) {
	config := &ToolConfirmationConfig{
		RequiresConfirmation: map[string]bool{
			"execute_sql":         true,
			"get_all_tables":      false,
			"get_table_schema":    false,
			"explain_query":       true,
			"get_table_indexes":   false,
			"get_table_stats":     false,
			"find_duplicate_data": true,
		},
	}

	// Test tools that require confirmation
	assert.True(t, config.RequiresConfirmation["execute_sql"])
	assert.True(t, config.RequiresConfirmation["explain_query"])
	assert.True(t, config.RequiresConfirmation["find_duplicate_data"])

	// Test tools that don't require confirmation
	assert.False(t, config.RequiresConfirmation["get_all_tables"])
	assert.False(t, config.RequiresConfirmation["get_table_schema"])
	assert.False(t, config.RequiresConfirmation["get_table_indexes"])
	assert.False(t, config.RequiresConfirmation["get_table_stats"])

	// Test non-existent tool (should return false)
	assert.False(t, config.RequiresConfirmation["non_existent_tool"])
}

func TestToolConfirmationConfig_RiskLevels(t *testing.T) {
	config := &ToolConfirmationConfig{
		RiskLevels: map[string]string{
			"execute_sql":         "high",
			"explain_query":       "medium",
			"find_duplicate_data": "medium",
			"get_all_tables":      "low",
			"get_table_schema":    "low",
			"get_table_stats":     "low",
		},
	}

	// Test high risk tools
	assert.Equal(t, "high", config.RiskLevels["execute_sql"])

	// Test medium risk tools
	assert.Equal(t, "medium", config.RiskLevels["explain_query"])
	assert.Equal(t, "medium", config.RiskLevels["find_duplicate_data"])

	// Test low risk tools
	assert.Equal(t, "low", config.RiskLevels["get_all_tables"])
	assert.Equal(t, "low", config.RiskLevels["get_table_schema"])
	assert.Equal(t, "low", config.RiskLevels["get_table_stats"])

	// Test non-existent tool
	assert.Equal(t, "", config.RiskLevels["non_existent_tool"])
}

func TestToolConfirmationConfig_Descriptions(t *testing.T) {
	config := &ToolConfirmationConfig{
		Descriptions: map[string]string{
			"execute_sql":    "Executes arbitrary SQL queries on the database",
			"get_all_tables": "Retrieves a list of all tables in the database",
			"explain_query":  "Shows the execution plan for a SQL query",
		},
	}

	// Test that descriptions are properly set
	assert.Equal(t, "Executes arbitrary SQL queries on the database", config.Descriptions["execute_sql"])
	assert.Equal(t, "Retrieves a list of all tables in the database", config.Descriptions["get_all_tables"])
	assert.Equal(t, "Shows the execution plan for a SQL query", config.Descriptions["explain_query"])

	// Test non-existent tool
	assert.Equal(t, "", config.Descriptions["non_existent_tool"])
}

func TestToolConfirmationConfig_CompleteExample(t *testing.T) {
	// Test a complete realistic configuration
	config := NewToolConfirmationConfig(
		map[string]bool{
			"execute_sql":         true,
			"explain_query":       true,
			"find_duplicate_data": true,
			"get_all_tables":      false,
			"get_table_schema":    false,
			"get_table_stats":     false,
		},
		map[string]string{
			"execute_sql":         "high",
			"explain_query":       "medium",
			"find_duplicate_data": "medium",
			"get_all_tables":      "low",
			"get_table_schema":    "low",
			"get_table_stats":     "low",
		},
		map[string]string{
			"execute_sql":         "Executes arbitrary SQL queries - can modify data",
			"explain_query":       "Analyzes query execution plan - read-only operation",
			"find_duplicate_data": "Searches for duplicate records - read-only but resource intensive",
			"get_all_tables":      "Lists database tables - safe read-only operation",
			"get_table_schema":    "Shows table structure - safe read-only operation",
			"get_table_stats":     "Shows table statistics - safe read-only operation",
		},
	)

	// Verify configuration is consistent
	for tool := range config.RequiresConfirmation {
		assert.Contains(t, config.RiskLevels, tool, "Tool %s missing from risk levels", tool)
		assert.Contains(t, config.Descriptions, tool, "Tool %s missing from descriptions", tool)
	}

	// Verify high-risk tools require confirmation
	for tool, riskLevel := range config.RiskLevels {
		if riskLevel == "high" {
			assert.True(t, config.RequiresConfirmation[tool],
				"High-risk tool %s should require confirmation", tool)
		}
	}
}

func BenchmarkNewToolConfirmationConfig(b *testing.B) {
	requiresConfirmation := map[string]bool{
		"execute_sql":    true,
		"get_all_tables": false,
	}
	riskLevels := map[string]string{
		"execute_sql":    "high",
		"get_all_tables": "low",
	}
	descriptions := map[string]string{
		"execute_sql":    "Executes arbitrary SQL queries",
		"get_all_tables": "Lists all database tables",
	}

	for i := 0; i < b.N; i++ {
		NewToolConfirmationConfig(requiresConfirmation, riskLevels, descriptions)
	}
}

func BenchmarkToolConfirmationConfigJSON(b *testing.B) {
	config := &ToolConfirmationConfig{
		RequiresConfirmation: map[string]bool{
			"execute_sql":    true,
			"get_all_tables": false,
		},
		RiskLevels: map[string]string{
			"execute_sql":    "high",
			"get_all_tables": "low",
		},
		Descriptions: map[string]string{
			"execute_sql":    "Executes arbitrary SQL queries",
			"get_all_tables": "Lists all database tables",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(config)
	}
}
