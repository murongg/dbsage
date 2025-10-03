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
	case "find_duplicate_data":
		return e.findDuplicateData(dbTools, args)
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
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SQL result: %w", err)
	}
	return string(resultJSON), nil
}

func (e *Executor) getAllTables(dbTools dbinterfaces.DatabaseInterface) (string, error) {
	tables, err := dbTools.GetAllTables()
	if err != nil {
		return "", err
	}
	resultJSON, err := json.Marshal(tables)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tables: %w", err)
	}
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
	resultJSON, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}
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
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal explain result: %w", err)
	}
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
	resultJSON, err := json.Marshal(indexes)
	if err != nil {
		return "", fmt.Errorf("failed to marshal indexes: %w", err)
	}
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
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal duplicate result: %w", err)
	}
	return string(resultJSON), nil
}
