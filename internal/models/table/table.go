// Package table provides domain entities for PDF table extraction.
package table

import (
	"fmt"
	"strings"
)

// Table represents an extracted table with cell content.
//
// A table is a rich domain entity that encapsulates:
//   - Cell content organized in rows and columns
//   - Metadata (page number, bounds, extraction method)
//   - Behavior for accessing and manipulating cells
//
// Tables are aggregates in DDD - they are the root entity that manages
// a collection of cells.
//
// This represents the output of Phase 2.7 (Table Extraction).
// Input is TableRegion from Phase 2.6 (Table Detection).
type Table struct {
	Rows     [][]*Cell // Cells organized by row (row-major order)
	RowCount int       // Number of rows
	ColCount int       // Number of columns
	PageNum  int       // Page number where table was found (0-based)
	Bounds   Rectangle // Bounding rectangle
	Method   string    // "Lattice" or "Stream"
}

// NewTable creates a new Table with the specified dimensions.
//
// All cells are initialized to empty cells with proper row/column indices.
//
// Parameters:
//   - rowCount: Number of rows
//   - colCount: Number of columns
//
// Returns an error if dimensions are invalid (< 1).
func NewTable(rowCount, colCount int) (*Table, error) {
	if rowCount < 1 {
		return nil, fmt.Errorf("invalid row count: %d (must be >= 1)", rowCount)
	}
	if colCount < 1 {
		return nil, fmt.Errorf("invalid column count: %d (must be >= 1)", colCount)
	}

	// Initialize cells
	rows := make([][]*Cell, rowCount)
	for r := 0; r < rowCount; r++ {
		rows[r] = make([]*Cell, colCount)
		for c := 0; c < colCount; c++ {
			rows[r][c] = NewCell("", r, c)
		}
	}

	return &Table{
		Rows:     rows,
		RowCount: rowCount,
		ColCount: colCount,
		Method:   "Unknown",
	}, nil
}

// GetCell returns the cell at the specified row and column.
//
// Returns nil if the position is out of bounds.
func (t *Table) GetCell(row, col int) *Cell {
	if row < 0 || row >= t.RowCount || col < 0 || col >= t.ColCount {
		return nil
	}
	return t.Rows[row][col]
}

// SetCell sets the cell at the specified row and column.
//
// Returns an error if the position is out of bounds.
func (t *Table) SetCell(row, col int, cell *Cell) error {
	if row < 0 || row >= t.RowCount {
		return fmt.Errorf("row index out of bounds: %d (table has %d rows)", row, t.RowCount)
	}
	if col < 0 || col >= t.ColCount {
		return fmt.Errorf("column index out of bounds: %d (table has %d columns)", col, t.ColCount)
	}

	// Update cell position to match table position
	cell.Row = row
	cell.Column = col

	t.Rows[row][col] = cell
	return nil
}

// GetRow returns all cells in the specified row.
//
// Returns nil if row is out of bounds.
func (t *Table) GetRow(row int) []*Cell {
	if row < 0 || row >= t.RowCount {
		return nil
	}
	return t.Rows[row]
}

// GetColumn returns all cells in the specified column.
//
// Returns nil if column is out of bounds.
func (t *Table) GetColumn(col int) []*Cell {
	if col < 0 || col >= t.ColCount {
		return nil
	}

	cells := make([]*Cell, t.RowCount)
	for r := 0; r < t.RowCount; r++ {
		cells[r] = t.Rows[r][col]
	}
	return cells
}

// IsEmpty returns true if all cells are empty.
func (t *Table) IsEmpty() bool {
	for _, row := range t.Rows {
		for _, cell := range row {
			if !cell.IsEmpty() {
				return false
			}
		}
	}
	return true
}

// CellCount returns the total number of cells in the table.
func (t *Table) CellCount() int {
	return t.RowCount * t.ColCount
}

// NonEmptyCellCount returns the number of non-empty cells.
func (t *Table) NonEmptyCellCount() int {
	count := 0
	for _, row := range t.Rows {
		for _, cell := range row {
			if !cell.IsEmpty() {
				count++
			}
		}
	}
	return count
}

// HasMergedCells returns true if any cell is merged (spans multiple rows/cols).
func (t *Table) HasMergedCells() bool {
	for _, row := range t.Rows {
		for _, cell := range row {
			if cell.IsMerged() {
				return true
			}
		}
	}
	return false
}

// ToStringGrid converts the table to a simple 2D string array.
//
// This is useful for export formats (CSV, JSON) that don't support
// merged cells or formatting.
//
// For merged cells, the text appears in the top-left cell,
// and merged positions contain empty strings.
func (t *Table) ToStringGrid() [][]string {
	grid := make([][]string, t.RowCount)
	for r := 0; r < t.RowCount; r++ {
		grid[r] = make([]string, t.ColCount)
		for c := 0; c < t.ColCount; c++ {
			grid[r][c] = t.Rows[r][c].Text
		}
	}
	return grid
}

// String returns a string representation of the table (for debugging).
func (t *Table) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Table{rows=%d, cols=%d, method=%s, page=%d}\n",
		t.RowCount, t.ColCount, t.Method, t.PageNum)

	for r, row := range t.Rows {
		fmt.Fprintf(&sb, "  Row %d: [", r)
		for c, cell := range row {
			if c > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%q", cell.Text)
		}
		sb.WriteString("]\n")
	}

	return sb.String()
}

// Validate checks if the table structure is valid.
//
// A valid table must have:
//   - At least 1 row and 1 column
//   - All rows have the same number of columns
//   - All cells have correct row/column indices
//
// Returns an error describing the first validation failure, or nil if valid.
func (t *Table) Validate() error {
	// Check dimensions
	if t.RowCount < 1 {
		return fmt.Errorf("invalid row count: %d", t.RowCount)
	}
	if t.ColCount < 1 {
		return fmt.Errorf("invalid column count: %d", t.ColCount)
	}

	// Check row count matches
	if len(t.Rows) != t.RowCount {
		return fmt.Errorf("row count mismatch: expected %d, got %d", t.RowCount, len(t.Rows))
	}

	// Check each row
	for r, row := range t.Rows {
		// Check column count
		if len(row) != t.ColCount {
			return fmt.Errorf("column count mismatch in row %d: expected %d, got %d",
				r, t.ColCount, len(row))
		}

		// Check cell indices
		for c, cell := range row {
			if cell.Row != r {
				return fmt.Errorf("cell row index mismatch at (%d,%d): expected %d, got %d",
					r, c, r, cell.Row)
			}
			if cell.Column != c {
				return fmt.Errorf("cell column index mismatch at (%d,%d): expected %d, got %d",
					r, c, c, cell.Column)
			}
		}
	}

	return nil
}
