package internal

import (
	"fmt"

	"github.com/coregx/gxpdf/creator"
	"github.com/coregx/gxpdf/layout"
)

// PDFRenderer walks a slice of layout.PageLayout values and emits PDF content
// by calling creator methods. It bridges the layout coordinate system
// (top-down, Y=0 at top) to the PDF coordinate system (bottom-up, Y=0 at bottom).
//
// Coordinate conversion:
//
//	pdfY = pageHeight - layoutY
//
// The renderer is stateless across pages; each page creates a new creator.Page.
type PDFRenderer struct {
	cr *creator.Creator
	// customFonts is the map of custom fonts for text rendering.
	customFonts map[string]*creator.CustomFont
}

// NewPDFRenderer creates a PDFRenderer using the given Creator and custom fonts.
func NewPDFRenderer(cr *creator.Creator, customFonts map[string]*creator.CustomFont) *PDFRenderer {
	if customFonts == nil {
		customFonts = make(map[string]*creator.CustomFont)
	}
	return &PDFRenderer{cr: cr, customFonts: customFonts}
}

// RenderDocument walks all PageLayouts, creating one creator.Page per layout
// page and rendering every Block tree onto it. If pages is empty (e.g. no
// content was added), a single blank A4 page is created so that the creator
// validation (which requires at least one page) does not fail.
func (r *PDFRenderer) RenderDocument(pages []layout.PageLayout, meta metadata) error {
	if meta.Title != "" {
		r.cr.SetTitle(meta.Title)
	}
	if meta.Author != "" {
		r.cr.SetAuthor(meta.Author)
	}

	// Ensure at least one page exists.
	if len(pages) == 0 {
		_, err := r.cr.NewPageWithDimensions(595.276, 841.890) // A4
		return err
	}

	for _, pl := range pages {
		page, err := r.cr.NewPageWithDimensions(pl.Size.Width, pl.Size.Height)
		if err != nil {
			return fmt.Errorf("create page (%.0fx%.0f): %w", pl.Size.Width, pl.Size.Height, err)
		}
		pr := &pageRenderer{
			page:       page,
			pageHeight: pl.Size.Height,
			fonts:      r.customFonts,
		}
		for i := range pl.Blocks {
			pr.renderBlock(&pl.Blocks[i])
		}
	}
	return nil
}

// metadata holds document-level metadata for PDF generation.
type metadata struct {
	Title  string
	Author string
}

// ExportedMetadata constructs a metadata value from the builder config.
// This is the only way for the builder package to create an internal metadata
// struct while keeping the type unexported within the package.
func ExportedMetadata(title, author string) metadata {
	return metadata{Title: title, Author: author}
}

// pageRenderer renders one physical page by translating layout coordinates to
// PDF coordinates and calling creator.Page methods.
type pageRenderer struct {
	page       *creator.Page
	pageHeight float64
	fonts      map[string]*creator.CustomFont
}

// renderBlock recursively renders a Block and its children. The baseX and baseY
// parameters are the accumulated absolute offset for nested blocks. Layout
// coordinates are top-left origin (Y increases downward); PDF coordinates are
// bottom-left origin (Y increases upward).
func (pr *pageRenderer) renderBlock(b *layout.Block) {
	// If this block has a Draw closure, invoke it with an adaptor Renderer
	// that translates top-left coordinates to PDF bottom-left coordinates.
	if b.Draw != nil {
		adaptor := &blockAdaptor{
			pr:    pr,
			baseX: b.X,
			baseY: b.Y,
		}
		b.Draw(adaptor)
	}
	// Recurse into children.
	for i := range b.Children {
		child := b.Children[i]
		// Accumulate the parent's offset for the child.
		child.X += b.X
		child.Y += b.Y
		pr.renderBlock(&child)
	}
}

// blockAdaptor implements layout.Renderer for a specific block, translating
// coordinates relative to the block's top-left corner into absolute PDF
// coordinates (bottom-left origin).
type blockAdaptor struct {
	pr    *pageRenderer
	baseX float64
	baseY float64
}

// toPDFCoords converts a layout point (top-down, relative to block origin) to
// an absolute PDF point (bottom-up, absolute page coordinates).
func (a *blockAdaptor) toPDFCoords(x, y float64) (float64, float64) {
	absX := a.baseX + x
	absY := a.pr.pageHeight - (a.baseY + y)
	return absX, absY
}

// DrawText implements layout.Renderer. It renders a text string at the given
// layout position. The Y coordinate is adjusted so that text is placed at the
// baseline, using the ascender to offset from the top of the line box.
func (a *blockAdaptor) DrawText(text string, x, y float64, font layout.FontRef, size float64, color layout.Color, options layout.TextDrawOptions) {
	if text == "" {
		return
	}

	pdfX, pdfYTop := a.toPDFCoords(x, y)

	// In PDF, AddText places text at the baseline. In layout, Y is the top of
	// the line box. We need to subtract the ascender to get the baseline Y.
	// The ascender moves us down from top in layout (= up in PDF).
	var ascender float64
	if cf, ok := a.pr.fonts[font.Family]; ok {
		ascender = cf.Ascender(size)
	} else {
		ascender = creator.FontAscender(resolveStandard14Name(font), size)
	}
	// PDF Y baseline = pageHeight - (layoutY + ascender)
	pdfY := pdfYTop - ascender

	creatorColor := creator.Color{
		R: color.R,
		G: color.G,
		B: color.B,
	}

	if cf, ok := a.pr.fonts[font.Family]; ok {
		// Custom font path.
		_ = a.pr.page.AddTextCustomFontColor(text, pdfX, pdfY, cf, size, creatorColor)
	} else {
		// Standard 14 font path.
		fontName := resolveStandard14Name(font)
		_ = a.pr.page.AddTextColor(text, pdfX, pdfY, fontName, size, creatorColor)
	}

	// Draw underline if requested.
	if options.Underline {
		underlineY := pdfY - size*0.1
		_ = a.pr.page.DrawLine(pdfX, underlineY, pdfX+creator.MeasureText(resolveStandard14Name(font), text, size), underlineY, &creator.LineOptions{
			Color: creatorColor,
			Width: size * 0.05,
		})
	}

	// Draw strikethrough if requested.
	if options.Strikethrough {
		midY := pdfY + size*0.3
		_ = a.pr.page.DrawLine(pdfX, midY, pdfX+creator.MeasureText(resolveStandard14Name(font), text, size), midY, &creator.LineOptions{
			Color: creatorColor,
			Width: size * 0.05,
		})
	}
}

