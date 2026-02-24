package creator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	c := New()

	assert.NotNil(t, c)
	assert.NotNil(t, c.doc)
	assert.Equal(t, 0, c.PageCount())

	// Check defaults
	assert.Equal(t, 72.0, c.defaultMargins.Top)
	assert.Equal(t, 72.0, c.defaultMargins.Right)
	assert.Equal(t, 72.0, c.defaultMargins.Bottom)
	assert.Equal(t, 72.0, c.defaultMargins.Left)
}

func TestCreator_SetMetadata(t *testing.T) {
	c := New()

	c.SetTitle("Test Document")
	c.SetAuthor("John Doe")
	c.SetSubject("Test Subject")

	doc := c.Document()
	assert.Equal(t, "Test Document", doc.Title())
	assert.Equal(t, "John Doe", doc.Author())
	assert.Equal(t, "Test Subject", doc.Subject())
}

func TestCreator_SetMetadata_AllAtOnce(t *testing.T) {
	c := New()

	c.SetMetadata("Title", "Author", "Subject")

	doc := c.Document()
	assert.Equal(t, "Title", doc.Title())
	assert.Equal(t, "Author", doc.Author())
	assert.Equal(t, "Subject", doc.Subject())
}

func TestCreator_SetKeywords(t *testing.T) {
	c := New()

	c.SetKeywords("pdf", "golang", "library")

	doc := c.Document()
	keywords := doc.Keywords()
	assert.Equal(t, 3, len(keywords))
	assert.Contains(t, keywords, "pdf")
	assert.Contains(t, keywords, "golang")
	assert.Contains(t, keywords, "library")
}

func TestCreator_NewPage(t *testing.T) {
	c := New()

	page, err := c.NewPage()
	require.NoError(t, err)
	assert.NotNil(t, page)
	assert.Equal(t, 1, c.PageCount())

	// Check page dimensions (A4 default)
	assert.Equal(t, 595.0, page.Width())
	assert.Equal(t, 842.0, page.Height())
}

func TestCreator_NewPageWithSize(t *testing.T) {
	c := New()

	page, err := c.NewPageWithSize(Letter)
	require.NoError(t, err)
	assert.NotNil(t, page)

	// Check page dimensions (Letter: 612 × 792)
	assert.Equal(t, 612.0, page.Width())
	assert.Equal(t, 792.0, page.Height())
}

func TestCreator_SetPageSize(t *testing.T) {
	c := New()
	c.SetPageSize(Letter)

	page, err := c.NewPage()
	require.NoError(t, err)

	// Should use Letter size (set as default)
	assert.Equal(t, 612.0, page.Width())
	assert.Equal(t, 792.0, page.Height())
}

func TestCreator_SetMargins(t *testing.T) {
	c := New()

	err := c.SetMargins(36, 36, 36, 36)
	require.NoError(t, err)

	page, err := c.NewPage()
	require.NoError(t, err)

	margins := page.Margins()
	assert.Equal(t, 36.0, margins.Top)
	assert.Equal(t, 36.0, margins.Right)
	assert.Equal(t, 36.0, margins.Bottom)
	assert.Equal(t, 36.0, margins.Left)
}

func TestCreator_SetMargins_Negative(t *testing.T) {
	c := New()

	err := c.SetMargins(-10, 0, 0, 0)
	assert.ErrorIs(t, err, ErrInvalidMargins)

	err = c.SetMargins(0, -10, 0, 0)
	assert.ErrorIs(t, err, ErrInvalidMargins)

	err = c.SetMargins(0, 0, -10, 0)
	assert.ErrorIs(t, err, ErrInvalidMargins)

	err = c.SetMargins(0, 0, 0, -10)
	assert.ErrorIs(t, err, ErrInvalidMargins)
}

func TestCreator_Validate_EmptyDocument(t *testing.T) {
	c := New()

	err := c.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pages")
}

func TestCreator_Validate_ValidDocument(t *testing.T) {
	c := New()

	_, err := c.NewPage()
	require.NoError(t, err)

	err = c.Validate()
	assert.NoError(t, err)
}

func TestCreator_PageCount(t *testing.T) {
	c := New()

	assert.Equal(t, 0, c.PageCount())

	_, _ = c.NewPage()
	assert.Equal(t, 1, c.PageCount())

	_, _ = c.NewPage()
	assert.Equal(t, 2, c.PageCount())

	_, _ = c.NewPage()
	assert.Equal(t, 3, c.PageCount())
}

func TestCreator_MultiplePages(t *testing.T) {
	c := New()

	// Add multiple pages with different sizes
	page1, err := c.NewPage() // A4 (default)
	require.NoError(t, err)
	assert.Equal(t, 595.0, page1.Width())

	c.SetPageSize(Letter)
	page2, err := c.NewPage() // Letter
	require.NoError(t, err)
	assert.Equal(t, 612.0, page2.Width())

	page3, err := c.NewPageWithSize(Legal) // Legal
	require.NoError(t, err)
	assert.Equal(t, 612.0, page3.Width())
	assert.Equal(t, 1008.0, page3.Height())

	assert.Equal(t, 3, c.PageCount())
}

