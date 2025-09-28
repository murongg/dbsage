package models

import (
	"time"

	"github.com/sashabaranov/go-openai"
)

// AppState represents the application state
type AppState int

const (
	StateWelcome AppState = iota
	StateInput
	StateThinking
	StateResponse
	StateHelp
	StateToolConfirmation
)

// ToolConfirmationInfo contains information about a tool that needs confirmation
type ToolConfirmationInfo struct {
	ToolName    string                 `json:"tool_name"`
	ToolCallID  string                 `json:"tool_call_id"`
	Arguments   map[string]interface{} `json:"arguments"`
	Description string                 `json:"description"`
	RiskLevel   string                 `json:"risk_level"` // "low", "medium", "high"
	Options     []ConfirmationOption   `json:"options"`
}

// ConfirmationOption represents an option in the confirmation dialog
type ConfirmationOption struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Action      string `json:"action"` // "execute", "cancel", "edit", "custom"
}

// ToolConfirmationConfig defines which tools need confirmation
type ToolConfirmationConfig struct {
	RequiresConfirmation map[string]bool   `json:"requires_confirmation"`
	RiskLevels           map[string]string `json:"risk_levels"`
	Descriptions         map[string]string `json:"descriptions"`
}

// PendingAIContext stores the context needed to resume AI processing after confirmation
type PendingAIContext struct {
	Messages          []openai.ChatCompletionMessage `json:"messages"`
	CompleteMessage   openai.ChatCompletionMessage   `json:"complete_message"`
	ToolCall          openai.ToolCall                `json:"tool_call"`
	StreamingCallback func(chunk string) error       `json:"-"` // Not serializable
}

// CommandInfo contains information about a command
type CommandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Category    string `json:"category"`
}

// Message types for Bubble Tea
type AIResponseMsg struct {
	Response string
	Err      error
}

type AIThinkingMsg struct{}

type CommandCompletedMsg struct{}

type TickMsg time.Time

type AIStreamChunkMsg struct {
	Chunk string
}

type AIStreamCompleteMsg struct {
	FullResponse string
}

type ToolConfirmationMsg struct {
	ToolInfo *ToolConfirmationInfo
}

type ToolConfirmationResponseMsg struct {
	Confirmed bool
	Action    string // "execute", "cancel", "edit"
}
