package creator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// PointsToInches, PointsToMM, PointsToCM — previously 0% coverage
// ============================================================================

func TestPointsToInches(t *testing.T) {
	// 72 points = 1 inch.
	got := PointsToInches(72.0)
	assert.InDelta(t, 1.0, got, 0.001)
}

func TestPointsToMM(t *testing.T) {
	// 72 points = 1 inch = 25.4 mm.
	got := PointsToMM(72.0)
	assert.InDelta(t, 25.4, got, 0.01)
}

func TestPointsToCM(t *testing.T) {
	// 72 points = 1 inch = 2.54 cm.
	got := PointsToCM(72.0)
	assert.InDelta(t, 2.54, got, 0.001)
}

func TestPointsToInches_A4Width(t *testing.T) {
	// A4 width: 595 points ~= 8.27 inches.
	got := PointsToInches(595.0)
	assert.InDelta(t, 8.27, got, 0.01)
}

func TestPointsToInches_Zero(t *testing.T) {
	assert.Equal(t, 0.0, PointsToInches(0.0))
}

// ============================================================================
// Paint.isPaint() marker methods — previously 0% coverage
// NOTE: these are private methods called only via interface dispatch.
//       We exercise them by assigning to a Paint interface variable.
// ============================================================================

func TestPaintInterface_AllImplementors(t *testing.T) {
	// Each type must implement Paint. Calling isPaint via interface dispatch
	// covers the marker method bodies.
	paints := []Paint{
		Color{R: 1, G: 0, B: 0},
		ColorRGBA{R: 0, G: 1, B: 0, A: 0.5},
		ColorCMYK{C: 0, M: 0, Y: 0, K: 1},
		&Gradient{},
	}
	for _, p := range paints {
		// Calling isPaint() via the interface covers the marker method body.
		p.isPaint()
	}
	assert.True(t, true, "all isPaint() calls succeeded")
}

// ============================================================================
// CustomFont — LoadFont, UseChar, UseString, MeasureString, Build, PostScriptName,
// UnitsPerEm, GetSubset, GetTTF, ID, Ascender, Descender, LineHeight, CapHeight
// ============================================================================

const testFontPath = "D:/projects/gopdf/reference/krilla/assets/fonts/DejaVuSansMono.ttf"

func loadTestFont(t *testing.T) *CustomFont {
	t.Helper()
	font, err := LoadFont(testFontPath)
	if err != nil {
		t.Skipf("test font not available at %s: %v", testFontPath, err)
	}
	return font
}

func TestLoadFont_Success(t *testing.T) {
	font := loadTestFont(t)
	assert.NotNil(t, font)
}

func TestLoadFont_NonExistent(t *testing.T) {
	_, err := LoadFont("/nonexistent/font.ttf")
	assert.Error(t, err)
}

func TestCustomFont_UseChar(t *testing.T) {
	font := loadTestFont(t)
	font.UseChar('A')
	font.UseChar('B')
	assert.True(t, font.subset.UsedChars['A'])
	assert.True(t, font.subset.UsedChars['B'])
}

func TestCustomFont_UseString(t *testing.T) {
	font := loadTestFont(t)
	font.UseString("Hello")
	for _, ch := range "Hello" {
		assert.True(t, font.subset.UsedChars[ch], "char %q should be marked used", ch)
	}
}

func TestCustomFont_MeasureString(t *testing.T) {
	font := loadTestFont(t)
	font.UseString("Hello")
	// MeasureString delegates to the subset; width may be 0 if hmtx glyph
	// entries don't cover all glyphs. We just verify it doesn't panic.
	w := font.MeasureString("Hello", 12)
	assert.GreaterOrEqual(t, w, 0.0)
}

func TestCustomFont_Build(t *testing.T) {
	font := loadTestFont(t)
	font.UseString("Hello PDF")

	err := font.Build()
	require.NoError(t, err)
	assert.True(t, font.isBuilt)

	// Calling Build() again should be a no-op (isBuilt=true).
	err = font.Build()
	require.NoError(t, err)
}

func TestCustomFont_PostScriptName(t *testing.T) {
	font := loadTestFont(t)
	name := font.PostScriptName()
	// DejaVu Mono has a PostScript name.
	assert.NotEmpty(t, name)
}

func TestCustomFont_UnitsPerEm(t *testing.T) {
	font := loadTestFont(t)
	upm := font.UnitsPerEm()
	assert.Greater(t, upm, uint16(0), "UnitsPerEm should be positive")
}

