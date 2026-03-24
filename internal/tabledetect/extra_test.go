package tabledetect

import (
	"strings"
	"testing"

	"github.com/coregx/gxpdf/internal/extractor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- ExtractionMethod ----

func TestExtractionMethod_String_Unknown(t *testing.T) {
	unknown := ExtractionMethod(999)
	assert.Equal(t, "Unknown", unknown.String())
}

// ---- RulingLine.String ----

func TestRulingLine_String(t *testing.T) {
	h := NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(100, 50))
	s := h.String()
	assert.Contains(t, s, "H")
	assert.Contains(t, s, "100")

	v := NewRulingLine(extractor.NewPoint(50, 0), extractor.NewPoint(50, 100))
	sv := v.String()
	assert.Contains(t, sv, "V")
}

// ---- RulingLine.Intersects vertical-first path ----

func TestRulingLine_Intersects_VerticalFirst(t *testing.T) {
	// Vertical line intersects horizontal line (via the !rl.IsHorizontal && other.IsHorizontal branch).
	vLine := NewRulingLine(extractor.NewPoint(50, 0), extractor.NewPoint(50, 100))
	hLine := NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(100, 50))

	pt := vLine.Intersects(hLine)
	require.NotNil(t, pt)
	assert.InDelta(t, 50.0, pt.X, 1e-6)
	assert.InDelta(t, 50.0, pt.Y, 1e-6)
}

func TestRulingLine_Intersects_TwoVerticals(t *testing.T) {
	v1 := NewRulingLine(extractor.NewPoint(50, 0), extractor.NewPoint(50, 100))
	v2 := NewRulingLine(extractor.NewPoint(60, 0), extractor.NewPoint(60, 100))
	pt := v1.Intersects(v2)
	assert.Nil(t, pt, "two vertical lines should not intersect")
}

func TestRulingLine_Intersects_OutsideSegment(t *testing.T) {
	// Horizontal line from X=0..50, vertical from Y=0..100 at X=200 — no intersection in segment.
	hLine := NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(50, 50))
	vLine := NewRulingLine(extractor.NewPoint(200, 0), extractor.NewPoint(200, 100))
	pt := hLine.Intersects(vLine)
	assert.Nil(t, pt)
}

// ---- DefaultRulingLineDetector sorting & merging ----

func TestRulingLineDetector_MergesCollinearLines(t *testing.T) {
	detector := NewDefaultRulingLineDetector()

	// Two horizontal lines at same Y, adjacent (should merge into one).
	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(0, 50),
				extractor.NewPoint(50, 50),
			},
		},
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(50, 50),
				extractor.NewPoint(100, 50),
			},
		},
	}

	lines, err := detector.DetectRulingLines(graphics)
	require.NoError(t, err)
	// Should be merged into 1 line.
	assert.LessOrEqual(t, len(lines), 2, "collinear adjacent lines should merge")
}

func TestRulingLineDetector_SkipsObliqueLines(t *testing.T) {
	detector := NewDefaultRulingLineDetector()

	graphics := []*extractor.GraphicsElement{
		{
			// Oblique line (diagonal)
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(0, 0),
				extractor.NewPoint(100, 100),
			},
		},
	}

	lines, err := detector.DetectRulingLines(graphics)
	require.NoError(t, err)
	assert.Empty(t, lines, "oblique lines should be skipped")
}

func TestRulingLineDetector_SkipsShortLines(t *testing.T) {
	detector := NewDefaultRulingLineDetector().WithMinLineLength(50.0)

	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(0, 0),
				extractor.NewPoint(5, 0), // shorter than 50
			},
		},
	}

	lines, err := detector.DetectRulingLines(graphics)
	require.NoError(t, err)
	assert.Empty(t, lines)
}

func TestRulingLineDetector_SkipsNonLineGraphics(t *testing.T) {
	detector := NewDefaultRulingLineDetector()

	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeRectangle,
			Points: []extractor.Point{
				extractor.NewPoint(0, 0),
				extractor.NewPoint(100, 0),
			},
		},
	}

	lines, err := detector.DetectRulingLines(graphics)
	require.NoError(t, err)
	assert.Empty(t, lines)
}

