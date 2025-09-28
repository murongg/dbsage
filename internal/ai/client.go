package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"dbsage/internal/ai/streaming"
	"dbsage/internal/ai/tools"
	"dbsage/pkg/database"

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

func NewClient(apiKey, baseURL string, dbTools *database.DatabaseTools) *Client {
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
func NewClientWithDynamicTools(apiKey, baseURL string, getDbTools func() *database.DatabaseTools) *Client {
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
		toolCall := message.ToolCalls[0]

		result, err := c.toolExecutor.Execute(toolCall)
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
		toolCall := completeMessage.ToolCalls[0]
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

// ContinueWithConfirmedTool continues AI processing after tool confirmation
func (c *Client) ContinueWithConfirmedTool(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, callback StreamingCallback) error {
	// Execute the confirmed tool
	result, err := c.toolExecutor.Execute(toolCall)
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
