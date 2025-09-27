package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// MarkdownRenderer handles markdown rendering with table optimization
type MarkdownRenderer struct {
	width int
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(width int) *MarkdownRenderer {
	return &MarkdownRenderer{
		width: width,
	}
}

// SetWidth updates the renderer width
func (mr *MarkdownRenderer) SetWidth(width int) {
	mr.width = width
}

// RenderMarkdown renders markdown content with enhanced styling
func (mr *MarkdownRenderer) RenderMarkdown(content string) string {
	// Calculate width
	maxWidth := mr.width - 4 // Leave some margin
	if maxWidth < 40 {
		maxWidth = 40
	}

	// Pre-process content to optimize tables
	tableRenderer := NewTableRenderer(maxWidth)
	content = tableRenderer.OptimizeTablesForDisplay(content)

	// Create a new renderer each time to avoid state issues
	// Use a minimal, safe configuration
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("notty"), // Use notty style to avoid terminal control sequences
		glamour.WithWordWrap(maxWidth),
	)
	if err != nil {
		// If glamour fails, return empty to fallback to plain text
		return ""
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		// If rendering fails, return empty to fallback to plain text
		return ""
	}

	// Clean up the rendered content
	rendered = strings.TrimRight(rendered, "\n")

	// Additional safety: remove any remaining control sequences
	rendered = mr.cleanControlSequences(rendered)

	return rendered
}

// cleanControlSequences removes terminal control sequences that might leak
func (mr *MarkdownRenderer) cleanControlSequences(text string) string {
	// Remove ANSI escape sequences that might cause issues
	// This is a simple cleanup - removes sequences like \x1b[...m or ]11;...
	lines := strings.Split(text, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Remove common problematic sequences
		line = strings.ReplaceAll(line, "\x1b", "") // Remove escape character
		line = strings.ReplaceAll(line, "\033", "") // Remove octal escape
		// Remove sequences that start with ] and contain rgb/color info
		if strings.Contains(line, "]11;rgb:") || strings.Contains(line, "]11;") {
			// Skip lines that contain terminal color setting sequences
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, "\n")
}
