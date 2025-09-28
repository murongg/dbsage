package streaming

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// StreamingCallback is called for each chunk of streaming response
// This is an alias to avoid circular imports
type StreamingCallback = func(chunk string) error

// StreamingHandler handles streaming responses from OpenAI
type StreamingHandler struct{}

func NewStreamingHandler() *StreamingHandler {
	return &StreamingHandler{}
}

// ProcessStream handles the streaming response and accumulates tool calls
func (h *StreamingHandler) ProcessStream(
	ctx context.Context,
	stream *openai.ChatCompletionStream,
	callback StreamingCallback,
) (openai.ChatCompletionMessage, error) {
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
			return completeMessage, fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta

			// Accumulate content
			if delta.Content != "" {
				contentBuilder.WriteString(delta.Content)
				// Send content chunks to callback for streaming display
				if err := callback(delta.Content); err != nil {
					return completeMessage, err
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

	return completeMessage, nil
}
