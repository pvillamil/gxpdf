package layout

// RichText is a multi-style inline text element. It lays out a sequence of
// styled fragments as a single paragraph with shared word-wrapping. Fragments
// with different font sizes share a common baseline within each line: smaller
// runs are shifted down so their baseline aligns with the tallest run's
// baseline.
//
// RichText implements both Element and Measurable.
type RichText struct {
	// Fragments is the ordered sequence of inline text segments. Each fragment
	// carries its own style (font, size, color, decoration). Together they form
	// a single paragraph.
	Fragments []RichTextFragment
	// Align controls horizontal alignment for every line in the paragraph.
	Align Align
	// LineHeight is a multiplier applied to the tallest font size on each line
	// to compute the line's vertical extent. Defaults to 1.2 when zero.
	LineHeight float64
	// Fonts is the font resolver used for measurement. When nil, the internal
	// mock approximation (0.5 * fontSize per character) is used.
	Fonts FontResolver
}

// RichTextFragment is one styled segment within a RichText paragraph. Multiple
// fragments are laid out inline; word-wrap boundaries may fall inside or
// between fragments.
type RichTextFragment struct {
	// Text is the content of this fragment.
	Text string
	// Style controls the typographic and decorative properties of this fragment.
	// FontSize, Font, Color, Bold, Italic, Underline, and Strikethrough are all
	// honored at the fragment level.
	Style Style
	// URL, when non-empty, makes the fragment a hyperlink. The renderer will
	// record this as a LinkArea on the containing Block.
	URL string
}

// richRun is a word-level unit produced by fragmentsToRichRuns. Each run
// carries a measured width, the effective resolved style, and a flag that
// distinguishes space runs from word runs.
type richRun struct {
	text     string
	style    Style // effective() already applied
	width    float64
	fontSize float64 // s.FontSize after effective()
	isSpace  bool
	url      string // non-empty for link fragments
}

// PlanLayout implements Element. It splits fragments into word-level runs,
// fills lines greedily, and returns positioned Blocks. Each Block corresponds
// to one wrapped line and contains a Draw closure that calls r.DrawText for
// every run on that line.
//
// If not all lines fit vertically, the remaining content is returned as a new
// RichText in Plan.Overflow.
func (rt *RichText) PlanLayout(area Area) Plan {
	resolver := rt.Fonts
	if resolver == nil {
		resolver = &MockFontResolver{}
	}

	runs := fragmentsToRichRuns(rt.Fragments, resolver)
	if len(runs) == 0 {
		return Plan{Status: Full, Consumed: 0}
	}

	lineHeight := rt.LineHeight
	if lineHeight <= 0 {
		lineHeight = 1.2
	}

	lines := fillRichLines(runs, area.Width)

	var blocks []Block
	cursorY := 0.0

	for i, line := range lines {
		maxFS := maxRichFontSize(line)
		lineSpacing := maxFS * lineHeight

		if cursorY+lineSpacing > area.Height+0.01 && area.Height < 1e8 {
			// Build overflow from remaining lines.
			overflow := rebuildRichOverflow(lines[i:], rt.Align, rt.LineHeight, rt.Fonts)
			status := Partial
			if len(blocks) == 0 {
				status = Nothing
			}
			return Plan{
				Status:   status,
				Consumed: cursorY,
				Blocks:   blocks,
				Overflow: overflow,
			}
		}

		isLastLine := i == len(lines)-1
		block := placeRichLine(line, rt.Align, area.Width, lineSpacing, maxFS, cursorY, isLastLine)
		blocks = append(blocks, block)
		cursorY += lineSpacing
	}

	return Plan{
		Status:   Full,
		Consumed: cursorY,
		Blocks:   blocks,
	}
}

// MinWidth implements Measurable. It returns the width of the widest single
// word across all fragments, which is the minimum width the element can occupy
// without losing content.
func (rt *RichText) MinWidth() float64 {
	resolver := rt.Fonts
	if resolver == nil {
		resolver = &MockFontResolver{}
	}
	max := 0.0
	for _, frag := range rt.Fragments {
		s := frag.Style.effective()
		words := splitRichWords(frag.Text)
		for _, w := range words {
			ww := resolver.MeasureString(s.Font, w, s.FontSize)
			if ww > max {
				max = ww
			}
		}
	}
	return max
}

// MaxWidth implements Measurable. It returns the width of all fragments
// rendered on a single line without wrapping.
func (rt *RichText) MaxWidth() float64 {
	resolver := rt.Fonts
	if resolver == nil {
		resolver = &MockFontResolver{}
	}
	total := 0.0
	for _, frag := range rt.Fragments {
		s := frag.Style.effective()
		total += resolver.MeasureString(s.Font, frag.Text, s.FontSize)
	}
	return total
}

