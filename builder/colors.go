package builder

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coregx/gxpdf/layout"
)

// Predefined colors available for use in builder options.
// All components are in the [0, 1] range as required by layout.Color.
var (
	// Black is opaque black (0, 0, 0).
	Black = layout.RGB(0, 0, 0)
	// White is opaque white (1, 1, 1).
	White = layout.RGB(1, 1, 1)
	// Red is pure red (1, 0, 0).
	Red = layout.RGB(1, 0, 0)
	// Green is mid-tone green (0, 0.5, 0) — matches CSS "green".
	Green = layout.RGB(0, 0.5, 0)
	// Blue is pure blue (0, 0, 1).
	Blue = layout.RGB(0, 0, 1)
	// Navy is dark navy blue (#1A237E).
	Navy = layout.RGB(0.102, 0.137, 0.494)
	// Gray is medium gray (0.5, 0.5, 0.5).
	Gray = layout.RGB(0.5, 0.5, 0.5)
	// LightGray is near-white gray (#F5F5F5) — good for zebra-stripe backgrounds.
	LightGray = layout.RGB(0.961, 0.961, 0.961)
	// DarkGray is dark gray (#555555).
	DarkGray = layout.RGB(0.333, 0.333, 0.333)
	// Yellow is pure yellow (1, 1, 0).
	Yellow = layout.RGB(1, 1, 0)
	// Orange is standard orange (1, 0.647, 0) — matches CSS "orange".
	Orange = layout.RGB(1, 0.647, 0)
	// Purple is standard purple (0.5, 0, 0.5) — matches CSS "purple".
	Purple = layout.RGB(0.5, 0, 0.5)
	// Cyan is pure cyan (0, 1, 1).
	Cyan = layout.RGB(0, 1, 1)
)

// Hex parses a CSS-style hex color string into a layout.Color.
// It accepts both "#RRGGBB" and "RRGGBB" formats (case-insensitive).
// If the string cannot be parsed, Black is returned.
//
// Example:
//
//	col := builder.Hex("#1A237E")  // Navy blue
//	col := builder.Hex("FF5722")   // Deep orange
func Hex(hex string) layout.Color {
	h := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(h) != 6 {
		return Black
	}
	r, err := strconv.ParseUint(h[0:2], 16, 8)
	if err != nil {
		return Black
	}
	g, err := strconv.ParseUint(h[2:4], 16, 8)
	if err != nil {
		return Black
	}
	b, err := strconv.ParseUint(h[4:6], 16, 8)
	if err != nil {
		return Black
	}
	return layout.RGB255(uint8(r), uint8(g), uint8(b))
}

// hexToString converts a layout.Color back to a "#RRGGBB" string for debugging.
func hexToString(c layout.Color) string {
	r := uint8(c.R * 255)
	g := uint8(c.G * 255)
	b := uint8(c.B * 255)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
