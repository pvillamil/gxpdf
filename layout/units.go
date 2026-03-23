package layout

// Conversion constants between PDF points and other units.
// 1 PDF point = 1/72 inch.
const (
	ptPerMm = 2.834645669  // 1 mm = 2.834645669 pt
	ptPerCm = 28.34645669  // 1 cm = 28.34645669 pt
	ptPerIn = 72.0         // 1 inch = 72 pt
)

// Unit identifies the measurement unit of a Value.
type Unit int

const (
	// UnitPt represents PDF points (1/72 inch). The native unit for PDF.
	UnitPt Unit = iota
	// UnitMm represents millimeters.
	UnitMm
	// UnitCm represents centimeters.
	UnitCm
	// UnitIn represents inches.
	UnitIn
	// UnitPct represents a percentage of the parent dimension.
	UnitPct
	// UnitFr represents a fractional share (for table/grid column sizing).
	UnitFr
	// UnitAuto indicates content-driven sizing.
	UnitAuto
)

// Value is a dimensional value with an associated unit. Use the constructor
// functions (Pt, Mm, Cm, In, Pct, Fr) to create values.
type Value struct {
	// Amount is the numeric quantity in the specified Unit.
	Amount float64
	// Unit specifies how Amount should be interpreted.
	Unit Unit
}

// Pt constructs a Value in PDF points.
func Pt(v float64) Value { return Value{Amount: v, Unit: UnitPt} }

// Mm constructs a Value in millimeters.
func Mm(v float64) Value { return Value{Amount: v, Unit: UnitMm} }

// Cm constructs a Value in centimeters.
func Cm(v float64) Value { return Value{Amount: v, Unit: UnitCm} }

// In constructs a Value in inches.
func In(v float64) Value { return Value{Amount: v, Unit: UnitIn} }

// Pct constructs a Value as a percentage of the parent dimension (0-100).
func Pct(v float64) Value { return Value{Amount: v, Unit: UnitPct} }

// Fr constructs a fractional Value for use in grid/table column definitions.
func Fr(v float64) Value { return Value{Amount: v, Unit: UnitFr} }

// Auto returns the zero Value with UnitAuto, requesting content-driven sizing.
func Auto() Value { return Value{Unit: UnitAuto} }

// IsAuto reports whether the value is an auto-sized value.
func (v Value) IsAuto() bool { return v.Unit == UnitAuto }

// Resolve converts the value to PDF points given the parent dimension and
// font size. UnitPct uses parentSize; UnitFr and UnitAuto return 0 (callers
// must handle those units separately).
func (v Value) Resolve(parentSize, fontSize float64) float64 {
	switch v.Unit {
	case UnitPt:
		return v.Amount
	case UnitMm:
		return v.Amount * ptPerMm
	case UnitCm:
		return v.Amount * ptPerCm
	case UnitIn:
		return v.Amount * ptPerIn
	case UnitPct:
		return v.Amount / 100.0 * parentSize
	case UnitFr, UnitAuto:
		return 0
	default:
		return v.Amount
	}
}

// Edges represents four-sided spacing (margin, padding, etc.) where each
// side is an independent Value.
type Edges struct {
	// Top is the top edge value.
	Top Value
	// Right is the right edge value.
	Right Value
	// Bottom is the bottom edge value.
	Bottom Value
	// Left is the left edge value.
	Left Value
}

// UniformEdges constructs Edges with the same value on all four sides.
func UniformEdges(v Value) Edges {
	return Edges{Top: v, Right: v, Bottom: v, Left: v}
}

// Resolve converts all edges to PDF points given the parent width, parent
// height, and font size. Horizontal edges (Left, Right) are resolved against
// parentW; vertical edges (Top, Bottom) are resolved against parentH.
func (e Edges) Resolve(parentW, parentH, fontSize float64) ResolvedEdges {
	return ResolvedEdges{
		Top:    e.Top.Resolve(parentH, fontSize),
		Right:  e.Right.Resolve(parentW, fontSize),
		Bottom: e.Bottom.Resolve(parentH, fontSize),
		Left:   e.Left.Resolve(parentW, fontSize),
	}
}

// ResolvedEdges holds four edge values already converted to PDF points.
type ResolvedEdges struct {
	// Top is the top edge in points.
	Top float64
	// Right is the right edge in points.
	Right float64
	// Bottom is the bottom edge in points.
	Bottom float64
	// Left is the left edge in points.
	Left float64
}

// Horizontal returns the sum of the Left and Right edges.
func (r ResolvedEdges) Horizontal() float64 { return r.Left + r.Right }

// Vertical returns the sum of the Top and Bottom edges.
func (r ResolvedEdges) Vertical() float64 { return r.Top + r.Bottom }

// Size represents a 2D size with width and height in PDF points.
type Size struct {
	// Width in PDF points.
	Width float64
	// Height in PDF points.
	Height float64
}

// Standard page sizes in PDF points (width × height for portrait orientation).
var (
	// PageA4 is the ISO A4 page size (210mm × 297mm).
	PageA4 = Size{Width: 595.276, Height: 841.890}
	// PageA3 is the ISO A3 page size (297mm × 420mm).
	PageA3 = Size{Width: 841.890, Height: 1190.551}
	// PageLetter is the US Letter page size (8.5in × 11in).
	PageLetter = Size{Width: 612, Height: 792}
	// PageLegal is the US Legal page size (8.5in × 14in).
	PageLegal = Size{Width: 612, Height: 1008}
)