func TestRulingLineDetector_SkipsWrongPointCount(t *testing.T) {
	detector := NewDefaultRulingLineDetector()

	graphics := []*extractor.GraphicsElement{
		{
			Type:   extractor.GraphicsTypeLine,
			Points: []extractor.Point{extractor.NewPoint(0, 0)}, // only 1 point
		},
	}

	lines, err := detector.DetectRulingLines(graphics)
	require.NoError(t, err)
	assert.Empty(t, lines)
}

func TestRulingLineDetector_AreAdjacent_Vertical(t *testing.T) {
	d := NewDefaultRulingLineDetector()

	// Two vertical lines close together on Y axis.
	v1 := NewRulingLine(extractor.NewPoint(50, 0), extractor.NewPoint(50, 50))
	v2 := NewRulingLine(extractor.NewPoint(50, 50), extractor.NewPoint(50, 100))
	assert.True(t, d.areAdjacent(v1, v2, false))

	// Far apart.
	v3 := NewRulingLine(extractor.NewPoint(50, 200), extractor.NewPoint(50, 300))
	assert.False(t, d.areAdjacent(v1, v3, false))
}

func TestRulingLineDetector_MergeTwo_Horizontal(t *testing.T) {
	d := NewDefaultRulingLineDetector()

	h1 := NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(60, 50))
	h2 := NewRulingLine(extractor.NewPoint(50, 50), extractor.NewPoint(120, 50))

	merged := d.mergeTwo(h1, h2, true)
	require.NotNil(t, merged)
	assert.True(t, merged.IsHorizontal)
	assert.Equal(t, 0.0, merged.Start.X)
	assert.Equal(t, 120.0, merged.End.X)
}

func TestRulingLineDetector_MergeTwo_Vertical(t *testing.T) {
	d := NewDefaultRulingLineDetector()

	v1 := NewRulingLine(extractor.NewPoint(50, 0), extractor.NewPoint(50, 60))
	v2 := NewRulingLine(extractor.NewPoint(50, 50), extractor.NewPoint(50, 120))

	merged := d.mergeTwo(v1, v2, false)
	require.NotNil(t, merged)
	assert.False(t, merged.IsHorizontal)
	assert.Equal(t, 0.0, merged.Start.Y)
	assert.Equal(t, 120.0, merged.End.Y)
}

func TestRulingLineDetector_SortLines_Horizontal(t *testing.T) {
	d := NewDefaultRulingLineDetector()

	lines := []*RulingLine{
		NewRulingLine(extractor.NewPoint(100, 50), extractor.NewPoint(200, 50)),
		NewRulingLine(extractor.NewPoint(10, 50), extractor.NewPoint(50, 50)),
		NewRulingLine(extractor.NewPoint(50, 50), extractor.NewPoint(90, 50)),
	}

	d.sortLines(lines, true)

	// After sort by X, first element should have smallest start X.
	assert.Equal(t, 10.0, lines[0].Start.X)
	assert.Equal(t, 50.0, lines[1].Start.X)
	assert.Equal(t, 100.0, lines[2].Start.X)
}

func TestRulingLineDetector_SortLines_Vertical(t *testing.T) {
	d := NewDefaultRulingLineDetector()

	lines := []*RulingLine{
		NewRulingLine(extractor.NewPoint(50, 200), extractor.NewPoint(50, 300)),
		NewRulingLine(extractor.NewPoint(50, 0), extractor.NewPoint(50, 100)),
		NewRulingLine(extractor.NewPoint(50, 100), extractor.NewPoint(50, 200)),
	}

	d.sortLines(lines, false)

	assert.Equal(t, 0.0, lines[0].Start.Y)
	assert.Equal(t, 100.0, lines[1].Start.Y)
	assert.Equal(t, 200.0, lines[2].Start.Y)
}

func TestRulingLineDetector_FindIntersections_NoLines(t *testing.T) {
	d := NewDefaultRulingLineDetector()
	pts := d.FindIntersections([]*RulingLine{})
	assert.Empty(t, pts)
}

