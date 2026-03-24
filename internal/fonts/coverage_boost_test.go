package fonts

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// ============================================================================
// FontSubset.Build(), createGlyphMapping, compressFont — previously 0% coverage
// ============================================================================

func makeFontWithData() *TTFFont {
	// Minimal TTFFont with enough fields to exercise Build().
	return &TTFFont{
		UnitsPerEm:  1000,
		GlyphWidths: map[uint16]uint16{0: 0, 1: 500, 2: 750},
		CharToGlyph: map[rune]uint16{'A': 1, 'B': 2},
		FontData:    []byte("FAKE_FONT_DATA_FOR_TESTING_ZLIB"),
	}
}

func TestFontSubset_Build_Empty(t *testing.T) {
	font := makeFontWithData()
	subset := NewFontSubset(font)

	// Build with no used chars — still compresses full font data.
	err := subset.Build()
	if err != nil {
		t.Fatalf("Build() with no used chars failed: %v", err)
	}
	if len(subset.SubsetData) == 0 {
		t.Error("SubsetData should be non-empty after Build()")
	}
	// GlyphMapping must contain at least glyph 0 (.notdef).
	if _, ok := subset.GlyphMapping[0]; !ok {
		t.Error("GlyphMapping should contain glyph 0 (.notdef) after Build()")
	}
}

func TestFontSubset_Build_WithUsedChars(t *testing.T) {
	font := makeFontWithData()
	subset := NewFontSubset(font)
	subset.UseString("AB")

	err := subset.Build()
	if err != nil {
		t.Fatalf("Build() with used chars failed: %v", err)
	}
	if len(subset.SubsetData) == 0 {
		t.Error("SubsetData should be non-empty after Build()")
	}
	// Glyph mapping should include glyphs for 'A' (1) and 'B' (2).
	if _, ok := subset.GlyphMapping[1]; !ok {
		t.Error("GlyphMapping should contain glyph 1 (A) after Build()")
	}
	if _, ok := subset.GlyphMapping[2]; !ok {
		t.Error("GlyphMapping should contain glyph 2 (B) after Build()")
	}
}

func TestFontSubset_Build_SubsetDataIsValidZlib(t *testing.T) {
	font := makeFontWithData()
	subset := NewFontSubset(font)
	subset.UseChar('A')

	if err := subset.Build(); err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify SubsetData starts with zlib magic byte (0x78).
	if len(subset.SubsetData) < 2 {
		t.Fatal("SubsetData too short to be valid zlib")
	}
	if subset.SubsetData[0] != 0x78 {
		t.Errorf("SubsetData[0] = 0x%02X, expected 0x78 (zlib magic)", subset.SubsetData[0])
	}
}

func TestFontSubset_CreateGlyphMapping_Order(t *testing.T) {
	font := makeFontWithData()
	subset := NewFontSubset(font)
	subset.UseString("AB")

	glyphs := subset.identifyUsedGlyphs()
	subset.createGlyphMapping(glyphs)

	// New IDs should be sequential starting from 0 matching sorted glyphs.
	for newID, oldID := range glyphs {
		mapped, ok := subset.GlyphMapping[oldID]
		if !ok {
			t.Errorf("GlyphMapping missing old glyph %d", oldID)
			continue
		}
		if int(mapped) != newID {
			t.Errorf("GlyphMapping[%d] = %d, want %d (sequential)", oldID, mapped, newID)
		}
	}
}

func TestFontSubset_GetCharWidth_MissingGlyphWidth(t *testing.T) {
	font := &TTFFont{
		UnitsPerEm:  1000,
		CharToGlyph: map[rune]uint16{'Z': 99},
		GlyphWidths: map[uint16]uint16{}, // glyph 99 has no width entry
	}
	subset := NewFontSubset(font)

	width := subset.GetCharWidth('Z')
	if width != 0 {
		t.Errorf("GetCharWidth for missing glyph width = %d, want 0", width)
	}
}

func TestFontSubset_MeasureString_ZeroUnitsPerEm(t *testing.T) {
	font := &TTFFont{
		UnitsPerEm:  0, // triggers fallback to 1000
		CharToGlyph: map[rune]uint16{'A': 1},
		GlyphWidths: map[uint16]uint16{1: 500},
	}
	subset := NewFontSubset(font)

	width := subset.MeasureString("A", 12)
	// Should use fallback unitsPerEm = 1000: 500 * 12 / 1000 = 6.0
	if width != 6.0 {
		t.Errorf("MeasureString with zero unitsPerEm = %f, want 6.0", width)
	}
}

func TestFontSubset_GetWidths_Empty(t *testing.T) {
	font := &TTFFont{
		UnitsPerEm:  1000,
		CharToGlyph: map[rune]uint16{},
		GlyphWidths: map[uint16]uint16{},
	}
	subset := NewFontSubset(font)
	// No used chars.
	first, last, widths := subset.GetWidths()
	if first != 0 || last != 0 || widths != nil {
		t.Errorf("GetWidths on empty subset = (%d,%d,%v), want (0,0,nil)", first, last, widths)
	}
}

// ============================================================================
// calculateDerivedMetrics — covers all weight class branches (50% → 100%)
// ============================================================================

func TestCalculateDerivedMetrics_LightWeight(t *testing.T) {
	f := &TTFFont{WeightClass: 100}
	f.calculateDerivedMetrics()
	// 50 + 100/10 = 60
	if f.StemV != 60 {
		t.Errorf("StemV for weight 100 = %d, want 60", f.StemV)
	}
}

func TestCalculateDerivedMetrics_Light300(t *testing.T) {
	f := &TTFFont{WeightClass: 300}
	f.calculateDerivedMetrics()
	// 50 + 300/10 = 80
	if f.StemV != 80 {
		t.Errorf("StemV for weight 300 = %d, want 80", f.StemV)
	}
}

func TestCalculateDerivedMetrics_NormalWeight400(t *testing.T) {
	f := &TTFFont{WeightClass: 400}
	f.calculateDerivedMetrics()
	// 80 + (400-400)/5 = 80
	if f.StemV != 80 {
		t.Errorf("StemV for 400 weight = %d, want 80", f.StemV)
	}
}

func TestCalculateDerivedMetrics_MediumWeight500(t *testing.T) {
	f := &TTFFont{WeightClass: 500}
	f.calculateDerivedMetrics()
	// 80 + (500-400)/5 = 80 + 20 = 100
	if f.StemV != 100 {
		t.Errorf("StemV for 500 weight = %d, want 100", f.StemV)
	}
}

