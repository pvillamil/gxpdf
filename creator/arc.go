package creator

import (
	"errors"
)

// ArcOptions configures arc drawing.
//
// An arc is a portion of an ellipse. For circular arcs, set rx = ry.
// The arc can be stroked only, or filled as a wedge (pie-slice) or chord.
type ArcOptions struct {
	// StrokeColor is the border color (nil = no stroke).
	// If StrokeColorCMYK is set, this field is ignored.
	StrokeColor *Color

	// StrokeColorCMYK is the border color in CMYK (nil = no stroke).
	// If set, this takes precedence over StrokeColor (RGB).
	StrokeColorCMYK *ColorCMYK

	// StrokeWidth is the border width in points (default: 1.0).
	StrokeWidth float64

	// FillColor is the fill color (nil = no fill).
	// Mutually exclusive with FillGradient and FillColorCMYK.
	// If FillColorCMYK is set, this field is ignored.
	FillColor *Color

	// FillColorCMYK is the fill color in CMYK (nil = no fill).
	// If set, this takes precedence over FillColor (RGB).
	// Mutually exclusive with FillGradient.
	FillColorCMYK *ColorCMYK

	// FillGradient is the gradient fill (nil = no gradient fill).
	// Mutually exclusive with FillColor and FillColorCMYK.
	FillGradient *Gradient

	// Opacity is the arc opacity (0.0 = transparent, 1.0 = opaque).
	// Optional. If set, applies transparency via ExtGState.
	// Affects both fill and stroke.
	// Range: [0.0, 1.0]
	Opacity *float64

	// Dashed enables dashed border rendering.
	Dashed bool

	// DashArray defines the dash pattern for the border.
	// Only used when Dashed is true.
	DashArray []float64

	// DashPhase is the starting offset into the dash pattern.
	// Only used when Dashed is true.
	DashPhase float64

	// Wedge controls fill mode when FillColor, FillColorCMYK, or FillGradient is set.
	// true (default) = pie-slice: lines are drawn from arc endpoints to center.
	// false = chord: a straight line connects the arc endpoints.
	// Has no effect when there is no fill.
	Wedge *bool
}

// DrawArc draws an elliptical arc centered at (cx, cy) with radii (rx, ry).
//
// startAngle is the angle in degrees from the positive X axis to the start of
// the arc. sweepAngle is the angular extent in degrees. Both angles follow the
// counter-clockwise (CCW) convention used by the PDF coordinate system (Y-up).
//
// A positive sweepAngle sweeps counter-clockwise; a negative value sweeps
// clockwise. A sweep of 360° or more draws a full ellipse.
//
// For a circular arc, set rx = ry = radius, or use [Page.DrawCircularArc].
//
// Parameters:
//   - cx, cy: Center coordinates
//   - rx: Horizontal radius (must be > 0)
//   - ry: Vertical radius (must be > 0)
//   - startAngle: Starting angle in degrees (0 = right / 3 o'clock)
//   - sweepAngle: Angular extent in degrees (positive = CCW)
//   - opts: Arc drawing options
//
// Example:
//
//	// Draw a 90° arc (quarter circle) with stroke only
//	opts := &creator.ArcOptions{
//	    StrokeColor: &creator.Black,
//	    StrokeWidth: 2.0,
//	}
//	err := page.DrawArc(200, 400, 80, 80, 0, 90, opts)
func (p *Page) DrawArc(cx, cy, rx, ry, startAngle, sweepAngle float64, opts *ArcOptions) error {
	if opts == nil {
		return errors.New("arc options cannot be nil")
	}
	if rx <= 0 {
		return errors.New("horizontal radius must be positive")
	}
	if ry <= 0 {
		return errors.New("vertical radius must be positive")
	}
	if err := validateArcOptions(opts); err != nil {
		return err
	}

	p.graphicsOps = append(p.graphicsOps, GraphicsOperation{
		Type:       GraphicsOpArc,
		X:          cx,
		Y:          cy,
		RX:         rx,
		RY:         ry,
		StartAngle: startAngle,
		SweepAngle: sweepAngle,
		ArcOpts:    opts,
	})

	return nil
}

// DrawCircularArc is a convenience wrapper around DrawArc for circular arcs
// where rx = ry = radius.
//
// Parameters:
//   - cx, cy: Center coordinates
//   - radius: Arc radius (must be > 0)
//   - startAngle: Starting angle in degrees (0 = right / 3 o'clock)
//   - sweepAngle: Angular extent in degrees (positive = CCW)
//   - opts: Arc drawing options
//
// Example:
//
//	// Filled pie-slice (wedge)
//	opts := &creator.ArcOptions{
//	    FillColor:   &creator.LightBlue,
//	    StrokeColor: &creator.Black,
//	    StrokeWidth: 1.0,
//	}
//	err := page.DrawCircularArc(300, 400, 60, 30, 120, opts)
func (p *Page) DrawCircularArc(cx, cy, radius, startAngle, sweepAngle float64, opts *ArcOptions) error {
	return p.DrawArc(cx, cy, radius, radius, startAngle, sweepAngle, opts)
}

// validateArcOptions validates arc drawing options.
func validateArcOptions(opts *ArcOptions) error {
	if opts.StrokeColor != nil {
		if err := validateColor(*opts.StrokeColor); err != nil {
			return errors.New("stroke " + err.Error())
		}
	}
	if opts.FillColor != nil {
		if err := validateColor(*opts.FillColor); err != nil {
			return errors.New("fill " + err.Error())
		}
	}
	if opts.StrokeWidth < 0 {
		return errors.New("stroke width must be non-negative")
	}

	hasFill := opts.FillColor != nil || opts.FillColorCMYK != nil || opts.FillGradient != nil
	if opts.StrokeColor == nil && opts.StrokeColorCMYK == nil && !hasFill {
		return errors.New("arc must have at least stroke, fill color, or gradient")
	}
	if opts.FillColor != nil && opts.FillGradient != nil {
		return errors.New("cannot use both fill color and fill gradient")
	}
	if opts.FillGradient != nil {
		if err := opts.FillGradient.Validate(); err != nil {
			return errors.New("fill gradient: " + err.Error())
		}
	}
	if err := validateOpacity(opts.Opacity); err != nil {
		return err
	}
	return nil
}
