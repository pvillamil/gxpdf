package creator

import "github.com/coregx/gxpdf/internal/document"

// PageSize represents standard PDF page sizes.
//
// Common page sizes are provided as constants (A4, Letter, etc.).
// For custom dimensions use NewPageWithDimensions on the Creator.
//
// The iota values mirror document.PageSize exactly so that casting between
// creator.PageSize and document.PageSize is always valid.
type PageSize int

const (
	// ISO 216 A series (most common international sizes)

	// A4 paper size (210 × 297 mm or 595 × 842 points).
	// This is the most common paper size worldwide.
	A4 PageSize = iota

	// A3 paper size (297 × 420 mm or 842 × 1191 points).
	// Twice the size of A4.
	A3

	// A5 paper size (148 × 210 mm or 420 × 595 points).
	// Half the size of A4.
	A5

	// ISO 216 B series

	// B4 paper size (250 × 353 mm or 709 × 1001 points).
	B4

	// B5 paper size (176 × 250 mm or 499 × 709 points).
	B5

	// North American sizes

	// Letter paper size (8.5 × 11 inches or 612 × 792 points).
	// This is the standard size in North America.
	Letter

	// Legal paper size (8.5 × 14 inches or 612 × 1008 points).
	Legal

	// Tabloid paper size (11 × 17 inches or 792 × 1224 points).
	// Also known as Ledger when in landscape orientation.
	Tabloid

	// Extended ISO A series

	// A0 paper size (841 × 1189 mm or 2384 × 3370 points).
	// Largest common A-series size.
	A0

	// A1 paper size (594 × 841 mm or 1684 × 2384 points).
	A1

	// A2 paper size (420 × 594 mm or 1191 × 1684 points).
	A2

	// A6 paper size (105 × 148 mm or 298 × 420 points). Postcard size.
	A6

	// A7 paper size (74 × 105 mm or 210 × 298 points).
	A7

	// A8 paper size (52 × 74 mm or 147 × 210 points).
	A8

	// Extended ISO B series

	// B0 paper size (1000 × 1414 mm or 2835 × 4008 points).
	B0

	// B1 paper size (707 × 1000 mm or 2004 × 2835 points).
	B1

	// B2 paper size (500 × 707 mm or 1417 × 2004 points).
	B2

	// B3 paper size (353 × 500 mm or 1001 × 1417 points).
	B3

	// B6 paper size (125 × 176 mm or 354 × 499 points).
	B6

	// ISO C / Envelope sizes

	// C4 envelope (229 × 324 mm or 649 × 918 points). Fits A4 unfolded.
	C4

	// C5 envelope (162 × 229 mm or 459 × 649 points). Fits A5 unfolded or A4 folded in half.
	C5

	// C6 envelope (114 × 162 mm or 323 × 459 points). Fits A6 unfolded or A4 folded in quarters.
	C6

	// DL envelope (110 × 220 mm or 312 × 624 points). Fits A4 folded in thirds.
	DL

	// North American extended sizes

	// Executive paper size (7.25 × 10.5 inches or 522 × 756 points).
	Executive

	// HalfLetter paper size (5.5 × 8.5 inches or 396 × 612 points).
	HalfLetter

	// ANSI Engineering sizes

	// ANSIC engineering drawing (17 × 22 inches or 1224 × 1584 points).
	ANSIC

	// ANSID engineering drawing (22 × 34 inches or 1584 × 2448 points).
	ANSID

	// ANSIE engineering drawing (34 × 44 inches or 2448 × 3168 points).
	ANSIE

	// Photo sizes

	// Photo4x6 is a standard 4 × 6 inch photo (288 × 432 points).
	Photo4x6

	// Photo5x7 is a standard 5 × 7 inch photo (360 × 504 points).
	Photo5x7

	// Photo8x10 is a standard 8 × 10 inch photo (576 × 720 points).
	Photo8x10

	// Book Publishing sizes

	// Digest is digest magazine size (5.5 × 8.5 inches or 396 × 612 points).
	Digest

	// USTradeBook is standard US trade book size (6 × 9 inches or 432 × 648 points).
	USTradeBook

	// Presentation sizes (unique — not found in most PDF libraries)

	// Slide16x9 is a widescreen presentation slide at 16:9 aspect ratio (960 × 540 points).
	// Equivalent to 13.33 × 7.5 inches — matches PowerPoint/Keynote widescreen default.
	Slide16x9

	// Slide4x3 is a standard presentation slide at 4:3 aspect ratio (720 × 540 points).
	// Equivalent to 10 × 7.5 inches — matches the traditional slide format.
	Slide4x3

	// US Envelope sizes

	// Envelope10 is a #10 commercial envelope (4.125 × 9.5 inches or 297 × 684 points).
	Envelope10

	// JIS B series (Japanese Industrial Standard — different from ISO B!)

	// JISB4 is Japanese B4 (257 × 364 mm or 729 × 1032 points).
	JISB4

	// JISB5 is Japanese B5 (182 × 257 mm or 516 × 729 points).
	JISB5
)

// toDomainSize converts creator PageSize to domain PageSize.
//
// Since both enums share identical iota ordering, a direct cast is safe.
// This is enforced by the compile-time check in the test file.
func (ps PageSize) toDomainSize() document.PageSize {
	return document.PageSize(ps)
}

// String returns the human-readable name of the page size.
func (ps PageSize) String() string {
	return ps.toDomainSize().String()
}

// Unit conversion constants.
//
// Use these with the conversion helper functions to convert between
// real-world units and PDF points before calling NewPageWithDimensions.
const (
	// PointsPerInch is the number of PDF points in one inch.
	// 1 inch = 72 points (PostScript/PDF standard).
	PointsPerInch = document.PointsPerInch

	// PointsPerMM is the number of PDF points in one millimeter.
	// 1 mm = 72/25.4 ≈ 2.83465 points.
	PointsPerMM = document.PointsPerMM

	// PointsPerCM is the number of PDF points in one centimeter.
	// 1 cm = 72/2.54 ≈ 28.3465 points.
	PointsPerCM = document.PointsPerCM
)

// InchesToPoints converts inches to PDF points.
//
// Example:
//
//	width := creator.InchesToPoints(8.5)  // 612 points (Letter width)
func InchesToPoints(inches float64) float64 {
	return document.InchesToPoints(inches)
}

// MMToPoints converts millimeters to PDF points.
//
// Example:
//
//	width := creator.MMToPoints(210)  // ≈ 595 points (A4 width)
func MMToPoints(mm float64) float64 {
	return document.MMToPoints(mm)
}

// CMToPoints converts centimeters to PDF points.
//
// Example:
//
//	width := creator.CMToPoints(21)  // ≈ 595 points (A4 width)
func CMToPoints(cm float64) float64 {
	return document.CMToPoints(cm)
}
