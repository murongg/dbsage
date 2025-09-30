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
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "analyze_query_performance",
				Description: "Analyze the performance of a specific query and provide optimization suggestions",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The SQL query to analyze",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "suggest_indexes",
				Description: "Suggest indexes for a specific table to improve performance",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"tableName": map[string]interface{}{
							"type":        "string",
							"description": "The name of the table to analyze for index suggestions",
						},
					},
					"required": []string{"tableName"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_query_patterns",
				Description: "Analyze query patterns from database statistics to identify optimization opportunities",
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
				Name:        "optimize_query",
				Description: "Provide optimization suggestions for a specific query",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The SQL query to optimize",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "analyze_table_performance",
				Description: "Analyze performance issues specific to a table",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"tableName": map[string]interface{}{
							"type":        "string",
							"description": "The name of the table to analyze",
						},
					},
					"required": []string{"tableName"},
				},
			},
		},
	}
}
