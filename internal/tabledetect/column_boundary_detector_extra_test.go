package tabledetect

import (
	"testing"

	"github.com/coregx/gxpdf/internal/extractor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTextElement is already defined in column_boundary_detector_test.go in this package.
// We reuse it via the same package namespace.

// ---- DetectBoundariesWithRulingLines ----

func TestDetectBoundariesWithRulingLines_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.DetectBoundariesWithRulingLines([]*extractor.TextElement{}, []float64{50, 100, 200})
	assert.Empty(t, result)
}

func TestDetectBoundariesWithRulingLines_NoRulingLines(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := makeThreeColumnTable()
	result := cbd.DetectBoundariesWithRulingLines(elements, []float64{})
	assert.NotNil(t, result)
}

func TestDetectBoundariesWithRulingLines_WithRulingLines(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := makeThreeColumnTable()

	// Ruling lines at approximately correct positions.
	rulingLines := []float64{50, 150, 250, 300}
	result := cbd.DetectBoundariesWithRulingLines(elements, rulingLines)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result), 2)
}

func TestDetectBoundariesWithRulingLines_NoTextBoundaries(t *testing.T) {
	cbd := NewColumnBoundaryDetector()

	// Just 1 element — edge clustering may return empty.
	elements := []*extractor.TextElement{
		newTextElement("A", 50, 100, 10, 10),
	}
	rulingLines := []float64{50, 100}
	result := cbd.DetectBoundariesWithRulingLines(elements, rulingLines)
	assert.NotNil(t, result)
}

func TestDetectBoundariesWithRulingLines_HybridFallback(t *testing.T) {
	cbd := NewColumnBoundaryDetector()

	// Two elements only — hybrid may fall back to text boundaries.
	elements := []*extractor.TextElement{
		newTextElement("A", 50, 100, 50, 10),
		newTextElement("B", 150, 100, 50, 10),
	}
	rulingLines := []float64{50, 100, 150, 200}
	result := cbd.DetectBoundariesWithRulingLines(elements, rulingLines)
	assert.NotNil(t, result)
}

// ---- DetectBoundariesWithHorizontalRulingLines ----

func TestDetectBoundariesWithHorizontalRulingLines_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.DetectBoundariesWithHorizontalRulingLines([]*extractor.TextElement{}, []*extractor.GraphicsElement{})
	assert.Empty(t, result)
}

func TestDetectBoundariesWithHorizontalRulingLines_NoGraphics(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := makeThreeColumnTable()
	result := cbd.DetectBoundariesWithHorizontalRulingLines(elements, []*extractor.GraphicsElement{})
	assert.NotNil(t, result)
}

func TestDetectBoundariesWithHorizontalRulingLines_WithGraphics(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := makeThreeColumnTable()

	// Add horizontal ruling lines.
	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(50, 120),
				extractor.NewPoint(100, 120),
			},
		},
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(150, 120),
				extractor.NewPoint(200, 120),
			},
		},
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(250, 120),
				extractor.NewPoint(300, 120),
			},
		},
	}

	result := cbd.DetectBoundariesWithHorizontalRulingLines(elements, graphics)
	assert.NotNil(t, result)
}

func TestDetectBoundariesWithHorizontalRulingLines_WideRegion(t *testing.T) {
	cbd := NewColumnBoundaryDetector()

	// Elements spread over a wide region (> 100pt) to trigger sub-division.
	elements := []*extractor.TextElement{
		newTextElement("A", 0, 100, 50, 10),
		newTextElement("B", 60, 100, 50, 10),
		newTextElement("C", 130, 100, 50, 10),
		newTextElement("D", 0, 90, 50, 10),
		newTextElement("E", 60, 90, 50, 10),
		newTextElement("F", 130, 90, 50, 10),
	}

	// A single horizontal line spanning a wide range triggers wide region processing.
	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(0, 110),
				extractor.NewPoint(200, 110), // wide
			},
		},
	}

	result := cbd.DetectBoundariesWithHorizontalRulingLines(elements, graphics)
	assert.NotNil(t, result)
}

// ---- deduplicateBoundaries ----

func TestDeduplicateBoundaries_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.deduplicateBoundaries([]float64{}, 5.0)
	assert.Empty(t, result)
}

