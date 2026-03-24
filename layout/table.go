package layout

// Table is a layout element that renders a structured grid of cells with
// optional repeating header and footer rows and page-split support.
//
// Column widths are resolved using a four-pass algorithm:
//   - Fixed (UnitPt/Mm/Cm/In): resolved to points directly
//   - Pct (UnitPct): resolved as a percentage of the available table width
//   - Fr (UnitFr): fractional share of remaining space after fixed/pct columns
//   - Auto (UnitAuto): intrinsic content width via the Measurable interface
//
// Table implements Element and Measurable.
type Table struct {
	// Columns defines the width specification for each column.
	// The number of entries here determines the column count.
	Columns []ColumnDef
	// Header rows are placed at the top of every page (repeated on overflow).
	Header []TableRow
	// Body rows form the main content; they split across pages.
	Body []TableRow
	// Footer rows are placed at the bottom of every page.
	Footer []TableRow
	// Style is applied to the whole table (background, border).
	Style Style
	// Fonts is the font resolver used by cell content measurement.
	// If nil, the MockFontResolver approximation is used.
	Fonts FontResolver
}

// ColumnDef describes how a single column's width is determined.
type ColumnDef struct {
	// Width specifies the column width using Value units.
	// UnitAuto means the column sizes itself to its content.
	// UnitFr distributes the remaining space fractionally.
	Width Value
}

// TableRow is a single row in a table section (header, body, or footer).
type TableRow struct {
	// Cells is the ordered list of cells in this row.
	Cells []TableCell
	// Style applies row-level styling (background color).
	Style Style
}

// TableCell is a single cell in a TableRow.
type TableCell struct {
	// Content is the list of elements rendered inside the cell.
	// Cells support any Element (Text, nested Table, Image, etc.).
	Content []Element
	// ColSpan is the number of columns this cell spans. 1 is normal.
	ColSpan int
	// RowSpan is the number of rows this cell spans. 1 is normal.
	RowSpan int
	// Style provides cell-level styling: padding, border, background, vertical align.
	Style Style
}

// effectiveColSpan returns ColSpan clamped to at least 1.
func (c *TableCell) effectiveColSpan() int {
	if c.ColSpan < 1 {
		return 1
	}
	return c.ColSpan
}

// effectiveRowSpan returns RowSpan clamped to at least 1.
func (c *TableCell) effectiveRowSpan() int {
	if c.RowSpan < 1 {
		return 1
	}
	return c.RowSpan
}

