package layout

import (
	"fmt"
	"strings"
)

// PageNumberPlaceholder is injected into text content where the current page
// number should appear. It is replaced in a second pass after all pages are
// known.
const PageNumberPlaceholder = "\x00PAGE\x00"

// TotalPagesPlaceholder is injected into text content where the total page
// count should appear.
const TotalPagesPlaceholder = "\x00TOTAL\x00"

// PageDef describes a page template: its physical size, margins, and the
// content elements to be rendered on each page of the document.
type PageDef struct {
	// Size is the physical page dimensions in PDF points.
	Size Size
	// Margins is the spacing between the page edge and the content area.
	Margins Edges
	// Header contains elements rendered at the top of every page.
	// The header is laid out with unlimited height to measure it, then
	// subtracted from the body area.
	Header []Element
	// Footer contains elements rendered at the bottom of every page.
	Footer []Element
	// Content contains the main body elements to paginate.
	Content []Element
}

// PageLayout holds all positioned blocks for a single rendered page,
// ready to be handed to the renderer.
type PageLayout struct {
	// Size is the physical page dimensions in PDF points.
	Size Size
	// Blocks contains all content blocks for this page, with coordinates
	// measured from the top-left corner of the page.
	Blocks []Block
}

// Paginator runs the layout engine across a sequence of PageDefs and
// produces a slice of PageLayout values — one per physical output page.
// Content that does not fit on a page overflows automatically to the next page.
type Paginator struct {
	// Fonts is the font resolver used by all layout operations.
	// If nil, MockFontResolver is used (suitable for tests).
	Fonts FontResolver
}

