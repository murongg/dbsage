package tools

import (
	"encoding/json"
	"fmt"

	"dbsage/pkg/dbinterfaces"

	"github.com/sashabaranov/go-openai"
)

// Executor handles tool execution
type Executor struct {
	dbTools    dbinterfaces.DatabaseInterface
	getDbTools func() dbinterfaces.DatabaseInterface
}

func NewExecutor(dbTools dbinterfaces.DatabaseInterface) *Executor {
	return &Executor{dbTools: dbTools}
}

func NewExecutorWithDynamicTools(getDbTools func() dbinterfaces.DatabaseInterface) *Executor {
	return &Executor{getDbTools: getDbTools}
}

// Execute executes a tool call
func (e *Executor) Execute(toolCall openai.ToolCall) (string, error) {
	// Get current database tools (either static or dynamic)
	var dbTools dbinterfaces.DatabaseInterface
	if e.getDbTools != nil {
		dbTools = e.getDbTools()
	} else {
		dbTools = e.dbTools
	}

	// Check if database tools are available
	if dbTools == nil {
		return `{"error": "No database connection available. Please add and switch to a database connection first using the /add command."}`, nil
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	switch toolCall.Function.Name {
	case "execute_sql":
		return e.executeSQL(dbTools, args)
	case "get_all_tables":
		return e.getAllTables(dbTools)
	case "get_table_schema":
		return e.getTableSchema(dbTools, args)
	case "explain_query":
		return e.explainQuery(dbTools, args)
	case "get_table_indexes":
		return e.getTableIndexes(dbTools, args)
	case "get_table_stats":
		return e.getTableStats(dbTools, args)
	case "find_duplicate_data":
		return e.findDuplicateData(dbTools, args)
	case "get_slow_queries":
		return e.getSlowQueries(dbTools)
	case "get_database_size":
		return e.getDatabaseSize(dbTools)
	case "get_table_sizes":
		return e.getTableSizes(dbTools)
	case "get_active_connections":
		return e.getActiveConnections(dbTools)
	case "analyze_query_performance":
		return e.analyzeQueryPerformance(dbTools, args)
	case "suggest_indexes":
		return e.suggestIndexes(dbTools, args)
	case "get_query_patterns":
		return e.getQueryPatterns(dbTools)
	case "optimize_query":
		return e.optimizeQuery(dbTools, args)
	case "analyze_table_performance":
		return e.analyzeTablePerformance(dbTools, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}
}

func (e *Executor) executeSQL(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	sql, ok := args["sql"].(string)
	if !ok {
		return "", fmt.Errorf("sql argument is required and must be a string")
	}
	result, err := dbTools.ExecuteSQL(sql)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (e *Executor) getAllTables(dbTools dbinterfaces.DatabaseInterface) (string, error) {
	tables, err := dbTools.GetAllTables()
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(tables)
	return string(resultJSON), nil
}

func (e *Executor) getTableSchema(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	tableName, ok := args["tableName"].(string)
	if !ok {
		return "", fmt.Errorf("tableName argument is required and must be a string")
	}
	schema, err := dbTools.GetTableSchema(tableName)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(schema)
	return string(resultJSON), nil
}

func (e *Executor) explainQuery(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	sql, ok := args["sql"].(string)
	if !ok {
		return "", fmt.Errorf("sql argument is required and must be a string")
	}
	result, err := dbTools.ExplainQuery(sql)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (e *Executor) getTableIndexes(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	tableName, ok := args["tableName"].(string)
	if !ok {
		return "", fmt.Errorf("tableName argument is required and must be a string")
	}
	indexes, err := dbTools.GetTableIndexes(tableName)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(indexes)
	return string(resultJSON), nil
}

func (e *Executor) getTableStats(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	tableName, ok := args["tableName"].(string)
	if !ok {
		return "", fmt.Errorf("tableName argument is required and must be a string")
	}
	stats, err := dbTools.GetTableStats(tableName)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(stats)
	return string(resultJSON), nil
}

func (e *Executor) findDuplicateData(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	tableName, ok := args["tableName"].(string)
	if !ok {
		return "", fmt.Errorf("tableName argument is required and must be a string")
	}
	columnsInterface, ok := args["columns"].([]interface{})
	if !ok {
		return "", fmt.Errorf("columns argument is required and must be an array")
	}
	var columns []string
	for _, col := range columnsInterface {
		if colStr, ok := col.(string); ok {
			columns = append(columns, colStr)
		}
	}
	result, err := dbTools.FindDuplicateData(tableName, columns)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (e *Executor) getSlowQueries(dbTools dbinterfaces.DatabaseInterface) (string, error) {
	queries, err := dbTools.GetSlowQueries()
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(queries)
	return string(resultJSON), nil
}

func (e *Executor) getDatabaseSize(dbTools dbinterfaces.DatabaseInterface) (string, error) {
	size, err := dbTools.GetDatabaseSize()
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(size)
	return string(resultJSON), nil
}

func (e *Executor) getTableSizes(dbTools dbinterfaces.DatabaseInterface) (string, error) {
	sizes, err := dbTools.GetTableSizes()
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(sizes)
	return string(resultJSON), nil
}

func (e *Executor) getActiveConnections(dbTools dbinterfaces.DatabaseInterface) (string, error) {
	connections, err := dbTools.GetActiveConnections()
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(connections)
	return string(resultJSON), nil
}

func (e *Executor) analyzeQueryPerformance(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query argument is required and must be a string")
	}
	analysis, err := dbTools.AnalyzeQueryPerformance(query)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(analysis)
	return string(resultJSON), nil
}

func (e *Executor) suggestIndexes(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	tableName, ok := args["tableName"].(string)
	if !ok {
		return "", fmt.Errorf("tableName argument is required and must be a string")
	}
	suggestions, err := dbTools.SuggestIndexes(tableName)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(suggestions)
	return string(resultJSON), nil
}

func (e *Executor) getQueryPatterns(dbTools dbinterfaces.DatabaseInterface) (string, error) {
	patterns, err := dbTools.GetQueryPatterns()
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(patterns)
	return string(resultJSON), nil
}

func (e *Executor) optimizeQuery(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query argument is required and must be a string")
	}
	suggestions, err := dbTools.OptimizeQuery(query)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(suggestions)
	return string(resultJSON), nil
}

func (e *Executor) analyzeTablePerformance(dbTools dbinterfaces.DatabaseInterface, args map[string]interface{}) (string, error) {
	tableName, ok := args["tableName"].(string)
	if !ok {
		return "", fmt.Errorf("tableName argument is required and must be a string")
	}
	analysis, err := dbTools.AnalyzeTablePerformance(tableName)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(analysis)
	return string(resultJSON), nil
}
