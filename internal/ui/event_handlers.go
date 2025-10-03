package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dbsage/internal/ai"
	"dbsage/internal/models"
	"dbsage/internal/ui/components"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"
)

// handleWindowSize handles window resize events
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	// Update renderer dimensions
	m.layoutRenderer.SetDimensions(m.width, m.height)
	m.contentRenderer.SetWidth(m.width)
	m.commandRenderer.SetWidth(m.width)

	// Update text input width
	components.UpdateTextInputWidth(&m.textInput, m.width)

	// Update confirmation list dimensions
	m.confirmationList.SetWidth(m.width - 4)
	m.confirmationList.SetHeight(10)

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		if m.stateManager.GetState() == models.StateToolConfirmation {
			return m, func() tea.Msg {
				return models.ToolConfirmationResponseMsg{Confirmed: false, Action: "cancel"}
			}
		}
		return m, tea.Quit

	case "ctrl+h", "?":
		showHelp := !m.stateManager.IsShowHelp()
		m.stateManager.SetShowHelp(showHelp)
		return m, nil

	case "q":
		// Dismiss guidance if it's currently shown
		if m.stateManager.GetCurrentGuidance() != nil {
			m.stateManager.DismissGuidance()
			return m, nil
		}
		// Dismiss version update notification if it's currently shown
		if m.stateManager.GetVersionUpdate() != nil {
			m.stateManager.DismissVersionUpdate()
			return m, nil
		}
		return m, nil

	case "tab":
		if m.stateManager.GetState() == models.StateInput {
			return m.handleTabCompletion()
		}
		return m, nil

	case "up":
		if m.stateManager.IsShowSuggestions() {
			m.stateManager.MoveSuggestionUp()
			return m, nil
		} else if m.stateManager.GetState() == models.StateToolConfirmation {
			var cmd tea.Cmd
			m.confirmationList, cmd = m.confirmationList.Update(msg)
			return m, cmd
		}
		return m, nil

	case "down":
		if m.stateManager.IsShowSuggestions() {
			m.stateManager.MoveSuggestionDown()
			return m, nil
		} else if m.stateManager.GetState() == models.StateToolConfirmation {
			var cmd tea.Cmd
			m.confirmationList, cmd = m.confirmationList.Update(msg)
			return m, cmd
		}
		return m, nil

	case "enter":
		if m.stateManager.IsShowSuggestions() {
			// Select the highlighted suggestion
			return m.handleSuggestionSelection()
		} else if m.stateManager.GetState() == models.StateInput && strings.TrimSpace(m.textInput.Value()) != "" {
			return m.handleInput()
		} else if m.stateManager.GetState() == models.StateToolConfirmation {
			if selectedItem, ok := m.confirmationList.SelectedItem().(components.ConfirmationItem); ok {
				confirmed := selectedItem.Title() == "Execute"
				action := "execute"
				if !confirmed {
					action = "cancel"
				}
				return m, func() tea.Msg {
					return models.ToolConfirmationResponseMsg{Confirmed: confirmed, Action: action}
				}
			}
		}
		return m, nil

	default:
		if m.stateManager.GetState() == models.StateToolConfirmation {
			var cmd tea.Cmd
			m.confirmationList, cmd = m.confirmationList.Update(msg)
			return m, cmd
		} else if m.stateManager.GetState() == models.StateInput || m.stateManager.GetState() == models.StateResponse {
			if m.stateManager.GetState() == models.StateResponse {
				m.stateManager.SetState(models.StateInput)
				m.stateManager.SetResponse("")
				m.stateManager.SetError(nil)
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)

			// Update command suggestions based on current input
			currentInput := m.textInput.Value()
			m.stateManager.UpdateCommandSuggestions(currentInput)

			return m, cmd
		}
		return m, nil
	}
}

// handleTabCompletion handles tab completion
func (m *Model) handleTabCompletion() (tea.Model, tea.Cmd) {
	if m.stateManager.IsShowSuggestions() {
		// If suggestions are showing, select the highlighted one
		return m.handleSuggestionSelection()
	}

	suggestions := m.stateManager.GetCommandSuggestions()
	m.inputHandler.HandleTabCompletion(&m.textInput, suggestions)

	// Clear suggestions after completion
	m.stateManager.SetShowSuggestions(false)
	m.stateManager.SetCommandSuggestions(nil)
	m.stateManager.SetShowParameterHelp(false)
	m.stateManager.SetParameterHelp("")

	return m, nil
}