// PlanLayout implements Element. It resolves column widths, builds the cell
// grid, measures row heights, places header and footer rows on every page,
// and splits body rows across pages when needed.
func (t *Table) PlanLayout(area Area) Plan {
	numCols := t.numCols()
	if numCols == 0 {
		return Plan{Status: Full}
	}

	colWidths := t.resolveColumnWidths(numCols, area.Width)

	// Measure all sections.
	headerRows := t.measureRows(t.Header, colWidths)
	bodyRows := t.measureRows(t.Body, colWidths)
	footerRows := t.measureRows(t.Footer, colWidths)

	headerHeight := sumHeights(headerRows)
	footerHeight := sumHeights(footerRows)

	// Check whether all content fits on one page.
	bodyHeight := sumHeights(bodyRows)
	totalHeight := headerHeight + bodyHeight + footerHeight

	if totalHeight <= area.Height || area.Height >= 1e8 {
		// Everything fits — produce a single plan.
		blocks := t.buildAllBlocks(headerRows, bodyRows, footerRows, colWidths, area.Width)
		return Plan{
			Status:   Full,
			Consumed: totalHeight,
			Blocks:   blocks,
		}
	}

	// Determine how many body rows fit on this page.
	available := area.Height - headerHeight - footerHeight
	if available <= 0 {
		// Not even a header+footer fits; force-place header only.
		if headerHeight == 0 {
			return Plan{Status: Nothing}
		}
		blocks := t.buildAllBlocks(headerRows, nil, footerRows, colWidths, area.Width)
		overflow := t.overflowTable(0)
		return Plan{
			Status:   Partial,
			Consumed: headerHeight + footerHeight,
			Blocks:   blocks,
			Overflow: overflow,
		}
	}

	// Walk body rows until we exceed available space.
	splitAt := 0
	usedHeight := 0.0
	for i := range bodyRows {
		if usedHeight+bodyRows[i].height > available {
			// This row doesn't fit; split before it.
			break
		}
		usedHeight += bodyRows[i].height
		splitAt = i + 1
	}

	if splitAt == 0 {
		// Nothing from the body fits on this page.
		if len(bodyRows) == 0 {
			// No body rows — full fit with header/footer only.
			blocks := t.buildAllBlocks(headerRows, nil, footerRows, colWidths, area.Width)
			return Plan{
				Status:   Full,
				Consumed: headerHeight + footerHeight,
				Blocks:   blocks,
			}
		}
		// No body row fits on this page.
		return Plan{Status: Nothing}
	}

	// Build plan for what fits.
	fittingBodyRows := bodyRows[:splitAt]
	consumed := headerHeight + sumHeights(fittingBodyRows) + footerHeight
	blocks := t.buildAllBlocks(headerRows, fittingBodyRows, footerRows, colWidths, area.Width)

	overflow := t.overflowTable(splitAt)
	status := Partial
	if splitAt >= len(bodyRows) {
		status = Full
		overflow = nil
	}

	return Plan{
		Status:   status,
		Consumed: consumed,
		Blocks:   blocks,
		Overflow: overflow,
	}
}

// MinWidth implements Measurable. It returns the sum of each column's minimum
// content width (longest unbreakable word) across all cells.
func (t *Table) MinWidth() float64 {
	if len(t.Columns) == 0 {
		return 0
	}
	numCols := t.numCols()
	colMin := make([]float64, numCols)
	allRows := combineRows(t.Header, t.Body, t.Footer)
	measureIntrinsicWidths(allRows, numCols, colMin, nil)
	total := 0.0
	for _, w := range colMin {
		total += w
	}
	return total
}

// MaxWidth implements Measurable. It returns the sum of each column's maximum
// content width (full content without wrapping) across all cells.
func (t *Table) MaxWidth() float64 {
	if len(t.Columns) == 0 {
		return 0
	}
	numCols := t.numCols()
	colMax := make([]float64, numCols)
	allRows := combineRows(t.Header, t.Body, t.Footer)
	measureIntrinsicWidths(allRows, numCols, nil, colMax)
	total := 0.0
	for _, w := range colMax {
		total += w
	}
	return total
}

// --- Internal types ---

// measuredRow is a TableRow with computed height and resolved cell widths.
type measuredRow struct {
	row      TableRow
	height   float64
	cellRecs []cellRecord
}

// cellRecord stores the resolved geometry of a single cell for rendering.
type cellRecord struct {
	cell      TableCell
	colIndex  int     // starting column (for X offset)
	cellWidth float64 // total width across spanned columns
	rowSpan   int     // effective row span
}

// --- numCols ---

// numCols returns the number of columns determined by t.Columns.
// If t.Columns is empty, it scans all rows to find the maximum column count.
func (t *Table) numCols() int {
	if len(t.Columns) > 0 {
		return len(t.Columns)
	}
	max := 0
	all := combineRows(t.Header, t.Body, t.Footer)
	for ri := range all {
		n := 0
		for ci := range all[ri].Cells {
			n += all[ri].Cells[ci].effectiveColSpan()
		}
		if n > max {
			max = n
		}
	}
	return max
}

// combineRows concatenates three row slices without allocating a large backing
// array eagerly — it returns a pre-allocated slice large enough for all rows.
func combineRows(a, b, c []TableRow) []TableRow {
	result := make([]TableRow, 0, len(a)+len(b)+len(c))
	result = append(result, a...)
	result = append(result, b...)
	result = append(result, c...)
	return result
}

// --- Column width resolution ---