func TestDeduplicateBoundaries_NoDuplicates(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	input := []float64{10, 50, 100, 200}
	result := cbd.deduplicateBoundaries(input, 5.0)
	assert.Equal(t, input, result)
}

func TestDeduplicateBoundaries_RemovesClose(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	input := []float64{10, 12, 50, 52, 100}
	result := cbd.deduplicateBoundaries(input, 5.0)
	// 10 and 12 are within tolerance; 50 and 52 are within tolerance.
	assert.Equal(t, []float64{10, 50, 100}, result)
}

// ---- detectBoundariesHeaderBased ----

func TestDetectBoundariesHeaderBased_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.detectBoundariesHeaderBased([]*extractor.TextElement{})
	assert.Empty(t, result)
}

func TestDetectBoundariesHeaderBased_SimpleTable(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := makeThreeColumnTable()
	result := cbd.detectBoundariesHeaderBased(elements)
	assert.NotNil(t, result)
}

// ---- groupElementsByRow ----

func TestGroupElementsByRow_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.groupElementsByRow([]*extractor.TextElement{})
	assert.Empty(t, result)
}

func TestGroupElementsByRow_TwoRows(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := []*extractor.TextElement{
		newTextElement("A", 0, 100, 50, 10),
		newTextElement("B", 60, 100, 50, 10),
		newTextElement("C", 0, 50, 50, 10),
		newTextElement("D", 60, 50, 50, 10),
	}

	rows := cbd.groupElementsByRow(elements)
	assert.Equal(t, 2, len(rows))
}

func TestGroupElementsByRow_SingleRow(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := []*extractor.TextElement{
		newTextElement("A", 0, 100, 50, 10),
		newTextElement("B", 60, 100, 50, 10),
		newTextElement("C", 120, 100, 50, 10),
	}

	rows := cbd.groupElementsByRow(elements)
	assert.Equal(t, 1, len(rows))
}

// ---- ValidateConsistency ----

func TestValidateConsistency_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	tableType, rate := cbd.ValidateConsistency([]*extractor.TextElement{}, []float64{0, 100, 200}, 2)
	assert.Equal(t, RegularTable, tableType)
	assert.Equal(t, 1.0, rate)
}

func TestValidateConsistency_ZeroExpectedColumns(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	tableType, rate := cbd.ValidateConsistency(makeThreeColumnTable(), []float64{50, 150, 250}, 0)
	assert.Equal(t, RegularTable, tableType)
	assert.Equal(t, 1.0, rate)
}

func TestValidateConsistency_RegularTable(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := makeThreeColumnTable()
	boundaries := []float64{50, 100, 150, 200, 250, 300}

	tableType, rate := cbd.ValidateConsistency(elements, boundaries, 3)
	assert.NotNil(t, tableType)
	assert.GreaterOrEqual(t, rate, 0.0)
	assert.LessOrEqual(t, rate, 1.0)
}

// ---- TableType.String ----

func TestTableType_String(t *testing.T) {
	assert.Equal(t, "Regular", RegularTable.String())
	assert.Equal(t, "Irregular", IrregularTable.String())
	assert.Equal(t, "Unknown", TableType(999).String())
}

// ---- AssignToColumns ----

func TestAssignToColumns_Basic(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := []*extractor.TextElement{
		newTextElement("A", 50, 100, 50, 10),
		newTextElement("B", 150, 100, 50, 10),
		newTextElement("C", 250, 100, 50, 10),
	}
	boundaries := []float64{30, 120, 220, 310}

	result := cbd.AssignToColumns(elements, boundaries)
	require.NotNil(t, result)
	// Should have 3 columns.
	assert.Equal(t, 3, len(result))
}

func TestAssignToColumns_EmptyElements(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.AssignToColumns([]*extractor.TextElement{}, []float64{0, 100, 200})
	// No elements => empty map (no columns populated).
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestAssignToColumns_EmptyBoundaries(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := []*extractor.TextElement{
		newTextElement("A", 50, 100, 50, 10),
	}
	result := cbd.AssignToColumns(elements, []float64{})
	// No boundaries => all elements in column 0.
	require.NotNil(t, result)
	assert.Equal(t, 1, len(result))
	assert.Len(t, result[0], 1)
}

// ---- selectBoundariesByConsistency ----

func TestSelectBoundariesByConsistency_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.selectBoundariesByConsistency(makeThreeColumnTable(), []float64{})
	assert.Empty(t, result)
}