// handleSuggestionSelection handles selecting a suggestion from the list
func (m *Model) handleSuggestionSelection() (tea.Model, tea.Cmd) {
	selectedCommand := m.stateManager.GetSelectedSuggestionCommand()
	if selectedCommand != "" {
		// Set the selected command as input value
		m.textInput.SetValue(selectedCommand + " ")
		m.textInput.SetCursor(len(selectedCommand) + 1)

		// Clear suggestions
		m.stateManager.SetShowSuggestions(false)
		m.stateManager.SetCommandSuggestions(nil)
		m.stateManager.SetShowParameterHelp(false)
		m.stateManager.SetParameterHelp("")
	}

	return m, nil
}

// handleInput handles user input submission
func (m *Model) handleInput() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.textInput.Value())

	// Clear suggestions
	m.stateManager.SetShowSuggestions(false)
	m.stateManager.SetCommandSuggestions(nil)
	m.stateManager.SetShowParameterHelp(false)
	m.stateManager.SetParameterHelp("")

	// Process input through state manager (handles commands)
	shouldContinue, _ := m.stateManager.ProcessInput(input)
	if !shouldContinue {
		// Exit was requested
		return m, tea.Quit
	}

	// Check if it was handled as a command
	if m.stateManager.GetState() == models.StateResponse {
		// Command was handled, reset input and ensure focus
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, func() tea.Msg { return models.CommandCompletedMsg{} }
	}

	// Not a command, process as AI query
	// Add user message to history
	m.stateManager.AddToHistory(openai.ChatMessageRoleUser, input)

	// Send thinking state message and start AI query
	return m, tea.Batch(
		func() tea.Msg { return models.AIThinkingMsg{} },
		m.queryAI(),
	)
}

// handleAIThinking handles AI thinking state
func (m *Model) handleAIThinking() (tea.Model, tea.Cmd) {
	m.stateManager.SetState(models.StateThinking)
	m.textInput.Blur()
	return m, m.tick()
}

// handleAIResponse handles AI response
func (m *Model) handleAIResponse(msg models.AIResponseMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.stateManager.SetError(msg.Err)
		m.stateManager.SetState(models.StateResponse)
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink
	} else {
		m.stateManager.SetResponse(msg.Response)
		m.stateManager.SetError(nil)
		m.stateManager.SetState(models.StateInput)
		m.stateManager.AddToHistory(openai.ChatMessageRoleAssistant, msg.Response)
	}

	m.textInput.SetValue("")
	m.textInput.Focus()
	return m, textinput.Blink
}

// handleCommandCompleted handles command completion
func (m *Model) handleCommandCompleted() (tea.Model, tea.Cmd) {
	m.stateManager.SetShowSuggestions(false)
	m.stateManager.SetCommandSuggestions(nil)
	m.stateManager.SetShowParameterHelp(false)
	m.stateManager.SetParameterHelp("")
	m.textInput.Focus()
	return m, textinput.Blink
}

// handleTick handles animation ticks
func (m *Model) handleTick() (tea.Model, tea.Cmd) {
	if m.stateManager.GetState() == models.StateThinking {
		return m, m.tick()
	}
	return m, nil
}

// handleStreamChunk handles streaming response chunks
func (m *Model) handleStreamChunk(msg models.AIStreamChunkMsg) (tea.Model, tea.Cmd) {
	m.streamingResponse += msg.Chunk

	if m.stateManager.GetState() != models.StateToolConfirmation {
		m.stateManager.SetState(models.StateResponse)
		m.stateManager.SetResponse(m.streamingResponse)
		m.stateManager.SetError(nil)
	}
	return m, nil
}

// handleStreamComplete handles streaming completion
func (m *Model) handleStreamComplete(msg models.AIStreamCompleteMsg) (tea.Model, tea.Cmd) {
	m.stateManager.AddToHistory(openai.ChatMessageRoleAssistant, msg.FullResponse)

	if m.stateManager.GetState() != models.StateToolConfirmation {
		m.stateManager.SetState(models.StateInput)
		m.textInput.SetValue("")
		m.textInput.Focus()
	}

	m.streamingResponse = ""
	return m, textinput.Blink
}

// handleToolConfirmation handles tool confirmation requests
func (m *Model) handleToolConfirmation(msg models.ToolConfirmationMsg) (tea.Model, tea.Cmd) {
	m.stateManager.SetPendingToolConfirmation(msg.ToolInfo)
	m.stateManager.SetState(models.StateToolConfirmation)
	m.textInput.Blur()

	components.SetupConfirmationList(&m.confirmationList, msg.ToolInfo)

	return m, nil
}

