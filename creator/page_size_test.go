package creator

import (
	"testing"

	"github.com/coregx/gxpdf/internal/document"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPageSize_IotaMirrorsDomain is a compile-time and runtime guard ensuring
// that creator.PageSize iota values map 1:1 to document.PageSize iota values.
// If the domain enum is ever reordered, this test will catch the mismatch before
// any PDF ends up with the wrong dimensions.
func TestPageSize_IotaMirrorsDomain(t *testing.T) {
	tests := []struct {
		creatorSize PageSize
		domainSize  document.PageSize
		name        string
	}{
		{A4, document.A4, "A4"},
		{A3, document.A3, "A3"},
		{A5, document.A5, "A5"},
		{B4, document.B4, "B4"},
		{B5, document.B5, "B5"},
		{Letter, document.Letter, "Letter"},
		{Legal, document.Legal, "Legal"},
		{Tabloid, document.Tabloid, "Tabloid"},
		{A0, document.A0, "A0"},
		{A1, document.A1, "A1"},
		{A2, document.A2, "A2"},
		{A6, document.A6, "A6"},
		{A7, document.A7, "A7"},
		{A8, document.A8, "A8"},
		{B0, document.B0, "B0"},
		{B1, document.B1, "B1"},
		{B2, document.B2, "B2"},
		{B3, document.B3, "B3"},
		{B6, document.B6, "B6"},
		{C4, document.C4, "C4"},
		{C5, document.C5, "C5"},
		{C6, document.C6, "C6"},
		{DL, document.DL, "DL"},
		{Executive, document.Executive, "Executive"},
		{HalfLetter, document.HalfLetter, "HalfLetter"},
		{ANSIC, document.ANSIC, "ANSIC"},
		{ANSID, document.ANSID, "ANSID"},
		{ANSIE, document.ANSIE, "ANSIE"},
		{Photo4x6, document.Photo4x6, "Photo4x6"},
		{Photo5x7, document.Photo5x7, "Photo5x7"},
		{Photo8x10, document.Photo8x10, "Photo8x10"},
		{Digest, document.Digest, "Digest"},
		{USTradeBook, document.USTradeBook, "USTradeBook"},
		{Slide16x9, document.Slide16x9, "Slide16x9"},
		{Slide4x3, document.Slide4x3, "Slide4x3"},
		{Envelope10, document.Envelope10, "Envelope10"},
		{JISB4, document.JISB4, "JISB4"},
		{JISB5, document.JISB5, "JISB5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The integer value must be identical.
			assert.Equal(t, int(tt.domainSize), int(tt.creatorSize),
				"creator.%s iota must equal document.%s iota", tt.name, tt.name)

			// The converted domain size must produce the same string.
			assert.Equal(t, tt.domainSize.String(), tt.creatorSize.String(),
				"String() must match for %s", tt.name)

			// The toDomainSize conversion must return the correct domain value.
			assert.Equal(t, tt.domainSize, tt.creatorSize.toDomainSize(),
				"toDomainSize() must return document.%s", tt.name)
		})
	}
}

// TestPageSize_String_AllSizes verifies human-readable names at the creator layer.
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
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.pageSize.String())
		})
	}
}

// TestNewPageWithSize_AllSizes verifies that Creator.NewPageWithSize works for
// every named size and produces a page with the correct dimensions.
func TestNewPageWithSize_AllSizes(t *testing.T) {
	sizes := []struct {
		size       PageSize
		wantWidth  float64
		wantHeight float64
	}{
		{A4, 595, 842},
		{A3, 842, 1191},
		{A5, 420, 595},
		{Letter, 612, 792},
		{Legal, 612, 1008},
		{Tabloid, 792, 1224},
		{A0, 2384, 3370},
		{Slide16x9, 960, 540},
		{Slide4x3, 720, 540},
		{JISB4, 729, 1032},
		{JISB5, 516, 729},
	}

	for _, tt := range sizes {
		t.Run(tt.size.String(), func(t *testing.T) {
			c := New()
			page, err := c.NewPageWithSize(tt.size)
			require.NoError(t, err)
			require.NotNil(t, page)

			assert.Equal(t, tt.wantWidth, page.Width(), "width mismatch for %s", tt.size)
			assert.Equal(t, tt.wantHeight, page.Height(), "height mismatch for %s", tt.size)
		})
	}
}

// TestCreatorPageSize_UnitConversionHelpers verifies the creator-level
// unit conversion constants and functions delegate correctly to the domain layer.
func TestCreatorPageSize_UnitConversionHelpers(t *testing.T) {
	t.Run("PointsPerInch", func(t *testing.T) {
		assert.Equal(t, 72.0, PointsPerInch)
	})

	t.Run("InchesToPoints(1) == 72", func(t *testing.T) {
		assert.Equal(t, 72.0, InchesToPoints(1.0))
	})

	t.Run("InchesToPoints(8.5) == 612 (Letter width)", func(t *testing.T) {
		assert.Equal(t, 612.0, InchesToPoints(8.5))
	})

	t.Run("MMToPoints(25.4) == 72 (1 inch)", func(t *testing.T) {
		assert.InDelta(t, 72.0, MMToPoints(25.4), 0.001)
	})

	t.Run("CMToPoints(2.54) == 72 (1 inch)", func(t *testing.T) {
		assert.InDelta(t, 72.0, CMToPoints(2.54), 0.001)
	})

	t.Run("PointsPerMM constant is correct", func(t *testing.T) {
		assert.InDelta(t, 72.0/25.4, PointsPerMM, 0.00001)
	})

	t.Run("PointsPerCM constant is correct", func(t *testing.T) {
		assert.InDelta(t, 72.0/2.54, PointsPerCM, 0.0001)
	})
}