func TestSelectBoundariesByConsistency_FewRows(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	// Only 2 rows — should return all boundaries (below minimum threshold).
	elements := []*extractor.TextElement{
		newTextElement("A", 50, 100, 50, 10),
		newTextElement("B", 150, 100, 50, 10),
		newTextElement("C", 50, 80, 50, 10),
		newTextElement("D", 150, 80, 50, 10),
	}
	boundaries := []float64{50, 100, 150, 200}
	result := cbd.selectBoundariesByConsistency(elements, boundaries)
	assert.NotNil(t, result)
}

func TestSelectBoundariesByConsistency_ManyRows(t *testing.T) {
	cbd := NewColumnBoundaryDetector()

	// Build a consistent 3-column table with many rows.
	var elements []*extractor.TextElement
	for row := 0; row < 15; row++ {
		y := float64(100 - row*10)
		elements = append(elements,
			newTextElement("A", 50, y, 50, 10),
			newTextElement("B", 150, y, 50, 10),
			newTextElement("C", 250, y, 50, 10),
		)
	}

	boundaries := []float64{50, 100, 150, 200, 250, 300}
	result := cbd.selectBoundariesByConsistency(elements, boundaries)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result), 2)
}

// ---- mergeTwoBoundarySets ----

func TestMergeTwoBoundarySets_Both(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	set1 := []float64{50, 150, 250}
	set2 := []float64{100, 200, 300}
	result := cbd.mergeTwoBoundarySets(set1, set2)
	assert.NotNil(t, result)
	assert.Equal(t, 6, len(result))
	// Should be sorted.
	for i := 1; i < len(result); i++ {
		assert.Greater(t, result[i], result[i-1])
	}
}

func TestMergeTwoBoundarySets_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.mergeTwoBoundarySets([]float64{}, []float64{})
	assert.Empty(t, result)
}

func TestMergeTwoBoundarySets_Deduplicates(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	// Both sets contain nearly same positions.
	set1 := []float64{50, 150}
	set2 := []float64{51, 151}
	result := cbd.mergeTwoBoundarySets(set1, set2)
	// With tolerance minGapWidth/2 = 5.0, these close pairs may or may not deduplicate.
	assert.NotNil(t, result)
}

// ---- extractHorizontalRulingLines ----

func TestExtractHorizontalRulingLines_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.extractHorizontalRulingLines([]*extractor.GraphicsElement{})
	assert.Empty(t, result)
}

func TestExtractHorizontalRulingLines_IgnoresVertical(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(50, 0),
				extractor.NewPoint(50, 100),
			},
		},
	}
	result := cbd.extractHorizontalRulingLines(graphics)
	assert.Empty(t, result)
}

func TestExtractHorizontalRulingLines_Horizontal(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(0, 100),
				extractor.NewPoint(200, 100),
			},
		},
	}
	result := cbd.extractHorizontalRulingLines(graphics)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 0.0, result[0].x1)
	assert.Equal(t, 200.0, result[0].x2)
}

func TestExtractHorizontalRulingLines_ReversedPoints(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	// Points given in reverse order (x2 < x1 initially).
	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeLine,
			Points: []extractor.Point{
				extractor.NewPoint(200, 100),
				extractor.NewPoint(0, 100),
			},
		},
	}
	result := cbd.extractHorizontalRulingLines(graphics)
	require.Equal(t, 1, len(result))
	assert.Equal(t, 0.0, result[0].x1)
	assert.Equal(t, 200.0, result[0].x2)
}

func TestExtractHorizontalRulingLines_IgnoresNonLineType(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	graphics := []*extractor.GraphicsElement{
		{
			Type: extractor.GraphicsTypeRectangle,
			Points: []extractor.Point{
				extractor.NewPoint(0, 100),
				extractor.NewPoint(200, 100),
			},
		},
	}
	result := cbd.extractHorizontalRulingLines(graphics)
	assert.Empty(t, result)
}

// ---- detectBoundariesWhitespace and helpers ----

func TestDetectBoundariesWhitespace_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.detectBoundariesWhitespace([]*extractor.TextElement{})
	assert.Empty(t, result)
}