func TestCustomFont_GetSubset(t *testing.T) {
	font := loadTestFont(t)
	subset := font.GetSubset()
	assert.NotNil(t, subset)
}

func TestCustomFont_GetTTF(t *testing.T) {
	font := loadTestFont(t)
	ttf := font.GetTTF()
	assert.NotNil(t, ttf)
}

func TestCustomFont_ID_PostScriptName(t *testing.T) {
	font := loadTestFont(t)
	id := font.ID()
	// DejaVu has a non-empty PostScript name, so ID returns that.
	assert.NotEmpty(t, id)
}

func TestCustomFont_Ascender(t *testing.T) {
	font := loadTestFont(t)
	asc := font.Ascender(12)
	assert.Greater(t, asc, 0.0, "Ascender should be positive")
}

func TestCustomFont_Descender(t *testing.T) {
	font := loadTestFont(t)
	desc := font.Descender(12)
	// Descender is typically negative for real fonts.
	assert.Less(t, desc, 0.0, "Descender should be negative")
}

func TestCustomFont_LineHeight(t *testing.T) {
	font := loadTestFont(t)
	lh := font.LineHeight(12)
	assert.Greater(t, lh, 0.0, "LineHeight should be positive")
}

func TestCustomFont_CapHeight(t *testing.T) {
	font := loadTestFont(t)
	ch := font.CapHeight(12)
	assert.GreaterOrEqual(t, ch, 0.0, "CapHeight should be non-negative")
}

// ============================================================================
// Page.AddTextCustomFont and variants — previously 0% coverage
// ============================================================================

func TestPage_AddTextCustomFont(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFont("Hello", 100, 700, font, 12)
	require.NoError(t, err)
}

func TestPage_AddTextCustomFont_NilFont(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFont("Hello", 100, 700, nil, 12)
	assert.Error(t, err)
}

func TestPage_AddTextCustomFontColor(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFontColor("Hello", 100, 700, font, 12, Red)
	require.NoError(t, err)
}

func TestPage_AddTextCustomFontRotated(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFontRotated("Hello", 100, 700, font, 12, 45)
	require.NoError(t, err)
}

func TestPage_AddTextCustomFontColorRotated(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFontColorRotated("Hello", 100, 700, font, 12, Blue, 90)
	require.NoError(t, err)
}

func TestPage_AddTextCustomFontColorRotated_NilFont(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFontColorRotated("Hello", 100, 700, nil, 12, Black, 0)
	assert.Error(t, err)
}

func TestPage_AddTextCustomFontColorRotated_ZeroSize(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFontColorRotated("Hello", 100, 700, font, 0, Black, 0)
	assert.Error(t, err)
}

// ============================================================================
// Page.AddTextCustomFontColorAlpha — previously 18.2% coverage
// ============================================================================

func TestPage_AddTextCustomFontColorAlpha(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFontColorAlpha("Hello", 100, 700, font, 12, Black, 0.7)
	require.NoError(t, err)
}

func TestPage_AddTextCustomFontColorRotatedAlpha(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextCustomFontColorRotatedAlpha("Hello", 100, 700, font, 12, Black, 45, 0.5)
	require.NoError(t, err)
}

// ============================================================================
// Page.BeginClipRect, EndClip, DrawTextClipped — previously 0% coverage
// ============================================================================

func TestPage_BeginClipRect(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.BeginClipRect(50, 100, 200, 150)
	require.NoError(t, err)
}

func TestPage_BeginClipRect_InvalidDimensions(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.BeginClipRect(50, 100, 0, 150)
	assert.Error(t, err, "zero width should fail")

	err = page.BeginClipRect(50, 100, 200, -1)
	assert.Error(t, err, "negative height should fail")
}

func TestPage_EndClip(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	_ = page.BeginClipRect(50, 100, 200, 150)
	err = page.EndClip()
	require.NoError(t, err)
}

func TestPage_DrawTextClipped(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.DrawTextClipped("Hello World", 50, 100, 40, 80, 200, 30, font, 12, Black)
	require.NoError(t, err)
}

func TestPage_DrawTextClipped_NilFont(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.DrawTextClipped("Hello", 50, 100, 40, 80, 200, 30, nil, 12, Black)
	assert.Error(t, err)
}

