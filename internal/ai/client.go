package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"dbsage/internal/ai/streaming"
	"dbsage/internal/ai/tools"
	"dbsage/pkg/dbinterfaces"

	"github.com/sashabaranov/go-openai"
)

// StreamingCallback is called for each chunk of streaming response
type StreamingCallback func(chunk string) error

type Client struct {
	client              *openai.Client
	toolExecutor        *tools.Executor
	streamingHandler    *streaming.StreamingHandler
	toolConfirmCallback func(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, callback StreamingCallback) (bool, error)
	toolConfirmConfig   *ToolConfirmationConfig
}

func NewClient(apiKey, baseURL string, dbTools dbinterfaces.DatabaseInterface) *Client {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)
	return &Client{
		client:           client,
		toolExecutor:     tools.NewExecutor(dbTools),
		streamingHandler: streaming.NewStreamingHandler(),
	}
}

// NewClientWithDynamicTools creates a new client with dynamic database tools getter
func NewClientWithDynamicTools(apiKey, baseURL string, getDbTools func() dbinterfaces.DatabaseInterface) *Client {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)
	return &Client{
		client:           client,
		toolExecutor:     tools.NewExecutorWithDynamicTools(getDbTools),
		streamingHandler: streaming.NewStreamingHandler(),
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
		}
	}

	// Execute the tool normally
	return c.toolExecutor.Execute(toolCall)
}

// QueryWithTools performs a query with tools support (non-streaming)
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
		// Process all tool calls
		toolMessages := []openai.ChatCompletionMessage{message}

		for _, toolCall := range message.ToolCalls {
			result, err := c.toolExecutor.Execute(toolCall)
			if err != nil {
				return "", fmt.Errorf("tool execution error: %w", err)
			}

			toolMessages = append(toolMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}

		updatedMessages := append(messages, toolMessages...)
		return c.QueryWithTools(ctx, updatedMessages)
	}

	return message.Content, nil
}

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

	// Process the stream
	completeMessage, err := c.streamingHandler.ProcessStream(ctx, stream, streaming.StreamingCallback(callback))
	if err != nil {
		return err
	}

	// If tools are needed, execute them
	if len(completeMessage.ToolCalls) > 0 {
		// For streaming with multiple tool calls, we need to handle them sequentially
		// because confirmation might be needed for some tools
		toolCall := completeMessage.ToolCalls[0]

		// Execute the first tool with confirmation check
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

		// Process all tool calls
		toolMessages := []openai.ChatCompletionMessage{completeMessage}

		for _, tc := range completeMessage.ToolCalls {
			var toolResult string
			if tc.ID == toolCall.ID {
				// Use the already executed result for the first tool
				toolResult = result
			} else {
				// Execute other tools (they don't need confirmation in this context)
				toolResult, err = c.toolExecutor.Execute(tc)
				if err != nil {
					return fmt.Errorf("tool execution error: %w", err)
				}
			}

			toolMessages = append(toolMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    toolResult,
				ToolCallID: tc.ID,
			})
		}

		updatedMessages := append(messages, toolMessages...)
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

// ContinueWithConfirmedTool continues AI processing after tool confirmation
func (c *Client) ContinueWithConfirmedTool(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, callback StreamingCallback) error {
	// Execute the confirmed tool
	result, err := c.toolExecutor.Execute(toolCall)
	if err != nil {
		return fmt.Errorf("tool execution error: %w", err)
	}

	// Process all tool calls
	toolMessages := []openai.ChatCompletionMessage{completeMessage}

	for _, tc := range completeMessage.ToolCalls {
		var toolResult string
		if tc.ID == toolCall.ID {
			// Use the confirmed tool result
			toolResult = result
		} else {
			// Execute other tools
			toolResult, err = c.toolExecutor.Execute(tc)
			if err != nil {
				return fmt.Errorf("tool execution error: %w", err)
			}
		}

		toolMessages = append(toolMessages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    toolResult,
			ToolCallID: tc.ID,
		})
	}

	updatedMessages := append(messages, toolMessages...)
	// Continue with streaming
	return c.QueryWithToolsStreaming(ctx, updatedMessages, callback)
}
