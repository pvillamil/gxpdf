package creator

import (
	"bytes"
	"strings"
	"testing"

	"github.com/coregx/gxpdf/internal/writer"
)

func TestDrawArc_StrokeOnly(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	opts := &ArcOptions{
		StrokeColor: &Black,
		StrokeWidth: 2.0,
	}
	err = page.DrawArc(200, 400, 80, 80, 0, 90, opts)
	if err != nil {
		t.Fatalf("DrawArc failed: %v", err)
	}

	ops := page.GraphicsOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	op := ops[0]
	if op.Type != GraphicsOpArc {
		t.Errorf("expected GraphicsOpArc, got %d", op.Type)
	}
	if op.X != 200 || op.Y != 400 {
		t.Errorf("expected center (200,400), got (%f,%f)", op.X, op.Y)
	}
	if op.RX != 80 || op.RY != 80 {
		t.Errorf("expected radii (80,80), got (%f,%f)", op.RX, op.RY)
	}
	if op.StartAngle != 0 || op.SweepAngle != 90 {
		t.Errorf("expected angles (0,90), got (%f,%f)", op.StartAngle, op.SweepAngle)
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}
}

func TestDrawArc_FilledWedge(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	wedge := true
	opts := &ArcOptions{
		StrokeColor: &Black,
		StrokeWidth: 1.0,
		FillColor:   &Blue,
		Wedge:       &wedge,
	}
	err = page.DrawArc(200, 400, 80, 60, 30, 120, opts)
	if err != nil {
		t.Fatalf("DrawArc wedge failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}

	// Verify fill+stroke operator via the content stream (uncompressed).
	gops := convertGraphicsOps(page.GraphicsOperations())
	content, _, err := writer.GenerateContentStreamWithGraphics(nil, gops)
	if err != nil {
		t.Fatalf("GenerateContentStreamWithGraphics failed: %v", err)
	}
	cs := string(content)
	// Wedge with both fill and stroke uses the B (FillAndStroke) operator.
	if !strings.Contains(cs, "B\n") {
		t.Errorf("expected FillAndStroke 'B' operator in content stream, got:\n%s", cs)
	}
}

func TestDrawArc_FilledChord(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	chord := false
	opts := &ArcOptions{
		FillColor: &Red,
		Wedge:     &chord,
	}
	err = page.DrawArc(200, 400, 80, 60, 45, 90, opts)
	if err != nil {
		t.Fatalf("DrawArc chord failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}

	// Verify fill operator via the content stream (uncompressed).
	gops := convertGraphicsOps(page.GraphicsOperations())
	content, _, err := writer.GenerateContentStreamWithGraphics(nil, gops)
	if err != nil {
		t.Fatalf("GenerateContentStreamWithGraphics failed: %v", err)
	}
	cs := string(content)
	// Chord with fill only uses the f (Fill) operator.
	if !strings.Contains(cs, "f\n") {
		t.Errorf("expected Fill 'f' operator in content stream, got:\n%s", cs)
	}
}

func TestDrawArc_CircularArc(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	opts := &ArcOptions{
		StrokeColor: &Black,
		StrokeWidth: 1.5,
	}
	err = page.DrawCircularArc(300, 400, 60, 0, 180, opts)
	if err != nil {
		t.Fatalf("DrawCircularArc failed: %v", err)
	}

	ops := page.GraphicsOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	op := ops[0]
	if op.RX != 60 || op.RY != 60 {
		t.Errorf("expected equal radii 60, got rx=%f ry=%f", op.RX, op.RY)
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}
}

func TestDrawArc_FullCircle(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	opts := &ArcOptions{
		StrokeColor: &Black,
		StrokeWidth: 1.0,
		FillColor:   &Green,
	}
	// 360° sweep should produce a full ellipse.
	err = page.DrawArc(200, 400, 80, 50, 0, 360, opts)
	if err != nil {
		t.Fatalf("DrawArc full circle failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}
}

func TestDrawArc_GradientFill(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	grad := NewLinearGradient(120, 400, 280, 400)
	grad.AddColorStop(0, Red)
	grad.AddColorStop(1, Blue)

	opts := &ArcOptions{
		FillGradient: grad,
		StrokeColor:  &Black,
		StrokeWidth:  1.0,
	}
	err = page.DrawArc(200, 400, 80, 80, 0, 180, opts)
	if err != nil {
		t.Fatalf("DrawArc gradient failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}

	// Verify gradient shading operator via the content stream (uncompressed).
	gops := convertGraphicsOps(page.GraphicsOperations())
	content, _, err := writer.GenerateContentStreamWithGraphics(nil, gops)
	if err != nil {
		t.Fatalf("GenerateContentStreamWithGraphics failed: %v", err)
	}
	cs := string(content)
	// Gradient fill uses the 'sh' shading operator.
	if !strings.Contains(cs, " sh\n") {
		t.Errorf("expected shading operator 'sh' in content stream, got:\n%s", cs)
	}
}

func TestDrawArc_Validation(t *testing.T) {
	tests := []struct {
		name         string
		cx, cy       float64
		rx, ry       float64
		start, sweep float64
		opts         *ArcOptions
		wantErr      bool
		errMsg       string
	}{
		{
			name: "nil options",
			cx:   200, cy: 400, rx: 80, ry: 80, start: 0, sweep: 90,
			opts:    nil,
			wantErr: true,
			errMsg:  "arc options cannot be nil",
		},
		{
			name: "zero horizontal radius",
			cx:   200, cy: 400, rx: 0, ry: 80, start: 0, sweep: 90,
			opts:    &ArcOptions{StrokeColor: &Black},
			wantErr: true,
			errMsg:  "horizontal radius must be positive",
		},
		{
			name: "negative horizontal radius",
			cx:   200, cy: 400, rx: -10, ry: 80, start: 0, sweep: 90,
			opts:    &ArcOptions{StrokeColor: &Black},
			wantErr: true,
			errMsg:  "horizontal radius must be positive",
		},
		{
			name: "zero vertical radius",
			cx:   200, cy: 400, rx: 80, ry: 0, start: 0, sweep: 90,
			opts:    &ArcOptions{StrokeColor: &Black},
			wantErr: true,
			errMsg:  "vertical radius must be positive",
		},
		{
			name: "no stroke or fill",
			cx:   200, cy: 400, rx: 80, ry: 80, start: 0, sweep: 90,
			opts:    &ArcOptions{},
			wantErr: true,
			errMsg:  "arc must have at least stroke, fill color, or gradient",
		},
		{
			name: "both fill color and gradient",
			cx:   200, cy: 400, rx: 80, ry: 80, start: 0, sweep: 90,
			opts: func() *ArcOptions {
				g := NewLinearGradient(0, 0, 100, 0)
				g.AddColorStop(0, Red)
				g.AddColorStop(1, Blue)
				return &ArcOptions{FillColor: &Red, FillGradient: g}
			}(),
			wantErr: true,
			errMsg:  "cannot use both fill color and fill gradient",
		},
		{
			name: "negative stroke width",
			cx:   200, cy: 400, rx: 80, ry: 80, start: 0, sweep: 90,
			opts:    &ArcOptions{StrokeColor: &Black, StrokeWidth: -1},
			wantErr: true,
			errMsg:  "stroke width must be non-negative",
		},
		{
			name: "valid stroke only",
			cx:   200, cy: 400, rx: 80, ry: 80, start: 0, sweep: 90,
			opts:    &ArcOptions{StrokeColor: &Black, StrokeWidth: 1.0},
			wantErr: false,
		},
		{
			name: "valid fill only",
			cx:   200, cy: 400, rx: 80, ry: 80, start: 0, sweep: 90,
			opts:    &ArcOptions{FillColor: &Blue},
			wantErr: false,
		},
		{
			name: "valid negative sweep (clockwise)",
			cx:   200, cy: 400, rx: 80, ry: 80, start: 90, sweep: -45,
			opts:    &ArcOptions{StrokeColor: &Black},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, err := c.NewPage()
			if err != nil {
				t.Fatalf("failed to create page: %v", err)
			}
			err = page.DrawArc(tt.cx, tt.cy, tt.rx, tt.ry, tt.start, tt.sweep, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDrawArc_OpacityField(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	opacity := 0.5
	opts := &ArcOptions{
		StrokeColor: &Black,
		StrokeWidth: 1.0,
		FillColor:   &Blue,
		Opacity:     &opacity,
	}
	err = page.DrawArc(200, 400, 80, 80, 0, 180, opts)
	if err != nil {
		t.Fatalf("DrawArc with opacity failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}
}

func TestDrawArc_WedgeDefaultTrue(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	// When Wedge is nil the conversion sets gop.Wedge = true (pie-slice default).
	opts := &ArcOptions{
		FillColor: &Blue,
	}
	err = page.DrawArc(200, 400, 80, 80, 0, 90, opts)
	if err != nil {
		t.Fatalf("DrawArc failed: %v", err)
	}

	ops := page.GraphicsOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	// ArcOpts.Wedge == nil means default = wedge (pie-slice).
	if ops[0].ArcOpts.Wedge != nil {
		t.Errorf("expected nil Wedge (default true), got %v", ops[0].ArcOpts.Wedge)
	}

	// Verify the writer GraphicsOp gets Wedge=true via conversion.
	gops := convertGraphicsOps(ops)
	if len(gops) != 1 {
		t.Fatalf("expected 1 writer op, got %d", len(gops))
	}
	if !gops[0].Wedge {
		t.Error("expected writer GraphicsOp.Wedge = true for nil ArcOptions.Wedge")
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}
}

func TestDrawArc_NegativeSweepClockwise(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}

	opts := &ArcOptions{
		StrokeColor: &Black,
		StrokeWidth: 1.0,
	}
	// Negative sweep = clockwise arc.
	err = page.DrawArc(200, 400, 80, 80, 90, -90, opts)
	if err != nil {
		t.Fatalf("DrawArc clockwise failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}
}
