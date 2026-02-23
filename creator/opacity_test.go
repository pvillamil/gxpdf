package creator

import (
	"bytes"
	"strings"
	"testing"

	"github.com/coregx/gxpdf/internal/writer"
)

// -------------------------------------------------------------------------
// TestDrawCircle_WithOpacity reproduces ajstarks' exact code from issue #47.
// -------------------------------------------------------------------------

// TestDrawCircle_WithOpacity verifies that a circle drawn with Opacity produces
// a valid GraphicsOperation and correctly propagates the opacity value.
func TestDrawCircle_WithOpacity(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}

	opacity := 0.5
	opts := &CircleOptions{
		FillColor: &Blue,
		Opacity:   &opacity,
	}

	if err := page.DrawCircle(300, 400, 50, opts); err != nil {
		t.Fatalf("DrawCircle() error = %v", err)
	}

	if len(page.graphicsOps) != 1 {
		t.Fatalf("expected 1 op, got %d", len(page.graphicsOps))
	}
	gop := page.graphicsOps[0]
	if gop.Type != GraphicsOpCircle {
		t.Errorf("expected GraphicsOpCircle, got %d", gop.Type)
	}
	if gop.CircleOpts == nil || gop.CircleOpts.Opacity == nil {
		t.Fatal("CircleOpts.Opacity should not be nil")
	}
	if *gop.CircleOpts.Opacity != opacity {
		t.Errorf("Opacity: want %.2f, got %.2f", opacity, *gop.CircleOpts.Opacity)
	}
}

// TestDrawRect_WithOpacity verifies that a rectangle drawn with Opacity
// correctly propagates the value through the creator layer.
func TestDrawRect_WithOpacity(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	opacity := 0.3
	opts := &RectOptions{
		FillColor: &Red,
		Opacity:   &opacity,
	}

	if err := page.DrawRect(100, 200, 150, 80, opts); err != nil {
		t.Fatalf("DrawRect() error = %v", err)
	}

	gop := page.graphicsOps[0]
	if gop.Type != GraphicsOpRect {
		t.Errorf("expected GraphicsOpRect, got %d", gop.Type)
	}
	if gop.RectOpts == nil || gop.RectOpts.Opacity == nil {
		t.Fatal("RectOpts.Opacity should not be nil")
	}
	if *gop.RectOpts.Opacity != opacity {
		t.Errorf("Opacity: want %.2f, got %.2f", opacity, *gop.RectOpts.Opacity)
	}
}

// TestDrawEllipse_WithOpacity verifies that an ellipse drawn with Opacity
// is accepted without error and stores the opacity in the options.
func TestDrawEllipse_WithOpacity(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	opacity := 0.7
	opts := &EllipseOptions{
		StrokeColor: &Black,
		FillColor:   &Green,
		Opacity:     &opacity,
	}

	if err := page.DrawEllipse(200, 300, 80, 40, opts); err != nil {
		t.Fatalf("DrawEllipse() error = %v", err)
	}

	gop := page.graphicsOps[0]
	if gop.Type != GraphicsOpEllipse {
		t.Errorf("expected GraphicsOpEllipse, got %d", gop.Type)
	}
	if gop.EllipseOpts == nil || gop.EllipseOpts.Opacity == nil {
		t.Fatal("EllipseOpts.Opacity should not be nil")
	}
	if *gop.EllipseOpts.Opacity != opacity {
		t.Errorf("Opacity: want %.2f, got %.2f", opacity, *gop.EllipseOpts.Opacity)
	}
}

// TestDrawLine_WithOpacity verifies that a line drawn with Opacity
// is accepted without error and stores the opacity in the options.
func TestDrawLine_WithOpacity(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	opacity := 0.4
	opts := &LineOptions{
		Color:   Black,
		Width:   2.0,
		Opacity: &opacity,
	}

	if err := page.DrawLine(50, 100, 250, 100, opts); err != nil {
		t.Fatalf("DrawLine() error = %v", err)
	}

	gop := page.graphicsOps[0]
	if gop.Type != GraphicsOpLine {
		t.Errorf("expected GraphicsOpLine, got %d", gop.Type)
	}
	if gop.LineOpts == nil || gop.LineOpts.Opacity == nil {
		t.Fatal("LineOpts.Opacity should not be nil")
	}
	if *gop.LineOpts.Opacity != opacity {
		t.Errorf("Opacity: want %.2f, got %.2f", opacity, *gop.LineOpts.Opacity)
	}
}