// resolveColumnWidths computes the width of every column in points.
//
// Resolution order:
//  1. Fixed units (Pt, Mm, Cm, In) — resolved directly.
//  2. Pct — percentage of available table width.
//  3. Auto — content-driven min/max measurement.
//  4. Fr — fractional share of remaining space after the above.
func (t *Table) resolveColumnWidths(numCols int, tableWidth float64) []float64 {
	widths := make([]float64, numCols)

	// Handle case where no ColumnDef are provided: distribute equally.
	if len(t.Columns) == 0 {
		w := tableWidth / float64(numCols)
		for i := range widths {
			widths[i] = w
		}
		return widths
	}

	usedWidth := 0.0
	totalFr := 0.0

	// First pass: resolve fixed and pct columns.
	for i := 0; i < numCols; i++ {
		col := t.columnAt(i)
		switch col.Width.Unit {
		case UnitPt, UnitMm, UnitCm, UnitIn:
			w := col.Width.Resolve(tableWidth, 12)
			widths[i] = w
			usedWidth += w
		case UnitPct:
			w := col.Width.Resolve(tableWidth, 12)
			widths[i] = w
			usedWidth += w
		case UnitFr:
			totalFr += col.Width.Amount
		case UnitAuto:
			// handled in second pass
		}
	}

	// Second pass: auto columns — measure content.
	allRows := combineRows(t.Header, t.Body, t.Footer)
	autoColMax := make([]float64, numCols)
	autoColMin := make([]float64, numCols)
	measureIntrinsicWidths(allRows, numCols, autoColMin, autoColMax)

	for i := 0; i < numCols; i++ {
		col := t.columnAt(i)
		if col.Width.Unit == UnitAuto {
			remaining := tableWidth - usedWidth
			if remaining < 0 {
				remaining = 0
			}
			w := autoColMax[i]
			if w > remaining {
				w = autoColMin[i]
			}
			widths[i] = w
			usedWidth += w
		}
	}

	// Third pass: Fr columns share remaining space.
	if totalFr > 0 {
		remaining := tableWidth - usedWidth
		if remaining < 0 {
			remaining = 0
		}
		for i := 0; i < numCols; i++ {
			col := t.columnAt(i)
			if col.Width.Unit == UnitFr {
				widths[i] = remaining * (col.Width.Amount / totalFr)
			}
		}
	}

	return widths
}

// columnAt returns the ColumnDef at index i, or an Auto column if out of range.
func (t *Table) columnAt(i int) ColumnDef {
	if i < len(t.Columns) {
		return t.Columns[i]
	}
	return ColumnDef{Width: Auto()}
}

// measureIntrinsicWidths measures min/max widths for Auto column sizing.
// For single-column cells it updates the column arrays directly.
// For colspan cells it distributes any deficit equally.
// Passing nil for colMin or colMax skips updating that slice.
func measureIntrinsicWidths(rows []TableRow, numCols int, colMin, colMax []float64) {
	// First pass: single-column cells only.
	for ri := range rows {
		colIdx := 0
		for ci := range rows[ri].Cells {
			cell := &rows[ri].Cells[ci]
			span := cell.effectiveColSpan()
			if colIdx >= numCols {
				break
			}
			if span == 1 {
				mn, mx := cellIntrinsicWidths(cell)
				if colMin != nil && mn > colMin[colIdx] {
					colMin[colIdx] = mn
				}
				if colMax != nil && mx > colMax[colIdx] {
					colMax[colIdx] = mx
				}
			}
			colIdx += span
		}
	}

	// Second pass: colspan cells — distribute deficit across spanned columns.
	for ri := range rows {
		colIdx := 0
		for ci := range rows[ri].Cells {
			cell := &rows[ri].Cells[ci]
			span := cell.effectiveColSpan()
			if colIdx >= numCols {
				break
			}
			if span > 1 {
				distributeColspanWidth(cell, colIdx, span, numCols, colMin, colMax)
			}
			colIdx += span
		}
	}
}