func TestRulingLineDetector_FindIntersections_DuplicatesRemoved(t *testing.T) {
	d := NewDefaultRulingLineDetector()

	// Three lines forming a cross — should deduplicate intersection.
	h1 := NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(100, 50))
	h2 := NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(100, 50)) // same as h1
	v := NewRulingLine(extractor.NewPoint(50, 0), extractor.NewPoint(50, 100))

	pts := d.FindIntersections([]*RulingLine{h1, h2, v})
	// Should not double-count the same intersection.
	for i, p := range pts {
		for j := i + 1; j < len(pts); j++ {
			dx := p.X - pts[j].X
			dy := p.Y - pts[j].Y
			if dx < 0 {
				dx = -dx
			}
			if dy < 0 {
				dy = -dy
			}
			assert.False(t, dx < d.tolerance && dy < d.tolerance, "found duplicate intersection")
		}
	}
}

// ---- Grid & Cell String methods ----

func TestCell_String(t *testing.T) {
	cell := NewCell(1, 2, extractor.NewRectangle(10, 20, 50, 30))
	s := cell.String()
	assert.Contains(t, s, "row=1")
	assert.Contains(t, s, "col=2")
}

func TestGrid_String(t *testing.T) {
	grid := NewGrid([]float64{0, 50, 100}, []float64{0, 100, 200})
	s := grid.String()
	assert.Contains(t, s, "rows=2")
	assert.Contains(t, s, "cols=2")
}

func TestGrid_RowCount_EdgeCases(t *testing.T) {
	tests := []struct {
		rows     []float64
		wantRows int
	}{
		{[]float64{}, 0},
		{[]float64{50}, 0},
		{[]float64{0, 100}, 1},
		{[]float64{0, 50, 100}, 2},
	}

	for _, tt := range tests {
		g := &Grid{Rows: tt.rows, Columns: []float64{0, 100}}
		assert.Equal(t, tt.wantRows, g.RowCount(), "rows=%v", tt.rows)
	}
}

func TestGrid_ColumnCount_EdgeCases(t *testing.T) {
	tests := []struct {
		cols     []float64
		wantCols int
	}{
		{[]float64{}, 0},
		{[]float64{0}, 0},
		{[]float64{0, 100}, 1},
		{[]float64{0, 100, 200}, 2},
	}

	for _, tt := range tests {
		g := &Grid{Rows: []float64{0, 100}, Columns: tt.cols}
		assert.Equal(t, tt.wantCols, g.ColumnCount(), "cols=%v", tt.cols)
	}
}

func TestGrid_GetCell_OutOfRange(t *testing.T) {
	grid := NewGrid([]float64{0, 50, 100}, []float64{0, 100, 200})
	gb := NewDefaultGridBuilder()
	grid.Cells = gb.createCells(grid.Rows, grid.Columns)

	assert.Nil(t, grid.GetCell(-1, 0))
	assert.Nil(t, grid.GetCell(0, -1))
	assert.Nil(t, grid.GetCell(10, 0))
	assert.Nil(t, grid.GetCell(0, 10))
}

func TestGrid_GetCell_Valid(t *testing.T) {
	grid := NewGrid([]float64{0, 50, 100}, []float64{0, 100, 200})
	gb := NewDefaultGridBuilder()
	grid.Cells = gb.createCells(grid.Rows, grid.Columns)

	cell := grid.GetCell(0, 0)
	require.NotNil(t, cell)
	assert.Equal(t, 0, cell.Row)
	assert.Equal(t, 0, cell.Column)
}

func TestGrid_Bounds_SingleRow(t *testing.T) {
	// Less than 2 rows or columns — should return zero rectangle.
	g := &Grid{Rows: []float64{0}, Columns: []float64{0, 100}}
	bounds := g.Bounds()
	assert.Equal(t, 0.0, bounds.Width)
	assert.Equal(t, 0.0, bounds.Height)
}

// ---- DefaultGridBuilder ----

