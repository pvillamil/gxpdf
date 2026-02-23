package creator

import (
	"errors"
)

// BezierSegment represents a cubic Bézier curve segment.
//
// A cubic Bézier curve is defined by:
//   - Start point (P0)
//   - First control point (C1)
//   - Second control point (C2)
//   - End point (P1)
//
// The curve starts at P0, is pulled toward C1 and C2, and ends at P1.
type BezierSegment struct {
	// Start is the starting point (P0).
	// For the first segment, this is the curve's starting point.
	// For subsequent segments, this should match the previous segment's End point.
	Start Point

	// C1 is the first control point.
	C1 Point

	// C2 is the second control point.
	C2 Point

	// End is the ending point (P1).
	End Point
}

// BezierOptions configures Bézier curve drawing.
type BezierOptions struct {
	// Color is the curve color (RGB, 0.0 to 1.0 range).
	// If ColorCMYK is set, this field is ignored.
	Color Color

	// ColorCMYK is the curve color in CMYK color space (optional).
	// If set, this takes precedence over Color (RGB).
	ColorCMYK *ColorCMYK

	// Width is the curve width in points (default: 1.0).
	Width float64

	// Dashed enables dashed curve rendering.
	Dashed bool

	// DashArray defines the dash pattern (e.g., [3, 1] for "3 on, 1 off").
	// Only used when Dashed is true.
	DashArray []float64

	// DashPhase is the starting offset into the dash pattern.
	// Only used when Dashed is true.
	DashPhase float64

	// Closed determines if the curve path should be closed.
	// If true, a line is drawn from the last segment's end point
	// back to the first segment's start point.
	Closed bool

	// FillColor is the fill color for closed curves (nil = no fill).
	// Only used when Closed is true.
	// Mutually exclusive with FillGradient.
	FillColor *Color

	// FillGradient is the gradient fill for closed curves (nil = no gradient fill).
	// Only used when Closed is true.
	// Mutually exclusive with FillColor.
	FillGradient *Gradient

	// Opacity is the bezier curve opacity (0.0 = transparent, 1.0 = opaque).
	// Optional. If set, applies transparency via ExtGState.
	// Affects both stroke and fill (if Closed is true).
	// Range: [0.0, 1.0]
	Opacity *float64
}

// DrawBezierCurve draws a complex curve composed of one or more cubic Bézier segments.
//
// A Bézier curve provides smooth, flowing curves that are widely used in vector graphics.
// Multiple segments can be connected to create complex paths.
//
// Parameters:
//   - segments: Array of Bézier curve segments (minimum 1 segment)
//   - opts: Curve options (color, width, dash pattern, closed path)
//
// Example (simple S-curve):
//
//	opts := &creator.BezierOptions{
//	    Color: creator.Blue,
//	    Width: 2.0,
//	}
//	segments := []creator.BezierSegment{
//	    {
//	        Start: creator.Point{X: 100, Y: 100},
//	        C1:    creator.Point{X: 150, Y: 200},
//	        C2:    creator.Point{X: 200, Y: 200},
//	        End:   creator.Point{X: 250, Y: 100},
//	    },
//	}
//	err := page.DrawBezierCurve(segments, opts)
//
// Example (closed filled shape):
//
//	opts := &creator.BezierOptions{
//	    Color:     creator.Black,
//	    Width:     1.0,
//	    Closed:    true,
//	    FillColor: &creator.Yellow,
//	}
//	segments := []creator.BezierSegment{
//	    // ... multiple segments forming a closed shape
//	}
//	err := page.DrawBezierCurve(segments, opts)
func (p *Page) DrawBezierCurve(segments []BezierSegment, opts *BezierOptions) error {
	if opts == nil {
		return errors.New("bezier curve options cannot be nil")
	}

	// Validate segments
	if len(segments) == 0 {
		return errors.New("bezier curve must have at least 1 segment")
	}

	// Validate segment continuity (each segment's start should match previous segment's end)
	for i := 1; i < len(segments); i++ {
		prev := segments[i-1].End
		curr := segments[i].Start
		// Allow small floating-point differences
		const epsilon = 0.001
		if abs(prev.X-curr.X) > epsilon || abs(prev.Y-curr.Y) > epsilon {
			return errors.New("bezier segments must be continuous (segment start point must match previous segment end point)")
		}
	}

	// Validate options
	if err := validateBezierOptions(opts); err != nil {
		return err
	}

	// Store graphics operation
	p.graphicsOps = append(p.graphicsOps, GraphicsOperation{
		Type:       GraphicsOpBezier,
		BezierSegs: segments,
		BezierOpts: opts,
	})

	return nil
}