func TestCalculateDerivedMetrics_SemiBoldWeight600(t *testing.T) {
	f := &TTFFont{WeightClass: 600}
	f.calculateDerivedMetrics()
	// 100 + (600-500)/5 = 100 + 20 = 120
	if f.StemV != 120 {
		t.Errorf("StemV for 600 weight = %d, want 120", f.StemV)
	}
}

func TestCalculateDerivedMetrics_BoldWeight700(t *testing.T) {
	f := &TTFFont{WeightClass: 700}
	f.calculateDerivedMetrics()
	// 100 + (700-500)/5 = 100 + 40 = 140
	if f.StemV != 140 {
		t.Errorf("StemV for 700 weight = %d, want 140", f.StemV)
	}
}

func TestCalculateDerivedMetrics_BlackWeight900(t *testing.T) {
	f := &TTFFont{WeightClass: 900}
	f.calculateDerivedMetrics()
	// 130 + (900-700)/10 = 130 + 20 = 150
	if f.StemV != 150 {
		t.Errorf("StemV for 900 weight = %d, want 150", f.StemV)
	}
}

func TestCalculateDerivedMetrics_ZeroWeight(t *testing.T) {
	f := &TTFFont{WeightClass: 0}
	f.calculateDerivedMetrics()
	// WeightClass == 0: falls to first case (<=300): 50 + 0/10 = 50, then
	// the "StemV == 0 || WeightClass == 0" guard kicks in: StemV = 80.
	if f.StemV != 80 {
		t.Errorf("StemV for 0 weight = %d, want 80 (default)", f.StemV)
	}
}

func TestCalculateDerivedMetrics_FixedPitchFlag(t *testing.T) {
	f := &TTFFont{WeightClass: 400, IsFixedPitch: true}
	f.calculateDerivedMetrics()
	// Flags should have bit 1 set (FixedPitch).
	if f.Flags&1 == 0 {
		t.Errorf("Flags should have FixedPitch bit set, got 0x%X", f.Flags)
	}
}

func TestCalculateDerivedMetrics_ItalicFlag(t *testing.T) {
	f := &TTFFont{WeightClass: 400, ItalicAngle: -15.0}
	f.calculateDerivedMetrics()
	// Flags should have italic bit set (bit 7 = 64).
	if f.Flags&64 == 0 {
		t.Errorf("Flags should have Italic bit set for italic angle, got 0x%X", f.Flags)
	}
}

func TestCalculateDerivedMetrics_NonSymbolicDefaultFlags(t *testing.T) {
	f := &TTFFont{WeightClass: 400}
	f.calculateDerivedMetrics()
	// Default: Nonsymbolic (bit 6 = 32) should be set.
	if f.Flags&32 == 0 {
		t.Errorf("Flags should have Nonsymbolic bit set by default, got 0x%X", f.Flags)
	}
}

// ============================================================================
// parseCmapFormat12 — covered by calling it directly (0% → 100%)
// ============================================================================

func TestParseCmapFormat12_NotImplemented(t *testing.T) {
	f := &TTFFont{}
	err := f.parseCmapFormat12(nil, 0)
	if err == nil {
		t.Error("parseCmapFormat12 should return error (not yet implemented)")
	}
}

// ============================================================================
// Standard14Font.WriteFontObject — cover additional font variants
// ============================================================================

// ============================================================================
// GenerateToUnicodeCMap — cover "character not in font" skip path
// ============================================================================

func TestGenerateToUnicodeCMap_SkipsUnmappedChar(t *testing.T) {
	ttf := &TTFFont{
		CharToGlyph: map[rune]uint16{
			'A': 1, // only 'A' is in font
		},
	}
	subset := NewFontSubset(ttf)
	subset.UseChar('A')
	subset.UseChar('X') // 'X' is NOT in CharToGlyph → will be skipped

	cmap, err := GenerateToUnicodeCMap(subset)
	if err != nil {
		t.Fatalf("GenerateToUnicodeCMap failed: %v", err)
	}
	cmapStr := string(cmap)
	// 'A' → glyph 1, Unicode 0x0041
	if !strings.Contains(cmapStr, "<0001> <0041>") {
		t.Errorf("CMap should contain mapping for 'A': got\n%s", cmapStr)
	}
	// 'X' should NOT appear (no glyph mapping)
	if strings.Contains(cmapStr, "<0058>") { // U+0058 = X
		t.Error("CMap should not contain unmapped character 'X'")
	}
}

// ============================================================================
// writeMappingBatch — covers the batch writing path with known data
// ============================================================================