// distributeColspanWidth distributes any width deficit from a colspan cell
// equally across the spanned columns.
func distributeColspanWidth(cell *TableCell, colIdx, span, numCols int, colMin, colMax []float64) {
	end := colIdx + span
	if end > numCols {
		end = numCols
		span = end - colIdx
	}
	if span <= 0 {
		return
	}
	mn, mx := cellIntrinsicWidths(cell)

	if colMin != nil {
		spanMin := 0.0
		for c := colIdx; c < end; c++ {
			spanMin += colMin[c]
		}
		if mn > spanMin {
			per := (mn - spanMin) / float64(span)
			for c := colIdx; c < end; c++ {
				colMin[c] += per
			}
		}
	}
	if colMax != nil {
		spanMax := 0.0
		for c := colIdx; c < end; c++ {
			spanMax += colMax[c]
		}
		if mx > spanMax {
			per := (mx - spanMax) / float64(span)
			for c := colIdx; c < end; c++ {
				colMax[c] += per
			}
		}
	}
}

// cellIntrinsicWidths returns the (minWidth, maxWidth) for a cell, including
// its padding. For cells with Measurable content, delegates to MinWidth/MaxWidth.
// The font resolver is embedded in each Element at construction time, so it is
// not needed here directly.
func cellIntrinsicWidths(cell *TableCell) (minW, maxW float64) {
	s := cell.Style.effective()
	fontSize := s.FontSize
	pad := s.Padding.Resolve(0, 0, fontSize)
	padH := pad.Left + pad.Right

	for _, elem := range cell.Content {
		if m, ok := elem.(Measurable); ok {
			mn := m.MinWidth()
			mx := m.MaxWidth()
			if mn+padH > minW {
				minW = mn + padH
			}
			if mx+padH > maxW {
				maxW = mx + padH
			}
		}
	}
	if minW == 0 && maxW == 0 {
		minW = padH
		maxW = padH
	}
	return minW, maxW
}

// --- Row measurement ---

// measureRows converts a slice of TableRows into measuredRows with heights
// pre-computed. It uses the colOccupied array to handle rowspan tracking.
func (t *Table) measureRows(rows []TableRow, colWidths []float64) []measuredRow {
	nCols := len(colWidths)
	colOccupied := make([]int, nCols)
	result := make([]measuredRow, 0, len(rows))

	for ri := range rows {
		mr := buildMeasuredRow(&rows[ri], colWidths, nCols, colOccupied)
		mr.height = computeRowHeight(mr.cellRecs)
		result = append(result, mr)
	}

	return result
}

// buildMeasuredRow constructs a measuredRow for a single TableRow, advancing
// the colOccupied tracker to account for rowspan cells from previous rows.
func buildMeasuredRow(row *TableRow, colWidths []float64, nCols int, colOccupied []int) measuredRow {
	mr := measuredRow{row: *row}
	cellIdx := 0
	col := 0

	for col < nCols && cellIdx < len(row.Cells) {
		col = skipOccupiedCols(colOccupied, col, nCols)
		if col >= nCols || cellIdx >= len(row.Cells) {
			break
		}

		cell := row.Cells[cellIdx]
		colspan := cell.effectiveColSpan()
		if col+colspan > nCols {
			colspan = nCols - col
		}
		rowspan := cell.effectiveRowSpan()

		cellWidth := sumColWidths(colWidths, col, col+colspan)
		mr.cellRecs = append(mr.cellRecs, cellRecord{
			cell:      cell,
			colIndex:  col,
			cellWidth: cellWidth,
			rowSpan:   rowspan,
		})

		markRowspanOccupancy(colOccupied, col, col+colspan, rowspan)

		col += colspan
		cellIdx++
	}

	// Drain remaining occupied columns for this row.
	for ; col < nCols; col++ {
		if colOccupied[col] > 0 {
			colOccupied[col]--
		}
	}

	return mr
}

