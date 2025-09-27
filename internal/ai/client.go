package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"dbsage/pkg/database"

	"github.com/sashabaranov/go-openai"
)

// ToolConfirmationConfig defines which tools need confirmation (copied from cli package to avoid circular import)
type ToolConfirmationConfig struct {
	RequiresConfirmation map[string]bool   `json:"requires_confirmation"`
	RiskLevels           map[string]string `json:"risk_levels"`
	Descriptions         map[string]string `json:"descriptions"`
}

type Client struct {
	client              *openai.Client
	dbTools             *database.DatabaseTools
	getDbTools          func() *database.DatabaseTools                                                                                                                                                       // Function to get current database tools
	toolConfirmCallback func(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, callback StreamingCallback) (bool, error) // Tool confirmation callback
	toolConfirmConfig   *ToolConfirmationConfig                                                                                                                                                              // Tool confirmation configuration
}

type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func NewClient(apiKey, baseURL string, dbTools *database.DatabaseTools) *Client {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)
	return &Client{
		client:  client,
		dbTools: dbTools,
	}
}

// NewClientWithDynamicTools creates a new client with dynamic database tools getter
func NewClientWithDynamicTools(apiKey, baseURL string, getDbTools func() *database.DatabaseTools) *Client {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)
	return &Client{
		client:     client,
		getDbTools: getDbTools,
	}
}

// executeToolWithConfirmation executes a tool with confirmation check
func (c *Client) executeToolWithConfirmation(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, callback StreamingCallback) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	// Check if tool requires confirmation based on config
	if c.toolConfirmConfig != nil && c.toolConfirmConfig.RequiresConfirmation[toolCall.Function.Name] {
		if c.toolConfirmCallback != nil {
			confirmed, err := c.toolConfirmCallback(ctx, messages, completeMessage, toolCall, callback)
			if err != nil {
				return "", fmt.Errorf("tool confirmation error: %w", err)
			}
			if !confirmed {
				// Tool confirmation is pending - return a special marker
				return "CONFIRMATION_PENDING", nil
			}
		} else {
			// Debug: callback is nil
			return fmt.Sprintf("DEBUG: toolConfirmCallback is nil for tool: %s", toolCall.Function.Name), nil
		}
	} else {
		// Debug: either config is nil or tool doesn't require confirmation
		configStatus := "nil"
		if c.toolConfirmConfig != nil {
			configStatus = fmt.Sprintf("exists, requires_confirmation=%v", c.toolConfirmConfig.RequiresConfirmation[toolCall.Function.Name])
		}
		return fmt.Sprintf("DEBUG: toolConfirmConfig=%s for tool: %s", configStatus, toolCall.Function.Name), nil
	}

	// Execute the tool normally
	return c.executeTool(toolCall)
}

func (c *Client) executeTool(toolCall openai.ToolCall) (string, error) {
	// Get current database tools (either static or dynamic)
	var dbTools *database.DatabaseTools
	if c.getDbTools != nil {
		dbTools = c.getDbTools()
	} else {
		dbTools = c.dbTools
	}

	// Check if database tools are available
	if dbTools == nil {
		return `{"error": "No database connection available. Please add and switch to a database connection first using the /add command."}`, nil
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	// Check if tool requires confirmation - for now, we'll handle this differently
	// The confirmation will be handled in the streaming response processing

	switch toolCall.Function.Name {
	case "execute_sql":
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

	case "get_all_tables":
		tables, err := dbTools.GetAllTables()
		if err != nil {
			return "", err
		}
		resultJSON, _ := json.Marshal(tables)
		return string(resultJSON), nil

	case "get_table_schema":
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

	case "explain_query":
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

	case "get_table_indexes":
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

	case "get_table_stats":
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

	case "find_duplicate_data":
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

	case "get_slow_queries":
		queries, err := dbTools.GetSlowQueries()
		if err != nil {
			return "", err
		}
		resultJSON, _ := json.Marshal(queries)
		return string(resultJSON), nil

	case "get_database_size":
		size, err := dbTools.GetDatabaseSize()
		if err != nil {
			return "", err
		}
		resultJSON, _ := json.Marshal(size)
		return string(resultJSON), nil

	case "get_table_sizes":
		sizes, err := dbTools.GetTableSizes()
		if err != nil {
			return "", err
		}
		resultJSON, _ := json.Marshal(sizes)
		return string(resultJSON), nil

	case "get_active_connections":
		connections, err := dbTools.GetActiveConnections()
		if err != nil {
			return "", err
		}
		resultJSON, _ := json.Marshal(connections)
		return string(resultJSON), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}
}

func (c *Client) QueryWithTools(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {

	systemMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: GetSystemPrompt(),
	}

	allMessages := append([]openai.ChatCompletionMessage{systemMessage}, messages...)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: allMessages,
		Tools:    GetTools(),
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	message := resp.Choices[0].Message

	if len(message.ToolCalls) > 0 {
		toolCall := message.ToolCalls[0]

		result, err := c.executeTool(toolCall)
		if err != nil {
			return "", fmt.Errorf("tool execution error: %w", err)
		}

		updatedMessages := append(messages,
			message,
			openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			},
		)

		return c.QueryWithTools(ctx, updatedMessages)
	}

	return message.Content, nil
}