// --- internal helpers ---

// fragmentsToRichRuns converts the fragment list into a flat slice of word-
// and space-level runs. Each run's width is measured via the font resolver.
func fragmentsToRichRuns(fragments []RichTextFragment, resolver FontResolver) []richRun {
	var runs []richRun
	for _, frag := range fragments {
		if frag.Text == "" {
			continue
		}
		s := frag.Style.effective()
		parts := splitIntoWordsAndSpaces(frag.Text)
		for _, part := range parts {
			isSpace := isAllSpaces(part)
			w := resolver.MeasureString(s.Font, part, s.FontSize)
			runs = append(runs, richRun{
				text:     part,
				style:    s,
				width:    w,
				fontSize: s.FontSize,
				isSpace:  isSpace,
				url:      frag.URL,
			})
		}
	}
	return runs
}

// splitIntoWordsAndSpaces splits text into alternating word and space tokens.
// For example "Hello  world" → ["Hello", "  ", "world"].
func splitIntoWordsAndSpaces(text string) []string {
	var parts []string
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		if runes[i] == ' ' {
			j := i
			for j < len(runes) && runes[j] == ' ' {
				j++
			}
			parts = append(parts, string(runes[i:j]))
			i = j
		} else {
			j := i
			for j < len(runes) && runes[j] != ' ' {
				j++
			}
			parts = append(parts, string(runes[i:j]))
			i = j
		}
	}
	return parts
}

// isAllSpaces reports whether s consists entirely of space characters.
func isAllSpaces(s string) bool {
	for _, r := range s {
		if r != ' ' {
			return false
		}
	}
	return len(s) > 0
}

// splitRichWords extracts non-space words from text (for MinWidth measurement).
func splitRichWords(text string) []string {
	var words []string
	for _, part := range splitIntoWordsAndSpaces(text) {
		if !isAllSpaces(part) && part != "" {
			words = append(words, part)
		}
	}
	return words
}

// fillRichLines distributes runs into wrapped lines using a greedy algorithm.
// Space runs at the start of a new line are discarded. Trailing spaces on
// each committed line are also stripped.
func fillRichLines(runs []richRun, availWidth float64) [][]richRun {
	if len(runs) == 0 {
		return nil
	}

	var lines [][]richRun
	var currentLine []richRun
	lineWidth := 0.0

	for _, run := range runs {
		if run.isSpace {
			if len(currentLine) == 0 {
				continue // skip leading spaces on a fresh line
			}
			if lineWidth+run.width <= availWidth {
				currentLine = append(currentLine, run)
				lineWidth += run.width
			} else {
				// Space falls exactly at line boundary — break here.
				lines = append(lines, trimTrailingRichSpaces(currentLine))
				currentLine = nil
				lineWidth = 0
			}
		} else {
			if len(currentLine) == 0 {
				// First word on a line always placed regardless of width.
				currentLine = append(currentLine, run)
				lineWidth = run.width
			} else if lineWidth+run.width <= availWidth {
				currentLine = append(currentLine, run)
				lineWidth += run.width
			} else {
				// Word overflows — commit current line, start fresh.
				lines = append(lines, trimTrailingRichSpaces(currentLine))
				currentLine = []richRun{run}
				lineWidth = run.width
			}
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, trimTrailingRichSpaces(currentLine))
	}

	return lines
}

// trimTrailingRichSpaces removes space runs from the end of a line slice.
func trimTrailingRichSpaces(runs []richRun) []richRun {
	for len(runs) > 0 && runs[len(runs)-1].isSpace {
		runs = runs[:len(runs)-1]
	}
	return runs
}

// maxRichFontSize returns the largest fontSize across all runs in a line.
// Returns 12 if the line is empty or all sizes are zero.
func maxRichFontSize(line []richRun) float64 {
	max := 0.0
	for _, run := range line {
		if run.fontSize > max {
			max = run.fontSize
		}
	}
	if max <= 0 {
		return 12
	}
	return max
}