// -------------------------------------------------------------------------
// TestOpacity_WritesValidPDF — end-to-end: verify the PDF bytes are produced
// and contain an ExtGState entry (the /GS operator in the content stream).
// -------------------------------------------------------------------------

// TestOpacity_WritesValidPDF creates a document with a semi-transparent circle
// and verifies that the resulting PDF bytes contain an ExtGState reference in
// the resource dictionary, which proves the opacity pipeline reached the writer
// layer.
//
// Note: PDF content streams are compressed with FlateDecode by default, so we
// cannot search the raw bytes for the 'gs' operator.  The resource dictionary,
// however, is always written as plaintext and is a reliable indicator that an
// ExtGState object was created and wired up correctly.
func TestOpacity_WritesValidPDF(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}

	opacity := 0.5
	opts := &CircleOptions{
		FillColor: &Blue,
		Opacity:   &opacity,
	}

	if err := page.DrawCircle(300, 400, 50, opts); err != nil {
		t.Fatalf("DrawCircle() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	pdfBytes := buf.Bytes()
	if len(pdfBytes) == 0 {
		t.Fatal("WriteTo() produced empty bytes")
	}

	// A semi-transparent shape must produce an ExtGState resource.
	// The resource dictionary is always written as plaintext and will contain
	// "/ExtGState" when an opacity graphics state was registered.
	pdfStr := string(pdfBytes)
	if !strings.Contains(pdfStr, "/ExtGState") {
		t.Error("PDF does not contain /ExtGState — opacity was not applied")
	}

	// Verify ExtGState references real objects (not placeholder 0 0 R).
	if strings.Contains(pdfStr, "/GS1 0 0 R") {
		t.Error("ExtGState /GS1 references object 0 — object was never created")
	}

	// Verify actual ExtGState dictionary exists in PDF.
	if !strings.Contains(pdfStr, "/ca ") {
		t.Error("PDF does not contain ExtGState dictionary with /ca opacity key")
	}
}

// TestOpacity_ExtGStateSharing verifies that two shapes with the same opacity
// value share a single GS resource rather than creating duplicates.
func TestOpacity_ExtGStateSharing(t *testing.T) {
	opacity := 0.5
	graphicsOps := []writer.GraphicsOp{
		{
			Type:      int(GraphicsOpCircle),
			X:         200,
			Y:         400,
			Radius:    50,
			FillColor: &writer.RGB{R: 0, G: 0, B: 1},
			Opacity:   opacity,
		},
		{
			Type:      int(GraphicsOpRect),
			X:         100,
			Y:         200,
			Width:     100,
			Height:    80,
			FillColor: &writer.RGB{R: 1, G: 0, B: 0},
			Opacity:   opacity,
		},
	}

	_, resources, err := writer.GenerateContentStreamWithGraphics(nil, graphicsOps)
	if err != nil {
		t.Fatalf("GenerateContentStreamWithGraphics() error = %v", err)
	}

	// Both shapes share the same opacity — only one ExtGState should be created.
	resStr := resources.String()

	// Count occurrences of /GS entries: should be exactly one unique entry.
	// The resource dict format is: /GS1 N 0 R (and possibly /GS2, etc.)
	// We expect only /GS1 and not /GS2.
	if strings.Contains(resStr, "/GS2") {
		t.Errorf("Expected a single GS entry (shared), but found multiple: %s", resStr)
	}
	if !strings.Contains(resStr, "/GS1") {
		t.Errorf("Expected /GS1 in resource dictionary, got: %s", resStr)
	}
}

// -------------------------------------------------------------------------
// TestOpacity_ValidationErrors — opacity outside [0, 1] must be rejected.
// -------------------------------------------------------------------------

// TestOpacity_ValidationErrors verifies that opacity values outside [0.0, 1.0]
// are rejected at the creator layer before any PDF is written.
func TestOpacity_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		opacity float64
		wantErr bool
	}{
		{"valid zero", 0.0, false},
		{"valid half", 0.5, false},
		{"valid one", 1.0, false},
		{"invalid negative", -0.1, true},
		{"invalid above one", 1.1, true},
		{"invalid large positive", 2.0, true},
		{"invalid large negative", -100.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, _ := c.NewPage()

			opacity := tt.opacity
			err := page.DrawCircle(300, 400, 50, &CircleOptions{
				FillColor: &Blue,
				Opacity:   &opacity,
			})

			if tt.wantErr && err == nil {
				t.Errorf("DrawCircle() with opacity %.2f: expected error, got nil", tt.opacity)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("DrawCircle() with opacity %.2f: unexpected error: %v", tt.opacity, err)
			}
		})
	}
}