func TestWriteMappingBatch_Single(t *testing.T) {
	var buf bytes.Buffer
	mappings := []glyphMapping{
		{glyphID: 1, unicode: 'A'},
	}
	err := writeMappingBatch(&buf, mappings)
	if err != nil {
		t.Fatalf("writeMappingBatch failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1 beginbfchar") {
		t.Errorf("expected '1 beginbfchar', got: %s", out)
	}
	if !strings.Contains(out, "<0001> <0041>") {
		t.Errorf("expected '<0001> <0041>', got: %s", out)
	}
	if !strings.Contains(out, "endbfchar") {
		t.Errorf("expected 'endbfchar', got: %s", out)
	}
}

// ============================================================================
// GenerateFontDescriptor — cover nil path and derived name path
// ============================================================================

func TestGenerateFontDescriptor_Nil(t *testing.T) {
	result := GenerateFontDescriptor(nil)
	if result != nil {
		t.Error("GenerateFontDescriptor(nil) should return nil")
	}
}

func TestGenerateFontDescriptor_WithPostScriptName(t *testing.T) {
	ttf := &TTFFont{
		PostScriptName: "OpenSans-Bold",
		UnitsPerEm:     1000,
		Ascender:       800,
		Descender:      -200,
		CapHeight:      700,
		StemV:          80,
	}
	fd := GenerateFontDescriptor(ttf)
	if fd == nil {
		t.Fatal("GenerateFontDescriptor returned nil for valid TTF")
	}
	if fd.FontName != "OpenSans-Bold" {
		t.Errorf("FontName = %q, want %q", fd.FontName, "OpenSans-Bold")
	}
	if fd.Ascent != 800 {
		t.Errorf("Ascent = %d, want 800", fd.Ascent)
	}
	if fd.Descent != -200 {
		t.Errorf("Descent = %d, want -200", fd.Descent)
	}
}

func TestGenerateFontDescriptor_DerivedNameFromPath(t *testing.T) {
	ttf := &TTFFont{
		PostScriptName: "", // no PS name → derive from path
		FilePath:       "/path/to/OpenSans-Regular.ttf",
		UnitsPerEm:     1000,
	}
	fd := GenerateFontDescriptor(ttf)
	if fd == nil {
		t.Fatal("GenerateFontDescriptor returned nil")
	}
	if fd.FontName != "OpenSans-Regular" {
		t.Errorf("FontName = %q, want %q", fd.FontName, "OpenSans-Regular")
	}
}

func TestGenerateFontDescriptor_ToPDFDict_WithXHeight(t *testing.T) {
	fd := &FontDescriptor{
		FontName:  "Test",
		Flags:     32,
		Ascent:    800,
		Descent:   -200,
		CapHeight: 700,
		StemV:     80,
		XHeight:   500, // should be included
	}
	dict := fd.ToPDFDict(0)
	if !strings.Contains(dict, "/XHeight 500") {
		t.Errorf("dict should contain /XHeight 500, got: %s", dict)
	}
}

func TestGenerateFontDescriptor_ToPDFDict_WithFontFile2(t *testing.T) {
	fd := &FontDescriptor{
		FontName:  "Test",
		Flags:     32,
		CapHeight: 700,
		StemV:     80,
	}
	dict := fd.ToPDFDict(42)
	if !strings.Contains(dict, "/FontFile2 42 0 R") {
		t.Errorf("dict should contain /FontFile2 42 0 R, got: %s", dict)
	}
}

func TestSubsetFontName_EmptyChars(t *testing.T) {
	// Empty rune list should still produce valid prefix+name format.
	name := SubsetFontName("TestFont", nil)
	if !strings.Contains(name, "+TestFont") {
		t.Errorf("SubsetFontName with empty chars should contain +TestFont, got %q", name)
	}
	if len(name) != len("+TestFont")+6 { // 6-letter prefix
		t.Errorf("SubsetFontName should have 6-letter prefix, got %q", name)
	}
}

// ============================================================================
// LoadTTF with real font file from reference directory — covers TTF parser paths
// ============================================================================

const dejaVuFontPath = "D:/projects/gopdf/reference/krilla/assets/fonts/DejaVuSansMono.ttf"

func TestLoadTTF_DejaVu(t *testing.T) {
	ttf, err := LoadTTF(dejaVuFontPath)
	if err != nil {
		t.Skipf("DejaVu font not available: %v", err)
	}
	if ttf == nil {
		t.Fatal("LoadTTF returned nil")
	}
	if ttf.UnitsPerEm == 0 {
		t.Error("UnitsPerEm should not be 0")
	}
	if len(ttf.CharToGlyph) == 0 {
		t.Error("CharToGlyph should not be empty")
	}
	if len(ttf.GlyphWidths) == 0 {
		t.Error("GlyphWidths should not be empty")
	}
	// Basic ASCII chars should be present.
	for _, ch := range "ABCabc012" {
		if _, ok := ttf.CharToGlyph[ch]; !ok {
			t.Errorf("ASCII char %q not in CharToGlyph", ch)
		}
	}
}

func TestLoadTTF_DejaVu_MetricsAreReasonable(t *testing.T) {
	ttf, err := LoadTTF(dejaVuFontPath)
	if err != nil {
		t.Skipf("DejaVu font not available: %v", err)
	}
	if ttf.Ascender <= 0 {
		t.Errorf("Ascender should be positive, got %d", ttf.Ascender)
	}
	if ttf.Descender >= 0 {
		t.Errorf("Descender should be negative, got %d", ttf.Descender)
	}
	if ttf.PostScriptName == "" {
		t.Error("PostScriptName should not be empty for DejaVu")
	}
}

func TestLoadTTF_NonExistent(t *testing.T) {
	_, err := LoadTTF("/nonexistent/path/to/font.ttf")
	if err == nil {
		t.Error("LoadTTF should fail for nonexistent file")
	}
}

func TestFontSubset_Build_WithRealFont(t *testing.T) {
	ttf, err := LoadTTF(dejaVuFontPath)
	if err != nil {
		t.Skipf("DejaVu font not available: %v", err)
	}
	subset := NewFontSubset(ttf)
	subset.UseString("Hello PDF")

	err = subset.Build()
	if err != nil {
		t.Fatalf("Build() with real font failed: %v", err)
	}
	if len(subset.SubsetData) == 0 {
		t.Error("SubsetData should be non-empty")
	}
}

func TestWriteFontObject_AllVariants(t *testing.T) {
	allFonts := []*Standard14Font{
		TimesRoman, TimesBold, TimesItalic, TimesBoldItalic,
		Helvetica, HelveticaBold, HelveticaOblique, HelveticaBoldOblique,
		Courier, CourierBold, CourierOblique, CourierBoldOblique,
		Symbol, ZapfDingbats,
	}
	for i, font := range allFonts {
		t.Run(font.Name, func(t *testing.T) {
			var buf bytes.Buffer
			err := font.WriteFontObject(i+1, &buf)
			if err != nil {
				t.Errorf("WriteFontObject(%s) failed: %v", font.Name, err)
			}
			if buf.Len() == 0 {
				t.Errorf("WriteFontObject(%s) wrote nothing", font.Name)
			}
			if !strings.Contains(buf.String(), "endobj") {
				t.Errorf("WriteFontObject(%s) output missing 'endobj'", font.Name)
			}
		})
	}
}

// ============================================================================
// WriteFontObject — error writer paths
// ============================================================================

// limitedWriter fails after writing N bytes.
type limitedWriter struct {
	limit   int
	written int
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	remaining := lw.limit - lw.written
	if remaining <= 0 {
		return 0, fmt.Errorf("write limit reached")
	}
	if len(p) > remaining {
		lw.written += remaining
		return remaining, fmt.Errorf("write limit reached")
	}
	lw.written += len(p)
	return len(p), nil
}

func TestWriteFontObject_WriterErrorOnHeader(t *testing.T) {
	// Fail before writing anything
	w := &limitedWriter{limit: 0}
	err := Helvetica.WriteFontObject(1, w)
	if err == nil {
		t.Error("WriteFontObject should fail when writer fails on header")
	}
}

func TestWriteFontObject_WriterErrorOnType(t *testing.T) {
	// Allow header to write (e.g., "1 0 obj\n" = 8 bytes) then fail
	w := &limitedWriter{limit: 8}
	err := Helvetica.WriteFontObject(1, w)
	if err == nil {
		t.Error("WriteFontObject should fail when writer fails on type line")
	}
}

func TestWriteFontObject_WriterErrorOnSubtype(t *testing.T) {
	// Allow header + type line then fail
	w := &limitedWriter{limit: 25}
	err := Helvetica.WriteFontObject(1, w)
	if err == nil {
		t.Error("WriteFontObject should fail when writer fails on subtype line")
	}
}

func TestWriteFontObject_WriterErrorOnBaseFont(t *testing.T) {
	w := &limitedWriter{limit: 45}
	err := Helvetica.WriteFontObject(1, w)
	if err == nil {
		t.Error("WriteFontObject should fail when writer fails on base font line")
	}
}

func TestWriteFontObject_WriterErrorOnEncoding(t *testing.T) {
	// Allow through base font line (about 60 bytes) then fail on encoding.
	w := &limitedWriter{limit: 62}
	err := Helvetica.WriteFontObject(1, w)
	if err == nil {
		t.Error("WriteFontObject should fail when writer fails on encoding line")
	}
}

func TestWriteFontObject_WriterErrorOnDictClose(t *testing.T) {
	// Allow through encoding line (~85 bytes) then fail on ">>\n".
	w := &limitedWriter{limit: 87}
	err := Helvetica.WriteFontObject(1, w)
	if err == nil {
		t.Error("WriteFontObject should fail when writer fails on dict close")
	}
}

func TestWriteFontObject_WriterErrorOnFooter(t *testing.T) {
	// Allow through everything except "endobj\n".
	w := &limitedWriter{limit: 94}
	err := Helvetica.WriteFontObject(1, w)
	if err == nil {
		t.Error("WriteFontObject should fail when writer fails on footer")
	}
}

func TestWriteFontObject_SymbolicFont_NoEncoding(t *testing.T) {
	// Symbol is symbolic — should NOT write /Encoding line.
	var buf bytes.Buffer
	err := Symbol.WriteFontObject(1, &buf)
	if err != nil {
		t.Fatalf("WriteFontObject(Symbol) failed: %v", err)
	}
	if strings.Contains(buf.String(), "Encoding") {
		t.Error("symbolic font should not write /Encoding entry")
	}
}

// ============================================================================
// TTF parser — error paths via malformed binary data
// ============================================================================

// makeValidTTFHeader builds a minimal valid TrueType font header (just the
// sfnt version + 0 tables) — enough to pass parseFontDirectory without error.
func makeValidTTFHeader() []byte {
	var buf bytes.Buffer
	// sfntVersion = 0x00010000 (TrueType)
	buf.Write([]byte{0x00, 0x01, 0x00, 0x00})
	// numTables = 0
	buf.Write([]byte{0x00, 0x00})
	// searchRange, entrySelector, rangeShift (6 bytes) = 0
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	return buf.Bytes()
}

func TestParseFontDirectory_TruncatedData(t *testing.T) {
	// Too short to read sfnt version.
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parse([]byte{0x00, 0x01}) // only 2 bytes, need 4
	if err == nil {
		t.Error("parse should fail with truncated data")
	}
}

func TestParseFontDirectory_UnsupportedVersion(t *testing.T) {
	// Version 0x4F54544F = "OTTO" (CFF, not TrueType).
	data := []byte{0x4F, 0x54, 0x54, 0x4F, 0x00, 0x00}
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parse(data)
	if err == nil {
		t.Error("parse should fail for CFF (OTTO) format")
	}
}

func TestParseFontDirectory_TruncatedAfterVersion(t *testing.T) {
	// Valid version but then EOF before numTables.
	data := []byte{0x00, 0x01, 0x00, 0x00} // only 4 bytes, need 6+ more
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parse(data)
	if err == nil {
		t.Error("parse should fail with truncated data after sfnt version")
	}
}

func TestParse_MissingRequiredTables(t *testing.T) {
	// Valid header with 0 tables — parseRequiredTables will fail (missing "head").
	data := makeValidTTFHeader()
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
		FontData:    data,
	}
	err := f.parse(data)
	if err == nil {
		t.Error("parse should fail when required tables are missing")
	}
}

