package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"dbsage/internal/ai"
	"dbsage/internal/models"

	"github.com/sashabaranov/go-openai"
)

// ToolHandler handles tool confirmation and execution
type ToolHandler struct{}

func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// CheckToolConfirmation checks if a tool requires confirmation
func (h *ToolHandler) CheckToolConfirmation(toolName string, config *models.ToolConfirmationConfig) bool {
	if config == nil {
		return false
	}
	return config.RequiresConfirmation[toolName]
}

// CreateToolConfirmationInfo creates tool confirmation information
func (h *ToolHandler) CreateToolConfirmationInfo(toolName, toolCallID string, args map[string]interface{}, config *models.ToolConfirmationConfig) *models.ToolConfirmationInfo {
	if config == nil {
		return nil
	}

	description := config.Descriptions[toolName]
	riskLevel := config.RiskLevels[toolName]

	// Create description based on tool and arguments
	if description == "" {
		switch toolName {
		case "execute_sql":
			if sql, ok := args["sql"].(string); ok {
				description = fmt.Sprintf("Execute SQL: %s", sql)
			}
		default:
			description = fmt.Sprintf("Execute %s", toolName)
		}
	}

	if riskLevel == "" {
		riskLevel = "medium"
	}

	// Create confirmation options
	options := []models.ConfirmationOption{
		{
			Key:         "1",
			Label:       "Execute",
			Description: "Execute the operation",
			Action:      "execute",
		},
		{
			Key:         "2",
			Label:       "Cancel",
			Description: "Cancel the operation",
			Action:      "cancel",
		},
	}

	return &models.ToolConfirmationInfo{
		ToolName:    toolName,
		ToolCallID:  toolCallID,
		Arguments:   args,
		Description: description,
		RiskLevel:   riskLevel,
		Options:     options,
	}
}

// HandleToolConfirmation handles tool confirmation requests
func (h *ToolHandler) HandleToolConfirmation(
	ctx context.Context,
	messages []openai.ChatCompletionMessage,
	completeMessage openai.ChatCompletionMessage,
	toolCall openai.ToolCall,
	streamingCallback ai.StreamingCallback,
	config *models.ToolConfirmationConfig,
) (*models.ToolConfirmationInfo, *models.PendingAIContext, error) {
	// Check if this tool requires confirmation
	if !h.CheckToolConfirmation(toolCall.Function.Name, config) {
		return nil, nil, nil // No confirmation needed
	}

	// Parse arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return nil, nil, fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	// Create tool confirmation info
	toolInfo := h.CreateToolConfirmationInfo(toolCall.Function.Name, toolCall.ID, args, config)
	if toolInfo == nil {
		return nil, nil, nil // Fallback to allowing execution
	}

	// Store the AI context for later resumption
	aiContext := &models.PendingAIContext{
		Messages:          messages,
		CompleteMessage:   completeMessage,
		ToolCall:          toolCall,
		StreamingCallback: streamingCallback,
	}

	return toolInfo, aiContext, nil
}