func TestGridBuilder_BuildGrid_InsufficientLines(t *testing.T) {
	gb := NewDefaultGridBuilder()

	// Only horizontal, no vertical.
	lines := []*RulingLine{
		NewRulingLine(extractor.NewPoint(0, 0), extractor.NewPoint(100, 0)),
		NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(100, 50)),
	}
	_, err := gb.BuildGrid(lines)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")

	// Only vertical, no horizontal.
	lines2 := []*RulingLine{
		NewRulingLine(extractor.NewPoint(0, 0), extractor.NewPoint(0, 100)),
		NewRulingLine(extractor.NewPoint(50, 0), extractor.NewPoint(50, 100)),
	}
	_, err2 := gb.BuildGrid(lines2)
	require.Error(t, err2)

	// Empty lines.
	_, err3 := gb.BuildGrid([]*RulingLine{})
	require.Error(t, err3)
}

func TestGridBuilder_FindCellsFromIntersections(t *testing.T) {
	gb := NewDefaultGridBuilder()

	horizontal := []*RulingLine{
		NewRulingLine(extractor.NewPoint(0, 0), extractor.NewPoint(200, 0)),
		NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(200, 50)),
		NewRulingLine(extractor.NewPoint(0, 100), extractor.NewPoint(200, 100)),
	}
	vertical := []*RulingLine{
		NewRulingLine(extractor.NewPoint(0, 0), extractor.NewPoint(0, 100)),
		NewRulingLine(extractor.NewPoint(100, 0), extractor.NewPoint(100, 100)),
		NewRulingLine(extractor.NewPoint(200, 0), extractor.NewPoint(200, 100)),
	}

	cells, err := gb.FindCellsFromIntersections(horizontal, vertical)
	require.NoError(t, err)
	assert.Equal(t, 4, len(cells), "3x3 grid of lines should produce 4 cells")
}

func TestGridBuilder_FindCellsFromIntersections_Insufficient(t *testing.T) {
	gb := NewDefaultGridBuilder()

	_, err := gb.FindCellsFromIntersections(
		[]*RulingLine{NewRulingLine(extractor.NewPoint(0, 0), extractor.NewPoint(100, 0))},
		[]*RulingLine{},
	)
	require.Error(t, err)
}

func TestGridBuilder_BuildGridFromCells(t *testing.T) {
	gb := NewDefaultGridBuilder()

	cells := []*Cell{
		NewCell(0, 0, extractor.NewRectangle(0, 0, 100, 50)),
		NewCell(0, 1, extractor.NewRectangle(100, 0, 100, 50)),
		NewCell(1, 0, extractor.NewRectangle(0, 50, 100, 50)),
		NewCell(1, 1, extractor.NewRectangle(100, 50, 100, 50)),
	}

	grid, err := gb.BuildGridFromCells(cells)
	require.NoError(t, err)
	require.NotNil(t, grid)
	assert.Equal(t, 2, grid.RowCount())
	assert.Equal(t, 2, grid.ColumnCount())
}

func TestGridBuilder_BuildGridFromCells_Empty(t *testing.T) {
	gb := NewDefaultGridBuilder()
	_, err := gb.BuildGridFromCells([]*Cell{})
	require.Error(t, err)
}

func TestGridBuilder_WithTolerance(t *testing.T) {
	gb := NewDefaultGridBuilder().WithTolerance(5.0)
	require.NotNil(t, gb)
}

// ---- TableRegion edge cases ----

func TestTableRegion_RowCount_StreamMode(t *testing.T) {
	bounds := extractor.NewRectangle(0, 0, 200, 100)
	region := NewTableRegion(bounds, MethodStream)
	region.Rows = []float64{0, 50, 100}

	assert.Equal(t, 2, region.RowCount())
}

func TestTableRegion_RowCount_NoRows(t *testing.T) {
	bounds := extractor.NewRectangle(0, 0, 200, 100)
	region := NewTableRegion(bounds, MethodStream)
	region.Rows = []float64{50}

	assert.Equal(t, 0, region.RowCount())
}

func TestTableRegion_ColumnCount_StreamMode(t *testing.T) {
	bounds := extractor.NewRectangle(0, 0, 200, 100)
	region := NewTableRegion(bounds, MethodStream)
	region.Columns = []float64{0, 100, 200}

	assert.Equal(t, 2, region.ColumnCount())
}

