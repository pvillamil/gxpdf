package layout

// Align specifies horizontal text alignment.
type Align int

const (
	// AlignLeft aligns text to the left edge (default).
	AlignLeft Align = iota
	// AlignCenter centers text within the available width.
	AlignCenter
	// AlignRight aligns text to the right edge.
	AlignRight
	// AlignJustify distributes words evenly across the full width.
	// The last line of a paragraph uses left alignment.
	AlignJustify
)

// VAlign specifies vertical content alignment within a container.
type VAlign int

const (
	// VAlignTop aligns content to the top of the container (default).
	VAlignTop VAlign = iota
	// VAlignMiddle centers content vertically within the container.
	VAlignMiddle
	// VAlignBottom aligns content to the bottom of the container.
	VAlignBottom
)

// Direction controls the stacking direction of a Box container.
type Direction int

const (
	// Vertical stacks children top-to-bottom (default, CSS block layout).
	Vertical Direction = iota
	// Horizontal places children left-to-right (CSS inline/flex row layout).
	Horizontal
)

// BorderSide describes a single border edge.
type BorderSide struct {
	// Width is the border line width in PDF points.
	Width float64
	// Color is the border line color.
	Color Color
}

// BorderEdges groups the four border sides of a box.
type BorderEdges struct {
	// Top is the top border.
	Top BorderSide
	// Right is the right border.
	Right BorderSide
	// Bottom is the bottom border.
	Bottom BorderSide
	// Left is the left border.
	Left BorderSide
}

// widths returns a ResolvedEdges containing only the border widths (not colors).
func (b BorderEdges) widths() ResolvedEdges {
	return ResolvedEdges{
		Top:    b.Top.Width,
		Right:  b.Right.Width,
		Bottom: b.Bottom.Width,
		Left:   b.Left.Width,
	}
}

// Style is the complete set of visual and layout attributes that can be
// applied to any Element. Zero values produce sensible defaults.
type Style struct {
	// --- Typography ---

	// Font identifies the typeface to use for text within this element.
	Font FontRef
	// FontSize is the text size in PDF points. Defaults to 12 when zero.
	FontSize float64
	// Color is the foreground text color.
	Color Color
	// Background is the optional fill color for the element's background.
	// Nil means transparent.
	Background *Color
	// TextAlign controls horizontal text alignment within the element.
	TextAlign Align

	// --- Line and letter spacing ---

	// LineHeight is a multiplier applied to FontSize to compute inter-line
	// spacing. A value of 1.2 (20% leading) is the default.
	LineHeight float64
	// LetterSpacing adds extra spacing between characters in PDF points.
	LetterSpacing float64

	// --- Box model ---

	// Margin is the space outside the border, separating this element from
	// adjacent elements.
	Margin Edges
	// Padding is the space inside the border, between the border and content.
	Padding Edges
	// Border defines the four border sides.
	Border BorderEdges

	// --- Text decoration ---

	// Bold applies bold weight to the font. Overrides Font.Weight.
	Bold bool
	// Italic applies italic style to the font. Overrides Font.Style.
	Italic bool
	// Underline adds an underline decoration to text.
	Underline bool
	// Strikethrough adds a strikethrough decoration to text.
	Strikethrough bool

	// --- Page break control (from QuestPDF) ---

	// KeepWithNext prevents a page break between this element and the next
	// sibling element.
	KeepWithNext bool
	// KeepTogether prevents this element from being split across pages.
	// If the element does not fit on the current page, it is pushed to the
	// next page as a whole. If it does not fit on a fresh page either, it
	// is placed anyway (forced layout).
	KeepTogether bool

	// --- Vertical alignment ---

	// VerticalAlign controls how content is aligned vertically within its
	// container (applies to table cells and columns).
	VerticalAlign VAlign
}

// DefaultStyle returns a Style with sensible defaults applied:
// 12pt Helvetica, black color, 1.2 line height, left alignment.
func DefaultStyle() Style {
	return Style{
		Font:       DefaultFont(),
		FontSize:   12,
		Color:      Black,
		TextAlign:  AlignLeft,
		LineHeight: 1.2,
	}
}

// effective returns a copy of the style with zero-value fields replaced
// by defaults from the DefaultStyle. This allows partial styles (e.g.
// only FontSize set) to be used safely by layout algorithms.
func (s Style) effective() Style {
	d := DefaultStyle()
	if s.Font.Family == "" {
		s.Font = d.Font
	}
	if s.Font.Weight == 0 {
		s.Font.Weight = d.Font.Weight
	}
	if s.FontSize <= 0 {
		s.FontSize = d.FontSize
	}
	if s.LineHeight <= 0 {
		s.LineHeight = d.LineHeight
	}
	// Apply Bold/Italic shortcuts.
	if s.Bold {
		s.Font.Weight = WeightBold
	}
	if s.Italic {
		s.Font.Style = StyleItalic
	}
	return s
}