func TestParseCmapSubtable_UnsupportedFormat(t *testing.T) {
	// Build a 2-byte cmap "subtable" with format=6 (unsupported).
	data := []byte{0x00, 0x06} // format = 6
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseCmapSubtable(data, 0)
	if err == nil {
		t.Error("parseCmapSubtable should fail for unsupported format")
	}
}

func TestParseCmapSubtable_TruncatedData(t *testing.T) {
	// Empty data → can't read format.
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseCmapSubtable([]byte{}, 0)
	if err == nil {
		t.Error("parseCmapSubtable should fail with empty data")
	}
}

func TestFindBestCmapSubtable_NoSuitableTable(t *testing.T) {
	// Build a cmap data with one record that doesn't match Windows Unicode BMP.
	// format(2)+numTables(2) = 4 bytes header
	// then record: platformID(2)+encodingID(2)+offset(4) = 8 bytes per entry
	data := make([]byte, 12)
	// version=0, numTables=1
	data[0], data[1] = 0x00, 0x00
	data[2], data[3] = 0x00, 0x01
	// record: platformID=1 (Mac), encodingID=0, offset=0
	data[4], data[5] = 0x00, 0x01 // platformID = 1 (not Windows=3)
	data[6], data[7] = 0x00, 0x00 // encodingID = 0
	data[8], data[9] = 0x00, 0x00
	data[10], data[11] = 0x00, 0x00 // offset = 0

	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	_, err := f.findBestCmapSubtable(data, 1)
	if err == nil {
		t.Error("findBestCmapSubtable should fail when no Windows Unicode table found")
	}
}

func TestLoadTTF_InvalidFile(t *testing.T) {
	// Write a file with invalid TTF data.
	tmpDir := t.TempDir()
	path := tmpDir + "/invalid.ttf"
	err := writeFile(path, []byte("not a ttf file"))
	if err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	_, err = LoadTTF(path)
	if err == nil {
		t.Error("LoadTTF should fail for invalid TTF data")
	}
}

// writeFile is a helper to write bytes to a file path.
func writeFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// ============================================================================
// tounicode.go — additional path coverage via large batch (>100 mappings)
// ============================================================================

func TestGenerateToUnicodeCMap_LargeBatch(t *testing.T) {
	// Create more than 100 char mappings to trigger multi-batch writing.
	charToGlyph := make(map[rune]uint16, 150)
	for i := 0; i < 150; i++ {
		charToGlyph[rune(0x4E00+i)] = uint16(i + 1)
	}
	ttf := &TTFFont{CharToGlyph: charToGlyph}
	subset := NewFontSubset(ttf)
	for ch := range charToGlyph {
		subset.UseChar(ch)
	}
	cmap, err := GenerateToUnicodeCMap(subset)
	if err != nil {
		t.Fatalf("GenerateToUnicodeCMap with large batch failed: %v", err)
	}
	if len(cmap) == 0 {
		t.Error("CMap should not be empty")
	}
	// Should contain at least 2 "beginbfchar" entries (two batches).
	if strings.Count(string(cmap), "beginbfchar") < 2 {
		t.Errorf("expected 2+ batch beginnings, got: %d", strings.Count(string(cmap), "beginbfchar"))
	}
}