// StreamingCallback is called for each chunk of streaming response
type StreamingCallback func(chunk string) error

// QueryWithToolsStreaming performs a streaming query with tools support
func (c *Client) QueryWithToolsStreaming(ctx context.Context, messages []openai.ChatCompletionMessage, callback StreamingCallback) error {
	systemMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: GetSystemPrompt(),
	}

	allMessages := append([]openai.ChatCompletionMessage{systemMessage}, messages...)

	// Create streaming request with tools
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: allMessages,
		Tools:    GetTools(),
		Stream:   true,
	})
	if err != nil {
		return fmt.Errorf("OpenAI streaming API error: %w", err)
	}
	defer stream.Close()

	// Collect the complete response to check for tool calls
	var completeMessage openai.ChatCompletionMessage
	var toolCalls []openai.ToolCall
	var contentBuilder strings.Builder

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta

			// Accumulate content
			if delta.Content != "" {
				contentBuilder.WriteString(delta.Content)
				// Send content chunks to callback for streaming display
				if err := callback(delta.Content); err != nil {
					return err
				}
			}

			// Accumulate tool calls
			if len(delta.ToolCalls) > 0 {
				for _, toolCall := range delta.ToolCalls {
					// If this is a new tool call or we need to extend existing ones
					if toolCall.Index != nil {
						// Ensure we have enough space in the slice
						for len(toolCalls) <= *toolCall.Index {
							toolCalls = append(toolCalls, openai.ToolCall{})
						}

						// Update the tool call at the specified index
						if toolCall.ID != "" {
							toolCalls[*toolCall.Index].ID = toolCall.ID
						}
						if toolCall.Type != "" {
							toolCalls[*toolCall.Index].Type = toolCall.Type
						}
						if toolCall.Function.Name != "" {
							if toolCalls[*toolCall.Index].Function.Name == "" {
								toolCalls[*toolCall.Index].Function = openai.FunctionCall{}
							}
							toolCalls[*toolCall.Index].Function.Name = toolCall.Function.Name
						}
						if toolCall.Function.Arguments != "" {
							if toolCalls[*toolCall.Index].Function.Name == "" && toolCalls[*toolCall.Index].Function.Arguments == "" {
								toolCalls[*toolCall.Index].Function = openai.FunctionCall{}
							}
							toolCalls[*toolCall.Index].Function.Arguments += toolCall.Function.Arguments
						}
					}
				}
			}
		}
	}

	// Build complete message
	completeMessage.Role = openai.ChatMessageRoleAssistant
	completeMessage.Content = contentBuilder.String()
	completeMessage.ToolCalls = toolCalls

	// If tools are needed, execute them
	if len(toolCalls) > 0 {
		toolCall := toolCalls[0]
		// Execute the tool with confirmation check
		result, err := c.executeToolWithConfirmation(ctx, messages, completeMessage, toolCall, callback)
		if err != nil {
			return fmt.Errorf("tool execution error: %w", err)
		}

		// Check if confirmation is pending
		if result == "CONFIRMATION_PENDING" {
			// Tool confirmation is pending - don't continue execution
			// The UI will handle the confirmation and call ContinueWithConfirmedTool when ready
			return nil
		}

		updatedMessages := append(messages,
			completeMessage,
			openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			},
		)

		// Recursively call with updated messages
		return c.QueryWithToolsStreaming(ctx, updatedMessages, callback)
	}

	// No tools needed, streaming is already complete
	return nil
}

// SetToolConfirmationCallback sets the tool confirmation callback
func (c *Client) SetToolConfirmationCallback(callback func(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, streamingCallback StreamingCallback) (bool, error)) {
	c.toolConfirmCallback = callback
}

// SetToolConfirmationConfig sets the tool confirmation configuration
func (c *Client) SetToolConfirmationConfig(config *ToolConfirmationConfig) {
	c.toolConfirmConfig = config
}

// NewToolConfirmationConfig creates a new tool confirmation config
func NewToolConfirmationConfig(requiresConfirmation map[string]bool, riskLevels map[string]string, descriptions map[string]string) *ToolConfirmationConfig {
	return &ToolConfirmationConfig{
		RequiresConfirmation: requiresConfirmation,
		RiskLevels:           riskLevels,
		Descriptions:         descriptions,
	}
}

// ContinueWithConfirmedTool continues AI processing after tool confirmation
func (c *Client) ContinueWithConfirmedTool(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, callback StreamingCallback) error {
	// Execute the confirmed tool
	result, err := c.executeTool(toolCall)
	if err != nil {
		return fmt.Errorf("tool execution error: %w", err)
	}

	updatedMessages := append(messages,
		completeMessage,
		openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    result,
			ToolCallID: toolCall.ID,
		},
	)

	// Continue with streaming
	return c.QueryWithToolsStreaming(ctx, updatedMessages, callback)
}
