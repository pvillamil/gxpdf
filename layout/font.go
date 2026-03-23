package layout

import "strings"

// FontResolver abstracts font measurement from the layout engine.
// Layout never imports creator/ or any PDF font package — it always goes
// through this interface. The concrete implementation in builder/internal
// bridges to the creator/ font subsystem.
type FontResolver interface {
	// MeasureString returns the width of text rendered at the given font and size,
	// in PDF points. It must NOT include trailing whitespace width.
	MeasureString(font FontRef, text string, size float64) float64
	// LineHeight returns the total line height (ascender + descender + leading)
	// for the given font and size, in PDF points.
	LineHeight(font FontRef, size float64) float64
	// Ascender returns the ascender height above the baseline for the given
	// font and size, in PDF points.
	Ascender(font FontRef, size float64) float64
	// Descender returns the magnitude of the descender below the baseline
	// for the given font and size, in PDF points. Always a positive value.
	Descender(font FontRef, size float64) float64
	// LineBreak splits text into lines that each fit within maxWidth points
	// when rendered at the given font and size. It wraps at word boundaries
	// where possible; it may split mid-word only when a single word exceeds
	// maxWidth. Returns at least one element even for empty text.
	LineBreak(font FontRef, text string, size float64, maxWidth float64) []string
}

// FontRef identifies a font by its typographic attributes. It is used as a
// key when looking up metrics from the FontResolver.
type FontRef struct {
	// Family is the font family name (e.g. "Helvetica", "Inter", "Times New Roman").
	Family string
	// Weight selects the font weight within the family.
	Weight FontWeight
	// Style selects the font style (normal or italic/oblique).
	Style FontStyle
}

// FontWeight represents the weight (boldness) of a font.
type FontWeight int

const (
	// WeightNormal is the standard weight (400).
	WeightNormal FontWeight = 400
	// WeightBold is the bold weight (700).
	WeightBold FontWeight = 700
)

// FontStyle represents the style variant of a font.
type FontStyle int

const (
	// StyleNormal is upright (roman) text.
	StyleNormal FontStyle = iota
	// StyleItalic is italic or oblique text.
	StyleItalic
)

// DefaultFont returns a FontRef for the default document font (Helvetica, normal weight and style).
func DefaultFont() FontRef {
	return FontRef{Family: "Helvetica", Weight: WeightNormal, Style: StyleNormal}
}

// MockFontResolver is a deterministic font resolver for use in tests.
// It approximates character width as 0.5 * fontSize and line height as 1.2 * fontSize.
// All methods are safe for concurrent use.
type MockFontResolver struct{}

// MeasureString returns an approximate width using 0.5 * fontSize per rune.
func (m *MockFontResolver) MeasureString(_ FontRef, text string, size float64) float64 {
	return float64(len([]rune(text))) * size * 0.5
}

// LineHeight returns 1.2 * fontSize as the total line height.
func (m *MockFontResolver) LineHeight(_ FontRef, size float64) float64 {
	return size * 1.2
}

// Ascender returns 0.8 * fontSize as the ascender height.
func (m *MockFontResolver) Ascender(_ FontRef, size float64) float64 {
	return size * 0.8
}

// Descender returns 0.2 * fontSize as the descender magnitude.
func (m *MockFontResolver) Descender(_ FontRef, size float64) float64 {
	return size * 0.2
}

// LineBreak splits text into lines that fit within maxWidth using the
// approximate 0.5 * fontSize per character metric.
func (m *MockFontResolver) LineBreak(font FontRef, text string, size float64, maxWidth float64) []string {
	if text == "" {
		return []string{""}
	}

	charWidth := size * 0.5
	if charWidth <= 0 || maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""

	for i, word := range words {
		if i == 0 {
			current = word
			continue
		}
		candidate := current + " " + word
		candidateWidth := m.MeasureString(font, candidate, size)
		if candidateWidth <= maxWidth {
			current = candidate
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	if len(lines) == 0 {
		return []string{text}
	}
	return lines
}
