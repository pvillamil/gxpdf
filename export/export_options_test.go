package export

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultExportOptions(t *testing.T) {
	opts := DefaultExportOptions()
	assert.NotNil(t, opts)
	assert.True(t, opts.IncludeEmpty)
	assert.False(t, opts.PreserveSpans)
	assert.Equal(t, ",", opts.Delimiter)
	assert.False(t, opts.IncludeMetadata)
	assert.False(t, opts.PrettyPrint)
}

func TestCSVExporter_NewWithOptions(t *testing.T) {
	opts := &ExportOptions{
		Delimiter: ";",
	}
	e := NewCSVExporterWithOptions(opts)
	assert.NotNil(t, e)
	assert.Equal(t, ";", e.options.Delimiter)
}

func TestCSVExporter_NewWithOptions_Nil(t *testing.T) {
	e := NewCSVExporterWithOptions(nil)
	assert.NotNil(t, e)
	assert.NotNil(t, e.options)
	assert.Equal(t, ",", e.options.Delimiter)
}

func TestJSONExporter_NewWithOptions(t *testing.T) {
	opts := &ExportOptions{
		PrettyPrint:     true,
		IncludeMetadata: true,
	}
	e := NewJSONExporterWithOptions(opts)
	assert.NotNil(t, e)
	assert.True(t, e.options.PrettyPrint)
	assert.True(t, e.options.IncludeMetadata)
}

func TestJSONExporter_NewWithOptions_Nil(t *testing.T) {
	e := NewJSONExporterWithOptions(nil)
	assert.NotNil(t, e)
	assert.NotNil(t, e.options)
}

// TestTableExporter_Interface verifies all exporters implement the interface
func TestTableExporter_Interface(t *testing.T) {
	// This ensures the interface is satisfied at compile time
	var _ TableExporter = NewCSVExporter()
	var _ TableExporter = NewJSONExporter()
	var _ TableExporter = NewExcelExporter()
}