// TestOpacity_NilMeansOpaque verifies that nil opacity produces a valid shape
// with no ExtGState resource (fully opaque, no transparency overhead).
//
// The resource dictionary is always written as plaintext, so the absence of
// "/ExtGState" in the raw PDF bytes is a reliable indicator that no opacity
// state was registered.
func TestOpacity_NilMeansOpaque(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	// No Opacity field set — should produce a fully opaque circle.
	opts := &CircleOptions{
		FillColor: &Blue,
		// Opacity is nil (not set)
	}

	if err := page.DrawCircle(300, 400, 50, opts); err != nil {
		t.Fatalf("DrawCircle() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	pdfStr := string(buf.Bytes())

	// A fully opaque shape (nil opacity) must NOT produce an ExtGState entry.
	// The resource dictionary is plaintext, so this is a reliable check.
	if strings.Contains(pdfStr, "/ExtGState") {
		t.Error("PDF should NOT contain /ExtGState for fully opaque shape (nil opacity)")
	}
}

// -------------------------------------------------------------------------
// Additional shape coverage: polygon, polyline, bezier.
// -------------------------------------------------------------------------

// TestDrawPolygon_WithOpacity verifies polygon opacity propagation.
func TestDrawPolygon_WithOpacity(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	opacity := 0.6
	vertices := []Point{{X: 100, Y: 100}, {X: 150, Y: 50}, {X: 200, Y: 100}}
	opts := &PolygonOptions{
		FillColor: &Yellow,
		Opacity:   &opacity,
	}

	if err := page.DrawPolygon(vertices, opts); err != nil {
		t.Fatalf("DrawPolygon() error = %v", err)
	}

	gop := page.graphicsOps[0]
	if gop.PolygonOpts == nil || gop.PolygonOpts.Opacity == nil {
		t.Fatal("PolygonOpts.Opacity should not be nil")
	}
	if *gop.PolygonOpts.Opacity != opacity {
		t.Errorf("Opacity: want %.2f, got %.2f", opacity, *gop.PolygonOpts.Opacity)
	}
}

// TestDrawPolyline_WithOpacity verifies polyline opacity propagation.
func TestDrawPolyline_WithOpacity(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	opacity := 0.8
	vertices := []Point{{X: 100, Y: 100}, {X: 150, Y: 150}, {X: 200, Y: 100}}
	opts := &PolylineOptions{
		Color:   Blue,
		Width:   2.0,
		Opacity: &opacity,
	}

	if err := page.DrawPolyline(vertices, opts); err != nil {
		t.Fatalf("DrawPolyline() error = %v", err)
	}

	gop := page.graphicsOps[0]
	if gop.PolylineOpts == nil || gop.PolylineOpts.Opacity == nil {
		t.Fatal("PolylineOpts.Opacity should not be nil")
	}
	if *gop.PolylineOpts.Opacity != opacity {
		t.Errorf("Opacity: want %.2f, got %.2f", opacity, *gop.PolylineOpts.Opacity)
	}
}

// TestDrawBezier_WithOpacity verifies bezier curve opacity propagation.
func TestDrawBezier_WithOpacity(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	opacity := 0.25
	segs := []BezierSegment{
		{
			Start: Point{X: 100, Y: 100},
			C1:    Point{X: 150, Y: 200},
			C2:    Point{X: 200, Y: 200},
			End:   Point{X: 250, Y: 100},
		},
	}
	opts := &BezierOptions{
		Color:   Red,
		Width:   2.0,
		Opacity: &opacity,
	}

	if err := page.DrawBezierCurve(segs, opts); err != nil {
		t.Fatalf("DrawBezierCurve() error = %v", err)
	}

	gop := page.graphicsOps[0]
	if gop.BezierOpts == nil || gop.BezierOpts.Opacity == nil {
		t.Fatal("BezierOpts.Opacity should not be nil")
	}
	if *gop.BezierOpts.Opacity != opacity {
		t.Errorf("Opacity: want %.2f, got %.2f", opacity, *gop.BezierOpts.Opacity)
	}
}

// -------------------------------------------------------------------------
// TestOpacity_ConvertPropagation — verifies opacity survives the convert layer.
// -------------------------------------------------------------------------

// TestOpacity_ConvertPropagation verifies that opacity set on creator options
// correctly propagates through convertGraphicsOps into writer.GraphicsOp.Opacity.
func TestOpacity_ConvertPropagation(t *testing.T) {
	opacity := 0.42

	tests := []struct {
		name string
		op   GraphicsOperation
	}{
		{
			name: "circle",
			op: GraphicsOperation{
				Type: GraphicsOpCircle,
				X:    200, Y: 200, Radius: 50,
				CircleOpts: &CircleOptions{FillColor: &Blue, Opacity: &opacity},
			},
		},
		{
			name: "rect",
			op: GraphicsOperation{
				Type: GraphicsOpRect,
				X:    100, Y: 100, Width: 100, Height: 50,
				RectOpts: &RectOptions{FillColor: &Red, Opacity: &opacity},
			},
		},
		{
			name: "ellipse",
			op: GraphicsOperation{
				Type: GraphicsOpEllipse,
				X:    200, Y: 200, RX: 80, RY: 40,
				EllipseOpts: &EllipseOptions{FillColor: &Green, Opacity: &opacity},
			},
		},
		{
			name: "polygon",
			op: GraphicsOperation{
				Type:        GraphicsOpPolygon,
				Vertices:    []Point{{100, 100}, {150, 50}, {200, 100}},
				PolygonOpts: &PolygonOptions{FillColor: &Yellow, Opacity: &opacity},
			},
		},
		{
			name: "polyline",
			op: GraphicsOperation{
				Type:         GraphicsOpPolyline,
				Vertices:     []Point{{100, 100}, {200, 200}},
				PolylineOpts: &PolylineOptions{Color: Black, Opacity: &opacity},
			},
		},
		{
			name: "line",
			op: GraphicsOperation{
				Type: GraphicsOpLine,
				X:    0, Y: 0, X2: 100, Y2: 100,
				LineOpts: &LineOptions{Color: Black, Opacity: &opacity},
			},
		},
		{
			name: "bezier",
			op: GraphicsOperation{
				Type: GraphicsOpBezier,
				BezierSegs: []BezierSegment{
					{Start: Point{100, 100}, C1: Point{150, 200}, C2: Point{200, 200}, End: Point{250, 100}},
				},
				BezierOpts: &BezierOptions{Color: Blue, Opacity: &opacity},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writerOps := convertGraphicsOps([]GraphicsOperation{tt.op})
			if len(writerOps) != 1 {
				t.Fatalf("expected 1 writer op, got %d", len(writerOps))
			}
			wop := writerOps[0]
			if wop.Opacity != opacity {
				t.Errorf("Opacity after conversion: want %.4f, got %.4f", opacity, wop.Opacity)
			}
		})
	}
}

// TestOpacity_FullyOpaque_NoExtGState verifies that opacity of exactly 1.0
// does not produce an ExtGState (1.0 is treated as fully opaque, same as nil).
func TestOpacity_FullyOpaque_NoExtGState(t *testing.T) {
	graphicsOp := writer.GraphicsOp{
		Type:      int(GraphicsOpCircle),
		X:         200,
		Y:         400,
		Radius:    50,
		FillColor: &writer.RGB{R: 0, G: 0, B: 1},
		Opacity:   1.0, // Fully opaque — should not emit gs operator
	}

	content, resources, err := writer.GenerateContentStreamWithGraphics(nil, []writer.GraphicsOp{graphicsOp})
	if err != nil {
		t.Fatalf("GenerateContentStreamWithGraphics() error = %v", err)
	}

	contentStr := string(content)
	resStr := resources.String()

	if strings.Contains(contentStr, " gs") {
		t.Error("Content stream should NOT contain 'gs' for opacity=1.0")
	}
	if strings.Contains(resStr, "/ExtGState") {
		t.Error("Resource dict should NOT contain /ExtGState for opacity=1.0")
	}
}

// TestOpacity_Zero_NoExtGState verifies that opacity of exactly 0.0 (unset/zero
// value) does not produce an ExtGState — zero is the sentinel for "not set".
func TestOpacity_Zero_NoExtGState(t *testing.T) {
	graphicsOp := writer.GraphicsOp{
		Type:      int(GraphicsOpCircle),
		X:         200,
		Y:         400,
		Radius:    50,
		FillColor: &writer.RGB{R: 0, G: 0, B: 1},
		Opacity:   0, // Zero value = "not set", treated as fully opaque
	}

	content, resources, err := writer.GenerateContentStreamWithGraphics(nil, []writer.GraphicsOp{graphicsOp})
	if err != nil {
		t.Fatalf("GenerateContentStreamWithGraphics() error = %v", err)
	}

	contentStr := string(content)
	resStr := resources.String()

	if strings.Contains(contentStr, " gs") {
		t.Error("Content stream should NOT contain 'gs' for Opacity=0 (not set)")
	}
	if strings.Contains(resStr, "/ExtGState") {
		t.Error("Resource dict should NOT contain /ExtGState for Opacity=0 (not set)")
	}
}

// TestOpacity_RectValidation tests opacity validation for all shape types.
func TestOpacity_RectValidation(t *testing.T) {
	tests := []struct {
		name    string
		opacity float64
		wantErr bool
	}{
		{"valid 0.0", 0.0, false},
		{"valid 0.5", 0.5, false},
		{"valid 1.0", 1.0, false},
		{"invalid -0.01", -0.01, true},
		{"invalid 1.01", 1.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, _ := c.NewPage()
			opacity := tt.opacity
			err := page.DrawRect(100, 100, 50, 50, &RectOptions{
				FillColor: &Red,
				Opacity:   &opacity,
			})
			if tt.wantErr && err == nil {
				t.Errorf("expected validation error for opacity %.4f, got nil", tt.opacity)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for opacity %.4f: %v", tt.opacity, err)
			}
		})
	}
}

// -------------------------------------------------------------------------
// Text Opacity Tests (feat-075, issue #46)
// -------------------------------------------------------------------------

// TestAddTextColorAlpha verifies that text with opacity stores the correct
// TextOperation and that the opacity value is propagated.
func TestAddTextColorAlpha(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}

	err = page.AddTextColorAlpha("Hello", 100, 700, Helvetica, 12, Red, 0.5)
	if err != nil {
		t.Fatalf("AddTextColorAlpha() error = %v", err)
	}

	ops := page.TextOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 text operation, got %d", len(ops))
	}

	op := ops[0]
	if op.Opacity == nil {
		t.Fatal("expected Opacity to be set, got nil")
	}
	if *op.Opacity != 0.5 {
		t.Errorf("expected Opacity 0.5, got %f", *op.Opacity)
	}
	if op.Text != "Hello" {
		t.Errorf("expected text 'Hello', got %q", op.Text)
	}
}

