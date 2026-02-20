package document

import "github.com/coregx/gxpdf/internal/models/types"

// PageSize represents standard PDF page sizes.
//
// Standard sizes are provided as constants for convenience.
// For custom sizes, use CustomPageSize().
type PageSize int

const (
	// ISO 216 A series (most common international sizes)

	// A4 is 210 × 297 mm (8.27 × 11.69 in) - Most common international paper size.
	A4 PageSize = iota

	// A3 is 297 × 420 mm (11.69 × 16.54 in) - Twice the area of A4.
	A3

	// A5 is 148 × 210 mm (5.83 × 8.27 in) - Half the area of A4.
	A5

	// ISO 216 B series

	// B4 is 250 × 353 mm (9.84 × 13.90 in) - Between A3 and A4.
	B4

	// B5 is 176 × 250 mm (6.93 × 9.84 in) - Between A4 and A5.
	B5

	// North American sizes

	// Letter is 8.5 × 11 in (215.9 × 279.4 mm) - Standard US/Canada paper size.
	Letter

	// Legal is 8.5 × 14 in (215.9 × 355.6 mm) - US legal documents.
	Legal

	// Tabloid is 11 × 17 in (279.4 × 431.8 mm) - Also known as Ledger.
	Tabloid

	// Extended ISO A series

	// A0 is 841 × 1189 mm (2384 × 3370 pt) - Largest common A-series size.
	A0

	// A1 is 594 × 841 mm (1684 × 2384 pt).
	A1

	// A2 is 420 × 594 mm (1191 × 1684 pt).
	A2

	// A6 is 105 × 148 mm (298 × 420 pt) - Postcard size.
	A6

	// A7 is 74 × 105 mm (210 × 298 pt).
	A7

	// A8 is 52 × 74 mm (147 × 210 pt).
	A8

	// Extended ISO B series

	// B0 is 1000 × 1414 mm (2835 × 4008 pt).
	B0

	// B1 is 707 × 1000 mm (2004 × 2835 pt).
	B1

	// B2 is 500 × 707 mm (1417 × 2004 pt).
	B2

	// B3 is 353 × 500 mm (1001 × 1417 pt).
	B3

	// B6 is 125 × 176 mm (354 × 499 pt).
	B6

	// ISO C / Envelope sizes

	// C4 is 229 × 324 mm (649 × 918 pt) - A4 envelope.
	C4

	// C5 is 162 × 229 mm (459 × 649 pt) - A5 envelope.
	C5

	// C6 is 114 × 162 mm (323 × 459 pt) - A6 envelope.
	C6

	// DL is 110 × 220 mm (312 × 624 pt) - Long envelope (fits A4 folded in thirds).
	DL

	// North American extended sizes

	// Executive is 7.25 × 10.5 in (522 × 756 pt) - US executive paper.
	Executive

	// HalfLetter is 5.5 × 8.5 in (396 × 612 pt) - Half US letter.
	HalfLetter

	// ANSI Engineering sizes

	// ANSIC is 17 × 22 in (1224 × 1584 pt) - ANSI C engineering drawing.
	ANSIC

	// ANSID is 22 × 34 in (1584 × 2448 pt) - ANSI D engineering drawing.
	ANSID

	// ANSIE is 34 × 44 in (2448 × 3168 pt) - ANSI E engineering drawing.
	ANSIE

	// Photo sizes

	// Photo4x6 is 4 × 6 in (288 × 432 pt) - Standard 4×6 photo print.
	Photo4x6

	// Photo5x7 is 5 × 7 in (360 × 504 pt) - Standard 5×7 photo print.
	Photo5x7

	// Photo8x10 is 8 × 10 in (576 × 720 pt) - Standard 8×10 photo print.
	Photo8x10

	// Book Publishing sizes

	// Digest is 5.5 × 8.5 in (396 × 612 pt) - Digest magazine size.
	Digest

	// USTradeBook is 6 × 9 in (432 × 648 pt) - Standard US trade book.
	USTradeBook

	// Presentation sizes (unique — not found in most PDF libraries)

	// Slide16x9 is a widescreen presentation slide at 16:9 aspect ratio (960 × 540 pt).
	// Equivalent to 13.33 × 7.5 inches — matches PowerPoint/Keynote widescreen default.
	Slide16x9

	// Slide4x3 is a standard presentation slide at 4:3 aspect ratio (720 × 540 pt).
	// Equivalent to 10 × 7.5 inches — matches the traditional slide format.
	Slide4x3

	// US Envelope sizes

	// Envelope10 is a #10 commercial envelope (4.125 × 9.5 in, 297 × 684 pt).
	Envelope10

	// JIS B series (Japanese Industrial Standard — different from ISO B!)

	// JISB4 is 257 × 364 mm (729 × 1032 pt) - Japanese B4.
	JISB4

	// JISB5 is 182 × 257 mm (516 × 729 pt) - Japanese B5.
	JISB5

	// Custom indicates a custom page size (use CustomPageSize function).
	Custom
)

