package creator

import (
	"math"
	"testing"
)

// quadFloatEq returns true if a and b differ by less than 1e-9.
// This is the precision used for degree-elevation coordinate comparisons.
func quadFloatEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// quadPtEq returns true if two Points are equal within 1e-9.
func quadPtEq(a, b Point) bool {
	return quadFloatEq(a.X, b.X) && quadFloatEq(a.Y, b.Y)
}

// --- ToCubic conversion tests -------------------------------------------------

// TestQuadBezierSegment_ToCubic verifies the degree-elevation formula using
// a simple, hand-calculable case.
//
// Given P0=(0,0), P1=(50,100), P2=(100,0):
//
//	Q0 = (0, 0)
//	Q1 = P0 + 2/3*(P1-P0) = (0,0) + 2/3*(50,100) = (33.333…, 66.666…)
//	Q2 = P2 + 2/3*(P1-P2) = (100,0) + 2/3*(-50,100) = (66.666…, 66.666…)
//	Q3 = (100, 0)
func TestQuadBezierSegment_ToCubic(t *testing.T) {
	q := QuadBezierSegment{
		Start:   Point{X: 0, Y: 0},
		Control: Point{X: 50, Y: 100},
		End:     Point{X: 100, Y: 0},
	}

	cubic := q.ToCubic()

	// Q0 must equal Start
	if !quadPtEq(cubic.Start, q.Start) {
		t.Errorf("cubic.Start: got (%v,%v), want (%v,%v)",
			cubic.Start.X, cubic.Start.Y, q.Start.X, q.Start.Y)
	}

	// Q3 must equal End
	if !quadPtEq(cubic.End, q.End) {
		t.Errorf("cubic.End: got (%v,%v), want (%v,%v)",
			cubic.End.X, cubic.End.Y, q.End.X, q.End.Y)
	}

	// Q1 = P0 + 2/3*(P1-P0)
	wantC1X := 0 + (2.0/3.0)*(50-0)
	wantC1Y := 0 + (2.0/3.0)*(100-0)
	if !quadFloatEq(cubic.C1.X, wantC1X) || !quadFloatEq(cubic.C1.Y, wantC1Y) {
		t.Errorf("cubic.C1: got (%.9f,%.9f), want (%.9f,%.9f)",
			cubic.C1.X, cubic.C1.Y, wantC1X, wantC1Y)
	}

	// Q2 = P2 + 2/3*(P1-P2)
	wantC2X := 100 + (2.0/3.0)*(50-100)
	wantC2Y := 0 + (2.0/3.0)*(100-0)
	if !quadFloatEq(cubic.C2.X, wantC2X) || !quadFloatEq(cubic.C2.Y, wantC2Y) {
		t.Errorf("cubic.C2: got (%.9f,%.9f), want (%.9f,%.9f)",
			cubic.C2.X, cubic.C2.Y, wantC2X, wantC2Y)
	}
}