// DrawRect implements layout.Renderer.
func (a *blockAdaptor) DrawRect(x, y, width, height float64, fill *layout.Color, stroke *layout.Color, strokeWidth float64) {
	pdfX, pdfYTop := a.toPDFCoords(x, y)
	// PDF rect origin is lower-left; layout rect origin is upper-left.
	// pdfY lower-left = pdfYTop - height
	pdfY := pdfYTop - height

	opts := &creator.RectOptions{}
	if fill != nil {
		c := creator.Color{R: fill.R, G: fill.G, B: fill.B}
		opts.FillColor = &c
	}
	if stroke != nil {
		c := creator.Color{R: stroke.R, G: stroke.G, B: stroke.B}
		opts.StrokeColor = &c
		opts.StrokeWidth = strokeWidth
		if opts.StrokeWidth <= 0 {
			opts.StrokeWidth = 1.0
		}
	}
	_ = a.pr.page.DrawRect(pdfX, pdfY, width, height, opts)
}

// DrawLine implements layout.Renderer.
func (a *blockAdaptor) DrawLine(x1, y1, x2, y2 float64, color layout.Color, width float64) {
	pdfX1, pdfY1 := a.toPDFCoords(x1, y1)
	pdfX2, pdfY2 := a.toPDFCoords(x2, y2)
	c := creator.Color{R: color.R, G: color.G, B: color.B}
	w := width
	if w <= 0 {
		w = 1.0
	}
	_ = a.pr.page.DrawLine(pdfX1, pdfY1, pdfX2, pdfY2, &creator.LineOptions{
		Color: c,
		Width: w,
	})
}

// DrawImage implements layout.Renderer.
// TODO(Phase 4): Load image bytes via creator.LoadImageFromBytes and render
// using creator.Page.DrawImage once the creator exposes byte-loading.
// For now this emits a placeholder gray rectangle.
func (a *blockAdaptor) DrawImage(data []byte, x, y, width, height float64) {
	pdfX, pdfYTop := a.toPDFCoords(x, y)
	pdfY := pdfYTop - height
	placeholder := creator.Color{R: 0.9, G: 0.9, B: 0.9}
	_ = a.pr.page.DrawRectFilled(pdfX, pdfY, width, height, placeholder)
}

// PushState implements layout.Renderer. Currently a no-op stub; clip state
// management is handled at the BeginClipRect/EndClip level.
func (a *blockAdaptor) PushState() {
	// Clip state is saved implicitly by BeginClipRect in creator.
}

// PopState implements layout.Renderer.
func (a *blockAdaptor) PopState() {
	_ = a.pr.page.EndClip()
}

// SetClipRect implements layout.Renderer. Sets a rectangular clipping region.
// Layout uses top-left origin; creator.BeginClipRect uses bottom-left origin.
func (a *blockAdaptor) SetClipRect(x, y, width, height float64) {
	pdfX, pdfYTop := a.toPDFCoords(x, y)
	pdfY := pdfYTop - height
	_ = a.pr.page.BeginClipRect(pdfX, pdfY, width, height)
}

// resolveStandard14Name maps a layout.FontRef to a creator.FontName for
// Standard 14 fonts. This is the same logic as FontBridge.resolveStandard14
// but local to the renderer to avoid package coupling.
func resolveStandard14Name(font layout.FontRef) creator.FontName {
	family := font.Family
	bold := font.Weight == layout.WeightBold
	italic := font.Style == layout.StyleItalic

	switch {
	case isTimesFamily(family):
		switch {
		case bold && italic:
			return creator.TimesBoldItalic
		case bold:
			return creator.TimesBold
		case italic:
			return creator.TimesItalic
		default:
			return creator.TimesRoman
		}
	case isCourierFamily(family):
		switch {
		case bold && italic:
			return creator.CourierBoldOblique
		case bold:
			return creator.CourierBold
		case italic:
			return creator.CourierOblique
		default:
			return creator.Courier
		}
	case family == "Symbol":
		return creator.Symbol
	case family == "ZapfDingbats":
		return creator.ZapfDingbats
	default:
		switch {
		case bold && italic:
			return creator.HelveticaBoldOblique
		case bold:
			return creator.HelveticaBold
		case italic:
			return creator.HelveticaOblique
		default:
			return creator.Helvetica
		}
	}
}

func isTimesFamily(family string) bool {
	return family == "Times" || family == "Times New Roman" || family == "Times-Roman"
}

func isCourierFamily(family string) bool {
	return family == "Courier" || family == "Courier New"
}