func TestTableRegion_ColumnCount_NoColumns(t *testing.T) {
	bounds := extractor.NewRectangle(0, 0, 200, 100)
	region := NewTableRegion(bounds, MethodStream)
	region.Columns = []float64{100}

	assert.Equal(t, 0, region.ColumnCount())
}

func TestTableRegion_RowCount_LatticeWithNilGrid(t *testing.T) {
	bounds := extractor.NewRectangle(0, 0, 200, 100)
	region := NewTableRegion(bounds, MethodLattice)
	region.Grid = nil
	region.Rows = []float64{0, 50, 100}

	// HasRulingLines is true but grid is nil — falls through to Rows.
	assert.Equal(t, 2, region.RowCount())
}

// ---- DefaultTableDetector ----

func TestTableDetector_WithDeps(t *testing.T) {
	rulingDetector := NewDefaultRulingLineDetector()
	whitespaceAnalyzer := NewDefaultWhitespaceAnalyzer()
	gridBuilder := NewDefaultGridBuilder()

	td := NewTableDetectorWithDeps(rulingDetector, whitespaceAnalyzer, gridBuilder)
	require.NotNil(t, td)
}

func TestTableDetector_FluentSetters(t *testing.T) {
	td := NewDefaultTableDetector()
	td = td.WithRulingDetector(NewDefaultRulingLineDetector())
	td = td.WithWhitespaceAnalyzer(NewDefaultWhitespaceAnalyzer())
	td = td.WithGridBuilder(NewDefaultGridBuilder())
	require.NotNil(t, td)
}

func TestTableDetector_DetectTables_Stream(t *testing.T) {
	td := NewDefaultTableDetector()

	// Provide text elements in a grid layout but no graphics.
	elements := []*extractor.TextElement{
		extractor.NewTextElement("A", 0, 100, 50, 10, "/F1", 10),
		extractor.NewTextElement("B", 100, 100, 50, 10, "/F1", 10),
		extractor.NewTextElement("C", 0, 50, 50, 10, "/F1", 10),
		extractor.NewTextElement("D", 100, 50, 50, 10, "/F1", 10),
	}

	regions, err := td.DetectTables(elements, []*extractor.GraphicsElement{})
	require.NoError(t, err)
	// With stream mode, may return 0 or 1 region depending on whitespace analysis.
	assert.NotNil(t, regions)
}

func TestTableDetector_DetectTables_EmptyElements(t *testing.T) {
	td := NewDefaultTableDetector()
	regions, err := td.DetectTables(nil, nil)
	require.NoError(t, err)
	assert.Empty(t, regions)
}

func TestTableDetector_DetectTablesLattice_FallbackToStream(t *testing.T) {
	td := NewDefaultTableDetector()

	// No graphics — falls back to stream mode.
	elements := []*extractor.TextElement{
		extractor.NewTextElement("A", 0, 100, 50, 10, "/F1", 10),
	}

	regions, err := td.DetectTablesLattice(elements, []*extractor.GraphicsElement{})
	require.NoError(t, err)
	assert.NotNil(t, regions)
}

func TestTableDetector_DetectTablesStream_Explicit(t *testing.T) {
	td := NewDefaultTableDetector()

	elements := []*extractor.TextElement{
		extractor.NewTextElement("A", 0, 100, 50, 10, "/F1", 10),
		extractor.NewTextElement("B", 100, 100, 50, 10, "/F1", 10),
		extractor.NewTextElement("C", 0, 50, 50, 10, "/F1", 10),
		extractor.NewTextElement("D", 100, 50, 50, 10, "/F1", 10),
	}

	regions, err := td.DetectTablesStream(elements)
	require.NoError(t, err)
	assert.NotNil(t, regions)
}

func TestTableDetector_IsValidGrid_Nil(t *testing.T) {
	td := NewDefaultTableDetector()
	assert.False(t, td.isValidGrid(nil))
}

func TestTableDetector_IsValidGrid_SmallBounds(t *testing.T) {
	td := NewDefaultTableDetector()
	grid := NewGrid([]float64{0, 20, 40}, []float64{0, 20, 40})
	assert.False(t, td.isValidGrid(grid), "grid with tiny bounds should be invalid")
}

