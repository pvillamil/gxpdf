package gxpdf

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/coregx/gxpdf/logging"
)

// TestErrorLogging verifies that errors are logged via slog
func TestErrorLogging(t *testing.T) {
	// Setup a buffer to capture log output
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	logging.SetLogger(logger)
	defer logging.SetLogger(nil) // Restore to discard after test

	doc, err := Open("testdata/pdfs/minimal.pdf")
	if err != nil {
		t.Skip("testdata/pdfs/minimal.pdf not available, skipping error logging tests")
	}
	defer doc.Close()

	// Test ExtractTablesFromPage with out of range - should log error
	buf.Reset()
	tables := doc.ExtractTablesFromPage(9999)
	if tables != nil {
		t.Error("expected nil tables for out of range page")
	}

	// Verify error was logged
	logOutput := buf.String()
	if !strings.Contains(logOutput, "page number out of range") {
		t.Errorf("expected error log for out of range page, got: %s", logOutput)
	}
}

// TestExtractTextFromPagePropagatesError verifies ExtractTextFromPage returns errors
func TestExtractTextFromPagePropagatesError(t *testing.T) {
	doc, err := Open("testdata/pdfs/minimal.pdf")
	if err != nil {
		t.Skip("testdata/pdfs/minimal.pdf not available, skipping test")
	}
	defer doc.Close()

	// Test out of range
	_, err = doc.ExtractTextFromPage(9999)
	if err == nil {
		t.Error("expected error for out of range page")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("expected 'out of range' error, got: %v", err)
	}

	// Test page 0 (invalid)
	_, err = doc.ExtractTextFromPage(0)
	if err == nil {
		t.Error("expected error for page 0")
	}
}

// TestConvenienceMethodsLogErrors verifies that convenience methods log errors
func TestConvenienceMethodsLogErrors(t *testing.T) {
	// Setup logger to capture errors
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	logging.SetLogger(logger)
	defer logging.SetLogger(nil)

	doc, err := Open("testdata/pdfs/minimal.pdf")
	if err != nil {
		t.Skip("testdata/pdfs/minimal.pdf not available, skipping test")
	}
	defer doc.Close()

	// Get a page
	page := doc.Page(0)
	if page == nil {
		t.Fatal("expected page 0")
	}

	// Close reader to force errors
	doc.reader.Close()

	// Call convenience methods - they should log but not panic
	_ = page.ExtractText()
	_ = page.ExtractTables()
	_ = page.GetImages()
	_ = doc.ExtractTables()
	_ = doc.GetImages()
	_ = doc.ExtractTablesFromPage(1)

	// Verify errors were logged
	logOutput := buf.String()
	if !strings.Contains(logOutput, "error") {
		t.Error("expected errors to be logged from convenience methods")
	}
}