// placeRichLine builds a single layout Block for one wrapped line of mixed-
// style runs. The block's Draw closure iterates each run and calls
// r.DrawText at the run's computed X offset.
//
// Baseline alignment: all runs share a common baseline. Smaller runs are
// shifted down by (maxFontSize - runFontSize) so their baseline matches that
// of the tallest run in the line.
//
// Half-leading: like the single-style Text element, the vertical offset for
// each run includes halfLeading = (lineSpacing - maxFontSize) / 2 so that
// leading is distributed equally above and below the text.
func placeRichLine(
	line []richRun,
	align Align,
	availWidth float64,
	lineSpacing float64,
	maxFS float64,
	cursorY float64,
	isLastLine bool,
) Block {
	// Measure total content width and count space runs.
	contentWidth := 0.0
	spaceCount := 0
	for _, run := range line {
		contentWidth += run.width
		if run.isSpace {
			spaceCount++
		}
	}

	// Compute X start and extra space-per-gap for justified lines.
	baseX := 0.0
	extraPerGap := 0.0

	switch align {
	case AlignCenter:
		baseX = (availWidth - contentWidth) / 2
	case AlignRight:
		baseX = availWidth - contentWidth
	case AlignJustify:
		if !isLastLine && spaceCount > 0 {
			extraPerGap = (availWidth - contentWidth) / float64(spaceCount)
		}
	}

	// Half-leading distributes (lineSpacing - maxFS) equally above and below.
	halfLeading := (lineSpacing - maxFS) / 2

	// Capture loop variables for the Draw closure.
	type runDraw struct {
		text    string
		x       float64
		yOffset float64
		style   Style
		url     string
	}

	// Pre-compute per-run draw data (X positions and Y offsets).
	drawRuns := make([]runDraw, 0, len(line))
	cursorX := baseX
	for _, run := range line {
		if run.isSpace {
			spaceW := run.width
			if align == AlignJustify && !isLastLine {
				spaceW += extraPerGap
			}
			cursorX += spaceW
			continue
		}
		// Baseline alignment: shift smaller runs down to share the baseline.
		yOffset := halfLeading + (maxFS - run.fontSize)
		if yOffset < 0 {
			yOffset = 0
		}
		drawRuns = append(drawRuns, runDraw{
			text:    run.text,
			x:       cursorX,
			yOffset: yOffset,
			style:   run.style,
			url:     run.url,
		})
		cursorX += run.width
	}

	// Build LinkAreas for any URL-bearing runs.
	var links []LinkArea
	for _, rd := range drawRuns {
		if rd.url != "" {
			// Approximate the run's width by re-deriving from its x span.
			// We record the X range by checking the next run or line end.
			// For simplicity we mark the full line height as the link area.
			links = append(links, LinkArea{
				X:      rd.x,
				Y:      0,
				Width:  availWidth - rd.x, // conservative; refined if needed
				Height: lineSpacing,
				URL:    rd.url,
			})
		}
	}
	// Refine link widths: use next run's X when available.
	for i := range links {
		// Find corresponding drawRun with URL.
		for j, rd := range drawRuns {
			if rd.url != "" && rd.x == links[i].X {
				// Use the gap to the next non-empty run as width.
				if j+1 < len(drawRuns) {
					links[i].Width = drawRuns[j+1].x - rd.x
				} else {
					links[i].Width = cursorX - rd.x
				}
				break
			}
		}
	}

	capturedRuns := drawRuns
	block := Block{
		X:      0,
		Y:      cursorY,
		Width:  availWidth,
		Height: lineSpacing,
		Links:  links,
		Draw: func(r Renderer) {
			for _, rd := range capturedRuns {
				r.DrawText(
					rd.text,
					rd.x,
					rd.yOffset,
					rd.style.Font,
					rd.style.FontSize,
					rd.style.Color,
					TextDrawOptions{
						LetterSpacing: rd.style.LetterSpacing,
						Underline:     rd.style.Underline,
						Strikethrough: rd.style.Strikethrough,
					},
				)
			}
		},
	}
	return block
}

// rebuildRichOverflow reconstructs a RichText element from remaining lines of
// runs, preserving styles for overflow rendering on the next page.
func rebuildRichOverflow(lines [][]richRun, align Align, lineHeight float64, fonts FontResolver) *RichText {
	var fragments []RichTextFragment
	for i, line := range lines {
		if i > 0 {
			// Re-insert a space between lines to restore word separation.
			if len(line) > 0 {
				fragments = append(fragments, RichTextFragment{
					Text:  " ",
					Style: line[0].style,
					URL:   line[0].url,
				})
			}
		}
		for _, run := range line {
			fragments = append(fragments, RichTextFragment{
				Text:  run.text,
				Style: run.style,
				URL:   run.url,
			})
		}
	}
	return &RichText{
		Fragments:  fragments,
		Align:      align,
		LineHeight: lineHeight,
		Fonts:      fonts,
	}
}