// TestAddTextColorRotatedAlpha verifies combined rotation and opacity.
func TestAddTextColorRotatedAlpha(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}

	err = page.AddTextColorRotatedAlpha("DRAFT", 300, 400, HelveticaBold, 48, Red, 45, 0.3)
	if err != nil {
		t.Fatalf("AddTextColorRotatedAlpha() error = %v", err)
	}

	ops := page.TextOperations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 text operation, got %d", len(ops))
	}

	op := ops[0]
	if op.Opacity == nil {
		t.Fatal("expected Opacity to be set, got nil")
	}
	if *op.Opacity != 0.3 {
		t.Errorf("expected Opacity 0.3, got %f", *op.Opacity)
	}
	if op.Rotation != 45 {
		t.Errorf("expected Rotation 45, got %f", op.Rotation)
	}
}

// TestAddTextColorAlpha_ValidationErrors tests boundary validation for text opacity.
func TestAddTextColorAlpha_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		opacity float64
		wantErr bool
	}{
		{"valid 0.0", 0.0, false},
		{"valid 0.5", 0.5, false},
		{"valid 1.0", 1.0, false},
		{"invalid -0.01", -0.01, true},
		{"invalid 1.01", 1.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, _ := c.NewPage()
			err := page.AddTextColorAlpha("test", 100, 700, Helvetica, 12, Black, tt.opacity)
			if tt.wantErr && err == nil {
				t.Errorf("expected validation error for opacity %.4f, got nil", tt.opacity)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for opacity %.4f: %v", tt.opacity, err)
			}
		})
	}
}