// pageSizeData holds the display name and point dimensions for a page size.
type pageSizeData struct {
	name          string
	width, height float64
}

// pageSizeMap is the single source of truth for all standard page size data.
// Using a map eliminates duplicated switch statements across the codebase.
var pageSizeMap = map[PageSize]pageSizeData{
	// ISO A series (original 8)
	A4:      {"A4", 595, 842},
	A3:      {"A3", 842, 1191},
	A5:      {"A5", 420, 595},
	B4:      {"B4", 709, 1001},
	B5:      {"B5", 499, 709},
	Letter:  {"Letter", 612, 792},
	Legal:   {"Legal", 612, 1008},
	Tabloid: {"Tabloid", 792, 1224},

	// Extended ISO A series
	A0: {"A0", 2384, 3370},
	A1: {"A1", 1684, 2384},
	A2: {"A2", 1191, 1684},
	A6: {"A6", 298, 420},
	A7: {"A7", 210, 298},
	A8: {"A8", 147, 210},

	// Extended ISO B series
	B0: {"B0", 2835, 4008},
	B1: {"B1", 2004, 2835},
	B2: {"B2", 1417, 2004},
	B3: {"B3", 1001, 1417},
	B6: {"B6", 354, 499},

	// ISO C / Envelope sizes
	C4: {"C4", 649, 918},
	C5: {"C5", 459, 649},
	C6: {"C6", 323, 459},
	DL: {"DL", 312, 624},

	// North American extended
	Executive:  {"Executive", 522, 756},
	HalfLetter: {"HalfLetter", 396, 612},

	// ANSI Engineering
	ANSIC: {"ANSIC", 1224, 1584},
	ANSID: {"ANSID", 1584, 2448},
	ANSIE: {"ANSIE", 2448, 3168},

	// Photo
	Photo4x6:  {"Photo4x6", 288, 432},
	Photo5x7:  {"Photo5x7", 360, 504},
	Photo8x10: {"Photo8x10", 576, 720},

	// Book Publishing
	Digest:      {"Digest", 396, 612},
	USTradeBook: {"USTradeBook", 432, 648},

	// Presentation (unique)
	Slide16x9: {"Slide16x9", 960, 540},
	Slide4x3:  {"Slide4x3", 720, 540},

	// US Envelope
	Envelope10: {"Envelope10", 297, 684},

	// JIS B series
	JISB4: {"JISB4", 729, 1032},
	JISB5: {"JISB5", 516, 729},

	// Custom placeholder
	Custom: {"Custom", 0, 0},
}

// ToRectangle converts PageSize to Rectangle (in points, 1 point = 1/72 inch).
//
// All standard page sizes are returned in portrait orientation.
// Use Page.SetRotation() for landscape orientation.
//
// Example:
//
//	rect := document.A4.ToRectangle()
//	// rect is 595×842 points (210×297mm)
func (ps PageSize) ToRectangle() types.Rectangle {
	if data, ok := pageSizeMap[ps]; ok && ps != Custom {
		return types.MustRectangle(0, 0, data.width, data.height)
	}
	// Default to A4 if unknown or Custom (Custom must use CustomPageSize)
	return types.MustRectangle(0, 0, 595, 842)
}

// String returns the name of the page size.
func (ps PageSize) String() string {
	if data, ok := pageSizeMap[ps]; ok {
		return data.name
	}
	return "Unknown"
}

// CustomPageSize creates a custom page size in points.
//
// Points are 1/72 of an inch.
//
// Example:
//
//	// Create a custom 6×9 inch page
//	customSize := document.CustomPageSize(6*72, 9*72)
func CustomPageSize(widthPt, heightPt float64) types.Rectangle {
	return types.MustRectangle(0, 0, widthPt, heightPt)
}

// Common conversion constants for convenience

const (
	// PointsPerInch is the number of points in one inch.
	// 1 inch = 72 points (PostScript/PDF standard)
	PointsPerInch = 72.0

	// PointsPerMM is the number of points in one millimeter.
	// 1 mm = 72/25.4 ≈ 2.83465 points
	PointsPerMM = 72.0 / 25.4

	// PointsPerCM is the number of points in one centimeter.
	// 1 cm = 72/2.54 ≈ 28.3465 points
	PointsPerCM = 72.0 / 2.54
)

// InchesToPoints converts inches to points.
func InchesToPoints(inches float64) float64 {
	return inches * PointsPerInch
}

// MMToPoints converts millimeters to points.
func MMToPoints(mm float64) float64 {
	return mm * PointsPerMM
}

// CMToPoints converts centimeters to points.
func CMToPoints(cm float64) float64 {
	return cm * PointsPerCM
}

// PointsToInches converts points to inches.
func PointsToInches(points float64) float64 {
	return points / PointsPerInch
}

// PointsToMM converts points to millimeters.
func PointsToMM(points float64) float64 {
	return points / PointsPerMM
}

// PointsToCM converts points to centimeters.
func PointsToCM(points float64) float64 {
	return points / PointsPerCM
}