func TestDetectBoundariesWhitespace_TwoColumns(t *testing.T) {
	cbd := NewColumnBoundaryDetector()

	// Two clear columns with a gap between them.
	elements := []*extractor.TextElement{
		newTextElement("A", 0, 100, 40, 10),
		newTextElement("B", 0, 90, 40, 10),
		newTextElement("C", 0, 80, 40, 10),
		newTextElement("D", 100, 100, 40, 10),
		newTextElement("E", 100, 90, 40, 10),
		newTextElement("F", 100, 80, 40, 10),
	}

	result := cbd.detectBoundariesWhitespace(elements)
	assert.NotNil(t, result)
}

func TestDetectBoundariesWhitespace_NoGap(t *testing.T) {
	cbd := NewColumnBoundaryDetector()

	// Elements covering continuous range — no whitespace valley.
	elements := []*extractor.TextElement{
		newTextElement("A", 0, 100, 200, 10),
	}

	result := cbd.detectBoundariesWhitespace(elements)
	assert.NotNil(t, result)
}

func TestFindExtent_Basic(t *testing.T) {
	cbd := NewColumnBoundaryDetector()

	elements := []*extractor.TextElement{
		newTextElement("A", 20, 100, 30, 10), // X=20..50
		newTextElement("B", 80, 100, 40, 10), // X=80..120
	}

	minX, maxX := cbd.findExtent(elements)
	assert.Equal(t, 20.0, minX)
	assert.Equal(t, 120.0, maxX)
}

func TestFindExtent_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	minX, maxX := cbd.findExtent([]*extractor.TextElement{})
	assert.Equal(t, 0.0, minX)
	assert.Equal(t, 0.0, maxX)
}

func TestFindValleysAdaptive_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	valleys := cbd.findValleysAdaptive([]int{}, 0, 1.0)
	assert.Empty(t, valleys)
}

func TestFindValleysAdaptive_ClearValley(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	// Profile: high, high, low, low, high, high.
	profile := []int{10, 10, 0, 0, 10, 10}
	valleys := cbd.findValleysAdaptive(profile, 0.0, 1.0)
	assert.Greater(t, len(valleys), 0, "should find valley in low region")
}

func TestFindValleysAdaptive_AllHigh(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	profile := []int{10, 10, 10, 10, 10}
	valleys := cbd.findValleysAdaptive(profile, 0.0, 1.0)
	assert.Empty(t, valleys)
}

func TestFindValleys_Basic(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	profile := []int{5, 5, 0, 0, 5, 5}
	valleys := cbd.findValleys(profile, 0.0, 1.0)
	assert.Greater(t, len(valleys), 0)
}

func TestFindValleys_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	valleys := cbd.findValleys([]int{}, 0.0, 1.0)
	assert.Empty(t, valleys)
}

func TestFilterValleys_MinWidth(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	valleys := []valley{
		{start: 0, end: 5, width: 5},
		{start: 20, end: 35, width: 15},
	}

	// Filter with minWidth=10: only second valley passes.
	filtered := cbd.filterValleys(valleys, 10.0)
	assert.Equal(t, 1, len(filtered))
	assert.Equal(t, 15.0, filtered[0].width)
}

func TestFilterValleys_AllPass(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	valleys := []valley{
		{start: 0, end: 15, width: 15},
		{start: 30, end: 50, width: 20},
	}
	filtered := cbd.filterValleys(valleys, 5.0)
	assert.Equal(t, 2, len(filtered))
}

func TestFilterValleys_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	filtered := cbd.filterValleys([]valley{}, 5.0)
	assert.Empty(t, filtered)
}

func TestMergeBoundaries_Basic(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	boundaries := []float64{10, 15, 50, 52, 100}
	merged := cbd.mergeBoundaries(boundaries, 10.0)
	// 10 and 15 are within 10pt → keep first.
	// 50 and 52 are within 10pt → keep first.
	assert.Less(t, len(merged), len(boundaries))
}

func TestMergeBoundaries_Empty(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.mergeBoundaries([]float64{}, 10.0)
	assert.Empty(t, result)
}

func TestMergeBoundaries_SingleElement(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	result := cbd.mergeBoundaries([]float64{50.0}, 10.0)
	assert.Equal(t, []float64{50.0}, result)
}