// QuadBezierSegment represents a quadratic Bézier curve segment.
//
// A quadratic Bézier curve is defined by three points:
//   - Start point (P0): the curve's starting anchor
//   - Control point (P1): the single control point that shapes the curve
//   - End point (P2): the curve's ending anchor
//
// Quadratic Bézier curves are the native curve primitive of TrueType font
// outlines and the SVG "Q" command. They offer one fewer degree of freedom
// than cubic Bézier curves and are therefore simpler to parameterize.
//
// Because the PDF specification (ISO 32000) only supports cubic Bézier curves
// (the "c" operator), GxPDF converts each quadratic segment to an equivalent
// cubic segment using exact degree elevation via [QuadBezierSegment.ToCubic].
// The conversion is mathematically exact — it is not an approximation.
type QuadBezierSegment struct {
	// Start is the starting anchor point (P0).
	// For the first segment this is the curve's entry point.
	// For subsequent segments this should match the previous segment's End point.
	Start Point

	// Control is the single quadratic control point (P1).
	// The curve does not pass through this point; it is attracted toward it.
	Control Point

	// End is the ending anchor point (P2).
	End Point
}

// ToCubic converts the quadratic Bézier segment to an equivalent cubic
// Bézier segment using degree elevation.
//
// The conversion formula (exact, not an approximation) is:
//
//	Q0 = P0
//	Q1 = P0 + (2/3) * (P1 - P0)
//	Q2 = P2 + (2/3) * (P1 - P2)
//	Q3 = P2
//
// where P0, P1, P2 are the quadratic start, control, and end points
// and Q0..Q3 are the resulting cubic start, first control, second control,
// and end points.
//
// This identity is derived from the de Casteljau algorithm: a degree-n
// polynomial can always be expressed exactly as a degree-(n+1) polynomial
// by splitting its control polygon in the ratio 2:1 from each endpoint.
//
// Reference: Farin, "Curves and Surfaces for CAGD", §5.2 — Degree Elevation.
func (q QuadBezierSegment) ToCubic() BezierSegment {
	// Q1 = P0 + (2/3)(P1 - P0)
	c1 := Point{
		X: q.Start.X + (2.0/3.0)*(q.Control.X-q.Start.X),
		Y: q.Start.Y + (2.0/3.0)*(q.Control.Y-q.Start.Y),
	}
	// Q2 = P2 + (2/3)(P1 - P2)
	c2 := Point{
		X: q.End.X + (2.0/3.0)*(q.Control.X-q.End.X),
		Y: q.End.Y + (2.0/3.0)*(q.Control.Y-q.End.Y),
	}
	return BezierSegment{
		Start: q.Start,
		C1:    c1,
		C2:    c2,
		End:   q.End,
	}
}

