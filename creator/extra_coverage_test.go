package creator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coregx/gxpdf/internal/document"
	"github.com/coregx/gxpdf/internal/models/types"
)

// TestChapter_Height covers the Chapter.Height method.
func TestChapter_Height(t *testing.T) {
	ch := NewChapter("Test Chapter")

	ctx := &LayoutContext{
		PageWidth:  595,
		PageHeight: 842,
	}

	h := ch.Height(ctx)
	// The default ChapterStyle has FontSize 18, so heading height ~= 18*1.2 = 21.6.
	assert.Greater(t, h, 0.0)
}

func TestChapter_Height_WithSubChapters(t *testing.T) {
	ch := NewChapter("Parent")
	_ = ch.NewSubChapter("Child")

	ctx := &LayoutContext{PageWidth: 595, PageHeight: 842}
	h := ch.Height(ctx)
	assert.Greater(t, h, 0.0)
}

// TestStampAnnotation_SetNote covers the SetNote method.
func TestStampAnnotation_SetNote(t *testing.T) {
	stamp := NewStampAnnotation(100, 700, 150, 40, StampApproved)
	require.NotNil(t, stamp)

	result := stamp.SetNote("Approved by J. Doe on 2025-01-01")
	// Method should return *StampAnnotation for chaining.
	assert.Same(t, stamp, result)
}

// TestSizeFromMediaBox_StandardSizes covers the sizeFromMediaBox helper.
func TestSizeFromMediaBox_StandardSizes(t *testing.T) {
	tests := []struct {
		name    string
		llx     float64
		lly     float64
		urx     float64
		ury     float64
		wantDoc document.PageSize
	}{
		{"A4 exact", 0, 0, 595, 842, document.A4},
		{"A4 near tolerance", 0, 0, 597, 842, document.A4}, // within 5pt tolerance
		{"Letter exact", 0, 0, 612, 792, document.Letter},
		{"custom", 0, 0, 400, 600, document.Custom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := types.MustRectangle(tt.llx, tt.lly, tt.urx, tt.ury)
			size := sizeFromMediaBox(mb)
			assert.Equal(t, tt.wantDoc, size)
		})
	}
}

// TestMerger_AddDocument covers the addDocument internal helper.
func TestMerger_AddDocument(t *testing.T) {
	doc := document.NewDocument()
	_, err := doc.AddPage(document.A4)
	require.NoError(t, err)
	_, err = doc.AddPage(document.A4)
	require.NoError(t, err)

	m := NewMerger()
	err = m.addDocument(doc)
	require.NoError(t, err)
	assert.Len(t, m.pageInfos, 2)
}

// TestMerger_Close_Empty covers Close() on a fresh merger with no readers.
func TestMerger_Close_Empty(t *testing.T) {
	m := NewMerger()
	err := m.Close()
	assert.NoError(t, err)
}

// TestMergeDocuments_ErrorMessage verifies the no-documents error message.
func TestMergeDocuments_ErrorMessage(t *testing.T) {
	err := MergeDocuments("/tmp/out.pdf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no documents")
}

// TestSplitter_Close_NilReader covers Close() when reader is nil.
func TestSplitter_Close_NilReader(t *testing.T) {
	s := &Splitter{reader: nil}
	err := s.Close()
	assert.NoError(t, err)
}

// TestSplitter_SetFilenamePattern_Extra covers the SetFilenamePattern method.
func TestSplitter_SetFilenamePattern_Extra(t *testing.T) {
	s := &Splitter{filenamePattern: "page_%03d.pdf"}
	s.SetFilenamePattern("doc_%04d.pdf")
	assert.Equal(t, "doc_%04d.pdf", s.filenamePattern)
}

// TestSplitter_ValidateRanges_Errors covers the validateRanges error paths.
func TestSplitter_ValidateRanges_Errors(t *testing.T) {
	// Create a document with 3 pages.
	doc := document.NewDocument()
	for i := 0; i < 3; i++ {
		_, err := doc.AddPage(document.A4)
		require.NoError(t, err)
	}
	s := &Splitter{sourceDoc: doc, filenamePattern: "p_%03d.pdf"}

	tests := []struct {
		name   string
		ranges []PageRange
	}{
		{"start<1", []PageRange{{Start: 0, End: 2, Output: "out.pdf"}}},
		{"end<start", []PageRange{{Start: 3, End: 1, Output: "out.pdf"}}},
		{"end>pageCount", []PageRange{{Start: 1, End: 10, Output: "out.pdf"}}},
		{"emptyOutput", []PageRange{{Start: 1, End: 2, Output: ""}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateRanges(tt.ranges)
			assert.Error(t, err)
		})
	}
}

// TestSplitter_ValidateRanges_Valid verifies that valid ranges pass.
func TestSplitter_ValidateRanges_Valid(t *testing.T) {
	doc := document.NewDocument()
	for i := 0; i < 5; i++ {
		_, err := doc.AddPage(document.A4)
		require.NoError(t, err)
	}
	s := &Splitter{sourceDoc: doc, filenamePattern: "p_%03d.pdf"}

	ranges := []PageRange{
		{Start: 1, End: 3, Output: "out1.pdf"},
		{Start: 4, End: 5, Output: "out2.pdf"},
	}
	err := s.validateRanges(ranges)
	assert.NoError(t, err)
}

// TestSplitter_ValidatePageNumbers_Extra covers validatePageNumbers paths.
func TestSplitter_ValidatePageNumbers_Extra(t *testing.T) {
	doc := document.NewDocument()
	for i := 0; i < 3; i++ {
		_, err := doc.AddPage(document.A4)
		require.NoError(t, err)
	}
	s := &Splitter{sourceDoc: doc}

	// Invalid: page 0 (< 1).
	err := s.validatePageNumbers([]int{0})
	assert.Error(t, err)

	// Invalid: page 10 (> 3).
	err = s.validatePageNumbers([]int{10})
	assert.Error(t, err)

	// Valid: pages 1-3.
	err = s.validatePageNumbers([]int{1, 2, 3})
	assert.NoError(t, err)
}

// TestSplitter_ExtractPages_Empty covers the zero-pages error.
func TestSplitter_ExtractPages_Empty(t *testing.T) {
	doc := document.NewDocument()
	_, err := doc.AddPage(document.A4)
	require.NoError(t, err)

	s := &Splitter{sourceDoc: doc}
	_, err = s.ExtractPages()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pages")
}

// TestSplitter_SplitByRanges_Empty covers the zero-ranges error path.
func TestSplitter_SplitByRanges_Empty(t *testing.T) {
	doc := document.NewDocument()
	_, err := doc.AddPage(document.A4)
	require.NoError(t, err)

	s := &Splitter{sourceDoc: doc, filenamePattern: "p_%03d.pdf"}
	err = s.SplitByRanges()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no ranges")
}

// TestSplitter_SplitContext_EmptyDoc covers the zero-page guard in SplitContext.
func TestSplitter_SplitContext_EmptyDoc(t *testing.T) {
	doc := document.NewDocument()
	s := &Splitter{sourceDoc: doc, filenamePattern: "p_%03d.pdf"}

	err := s.Split("/tmp/output")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pages")
}