// TestQuadBezierToCubic_EquivalenceWithKnownValues covers additional known
// configurations to ensure the formula is applied correctly in all quadrants.
func TestQuadBezierToCubic_EquivalenceWithKnownValues(t *testing.T) {
	tests := []struct {
		name    string
		seg     QuadBezierSegment
		wantC1  Point
		wantC2  Point
		wantEnd Point
	}{
		{
			// Straight horizontal line: control point on the midpoint.
			// Degree elevation of a straight line gives a straight line.
			name: "horizontal straight line",
			seg: QuadBezierSegment{
				Start:   Point{X: 0, Y: 0},
				Control: Point{X: 50, Y: 0},
				End:     Point{X: 100, Y: 0},
			},
			wantC1:  Point{X: 100.0 / 3.0, Y: 0},
			wantC2:  Point{X: 200.0 / 3.0, Y: 0},
			wantEnd: Point{X: 100, Y: 0},
		},
		{
			// Control point at the start: the curve degenerates to a
			// line from start to end.
			name: "control point coincides with start",
			seg: QuadBezierSegment{
				Start:   Point{X: 10, Y: 20},
				Control: Point{X: 10, Y: 20},
				End:     Point{X: 110, Y: 120},
			},
			// C1 = Start + 2/3*(Start-Start) = Start
			wantC1: Point{X: 10, Y: 20},
			// C2 = End + 2/3*(Start-End) = End + 2/3*(-100,-100)
			wantC2:  Point{X: 110 + (2.0/3.0)*(10-110), Y: 120 + (2.0/3.0)*(20-120)},
			wantEnd: Point{X: 110, Y: 120},
		},
		{
			// Symmetric arch: P1 directly above the midpoint.
			name: "symmetric arch",
			seg: QuadBezierSegment{
				Start:   Point{X: 0, Y: 0},
				Control: Point{X: 100, Y: 200},
				End:     Point{X: 200, Y: 0},
			},
			wantC1:  Point{X: (2.0 / 3.0) * 100, Y: (2.0 / 3.0) * 200},
			wantC2:  Point{X: 200 - (2.0/3.0)*100, Y: (2.0 / 3.0) * 200},
			wantEnd: Point{X: 200, Y: 0},
		},
		{
			// Negative coordinates.
			name: "negative coordinates",
			seg: QuadBezierSegment{
				Start:   Point{X: -100, Y: -50},
				Control: Point{X: 0, Y: 100},
				End:     Point{X: 100, Y: -50},
			},
			wantC1:  Point{X: -100 + (2.0/3.0)*(0-(-100)), Y: -50 + (2.0/3.0)*(100-(-50))},
			wantC2:  Point{X: 100 + (2.0/3.0)*(0-100), Y: -50 + (2.0/3.0)*(100-(-50))},
			wantEnd: Point{X: 100, Y: -50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cubic := tt.seg.ToCubic()

			if !quadPtEq(cubic.Start, tt.seg.Start) {
				t.Errorf("Start mismatch: got %+v, want %+v", cubic.Start, tt.seg.Start)
			}
			if !quadPtEq(cubic.C1, tt.wantC1) {
				t.Errorf("C1 mismatch: got (%.9f,%.9f), want (%.9f,%.9f)",
					cubic.C1.X, cubic.C1.Y, tt.wantC1.X, tt.wantC1.Y)
			}
			if !quadPtEq(cubic.C2, tt.wantC2) {
				t.Errorf("C2 mismatch: got (%.9f,%.9f), want (%.9f,%.9f)",
					cubic.C2.X, cubic.C2.Y, tt.wantC2.X, tt.wantC2.Y)
			}
			if !quadPtEq(cubic.End, tt.wantEnd) {
				t.Errorf("End mismatch: got %+v, want %+v", cubic.End, tt.wantEnd)
			}
		})
	}
}

// --- DrawQuadBezierCurve validation tests ------------------------------------

