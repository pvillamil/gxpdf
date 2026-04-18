package fonts

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"sort"
)

// FontSubset represents a subset of a font containing only used glyphs.
//
// Font subsetting reduces PDF file size by embedding only the glyphs
// that are actually used in the document, rather than the entire font.
//
// Benefits:
//   - Smaller file size (especially for large CJK fonts)
//   - Faster loading
//   - Lower memory usage
type FontSubset struct {
	// BaseFont is the original font.
	BaseFont *TTFFont

	// UsedChars is the set of characters used in the document.
	UsedChars map[rune]bool

	// GlyphMapping maps old glyph IDs to new glyph IDs.
	GlyphMapping map[uint16]uint16

	// SubsetData is the compressed font data (for embedding).
	SubsetData []byte
}

// NewFontSubset creates a new font subset from a TTF font.
func NewFontSubset(font *TTFFont) *FontSubset {
	return &FontSubset{
		BaseFont:     font,
		UsedChars:    make(map[rune]bool),
		GlyphMapping: make(map[uint16]uint16),
	}
}

// UseChar marks a character as used in the document.
func (s *FontSubset) UseChar(ch rune) {
	s.UsedChars[ch] = true
}

// UseString marks all characters in a string as used.
func (s *FontSubset) UseString(text string) {
	for _, ch := range text {
		s.UseChar(ch)
	}
}

// Build builds the font subset.
//
// This process:
//  1. Identifies all glyphs used (via character mapping)
//  2. Creates a new glyph ID mapping
//  3. Builds subset font data
//  4. Compresses the data
//
// Returns an error if subsetting fails.
func (s *FontSubset) Build() error {
	// Identify used glyphs.
	usedGlyphs := s.identifyUsedGlyphs()

	// Create glyph mapping (old ID -> new ID).
	s.createGlyphMapping(usedGlyphs)

	// For MVP, we'll embed the full font data (no actual subsetting).
	// Real subsetting requires rebuilding TTF tables, which is complex.
	// This is acceptable for MVP - subsetting can be optimized later.
	if err := s.compressFont(); err != nil {
		return fmt.Errorf("compress font: %w", err)
	}

	return nil
}

// identifyUsedGlyphs identifies which glyphs are used.
func (s *FontSubset) identifyUsedGlyphs() []uint16 {
	glyphSet := make(map[uint16]bool)

	// Always include glyph 0 (.notdef).
	glyphSet[0] = true

	// Add glyphs for used characters.
	for ch := range s.UsedChars {
		if glyphID, ok := s.BaseFont.CharToGlyph[ch]; ok {
			glyphSet[glyphID] = true
		}
	}

	// Convert to sorted slice.
	glyphs := make([]uint16, 0, len(glyphSet))
	for gid := range glyphSet {
		glyphs = append(glyphs, gid)
	}
	sort.Slice(glyphs, func(i, j int) bool {
		return glyphs[i] < glyphs[j]
	})

	return glyphs
}

// createGlyphMapping creates mapping from old to new glyph IDs.
func (s *FontSubset) createGlyphMapping(usedGlyphs []uint16) {
	for i, oldID := range usedGlyphs {
		//nolint:gosec // Index is bounded by usedGlyphs length (< 65536).
		s.GlyphMapping[oldID] = uint16(i)
	}
}

// compressFont compresses the font data using FlateDecode.
func (s *FontSubset) compressFont() error {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)

	if _, err := w.Write(s.BaseFont.FontData); err != nil {
		_ = w.Close() // Best effort cleanup.
		return fmt.Errorf("write font data: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close zlib writer: %w", err)
	}

	s.SubsetData = buf.Bytes()
	return nil
}

// GetCharWidth returns the width of a character in font units.
func (s *FontSubset) GetCharWidth(ch rune) uint16 {
	glyphID, ok := s.BaseFont.CharToGlyph[ch]
	if !ok {
		return 0
	}

	width, ok := s.BaseFont.GlyphWidths[glyphID]
	if !ok {
		return 0
	}

	return width
}

// MeasureString returns the width of a string in points.
func (s *FontSubset) MeasureString(text string, size float64) float64 {
	var totalWidth int
	for _, ch := range text {
		totalWidth += int(s.GetCharWidth(ch))
	}

	// Convert from font units to points.
	unitsPerEm := float64(s.BaseFont.UnitsPerEm)
	if unitsPerEm == 0 {
		unitsPerEm = 1000 // Fallback.
	}

	return float64(totalWidth) * size / unitsPerEm
}

// GetWidths returns an array of character widths for PDF /Widths array.
//
// The /Widths array in PDF specifies the width of each character
// from firstChar to lastChar.
func (s *FontSubset) GetWidths() (firstChar, lastChar int, widths []int) {
	if len(s.UsedChars) == 0 {
		return 0, 0, nil
	}

	// Find min/max character codes.
	first := int(^uint(0) >> 1) // Max int.
	last := 0

	for ch := range s.UsedChars {
		code := int(ch)
		if code < first {
			first = code
		}
		if code > last {
			last = code
		}
	}

	// Build widths array.
	widths = make([]int, last-first+1)
	for i := range widths {
		ch := rune(first + i)
		width := s.GetCharWidth(ch)
		widths[i] = int(width)
	}

	return first, last, widths
}
