package ui

import (
	"strings"
	"unicode/utf8"
)

// TruncateText truncates text intelligently using generic algorithms
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	if maxLength <= 3 {
		return "..."
	}

	// Strategy 1: Try to find natural break points (separators)
	if truncated := truncateAtSeparator(text, maxLength); truncated != "" {
		return truncated
	}

	// Strategy 2: Try to truncate at word boundary
	if truncated := truncateAtWordBoundary(text, maxLength); truncated != "" {
		return truncated
	}

	// Strategy 3: Simple truncation with ellipsis
	return text[:maxLength-3] + "..."
}

// truncateAtSeparator tries to truncate at natural separators (/, -, _, etc.)
func truncateAtSeparator(text string, maxLength int) string {
	separators := []string{"/", "-", "_", ".", ":", " "}

	for _, sep := range separators {
		parts := strings.Split(text, sep)
		if len(parts) <= 1 {
			continue
		}

		// Try to keep the most informative part (usually the last part)
		lastPart := parts[len(parts)-1]
		if len(lastPart)+3 <= maxLength {
			return "..." + lastPart
		}

		// Try to keep the first part
		firstPart := parts[0]
		if len(firstPart)+3 <= maxLength {
			return firstPart + "..."
		}

		// Try to combine first and last parts
		if len(firstPart)+len(lastPart)+6 <= maxLength {
			return firstPart + "..." + lastPart
		}
	}

	return ""
}

// truncateAtWordBoundary tries to truncate at word boundaries
func truncateAtWordBoundary(text string, maxLength int) string {
	if maxLength <= 3 {
		return ""
	}

	// Find the best position to truncate
	targetLength := maxLength - 3

	// If there are spaces, try to break at word boundary
	if strings.Contains(text, " ") {
		words := strings.Fields(text)
		if len(words) <= 1 {
			return ""
		}

		// Try to fit as many complete words as possible
		var result strings.Builder
		for _, word := range words {
			if result.Len()+len(word)+1 > targetLength {
				break
			}
			if result.Len() > 0 {
				result.WriteString(" ")
			}
			result.WriteString(word)
		}

		if result.Len() > 0 && result.Len() < len(text) {
			return result.String() + "..."
		}
	}

	return ""
}

// WrapText wraps text to fit within the specified width
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if utf8.RuneCountInString(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Wrap long lines
		words := strings.Fields(line)
		if len(words) == 0 {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		currentLine := ""
		for _, word := range words {
			// If adding this word would exceed the width
			if utf8.RuneCountInString(currentLine+" "+word) > width {
				if currentLine != "" {
					wrappedLines = append(wrappedLines, currentLine)
					currentLine = word
				} else {
					// Single word is too long, break it
					wrappedLines = append(wrappedLines, word)
				}
			} else {
				if currentLine == "" {
					currentLine = word
				} else {
					currentLine += " " + word
				}
			}
		}

		if currentLine != "" {
			wrappedLines = append(wrappedLines, currentLine)
		}
	}

	return strings.Join(wrappedLines, "\n")
}
