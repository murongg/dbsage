package ai

import "github.com/sashabaranov/go-openai"

// GetTools returns the available tools for the database AI assistant
func GetTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "execute_sql",
				Description: "Execute a SQL query",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"sql": map[string]interface{}{
							"type":        "string",
							"description": "The SQL query to execute",
						},
					},
					"required": []string{"sql"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_all_tables",
				Description: "Get all tables in the database",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_table_schema",
				Description: "Get detailed schema of a table including columns, types, nullability, defaults",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"tableName": map[string]interface{}{
							"type":        "string",
							"description": "The name of the table",
						},
					},
					"required": []string{"tableName"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "explain_query",
				Description: "Analyze query performance with EXPLAIN ANALYZE",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"sql": map[string]interface{}{
							"type":        "string",
							"description": "The SQL query to analyze",
						},
					},
					"required": []string{"sql"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_table_indexes",
				Description: "Get all indexes for a specific table",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"tableName": map[string]interface{}{
							"type":        "string",
							"description": "The name of the table",
						},
					},
					"required": []string{"tableName"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_table_stats",
				Description: "Get statistical information about table columns",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"tableName": map[string]interface{}{
							"type":        "string",
							"description": "The name of the table",
						},
					},
					"required": []string{"tableName"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "find_duplicate_data",
				Description: "Find duplicate records in a table based on specified columns",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"tableName": map[string]interface{}{
							"type":        "string",
							"description": "The name of the table",
						},
						"columns": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
							"description": "Array of column names to check for duplicates",
						},
					},
					"required": []string{"tableName", "columns"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_slow_queries",
				Description: "Get the slowest queries from pg_stat_statements",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_database_size",
				Description: "Get the size of the current database",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_table_sizes",
				Description: "Get sizes of all tables including table and index sizes",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_active_connections",
				Description: "Get information about active database connections",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
	}
}
