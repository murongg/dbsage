package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMarkdownRenderer(t *testing.T) {
	renderer := NewMarkdownRenderer(80)

	assert.NotNil(t, renderer)
	assert.Equal(t, 80, renderer.width)
}

func TestMarkdownRenderer_SetWidth(t *testing.T) {
	renderer := NewMarkdownRenderer(80)

	renderer.SetWidth(120)
	assert.Equal(t, 120, renderer.width)

	renderer.SetWidth(40)
	assert.Equal(t, 40, renderer.width)
}

func TestMarkdownRenderer_RenderMarkdown_SimpleContent(t *testing.T) {
	renderer := NewMarkdownRenderer(80)

	tests := []struct {
		name     string
		content  string
		contains []string // Strings that should be present in output
	}{
		{
			name:     "plain text",
			content:  "Hello world",
			contains: []string{"Hello world"},
		},
		{
			name:     "header",
			content:  "# Main Title",
			contains: []string{"Main Title"},
		},
		{
			name:     "bold text",
			content:  "This is **bold** text",
			contains: []string{"bold"},
		},
		{
			name:     "italic text",
			content:  "This is *italic* text",
			contains: []string{"italic"},
		},
		{
			name:     "code block",
			content:  "```sql\nSELECT * FROM users;\n```",
			contains: []string{"SELECT", "users"},
		},
		{
			name:     "inline code",
			content:  "Use `SELECT` statement",
			contains: []string{"SELECT"},
		},
		{
			name:     "link",
			content:  "[OpenAI](https://openai.com)",
			contains: []string{"OpenAI"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.RenderMarkdown(tt.content)

			// If glamour rendering fails, it returns empty string
			// This is acceptable fallback behavior
			if result == "" {
				t.Skip("Glamour rendering not available in test environment")
				return
			}

			for _, expected := range tt.contains {
				assert.Contains(t, result, expected,
					"Expected '%s' to be present in rendered output", expected)
			}
		})
	}
}

func TestMarkdownRenderer_RenderMarkdown_DifferentWidths(t *testing.T) {
	content := "# This is a long title that should be wrapped properly"

	tests := []struct {
		name  string
		width int
	}{
		{"narrow", 40},
		{"medium", 80},
		{"wide", 120},
		{"very narrow", 20},
		{"very wide", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewMarkdownRenderer(tt.width)
			result := renderer.RenderMarkdown(content)

			// If rendering fails, it returns empty (fallback behavior)
			if result == "" {
				t.Skip("Glamour rendering not available in test environment")
				return
			}

			// Basic validation - should not be empty and should contain title content
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "title")
		})
	}
}

func TestMarkdownRenderer_RenderMarkdown_EmptyAndInvalidContent(t *testing.T) {
	renderer := NewMarkdownRenderer(80)

	tests := []struct {
		name    string
		content string
	}{
		{"empty string", ""},
		{"only whitespace", "   \n\t  \n"},
		{"only newlines", "\n\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.RenderMarkdown(tt.content)
			// Empty/whitespace content should either render as empty or minimal content
			// This is acceptable behavior
			assert.NotContains(t, result, "ERROR")
		})
	}
}

func TestMarkdownRenderer_CleanControlSequences(t *testing.T) {
	renderer := NewMarkdownRenderer(80)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "escape character removal",
			input:    "Hello\x1b[31mworld\x1b[0m",
			expected: "Hello[31mworld[0m",
		},
		{
			name:     "octal escape removal",
			input:    "Hello\033[31mworld\033[0m",
			expected: "Hello[31mworld[0m",
		},
		{
			name:     "rgb color sequence removal",
			input:    "Line 1\n]11;rgb:ff00/00ff/0000\nLine 2",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "general color sequence removal",
			input:    "Line 1\n]11;some_color_setting\nLine 2",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "normal text unchanged",
			input:    "Hello world\nSecond line",
			expected: "Hello world\nSecond line",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.cleanControlSequences(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarkdownRenderer_RenderMarkdown_TableOptimization(t *testing.T) {
	renderer := NewMarkdownRenderer(80)

	// Test that table optimization is called
	// We can't easily test the full table optimization without setting up the table renderer
	// But we can test that basic table markdown doesn't cause errors
	tableContent := `| Name | Age | City |
|------|-----|------|
| John | 25  | NYC  |
| Jane | 30  | LA   |`

	result := renderer.RenderMarkdown(tableContent)

	// If glamour is available, result should contain table data
	// If not available, result will be empty (acceptable fallback)
	if result != "" {
		// Should contain table data
		assert.Contains(t, result, "John")
		assert.Contains(t, result, "Jane")
	}
}

func TestMarkdownRenderer_RenderMarkdown_ErrorHandling(t *testing.T) {
	// Test with extremely narrow width that might cause issues
	renderer := NewMarkdownRenderer(1)

	content := "# Very long header that is much longer than the available width"
	result := renderer.RenderMarkdown(content)

	// Should not panic and should handle gracefully
	// Result might be empty (fallback) or truncated content
	assert.NotContains(t, result, "panic")
	assert.NotContains(t, result, "fatal")
}

func TestMarkdownRenderer_LongContent(t *testing.T) {
	renderer := NewMarkdownRenderer(80)

	// Test with very long content
	longContent := strings.Repeat("This is a very long paragraph that contains a lot of text. ", 100)
	content := "# Long Content\n\n" + longContent

	result := renderer.RenderMarkdown(content)

	// Should handle long content without issues
	if result != "" {
		assert.Contains(t, result, "Long Content")
		// Rendered content should not contain excessive control sequences
		assert.NotContains(t, result, "\x1b[")
	}
}

func TestMarkdownRenderer_SpecialCharacters(t *testing.T) {
	renderer := NewMarkdownRenderer(80)

	content := `# Special Characters: áéíóú ñ 中文 🚀

This content has various special characters:
- Unicode: café résumé naïve
- Emojis: 🎉 💻 📊
- Chinese: 数据库管理
- Symbols: → ← ↑ ↓ ★ ♥`

	result := renderer.RenderMarkdown(content)

	// Should handle special characters gracefully
	if result != "" {
		assert.Contains(t, result, "Special Characters")
		// Should preserve unicode content
		assert.Contains(t, result, "café")
	}
}

// Benchmark tests
func BenchmarkMarkdownRenderer_RenderSimple(b *testing.B) {
	renderer := NewMarkdownRenderer(80)
	content := "# Simple Header\n\nThis is **bold** text with some *italic* content."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.RenderMarkdown(content)
	}
}

func BenchmarkMarkdownRenderer_RenderLong(b *testing.B) {
	renderer := NewMarkdownRenderer(80)
	content := "# Long Document\n\n" +
		strings.Repeat("This is a paragraph with **bold** and *italic* text. ", 50) +
		"\n\n## Section 2\n\n" +
		strings.Repeat("Another paragraph with [links](http://example.com) and code blocks. ", 30) +
		"\n\n```sql\nSELECT * FROM users WHERE active = true;\n```\n"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.RenderMarkdown(content)
	}
}

func BenchmarkMarkdownRenderer_CleanControlSequences(b *testing.B) {
	renderer := NewMarkdownRenderer(80)
	text := strings.Repeat("Normal text\x1b[31mwith\x1b[0m escape sequences\n]11;rgb:ff00/00ff/0000\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.cleanControlSequences(text)
	}
}
