package document

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPageSize_ToRectangle_AllSizes verifies that every named PageSize constant
// produces the correct width and height when converted to a Rectangle.
// This is the canonical correctness test: if a size is in the map, its dimensions
// must match the PDF/ISO specification.
func TestPageSize_ToRectangle_AllSizes(t *testing.T) {
	tests := []struct {
		name       string
		pageSize   PageSize
		wantWidth  float64
		wantHeight float64
	}{
		// Original 8 sizes — values MUST NOT change (backward compatibility)
		{"A4", A4, 595, 842},
		{"A3", A3, 842, 1191},
		{"A5", A5, 420, 595},
		{"B4", B4, 709, 1001},
		{"B5", B5, 499, 709},
		{"Letter", Letter, 612, 792},
		{"Legal", Legal, 612, 1008},
		{"Tabloid", Tabloid, 792, 1224},

		// Extended ISO A series
		{"A0", A0, 2384, 3370},
		{"A1", A1, 1684, 2384},
		{"A2", A2, 1191, 1684},
		{"A6", A6, 298, 420},
		{"A7", A7, 210, 298},
		{"A8", A8, 147, 210},

		// Extended ISO B series
		{"B0", B0, 2835, 4008},
		{"B1", B1, 2004, 2835},
		{"B2", B2, 1417, 2004},
		{"B3", B3, 1001, 1417},
		{"B6", B6, 354, 499},

		// ISO C / Envelope sizes
		{"C4", C4, 649, 918},
		{"C5", C5, 459, 649},
		{"C6", C6, 323, 459},
		{"DL", DL, 312, 624},

		// North American extended
		{"Executive", Executive, 522, 756},
		{"HalfLetter", HalfLetter, 396, 612},

		// ANSI Engineering
		{"ANSIC", ANSIC, 1224, 1584},
		{"ANSID", ANSID, 1584, 2448},
		{"ANSIE", ANSIE, 2448, 3168},

		// Photo
		{"Photo4x6", Photo4x6, 288, 432},
		{"Photo5x7", Photo5x7, 360, 504},
		{"Photo8x10", Photo8x10, 576, 720},

		// Book Publishing
		{"Digest", Digest, 396, 612},
		{"USTradeBook", USTradeBook, 432, 648},

		// Presentation (unique feature)
		{"Slide16x9", Slide16x9, 960, 540},
		{"Slide4x3", Slide4x3, 720, 540},

		// US Envelope
		{"Envelope10", Envelope10, 297, 684},

		// JIS B series (Japanese — different from ISO B)
		{"JISB4", JISB4, 729, 1032},
		{"JISB5", JISB5, 516, 729},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rect := tt.pageSize.ToRectangle()

			llx, lly := rect.LowerLeft()
			assert.Equal(t, 0.0, llx, "lower-left X must be 0")
			assert.Equal(t, 0.0, lly, "lower-left Y must be 0")
			assert.Equal(t, tt.wantWidth, rect.Width(), "width mismatch for %s", tt.name)
			assert.Equal(t, tt.wantHeight, rect.Height(), "height mismatch for %s", tt.name)
		})
	}
}

// TestPageSize_ToRectangle_LegacyUnknown verifies that an unrecognized PageSize
// falls back to A4 (the historical default) and does not panic.
func TestPageSize_ToRectangle_LegacyUnknown(t *testing.T) {
	rect := PageSize(999).ToRectangle()
	// Falls back to A4
	assert.Equal(t, 595.0, rect.Width(), "unknown size should fall back to A4 width")
	assert.Equal(t, 842.0, rect.Height(), "unknown size should fall back to A4 height")
}

// TestPageSize_ToRectangle_Custom verifies that the Custom sentinel falls back to
// A4 (users must call CustomPageSize to get a proper custom size).
func TestPageSize_ToRectangle_Custom(t *testing.T) {
	rect := Custom.ToRectangle()
	// Custom without explicit dims falls back to A4
	assert.Equal(t, 595.0, rect.Width())
	assert.Equal(t, 842.0, rect.Height())
}

// TestPageSize_String_AllSizes verifies that every named PageSize constant
// returns a non-empty, meaningful string representation.
func TestPageSize_String_AllSizes(t *testing.T) {
	tests := []struct {
		pageSize PageSize
		want     string
	}{
		{A4, "A4"}, {A3, "A3"}, {A5, "A5"},
		{B4, "B4"}, {B5, "B5"},
		{Letter, "Letter"}, {Legal, "Legal"}, {Tabloid, "Tabloid"},

		{A0, "A0"}, {A1, "A1"}, {A2, "A2"}, {A6, "A6"}, {A7, "A7"}, {A8, "A8"},
		{B0, "B0"}, {B1, "B1"}, {B2, "B2"}, {B3, "B3"}, {B6, "B6"},
		{C4, "C4"}, {C5, "C5"}, {C6, "C6"}, {DL, "DL"},
		{Executive, "Executive"}, {HalfLetter, "HalfLetter"},
		{ANSIC, "ANSIC"}, {ANSID, "ANSID"}, {ANSIE, "ANSIE"},
		{Photo4x6, "Photo4x6"}, {Photo5x7, "Photo5x7"}, {Photo8x10, "Photo8x10"},
		{Digest, "Digest"}, {USTradeBook, "USTradeBook"},
		{Slide16x9, "Slide16x9"}, {Slide4x3, "Slide4x3"},
		{Envelope10, "Envelope10"},
		{JISB4, "JISB4"}, {JISB5, "JISB5"},
		{Custom, "Custom"},
		{PageSize(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.pageSize.String())
		})
	}
}