// Paginate processes the given page definitions and returns one PageLayout
// per physical output page. The algorithm:
//  1. For each PageDef, resolve margins.
//  2. Layout header and footer at unlimited height to measure their heights.
//  3. Compute bodyHeight = pageHeight - margins - headerH - footerH.
//  4. Iterate content elements, calling PlanLayout on each.
//  5. Full → accumulate blocks, advance cursor.
//  6. Partial → flush page, push Overflow to next page.
//  7. Nothing at page top → force layout with unlimited height (oversized element).
//  8. Nothing mid-page → flush page, retry on fresh page.
//  9. After all pages are generated, resolve page number placeholders.
func (p *Paginator) Paginate(pages []*PageDef) []PageLayout {
	fonts := p.Fonts
	if fonts == nil {
		fonts = &MockFontResolver{}
	}

	var result []PageLayout

	for _, def := range pages {
		pageSize := def.Size
		if pageSize.Width <= 0 || pageSize.Height <= 0 {
			pageSize = PageA4
		}

		resolvedMargins := def.Margins.Resolve(pageSize.Width, pageSize.Height, 12)

		contentWidth := pageSize.Width - resolvedMargins.Horizontal()
		contentHeight := pageSize.Height - resolvedMargins.Vertical()
		if contentWidth < 0 {
			contentWidth = 0
		}
		if contentHeight < 0 {
			contentHeight = 0
		}

		// Measure header and footer once per page definition.
		headerBlocks, headerHeight := measureSection(def.Header, contentWidth, fonts)
		footerBlocks, footerHeight := measureSection(def.Footer, contentWidth, fonts)

		bodyHeight := contentHeight - headerHeight - footerHeight
		if bodyHeight < 0 {
			bodyHeight = 0
		}

		// State for paginating the content elements.
		remaining := make([]Element, len(def.Content))
		copy(remaining, def.Content)

		for len(remaining) > 0 {
			var pageBlocks []Block
			cursorY := 0.0
			var nextRemaining []Element

			for len(remaining) > 0 {
				elem := remaining[0]
				availH := bodyHeight - cursorY

				plan := elem.PlanLayout(Area{Width: contentWidth, Height: availH})

				switch plan.Status {
				case Full:
					placed := cloneBlocks(plan.Blocks)
					offsetBlocks(placed, 0, cursorY)
					pageBlocks = append(pageBlocks, placed...)
					cursorY += plan.Consumed
					remaining = remaining[1:]

				case Partial:
					placed := cloneBlocks(plan.Blocks)
					offsetBlocks(placed, 0, cursorY)
					pageBlocks = append(pageBlocks, placed...)
					cursorY += plan.Consumed
					// Overflow goes to the next page; include remaining elements after it.
					nextRemaining = make([]Element, 0, 1+len(remaining[1:]))
					nextRemaining = append(nextRemaining, plan.Overflow)
					nextRemaining = append(nextRemaining, remaining[1:]...)
					remaining = nil // flush page

				case Nothing:
					if cursorY == 0 {
						// Nothing at page top — force with unlimited height.
						forcedPlan := elem.PlanLayout(Area{Width: contentWidth, Height: 1e9})
						placed := cloneBlocks(forcedPlan.Blocks)
						offsetBlocks(placed, 0, cursorY)
						pageBlocks = append(pageBlocks, placed...)
						cursorY += forcedPlan.Consumed
						remaining = remaining[1:]
						if forcedPlan.Overflow != nil {
							nextRemaining = make([]Element, 0, 1+len(remaining))
							nextRemaining = append(nextRemaining, forcedPlan.Overflow)
							nextRemaining = append(nextRemaining, remaining...)
							remaining = nil
						}
					} else {
						// Nothing mid-page — flush page, retry this element on next page.
						nextRemaining = remaining
						remaining = nil
					}
				}
			}

			// Compose the full page by offsetting body, header, and footer by margins.
			mx := resolvedMargins.Left
			my := resolvedMargins.Top

			var allBlocks []Block

			// Header at top of content area.
			if len(headerBlocks) > 0 {
				hb := cloneBlocks(headerBlocks)
				offsetBlocks(hb, mx, my)
				allBlocks = append(allBlocks, hb...)
			}

			// Body below header.
			bodyOffY := my + headerHeight
			if len(pageBlocks) > 0 {
				bb := cloneBlocks(pageBlocks)
				offsetBlocks(bb, mx, bodyOffY)
				allBlocks = append(allBlocks, bb...)
			}

			// Footer at bottom of content area.
			if len(footerBlocks) > 0 {
				fb := cloneBlocks(footerBlocks)
				footerY := my + contentHeight - footerHeight
				offsetBlocks(fb, mx, footerY)
				allBlocks = append(allBlocks, fb...)
			}

			result = append(result, PageLayout{
				Size:   pageSize,
				Blocks: allBlocks,
			})

			remaining = nextRemaining
		}
	}

	// Two-pass page number resolution.
	ResolvePageNumbers(result)

	return result
}

// measureSection lays out a slice of elements at unlimited height and returns
// their blocks and total consumed height. This is used to measure header and
// footer sections.
func measureSection(elements []Element, width float64, fonts FontResolver) ([]Block, float64) {
	if len(elements) == 0 {
		return nil, 0
	}
	container := &Box{
		Children: elements,
	}
	plan := container.PlanLayout(Area{Width: width, Height: 1e9})
	return plan.Blocks, plan.Consumed
}

// ResolvePageNumbers performs a second pass over all pages, replacing
// PageNumberPlaceholder and TotalPagesPlaceholder strings inside Draw
// closures with the actual page number and total page count.
//
// Because Draw closures are opaque functions, page number injection is
// implemented by wrapping Draw closures in Text blocks that contain the
// placeholder strings. The paginator substitutes those strings when the
// Text element is created with PageNumberPlaceholder or TotalPagesPlaceholder
// as content.
//
// This function walks all blocks recursively and replaces placeholder content
// in any TextBlock created by the text layout.
func ResolvePageNumbers(pages []PageLayout) {
	total := len(pages)
	for i := range pages {
		resolveBlockPageNumbers(pages[i].Blocks, i+1, total)
	}
}

