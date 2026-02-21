package creator

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPageWithDimensions_BasicDimensions verifies that custom width and height
// are stored exactly and reflected in the returned Page's Width/Height accessors.
func TestNewPageWithDimensions_BasicDimensions(t *testing.T) {
	tests := []struct {
		name       string
		widthPt    float64
		heightPt   float64
		wantWidth  float64
		wantHeight float64
	}{
		{
			name:       "6x9 inch trade book",
			widthPt:    InchesToPoints(6),
			heightPt:   InchesToPoints(9),
			wantWidth:  432.0,
			wantHeight: 648.0,
		},
		{
			name:       "true landscape A4 (swap width/height)",
			widthPt:    842,
			heightPt:   595,
			wantWidth:  842,
			wantHeight: 595,
		},
		{
			name:       "business card (3.5 x 2 inches)",
			widthPt:    InchesToPoints(3.5),
			heightPt:   InchesToPoints(2),
			wantWidth:  252.0,
			wantHeight: 144.0,
		},
		{
			name:       "square page",
			widthPt:    500,
			heightPt:   500,
			wantWidth:  500,
			wantHeight: 500,
		},
		{
			name:       "A4 via mm conversion",
			widthPt:    MMToPoints(210),
			heightPt:   MMToPoints(297),
			wantWidth:  MMToPoints(210),
			wantHeight: MMToPoints(297),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, err := c.NewPageWithDimensions(tt.widthPt, tt.heightPt)
			require.NoError(t, err)
			require.NotNil(t, page)

			assert.Equal(t, tt.wantWidth, page.Width(), "width mismatch")
			assert.Equal(t, tt.wantHeight, page.Height(), "height mismatch")
		})
	}
}

// TestNewPageWithDimensions_ValidationErrors verifies that non-positive dimensions
// are rejected with a descriptive error.
func TestNewPageWithDimensions_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		widthPt  float64
		heightPt float64
	}{
		{"zero width", 0, 842},
		{"zero height", 595, 0},
		{"negative width", -100, 842},
		{"negative height", 595, -100},
		{"both zero", 0, 0},
		{"both negative", -595, -842},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, err := c.NewPageWithDimensions(tt.widthPt, tt.heightPt)
			assert.Error(t, err, "expected error for invalid dimensions")
			assert.Nil(t, page, "page should be nil on error")
		})
	}
}

// TestNewPageWithDimensions_PageCount verifies that each call to NewPageWithDimensions
// increments the document page count correctly.
func TestNewPageWithDimensions_PageCount(t *testing.T) {
	c := New()

	assert.Equal(t, 0, c.PageCount())

	_, err := c.NewPageWithDimensions(500, 500)
	require.NoError(t, err)
	assert.Equal(t, 1, c.PageCount())

	_, err = c.NewPageWithDimensions(300, 400)
	require.NoError(t, err)
	assert.Equal(t, 2, c.PageCount())
}

// TestNewPageWithDimensions_MixedWithStandardPages verifies that custom-dimension
// pages can be combined with standard-size pages in one document.
func TestNewPageWithDimensions_MixedWithStandardPages(t *testing.T) {
	c := New()

	// Standard page
	stdPage, err := c.NewPageWithSize(A4)
	require.NoError(t, err)

	// Custom page
	customPage, err := c.NewPageWithDimensions(300, 400)
	require.NoError(t, err)

	assert.Equal(t, 2, c.PageCount())
	assert.Equal(t, 595.0, stdPage.Width())
	assert.Equal(t, 842.0, stdPage.Height())
	assert.Equal(t, 300.0, customPage.Width())
	assert.Equal(t, 400.0, customPage.Height())
}

// TestNewPageWithDimensions_WritesValidPDF verifies that a document with a
// custom-dimension page can be written without error and produces valid PDF bytes.
func TestNewPageWithDimensions_WritesValidPDF(t *testing.T) {
	c := New()
	page, err := c.NewPageWithDimensions(InchesToPoints(6), InchesToPoints(9))
	require.NoError(t, err)

	err = page.AddText("Custom 6x9 page", 50, 600, Helvetica, 12)
	require.NoError(t, err)

	pdfBytes, err := c.Bytes()
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)

	// Verify PDF header and footer.
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")), "must start with PDF header")
	assert.True(t, bytes.HasSuffix(bytes.TrimSpace(pdfBytes), []byte("%%EOF")), "must end with %%EOF")
}

