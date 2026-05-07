package gxpdf

// TextElement represents a single text run extracted from a PDF page,
// with full positional and font metadata.
//
// Coordinates follow the PDF coordinate system (ISO 32000-1, §8.3.2):
//   - Origin (0, 0) is at the bottom-left corner of the page.
//   - X increases to the right, Y increases upward.
//   - All measurements are in points (1 pt = 1/72 inch).
//
// Use Page.ExtractTextElements or Document.ExtractTextElementsFromPage to
// obtain slices of TextElement.
type TextElement struct {
	// Text is the actual string content of this run.
	Text string

	// X is the left edge of the text run in points.
	X float64

	// Y is the bottom edge of the text run in points.
	Y float64

	// Width is the horizontal extent of the text run in points.
	Width float64

	// Height is the vertical extent of the text run in points.
	Height float64

	// FontName is the PDF internal font resource name (e.g. "/F1", "/Helvetica").
	FontName string

	// FontSize is the rendered font size in points.
	FontSize float64
}