// resolveBlockPageNumbers recursively finds page-number blocks and updates
// the shared textPtr that their Draw closures read from.
func resolveBlockPageNumbers(blocks []Block, pageNum, totalPages int) {
	for i := range blocks {
		if blocks[i].Tag == "__pagenumber__" && blocks[i].AltText != "" {
			resolved := blocks[i].AltText
			resolved = strings.ReplaceAll(resolved, PageNumberPlaceholder, fmt.Sprintf("%d", pageNum))
			resolved = strings.ReplaceAll(resolved, TotalPagesPlaceholder, fmt.Sprintf("%d", totalPages))
			blocks[i].AltText = resolved

			// Replace the Draw closure with one that uses the resolved text.
			// The old closure captured styling info — we rebuild it here.
			if blocks[i].drawData != nil {
				dd := blocks[i].drawData
				// Draw at (0,0) — the block's X/Y handles positioning.
				// Update block.X to the new alignment offset for the resolved text.
				w := dd.fonts.MeasureString(dd.font, resolved, dd.size)
				rx, _ := computeTextX(dd.align, w, dd.areaWidth, resolved, true)
				blocks[i].X = rx
				blocks[i].Width = w
				blocks[i].Draw = func(r Renderer) {
					r.DrawText(resolved, 0, 0, dd.font, dd.size, dd.color, TextDrawOptions{})
				}
			}
		}
		if len(blocks[i].Children) > 0 {
			resolveBlockPageNumbers(blocks[i].Children, pageNum, totalPages)
		}
	}
}

// PageNumber is a special Element that renders the current page number.
// It uses a two-pass approach: on the first pass it inserts a placeholder
// that the paginator replaces with the actual page number after all pages
// are known.
type PageNumber struct {
	// Format is the format string. Use PageNumberPlaceholder and
	// TotalPagesPlaceholder as substitution markers.
	// Example: PageNumberPlaceholder + " / " + TotalPagesPlaceholder
	Format string
	// Style controls the appearance of the page number text.
	Style Style
	// Fonts is the font resolver.
	Fonts FontResolver
}

// pageNumberDrawData holds the styling data needed to rebuild a page number
// Draw closure after ResolvePageNumbers replaces placeholder text.
type pageNumberDrawData struct {
	fonts     FontResolver
	font      FontRef
	size      float64
	color     Color
	align     Align
	areaWidth float64
}

// PlanLayout implements Element. It creates blocks tagged with "__pagenumber__"
// whose Draw closures are rebuilt by ResolvePageNumbers after pagination.
func (pn *PageNumber) PlanLayout(area Area) Plan {
	s := pn.Style.effective()
	fontSize := s.FontSize
	lineSpacing := fontSize * s.LineHeight

	resolver := pn.Fonts
	if resolver == nil {
		resolver = &MockFontResolver{}
	}
	font := s.Font

	lineWidth := resolver.MeasureString(font, pn.Format, fontSize)
	xPos, _ := computeTextX(s.TextAlign, lineWidth, area.Width, pn.Format, true)

	dd := &pageNumberDrawData{
		fonts:     resolver,
		font:      font,
		size:      fontSize,
		color:     s.Color,
		align:     s.TextAlign,
		areaWidth: area.Width,
	}

	// Initial Draw uses placeholder text; ResolvePageNumbers replaces it.
	capturedFormat := pn.Format
	block := Block{
		X:        xPos,
		Y:        0,
		Width:    lineWidth,
		Height:   lineSpacing,
		Tag:      "__pagenumber__",
		AltText:  pn.Format,
		drawData: dd,
		Draw: func(r Renderer) {
			r.DrawText(capturedFormat, 0, 0, dd.font, dd.size, dd.color, TextDrawOptions{})
		},
	}

	return Plan{
		Status:   Full,
		Consumed: lineSpacing,
		Blocks:   []Block{block},
	}
}