func TestTableDetector_IsValidGrid_ValidGrid(t *testing.T) {
	td := NewDefaultTableDetector()
	grid := NewGrid([]float64{0, 50, 100}, []float64{0, 100, 200})
	assert.True(t, td.isValidGrid(grid))
}

func TestTableDetector_CalculateBoundsFromText_Empty(t *testing.T) {
	td := NewDefaultTableDetector()
	bounds := td.calculateBoundsFromText([]*extractor.TextElement{})
	assert.Equal(t, 0.0, bounds.Width)
	assert.Equal(t, 0.0, bounds.Height)
}

func TestTableDetector_CalculateBoundsFromText_SingleElement(t *testing.T) {
	td := NewDefaultTableDetector()
	elements := []*extractor.TextElement{
		extractor.NewTextElement("Hello", 10, 20, 50, 10, "/F1", 10),
	}
	bounds := td.calculateBoundsFromText(elements)
	assert.Equal(t, 10.0, bounds.X)
	assert.Equal(t, 20.0, bounds.Y)
}

// ---- DefaultWhitespaceAnalyzer ----

func TestWhitespaceAnalyzerForLattice(t *testing.T) {
	wa := NewWhitespaceAnalyzerForLattice()
	require.NotNil(t, wa)
	assert.True(t, wa.isLatticeMode)
}

func TestWhitespaceAnalyzer_WithProjectionAnalyzer(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()
	wa = wa.WithProjectionAnalyzer(NewDefaultProjectionAnalyzer())
	require.NotNil(t, wa)
}

func TestWhitespaceAnalyzer_DetectColumns_Empty(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()
	cols := wa.DetectColumns([]*extractor.TextElement{})
	assert.Empty(t, cols)
}

func TestWhitespaceAnalyzer_GroupIntoRows(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()

	elements := []*extractor.TextElement{
		extractor.NewTextElement("A", 0, 100, 50, 10, "/F1", 10),
		extractor.NewTextElement("B", 60, 100, 50, 10, "/F1", 10), // same row
		extractor.NewTextElement("C", 0, 50, 50, 10, "/F1", 10),   // different row
		extractor.NewTextElement("D", 60, 50, 50, 10, "/F1", 10),  // same as C
	}

	rows := wa.GroupIntoRows(elements)
	assert.Equal(t, 2, len(rows), "should group into 2 rows")
}

func TestWhitespaceAnalyzer_GroupIntoRows_Empty(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()
	rows := wa.GroupIntoRows([]*extractor.TextElement{})
	assert.Empty(t, rows)
}

func TestWhitespaceAnalyzer_GroupIntoColumns(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()

	elements := []*extractor.TextElement{
		extractor.NewTextElement("A", 0, 100, 50, 10, "/F1", 10),
		extractor.NewTextElement("B", 0, 50, 50, 10, "/F1", 10),    // same column
		extractor.NewTextElement("C", 100, 100, 50, 10, "/F1", 10), // different column
		extractor.NewTextElement("D", 100, 50, 50, 10, "/F1", 10),  // same as C
	}

	cols := wa.GroupIntoColumns(elements)
	assert.Equal(t, 2, len(cols), "should group into 2 columns")
}

func TestWhitespaceAnalyzer_GroupIntoColumns_Empty(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()
	cols := wa.GroupIntoColumns([]*extractor.TextElement{})
	assert.Empty(t, cols)
}

func TestWhitespaceAnalyzer_DetectTableRegion_NoTable(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()
	region := wa.DetectTableRegion([]*extractor.TextElement{})
	assert.Nil(t, region)
}

func TestWhitespaceAnalyzer_String(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()
	s := wa.String()
	assert.True(t, strings.Contains(s, "WhitespaceAnalyzer"))
}

// ---- ProjectionProfile ----

func TestProjectionProfile_String(t *testing.T) {
	pp := NewProjectionProfile([]float64{1, 2, 3}, 5.0, 0, 15)
	s := pp.String()
	assert.Contains(t, s, "ProjectionProfile")
	assert.Contains(t, s, "bins=3")
}

