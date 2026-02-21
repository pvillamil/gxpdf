// Package document provides the domain model for PDF document creation.
//
// This package contains the Document aggregate root and related entities
// following Domain-Driven Design (DDD) principles.
package document

import (
	"errors"
	"fmt"
	"time"

	"github.com/coregx/gxpdf/internal/models/content"
	"github.com/coregx/gxpdf/internal/models/types"
)

// Document is the aggregate root for PDF document creation.
//
// It manages the collection of pages and ensures document-level consistency.
// All modifications to the document go through this entity (Rich Domain Model).
//
// Example:
//
//	doc := document.NewDocument()
//	doc.SetMetadata("My Document", "Author Name", "Subject")
//	page, _ := doc.AddPage(document.A4)
//	doc.Validate()
type Document struct {
	// Identity
	id string // Internal ID (for tracking)

	// Document properties
	version      types.Version
	title        string
	author       string
	subject      string
	keywords     []string
	creator      string
	producer     string
	creationDate time.Time
	modDate      time.Time

	// Content
	pages []*Page

	// Behavior (Rich Domain Model)
	// pageNumbering could be added here for custom page numbering strategies
}

// NewDocument creates a new empty PDF document.
//
// The document is initialized with:
// - PDF version 1.7
// - Creation/modification dates set to now
// - Creator and producer set to "gxpdf"
// - Empty pages collection
func NewDocument() *Document {
	now := time.Now()
	return &Document{
		id:           generateID(),
		version:      types.PDF17, // PDF 1.7
		creator:      "gxpdf",
		producer:     "gxpdf (github.com/coregx/gxpdf)",
		creationDate: now,
		modDate:      now,
		pages:        make([]*Page, 0),
	}
}

// AddPage adds a new page to the document.
//
// Returns the newly created page for method chaining.
//
// Example:
//
//	page, err := doc.AddPage(document.A4)
//	if err != nil {
//	    return err
//	}
//	// Use page...
func (d *Document) AddPage(pageSize PageSize) (*Page, error) {
	page := NewPage(len(d.pages), pageSize)
	d.pages = append(d.pages, page)
	d.modDate = time.Now()
	return page, nil
}

// AddPageWithRect adds a new page to the document with explicitly specified dimensions.
//
// widthPt and heightPt are the page dimensions in PDF points (1 pt = 1/72 inch).
// Use CustomPageSize to build the rectangle, or InchesToPoints/MMToPoints to convert.
//
// Returns the newly created page for method chaining.
//
// Example:
//
//	// A custom 6×9 inch page
//	rect := document.CustomPageSize(6*72, 9*72)
//	page, err := doc.AddPageWithRect(rect)
func (d *Document) AddPageWithRect(rect types.Rectangle) (*Page, error) {
	page := &Page{
		number:            len(d.pages),
		mediaBox:          rect,
		rotation:          0,
		contents:          make([]content.Content, 0),
		linkAnnotations:   make([]*LinkAnnotation, 0),
		textAnnotations:   make([]*TextAnnotation, 0),
		markupAnnotations: make([]*MarkupAnnotation, 0),
		stampAnnotations:  make([]*StampAnnotation, 0),
		formFields:        make([]*FormField, 0),
	}
	d.pages = append(d.pages, page)
	d.modDate = time.Now()
	return page, nil
}

// InsertPage inserts a page at the specified index.
//
// This will renumber all subsequent pages.
//
// Returns an error if the index is out of bounds.
func (d *Document) InsertPage(index int, pageSize PageSize) (*Page, error) {
	if index < 0 || index > len(d.pages) {
		return nil, fmt.Errorf("%w: index %d out of range [0, %d]", ErrInvalidPageIndex, index, len(d.pages))
	}

	page := NewPage(index, pageSize)
	d.pages = append(d.pages[:index], append([]*Page{page}, d.pages[index:]...)...)

	// Renumber pages after insertion
	d.renumberPages()
	d.modDate = time.Now()

	return page, nil
}

