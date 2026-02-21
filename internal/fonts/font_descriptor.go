package fonts

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FontDescriptor represents a PDF FontDescriptor dictionary.
//
// The FontDescriptor specifies metrics and other attributes of a font.
// It is required for embedded fonts in PDF documents.
//
// Reference: PDF Reference 1.7, Section 9.8.
type FontDescriptor struct {
	// FontName is the PostScript name of the font.
	FontName string

	// Flags is the font flags bitmap (PDF spec Table 123).
	Flags uint32

	// FontBBox is the bounding box [llx lly urx ury] in glyph space.
	FontBBox [4]int

	// ItalicAngle is the angle of italic text in degrees.
	ItalicAngle float64

	// Ascent is the maximum height above baseline.
	Ascent int

	// Descent is the maximum depth below baseline (negative).
	Descent int

	// CapHeight is the height of capital letters.
	CapHeight int

	// StemV is the dominant vertical stem width.
	StemV int

	// XHeight is the height of lowercase x (optional).
	XHeight int

	// Leading is the spacing between lines (optional).
	Leading int

	// FontFile2Ref is the object number of the embedded font stream.
	// Set to 0 if font is not embedded.
	FontFile2Ref int
}

// GenerateFontDescriptor creates a FontDescriptor from TTF font data.
//
// This extracts all required metrics from the parsed TTF font and
// converts them to PDF glyph space (scaled by 1000/UnitsPerEm).
func GenerateFontDescriptor(ttf *TTFFont) *FontDescriptor {
	if ttf == nil {
		return nil
	}

	// Calculate scale factor (PDF uses 1000 units per em).
	scale := 1000.0 / float64(ttf.UnitsPerEm)

	// Get PostScript name or derive from filename.
	fontName := ttf.PostScriptName
	if fontName == "" {
		// Derive from filename: /path/to/OpenSans-Regular.ttf -> OpenSans-Regular
		base := filepath.Base(ttf.FilePath)
		fontName = strings.TrimSuffix(base, filepath.Ext(base))
		// Remove spaces (PostScript names can't have spaces).
		fontName = strings.ReplaceAll(fontName, " ", "")
	}

	return &FontDescriptor{
		FontName:    fontName,
		Flags:       ttf.Flags,
		FontBBox:    scaleFontBBox(ttf.FontBBox, scale),
		ItalicAngle: ttf.ItalicAngle,
		Ascent:      scaleMetric(ttf.Ascender, scale),
		Descent:     scaleMetric(ttf.Descender, scale),
		CapHeight:   scaleMetric(ttf.CapHeight, scale),
		StemV:       int(ttf.StemV),
		XHeight:     scaleMetric(ttf.XHeight, scale),
		Leading:     scaleMetric(ttf.LineGap, scale),
	}
}

// scaleFontBBox scales the font bounding box to PDF units.
func scaleFontBBox(bbox [4]int16, scale float64) [4]int {
	return [4]int{
		int(float64(bbox[0]) * scale),
		int(float64(bbox[1]) * scale),
		int(float64(bbox[2]) * scale),
		int(float64(bbox[3]) * scale),
	}
}

// scaleMetric scales a single metric value to PDF units.
func scaleMetric(value int16, scale float64) int {
	return int(float64(value) * scale)
}

// ToPDFDict generates the PDF FontDescriptor dictionary as bytes.
//
// The output format:
//
//	<<
//	/Type /FontDescriptor
//	/FontName /FontName
//	/Flags 32
//	/FontBBox [0 -200 1000 800]
//	/ItalicAngle 0
//	/Ascent 800
//	/Descent -200
//	/CapHeight 700
//	/StemV 80
//	/FontFile2 X 0 R
//	>>
func (fd *FontDescriptor) ToPDFDict(fontFile2ObjNum int) string {
	var sb strings.Builder

	sb.WriteString("<<\n")
	sb.WriteString("/Type /FontDescriptor\n")
	fmt.Fprintf(&sb, "/FontName /%s\n", fd.FontName)
	fmt.Fprintf(&sb, "/Flags %d\n", fd.Flags)
	fmt.Fprintf(&sb, "/FontBBox [%d %d %d %d]\n",
		fd.FontBBox[0], fd.FontBBox[1], fd.FontBBox[2], fd.FontBBox[3])
	fmt.Fprintf(&sb, "/ItalicAngle %.1f\n", fd.ItalicAngle)
	fmt.Fprintf(&sb, "/Ascent %d\n", fd.Ascent)
	fmt.Fprintf(&sb, "/Descent %d\n", fd.Descent)
	fmt.Fprintf(&sb, "/CapHeight %d\n", fd.CapHeight)
	fmt.Fprintf(&sb, "/StemV %d\n", fd.StemV)

	if fd.XHeight > 0 {
		fmt.Fprintf(&sb, "/XHeight %d\n", fd.XHeight)
	}

	if fontFile2ObjNum > 0 {
		fmt.Fprintf(&sb, "/FontFile2 %d 0 R\n", fontFile2ObjNum)
	}

	sb.WriteString(">>")

	return sb.String()
}

// SubsetFontName generates a subset font name with random prefix.
//
// PDF subset font names use a 6-letter uppercase prefix followed by '+'.
// Example: ABCDEF+OpenSans-Regular
//
// The prefix should be unique to allow multiple subsets of the same font.
func SubsetFontName(baseName string, usedChars []rune) string {
	// Generate prefix from hash of used characters.
	// This ensures same characters = same prefix (deterministic).
	hash := uint32(0)
	for _, r := range usedChars {
		hash = hash*31 + uint32(r)
	}

	// Convert to 6 uppercase letters (A-Z).
	prefix := make([]byte, 6)
	for i := 0; i < 6; i++ {
		prefix[i] = byte('A' + (hash % 26))
		hash /= 26
	}

	return string(prefix) + "+" + baseName
}
