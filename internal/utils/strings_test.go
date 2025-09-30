package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "string shorter than maxLength",
			input:     "hello",
			maxLength: 10,
			expected:  "hello",
		},
		{
			name:      "string equal to maxLength",
			input:     "hello",
			maxLength: 5,
			expected:  "hello",
		},
		{
			name:      "string longer than maxLength",
			input:     "hello world",
			maxLength: 8,
			expected:  "hello...",
		},
		{
			name:      "empty string",
			input:     "",
			maxLength: 5,
			expected:  "",
		},
		{
			name:      "maxLength of 3 (minimum for truncation)",
			input:     "hello",
			maxLength: 3,
			expected:  "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLength)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple spaces",
			input:    "hello    world",
			expected: "hello world",
		},
		{
			name:     "tabs and spaces",
			input:    "hello\t\t  world",
			expected: "hello world",
		},
		{
			name:     "leading and trailing whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "newlines and mixed whitespace",
			input:    "hello\n\t world\n",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name:     "single word",
			input:    "hello",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanWhitespace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: true,
		},
		{
			name:     "only tabs",
			input:    "\t\t",
			expected: true,
		},
		{
			name:     "only newlines",
			input:    "\n\n",
			expected: true,
		},
		{
			name:     "mixed whitespace",
			input:    " \t\n ",
			expected: true,
		},
		{
			name:     "non-empty string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "string with content and whitespace",
			input:    " hello ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase word",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "uppercase word",
			input:    "HELLO",
			expected: "Hello",
		},
		{
			name:     "mixed case word",
			input:    "hELLo",
			expected: "Hello",
		},
		{
			name:     "multiple words",
			input:    "hello world",
			expected: "Hello World",
		},
		{
			name:     "words with underscores",
			input:    "hello_world",
			expected: "Hello_world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid identifiers
		{
			name:     "simple letter",
			input:    "a",
			expected: true,
		},
		{
			name:     "underscore only",
			input:    "_",
			expected: true,
		},
		{
			name:     "letter and numbers",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "underscore and letters",
			input:    "_hello",
			expected: true,
		},
		{
			name:     "mixed valid characters",
			input:    "hello_world_123",
			expected: true,
		},
		{
			name:     "uppercase letters",
			input:    "HELLO",
			expected: true,
		},
		{
			name:     "mixed case",
			input:    "HelloWorld",
			expected: true,
		},
		// Invalid identifiers
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "starts with number",
			input:    "123abc",
			expected: false,
		},
		{
			name:     "contains space",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "contains special characters",
			input:    "hello@world",
			expected: false,
		},
		{
			name:     "contains hyphen",
			input:    "hello-world",
			expected: false,
		},
		{
			name:     "contains dot",
			input:    "hello.world",
			expected: false,
		},
		{
			name:     "only numbers",
			input:    "123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkTruncateString(b *testing.B) {
	longString := "This is a very long string that will be truncated in the benchmark test"
	for i := 0; i < b.N; i++ {
		TruncateString(longString, 20)
	}
}

func BenchmarkCleanWhitespace(b *testing.B) {
	dirtyString := "  hello    world  \t\n  with   lots   of   whitespace  "
	for i := 0; i < b.N; i++ {
		CleanWhitespace(dirtyString)
	}
}

func BenchmarkIsValidIdentifier(b *testing.B) {
	identifier := "valid_identifier_123"
	for i := 0; i < b.N; i++ {
		IsValidIdentifier(identifier)
	}
}
