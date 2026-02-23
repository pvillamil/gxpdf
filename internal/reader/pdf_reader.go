// Package reader provides application-layer PDF reading functionality.
//
// This package wraps the infrastructure-layer parser to provide
// a clean application-layer API for PDF document reading.
package reader

import (
	"github.com/coregx/gxpdf/internal/parser"
)

// PdfReader is an application-layer wrapper around the infrastructure parser.Reader.
//
// It provides a clean separation between infrastructure (low-level parsing)
// and application concerns (document structure navigation).
type PdfReader struct {
	reader *parser.Reader
}

// NewPdfReader creates a new PDF reader from a file path.
//
// The file is opened and parsed immediately. Remember to call Close()
// when done to release resources.
//
// Example:
//
//	reader, err := reader.NewPdfReader("document.pdf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
func NewPdfReader(path string) (*PdfReader, error) {
	r, err := parser.OpenPDF(path)
	if err != nil {
		return nil, err
	}
	return &PdfReader{reader: r}, nil
}

// NewPdfReaderWithPassword creates a new PDF reader from a password-protected file.
//
// Use this for encrypted PDFs that require a non-empty password.
// For PDFs with empty user password, NewPdfReader handles them transparently.
func NewPdfReaderWithPassword(path, password string) (*PdfReader, error) {
	r, err := parser.OpenPDFWithPassword(path, password)
	if err != nil {
		return nil, err
	}
	return &PdfReader{reader: r}, nil
}

// Close closes the PDF file and releases resources.
func (r *PdfReader) Close() error {
	return r.reader.Close()
}

// PageCount returns the total number of pages in the PDF.
func (r *PdfReader) PageCount() int {
	count, err := r.reader.GetPageCount()
	if err != nil {
		return 0
	}
	return count
}

// GetPage returns the page dictionary for the specified page index (0-based).
func (r *PdfReader) GetPage(pageIndex int) (*parser.Dictionary, error) {
	return r.reader.GetPage(pageIndex)
}

// Version returns the PDF version string (e.g., "1.7").
func (r *PdfReader) Version() string {
	return r.reader.Version()
}

// GetParserReader returns the underlying parser.Reader for advanced operations.
//
// This is used internally by extractors that need direct access to the parser.
// Most users should not need this method - use the higher-level methods instead.
func (r *PdfReader) GetParserReader() *parser.Reader {
	return r.reader
}