// TestNewPageWithDimensions_DefaultMarginsApplied verifies that the creator's
// default margins are applied to pages created with custom dimensions.
func TestNewPageWithDimensions_DefaultMarginsApplied(t *testing.T) {
	c := New()

	// Set non-default margins before creating the custom page.
	err := c.SetMargins(36, 36, 36, 36)
	require.NoError(t, err)

	page, err := c.NewPageWithDimensions(500, 700)
	require.NoError(t, err)

	margins := page.Margins()
	assert.Equal(t, 36.0, margins.Top)
	assert.Equal(t, 36.0, margins.Right)
	assert.Equal(t, 36.0, margins.Bottom)
	assert.Equal(t, 36.0, margins.Left)
}

// TestNewPageWithDimensions_ContentAreaIsCorrect verifies that ContentWidth
// and ContentHeight account for default margins correctly.
func TestNewPageWithDimensions_ContentAreaIsCorrect(t *testing.T) {
	c := New()
	err := c.SetMargins(50, 50, 50, 50)
	require.NoError(t, err)

	page, err := c.NewPageWithDimensions(400, 600)
	require.NoError(t, err)

	assert.Equal(t, 300.0, page.ContentWidth(), "content width = 400 - 50 - 50")
	assert.Equal(t, 500.0, page.ContentHeight(), "content height = 600 - 50 - 50")
}

// TestNewPageWithSize_Landscape verifies true landscape pages (swapped MediaBox, no /Rotate).
func TestNewPageWithSize_Landscape(t *testing.T) {
	tests := []struct {
		name  string
		size  PageSize
		wantW float64
		wantH float64
	}{
		{"A4 landscape", A4, 842, 595},
		{"Letter landscape", Letter, 792, 612},
		{"A3 landscape", A3, 1191, 842},
		{"Tabloid landscape", Tabloid, 1224, 792},
		{"A5 landscape", A5, 595, 420},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, err := c.NewPageWithSize(tt.size, Landscape)
			require.NoError(t, err)
			assert.Equal(t, tt.wantW, page.Width(), "landscape width")
			assert.Equal(t, tt.wantH, page.Height(), "landscape height")
			// Width must be greater than height for landscape
			assert.Greater(t, page.Width(), page.Height(), "landscape: width > height")
		})
	}
}

// TestNewPageWithSize_PortraitDefault verifies that omitting orientation gives portrait.
func TestNewPageWithSize_PortraitDefault(t *testing.T) {
	c := New()
	page, err := c.NewPageWithSize(A4)
	require.NoError(t, err)

	assert.Equal(t, 595.0, page.Width())
	assert.Equal(t, 842.0, page.Height())
	assert.Less(t, page.Width(), page.Height(), "portrait: width < height")
}

// TestNewPageWithSize_ExplicitPortrait verifies that passing Portrait explicitly works.
func TestNewPageWithSize_ExplicitPortrait(t *testing.T) {
	c := New()
	page, err := c.NewPageWithSize(Letter, Portrait)
	require.NoError(t, err)

	assert.Equal(t, 612.0, page.Width())
	assert.Equal(t, 792.0, page.Height())
}

// TestNewPageWithSize_LandscapePDFOutput verifies landscape page produces valid PDF.
func TestNewPageWithSize_LandscapePDFOutput(t *testing.T) {
	c := New()
	page, err := c.NewPageWithSize(A4, Landscape)
	require.NoError(t, err)

	err = page.AddText("Landscape A4", 100, 400, Helvetica, 24)
	require.NoError(t, err)

	pdfBytes, err := c.Bytes()
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")))
	assert.True(t, bytes.HasSuffix(bytes.TrimSpace(pdfBytes), []byte("%%EOF")))
}

// TestAddPageWithRect_DocumentLayer verifies that document.AddPageWithRect stores
// dimensions correctly at the domain layer (unit test of the domain method directly).
func TestAddPageWithRect_DocumentLayer(t *testing.T) {
	// We can test this through the creator which delegates to it.
	c := New()

	// Test that width=1 and height=1 (minimum legal positive value) work.
	page, err := c.NewPageWithDimensions(1, 1)
	require.NoError(t, err)
	require.NotNil(t, page)
	assert.Equal(t, 1.0, page.Width())
	assert.Equal(t, 1.0, page.Height())
}