// skipOccupiedCols advances col past any columns still occupied by a rowspan
// from a prior row, decrementing their occupancy counters as it goes.
func skipOccupiedCols(colOccupied []int, col, nCols int) int {
	for col < nCols && colOccupied[col] > 0 {
		colOccupied[col]--
		col++
	}
	return col
}

// sumColWidths returns the total width of columns [start, end).
func sumColWidths(colWidths []float64, start, end int) float64 {
	total := 0.0
	for c := start; c < end; c++ {
		total += colWidths[c]
	}
	return total
}

// markRowspanOccupancy records rowspan occupancy for columns [start, end)
// when rowspan > 1 so that subsequent rows skip those columns.
func markRowspanOccupancy(colOccupied []int, start, end, rowspan int) {
	if rowspan > 1 {
		for c := start; c < end; c++ {
			colOccupied[c] = rowspan - 1
		}
	}
}

// computeRowHeight returns the maximum cell height across all cellRecords,
// dividing rowspan cells proportionally.
func computeRowHeight(cellRecs []cellRecord) float64 {
	maxH := 0.0
	for i := range cellRecs {
		rec := &cellRecs[i]
		h := measureCellHeight(&rec.cell, rec.cellWidth)
		if rec.rowSpan > 1 {
			h /= float64(rec.rowSpan)
		}
		if h > maxH {
			maxH = h
		}
	}
	return maxH
}

// measureCellHeight returns the total height a cell needs including padding.
// Font resolvers are embedded within each Element (e.g. Text.Fonts) at
// construction time, so no FontResolver argument is required here.
func measureCellHeight(cell *TableCell, cellWidth float64) float64 {
	s := cell.Style.effective()
	fontSize := s.FontSize
	pad := s.Padding.Resolve(cellWidth, 0, fontSize)
	padV := pad.Top + pad.Bottom
	innerWidth := cellWidth - pad.Left - pad.Right
	if innerWidth < 0 {
		innerWidth = 0
	}

	if len(cell.Content) == 0 {
		// Empty cell: at least one line height.
		return fontSize*s.LineHeight + padV
	}

	// Measure by running layout on each content element with unlimited height.
	contentH := 0.0
	for _, elem := range cell.Content {
		plan := elem.PlanLayout(Area{Width: innerWidth, Height: 1e9})
		contentH += plan.Consumed
	}
	return contentH + padV
}

// sumHeights sums the heights of measured rows.
func sumHeights(rows []measuredRow) float64 {
	total := 0.0
	for i := range rows {
		total += rows[i].height
	}
	return total
}

// --- Block building ---

// buildAllBlocks converts measured rows into positioned Blocks for rendering.
// It emits header, then body, then footer rows in sequence starting at Y=0.
func (t *Table) buildAllBlocks(header, body, footer []measuredRow, colWidths []float64, tableWidth float64) []Block {
	var blocks []Block
	cursorY := 0.0

	for i := range header {
		rowBlocks := buildRowBlocks(&header[i], colWidths, tableWidth, cursorY)
		blocks = append(blocks, rowBlocks...)
		cursorY += header[i].height
	}
	for i := range body {
		rowBlocks := buildRowBlocks(&body[i], colWidths, tableWidth, cursorY)
		blocks = append(blocks, rowBlocks...)
		cursorY += body[i].height
	}
	for i := range footer {
		rowBlocks := buildRowBlocks(&footer[i], colWidths, tableWidth, cursorY)
		blocks = append(blocks, rowBlocks...)
		cursorY += footer[i].height
	}

	return blocks
}