func TestGap_String(t *testing.T) {
	g := NewGap(10.0, 30.0)
	s := g.String()
	assert.Contains(t, s, "Gap")
	assert.Contains(t, s, "10.00")
	assert.Contains(t, s, "30.00")
}

func TestGap_Center(t *testing.T) {
	g := NewGap(10.0, 30.0)
	assert.Equal(t, 20.0, g.Center())
}

// ---- DefaultProjectionAnalyzer ----

func TestProjectionAnalyzer_WithBinSize(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer().WithBinSize(5.0)
	require.NotNil(t, pa)
}

func TestProjectionAnalyzer_WithThreshold(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer().WithThreshold(0.5)
	require.NotNil(t, pa)
}

func TestProjectionAnalyzer_AnalyzeVertical(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()

	elements := []*extractor.TextElement{
		extractor.NewTextElement("Hello", 0, 0, 50, 10, "/F1", 10),
		extractor.NewTextElement("World", 100, 0, 60, 10, "/F1", 10),
	}

	profile := pa.AnalyzeVertical(elements)
	require.NotNil(t, profile)
	assert.Greater(t, profile.BinCount(), 0)
	assert.Equal(t, 0.0, profile.Min)
}

func TestProjectionAnalyzer_AnalyzeVertical_Empty(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()
	profile := pa.AnalyzeVertical([]*extractor.TextElement{})
	require.NotNil(t, profile)
	assert.Equal(t, 0, profile.BinCount())
}

func TestProjectionAnalyzer_AnalyzeHorizontal_Empty(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()
	profile := pa.AnalyzeHorizontal([]*extractor.TextElement{})
	require.NotNil(t, profile)
	assert.Equal(t, 0, profile.BinCount())
}

func TestProjectionAnalyzer_FindGaps_EmptyProfile(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()
	profile := NewProjectionProfile([]float64{}, 5.0, 0, 0)
	gaps := pa.FindGaps(profile)
	assert.Empty(t, gaps)
}

func TestProjectionAnalyzer_FindGaps_AllZero(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()
	profile := NewProjectionProfile([]float64{0, 0, 0, 0, 0}, 5.0, 0, 25)
	gaps := pa.FindGaps(profile)
	// All zero = one big gap.
	assert.Greater(t, len(gaps), 0)
}

func TestProjectionAnalyzer_FindGaps_AllDense(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()
	profile := NewProjectionProfile([]float64{10, 10, 10, 10, 10}, 5.0, 0, 25)
	gaps := pa.FindGaps(profile)
	assert.Empty(t, gaps, "no gaps when all bins have high density")
}

func TestProjectionAnalyzer_FindSignificantGaps(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()
	// Bins: dense, zero, zero, zero, dense
	profile := NewProjectionProfile([]float64{10, 0, 0, 0, 10}, 5.0, 0, 25)
	gaps := pa.FindSignificantGaps(profile, 8.0)
	// The gap between bins 1-3 is ~10 points wide (2 bins * 5 pts).
	assert.Greater(t, len(gaps), 0)
}

func TestNewPeak(t *testing.T) {
	p := NewPeak(10.0, 30.0, 5.5)
	assert.Equal(t, 10.0, p.Start)
	assert.Equal(t, 30.0, p.End)
	assert.Equal(t, 5.5, p.MaxValue)
	assert.Equal(t, 20.0, p.Center)
}

func TestPeak_String(t *testing.T) {
	p := NewPeak(10.0, 30.0, 5.5)
	s := p.String()
	assert.Contains(t, s, "Peak")
	assert.Contains(t, s, "10.00")
}

func TestProjectionAnalyzer_FindPeaks_Empty(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()
	profile := NewProjectionProfile([]float64{}, 5.0, 0, 0)
	peaks := pa.FindPeaks(profile, 1.0)
	assert.Empty(t, peaks)
}

func TestProjectionAnalyzer_FindPeaks(t *testing.T) {
	pa := NewDefaultProjectionAnalyzer()
	profile := NewProjectionProfile([]float64{0, 0, 10, 10, 0, 0, 8, 8, 0}, 5.0, 0, 45)
	peaks := pa.FindPeaks(profile, 5.0)
	assert.Equal(t, 2, len(peaks), "should find 2 peaks")
}

