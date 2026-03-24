package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/coregx/gxpdf/internal/models/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// CSVExporter — validate error path (invalid table dimensions)
// ============================================================================

func TestCSVExporter_Export_InvalidTable(t *testing.T) {
	// NewTable with 0 rows returns an error, so we cannot call NewTable(0, 2).
	// Instead manually construct a Table with mismatched RowCount to trigger Validate.
	tbl := &table.Table{
		RowCount: -1,
		ColCount: 2,
	}
	e := NewCSVExporter()
	var buf bytes.Buffer
	err := e.Export(tbl, &buf)
	assert.Error(t, err)
}

// ============================================================================
// CSVExporter — ExportToString error propagation
// ============================================================================

func TestCSVExporter_ExportToString_InvalidTable(t *testing.T) {
	tbl := &table.Table{RowCount: -1, ColCount: 1}
	e := NewCSVExporter()
	_, err := e.ExportToString(tbl)
	assert.Error(t, err)
}

// ============================================================================
// JSONExporter — invalid table (validate error path)
// ============================================================================

func TestJSONExporter_Export_InvalidTable(t *testing.T) {
	tbl := &table.Table{RowCount: -1, ColCount: 1}
	e := NewJSONExporter()
	var buf bytes.Buffer
	err := e.Export(tbl, &buf)
	assert.Error(t, err)
}

// ============================================================================
// JSONExporter — ExportToString error propagation
// ============================================================================

func TestJSONExporter_ExportToString_InvalidTable(t *testing.T) {
	tbl := &table.Table{RowCount: -1, ColCount: 1}
	e := NewJSONExporter()
	_, err := e.ExportToString(tbl)
	assert.Error(t, err)
}

// ============================================================================
// JSONExporter — WithMetadata false (no metadata in output)
// ============================================================================

func TestJSONExporter_WithoutMetadata(t *testing.T) {
	tbl := createTestTable(t)
	e := NewJSONExporter().WithMetadata(false)

	result, err := e.ExportToString(tbl)
	require.NoError(t, err)

	// metadata field should be absent when not requested.
	assert.NotContains(t, result, `"metadata"`)
}

// ============================================================================
// ExcelExporter — cell with very long text (triggers maxWidth cap)
// ============================================================================

func TestExcelExporter_LongCellText(t *testing.T) {
	tbl, err := table.NewTable(1, 1)
	require.NoError(t, err)
	// A string longer than 50 chars to exercise the maxWidth cap.
	longText := strings.Repeat("A", 60)
	tbl.SetCell(0, 0, table.NewCell(longText, 0, 0))

	e := NewExcelExporter()
	var buf bytes.Buffer
	err = e.Export(tbl, &buf)
	require.NoError(t, err)
	assert.Greater(t, len(buf.Bytes()), 0)
}

// ============================================================================
// ExcelExporter — row 0 (header style) with no alignment
// ============================================================================

func TestExcelExporter_HeaderRow(t *testing.T) {
	tbl, err := table.NewTable(2, 2)
	require.NoError(t, err)
	tbl.SetCell(0, 0, table.NewCell("Header1", 0, 0))
	tbl.SetCell(0, 1, table.NewCell("Header2", 0, 1))
	tbl.SetCell(1, 0, table.NewCell("data1", 1, 0))
	tbl.SetCell(1, 1, table.NewCell("data2", 1, 1))

	e := NewExcelExporter()
	var buf bytes.Buffer
	err = e.Export(tbl, &buf)
	require.NoError(t, err)
	assert.Greater(t, len(buf.Bytes()), 100)
}

// ============================================================================
// ExcelExporter — data row with AlignLeft (default, no style applied)
// ============================================================================

func TestExcelExporter_DataRow_DefaultAlign(t *testing.T) {
	tbl, err := table.NewTable(2, 1)
	require.NoError(t, err)
	tbl.SetCell(0, 0, table.NewCell("H", 0, 0))
	tbl.SetCell(1, 0, table.NewCell("val", 1, 0)) // AlignLeft = default, no style

	e := NewExcelExporter()
	var buf bytes.Buffer
	err = e.Export(tbl, &buf)
	require.NoError(t, err)
}
