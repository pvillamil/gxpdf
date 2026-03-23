package builder

import (
	"fmt"
	"strconv"
	"strings"
)

// Predefined colors available for use in builder options.
// All components are in the [0, 1] range.
var (
	// Black is opaque black (0, 0, 0).
	Black = Color{R: 0, G: 0, B: 0}
	// White is opaque white (1, 1, 1).
	White = Color{R: 1, G: 1, B: 1}
	// Red is pure red (1, 0, 0).
	Red = Color{R: 1, G: 0, B: 0}
	// Green is mid-tone green (0, 0.5, 0) — matches CSS "green".
	Green = Color{R: 0, G: 0.5, B: 0}
	// Blue is pure blue (0, 0, 1).
	Blue = Color{R: 0, G: 0, B: 1}
	// Navy is dark navy blue (#1A237E).
	Navy = Color{R: 0.102, G: 0.137, B: 0.494}
	// Gray is medium gray (0.5, 0.5, 0.5).
	Gray = Color{R: 0.5, G: 0.5, B: 0.5}
	// LightGray is near-white gray (#F5F5F5) — good for zebra-stripe backgrounds.
	LightGray = Color{R: 0.961, G: 0.961, B: 0.961}
	// DarkGray is dark gray (#555555).
	DarkGray = Color{R: 0.333, G: 0.333, B: 0.333}
	// Yellow is pure yellow (1, 1, 0).
	Yellow = Color{R: 1, G: 1, B: 0}
	// Orange is standard orange (1, 0.647, 0) — matches CSS "orange".
	Orange = Color{R: 1, G: 0.647, B: 0}
	// Purple is standard purple (0.5, 0, 0.5) — matches CSS "purple".
	Purple = Color{R: 0.5, G: 0, B: 0.5}
	// Cyan is pure cyan (0, 1, 1).
	Cyan = Color{R: 0, G: 1, B: 1}
)

// Hex parses a CSS-style hex color string into a Color.
// It accepts both "#RRGGBB" and "RRGGBB" formats (case-insensitive).
// If the string cannot be parsed, Black is returned.
//
// Example:
//
//	col := builder.Hex("#1A237E")  // Navy blue
//	col := builder.Hex("FF5722")   // Deep orange
func Hex(hex string) Color {
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
	return RGB255(uint8(r), uint8(g), uint8(b))
}

// hexToString converts a Color back to a "#RRGGBB" string for debugging.
func hexToString(c Color) string {
	r := uint8(c.R * 255)
	g := uint8(c.G * 255)
	b := uint8(c.B * 255)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