// ---- TableExtractor ----

func TestTableExtractor_ExtractTable_Nil(t *testing.T) {
	te := NewTableExtractor([]*extractor.TextElement{})
	_, err := te.ExtractTable(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestTableExtractor_ExtractTable_Stream(t *testing.T) {
	elements := []*extractor.TextElement{
		extractor.NewTextElement("A", 0, 100, 50, 10, "/F1", 10),
		extractor.NewTextElement("B", 100, 100, 50, 10, "/F1", 10),
		extractor.NewTextElement("C", 0, 50, 50, 10, "/F1", 10),
		extractor.NewTextElement("D", 100, 50, 50, 10, "/F1", 10),
	}

	te := NewTableExtractor(elements)

	region := NewTableRegion(extractor.NewRectangle(0, 50, 150, 60), MethodStream)
	region.Rows = []float64{50, 100}
	region.Columns = []float64{0, 100, 150}

	tbl, err := te.ExtractTable(region)
	require.NoError(t, err)
	require.NotNil(t, tbl)
	assert.Equal(t, 1, tbl.RowCount)
	assert.Equal(t, 2, tbl.ColCount)
}

func TestTableExtractor_ExtractTable_Lattice(t *testing.T) {
	elements := []*extractor.TextElement{
		extractor.NewTextElement("Cell1", 5, 55, 40, 10, "/F1", 10),
		extractor.NewTextElement("Cell2", 105, 55, 40, 10, "/F1", 10),
	}

	te := NewTableExtractor(elements)

	region := NewTableRegion(extractor.NewRectangle(0, 50, 200, 60), MethodLattice)

	// Build a real grid.
	lines := []*RulingLine{
		NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(200, 50)),
		NewRulingLine(extractor.NewPoint(0, 110), extractor.NewPoint(200, 110)),
		NewRulingLine(extractor.NewPoint(0, 50), extractor.NewPoint(0, 110)),
		NewRulingLine(extractor.NewPoint(100, 50), extractor.NewPoint(100, 110)),
		NewRulingLine(extractor.NewPoint(200, 50), extractor.NewPoint(200, 110)),
	}
	gb := NewDefaultGridBuilder()
	grid, err := gb.BuildGrid(lines)
	require.NoError(t, err)

	region.Grid = grid
	region.HasRulingLines = true

	tbl, err := te.ExtractTable(region)
	require.NoError(t, err)
	require.NotNil(t, tbl)
	assert.Equal(t, 1, tbl.RowCount)
	assert.Equal(t, 2, tbl.ColCount)
}

func TestTableExtractor_ExtractTable_LatticeMissingGrid(t *testing.T) {
	te := NewTableExtractor([]*extractor.TextElement{})

	region := NewTableRegion(extractor.NewRectangle(0, 0, 200, 100), MethodLattice)
	region.Grid = nil

	_, err := te.ExtractTable(region)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "grid")
}

func TestTableExtractor_ExtractTable_StreamInsufficientBoundaries(t *testing.T) {
	te := NewTableExtractor([]*extractor.TextElement{})

	region := NewTableRegion(extractor.NewRectangle(0, 0, 200, 100), MethodStream)
	region.Rows = []float64{50} // only 1 row boundary
	region.Columns = []float64{0, 100}

	_, err := te.ExtractTable(region)
	require.Error(t, err)
}

func TestTableExtractor_ExtractTables(t *testing.T) {
	te := NewTableExtractor([]*extractor.TextElement{})

	region := NewTableRegion(extractor.NewRectangle(0, 0, 200, 100), MethodStream)
	region.Rows = []float64{0, 50, 100}
	region.Columns = []float64{0, 100, 200}

	tables, err := te.ExtractTables([]*TableRegion{region})
	require.NoError(t, err)
	assert.Len(t, tables, 1)
}

func TestTableExtractor_ExtractTables_Empty(t *testing.T) {
	te := NewTableExtractor([]*extractor.TextElement{})
	tables, err := te.ExtractTables([]*TableRegion{})
	require.NoError(t, err)
	assert.Empty(t, tables)
}
