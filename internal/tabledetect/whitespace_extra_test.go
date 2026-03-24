package tabledetect

import (
	"testing"

	"github.com/coregx/gxpdf/internal/extractor"
	"github.com/stretchr/testify/assert"
)

// makeTextElemAt creates a minimal TextElement positioned at (x, y) with given width.
func makeTextElemAt(x, y, width, height float64, text string) *extractor.TextElement {
	return &extractor.TextElement{
		Text:     text,
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		FontSize: 10,
	}
}

// TestDetectColumnsWithRulingLines_Empty verifies empty input returns empty slice.
func TestDetectColumnsWithRulingLines_Empty(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()
	result := wa.DetectColumnsWithRulingLines(nil, []float64{100, 200, 300})
	assert.Empty(t, result)
}

// TestDetectColumnsWithRulingLines_WithElements exercises the fallback branch.
func TestDetectColumnsWithRulingLines_WithElements(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()

	elements := []*extractor.TextElement{
		makeTextElemAt(50, 100, 80, 12, "Item"),
		makeTextElemAt(200, 100, 60, 12, "Price"),
		makeTextElemAt(50, 80, 80, 12, "Widget"),
		makeTextElemAt(200, 80, 60, 12, "9.99"),
	}

	rulingLines := []float64{150.0, 300.0}
	result := wa.DetectColumnsWithRulingLines(elements, rulingLines)
	// Should return at least some column boundaries.
	assert.NotNil(t, result)
}

// TestFindVerticalAlignments_FewElements verifies the function handles small input.
func TestFindVerticalAlignments_FewElements(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()

	// 2 elements — fewer than the minimum alignment threshold of 3.
	elements := []*extractor.TextElement{
		makeTextElemAt(50, 100, 80, 12, "A"),
		makeTextElemAt(50, 80, 80, 12, "B"),
	}

	result := wa.findVerticalAlignments(elements)
	// With only 2 elements aligned, should not meet the 3-element threshold.
	// Result may be nil or empty slice — both are valid.
	assert.True(t, result == nil || len(result) == 0)
}

// TestFindVerticalAlignments_ManyAligned ensures the 3-element threshold is met.
func TestFindVerticalAlignments_ManyAligned(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()

	// 4 elements all aligned at x=50 (left edge) — should trigger threshold.
	elements := []*extractor.TextElement{
		makeTextElemAt(50, 100, 80, 12, "Row1"),
		makeTextElemAt(50, 88, 80, 12, "Row2"),
		makeTextElemAt(50, 76, 80, 12, "Row3"),
		makeTextElemAt(50, 64, 80, 12, "Row4"),
	}

	result := wa.findVerticalAlignments(elements)
	assert.NotEmpty(t, result, "should find left-edge alignment at x=50")
}

// TestCountEdgeOccurrences_Left covers the left=true branch.
func TestCountEdgeOccurrences_Left(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()

	elements := []*extractor.TextElement{
		makeTextElemAt(50, 100, 80, 12, "A"),
		makeTextElemAt(50, 88, 80, 12, "B"),
		makeTextElemAt(200, 76, 60, 12, "C"),
	}

	counts := wa.countEdgeOccurrences(elements, true) // left edges
	assert.NotNil(t, counts)
	// Two elements at x=50; should appear as count=2 for the same key.
	var maxCount int
	for _, v := range counts {
		if v > maxCount {
			maxCount = v
		}
	}
	assert.GreaterOrEqual(t, maxCount, 2)
}

// TestCountEdgeOccurrences_Right covers the left=false (right edge) branch.
func TestCountEdgeOccurrences_Right(t *testing.T) {
	wa := NewDefaultWhitespaceAnalyzer()

	// Both elements have right edge at 130 (50+80).
	elements := []*extractor.TextElement{
		makeTextElemAt(50, 100, 80, 12, "A"),
		makeTextElemAt(50, 88, 80, 12, "B"),
		makeTextElemAt(50, 76, 80, 12, "C"),
	}

	counts := wa.countEdgeOccurrences(elements, false) // right edges
	assert.NotNil(t, counts)
}
