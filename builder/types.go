package builder

import "github.com/coregx/gxpdf/layout"

// Value represents a dimension with a unit.
// Use constructor functions Pt, Mm, Cm, In, Pct, Fr to create values.
type Value struct {
	amount float64
	unit   unit
}

type unit int

const (
	unitPt  unit = iota
	unitMm
	unitCm
	unitIn
	unitPct
	unitFr
	unitAuto
)

// Pt creates a value in PDF points (1/72 inch).
func Pt(v float64) Value { return Value{amount: v, unit: unitPt} }

// Mm creates a value in millimeters.
func Mm(v float64) Value { return Value{amount: v, unit: unitMm} }

// Cm creates a value in centimeters.
func Cm(v float64) Value { return Value{amount: v, unit: unitCm} }

// In creates a value in inches.
func In(v float64) Value { return Value{amount: v, unit: unitIn} }

// Pct creates a percentage value relative to the parent dimension.
func Pct(v float64) Value { return Value{amount: v, unit: unitPct} }

// Fr creates a fractional value for proportional column sizing in tables.
func Fr(v float64) Value { return Value{amount: v, unit: unitFr} }

// Auto returns a value indicating content-driven sizing.
func Auto() Value { return Value{unit: unitAuto} }

// toLayout converts builder.Value to layout.Value.
func (v Value) toLayout() layout.Value {
	switch v.unit {
	case unitPt:
		return layout.Pt(v.amount)
	case unitMm:
		return layout.Mm(v.amount)
	case unitCm:
		return layout.Cm(v.amount)
	case unitIn:
		return layout.In(v.amount)
	case unitPct:
		return layout.Pct(v.amount)
	case unitFr:
		return layout.Fr(v.amount)
	case unitAuto:
		return layout.Auto()
	default:
		return layout.Pt(v.amount)
	}
}

// Size represents page dimensions.
type Size struct {
	Width  float64 // in points
	Height float64 // in points
}

// toLayout converts builder.Size to layout.Size.
func (s Size) toLayout() layout.Size {
	return layout.Size{Width: s.Width, Height: s.Height}
}

// Predefined page sizes.
var (
	A4     = Size{Width: 595.276, Height: 841.890}
	A3     = Size{Width: 841.890, Height: 1190.551}
	Letter = Size{Width: 612, Height: 792}
	Legal  = Size{Width: 612, Height: 1008}
)

// Color represents an RGB color.
type Color struct {
	R, G, B float64 // 0.0 to 1.0
}

// toLayout converts builder.Color to layout.Color.
func (c Color) toLayout() layout.Color {
	return layout.RGB(c.R, c.G, c.B)
}

// RGB creates a color from float components (0.0 to 1.0).
func RGB(r, g, b float64) Color { return Color{R: r, G: g, B: b} }

// RGB255 creates a color from byte components (0 to 255).
func RGB255(r, g, b uint8) Color {
	return Color{
		R: float64(r) / 255.0,
		G: float64(g) / 255.0,
		B: float64(b) / 255.0,
	}
}

// Page number placeholders for use with Container.PageNumber().
// These must match layout.PageNumberPlaceholder / layout.TotalPagesPlaceholder.
const (
	// PageNum is replaced with the current page number after layout.
	PageNum = "\x00PAGE\x00"
	// TotalPages is replaced with the total page count after layout.
	TotalPages = "\x00TOTAL\x00"
)