// buildRowBlocks creates Blocks for a single measured row.
func buildRowBlocks(mr *measuredRow, colWidths []float64, tableWidth, rowY float64) []Block {
	var blocks []Block

	rowH := mr.height
	rowStyle := mr.row.Style.effective()

	// Row background block.
	if rowStyle.Background != nil {
		bg := *rowStyle.Background
		capW := tableWidth
		capH := rowH
		block := Block{
			X:      0,
			Y:      rowY,
			Width:  tableWidth,
			Height: rowH,
			Draw: func(r Renderer) {
				r.DrawRect(0, 0, capW, capH, &bg, nil, 0)
			},
		}
		blocks = append(blocks, block)
	}

	// Compute column offsets.
	nCols := len(colWidths)
	colOffsets := make([]float64, nCols)
	for i := 1; i < nCols; i++ {
		colOffsets[i] = colOffsets[i-1] + colWidths[i-1]
	}

	// Cell blocks.
	for i := range mr.cellRecs {
		cellBlock := buildCellBlock(&mr.cellRecs[i], colOffsets, rowH, rowY)
		blocks = append(blocks, cellBlock)
	}

	return blocks
}

// buildCellBlock creates the Block for a single cell, including background,
// border, and content layout.
func buildCellBlock(rec *cellRecord, colOffsets []float64, rowH, rowY float64) Block {
	s := rec.cell.Style.effective()
	fontSize := s.FontSize

	cellX := colOffsets[rec.colIndex]
	cellW := rec.cellWidth
	cellH := rowH

	pad := s.Padding.Resolve(cellW, cellH, fontSize)
	innerW := cellW - pad.Left - pad.Right
	innerH := cellH - pad.Top - pad.Bottom
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	// Layout cell content to get actual content height (for vertical alignment).
	var contentBlocks []Block
	innerCursorY := 0.0
	for _, elem := range rec.cell.Content {
		plan := elem.PlanLayout(Area{Width: innerW, Height: innerH})
		placed := cloneBlocks(plan.Blocks)
		offsetBlocks(placed, 0, innerCursorY)
		contentBlocks = append(contentBlocks, placed...)
		innerCursorY += plan.Consumed
	}
	contentH := innerCursorY

	// Vertical alignment offset.
	var vOffset float64
	switch s.VerticalAlign {
	case VAlignMiddle:
		vOffset = (innerH - contentH) / 2
	case VAlignBottom:
		vOffset = innerH - contentH
	default: // VAlignTop
		vOffset = 0
	}
	if vOffset < 0 {
		vOffset = 0
	}

	// Offset content blocks by padding + vertical alignment.
	offsetBlocks(contentBlocks, pad.Left, pad.Top+vOffset)

	// Capture values for the Draw closure.
	// Draw at (0,0) relative to block origin — renderer adds block.X/Y.
	capW := cellW
	capH := cellH
	capBg := s.Background
	capBorder := s.Border

	outerBlock := Block{
		X:        cellX,
		Y:        rowY,
		Width:    cellW,
		Height:   cellH,
		Tag:      "TD",
		Children: contentBlocks,
		Draw: func(r Renderer) {
			if capBg != nil {
				bg := *capBg
				r.DrawRect(0, 0, capW, capH, &bg, nil, 0)
			}
			if capBorder.Top.Width > 0 {
				r.DrawLine(0, 0, capW, 0, capBorder.Top.Color, capBorder.Top.Width)
			}
			if capBorder.Right.Width > 0 {
				r.DrawLine(capW, 0, capW, capH, capBorder.Right.Color, capBorder.Right.Width)
			}
			if capBorder.Bottom.Width > 0 {
				r.DrawLine(0, capH, capW, capH, capBorder.Bottom.Color, capBorder.Bottom.Width)
			}
			if capBorder.Left.Width > 0 {
				r.DrawLine(0, 0, 0, capH, capBorder.Left.Color, capBorder.Left.Width)
			}
		},
	}

	return outerBlock
}

// --- Overflow ---

// overflowTable returns a new Table carrying the remaining body rows starting
// at bodyStart, with the same Header, Footer, Columns, Style and Fonts.
func (t *Table) overflowTable(bodyStart int) *Table {
	if bodyStart >= len(t.Body) {
		return nil
	}
	return &Table{
		Columns: t.Columns,
		Header:  t.Header,
		Body:    t.Body[bodyStart:],
		Footer:  t.Footer,
		Style:   t.Style,
		Fonts:   t.Fonts,
	}
}