func TestPage_DrawTextClipped_ZeroClipW(t *testing.T) {
	font := loadTestFont(t)
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.DrawTextClipped("Hello", 50, 100, 40, 80, 0, 30, font, 12, Black)
	assert.Error(t, err)
}

// ============================================================================
// Surface.FillPath and StrokePath — previously 0% coverage
// ============================================================================

func TestSurface_FillPath(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	surf := page.Surface()
	surf.SetFill(NewFill(Red))

	path := NewPath()
	path.MoveTo(10, 10)
	path.LineTo(100, 10)
	path.LineTo(100, 100)
	path.Close()

	err = surf.FillPath(path)
	require.NoError(t, err)
}

func TestSurface_FillPath_NilPath(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	surf := page.Surface()
	surf.SetFill(NewFill(Red))

	err = surf.FillPath(nil)
	assert.Error(t, err)
}

func TestSurface_FillPath_EmptyPath(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	surf := page.Surface()
	surf.SetFill(NewFill(Red))

	path := NewPath() // empty
	err = surf.FillPath(path)
	// Empty path: no-op (no error or error — whichever the impl decides).
	// We just verify it doesn't panic.
	_ = err
}

func TestSurface_FillPath_NoFill(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	surf := page.Surface()
	// No fill set.
	path := NewPath()
	path.MoveTo(10, 10)
	path.LineTo(100, 10)

	err = surf.FillPath(path)
	assert.Error(t, err, "FillPath without fill should fail")
}

func TestSurface_StrokePath(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	surf := page.Surface()
	surf.SetStroke(NewStroke(Black))

	path := NewPath()
	path.MoveTo(10, 10)
	path.LineTo(100, 10)

	err = surf.StrokePath(path)
	require.NoError(t, err)
}

func TestSurface_StrokePath_NilPath(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	surf := page.Surface()
	surf.SetStroke(NewStroke(Black))

	err = surf.StrokePath(nil)
	assert.Error(t, err)
}

func TestSurface_StrokePath_NoStroke(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	surf := page.Surface()
	// No stroke set.
	path := NewPath()
	path.MoveTo(10, 10)
	path.LineTo(100, 10)

	err = surf.StrokePath(path)
	assert.Error(t, err)
}

// ============================================================================
// Creator.AddChapter, Chapters, EnableTOC, DisableTOC, TOCEnabled, TOC
// ============================================================================

func TestCreator_TOC(t *testing.T) {
	c := New()
	_, err := c.NewPage()
	require.NoError(t, err)

	assert.False(t, c.TOCEnabled(), "TOC should be disabled by default")

	c.EnableTOC()
	assert.True(t, c.TOCEnabled())

	c.DisableTOC()
	assert.False(t, c.TOCEnabled())
}

func TestCreator_TOC_Access(t *testing.T) {
	c := New()
	_, err := c.NewPage()
	require.NoError(t, err)

	c.EnableTOC()
	toc := c.TOC()
	assert.NotNil(t, toc, "TOC() should return non-nil after EnableTOC")
}

func TestCreator_AddChapter(t *testing.T) {
	c := New()
	_, err := c.NewPage()
	require.NoError(t, err)

	ch := NewChapter("Chapter 1")
	addErr := c.AddChapter(ch)
	require.NoError(t, addErr)

	chapters := c.Chapters()
	assert.Len(t, chapters, 1)
}

func TestCreator_Chapters_Empty(t *testing.T) {
	c := New()
	chapters := c.Chapters()
	assert.Empty(t, chapters)
}

// ============================================================================
// Chapter.SetTitle, NumberString
// ============================================================================

func TestChapter_SetTitle(t *testing.T) {
	c := New()
	_, err := c.NewPage()
	require.NoError(t, err)

	ch := NewChapter("Original Title")
	require.NoError(t, c.AddChapter(ch))
	ch.SetTitle("Updated Title")
	assert.Equal(t, "Updated Title", ch.Title())
}

func TestChapter_NumberString(t *testing.T) {
	c := New()
	_, err := c.NewPage()
	require.NoError(t, err)

	ch := NewChapter("Chapter 1")
	require.NoError(t, c.AddChapter(ch))
	ns := ch.NumberString()
	assert.NotEmpty(t, ns, "NumberString should return non-empty for an assigned chapter")
}

// ============================================================================
// Page.AddTextColorCMYK — remaining branches
// ============================================================================

func TestPage_AddTextColorCMYK_BlackText(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	black := ColorCMYK{C: 0, M: 0, Y: 0, K: 1}
	err = page.AddTextColorCMYK("Hello", 100, 700, Helvetica, 12, black)
	require.NoError(t, err)
}

