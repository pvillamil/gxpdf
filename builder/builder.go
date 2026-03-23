// Package builder provides a declarative, enterprise-grade API for generating
// PDF documents with GxPDF.
//
// It sits on top of the layout/ engine (pure computation) and the creator/
// package (PDF rendering), connecting them via a FontResolver bridge and a
// layout-to-PDF renderer.
//
// # Quick Start
//
//	b := builder.NewBuilder(
//	    builder.WithPageSize(layout.PageA4),
//	    builder.WithMargins(layout.Mm(20), layout.Mm(15), layout.Mm(20), layout.Mm(15)),
//	    builder.WithTitle("My Document"),
//	)
//
//	b.Page(func(p *builder.PageBuilder) {
//	    p.Content(func(c *builder.Container) {
//	        c.Text("Hello World", builder.Bold(), builder.FontSize(18))
//	        c.Spacer(layout.Mm(5))
//	        c.Text("This is the first paragraph.")
//	    })
//	})
//
//	pdfBytes, err := b.Build()
//
// # Architecture
//
// The builder package connects three layers:
//
//   - builder/    — user-facing API (this package)
//   - layout/     — pure layout engine (no PDF dependencies)
//   - creator/    — low-level PDF rendering backend
//
// Error handling: all errors are accumulated and returned as a joined error
// from Build(). This means layout definitions can be built without checking
// errors at each step; the first Build() call surfaces all problems at once.
package builder

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	builderinternal "github.com/coregx/gxpdf/builder/internal"
	"github.com/coregx/gxpdf/creator"
	"github.com/coregx/gxpdf/layout"
)

// Builder is the document builder entry point. It accumulates page definitions
// and document-level configuration, then generates a PDF on Build().
//
// Builder instances are not safe for concurrent use. Use one Builder per
// goroutine.
type Builder struct {
	cfg   config
	pages []*pageDef
	errs  []error
}

// NewBuilder creates a new Builder with the given document-level options.
// Sensible defaults are applied for any options not provided:
//   - Page size: A4
//   - Margins: 20mm top/bottom, 15mm left/right
//   - Default style: Helvetica 12pt, black text, 1.2 line height
//
// Example:
//
//	b := builder.NewBuilder(
//	    builder.WithPageSize(layout.PageLetter),
//	    builder.WithFont("Inter", interTTF),
//	)
func NewBuilder(opts ...Option) *Builder {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Builder{cfg: cfg}
}

// Page adds a page definition to the document. The fn callback receives a
// PageBuilder for defining header, content, and footer zones.
//
// A single Page definition may produce multiple physical PDF pages if the
// content overflows — the paginator handles splitting automatically.
//
// Example:
//
//	b.Page(func(p *builder.PageBuilder) {
//	    p.Header(func(h *builder.Container) { h.Text("Header") })
//	    p.Content(func(c *builder.Container) { c.Text("Body text") })
//	    p.Footer(func(f *builder.Container) {
//	        f.PageNumber(layout.PageNumberPlaceholder + " / " + layout.TotalPagesPlaceholder)
//	    })
//	})
func (b *Builder) Page(fn func(*PageBuilder)) {
	pb := &PageBuilder{b: b}
	fn(pb)
	b.pages = append(b.pages, &pb.def)
}

// Build runs the full PDF generation pipeline and returns the resulting PDF
// bytes. All accumulated errors (from font loading, invalid options, etc.) are
// returned as a joined error.
//
// The pipeline:
//  1. Check accumulated errors.
//  2. Convert page definitions to layout.PageDef values.
//  3. Run the paginator to produce positioned blocks.
//  4. Walk the block tree with the renderer to emit PDF content via creator.
//  5. Serialise to bytes and return.
func (b *Builder) Build() ([]byte, error) {
	if err := b.accumulatedError(); err != nil {
		return nil, err
	}

	cr := creator.New()
	if err := b.renderTo(cr); err != nil {
		return nil, err
	}

	pdfBytes, err := cr.Bytes()
	if err != nil {
		return nil, fmt.Errorf("serialise PDF: %w", err)
	}
	return pdfBytes, nil
}

// BuildTo writes the generated PDF to the given io.Writer.
// Returns an error if PDF generation or writing fails.
func (b *Builder) BuildTo(w io.Writer) error {
	pdfBytes, err := b.Build()
	if err != nil {
		return err
	}
	_, err = w.Write(pdfBytes)
	return err
}

// BuildToFile writes the generated PDF to the file at path, creating or
// truncating the file as needed.
func (b *Builder) BuildToFile(path string) error {
	pdfBytes, err := b.Build()
	if err != nil {
		return err
	}
	return os.WriteFile(path, pdfBytes, 0644)
}

// --- Internal helpers ---

// defaultStyle returns the document-level default style. Consumers (Container,
// PageBuilder) call this to initialise per-element styles.
func (b *Builder) defaultStyle() layout.Style {
	return b.cfg.defaultStyle
}

// fontResolver returns a layout.FontResolver backed by the registered custom
// fonts and the Standard 14 font metrics.
func (b *Builder) fontResolver() layout.FontResolver {
	return builderinternal.NewFontBridge(b.cfg.fonts)
}

// addError appends an error to the accumulated error list.
func (b *Builder) addError(err error) {
	if err != nil {
		b.errs = append(b.errs, err)
	}
}

// accumulatedError returns all accumulated errors joined into a single error,
// or nil if there are none.
func (b *Builder) accumulatedError() error {
	if len(b.errs) == 0 {
		return nil
	}
	return errors.Join(b.errs...)
}

// renderTo runs the layout pipeline and emits PDF content into cr.
func (b *Builder) renderTo(cr *creator.Creator) error {
	// Build layout page definitions.
	layoutPages := make([]*layout.PageDef, 0, len(b.pages))
	for _, pd := range b.pages {
		layoutPages = append(layoutPages, pd.toLayoutPageDef(b.cfg.pageSize, b.cfg.margins))
	}

	// Run the paginator.
	paginator := &layout.Paginator{
		Fonts: b.fontResolver(),
	}
	pageLayouts := paginator.Paginate(layoutPages)

	// Render page layouts via the PDF renderer.
	renderer := builderinternal.NewPDFRenderer(cr, b.cfg.fonts)
	meta := builderinternal.ExportedMetadata(b.cfg.meta.title, b.cfg.meta.author)
	if err := renderer.RenderDocument(pageLayouts, meta); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	return nil
}

// Bytes is an alias for Build() for callers that prefer the creator-style API.
func (b *Builder) Bytes() ([]byte, error) {
	return b.Build()
}

// WriteToFile is an alias for BuildToFile() for callers that prefer the
// creator-style API.
func (b *Builder) WriteToFile(path string) error {
	return b.BuildToFile(path)
}

// WriteTo implements io.WriterTo.
func (b *Builder) WriteTo(w io.Writer) (int64, error) {
	pdfBytes, err := b.Build()
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(w, bytes.NewReader(pdfBytes))
	return n, err
}