// handleToolConfirmationResponse handles tool confirmation responses
func (m *Model) handleToolConfirmationResponse(msg models.ToolConfirmationResponseMsg) (tea.Model, tea.Cmd) {
	if msg.Confirmed && msg.Action == "execute" {
		return m.executePendingTool()
	} else {
		toolInfo := m.stateManager.GetPendingToolConfirmation()
		m.stateManager.ClearPendingToolConfirmation()
		m.stateManager.SetState(models.StateInput)
		m.textInput.Focus()

		if msg.Action == "edit" && toolInfo != nil {
			if toolInfo.ToolName == "execute_sql" {
				if sql, ok := toolInfo.Arguments["sql"].(string); ok {
					m.textInput.SetValue("Execute this SQL: " + sql)
				}
			} else {
				m.textInput.SetValue("Execute " + toolInfo.Description)
			}
		}
		return m, textinput.Blink
	}
}

// tick generates animation tick commands
func (m *Model) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return models.TickMsg(t)
	})
}

// queryAI queries AI with streaming support
func (m *Model) queryAI() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		history := m.stateManager.GetHistory()
		aiClient := m.stateManager.GetAIClient()

		go func() {
			var fullResponse strings.Builder

			err := aiClient.QueryWithToolsStreaming(ctx, history, func(chunk string) error {
				fullResponse.WriteString(chunk)

				if m.program != nil {
					m.program.Send(models.AIStreamChunkMsg{Chunk: chunk})
				}
				return nil
			})

			if m.program != nil {
				if err != nil {
					m.program.Send(models.AIResponseMsg{Response: "", Err: err})
				} else {
					m.program.Send(models.AIStreamCompleteMsg{FullResponse: fullResponse.String()})
				}
			}
		}()

		return nil
	}
}

// executePendingTool executes the pending tool
func (m *Model) executePendingTool() (tea.Model, tea.Cmd) {
	aiContext := m.stateManager.GetPendingAIContext()
	if aiContext == nil {
		m.stateManager.SetError(fmt.Errorf("no pending AI context to resume"))
		m.stateManager.SetState(models.StateResponse)
		m.textInput.Focus()
		return m, textinput.Blink
	}

	aiClient := m.stateManager.GetAIClient()
	if aiClient == nil {
		m.stateManager.SetError(fmt.Errorf("no AI client available"))
		m.stateManager.SetState(models.StateResponse)
		m.stateManager.ClearPendingToolConfirmation()
		m.textInput.Focus()
		return m, textinput.Blink
	}

	m.stateManager.ClearPendingToolConfirmation()
	m.stateManager.SetState(models.StateThinking)

	return m, func() tea.Msg {
		ctx := context.Background()
		var fullResponse strings.Builder

		streamingCallback := func(chunk string) error {
			if m.program != nil {
				m.program.Send(models.AIStreamChunkMsg{Chunk: chunk})
			}
			fullResponse.WriteString(chunk)
			return nil
		}

		err := aiClient.ContinueWithConfirmedTool(
			ctx,
			aiContext.Messages,
			aiContext.CompleteMessage,
			aiContext.ToolCall,
			streamingCallback,
		)

		if m.program != nil {
			if err != nil {
				m.program.Send(models.AIResponseMsg{Response: "", Err: err})
			} else {
				m.program.Send(models.AIStreamCompleteMsg{FullResponse: fullResponse.String()})
			}
		}

		return nil
	}
}

// handleToolConfirmationFromAI handles tool confirmation requests from the AI client
func (m *Model) handleToolConfirmationFromAI(ctx context.Context, messages []openai.ChatCompletionMessage, completeMessage openai.ChatCompletionMessage, toolCall openai.ToolCall, streamingCallback ai.StreamingCallback) (bool, error) {
	if !m.stateManager.RequiresConfirmation(toolCall.Function.Name) {
		return true, nil
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return false, fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	toolInfo := m.stateManager.CreateToolConfirmationInfo(toolCall.Function.Name, toolCall.ID, args)
	if toolInfo == nil {
		return true, nil
	}

	aiContext := &models.PendingAIContext{
		Messages:          messages,
		CompleteMessage:   completeMessage,
		ToolCall:          toolCall,
		StreamingCallback: streamingCallback,
	}
	m.stateManager.SetPendingAIContext(aiContext)

	if m.program != nil {
		m.program.Send(models.ToolConfirmationMsg{ToolInfo: toolInfo})
		return false, nil
	}

	return true, nil
}

// handleVersionUpdate handles version update notifications
func (m *Model) handleVersionUpdate(msg models.VersionUpdateMsg) (tea.Model, tea.Cmd) {
	m.stateManager.SetVersionUpdate(msg.UpdateInfo)
	return m, nil
}