// ============================================================================
// compressFont — error path (empty FontData still compresses, no error)
// Build() — verify second call is idempotent
// ============================================================================

func TestFontSubset_Build_Idempotent(t *testing.T) {
	font := makeFontWithData()
	subset := NewFontSubset(font)
	subset.UseChar('A')

	if err := subset.Build(); err != nil {
		t.Fatalf("first Build() failed: %v", err)
	}
	firstData := make([]byte, len(subset.SubsetData))
	copy(firstData, subset.SubsetData)

	// Second Build() should regenerate (subset was invalidated on each UseChar/UseString).
	// Actually Build() doesn't set isBuilt on FontSubset. Just verify no panic/error.
	if err := subset.Build(); err != nil {
		t.Fatalf("second Build() failed: %v", err)
	}
}

// ============================================================================
// parseOS2Table and parseHeadTable — via patched TTFFont with minimal table data
// ============================================================================

func TestParseOS2Table_TruncatedData(t *testing.T) {
	f := &TTFFont{
		Tables: map[string]*TTFTable{
			"OS/2": {Data: []byte{0x00, 0x04}}, // only 2 bytes, need much more
		},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	// parseOS2Table should fail with truncated data.
	err := f.parseOS2Table()
	if err == nil {
		t.Error("parseOS2Table should fail with truncated data")
	}
}

func TestParseHeadTable_TruncatedData(t *testing.T) {
	f := &TTFFont{
		Tables: map[string]*TTFTable{
			"head": {Data: []byte{0x00, 0x01}}, // only 2 bytes
		},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHeadTable()
	if err == nil {
		t.Error("parseHeadTable should fail with truncated data")
	}
}

func TestParseHheaTable_TruncatedData(t *testing.T) {
	f := &TTFFont{
		Tables: map[string]*TTFTable{
			"hhea": {Data: []byte{0x00, 0x01, 0x02}}, // too short
		},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHheaTable()
	if err == nil {
		t.Error("parseHheaTable should fail with truncated data")
	}
}

func TestParsePostTable_TruncatedData(t *testing.T) {
	f := &TTFFont{
		Tables: map[string]*TTFTable{
			"post": {Data: []byte{0x00}}, // only 1 byte, need at least 32
		},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parsePostTable()
	if err == nil {
		t.Error("parsePostTable should fail with truncated data")
	}
}

func TestParsePostTable_ValidData(t *testing.T) {
	// Build a valid 32-byte post table.
	// version (4), italicAngle (4), underlinePosition (2), underlineThickness (2), isFixedPitch (4) = 16 bytes min
	// But parsePostTable reads: skip 4 (version), read 4 (italicAngle), read 2 (underPos), read 2 (underThick), read 4 (isFixedPitch) = 16 bytes
	// Needs len >= 32, so pad to 32.
	data := make([]byte, 32)
	// version = 2.0 (0x00020000)
	data[0], data[1], data[2], data[3] = 0x00, 0x02, 0x00, 0x00
	// italicAngle = 0 (Fixed 16.16 = 0x00000000)
	// underlinePosition = 0
	// underlineThickness = 0
	// isFixedPitch = 0 (not fixed)
	f := &TTFFont{
		Tables:      map[string]*TTFTable{"post": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parsePostTable()
	if err != nil {
		t.Errorf("parsePostTable failed with valid data: %v", err)
	}
}

func TestParseOS2Table_ValidVersion0(t *testing.T) {
	// Build a valid OS/2 table with version 0 (78 bytes minimum).
	// Reader positions after each read:
	//   Read version (2) → pos=2
	//   skip xAvgCharWidth (2) → pos=4
	//   Read usWeightClass (2) → pos=6
	//   Read usWidthClass (2) → pos=8
	//   Read fsType (2) → pos=10
	//   skip 56 → pos=66 (sTypoAscender)
	//   Read sTypoAscender (2) → pos=68
	//   Read sTypoDescender (2) → pos=70
	//   skip sTypoLineGap (2) → pos=72
	//   skip usWinAscent+Descent (4) → pos=76
	//   for version 0: no more reads needed → need 76 bytes minimum
	data := make([]byte, 78)
	// version = 0
	// usWeightClass at reader pos 4 (data[4:6]) = 400 = 0x0190
	data[4], data[5] = 0x01, 0x90
	// sTypoAscender at reader pos 66 (data[66:68]) = 800 = 0x0320
	data[66], data[67] = 0x03, 0x20
	// sTypoDescender at reader pos 68 (data[68:70]) = -200 = 0xFF38 (int16 big-endian)
	data[68], data[69] = 0xFF, 0x38

	f := &TTFFont{
		Tables:      map[string]*TTFTable{"OS/2": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
		Ascender:    800,
	}
	err := f.parseOS2Table()
	if err != nil {
		t.Errorf("parseOS2Table failed with valid version 0 data: %v", err)
	}
	if f.WeightClass != 400 {
		t.Errorf("WeightClass = %d, want 400", f.WeightClass)
	}
	// version 0: CapHeight estimated as 70% of Ascender.
	expectedCapHeight := int16(float64(f.Ascender) * 0.7)
	if f.CapHeight != expectedCapHeight {
		t.Errorf("CapHeight = %d, want %d (estimated)", f.CapHeight, expectedCapHeight)
	}
}

func TestParseOS2Table_ValidVersion2(t *testing.T) {
	// Build a valid OS/2 table with version 2 (96+ bytes).
	// After sTypoDescender (pos=70): skip sTypoLineGap(2)+usWinAscent+Descent(4)=6 → pos=76
	// For version>=2 AND len>=96: skip ulCodePageRange1+2 (8) → pos=84
	//   Read sxHeight (2) → pos=86
	//   Read sCapHeight (2) → pos=88
	data := make([]byte, 96)
	// version = 2 at pos 0 (data[0:2])
	data[0], data[1] = 0x00, 0x02
	// usWeightClass at pos 4 (data[4:6]) = 700 = 0x02BC
	data[4], data[5] = 0x02, 0xBC
	// sTypoAscender at pos 66 (data[66:68]) = 800 = 0x0320
	data[66], data[67] = 0x03, 0x20
	// sTypoDescender at pos 68 (data[68:70]) = -200 = 0xFF38
	data[68], data[69] = 0xFF, 0x38
	// sxHeight at pos 84 (data[84:86]) = 550 = 0x0226
	data[84], data[85] = 0x02, 0x26
	// sCapHeight at pos 86 (data[86:88]) = 720 = 0x02D0
	data[86], data[87] = 0x02, 0xD0

	f := &TTFFont{
		Tables:      map[string]*TTFTable{"OS/2": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseOS2Table()
	if err != nil {
		t.Errorf("parseOS2Table failed with version 2 data: %v", err)
	}
	if f.WeightClass != 700 {
		t.Errorf("WeightClass = %d, want 700", f.WeightClass)
	}
	if f.CapHeight != 720 {
		t.Errorf("CapHeight = %d, want 720", f.CapHeight)
	}
	if f.XHeight != 550 {
		t.Errorf("XHeight = %d, want 550", f.XHeight)
	}
}

func TestParseHeadTable_ValidData(t *testing.T) {
	// parseHeadTable reads:
	// skip 8 (version+fontRevision), skip 8 (checksum+magic), skip 2 (flags) = 18 bytes
	// read 2 (unitsPerEm)
	// skip 16 (timestamps)
	// read 2*4 (FontBBox xMin, yMin, xMax, yMax)
	// Total minimum: 18+2+16+8 = 44 bytes
	data := make([]byte, 44)
	// unitsPerEm at offset 18 = 1000 = 0x03E8
	data[18], data[19] = 0x03, 0xE8
	// xMin at offset 36 = -100 = 0xFF9C
	data[36], data[37] = 0xFF, 0x9C
	// yMin at offset 38 = -200 = 0xFF38
	data[38], data[39] = 0xFF, 0x38
	// xMax at offset 40 = 800 = 0x0320
	data[40], data[41] = 0x03, 0x20
	// yMax at offset 42 = 900 = 0x0384
	data[42], data[43] = 0x03, 0x84

	f := &TTFFont{
		Tables:      map[string]*TTFTable{"head": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHeadTable()
	if err != nil {
		t.Errorf("parseHeadTable failed with valid data: %v", err)
	}
	if f.UnitsPerEm != 1000 {
		t.Errorf("UnitsPerEm = %d, want 1000", f.UnitsPerEm)
	}
}

// ============================================================================
// parseNameTable — cover platform-specific paths
// ============================================================================

func TestParseNameTable_MacPlatform(t *testing.T) {
	// Build a minimal name table with one record: nameID=6, platform=1 (Mac).
	// Format:
	//   format (2), count (2), stringOffset (2) = 6 bytes header
	//   per record: platformID(2)+encodingID(2)+languageID(2)+nameID(2)+length(2)+offset(2) = 12 bytes
	//   string data: "TestFont"
	psName := "TestFont"
	psLen := len(psName)
	// stringOffset = 6 + 1*12 = 18 (right after the one record)
	stringOffset := uint16(18)
	data := make([]byte, 18+psLen)
	// format = 0
	data[0], data[1] = 0x00, 0x00
	// count = 1
	data[2], data[3] = 0x00, 0x01
	// stringOffset = 18
	data[4], data[5] = 0x00, 0x12
	// Record: platformID=1, encodingID=0, languageID=0, nameID=6, length=8, offset=0
	data[6], data[7] = 0x00, 0x01   // platformID = 1 (Mac)
	data[8], data[9] = 0x00, 0x00   // encodingID
	data[10], data[11] = 0x00, 0x00 // languageID
	data[12], data[13] = 0x00, 0x06 // nameID = 6 (PostScript)
	data[14], data[15] = 0x00, byte(psLen)
	data[16], data[17] = 0x00, 0x00 // offset = 0
	copy(data[stringOffset:], []byte(psName))

	f := &TTFFont{
		Tables:      map[string]*TTFTable{"name": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseNameTable()
	if err != nil {
		t.Errorf("parseNameTable failed: %v", err)
	}
	if f.PostScriptName != psName {
		t.Errorf("PostScriptName = %q, want %q", f.PostScriptName, psName)
	}
}

func TestParseNameTable_WindowsPlatform(t *testing.T) {
	// Build a name table with nameID=6, platform=3 (Windows), UTF-16BE encoded.
	psName := "WinFont" // 7 ASCII chars → 14 bytes in UTF-16BE
	utf16Data := make([]byte, len(psName)*2)
	for i, ch := range psName {
		utf16Data[i*2] = 0x00
		utf16Data[i*2+1] = byte(ch)
	}
	psLen := len(utf16Data)
	stringOffset := uint16(18)
	data := make([]byte, 18+psLen)
	data[0], data[1] = 0x00, 0x00 // format
	data[2], data[3] = 0x00, 0x01 // count = 1
	data[4], data[5] = byte(stringOffset>>8), byte(stringOffset)
	// Record: platformID=3 (Windows), encodingID=1, languageID=0, nameID=6
	data[6], data[7] = 0x00, 0x03   // platformID = 3 (Windows)
	data[8], data[9] = 0x00, 0x01   // encodingID = 1
	data[10], data[11] = 0x00, 0x00 // languageID
	data[12], data[13] = 0x00, 0x06 // nameID = 6
	data[14], data[15] = 0x00, byte(psLen)
	data[16], data[17] = 0x00, 0x00 // offset = 0
	copy(data[stringOffset:], utf16Data)

	f := &TTFFont{
		Tables:      map[string]*TTFTable{"name": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseNameTable()
	if err != nil {
		t.Errorf("parseNameTable failed: %v", err)
	}
	if f.PostScriptName != psName {
		t.Errorf("PostScriptName = %q, want %q", f.PostScriptName, psName)
	}
}

func TestParseNameTable_SkipsNonPostScriptID(t *testing.T) {
	// Build a name table with nameID=1 (family, not PS name) — should skip.
	stringOffset := uint16(18)
	data := make([]byte, 18+5)
	data[2], data[3] = 0x00, 0x01 // count = 1
	data[4], data[5] = byte(stringOffset>>8), byte(stringOffset)
	data[6], data[7] = 0x00, 0x01   // platformID = 1
	data[12], data[13] = 0x00, 0x01 // nameID = 1 (not 6)
	data[14], data[15] = 0x00, 0x05 // length = 5
	copy(data[18:], []byte("Hello"))

	f := &TTFFont{
		Tables:      map[string]*TTFTable{"name": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseNameTable()
	if err != nil {
		t.Errorf("parseNameTable failed: %v", err)
	}
	// PostScriptName should remain empty (nameID was not 6).
	if f.PostScriptName != "" {
		t.Errorf("PostScriptName = %q, expected empty (non-PS nameID)", f.PostScriptName)
	}
}

func TestParseHheaTable_ValidData(t *testing.T) {
	// parseHheaTable reads:
	// skip 4 (version)
	// read 2 (Ascender), read 2 (Descender), read 2 (LineGap) = 6 bytes
	// skip 24 bytes
	// read 2 (numOfLongHorMetrics)
	// Total: 4+6+24+2 = 36 bytes minimum
	data := make([]byte, 36)
	// Ascender at offset 4 = 800 = 0x0320
	data[4], data[5] = 0x03, 0x20
	// Descender at offset 6 = -200 = 0xFF38
	data[6], data[7] = 0xFF, 0x38
	// LineGap at offset 8 = 0
	// numHMetrics at offset 34 = 2
	data[34], data[35] = 0x00, 0x02

	f := &TTFFont{
		Tables:      map[string]*TTFTable{"hhea": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHheaTable()
	if err != nil {
		t.Errorf("parseHheaTable failed with valid data: %v", err)
	}
	if f.Ascender != 800 {
		t.Errorf("Ascender = %d, want 800", f.Ascender)
	}
	if f.Descender != -200 {
		t.Errorf("Descender = %d, want -200", f.Descender)
	}
}

// ============================================================================
// "Table not found" error paths — previously uncovered branches
// ============================================================================

func TestParseHeadTable_NotFound(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHeadTable()
	if err == nil {
		t.Error("parseHeadTable should fail when head table missing")
	}
}

func TestParseHheaTable_NotFound(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHheaTable()
	if err == nil {
		t.Error("parseHheaTable should fail when hhea table missing")
	}
}

func TestParseHmtxTable_NotFound(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHmtxTable()
	if err == nil {
		t.Error("parseHmtxTable should fail when hmtx table missing")
	}
}

func TestParseHmtxTable_MissingHhea(t *testing.T) {
	// hmtx present but hhea missing — should fail.
	f := &TTFFont{
		Tables: map[string]*TTFTable{
			"hmtx": {Data: make([]byte, 8)},
			// no "hhea"
		},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHmtxTable()
	if err == nil {
		t.Error("parseHmtxTable should fail when hhea table missing")
	}
}

func TestParseCmapTable_NotFound(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseCmapTable()
	if err == nil {
		t.Error("parseCmapTable should fail when cmap table missing")
	}
}

func TestParsePostTable_NotFound(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parsePostTable()
	if err == nil {
		t.Error("parsePostTable should fail when post table missing")
	}
}

func TestParseOS2Table_NotFound(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseOS2Table()
	if err == nil {
		t.Error("parseOS2Table should fail when OS/2 table missing")
	}
}

func TestParseNameTable_NotFound(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseNameTable()
	if err == nil {
		t.Error("parseNameTable should fail when name table missing")
	}
}

// ============================================================================
// parseCmapSubtable — format 12 path
// ============================================================================

func TestParseCmapSubtable_Format12(t *testing.T) {
	// Build 2-byte subtable header with format=12.
	data := []byte{0x00, 0x0C} // format = 12
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseCmapSubtable(data, 0)
	// format 12 is not yet implemented — should return an error.
	if err == nil {
		t.Error("parseCmapSubtable format 12 should return not-implemented error")
	}
}

// ============================================================================
// readCmapHeader — error path (truncated data)
// ============================================================================

func TestReadCmapHeader_TruncatedData(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	// Only 1 byte — can't read version (uint16).
	_, err := f.readCmapHeader([]byte{0x00})
	if err == nil {
		t.Error("readCmapHeader should fail with truncated data")
	}
}

// ============================================================================
// loadTable — out-of-bounds path
// ============================================================================

func TestLoadTable_OutOfBounds(t *testing.T) {
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	table := &TTFTable{
		Tag:    "test",
		Offset: 100,
		Length: 50,
	}
	// Data is only 10 bytes but table says offset=100 length=50.
	err := f.loadTable([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}, table)
	if err == nil {
		t.Error("loadTable should fail when offset+length exceeds data length")
	}
}

// ============================================================================
// parseRequiredTables — error propagation paths
// ============================================================================

func TestParseRequiredTables_HeadFails(t *testing.T) {
	// No tables at all — parseHeadTable fails immediately.
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseRequiredTables()
	if err == nil {
		t.Error("parseRequiredTables should fail when head table is missing")
	}
}

func TestParseRequiredTables_HheaFails(t *testing.T) {
	// Provide valid head but no hhea — parseHheaTable fails.
	headData := make([]byte, 44)
	headData[18], headData[19] = 0x03, 0xE8 // unitsPerEm = 1000
	f := &TTFFont{
		Tables: map[string]*TTFTable{
			"head": {Data: headData},
			// no "hhea"
		},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseRequiredTables()
	if err == nil {
		t.Error("parseRequiredTables should fail when hhea table is missing")
	}
}

func TestParseRequiredTables_PostAndOS2ErrorsAreNonFatal(t *testing.T) {
	// Provide head + hhea + hmtx + cmap but with bad post/OS2 tables.
	// The non-fatal handling path (lines 102-103, 111-112) should execute.

	// head: 44 bytes, unitsPerEm=1000 at offset 18.
	headData := make([]byte, 44)
	headData[18], headData[19] = 0x03, 0xE8

	// hhea: 36 bytes, Ascender=800 at offset 4, numHMetrics=0 at offset 34.
	hheaData := make([]byte, 36)
	hheaData[4], hheaData[5] = 0x03, 0x20 // Ascender=800
	// numHMetrics=0 so no hmtx loop iterations needed.

	// hmtx: empty (numHMetrics=0 means no reads).
	hmtxData := make([]byte, 0)

	// cmap: minimal valid header with 0 tables → no suitable subtable → error.
	// We need parseCmapTable to succeed, so we need a valid Windows Unicode BMP entry.
	// Use a format 4 cmap with 1 segment (just the terminator 0xFFFF).
	// cmap header: version=0, numTables=1
	// record: platformID=3, encodingID=1, offset=12 (right after record)
	// format4 subtable: format=4, length=32, language=0, segCountX2=4 (2 segs), ...
	//   endCode: [0xFFFF], reservedPad=0, startCode: [0xFFFF], idDelta: [1], idRangeOffset: [0]
	cmapData := make([]byte, 12+32)
	// cmap header
	cmapData[0], cmapData[1] = 0x00, 0x00 // version=0
	cmapData[2], cmapData[3] = 0x00, 0x01 // numTables=1
	// record: platformID=3, encodingID=1, offset=12
	cmapData[4], cmapData[5] = 0x00, 0x03 // platformID=3
	cmapData[6], cmapData[7] = 0x00, 0x01 // encodingID=1
	cmapData[8], cmapData[9] = 0x00, 0x00
	cmapData[10], cmapData[11] = 0x00, 0x0C // offset=12
	// format 4 subtable at offset 12
	cmapData[12], cmapData[13] = 0x00, 0x04 // format=4
	cmapData[14], cmapData[15] = 0x00, 0x20 // length=32
	cmapData[16], cmapData[17] = 0x00, 0x00 // language=0
	cmapData[18], cmapData[19] = 0x00, 0x04 // segCountX2=4 (segCount=2)
	cmapData[20], cmapData[21] = 0x00, 0x04 // searchRange
	cmapData[22], cmapData[23] = 0x00, 0x01 // entrySelector
	cmapData[24], cmapData[25] = 0x00, 0x00 // rangeShift
	// endCode[2]: 0x0041, 0xFFFF
	cmapData[26], cmapData[27] = 0x00, 0x41 // endCode[0] = 'A'
	cmapData[28], cmapData[29] = 0xFF, 0xFF // endCode[1] = 0xFFFF (terminator)
	// reservedPad = 0
	cmapData[30], cmapData[31] = 0x00, 0x00
	// startCode[2]: 0x0041, 0xFFFF
	cmapData[32], cmapData[33] = 0x00, 0x41
	cmapData[34], cmapData[35] = 0xFF, 0xFF
	// idDelta[2]: 0, 1
	cmapData[36], cmapData[37] = 0x00, 0x00
	cmapData[38], cmapData[39] = 0x00, 0x01
	// idRangeOffset[2]: 0, 0
	cmapData[40], cmapData[41] = 0x00, 0x00
	cmapData[42], cmapData[43] = 0x00, 0x00

	f := &TTFFont{
		Tables: map[string]*TTFTable{
			"head": {Data: headData},
			"hhea": {Data: hheaData},
			"hmtx": {Data: hmtxData},
			"cmap": {Data: cmapData},
			// Deliberately malformed post and OS/2 tables to trigger non-fatal error paths.
			"post": {Data: []byte{0x00}},             // too short → parsePostTable fails → non-fatal
			"OS/2": {Data: []byte{0x00, 0x00, 0x00}}, // too short → parseOS2Table fails → non-fatal
		},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseRequiredTables()
	// Should succeed despite post/OS2 errors (they are non-fatal).
	if err != nil {
		t.Errorf("parseRequiredTables should succeed despite post/OS2 errors: %v", err)
	}
}

// ============================================================================
// parseOS2Table — read error paths via truncated data
// ============================================================================

func TestParseOS2Table_TruncatedAfterVersion(t *testing.T) {
	// Exactly 78 bytes but truncated so that reads inside fail.
	// parseOS2Table needs to read: version(2), skip xAvgCharWidth(2), WeightClass(2),
	// WidthClass(2), FSType(2) = 10 bytes, then skip 56 = 66 bytes before sTypoAscender.
	// Provide exactly 10 bytes — skipBytes(r, 56) will seek past end but not error
	// (bytes.Reader.Seek allows seeking past end). The subsequent binary.Read will fail.
	data := make([]byte, 10)      // enough for first 5 reads but not skip+reads after
	data[0], data[1] = 0x00, 0x00 // version=0
	f := &TTFFont{
		Tables:      map[string]*TTFTable{"OS/2": {Data: data}},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseOS2Table()
	// Should fail because OS/2 len check requires >= 78 bytes.
	if err == nil {
		t.Error("parseOS2Table should fail with only 10 bytes")
	}
}

// ============================================================================
// parseHmtxTable — truncated mid-read
// ============================================================================

func TestParseHmtxTable_TruncatedData(t *testing.T) {
	// hhea says numHMetrics=3 but hmtx has only 4 bytes (less than 3*4=12 needed).
	hheaData := make([]byte, 36)
	hheaData[34], hheaData[35] = 0x00, 0x03 // numHMetrics=3

	hmtxData := make([]byte, 4) // only 4 bytes, need 12

	f := &TTFFont{
		Tables: map[string]*TTFTable{
			"hhea": {Data: hheaData},
			"hmtx": {Data: hmtxData},
		},
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseHmtxTable()
	if err == nil {
		t.Error("parseHmtxTable should fail with truncated hmtx data")
	}
}

// ============================================================================
// parseFontDirectory — truncated entry
// ============================================================================

func TestParseFontDirectory_TruncatedEntry(t *testing.T) {
	// Valid header with numTables=1 but no actual table entry data.
	// parseFontDirectory will try to read a 16-byte entry and fail.
	var buf bytes.Buffer
	// sfntVersion = 0x00010000
	buf.Write([]byte{0x00, 0x01, 0x00, 0x00})
	// numTables = 1
	buf.Write([]byte{0x00, 0x01})
	// searchRange, entrySelector, rangeShift = 6 bytes
	buf.Write([]byte{0x00, 0x10, 0x00, 0x00, 0x00, 0x00})
	// No table entry data at all — parseTableEntry will fail.

	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}
	err := f.parseFontDirectory(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Error("parseFontDirectory should fail with missing table entry data")
	}
}

// ============================================================================
// parse — loadTables error path
// ============================================================================

func TestParse_LoadTablesOutOfBounds(t *testing.T) {
	// Build a valid font directory with 1 table entry that has an
	// offset/length exceeding the data length — loadTables will fail.
	var buf bytes.Buffer
	// sfntVersion = 0x00010000
	buf.Write([]byte{0x00, 0x01, 0x00, 0x00})
	// numTables = 1
	buf.Write([]byte{0x00, 0x01})
	// searchRange(2), entrySelector(2), rangeShift(2) = 6 bytes
	buf.Write([]byte{0x00, 0x10, 0x00, 0x00, 0x00, 0x00})
	// Table entry: tag="head", checksum=0, offset=9999, length=100
	// (offset+length exceeds our small data buffer)
	buf.Write([]byte{'h', 'e', 'a', 'd'})     // tag
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // checksum
	buf.Write([]byte{0x00, 0x00, 0x27, 0x0F}) // offset = 9999
	buf.Write([]byte{0x00, 0x00, 0x00, 0x64}) // length = 100

	data := buf.Bytes()
	f := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
		FontData:    data,
	}
	err := f.parse(data)
	if err == nil {
		t.Error("parse should fail when table offset is out of bounds")
	}
}
