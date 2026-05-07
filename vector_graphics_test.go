package gxpdf_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/coregx/gxpdf"
	"github.com/coregx/gxpdf/creator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createVectorTestPDF builds a small PDF in memory using the Creator API and
// returns its raw bytes. The document contains:
//   - A stroked horizontal line (black, 2pt width)
//   - A filled rectangle (blue fill)
//   - A stroked+filled rectangle (red stroke, green fill)
//   - A cubic Bézier curve
func createVectorTestPDF(t *testing.T) []byte {
	t.Helper()

	c := creator.New()

	page, err := c.NewPage()
	require.NoError(t, err, "NewPage failed")

	// Black horizontal line, 2pt wide.
	lineOpts := &creator.LineOptions{
		Color: creator.Color{R: 0, G: 0, B: 0},
		Width: 2.0,
	}
	err = page.DrawLine(50, 700, 400, 700, lineOpts)
	require.NoError(t, err, "DrawLine failed")

	// Blue filled rectangle.
	rectOpts := &creator.RectOptions{
		FillColor: &creator.Color{R: 0, G: 0, B: 1},
	}
	err = page.DrawRect(100, 600, 150, 80, rectOpts)
	require.NoError(t, err, "DrawRect (fill) failed")

	// Red-stroked, green-filled rectangle.
	rectOpts2 := &creator.RectOptions{
		StrokeColor: &creator.Color{R: 1, G: 0, B: 0},
		FillColor:   &creator.Color{R: 0, G: 1, B: 0},
		StrokeWidth: 1.5,
	}
	err = page.DrawRect(300, 600, 150, 80, rectOpts2)
	require.NoError(t, err, "DrawRect (fill+stroke) failed")

	// Cubic Bézier curve.
	segments := []creator.BezierSegment{
		{
			Start: creator.Point{X: 50, Y: 500},
			C1:    creator.Point{X: 100, Y: 550},
			C2:    creator.Point{X: 200, Y: 450},
			End:   creator.Point{X: 300, Y: 500},
		},
	}
	bezierOpts := &creator.BezierOptions{
		Color: creator.Color{R: 0.5, G: 0, B: 0.5},
		Width: 1.0,
	}
	err = page.DrawBezierCurve(segments, bezierOpts)
	require.NoError(t, err, "DrawBezierCurve failed")

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	require.NoError(t, err, "WriteTo failed")

	return buf.Bytes()
}

// openFromBytes opens a PDF from in-memory bytes.
// It writes the bytes to a temp file that is cleaned up after the test.
func openFromBytes(t *testing.T, data []byte) *gxpdf.Document {
	t.Helper()

	tmpFile := t.TempDir() + "/test.pdf"
	err := writeBytesToFile(tmpFile, data)
	require.NoError(t, err, "writeBytesToFile failed")

	doc, err := gxpdf.Open(tmpFile)
	require.NoError(t, err, "gxpdf.Open failed")
	t.Cleanup(func() { _ = doc.Close() })
	return doc
}

func writeBytesToFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

// ── PathVerb public API tests ─────────────────────────────────────────────────

func TestPathVerb_String_Public(t *testing.T) {
	tests := []struct {
		verb gxpdf.PathVerb
		want string
	}{
		{gxpdf.PathVerbMoveTo, "MoveTo"},
		{gxpdf.PathVerbLineTo, "LineTo"},
		{gxpdf.PathVerbCubicTo, "CubicTo"},
		{gxpdf.PathVerbQuadTo, "QuadTo"},
		{gxpdf.PathVerbClose, "Close"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.verb.String())
		})
	}
}

func TestVerbCoordCount_Public(t *testing.T) {
	assert.Equal(t, 2, gxpdf.VerbCoordCount(gxpdf.PathVerbMoveTo))
	assert.Equal(t, 2, gxpdf.VerbCoordCount(gxpdf.PathVerbLineTo))
	assert.Equal(t, 6, gxpdf.VerbCoordCount(gxpdf.PathVerbCubicTo))
	assert.Equal(t, 4, gxpdf.VerbCoordCount(gxpdf.PathVerbQuadTo))
	assert.Equal(t, 0, gxpdf.VerbCoordCount(gxpdf.PathVerbClose))
}

func TestPaintMode_String_Public(t *testing.T) {
	assert.Equal(t, "Stroke", gxpdf.PaintModeStroke.String())
	assert.Equal(t, "Fill", gxpdf.PaintModeFill.String())
	assert.Equal(t, "FillStroke", gxpdf.PaintModeFillStroke.String())
}

// ── VectorPath String ─────────────────────────────────────────────────────────

func TestVectorPath_String_Public(t *testing.T) {
	vp := &gxpdf.VectorPath{
		PageNum:   0,
		Verbs:     []gxpdf.PathVerb{gxpdf.PathVerbMoveTo, gxpdf.PathVerbLineTo},
		PaintMode: gxpdf.PaintModeStroke,
		LineWidth: 2.0,
		Opacity:   1.0,
	}
	s := vp.String()
	assert.Contains(t, s, "page=0")
	assert.Contains(t, s, "verbs=2")
	assert.Contains(t, s, "Stroke")
}

