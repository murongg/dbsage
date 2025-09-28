package handlers

import (
	"strings"

	"dbsage/internal/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// InputHandler handles input-related functionality
type InputHandler struct{}

func NewInputHandler() *InputHandler {
	return &InputHandler{}
}

// HandleTabCompletion handles tab key for command completion
func (h *InputHandler) HandleTabCompletion(textInput *textinput.Model, suggestions []*models.CommandInfo) {
	currentInput := strings.TrimSpace(textInput.Value())

	// Complete if input starts with "/" or "@"
	if !strings.HasPrefix(currentInput, "/") && !strings.HasPrefix(currentInput, "@") {
		return
	}

	if len(suggestions) == 1 {
		// If there's only one suggestion, complete it
		completed := suggestions[0].Name
		if !strings.HasSuffix(completed, " ") {
			completed += " "
		}
		textInput.SetValue(completed)
	} else if len(suggestions) > 1 {
		// Find common prefix for partial completion
		commonPrefix := h.findCommonPrefix(suggestions)
		if len(commonPrefix) > len(currentInput) {
			textInput.SetValue(commonPrefix)
		}
	}
}

// findCommonPrefix finds the common prefix among command suggestions
func (h *InputHandler) findCommonPrefix(suggestions []*models.CommandInfo) string {
	if len(suggestions) == 0 {
		return ""
	}

	prefix := suggestions[0].Name
	for _, cmd := range suggestions[1:] {
		// Find common prefix between current prefix and this command
		for i := 0; i < len(prefix) && i < len(cmd.Name); i++ {
			if prefix[i] != cmd.Name[i] {
				prefix = prefix[:i]
				break
			}
		}
		if len(prefix) == 0 {
			break
		}
	}

	return prefix
}

// ProcessInput processes user input and determines the next action
func (h *InputHandler) ProcessInput(input string) (bool, tea.Cmd) {
	input = strings.TrimSpace(input)

	// Check for exit commands
	if input == "exit" || input == "quit" {
		return false, tea.Quit
	}

	// Return true to continue processing
	return true, nil
}
