package extractor

import "fmt"

// PathVerb identifies a single step in a vector path.
//
// The verb+coords model is compatible with the gogpu/gg PathVerb pattern:
// each verb describes the operation and how many coordinate values follow
// in the flattened Coords slice.
//
//   - VerbMoveTo  — 2 coords: x, y
//   - VerbLineTo  — 2 coords: x, y
//   - VerbCubicTo — 6 coords: c1x, c1y, c2x, c2y, x, y
//   - VerbQuadTo  — 4 coords: cx, cy, x, y
//   - VerbClose   — 0 coords
//
// Reference: PDF 1.7 specification, Section 8.5.2 (Path Construction Operators).
type PathVerb int

const (
	// VerbMoveTo starts a new subpath at (x, y).
	// Consumes 2 coords: x, y.
	VerbMoveTo PathVerb = iota

	// VerbLineTo adds a straight-line segment to (x, y).
	// Consumes 2 coords: x, y.
	VerbLineTo

	// VerbCubicTo adds a cubic Bézier segment.
	// Consumes 6 coords: c1x, c1y, c2x, c2y, x, y.
	// PDF operators: c (full), v (c1=current point), y (c2=endpoint).
	VerbCubicTo

	// VerbQuadTo adds a quadratic Bézier segment.
	// Consumes 4 coords: cx, cy, x, y.
	// Produced by approximating when the PDF source already encodes quads.
	VerbQuadTo

	// VerbClose closes the current subpath with a straight line back to start.
	// Consumes 0 coords.
	// PDF operator: h.
	VerbClose
)

// String returns a human-readable label for the verb.
func (v PathVerb) String() string {
	switch v {
	case VerbMoveTo:
		return "MoveTo"
	case VerbLineTo:
		return "LineTo"
	case VerbCubicTo:
		return "CubicTo"
	case VerbQuadTo:
		return "QuadTo"
	case VerbClose:
		return "Close"
	default:
		return fmt.Sprintf("PathVerb(%d)", int(v))
	}
}

// VerbCoordCount returns the number of coordinate values consumed by the verb.
func VerbCoordCount(v PathVerb) int {
	switch v {
	case VerbMoveTo:
		return 2
	case VerbLineTo:
		return 2
	case VerbCubicTo:
		return 6
	case VerbQuadTo:
		return 4
	case VerbClose:
		return 0
	default:
		return 0
	}
}

// PaintMode describes how a path is painted.
//
// Reference: PDF 1.7 specification, Section 8.5.3 (Path Painting Operators).
type PaintMode int

const (
	// PaintStroke strokes the path outline.
	// PDF operators: S, s.
	PaintStroke PaintMode = iota

	// PaintFill fills the path interior.
	// PDF operators: f, F, f*.
	PaintFill

	// PaintFillStroke fills and strokes the path.
	// PDF operators: B, B*, b, b*.
	PaintFillStroke
)

// String returns a human-readable label for the paint mode.
func (pm PaintMode) String() string {
	switch pm {
	case PaintStroke:
		return "Stroke"
	case PaintFill:
		return "Fill"
	case PaintFillStroke:
		return "FillStroke"
	default:
		return fmt.Sprintf("PaintMode(%d)", int(pm))
	}
}

// VectorPath represents a fully resolved vector path extracted from a PDF content stream.
//
// All coordinates are in page user space (points, origin at bottom-left),
// after the Current Transformation Matrix (CTM) has been applied.
//
// The verb+coords encoding is compact and compatible with gogpu/gg rendering:
//   - Iterate Verbs; for each verb consume VerbCoordCount(verb) values from Coords.
//
// Example — a filled blue rectangle:
//
//	path.Verbs  = [VerbMoveTo, VerbLineTo, VerbLineTo, VerbLineTo, VerbClose]
//	path.Coords = [x0, y0, x1, y1, x2, y2, x3, y3]
//	path.FillColor  = [0, 0, 1, 1]   // blue, fully opaque
//	path.PaintMode  = PaintFill
type VectorPath struct {
	// PageNum is the 0-based page index this path was extracted from.
	PageNum int

	// Verbs is the sequence of path construction verbs.
	// Each verb describes the type of segment and how many Coords it consumes.
	Verbs []PathVerb

	// Coords stores the flattened coordinates for all verbs.
	// Coordinate count per verb: see VerbCoordCount.
	Coords []float64

	// StrokeColor is the stroke color in RGBA (components 0–1).
	// Alpha is always 1.0 for DeviceRGB/DeviceGray/DeviceCMYK sources;
	// it reflects the ExtGState /CA value when set.
	StrokeColor [4]float64

	// FillColor is the fill color in RGBA (components 0–1).
	// Alpha reflects the ExtGState /ca value when set.
	FillColor [4]float64

	// LineWidth is the stroke line width in points.
	LineWidth float64

	// LineCap is the PDF line cap style (0=butt, 1=round, 2=square).
	LineCap int

	// LineJoin is the PDF line join style (0=miter, 1=round, 2=bevel).
	LineJoin int

	// MiterLimit is the miter join limit (default 10.0 per PDF spec).
	MiterLimit float64

	// Opacity is the effective opacity (0=transparent, 1=opaque).
	// Derived from ExtGState /ca for fill or /CA for stroke (whichever is
	// more restrictive). Defaults to 1.0 when no ExtGState is present.
	Opacity float64

	// PaintMode describes whether the path is stroked, filled, or both.
	PaintMode PaintMode
}

// String returns a compact human-readable representation of the vector path.
func (vp *VectorPath) String() string {
	return fmt.Sprintf(
		"VectorPath{page=%d verbs=%d coords=%d paint=%s lw=%.2f opacity=%.2f}",
		vp.PageNum,
		len(vp.Verbs),
		len(vp.Coords),
		vp.PaintMode.String(),
		vp.LineWidth,
		vp.Opacity,
	)
}