// ── GetVectorGraphicsForPage out-of-range ─────────────────────────────────────

func TestDocument_GetVectorGraphicsForPage_OutOfRange(t *testing.T) {
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	_, err := doc.GetVectorGraphicsForPage(-1)
	assert.Error(t, err, "negative page should return error")

	_, err = doc.GetVectorGraphicsForPage(9999)
	assert.Error(t, err, "page beyond count should return error")
}

// ── Round-trip: Creator → extract ────────────────────────────────────────────

func TestGetVectorGraphicsForPage_ReturnsNonEmptySlice(t *testing.T) {
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsForPage(0)
	require.NoError(t, err)
	assert.NotEmpty(t, paths, "expected at least one vector path on page 0")
}

func TestGetVectorGraphics_AllPages(t *testing.T) {
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths := doc.GetVectorGraphics()
	assert.NotEmpty(t, paths)
}

func TestGetVectorGraphicsWithError_AllPages(t *testing.T) {
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsWithError()
	require.NoError(t, err)
	assert.NotEmpty(t, paths)
}

func TestGetVectorGraphicsForPage_PageNumSet(t *testing.T) {
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsForPage(0)
	require.NoError(t, err)

	for _, p := range paths {
		assert.Equal(t, 0, p.PageNum, "all paths should be from page 0")
	}
}

func TestGetVectorGraphicsForPage_VerbCoordsConsistency(t *testing.T) {
	// Every verb must have the correct number of following coords.
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsForPage(0)
	require.NoError(t, err)

	for _, path := range paths {
		idx := 0
		for vi, verb := range path.Verbs {
			n := gxpdf.VerbCoordCount(verb)
			assert.GreaterOrEqual(t, len(path.Coords)-idx, n,
				"path verb[%d]=%s needs %d coords but only %d remain",
				vi, verb, n, len(path.Coords)-idx)
			idx += n
		}
		assert.Equal(t, len(path.Coords), idx,
			"all %d coords must be consumed by verbs", len(path.Coords))
	}
}

func TestGetVectorGraphicsForPage_HasStrokePaths(t *testing.T) {
	// The test PDF has a DrawLine which should produce a stroke path.
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsForPage(0)
	require.NoError(t, err)

	var foundStroke bool
	for _, p := range paths {
		if p.PaintMode == gxpdf.PaintModeStroke {
			foundStroke = true
			break
		}
	}
	assert.True(t, foundStroke, "expected at least one stroke path (from DrawLine)")
}

func TestGetVectorGraphicsForPage_HasFillPaths(t *testing.T) {
	// The test PDF has a blue filled rectangle.
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsForPage(0)
	require.NoError(t, err)

	var foundFill bool
	for _, p := range paths {
		if p.PaintMode == gxpdf.PaintModeFill || p.PaintMode == gxpdf.PaintModeFillStroke {
			foundFill = true
			break
		}
	}
	assert.True(t, foundFill, "expected at least one fill path (from DrawRect blue fill)")
}

func TestGetVectorGraphicsForPage_HasCubicBezier(t *testing.T) {
	// The test PDF has a Bézier curve drawn via DrawBezierCurve.
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsForPage(0)
	require.NoError(t, err)

	var foundCubic bool
	for _, path := range paths {
		for _, verb := range path.Verbs {
			if verb == gxpdf.PathVerbCubicTo {
				foundCubic = true
				break
			}
		}
		if foundCubic {
			break
		}
	}
	assert.True(t, foundCubic, "expected at least one CubicTo verb from DrawBezierCurve")
}

func TestGetVectorGraphicsForPage_ColorValues(t *testing.T) {
	// Paths must have valid RGB color values in [0,1].
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsForPage(0)
	require.NoError(t, err)

	for _, p := range paths {
		for i, v := range p.StrokeColor {
			assert.GreaterOrEqual(t, v, 0.0, "StrokeColor[%d] out of range", i)
			assert.LessOrEqual(t, v, 1.0, "StrokeColor[%d] out of range", i)
		}
		for i, v := range p.FillColor {
			assert.GreaterOrEqual(t, v, 0.0, "FillColor[%d] out of range", i)
			assert.LessOrEqual(t, v, 1.0, "FillColor[%d] out of range", i)
		}
		assert.GreaterOrEqual(t, p.Opacity, 0.0)
		assert.LessOrEqual(t, p.Opacity, 1.0)
		assert.GreaterOrEqual(t, p.LineWidth, 0.0)
	}
}

func TestGetVectorGraphicsForPage_DefaultMiterLimit(t *testing.T) {
	// PDF default miter limit is 10.0.
	data := createVectorTestPDF(t)
	doc := openFromBytes(t, data)

	paths, err := doc.GetVectorGraphicsForPage(0)
	require.NoError(t, err)
	require.NotEmpty(t, paths)

	// At least the first path should have the default miter limit.
	assert.InDelta(t, 10.0, paths[0].MiterLimit, 1e-9)
}