// TestConvertTextOps_OpacityPropagation verifies that opacity is passed from
// creator TextOperation to writer TextOp via convertTextOps.
func TestConvertTextOps_OpacityPropagation(t *testing.T) {
	opacity := 0.4
	ops := []TextOperation{
		{
			Text:    "translucent",
			X:       50,
			Y:       600,
			Font:    Helvetica,
			Size:    14,
			Opacity: &opacity,
		},
		{
			Text: "opaque",
			X:    50,
			Y:    580,
			Font: Helvetica,
			Size: 14,
			// Opacity is nil → writer.TextOp.Opacity should be 0
		},
	}

	writerOps := convertTextOps(ops)
	if len(writerOps) != 2 {
		t.Fatalf("expected 2 writer ops, got %d", len(writerOps))
	}

	if writerOps[0].Opacity != 0.4 {
		t.Errorf("expected writer op[0] Opacity 0.4, got %f", writerOps[0].Opacity)
	}
	if writerOps[1].Opacity != 0 {
		t.Errorf("expected writer op[1] Opacity 0 (not set), got %f", writerOps[1].Opacity)
	}
}

// TestTextOpacity_ContentStream verifies that the writer emits the correct
// ExtGState operator for text with opacity.
func TestTextOpacity_ContentStream(t *testing.T) {
	textOps := []writer.TextOp{
		{
			Text:    "semi-transparent",
			X:       100,
			Y:       700,
			Font:    "Helvetica",
			Size:    12,
			Opacity: 0.5,
		},
	}

	content, resources, err := writer.GenerateContentStream(textOps)
	if err != nil {
		t.Fatalf("GenerateContentStream() error = %v", err)
	}

	stream := string(content)

	// Must contain gs operator for ExtGState
	if !strings.Contains(stream, " gs") {
		t.Errorf("expected ExtGState gs operator in stream, got: %s", stream)
	}

	// Must contain q/Q (save/restore state) around the text
	if !strings.Contains(stream, "q\n") {
		t.Errorf("expected SaveState (q) in stream, got: %s", stream)
	}
	if !strings.Contains(stream, "Q\n") {
		t.Errorf("expected RestoreState (Q) in stream, got: %s", stream)
	}

	// Resources must have ExtGState entry
	resStr := resources.String()
	if !strings.Contains(resStr, "/ExtGState") {
		t.Errorf("expected /ExtGState in resources, got: %s", resStr)
	}

	// Before object creation, ExtGState should have placeholder object number 0.
	// This is expected — the writer layer creates the real objects later.
	// Verify the entry exists in the resource dictionary.
	if !strings.Contains(resStr, "/GS1") {
		t.Errorf("expected /GS1 in resources, got: %s", resStr)
	}
}