// TestPageSize_IotaOrder_OriginalEightUnchanged ensures that the iota values for
// the original 8 sizes (the public API contract) have not been accidentally shifted.
// This is a backward-compatibility guard: changing these would break existing code
// that stores or compares PageSize values numerically.
func TestPageSize_IotaOrder_OriginalEightUnchanged(t *testing.T) {
	require.Equal(t, PageSize(0), A4, "A4 must be iota 0")
	require.Equal(t, PageSize(1), A3, "A3 must be iota 1")
	require.Equal(t, PageSize(2), A5, "A5 must be iota 2")
	require.Equal(t, PageSize(3), B4, "B4 must be iota 3")
	require.Equal(t, PageSize(4), B5, "B5 must be iota 4")
	require.Equal(t, PageSize(5), Letter, "Letter must be iota 5")
	require.Equal(t, PageSize(6), Legal, "Legal must be iota 6")
	require.Equal(t, PageSize(7), Tabloid, "Tabloid must be iota 7")
}

// TestPageSize_PresentationSizes verifies the unique Slide sizes that distinguish
// gxpdf from other Go PDF libraries.
func TestPageSize_PresentationSizes(t *testing.T) {
	t.Run("Slide16x9 is wider than tall (landscape)", func(t *testing.T) {
		rect := Slide16x9.ToRectangle()
		assert.Greater(t, rect.Width(), rect.Height(), "16:9 slide must be landscape (width > height)")
		// Aspect ratio: 960/540 = 16/9
		ratio := rect.Width() / rect.Height()
		assert.InDelta(t, 16.0/9.0, ratio, 0.001, "16:9 aspect ratio")
	})

	t.Run("Slide4x3 is wider than tall (landscape)", func(t *testing.T) {
		rect := Slide4x3.ToRectangle()
		assert.Greater(t, rect.Width(), rect.Height(), "4:3 slide must be landscape (width > height)")
		// Aspect ratio: 720/540 = 4/3
		ratio := rect.Width() / rect.Height()
		assert.InDelta(t, 4.0/3.0, ratio, 0.001, "4:3 aspect ratio")
	})
}

// TestPageSize_JISvsISO_BSeriesDifferent verifies that JIS B sizes are distinct
// from ISO B sizes (a common source of confusion).
func TestPageSize_JISvsISO_BSeriesDifferent(t *testing.T) {
	isob4 := B4.ToRectangle()
	jisb4 := JISB4.ToRectangle()

	assert.NotEqual(t, isob4.Width(), jisb4.Width(), "ISO B4 and JIS B4 widths must differ")
	assert.NotEqual(t, isob4.Height(), jisb4.Height(), "ISO B4 and JIS B4 heights must differ")

	isob5 := B5.ToRectangle()
	jisb5 := JISB5.ToRectangle()

	assert.NotEqual(t, isob5.Width(), jisb5.Width(), "ISO B5 and JIS B5 widths must differ")
	assert.NotEqual(t, isob5.Height(), jisb5.Height(), "ISO B5 and JIS B5 heights must differ")
}

// TestPageSize_ASeriesHalving verifies the mathematical relationship between
// consecutive A-series sizes (each is half the area of the previous).
func TestPageSize_ASeriesHalving(t *testing.T) {
	pairs := []struct{ larger, smaller PageSize }{
		{A3, A4},
		{A4, A5},
		{A5, A6},
		{A6, A7},
		{A7, A8},
	}

	for _, pair := range pairs {
		larger := pair.larger.ToRectangle()
		smaller := pair.smaller.ToRectangle()

		largerArea := larger.Width() * larger.Height()
		smallerArea := smaller.Width() * smaller.Height()

		// Smaller is approximately half the area of larger (ISO 216 definition).
		ratio := largerArea / smallerArea
		assert.InDelta(t, 2.0, ratio, 0.05,
			"%s area should be ~2x %s area", pair.larger.String(), pair.smaller.String())
	}
}