// DrawQuadBezierCurve draws a curve composed of one or more quadratic Bézier
// segments.
//
// Quadratic Bézier curves are defined by a single control point per segment
// (compared to two control points for cubic Bézier curves). They are the
// curve primitive used in TrueType font outlines and the SVG "Q" command.
//
// Because the PDF specification (ISO 32000, section 8.5.2) only supports
// cubic Bézier curves via the "c" content-stream operator, each quadratic
// segment is converted to an exact cubic equivalent using degree elevation
// before writing to the PDF. The conversion is lossless — the rendered curve
// is identical to the original quadratic specification.
//
// Multiple segments are connected end-to-end to form a compound path. Each
// segment's Start point must match the previous segment's End point (within a
// floating-point epsilon of 0.001). This mirrors the continuity requirement of
// [DrawBezierCurve].
//
// All [BezierOptions] apply equally to quadratic and cubic curves: stroke
// color, width, dash patterns, closed paths, fill colors, and opacity.
//
// Parameters:
//   - segments: One or more quadratic Bézier segments (minimum 1).
//   - opts: Curve drawing options (must not be nil).
//
// Example (simple quadratic arc):
//
//	opts := &creator.BezierOptions{
//	    Color: creator.Blue,
//	    Width: 2.0,
//	}
//	segments := []creator.QuadBezierSegment{
//	    {
//	        Start:   creator.Point{X: 100, Y: 100},
//	        Control: creator.Point{X: 175, Y: 200},
//	        End:     creator.Point{X: 250, Y: 100},
//	    },
//	}
//	err := page.DrawQuadBezierCurve(segments, opts)
//
// Example (multi-segment wave):
//
//	segments := []creator.QuadBezierSegment{
//	    {Start: creator.Point{X: 50, Y: 100}, Control: creator.Point{X: 100, Y: 150}, End: creator.Point{X: 150, Y: 100}},
//	    {Start: creator.Point{X: 150, Y: 100}, Control: creator.Point{X: 200, Y: 50}, End: creator.Point{X: 250, Y: 100}},
//	}
//	err := page.DrawQuadBezierCurve(segments, opts)
func (p *Page) DrawQuadBezierCurve(segments []QuadBezierSegment, opts *BezierOptions) error {
	if opts == nil {
		return errors.New("bezier curve options cannot be nil")
	}

	// Validate segments
	if len(segments) == 0 {
		return errors.New("bezier curve must have at least 1 segment")
	}

	// Validate segment continuity (each segment's start should match previous segment's end)
	for i := 1; i < len(segments); i++ {
		prev := segments[i-1].End
		curr := segments[i].Start
		const epsilon = 0.001
		if abs(prev.X-curr.X) > epsilon || abs(prev.Y-curr.Y) > epsilon {
			return errors.New("bezier segments must be continuous (segment start point must match previous segment end point)")
		}
	}

	// Validate options
	if err := validateBezierOptions(opts); err != nil {
		return err
	}

	// Convert all quadratic segments to cubic using degree elevation.
	// This is exact (lossless) — no approximation is involved.
	cubicSegs := make([]BezierSegment, len(segments))
	for i, q := range segments {
		cubicSegs[i] = q.ToCubic()
	}

	// Reuse the existing cubic Bézier pipeline entirely.
	p.graphicsOps = append(p.graphicsOps, GraphicsOperation{
		Type:       GraphicsOpBezier,
		BezierSegs: cubicSegs,
		BezierOpts: opts,
	})

	return nil
}

// validateBezierOptions validates Bézier curve drawing options.
func validateBezierOptions(opts *BezierOptions) error {
	// Validate color components
	if err := validateColor(opts.Color); err != nil {
		return err
	}

	// Validate width
	if opts.Width < 0 {
		return errors.New("curve width must be non-negative")
	}

	// Validate fill color if provided
	if opts.FillColor != nil {
		if err := validateColor(*opts.FillColor); err != nil {
			return errors.New("fill " + err.Error())
		}
	}

	// Fill color only makes sense for closed curves
	if opts.FillColor != nil && !opts.Closed {
		return errors.New("fill color requires closed curve (set Closed: true)")
	}

	// FillColor and FillGradient are mutually exclusive
	if opts.FillColor != nil && opts.FillGradient != nil {
		return errors.New("cannot use both fill color and fill gradient")
	}

	// Validate gradient if provided
	if opts.FillGradient != nil {
		if !opts.Closed {
			return errors.New("fill gradient requires closed curve (set Closed: true)")
		}
		if err := opts.FillGradient.Validate(); err != nil {
			return errors.New("fill gradient: " + err.Error())
		}
	}

	// Validate opacity if provided.
	if err := validateOpacity(opts.Opacity); err != nil {
		return err
	}

	return nil
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
