package components

import (
	"github.com/charmbracelet/bubbles/textinput"
)

// CreateTextInput creates a new text input component with default settings
func CreateTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = "> "
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 70
	return ti
}

// UpdateTextInputWidth updates the text input width based on window dimensions
func UpdateTextInputWidth(ti *textinput.Model, windowWidth int) {
	ti.Width = windowWidth - 12 // Reserve space for borders and labels
	if ti.Width < 20 {
		ti.Width = 20
	}
}
