package ui

import (
	"strings"
)

// TableRenderer handles table-specific rendering and optimization
type TableRenderer struct {
	maxWidth int
}

// NewTableRenderer creates a new table renderer
func NewTableRenderer(maxWidth int) *TableRenderer {
	return &TableRenderer{
		maxWidth: maxWidth,
	}
}

// OptimizeTablesForDisplay optimizes markdown tables for better terminal display
func (tr *TableRenderer) OptimizeTablesForDisplay(content string) string {
	lines := strings.Split(content, "\n")
	var optimizedLines []string
	inTable := false

	for _, line := range lines {
		// Detect table lines (contain | characters)
		if strings.Contains(line, "|") && strings.Count(line, "|") >= 2 {
			if !inTable {
				inTable = true
			}

			// Process table row
			optimizedLine := tr.optimizeTableRow(line)
			optimizedLines = append(optimizedLines, optimizedLine)

		} else if inTable && strings.TrimSpace(line) == "" {
			// End of table
			inTable = false
			optimizedLines = append(optimizedLines, line)
		} else if inTable && strings.Contains(line, "---") {
			// Table separator line, keep as is but optimize length
			optimizedLines = append(optimizedLines, line)
		} else {
			// Regular line or not in table
			inTable = false
			optimizedLines = append(optimizedLines, line)
		}
	}

	return strings.Join(optimizedLines, "\n")
}

// optimizeTableRow optimizes a single table row by truncating long content
func (tr *TableRenderer) optimizeTableRow(row string) string {
	// Split by | to get cells
	cells := strings.Split(row, "|")

	// Calculate available width per cell
	numCells := len(cells)
	if numCells <= 2 {
		return row // Not a proper table row
	}

	// Get content cells (excluding empty cells at start/end)
	var contentCells []string
	for _, cell := range cells {
		trimmed := strings.TrimSpace(cell)
		if trimmed != "" {
			contentCells = append(contentCells, trimmed)
		}
	}

	if len(contentCells) == 0 {
		return row
	}

	// Calculate dynamic cell widths based on content
	cellWidths := tr.calculateOptimalCellWidths(contentCells)

	// Process each cell
	var optimizedCells []string
	contentIndex := 0

	for _, cell := range cells {
		trimmedCell := strings.TrimSpace(cell)

		// Skip empty cells (usually first and last in markdown tables)
		if trimmedCell == "" {
			optimizedCells = append(optimizedCells, cell)
			continue
		}

		// Get the calculated width for this content cell
		if contentIndex < len(cellWidths) {
			maxCellWidth := cellWidths[contentIndex]
			contentIndex++

			// Truncate if necessary
			if len(trimmedCell) > maxCellWidth {
				truncated := TruncateText(trimmedCell, maxCellWidth)
				optimizedCells = append(optimizedCells, " "+truncated+" ")
			} else {
				optimizedCells = append(optimizedCells, cell)
			}
		} else {
			optimizedCells = append(optimizedCells, cell)
		}
	}

	return strings.Join(optimizedCells, "|")
}

// calculateOptimalCellWidths calculates optimal width for each cell based on content
func (tr *TableRenderer) calculateOptimalCellWidths(cells []string) []int {
	numCells := len(cells)
	if numCells == 0 {
		return []int{}
	}

	// Reserve space for separators and padding
	separatorWidth := (numCells + 1) * 3 // " | " between cells and at edges
	availableWidth := tr.maxWidth - separatorWidth

	if availableWidth <= 0 {
		availableWidth = numCells * 8 // Minimum width
	}

	// Calculate natural width for each cell (up to a reasonable limit)
	naturalWidths := make([]int, numCells)
	totalNaturalWidth := 0

	for i, cell := range cells {
		naturalWidth := len(cell)
		if naturalWidth > 40 { // Cap natural width
			naturalWidth = 40
		}
		if naturalWidth < 8 { // Minimum width
			naturalWidth = 8
		}
		naturalWidths[i] = naturalWidth
		totalNaturalWidth += naturalWidth
	}

	// If natural widths fit, use them
	if totalNaturalWidth <= availableWidth {
		return naturalWidths
	}

	// Otherwise, scale down proportionally
	widths := make([]int, numCells)
	for i, naturalWidth := range naturalWidths {
		scaledWidth := (naturalWidth * availableWidth) / totalNaturalWidth
		if scaledWidth < 8 {
			scaledWidth = 8
		}
		widths[i] = scaledWidth
	}

	return widths
}
