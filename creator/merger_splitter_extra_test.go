package creator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makePDF creates a simple valid PDF with the given number of pages at the given path.
func makePDF(t *testing.T, path string, pageCount int) {
	t.Helper()
	c := New()
	c.SetPageSize(A4)
	for i := 0; i < pageCount; i++ {
		page, err := c.NewPage()
		require.NoError(t, err)
		require.NoError(t, page.AddText("Page content", 100, 700, Helvetica, 12))
	}
	require.NoError(t, c.WriteToFile(path))
}

// ============================================================================
// Merger — AddPages / addPagesFromFile
// ============================================================================

func TestMerger_AddPages_Valid(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 5)

	m := NewMerger()
	err := m.AddPages(src, 1, 3, 5)
	require.NoError(t, err)
	assert.Len(t, m.pageInfos, 3)

	defer func() { _ = m.Close() }()
}

func TestMerger_AddPages_NoPageNumbers(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 3)

	m := NewMerger()
	err := m.AddPages(src) // no page numbers
	assert.Error(t, err)
}

func TestMerger_AddPages_InvalidPageNum(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 3)

	m := NewMerger()
	defer func() { _ = m.Close() }()
	err := m.AddPages(src, 1, 99) // page 99 doesn't exist
	assert.Error(t, err)
}

func TestMerger_AddPages_NonExistentFile(t *testing.T) {
	m := NewMerger()
	err := m.AddPages("/nonexistent/file.pdf", 1)
	assert.Error(t, err)
}

// ============================================================================
// Merger — AddPageRange
// ============================================================================

func TestMerger_AddPageRange_Valid(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 10)

	m := NewMerger()
	err := m.AddPageRange(src, 3, 7) // pages 3-7
	require.NoError(t, err)
	assert.Len(t, m.pageInfos, 5)

	defer func() { _ = m.Close() }()
}

func TestMerger_AddPageRange_StartLessThanOne(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 5)

	m := NewMerger()
	err := m.AddPageRange(src, 0, 3)
	assert.Error(t, err)
}

func TestMerger_AddPageRange_EndBeforeStart(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 5)

	m := NewMerger()
	err := m.AddPageRange(src, 4, 2)
	assert.Error(t, err)
}

func TestMerger_AddPageRange_EndExceedsPageCount(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 3)

	m := NewMerger()
	defer func() { _ = m.Close() }()
	err := m.AddPageRange(src, 1, 10)
	assert.Error(t, err)
}

func TestMerger_AddPageRange_NonExistentFile(t *testing.T) {
	m := NewMerger()
	err := m.AddPageRange("/nonexistent.pdf", 1, 3)
	assert.Error(t, err)
}

// ============================================================================
// Merger — AddAllPages
// ============================================================================

func TestMerger_AddAllPages_Valid(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 4)

	m := NewMerger()
	err := m.AddAllPages(src)
	require.NoError(t, err)
	assert.Len(t, m.pageInfos, 4)

	defer func() { _ = m.Close() }()
}

func TestMerger_AddAllPages_NonExistentFile(t *testing.T) {
	m := NewMerger()
	err := m.AddAllPages("/nonexistent.pdf")
	assert.Error(t, err)
}

// ============================================================================
// Merger — Write / copyPagesToOutput / writeOutput
// ============================================================================

func TestMerger_Write_Valid(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	out := filepath.Join(dir, "merged.pdf")
	makePDF(t, src, 3)

	m := NewMerger()
	require.NoError(t, m.AddAllPages(src))
	err := m.Write(out)
	require.NoError(t, err)

	info, err := os.Stat(out)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestMerger_Write_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	src1 := filepath.Join(dir, "s1.pdf")
	src2 := filepath.Join(dir, "s2.pdf")
	out := filepath.Join(dir, "merged.pdf")

	makePDF(t, src1, 2)
	makePDF(t, src2, 3)

	m := NewMerger()
	require.NoError(t, m.AddAllPages(src1))
	require.NoError(t, m.AddAllPages(src2))
	err := m.Write(out)
	require.NoError(t, err)

	_, err = os.Stat(out)
	assert.NoError(t, err)
}

// TestMerger_Write_InvalidOutputPath verifies error on bad path.
func TestMerger_Write_InvalidOutputPath(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 1)

	m := NewMerger()
	require.NoError(t, m.AddAllPages(src))
	err := m.Write("/nonexistent/deeply/nested/out.pdf")
	assert.Error(t, err)
}

// ============================================================================
// Splitter — Split / SplitByRanges / ExtractPages
// ============================================================================

func TestSplitter_Split_Valid(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	outDir := filepath.Join(dir, "out")
	require.NoError(t, os.MkdirAll(outDir, 0755))
	makePDF(t, src, 3)

	s, err := NewSplitter(src)
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	err = s.Split(outDir)
	require.NoError(t, err)

	// Verify 3 output files.
	for i := 1; i <= 3; i++ {
		name := filepath.Join(outDir, "page_00"+string(rune('0'+i))+".pdf")
		_, statErr := os.Stat(name)
		assert.NoError(t, statErr, "page_%03d.pdf should exist", i)
	}
}

func TestSplitter_SetFilenamePattern_Applied(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	outDir := filepath.Join(dir, "out")
	require.NoError(t, os.MkdirAll(outDir, 0755))
	makePDF(t, src, 2)

	s, err := NewSplitter(src)
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	s.SetFilenamePattern("doc_%04d.pdf")
	require.NoError(t, s.Split(outDir))

	_, err = os.Stat(filepath.Join(outDir, "doc_0001.pdf"))
	assert.NoError(t, err, "doc_0001.pdf should exist")
	_, err = os.Stat(filepath.Join(outDir, "doc_0002.pdf"))
	assert.NoError(t, err, "doc_0002.pdf should exist")
}

func TestSplitter_SplitByRanges_Valid(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 6)

	s, err := NewSplitter(src)
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	part1 := filepath.Join(dir, "part1.pdf")
	part2 := filepath.Join(dir, "part2.pdf")

	err = s.SplitByRanges(
		PageRange{Start: 1, End: 3, Output: part1},
		PageRange{Start: 4, End: 6, Output: part2},
	)
	require.NoError(t, err)

	_, err = os.Stat(part1)
	assert.NoError(t, err)
	_, err = os.Stat(part2)
	assert.NoError(t, err)
}

func TestSplitter_SplitByRanges_InvalidRange_ValidSplitter(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 3)

	s, err := NewSplitter(src)
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	// End exceeds page count.
	err = s.SplitByRanges(
		PageRange{Start: 1, End: 10, Output: filepath.Join(dir, "out.pdf")},
	)
	assert.Error(t, err)
}

func TestSplitter_ExtractPages_Valid(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 5)

	s, err := NewSplitter(src)
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	doc, err := s.ExtractPages(1, 3, 5)
	require.NoError(t, err)
	assert.Equal(t, 3, doc.PageCount())
}

func TestSplitter_ExtractPages_InvalidPage_ValidSplitter(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 3)

	s, err := NewSplitter(src)
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	_, err = s.ExtractPages(99)
	assert.Error(t, err)
}

func TestSplitter_Close_WithReader(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pdf")
	makePDF(t, src, 1)

	s, err := NewSplitter(src)
	require.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)
}