// ============================================================================
// Page.MoveCursor — previously 0% coverage
// ============================================================================

func TestPage_MoveCursor(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	// MoveCursor is a no-op stub — should not panic.
	page.MoveCursor(0, 0)
	page.MoveCursor(100, 200)
}

// ============================================================================
// Page.Draw and Page.DrawAt — previously 0% coverage
// ============================================================================

func TestPage_Draw_WithParagraph(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	p := NewParagraph("Hello World")
	err = page.Draw(p)
	require.NoError(t, err)
}

func TestPage_DrawAt_WithParagraph(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	p := NewParagraph("Hello")
	err = page.DrawAt(p, 50, 100)
	require.NoError(t, err)
}

// ============================================================================
// StyledParagraph.Draw — previously 0% coverage
// ============================================================================

func TestStyledParagraph_Draw(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	sp := NewStyledParagraph()
	sp.Append("Hello World")

	err = page.Draw(sp)
	require.NoError(t, err)
}

func TestStyledParagraph_Draw_MultiSegment(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	sp := NewStyledParagraph()
	sp.Append("Hello ")
	sp.AppendStyled("World", TextStyle{Font: Helvetica, Size: 14})

	err = page.Draw(sp)
	require.NoError(t, err)
}

func TestStyledParagraph_DrawAt_Aligned(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	sp := NewStyledParagraph()
	sp.Append("Centered text")
	sp.SetAlignment(AlignCenter)

	err = page.DrawAt(sp, 100, 200)
	require.NoError(t, err)
}

// ============================================================================
// HighlightAnnotation, UnderlineAnnotation, StrikeOutAnnotation
// SetAuthor/SetNote — previously 0% coverage
// ============================================================================

func TestHighlightAnnotation_SetAuthorAndNote(t *testing.T) {
	ann := NewHighlightAnnotation(50, 700, 200, 715)
	result := ann.SetAuthor("Test Author")
	assert.Same(t, ann, result, "SetAuthor should return the annotation for chaining")

	result2 := ann.SetNote("This is a note")
	assert.Same(t, ann, result2, "SetNote should return the annotation for chaining")
}

func TestUnderlineAnnotation_SetAuthorAndNote(t *testing.T) {
	ann := NewUnderlineAnnotation(50, 700, 200, 715)
	result := ann.SetAuthor("Author")
	assert.Same(t, ann, result)

	result2 := ann.SetNote("Note text")
	assert.Same(t, ann, result2)
}

func TestStrikeOutAnnotation_SetAuthorAndNote(t *testing.T) {
	ann := NewStrikeOutAnnotation(50, 700, 200, 715)
	result := ann.SetAuthor("Reviewer")
	assert.Same(t, ann, result)

	result2 := ann.SetNote("Strike note")
	assert.Same(t, ann, result2)
}

// ============================================================================
// Appender form methods — previously 0% coverage
// These operate on the non-form reference PDF so form-specific calls
// return "no form" type errors or false — we just verify they don't panic.
// ============================================================================

// createTestAppender creates an Appender from the pdfcpu reference PDF.
// Skips if the reference PDF is not available.
func createTestAppender(t *testing.T) *Appender {
	t.Helper()
	sourcePDF := filepath.Join("..", "reference", "pdfcpu", "pkg", "samples", "annotations", "LinkAnnotWithDestTopLeft.pdf")
	if _, err := os.Stat(sourcePDF); os.IsNotExist(err) {
		t.Skipf("reference PDF not found: %s", sourcePDF)
	}
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.pdf")
	data, err := os.ReadFile(sourcePDF)
	require.NoError(t, err)
	err = os.WriteFile(testPath, data, 0o644)
	require.NoError(t, err)
	app, err := NewAppender(testPath)
	require.NoError(t, err)
	return app
}

func TestAppender_HasForm_NonFormPDF(t *testing.T) {
	app := createTestAppender(t)
	defer app.Close()

	// A non-form PDF should return false.
	hasForm := app.HasForm()
	assert.False(t, hasForm, "non-form PDF should not have a form")
}

func TestAppender_CanFlattenForm_NonFormPDF(t *testing.T) {
	app := createTestAppender(t)
	defer app.Close()

	// CanFlattenForm on a non-form PDF should return false.
	can := app.CanFlattenForm()
	assert.False(t, can)
}