// RemovePage removes the page at the specified index.
//
// This will renumber all subsequent pages.
//
// Returns an error if the index is out of bounds.
func (d *Document) RemovePage(index int) error {
	if index < 0 || index >= len(d.pages) {
		return fmt.Errorf("%w: index %d out of range [0, %d)", ErrInvalidPageIndex, index, len(d.pages))
	}

	d.pages = append(d.pages[:index], d.pages[index+1:]...)
	d.renumberPages()
	d.modDate = time.Now()

	return nil
}

// Page returns the page at the specified index (0-based).
//
// Returns an error if the index is out of bounds.
func (d *Document) Page(index int) (*Page, error) {
	if index < 0 || index >= len(d.pages) {
		return nil, fmt.Errorf("%w: index %d out of range [0, %d)", ErrInvalidPageIndex, index, len(d.pages))
	}
	return d.pages[index], nil
}

// Pages returns all pages in the document.
//
// The returned slice is a copy to prevent external modifications.
func (d *Document) Pages() []*Page {
	// Return a copy to prevent external modifications
	result := make([]*Page, len(d.pages))
	copy(result, d.pages)
	return result
}

// PageCount returns the number of pages in the document.
func (d *Document) PageCount() int {
	return len(d.pages)
}

// SetMetadata sets document metadata.
//
// Empty strings are ignored (existing values are kept).
//
// Example:
//
//	doc.SetMetadata("My Document", "John Doe", "Annual Report", "2025", "PDF", "Report")
func (d *Document) SetMetadata(title, author, subject string, keywords ...string) {
	if title != "" {
		d.title = title
	}
	if author != "" {
		d.author = author
	}
	if subject != "" {
		d.subject = subject
	}
	if len(keywords) > 0 {
		d.keywords = keywords
	}
	d.modDate = time.Now()
}

// Title returns the document title.
func (d *Document) Title() string {
	return d.title
}

// Author returns the document author.
func (d *Document) Author() string {
	return d.author
}

// Subject returns the document subject.
func (d *Document) Subject() string {
	return d.subject
}

// Keywords returns the document keywords.
func (d *Document) Keywords() []string {
	// Return a copy to prevent external modifications
	result := make([]string, len(d.keywords))
	copy(result, d.keywords)
	return result
}

// Version returns the PDF version.
func (d *Document) Version() types.Version {
	return d.version
}

// Creator returns the creator application.
func (d *Document) Creator() string {
	return d.creator
}

// Producer returns the producer application.
func (d *Document) Producer() string {
	return d.producer
}

// CreationDate returns the document creation date.
func (d *Document) CreationDate() time.Time {
	return d.creationDate
}

// ModificationDate returns the last modification date.
func (d *Document) ModificationDate() time.Time {
	return d.modDate
}

// renumberPages updates page numbers after insertion/deletion.
//
// This is an internal method that maintains consistency.
func (d *Document) renumberPages() {
	for i, page := range d.pages {
		page.number = i
	}
}

// Validate checks document consistency.
//
// Returns an error if:
// - Document has no pages
// - Any page is invalid
func (d *Document) Validate() error {
	if len(d.pages) == 0 {
		return ErrEmptyDocument
	}

	for i, page := range d.pages {
		if page == nil {
			return fmt.Errorf("page %d is nil", i)
		}
		if err := page.Validate(); err != nil {
			return fmt.Errorf("page %d validation failed: %w", i, err)
		}
	}

	return nil
}

// Domain errors
var (
	// ErrInvalidPageIndex is returned when a page index is out of bounds.
	ErrInvalidPageIndex = errors.New("invalid page index")

	// ErrEmptyDocument is returned when validating a document with no pages.
	ErrEmptyDocument = errors.New("document has no pages")
)

// generateID generates a unique document ID.
//
// This is a simple implementation using timestamp.
// In production, you might want to use UUID or similar.
func generateID() string {
	return fmt.Sprintf("doc_%d", time.Now().UnixNano())
}
