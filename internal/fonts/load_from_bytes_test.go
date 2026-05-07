package fonts

import (
	"os"
	"testing"
)

// dejaVuFontPath is defined in coverage_boost_test.go within this package.
// We reuse it here to avoid duplication.

// TestLoadTTFFromBytes_RoundTrip verifies that LoadTTFFromBytes produces the
// same parsed metrics as LoadTTF when given the same raw bytes.
func TestLoadTTFFromBytes_RoundTrip(t *testing.T) {
	// Load from file first.
	fileFont, err := LoadTTF(dejaVuFontPath)
	if err != nil {
		t.Skipf("DejaVu font not available: %v", err)
	}

	// Load the same bytes via LoadTTFFromBytes.
	byteFont, err := LoadTTFFromBytes(fileFont.FontData)
	if err != nil {
		t.Fatalf("LoadTTFFromBytes failed: %v", err)
	}

	if byteFont == nil {
		t.Fatal("LoadTTFFromBytes returned nil")
	}

	// FilePath should be empty (loaded from bytes, not disk).
	if byteFont.FilePath != "" {
		t.Errorf("FilePath = %q, want empty string for in-memory font", byteFont.FilePath)
	}

	// Core metrics must match.
	if byteFont.PostScriptName != fileFont.PostScriptName {
		t.Errorf("PostScriptName: from bytes=%q, from file=%q", byteFont.PostScriptName, fileFont.PostScriptName)
	}
	if byteFont.UnitsPerEm != fileFont.UnitsPerEm {
		t.Errorf("UnitsPerEm: from bytes=%d, from file=%d", byteFont.UnitsPerEm, fileFont.UnitsPerEm)
	}
	if byteFont.Ascender != fileFont.Ascender {
		t.Errorf("Ascender: from bytes=%d, from file=%d", byteFont.Ascender, fileFont.Ascender)
	}
	if byteFont.Descender != fileFont.Descender {
		t.Errorf("Descender: from bytes=%d, from file=%d", byteFont.Descender, fileFont.Descender)
	}
	if byteFont.CapHeight != fileFont.CapHeight {
		t.Errorf("CapHeight: from bytes=%d, from file=%d", byteFont.CapHeight, fileFont.CapHeight)
	}

	// Character mapping completeness.
	if len(byteFont.CharToGlyph) != len(fileFont.CharToGlyph) {
		t.Errorf("CharToGlyph len: from bytes=%d, from file=%d",
			len(byteFont.CharToGlyph), len(fileFont.CharToGlyph))
	}

	// Glyph widths completeness.
	if len(byteFont.GlyphWidths) != len(fileFont.GlyphWidths) {
		t.Errorf("GlyphWidths len: from bytes=%d, from file=%d",
			len(byteFont.GlyphWidths), len(fileFont.GlyphWidths))
	}

	// FontData must be preserved as-is.
	if len(byteFont.FontData) != len(fileFont.FontData) {
		t.Errorf("FontData len: from bytes=%d, from file=%d",
			len(byteFont.FontData), len(fileFont.FontData))
	}
}

// TestLoadTTFFromBytes_MetricsValid verifies basic sanity of parsed metrics.
func TestLoadTTFFromBytes_MetricsValid(t *testing.T) {
	fileFont, err := LoadTTF(dejaVuFontPath)
	if err != nil {
		t.Skipf("DejaVu font not available: %v", err)
	}

	byteFont, err := LoadTTFFromBytes(fileFont.FontData)
	if err != nil {
		t.Fatalf("LoadTTFFromBytes failed: %v", err)
	}

	if byteFont.UnitsPerEm == 0 {
		t.Error("UnitsPerEm should not be 0")
	}
	if byteFont.Ascender <= 0 {
		t.Errorf("Ascender should be positive, got %d", byteFont.Ascender)
	}
	if byteFont.Descender >= 0 {
		t.Errorf("Descender should be negative, got %d", byteFont.Descender)
	}
	if len(byteFont.CharToGlyph) == 0 {
		t.Error("CharToGlyph should not be empty")
	}
	if len(byteFont.GlyphWidths) == 0 {
		t.Error("GlyphWidths should not be empty")
	}

	// ASCII characters should all be present in a real font.
	for _, ch := range "ABCabc012" {
		if _, ok := byteFont.CharToGlyph[ch]; !ok {
			t.Errorf("ASCII char %q not in CharToGlyph", ch)
		}
	}
}

// TestLoadTTFFromBytes_InvalidData verifies rejection of non-TTF bytes.
func TestLoadTTFFromBytes_InvalidData(t *testing.T) {
	_, err := LoadTTFFromBytes([]byte("this is not a TTF font"))
	if err == nil {
		t.Error("LoadTTFFromBytes should fail for invalid data")
	}
}

// TestLoadTTFFromBytes_EmptyData verifies rejection of empty byte slice.
func TestLoadTTFFromBytes_EmptyData(t *testing.T) {
	_, err := LoadTTFFromBytes([]byte{})
	if err == nil {
		t.Error("LoadTTFFromBytes should fail for empty data")
	}
}

// TestLoadTTFFromBytes_Nil verifies rejection of nil byte slice.
func TestLoadTTFFromBytes_Nil(t *testing.T) {
	_, err := LoadTTFFromBytes(nil)
	if err == nil {
		t.Error("LoadTTFFromBytes should fail for nil data")
	}
}

// TestLoadTTFFromBytes_PreservesAllBytes verifies FontData is not truncated or altered.
func TestLoadTTFFromBytes_PreservesAllBytes(t *testing.T) {
	data, err := os.ReadFile(dejaVuFontPath)
	if err != nil {
		t.Skipf("DejaVu font not available: %v", err)
	}

	byteFont, err := LoadTTFFromBytes(data)
	if err != nil {
		t.Fatalf("LoadTTFFromBytes failed: %v", err)
	}

	if len(byteFont.FontData) != len(data) {
		t.Errorf("FontData len mismatch: got %d, want %d", len(byteFont.FontData), len(data))
	}
	for i := range data {
		if byteFont.FontData[i] != data[i] {
			t.Errorf("FontData differs at byte %d", i)
			break
		}
	}
}

// TestLoadTTFFromBytes_CanBuildSubset verifies that a font loaded from bytes
// can have a subset built from it, confirming full round-trip usability.
func TestLoadTTFFromBytes_CanBuildSubset(t *testing.T) {
	fileFont, err := LoadTTF(dejaVuFontPath)
	if err != nil {
		t.Skipf("DejaVu font not available: %v", err)
	}

	byteFont, err := LoadTTFFromBytes(fileFont.FontData)
	if err != nil {
		t.Fatalf("LoadTTFFromBytes failed: %v", err)
	}

	subset := NewFontSubset(byteFont)
	subset.UseString("Hello PDF")

	if err := subset.Build(); err != nil {
		t.Fatalf("Build() on bytes-loaded font failed: %v", err)
	}
	if len(subset.SubsetData) == 0 {
		t.Error("SubsetData should be non-empty after Build()")
	}
}
