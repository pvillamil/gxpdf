package builder

import (
	"github.com/coregx/gxpdf/layout"
)

// RowBuilder constructs a single row in the 12-column grid. Columns are added
// via Col(span, fn), where span is a value from 1 to 12. All column spans in a
// row should sum to 12 for a full-width row; partial sums are valid and leave
// the remainder empty.
type RowBuilder struct {
	// b is the owning Builder — propagated to column containers.
	b *Builder
	// columns holds the column definitions in order.
	columns []columnEntry
}

// columnEntry pairs a column span with its content callback.
type columnEntry struct {
	// span is the number of grid columns this entry occupies (1-12).
	span int
	// fn is the callback that populates the column's Container.
	fn func(*ColBuilder)
}

// Col adds a column to the row.
//
// span must be between 1 and 12 inclusive. Values outside this range are
// clamped. The fn callback receives a ColBuilder (which embeds Container)
// for adding text, images, nested rows, etc.
//
// Example:
//
//	r.Col(8, func(c *builder.ColBuilder) { c.Text("Wide column") })
//	r.Col(4, func(c *builder.ColBuilder) { c.Text("Narrow column") })
func (r *RowBuilder) Col(span int, fn func(*ColBuilder)) {
	if span < 1 {
		span = 1
	}
	if span > 12 {
		span = 12
	}
	r.columns = append(r.columns, columnEntry{span: span, fn: fn})
}

// build converts the RowBuilder into a layout.Element (a horizontal Box).
// Each column becomes a child Box with its width set as a fraction of the
// available width proportional to its span.
func (r *RowBuilder) build(cfg rowConfig) layout.Element {
	// Calculate total span for proportional sizing.
	totalSpan := 0
	for _, col := range r.columns {
		totalSpan += col.span
	}
	if totalSpan == 0 {
		totalSpan = 12
	}

	children := make([]layout.Element, 0, len(r.columns))
	for _, col := range r.columns {
		// Build column content.
		cb := &ColBuilder{Container: Container{b: r.b}}
		col.fn(cb)

		// Width as percentage of available space.
		pct := float64(col.span) / float64(totalSpan) * 100.0

		colStyle := layout.Style{}
		colBox := &layout.Box{
			Children:  cb.elements,
			Direction: layout.Vertical,
			Style:     colStyle,
			Width:     layout.Pct(pct),
		}
		children = append(children, colBox)
	}

	rowStyle := layout.Style{}
	if cfg.bgColor != nil {
		rowStyle.Background = cfg.bgColor
	}
	if cfg.padding != nil {
		p := *cfg.padding
		rowStyle.Padding = layout.Edges{Top: p, Right: p, Bottom: p, Left: p}
	}

	// Apply explicit height if configured.
	rowBox := &layout.Box{
		Children:  children,
		Direction: layout.Horizontal,
		Style:     rowStyle,
	}
	if cfg.height != nil {
		rowBox.Height = *cfg.height
	}

	return rowBox
}