func TestCreator_MetadataWrittenToPDF(t *testing.T) {
	c := New()
	c.SetTitle("Enterprise Report")
	c.SetAuthor("Jane Smith")
	c.SetSubject("Annual Analysis")

	_, err := c.NewPage()
	require.NoError(t, err)

	pdfBytes, err := c.Bytes()
	require.NoError(t, err)

	pdfStr := string(pdfBytes)

	// Info dictionary must contain metadata as proper PDF objects
	assert.True(t, strings.Contains(pdfStr, "/Title (Enterprise Report)"),
		"PDF output must contain /Title in Info dictionary")
	assert.True(t, strings.Contains(pdfStr, "/Author (Jane Smith)"),
		"PDF output must contain /Author in Info dictionary")
	assert.True(t, strings.Contains(pdfStr, "/Subject (Annual Analysis)"),
		"PDF output must contain /Subject in Info dictionary")

	// Trailer must reference Info with real object number (not 0)
	assert.True(t, strings.Contains(pdfStr, "/Info "),
		"trailer must contain /Info reference")
	assert.False(t, strings.Contains(pdfStr, "/Info 0 0 R"),
		"/Info must not reference object 0")
}

func TestGradientFill_LinearRect(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	// Draw a rectangle with a linear gradient fill.
	grad := NewLinearGradient(50, 700, 250, 700)
	err = grad.AddColorStop(0, Red)
	require.NoError(t, err)
	err = grad.AddColorStop(1, Blue)
	require.NoError(t, err)

	err = page.DrawRect(50, 700, 200, 60, &RectOptions{
		FillGradient: grad,
		StrokeColor:  &Black,
		StrokeWidth:  1.0,
	})
	require.NoError(t, err)

	// Generate PDF bytes.
	pdfBytes, err := c.Bytes()
	require.NoError(t, err)

	pdfStr := string(pdfBytes)

	// Must contain shading resource.
	assert.True(t, strings.Contains(pdfStr, "/Shading <<"),
		"PDF must contain /Shading resource dictionary")
	assert.True(t, strings.Contains(pdfStr, "/Sh1 "),
		"PDF must contain /Sh1 shading reference")

	// Must contain shading dictionary.
	assert.True(t, strings.Contains(pdfStr, "/ShadingType 2"),
		"PDF must contain /ShadingType 2 (axial)")
	assert.True(t, strings.Contains(pdfStr, "/ColorSpace /DeviceRGB"),
		"PDF must contain /ColorSpace /DeviceRGB")

	// Must contain function objects.
	assert.True(t, strings.Contains(pdfStr, "/FunctionType 2"),
		"PDF must contain /FunctionType 2 (exponential)")

	// Must NOT contain "0 0 R" for shading references.
	// The /Sh1 reference should have a real object number.
	assert.False(t, strings.Contains(pdfStr, "/Sh1 0 0 R"),
		"/Sh1 must not reference placeholder object 0")
}

func TestGradientFill_RadialCircle(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	// Draw a circle with a radial gradient fill.
	grad := NewRadialGradient(200, 500, 0, 200, 500, 50)
	err = grad.AddColorStop(0, White)
	require.NoError(t, err)
	err = grad.AddColorStop(1, Blue)
	require.NoError(t, err)

	err = page.DrawCircle(200, 500, 50, &CircleOptions{
		FillGradient: grad,
	})
	require.NoError(t, err)

	pdfBytes, err := c.Bytes()
	require.NoError(t, err)

	pdfStr := string(pdfBytes)
	assert.True(t, strings.Contains(pdfStr, "/ShadingType 3"),
		"PDF must contain /ShadingType 3 (radial)")
}

func TestGradientFill_MultiStop(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	// Draw a rect with a 3-stop gradient.
	grad := NewLinearGradient(50, 400, 250, 400)
	err = grad.AddColorStop(0, Red)
	require.NoError(t, err)
	err = grad.AddColorStop(0.5, Yellow)
	require.NoError(t, err)
	err = grad.AddColorStop(1, Green)
	require.NoError(t, err)

	err = page.DrawRect(50, 400, 200, 60, &RectOptions{
		FillGradient: grad,
	})
	require.NoError(t, err)

	pdfBytes, err := c.Bytes()
	require.NoError(t, err)

	pdfStr := string(pdfBytes)

	// Multi-stop should produce a Type 3 stitching function.
	assert.True(t, strings.Contains(pdfStr, "/FunctionType 3"),
		"PDF must contain /FunctionType 3 (stitching) for multi-stop gradient")
	assert.True(t, strings.Contains(pdfStr, "/Bounds"),
		"PDF must contain /Bounds in stitching function")
}