// TestCustomPageSize verifies that custom dimensions pass through correctly.
func TestCustomPageSize(t *testing.T) {
	tests := []struct {
		name       string
		width      float64
		height     float64
		wantWidth  float64
		wantHeight float64
	}{
		{
			name:       "6x9 inches in points",
			width:      432.0, // 6 * 72
			height:     648.0, // 9 * 72
			wantWidth:  432.0,
			wantHeight: 648.0,
		},
		{
			name:       "custom square",
			width:      500.0,
			height:     500.0,
			wantWidth:  500.0,
			wantHeight: 500.0,
		},
		{
			name:       "landscape A4 (swapped)",
			width:      842.0,
			height:     595.0,
			wantWidth:  842.0,
			wantHeight: 595.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rect := CustomPageSize(tt.width, tt.height)

			llx, lly := rect.LowerLeft()
			assert.Equal(t, 0.0, llx)
			assert.Equal(t, 0.0, lly)
			assert.Equal(t, tt.wantWidth, rect.Width())
			assert.Equal(t, tt.wantHeight, rect.Height())
		})
	}
}

// TestConversionFunctions verifies all unit conversion helpers.
func TestConversionFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
		convert  func(float64) float64
	}{
		{
			name:     "1 inch to points",
			input:    1.0,
			expected: 72.0,
			convert:  InchesToPoints,
		},
		{
			name:     "8.5 inches to points (Letter width)",
			input:    8.5,
			expected: 612.0,
			convert:  InchesToPoints,
		},
		{
			name:     "72 points to inches",
			input:    72.0,
			expected: 1.0,
			convert:  PointsToInches,
		},
		{
			name:     "25.4 mm to points (= 1 inch)",
			input:    25.4,
			expected: 72.0,
			convert:  MMToPoints,
		},
		{
			name:     "210 mm to points (A4 width)",
			input:    210.0,
			expected: 595.27559055118, // 210 * 72/25.4
			convert:  MMToPoints,
		},
		{
			name:     "595 points to mm (approximately 210)",
			input:    595.0,
			expected: 209.90277777778, // 595 * 25.4/72
			convert:  PointsToMM,
		},
		{
			name:     "1 cm to points",
			input:    1.0,
			expected: 28.346456692913, // 72/2.54
			convert:  CMToPoints,
		},
		{
			name:     "72 points to cm",
			input:    72.0,
			expected: 2.54,
			convert:  PointsToCM,
		},
		{
			name:     "2.54 cm to points (= 1 inch)",
			input:    2.54,
			expected: 72.0,
			convert:  CMToPoints,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.convert(tt.input)
			assert.InDelta(t, tt.expected, result, 0.0001, "conversion mismatch for %s", tt.name)
		})
	}
}

// TestConversionConstants verifies the unit conversion constants.
func TestConversionConstants(t *testing.T) {
	assert.Equal(t, 72.0, PointsPerInch, "1 inch = 72 points")
	assert.InDelta(t, 2.83465, PointsPerMM, 0.00001, "1 mm ≈ 2.83465 points")
	assert.InDelta(t, 28.3465, PointsPerCM, 0.0001, "1 cm ≈ 28.3465 points")
}

// TestRealWorldSizes verifies that common sizes match their real-world specifications.
func TestRealWorldSizes(t *testing.T) {
	t.Run("A4 dimensions match standard (210x297mm)", func(t *testing.T) {
		widthPt := MMToPoints(210.0)
		heightPt := MMToPoints(297.0)
		assert.InDelta(t, 595.0, widthPt, 1.0, "A4 width ≈ 595pt")
		assert.InDelta(t, 842.0, heightPt, 1.0, "A4 height ≈ 842pt")
	})

	t.Run("Letter dimensions match standard (8.5x11in)", func(t *testing.T) {
		widthPt := InchesToPoints(8.5)
		heightPt := InchesToPoints(11.0)
		assert.Equal(t, 612.0, widthPt, "Letter width = 612pt")
		assert.Equal(t, 792.0, heightPt, "Letter height = 792pt")
	})

	t.Run("ANSIE is larger than ANSID which is larger than ANSIC", func(t *testing.T) {
		c := ANSIC.ToRectangle()
		d := ANSID.ToRectangle()
		e := ANSIE.ToRectangle()

		assert.Less(t, c.Width()*c.Height(), d.Width()*d.Height(), "ANSIC area < ANSID area")
		assert.Less(t, d.Width()*d.Height(), e.Width()*e.Height(), "ANSID area < ANSIE area")
	})
}

// Benchmark tests

func BenchmarkPageSize_ToRectangle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = A4.ToRectangle()
	}
}

func BenchmarkPageSize_ToRectangle_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = A0.ToRectangle()
	}
}

func BenchmarkCustomPageSize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CustomPageSize(595, 842)
	}
}

func BenchmarkInchesToPoints(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = InchesToPoints(8.5)
	}
}

func BenchmarkMMToPoints(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MMToPoints(210.0)
	}
}
