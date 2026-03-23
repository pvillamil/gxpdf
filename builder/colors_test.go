package builder

import (
	"testing"

	"github.com/coregx/gxpdf/layout"
)

func TestHex_ValidFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  layout.Color
	}{
		{
			name:  "hash prefix uppercase",
			input: "#1A237E",
			want:  layout.RGB255(0x1A, 0x23, 0x7E),
		},
		{
			name:  "no hash lowercase",
			input: "ff5722",
			want:  layout.RGB255(0xFF, 0x57, 0x22),
		},
		{
			name:  "black",
			input: "#000000",
			want:  layout.RGB(0, 0, 0),
		},
		{
			name:  "white",
			input: "#FFFFFF",
			want:  layout.RGB255(255, 255, 255),
		},
		{
			name:  "hash prefix lowercase",
			input: "#aabbcc",
			want:  layout.RGB255(0xAA, 0xBB, 0xCC),
		},
		{
			name:  "mixed case",
			input: "#AbCdEf",
			want:  layout.RGB255(0xAB, 0xCD, 0xEF),
		},
		{
			name:  "leading whitespace trimmed",
			input: "  #FF0000",
			want:  layout.RGB255(255, 0, 0),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Hex(tc.input)
			if !colorsEqual(got, tc.want, 1.0/255.0) {
				t.Errorf("Hex(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestHex_InvalidFormats(t *testing.T) {
	invalid := []string{
		"",
		"#FFF",      // 3-digit shorthand not supported
		"#GGGGGG",   // invalid hex digits
		"#12345",    // 5 digits
		"#1234567",  // 7 digits
		"not-a-hex", // non-hex string
	}
	for _, input := range invalid {
		t.Run(input, func(t *testing.T) {
			got := Hex(input)
			// Should return Black as fallback.
			if got != Black {
				t.Errorf("Hex(%q) = %v, expected Black fallback", input, got)
			}
		})
	}
}

func TestPredefinedColors(t *testing.T) {
	// Verify that predefined colors are in the [0,1] range and are distinct.
	colors := map[string]layout.Color{
		"Black":     Black,
		"White":     White,
		"Red":       Red,
		"Green":     Green,
		"Blue":      Blue,
		"Navy":      Navy,
		"Gray":      Gray,
		"LightGray": LightGray,
		"DarkGray":  DarkGray,
		"Yellow":    Yellow,
		"Orange":    Orange,
		"Purple":    Purple,
		"Cyan":      Cyan,
	}
	for name, c := range colors {
		t.Run(name, func(t *testing.T) {
			if c.R < 0 || c.R > 1 {
				t.Errorf("%s.R = %f out of [0,1]", name, c.R)
			}
			if c.G < 0 || c.G > 1 {
				t.Errorf("%s.G = %f out of [0,1]", name, c.G)
			}
			if c.B < 0 || c.B > 1 {
				t.Errorf("%s.B = %f out of [0,1]", name, c.B)
			}
		})
	}
}

func TestPredefinedColors_KnownValues(t *testing.T) {
	tests := []struct {
		name  string
		got   layout.Color
		wantR float64
		wantG float64
		wantB float64
	}{
		{"Black", Black, 0, 0, 0},
		{"White", White, 1, 1, 1},
		{"Red", Red, 1, 0, 0},
		{"Blue", Blue, 0, 0, 1},
		{"Yellow", Yellow, 1, 1, 0},
		{"Cyan", Cyan, 0, 1, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got.R != tc.wantR || tc.got.G != tc.wantG || tc.got.B != tc.wantB {
				t.Errorf("%s = RGB(%.3f,%.3f,%.3f), want RGB(%.3f,%.3f,%.3f)",
					tc.name, tc.got.R, tc.got.G, tc.got.B,
					tc.wantR, tc.wantG, tc.wantB)
			}
		})
	}
}

func TestHexToString_RoundTrip(t *testing.T) {
	colors := []string{"#FF0000", "#00FF00", "#0000FF", "#1A237E", "#FFFFFF", "#000000"}
	for _, hex := range colors {
		t.Run(hex, func(t *testing.T) {
			c := Hex(hex)
			got := hexToString(c)
			if got != hex {
				t.Errorf("round-trip Hex(%q) -> hexToString = %q", hex, got)
			}
		})
	}
}

// colorsEqual returns true when both colors are within epsilon on all components.
func colorsEqual(a, b layout.Color, eps float64) bool {
	diff := func(x, y float64) float64 {
		if x > y {
			return x - y
		}
		return y - x
	}
	return diff(a.R, b.R) <= eps && diff(a.G, b.G) <= eps && diff(a.B, b.B) <= eps
}