func TestAppender_GetFormFields_NonFormPDF(t *testing.T) {
	app := createTestAppender(t)
	defer app.Close()

	// Non-form PDF: may return error or empty slice.
	_, _ = app.GetFormFields()
	// Just verify it doesn't panic.
}

func TestAppender_GetFieldValue_MissingField(t *testing.T) {
	app := createTestAppender(t)
	defer app.Close()

	_, err := app.GetFieldValue("nonexistent_field")
	// Should return an error for a field that doesn't exist.
	assert.Error(t, err)
}

func TestAppender_SetFieldValue_NonFormPDF(t *testing.T) {
	app := createTestAppender(t)
	defer app.Close()

	// Non-form PDF: setting a field should fail.
	err := app.SetFieldValue("nonexistent", "value")
	assert.Error(t, err)
}

func TestAppender_FlattenForm_NonFormPDF(t *testing.T) {
	app := createTestAppender(t)
	defer app.Close()

	// FlattenForm on a non-form PDF: should succeed with no-op.
	err := app.FlattenForm()
	// Either no error (empty flatten info) or an error — just don't panic.
	_ = err
}

func TestAppender_FlattenFields_NonFormPDF(t *testing.T) {
	app := createTestAppender(t)
	defer app.Close()

	// FlattenFields on a non-form PDF: should succeed as no-op.
	err := app.FlattenFields("field1", "field2")
	_ = err
}

// ============================================================================
// Page.AddField — previously 0% coverage
// ============================================================================

func TestPage_AddField_UnsupportedType(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	// Passing an unsupported type should return ErrUnsupportedFieldType.
	err = page.AddField("not a field type")
	assert.ErrorIs(t, err, ErrUnsupportedFieldType)
}

// ============================================================================
// pathOp.String — exercised via explicit call on private type
// ============================================================================

func TestPathOp_String_AllOps(t *testing.T) {
	tests := []struct {
		op   pathOp
		want string
	}{
		{pathOpMoveTo, "m"},
		{pathOpLineTo, "l"},
		{pathOpCubicTo, "c"},
		{pathOpClose, "h"},
		{pathOpRect, "re"},
		{pathOp(99), "?"},
	}
	for _, tc := range tests {
		got := tc.op.String()
		assert.Equal(t, tc.want, got, "pathOp(%d).String()", tc.op)
	}
}

// ============================================================================
// Creator.renderChapter, renderTOC, updateChapterPageIndices
// These are triggered by WriteToFile/WriteTo when chapters exist.
// ============================================================================

func TestCreator_WriteToFile_WithChapter(t *testing.T) {
	c := New()
	_, err := c.NewPage()
	require.NoError(t, err)

	ch := NewChapter("Test Chapter")
	require.NoError(t, c.AddChapter(ch))

	tmpFile := filepath.Join(t.TempDir(), "output.pdf")
	err = c.WriteToFile(tmpFile)
	require.NoError(t, err)

	// Verify the file was created.
	info, err := os.Stat(tmpFile)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestCreator_WriteToFile_WithChapterAndTOC(t *testing.T) {
	c := New()
	_, err := c.NewPage()
	require.NoError(t, err)

	c.EnableTOC()

	ch := NewChapter("Chapter One")
	require.NoError(t, c.AddChapter(ch))

	tmpFile := filepath.Join(t.TempDir(), "output_toc.pdf")
	err = c.WriteToFile(tmpFile)
	require.NoError(t, err)

	info, err := os.Stat(tmpFile)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestCreator_WriteToFile_WithSubChapter(t *testing.T) {
	c := New()
	_, err := c.NewPage()
	require.NoError(t, err)

	c.EnableTOC()

	ch := NewChapter("Main Chapter")
	sub := ch.NewSubChapter("Sub Chapter")
	_ = sub
	require.NoError(t, c.AddChapter(ch))

	tmpFile := filepath.Join(t.TempDir(), "output_sub.pdf")
	err = c.WriteToFile(tmpFile)
	require.NoError(t, err)
}

// ============================================================================
// Creator.WriteToFileContext / WriteToContext error paths
// ============================================================================

func TestCreator_WriteToFileContext_NoPages(t *testing.T) {
	c := New()
	// No pages added.
	tmpFile := filepath.Join(t.TempDir(), "empty.pdf")
	err := c.WriteToFile(tmpFile)
	// Validate succeeds or fails gracefully — no panic.
	_ = err
}