// TestDrawQuadBezierCurve_ValidationErrors exercises all error paths in
// DrawQuadBezierCurve to ensure correct messages and guard conditions.
func TestDrawQuadBezierCurve_ValidationErrors(t *testing.T) {
	validSeg := QuadBezierSegment{
		Start:   Point{X: 100, Y: 100},
		Control: Point{X: 175, Y: 200},
		End:     Point{X: 250, Y: 100},
	}

	tests := []struct {
		name       string
		segments   []QuadBezierSegment
		opts       *BezierOptions
		wantErrMsg string
	}{
		{
			name:       "nil options",
			segments:   []QuadBezierSegment{validSeg},
			opts:       nil,
			wantErrMsg: "bezier curve options cannot be nil",
		},
		{
			name:       "empty segments",
			segments:   []QuadBezierSegment{},
			opts:       &BezierOptions{Color: Black, Width: 1.0},
			wantErrMsg: "bezier curve must have at least 1 segment",
		},
		{
			name: "discontinuous segments",
			segments: []QuadBezierSegment{
				{
					Start:   Point{X: 100, Y: 100},
					Control: Point{X: 175, Y: 200},
					End:     Point{X: 250, Y: 100},
				},
				{
					// Start does not match previous End
					Start:   Point{X: 300, Y: 100},
					Control: Point{X: 375, Y: 200},
					End:     Point{X: 450, Y: 100},
				},
			},
			opts:       &BezierOptions{Color: Black, Width: 1.0},
			wantErrMsg: "bezier segments must be continuous (segment start point must match previous segment end point)",
		},
		{
			name:       "negative width",
			segments:   []QuadBezierSegment{validSeg},
			opts:       &BezierOptions{Color: Black, Width: -0.5},
			wantErrMsg: "curve width must be non-negative",
		},
		{
			name:       "invalid stroke color component",
			segments:   []QuadBezierSegment{validSeg},
			opts:       &BezierOptions{Color: Color{R: 2.0, G: 0, B: 0}, Width: 1.0},
			wantErrMsg: "color components must be in range [0.0, 1.0]",
		},
		{
			name:     "fill color without closed path",
			segments: []QuadBezierSegment{validSeg},
			opts: &BezierOptions{
				Color:     Black,
				Width:     1.0,
				FillColor: &Yellow,
			},
			wantErrMsg: "fill color requires closed curve (set Closed: true)",
		},
		{
			name:     "invalid fill color component",
			segments: []QuadBezierSegment{validSeg},
			opts: &BezierOptions{
				Color:     Black,
				Width:     1.0,
				Closed:    true,
				FillColor: &Color{R: 0, G: -0.1, B: 0},
			},
			wantErrMsg: "fill color components must be in range [0.0, 1.0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, err := c.NewPage()
			if err != nil {
				t.Fatalf("failed to create page: %v", err)
			}

			err = page.DrawQuadBezierCurve(tt.segments, tt.opts)
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErrMsg)
			}
			if err.Error() != tt.wantErrMsg {
				t.Errorf("error message: got %q, want %q", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

// --- DrawQuadBezierCurve rendering tests -------------------------------------

// TestDrawQuadBezierCurve_SingleSegment verifies that a single quadratic
// segment is stored as one GraphicsOpBezier operation with one cubic segment.
func TestDrawQuadBezierCurve_SingleSegment(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	seg := QuadBezierSegment{
		Start:   Point{X: 100, Y: 100},
		Control: Point{X: 175, Y: 200},
		End:     Point{X: 250, Y: 100},
	}
	opts := &BezierOptions{
		Color: Blue,
		Width: 2.0,
	}

	if err = page.DrawQuadBezierCurve([]QuadBezierSegment{seg}, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ops := page.GraphicsOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 graphics operation, got %d", len(ops))
	}

	op := ops[0]
	if op.Type != GraphicsOpBezier {
		t.Errorf("operation type: got %d, want GraphicsOpBezier (%d)", op.Type, GraphicsOpBezier)
	}
	if len(op.BezierSegs) != 1 {
		t.Errorf("expected 1 cubic segment (converted from 1 quadratic), got %d", len(op.BezierSegs))
	}

	// The stored segment must be the cubic elevation of the original quadratic.
	wantCubic := seg.ToCubic()
	got := op.BezierSegs[0]
	if !quadPtEq(got.Start, wantCubic.Start) ||
		!quadPtEq(got.C1, wantCubic.C1) ||
		!quadPtEq(got.C2, wantCubic.C2) ||
		!quadPtEq(got.End, wantCubic.End) {
		t.Errorf("stored cubic segment mismatch:\n  got  Start=%v C1=%v C2=%v End=%v\n  want Start=%v C1=%v C2=%v End=%v",
			got.Start, got.C1, got.C2, got.End,
			wantCubic.Start, wantCubic.C1, wantCubic.C2, wantCubic.End)
	}

	// Options pointer must be preserved.
	if op.BezierOpts != opts {
		t.Errorf("BezierOpts pointer not preserved")
	}
}

// TestDrawQuadBezierCurve_MultiSegment verifies that multiple continuous
// quadratic segments are all converted and stored in correct order.
func TestDrawQuadBezierCurve_MultiSegment(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	segs := []QuadBezierSegment{
		{
			Start:   Point{X: 50, Y: 100},
			Control: Point{X: 100, Y: 150},
			End:     Point{X: 150, Y: 100},
		},
		{
			Start:   Point{X: 150, Y: 100},
			Control: Point{X: 200, Y: 50},
			End:     Point{X: 250, Y: 100},
		},
		{
			Start:   Point{X: 250, Y: 100},
			Control: Point{X: 300, Y: 150},
			End:     Point{X: 350, Y: 100},
		},
	}
	opts := &BezierOptions{Color: Red, Width: 1.5}

	if err = page.DrawQuadBezierCurve(segs, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ops := page.GraphicsOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 graphics operation, got %d", len(ops))
	}

	// Must have exactly as many cubic segments as input quadratic segments.
	if len(ops[0].BezierSegs) != len(segs) {
		t.Errorf("expected %d cubic segments, got %d", len(segs), len(ops[0].BezierSegs))
	}

	// Verify each converted segment independently.
	for i, q := range segs {
		want := q.ToCubic()
		got := ops[0].BezierSegs[i]
		if !quadPtEq(got.Start, want.Start) || !quadPtEq(got.C1, want.C1) ||
			!quadPtEq(got.C2, want.C2) || !quadPtEq(got.End, want.End) {
			t.Errorf("segment %d mismatch:\n  got  %+v\n  want %+v", i, got, want)
		}
	}
}

// TestDrawQuadBezierCurve_ContinuityEpsilon verifies that segments whose
// endpoints agree within the floating-point epsilon (0.001) are accepted.
func TestDrawQuadBezierCurve_ContinuityEpsilon(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	segs := []QuadBezierSegment{
		{
			Start:   Point{X: 100, Y: 100},
			Control: Point{X: 150, Y: 160},
			End:     Point{X: 200.0000005, Y: 99.9999995}, // within epsilon
		},
		{
			Start:   Point{X: 200, Y: 100}, // matches End within 0.001
			Control: Point{X: 250, Y: 40},
			End:     Point{X: 300, Y: 100},
		},
	}
	opts := &BezierOptions{Color: Green, Width: 1.0}

	if err = page.DrawQuadBezierCurve(segs, opts); err != nil {
		t.Errorf("expected near-continuous segments to be accepted, got: %v", err)
	}
}

// TestDrawQuadBezierCurve_ClosedWithFill verifies that a closed quadratic
// Bézier curve with fill color is accepted and stored correctly.
func TestDrawQuadBezierCurve_ClosedWithFill(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	// Two segments forming a simple closed teardrop.
	segs := []QuadBezierSegment{
		{
			Start:   Point{X: 150, Y: 50},
			Control: Point{X: 210, Y: 150},
			End:     Point{X: 150, Y: 200},
		},
		{
			Start:   Point{X: 150, Y: 200},
			Control: Point{X: 90, Y: 150},
			End:     Point{X: 150, Y: 50},
		},
	}
	fillColor := Yellow
	opts := &BezierOptions{
		Color:     Black,
		Width:     1.0,
		Closed:    true,
		FillColor: &fillColor,
	}

	if err = page.DrawQuadBezierCurve(segs, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ops := page.GraphicsOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 graphics operation, got %d", len(ops))
	}
	if ops[0].BezierOpts.Closed != true {
		t.Errorf("expected Closed=true on stored operation")
	}
	if ops[0].BezierOpts.FillColor == nil {
		t.Errorf("expected non-nil FillColor on stored operation")
	}
}

// TestDrawQuadBezierCurve_DashedLine verifies that dashed quadratic curves
// are accepted and the dash settings are stored correctly.
func TestDrawQuadBezierCurve_DashedLine(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	segs := []QuadBezierSegment{
		{
			Start:   Point{X: 100, Y: 100},
			Control: Point{X: 175, Y: 200},
			End:     Point{X: 250, Y: 100},
		},
	}
	opts := &BezierOptions{
		Color:     Green,
		Width:     2.0,
		Dashed:    true,
		DashArray: []float64{8, 4},
		DashPhase: 0,
	}

	if err = page.DrawQuadBezierCurve(segs, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ops := page.GraphicsOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 graphics operation, got %d", len(ops))
	}
	stored := ops[0].BezierOpts
	if !stored.Dashed {
		t.Errorf("expected Dashed=true")
	}
	if len(stored.DashArray) != 2 || stored.DashArray[0] != 8 || stored.DashArray[1] != 4 {
		t.Errorf("DashArray mismatch: got %v", stored.DashArray)
	}
}

// TestDrawQuadBezierCurve_WritesValidPDF is an end-to-end test that confirms
// DrawQuadBezierCurve produces a PDF that can be written without error.
func TestDrawQuadBezierCurve_WritesValidPDF(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	segs := []QuadBezierSegment{
		{
			Start:   Point{X: 100, Y: 500},
			Control: Point{X: 200, Y: 650},
			End:     Point{X: 300, Y: 500},
		},
		{
			Start:   Point{X: 300, Y: 500},
			Control: Point{X: 400, Y: 350},
			End:     Point{X: 500, Y: 500},
		},
	}
	opts := &BezierOptions{
		Color: Color{R: 0.2, G: 0.4, B: 0.8},
		Width: 2.5,
	}

	if err = page.DrawQuadBezierCurve(segs, opts); err != nil {
		t.Fatalf("DrawQuadBezierCurve returned error: %v", err)
	}

	// Write to bytes — this exercises the full creator→writer pipeline.
	data, err := c.Bytes()
	if err != nil {
		t.Fatalf("WriteToBytes returned error: %v", err)
	}

	// A minimal valid PDF must be at least a few hundred bytes and start with %PDF.
	if len(data) < 100 {
		t.Errorf("PDF output too small: %d bytes", len(data))
	}
	if string(data[:4]) != "%PDF" {
		t.Errorf("PDF does not start with %%PDF header")
	}
}

// TestDrawQuadBezierCurve_NoDependencyOnWriterLayer confirms that quadratic
// Bézier curves are stored internally as cubic segments and that no new
// writer-layer type or constant is required.
func TestDrawQuadBezierCurve_NoDependencyOnWriterLayer(t *testing.T) {
	q := QuadBezierSegment{
		Start:   Point{X: 0, Y: 0},
		Control: Point{X: 50, Y: 100},
		End:     Point{X: 100, Y: 0},
	}

	// ToCubic must produce a BezierSegment (not a new type).
	var _ BezierSegment = q.ToCubic()

	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	opts := &BezierOptions{Color: Black, Width: 1.0}
	if err = page.DrawQuadBezierCurve([]QuadBezierSegment{q}, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The stored operation must use the existing GraphicsOpBezier constant.
	ops := page.GraphicsOperations()
	if len(ops) != 1 || ops[0].Type != GraphicsOpBezier {
		t.Errorf("expected GraphicsOpBezier operation, got type %d", ops[0].Type)
	}
}