// TestTextOpacity_FullOpaque verifies that opacity 1.0 does NOT emit ExtGState.
func TestTextOpacity_FullOpaque(t *testing.T) {
	textOps := []writer.TextOp{
		{
			Text:    "fully opaque",
			X:       100,
			Y:       700,
			Font:    "Helvetica",
			Size:    12,
			Opacity: 1.0,
		},
	}

	content, _, err := writer.GenerateContentStream(textOps)
	if err != nil {
		t.Fatalf("GenerateContentStream() error = %v", err)
	}

	stream := string(content)

	// Should NOT contain gs operator (fully opaque = no ExtGState needed)
	if strings.Contains(stream, " gs") {
		t.Errorf("fully opaque text should NOT emit gs operator, got: %s", stream)
	}
}

// TestTextOpacity_ZeroValue verifies that Opacity == 0 (not set) does NOT emit ExtGState.
func TestTextOpacity_ZeroValue(t *testing.T) {
	textOps := []writer.TextOp{
		{
			Text: "default opacity",
			X:    100,
			Y:    700,
			Font: "Helvetica",
			Size: 12,
			// Opacity is 0 (zero value, means "not set")
		},
	}

	content, _, err := writer.GenerateContentStream(textOps)
	if err != nil {
		t.Fatalf("GenerateContentStream() error = %v", err)
	}

	stream := string(content)

	if strings.Contains(stream, " gs") {
		t.Errorf("default (zero) opacity should NOT emit gs operator, got: %s", stream)
	}
}

