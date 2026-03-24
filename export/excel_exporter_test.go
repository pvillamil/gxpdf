package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/coregx/gxpdf/internal/models/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExcelExporter(t *testing.T) {
	e := NewExcelExporter()
	assert.NotNil(t, e)
	assert.NotNil(t, e.options)
	assert.Equal(t, "Table", e.sheetName)
}

func TestNewExcelExporterWithOptions(t *testing.T) {
	opts := &ExportOptions{
		IncludeEmpty:  true,
		PreserveSpans: true,
		Delimiter:     ",",
	}
	e := NewExcelExporterWithOptions(opts)
	assert.NotNil(t, e)
	assert.True(t, e.options.PreserveSpans)
}

func TestNewExcelExporterWithOptions_Nil(t *testing.T) {
	e := NewExcelExporterWithOptions(nil)
	assert.NotNil(t, e)
	assert.NotNil(t, e.options)
}

func TestExcelExporter_WithSheetName(t *testing.T) {
	e := NewExcelExporter().WithSheetName("MySheet")
	assert.Equal(t, "MySheet", e.sheetName)
}

func TestExcelExporter_WithMergedCells(t *testing.T) {
	e := NewExcelExporter().WithMergedCells(true)
	assert.True(t, e.options.PreserveSpans)

	e2 := NewExcelExporter().WithMergedCells(false)
	assert.False(t, e2.options.PreserveSpans)
}

func TestExcelExporter_ContentType(t *testing.T) {
	e := NewExcelExporter()
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", e.ContentType())
}

func TestExcelExporter_FileExtension(t *testing.T) {
	e := NewExcelExporter()
	assert.Equal(t, ".xlsx", e.FileExtension())
}

func TestExcelExporter_ExportToString(t *testing.T) {
	e := NewExcelExporter()
	tbl := createTestTable(t)
	_, err := e.ExportToString(tbl)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "binary")
}

func TestExcelExporter_Export_NilTable(t *testing.T) {
	e := NewExcelExporter()
	var buf bytes.Buffer
	err := e.Export(nil, &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestExcelExporter_Export_BasicTable(t *testing.T) {
	tbl := createTestTable(t)
	e := NewExcelExporter()

	var buf bytes.Buffer
	err := e.Export(tbl, &buf)
	require.NoError(t, err)

	// Should produce non-empty XLSX data
	data := buf.Bytes()
	assert.Greater(t, len(data), 100)
	// XLSX files start with PK (zip magic bytes)
	assert.Equal(t, byte('P'), data[0])
	assert.Equal(t, byte('K'), data[1])
}

func TestExcelExporter_ExportToBytes(t *testing.T) {
	tbl := createTestTable(t)
	e := NewExcelExporter()

	data, err := e.ExportToBytes(tbl)
	require.NoError(t, err)
	assert.Greater(t, len(data), 100)
}

func TestExcelExporter_ExportToBytes_NilTable(t *testing.T) {
	e := NewExcelExporter()
	_, err := e.ExportToBytes(nil)
	assert.Error(t, err)
}

func TestExcelExporter_Export_CustomSheetName(t *testing.T) {
	tbl := createTestTable(t)
	e := NewExcelExporter().WithSheetName("Results")

	var buf bytes.Buffer
	err := e.Export(tbl, &buf)
	require.NoError(t, err)
	assert.Greater(t, len(buf.Bytes()), 100)
}

func TestExcelExporter_Export_DefaultSheet(t *testing.T) {
	tbl := createTestTable(t)
	e := NewExcelExporter().WithSheetName("Sheet1")

	var buf bytes.Buffer
	err := e.Export(tbl, &buf)
	require.NoError(t, err)
	assert.Greater(t, len(buf.Bytes()), 100)
}

func TestExcelExporter_Export_WithMergedCells(t *testing.T) {
	tbl := createTestTable(t)
	mergedCell := tbl.GetCell(0, 0).WithRowSpan(2).WithColSpan(2)
	tbl.SetCell(0, 0, mergedCell)

	e := NewExcelExporter().WithMergedCells(true)
	var buf bytes.Buffer
	err := e.Export(tbl, &buf)
	require.NoError(t, err)
	assert.Greater(t, len(buf.Bytes()), 100)
}

func TestExcelExporter_Export_WithAlignment(t *testing.T) {
	tbl, err := table.NewTable(2, 3)
	require.NoError(t, err)

	tbl.SetCell(0, 0, table.NewCell("Header", 0, 0))
	tbl.SetCell(0, 1, table.NewCell("Center", 0, 1).WithAlignment(table.AlignCenter))
	tbl.SetCell(0, 2, table.NewCell("Right", 0, 2).WithAlignment(table.AlignRight))
	tbl.SetCell(1, 0, table.NewCell("data1", 1, 0))
	tbl.SetCell(1, 1, table.NewCell("data2", 1, 1).WithAlignment(table.AlignCenter))
	tbl.SetCell(1, 2, table.NewCell("data3", 1, 2).WithAlignment(table.AlignRight))

	e := NewExcelExporter()
	var buf bytes.Buffer
	err = e.Export(tbl, &buf)
	require.NoError(t, err)
	assert.Greater(t, len(buf.Bytes()), 100)
}

func TestExcelExporter_Export_EmptyTable(t *testing.T) {
	tbl, err := table.NewTable(1, 1)
	require.NoError(t, err)
	tbl.SetCell(0, 0, table.NewCell("", 0, 0))

	e := NewExcelExporter()
	var buf bytes.Buffer
	err = e.Export(tbl, &buf)
	require.NoError(t, err)
}

func TestExcelExporter_Export_LargeTable(t *testing.T) {
	tbl, err := table.NewTable(10, 5)
	require.NoError(t, err)
	for r := 0; r < 10; r++ {
		for c := 0; c < 5; c++ {
			tbl.SetCell(r, c, table.NewCell("cell", r, c))
		}
	}

	e := NewExcelExporter()
	var buf bytes.Buffer
	err = e.Export(tbl, &buf)
	require.NoError(t, err)
	assert.Greater(t, len(buf.Bytes()), 100)
}

func TestExcelExporter_Export_LongTextAutoFit(t *testing.T) {
	tbl, err := table.NewTable(1, 2)
	require.NoError(t, err)
	longText := strings.Repeat("A", 200) // Longer than maxWidth
	tbl.SetCell(0, 0, table.NewCell(longText, 0, 0))
	tbl.SetCell(0, 1, table.NewCell("short", 0, 1))

	e := NewExcelExporter()
	var buf bytes.Buffer
	err = e.Export(tbl, &buf)
	require.NoError(t, err)
}