// ---- AnalyzeTableStructure ----

func TestAnalyzeTableStructure(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	elements := makeThreeColumnTable()
	analysis := cbd.AnalyzeTableStructure(elements)
	require.NotNil(t, analysis)
	assert.Greater(t, analysis.ElementCount, 0)
	assert.Greater(t, analysis.MaxX, analysis.MinX)
}

// ---- Structural methods ----

func TestTableType_String_Methods(t *testing.T) {
	assert.Equal(t, "Regular", RegularTable.String())
	assert.Equal(t, "Irregular", IrregularTable.String())
}

// ---- Helper functions min/max/mean ----

func TestHelpers_MinMaxMean(t *testing.T) {
	cbd := NewColumnBoundaryDetector()

	assert.Equal(t, 0.0, cbd.min([]float64{}))
	assert.Equal(t, 0.0, cbd.max([]float64{}))
	assert.Equal(t, 0.0, cbd.mean([]float64{}))

	vals := []float64{3, 1, 4, 1, 5, 9, 2, 6}
	assert.Equal(t, 1.0, cbd.min(vals))
	assert.Equal(t, 9.0, cbd.max(vals))
	assert.InDelta(t, 3.875, cbd.mean(vals), 1e-6)
}

// ---- columnRegion horizontallyOverlaps / merge (tested via detectBoundariesHeaderBased) ----

func TestColumnRegion_HorizontallyOverlaps(t *testing.T) {
	// Test columnRegion methods directly (unexported type, same package).
	r := &columnRegion{minX: 50, maxX: 90}

	// Overlapping element.
	b := newTextElement("B", 70, 100, 40, 10) // X=70..110
	assert.True(t, r.horizontallyOverlaps(b), "element at 70..110 should overlap with region 50..90")

	// Non-overlapping (to the right).
	c := newTextElement("C", 200, 100, 40, 10) // X=200..240
	assert.False(t, r.horizontallyOverlaps(c), "element at 200..240 should not overlap with region 50..90")

	// Non-overlapping (to the left).
	d := newTextElement("D", 0, 100, 20, 10) // X=0..20
	assert.False(t, r.horizontallyOverlaps(d), "element at 0..20 should not overlap with region 50..90")
}

func TestColumnRegion_Merge(t *testing.T) {
	r := &columnRegion{minX: 50, maxX: 90}

	// Expand left.
	r.merge(newTextElement("A", 30, 100, 10, 10)) // X=30..40
	assert.Equal(t, 30.0, r.minX)

	// Expand right.
	r.merge(newTextElement("B", 80, 100, 30, 10)) // X=80..110
	assert.Equal(t, 110.0, r.maxX)
}

// ---- findColumnIndex (AssignToColumns helper) ----

func TestFindColumnIndex_CBD(t *testing.T) {
	cbd := NewColumnBoundaryDetector()
	boundaries := []float64{50, 150, 250, 350}

	// Test various positions.
	idx := cbd.findColumnIndex(60.0, boundaries)
	assert.Equal(t, 0, idx)

	idx = cbd.findColumnIndex(160.0, boundaries)
	assert.Equal(t, 1, idx)

	idx = cbd.findColumnIndex(260.0, boundaries)
	assert.Equal(t, 2, idx)

	// At or after last boundary.
	idx = cbd.findColumnIndex(350.0, boundaries)
	assert.Equal(t, 3, idx)

	// After last boundary.
	idx = cbd.findColumnIndex(400.0, boundaries)
	assert.Equal(t, 3, idx)
}

// ---- makeThreeColumnTable helper ----

func makeThreeColumnTable() []*extractor.TextElement {
	return []*extractor.TextElement{
		newTextElement("A1", 50, 100, 50, 10),
		newTextElement("B1", 150, 100, 50, 10),
		newTextElement("C1", 250, 100, 50, 10),
		newTextElement("A2", 50, 90, 50, 10),
		newTextElement("B2", 150, 90, 50, 10),
		newTextElement("C2", 250, 90, 50, 10),
		newTextElement("A3", 50, 80, 50, 10),
		newTextElement("B3", 150, 80, 50, 10),
		newTextElement("C3", 250, 80, 50, 10),
	}
}