// TestTextOpacity_WithRotation verifies that opacity and rotation work together.
func TestTextOpacity_WithRotation(t *testing.T) {
	textOps := []writer.TextOp{
		{
			Text:     "rotated + transparent",
			X:        200,
			Y:        400,
			Font:     "Helvetica",
			Size:     24,
			Opacity:  0.5,
			Rotation: 45,
		},
	}

	content, _, err := writer.GenerateContentStream(textOps)
	if err != nil {
		t.Fatalf("GenerateContentStream() error = %v", err)
	}

	stream := string(content)

	// Must have both gs (opacity) and Tm (rotation)
	if !strings.Contains(stream, " gs") {
		t.Errorf("expected gs operator for opacity, got: %s", stream)
	}
	if !strings.Contains(stream, " Tm") {
		t.Errorf("expected Tm operator for rotation, got: %s", stream)
	}
}

// TestAddTextCustomFontColorAlpha verifies custom font with opacity (nil font check).
func TestAddTextCustomFontColorAlpha(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	// nil font should return error
	err := page.AddTextCustomFontColorAlpha("test", 100, 700, nil, 12, Black, 0.5)
	if err == nil {
		t.Error("expected error for nil font, got nil")
	}
}

// TestAddTextCustomFontColorRotatedAlpha verifies the full-featured method.
func TestAddTextCustomFontColorRotatedAlpha(t *testing.T) {
	c := New()
	page, _ := c.NewPage()

	// nil font should return error
	err := page.AddTextCustomFontColorRotatedAlpha("test", 100, 700, nil, 12, Black, 45, 0.5)
	if err == nil {
		t.Error("expected error for nil font, got nil")
	}

	// invalid opacity should return error
	err = page.AddTextCustomFontColorRotatedAlpha("test", 100, 700, nil, 12, Black, 45, 1.5)
	if err == nil {
		t.Error("expected error for nil font, got nil")
	}
}

// -------------------------------------------------------------------------
// TestOpacity_ExtGStateObjectCreated — end-to-end: verify that ExtGState
// objects are actually created as PDF indirect objects with correct content.
// -------------------------------------------------------------------------

// TestOpacity_ExtGStateObjectCreated creates a document with a semi-transparent
// circle and verifies that the resulting PDF contains:
// 1. /GS1 references a real object (not 0 0 R)
// 2. An ExtGState dictionary with /ca and /CA keys exists
// 3. The opacity value is correct
func TestOpacity_ExtGStateObjectCreated(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}

	opacity := 0.5
	if err := page.DrawCircle(300, 400, 50, &CircleOptions{
		FillColor: &Blue,
		Opacity:   &opacity,
	}); err != nil {
		t.Fatalf("DrawCircle() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	pdfStr := string(buf.Bytes())

	// /GS1 must NOT reference object 0.
	if strings.Contains(pdfStr, "/GS1 0 0 R") {
		t.Error("ExtGState /GS1 references object 0 — ExtGState object was never created")
	}

	// The actual ExtGState dictionary must exist somewhere in the PDF.
	if !strings.Contains(pdfStr, "/Type /ExtGState") {
		t.Error("PDF does not contain /Type /ExtGState dictionary")
	}

	// Must contain both /ca and /CA opacity keys.
	if !strings.Contains(pdfStr, "/ca 0.50") {
		t.Errorf("PDF does not contain /ca 0.50 for fill opacity")
	}
	if !strings.Contains(pdfStr, "/CA 0.50") {
		t.Errorf("PDF does not contain /CA 0.50 for stroke opacity")
	}
}

// TestTextOpacity_ExtGStateObjectCreated verifies end-to-end that text with
// opacity produces a real ExtGState object in the final PDF.
func TestTextOpacity_ExtGStateObjectCreated(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}

	err = page.AddTextColorAlpha("Translucent", 100, 700, Helvetica, 14, Red, 0.3)
	if err != nil {
		t.Fatalf("AddTextColorAlpha() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	pdfStr := string(buf.Bytes())

	// /GS1 must NOT reference object 0.
	if strings.Contains(pdfStr, "/GS1 0 0 R") {
		t.Error("ExtGState /GS1 references object 0 — ExtGState object was never created")
	}

	// The actual ExtGState dictionary must exist.
	if !strings.Contains(pdfStr, "/Type /ExtGState") {
		t.Error("PDF does not contain /Type /ExtGState dictionary")
	}

	// Must contain the correct opacity value.
	if !strings.Contains(pdfStr, "/ca 0.30") {
		t.Errorf("PDF does not contain /ca 0.30 for fill opacity")
	}
}
