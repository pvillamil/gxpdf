// Package internal provides internal implementation details for the builder
// package. It must not be imported by external packages.
package internal

import (
	"strings"

	"github.com/coregx/gxpdf/creator"
	"github.com/coregx/gxpdf/layout"
)

// FontBridge implements layout.FontResolver by delegating to creator's font
// subsystem. It bridges between the layout engine (which uses FontRef) and
// the creator package (which has Standard 14 fonts and CustomFont).
//
// Standard 14 font families are resolved by mapping FontRef.Family to
// creator.FontName constants. Custom fonts registered via WithFont/WithFontFile
// are looked up by family name in the customFonts map.
type FontBridge struct {
	// customFonts maps family name to a loaded CustomFont.
	customFonts map[string]*creator.CustomFont
}

// NewFontBridge creates a FontBridge with the given custom font map.
// The map may be empty; Standard 14 fonts are always available.
func NewFontBridge(customFonts map[string]*creator.CustomFont) *FontBridge {
	if customFonts == nil {
		customFonts = make(map[string]*creator.CustomFont)
	}
	return &FontBridge{customFonts: customFonts}
}

// MeasureString implements layout.FontResolver.
// Returns the width of text in PDF points at the given font and size.
func (fb *FontBridge) MeasureString(font layout.FontRef, text string, size float64) float64 {
	if cf, ok := fb.customFonts[font.Family]; ok {
		return cf.MeasureString(text, size)
	}
	return creator.MeasureText(fb.resolveStandard14(font), text, size)
}

// LineHeight implements layout.FontResolver.
// Returns the total line height (ascender + |descender|) in PDF points.
func (fb *FontBridge) LineHeight(font layout.FontRef, size float64) float64 {
	if cf, ok := fb.customFonts[font.Family]; ok {
		return cf.LineHeight(size)
	}
	return creator.FontLineHeight(fb.resolveStandard14(font), size)
}

// Ascender implements layout.FontResolver.
// Returns the ascender height above the baseline in PDF points.
func (fb *FontBridge) Ascender(font layout.FontRef, size float64) float64 {
	if cf, ok := fb.customFonts[font.Family]; ok {
		return cf.Ascender(size)
	}
	return creator.FontAscender(fb.resolveStandard14(font), size)
}

// Descender implements layout.FontResolver.
// Returns the magnitude of the descender below the baseline in PDF points.
// Always returns a positive value.
func (fb *FontBridge) Descender(font layout.FontRef, size float64) float64 {
	if cf, ok := fb.customFonts[font.Family]; ok {
		d := cf.Descender(size)
		if d < 0 {
			d = -d
		}
		return d
	}
	d := creator.FontDescender(fb.resolveStandard14(font), size)
	if d < 0 {
		d = -d
	}
	return d
}

// LineBreak implements layout.FontResolver. It splits text into lines that
// each fit within maxWidth points, wrapping at word boundaries. CJK characters
// are treated as their own break opportunities.
//
// The algorithm:
//  1. Split text into words on whitespace.
//  2. Accumulate words into a line until the next word would exceed maxWidth.
//  3. For words wider than maxWidth (single long token), split at the
//     character level.
//  4. Handle CJK: any CJK rune is a valid break opportunity.
func (fb *FontBridge) LineBreak(font layout.FontRef, text string, size float64, maxWidth float64) []string {
	if text == "" {
		return []string{""}
	}
	if maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""

	for _, word := range words {
		// Check if the word itself contains CJK; if so, handle character-level.
		if containsCJK(word) {
			// Flush current line first.
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			cjkLines := fb.breakCJK(font, word, size, maxWidth)
			if len(cjkLines) > 1 {
				lines = append(lines, cjkLines[:len(cjkLines)-1]...)
				current = cjkLines[len(cjkLines)-1]
			} else if len(cjkLines) == 1 {
				current = cjkLines[0]
			}
			continue
		}

		// Non-CJK word.
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}

		if fb.MeasureString(font, candidate, size) <= maxWidth {
			current = candidate
		} else {
			// Word doesn't fit on current line.
			if current != "" {
				lines = append(lines, current)
			}
			// Check if the word alone exceeds maxWidth.
			if fb.MeasureString(font, word, size) > maxWidth {
				// Force-split the oversized word at character level.
				charLines := fb.breakAtChars(font, word, size, maxWidth)
				if len(charLines) > 1 {
					lines = append(lines, charLines[:len(charLines)-1]...)
					current = charLines[len(charLines)-1]
				} else {
					current = word
				}
			} else {
				current = word
			}
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

// breakAtChars splits a single token at character boundaries to fit maxWidth.
func (fb *FontBridge) breakAtChars(font layout.FontRef, word string, size, maxWidth float64) []string {
	var lines []string
	current := ""

	for _, r := range word {
		candidate := current + string(r)
		if fb.MeasureString(font, candidate, size) <= maxWidth {
			current = candidate
		} else {
			if current != "" {
				lines = append(lines, current)
			}
			current = string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// breakCJK splits CJK text at individual character boundaries (every character
// is a valid break point in CJK text), while still keeping non-CJK runs together.
func (fb *FontBridge) breakCJK(font layout.FontRef, text string, size, maxWidth float64) []string {
	var lines []string
	current := ""

	for _, r := range text {
		candidate := current + string(r)
		if fb.MeasureString(font, candidate, size) <= maxWidth {
			current = candidate
		} else {
			if current != "" {
				lines = append(lines, current)
			}
			current = string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// resolveStandard14 maps a FontRef to a creator.FontName for Standard 14 fonts.
// The mapping uses the family name (case-insensitive prefix matching) and
// falls back to Helvetica if the family is unknown.
func (fb *FontBridge) resolveStandard14(font layout.FontRef) creator.FontName {
	family := strings.ToLower(font.Family)
	bold := font.Weight == layout.WeightBold
	italic := font.Style == layout.StyleItalic

	switch {
	case strings.HasPrefix(family, "times") || strings.HasPrefix(family, "times new roman"):
		switch {
		case bold && italic:
			return creator.TimesBoldItalic
		case bold:
			return creator.TimesBold
		case italic:
			return creator.TimesItalic
		default:
			return creator.TimesRoman
		}
	case strings.HasPrefix(family, "courier"):
		switch {
		case bold && italic:
			return creator.CourierBoldOblique
		case bold:
			return creator.CourierBold
		case italic:
			return creator.CourierOblique
		default:
			return creator.Courier
		}
	case strings.HasPrefix(family, "symbol"):
		return creator.Symbol
	case strings.HasPrefix(family, "zapf"):
		return creator.ZapfDingbats
	default:
		// Default: Helvetica family.
		switch {
		case bold && italic:
			return creator.HelveticaBoldOblique
		case bold:
			return creator.HelveticaBold
		case italic:
			return creator.HelveticaOblique
		default:
			return creator.Helvetica
		}
	}
}

// containsCJK returns true if the string contains any CJK unified ideograph,
// Hiragana, Katakana, Hangul, or CJK compatibility character.
func containsCJK(s string) bool {
	for _, r := range s {
		if isCJKRune(r) {
			return true
		}
	}
	return false
}

// isCJKRune returns true for runes in the primary CJK Unicode ranges.
func isCJKRune(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x20000 && r <= 0x2A6DF) || // CJK Extension B
		(r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana
		(r >= 0xAC00 && r <= 0xD7AF) || // Hangul Syllables
		(r >= 0xF900 && r <= 0xFAFF) // CJK Compatibility Ideographs
}
