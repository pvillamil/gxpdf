# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - Unreleased "Opacity & Bezier"

### Added
- **Text Opacity** - New `AddTextColorAlpha`, `AddTextColorRotatedAlpha`, `AddTextCustomFontColorAlpha`, `AddTextCustomFontColorRotatedAlpha` methods (#46)
  - ExtGState transparency via `/ca` and `/CA` keys per ISO 32000 §11.6.4.4
  - Works with both standard 14 fonts and custom TTF/OTF fonts
  - Combines with rotation for watermark-style effects
- **Quadratic Bezier Curves** - `DrawQuadBezierCurve` with `QuadBezierSegment` struct (#45)
  - Exact degree elevation to cubic Bezier (not approximation)
  - `QuadBezierSegment.ToCubic()` for manual conversion
  - Multi-segment paths with full styling (stroke, fill, dash, opacity, gradient)

### Fixed
- **Shape Opacity** - `Opacity` on shape option structs now works correctly (#47)
  - Root cause: 3-layer pipeline gap — opacity accepted by creator but dropped during conversion
  - Fix propagates opacity through `convertOptions` → `writer.GraphicsOp` → ExtGState `gs` operator
  - Affects all shape types: circles, ellipses, rectangles, polygons, polylines, lines, Bezier curves

---

## [0.4.0] - 2026-02-21 "Creator API"

### Added
- **35+ Page Sizes** - Expanded from 8 to 38 built-in page sizes
  - ISO A series (A0–A8), ISO B series (B0–B6), ISO C/DL envelopes
  - ANSI engineering (C, D, E), photo sizes (4×6, 5×7, 8×10)
  - Book publishing (Digest, US Trade Book), JIS B series (B4, B5)
  - Presentation slides (16:9, 4:3) — PowerPoint/Keynote defaults
  - US #10 envelope, Executive, Half Letter
  - Map-based architecture replaces dual-switch for maintainability
- **Custom Page Dimensions** - `NewPageWithDimensions(widthPt, heightPt)` for arbitrary sizes (#41)
  - Unit conversion helpers: `InchesToPoints`, `MMToPoints`, `CMToPoints`
  - Reverse conversions: `PointsToInches`, `PointsToMM`, `PointsToCM`
- **Landscape Orientation** - `NewPageWithSize(size, Landscape)` parameter (#41)
  - Industry-standard swapped-MediaBox approach (no `/Rotate`)
  - `Portrait` / `Landscape` typed constants
- **Text Rotation** - `AddTextRotated` and `AddTextColorRotated` methods (#42)
  - Uses PDF `Tm` (text matrix) operator per ISO 32000 §9.4.2
  - Counter-clockwise rotation in degrees around origin point
  - Custom font variants: `AddTextCustomFontRotated`, `AddTextCustomFontColorRotated`

### Fixed
- **staticcheck QF1012** - Use `fmt.Fprintf` instead of `WriteString(Sprintf)` in font descriptors and table formatting

---

## [0.3.0] - 2026-02-16 "Parser Hardening"

### Added
- **Logging Package** - slog-based configurable logging (#21)
  - `logging.SetLogger()` / `logging.Logger()` API
  - Silent by default, opt-in via `slog.Handler`
  - All convenience methods (`ExtractText`, `ExtractTables`, `GetImages`) log errors via slog
- **Image XObject Rendering** - Complete image rendering in Writer (#36, #37)
  - `DrawImage()` and `DrawImageFit()` now produce visible images in PDFs
  - JPEG support via `/Filter /DCTDecode`
  - PNG support via `/Filter /FlateDecode` with `/SMask` for alpha channel
  - Proper CTM transformation for positioning and scaling
- **Watermark Rendering** - Writer now renders watermarks (#38)
  - Text watermarks with rotation, opacity, and font support
  - ExtGState for transparency
- **Text Extraction Example** - Added to README

### Fixed
- **Error Propagation** - Public API no longer silently swallows errors (#35, #39)
  - `ExtractTextFromPage()` now properly returns errors
  - Convenience methods log errors via slog instead of discarding them
  - `PageCount()`, `ExtractTables()`, `GetImages()` all log on failure
- **Parser Robustness** - 9 parser fixes from community contributions (#21-#33)
  - Leading whitespace before `%PDF-` header (#23)
  - CR line endings in `startxref` (#25)
  - Trailing garbage after `%%EOF` with progressive search (#27)
  - CMap `uint16` infinite loop — DoS vulnerability fix (#28)
  - Token position after indirect `Length` (#32)
  - Progressive xref stream buffer 1KB→4KB (#30)
  - `/W [0 0 0]` in xref streams (#31)
  - PNG predictor support for xref streams — all 5 filter types (#24)
  - Off-by-one xref object recovery with lenient parsing (#33)
- **Redundant `min()` helper** - Removed in favor of Go 1.21+ builtin (#29)

### Contributors
- [@mikeschinkel](https://github.com/mikeschinkel) — 11 PRs merged (parser hardening, logging)

---

## [0.2.1] - 2026-02-05

### Fixed
- **Hybrid-Reference PDF Support** - Fix `/Prev` chain and `/XRefStm` parsing (#19)
  - Follow `/Prev` links in trailer for incremental updates
  - Parse `/XRefStm` supplementary cross-reference streams
  - Cycle detection and depth limit for malformed PDFs
  - Fixes "object N not found in xref table" for MS Word-generated PDFs

---

## [0.2.0] - 2026-02-03 "Graphics Revolution"

### Added

#### Skia-like Graphics API (for GoGPU/gg integration)
- **Alpha Channel Support** - Transparency via ExtGState
  - `ColorRGBA` struct with alpha channel (0.0-1.0)
  - `Opacity` parameter on all drawing operations
  - ExtGState caching for efficient PDF output
  - 12 standard PDF blend modes
- **Push/Pop Graphics State** - Skia-like state stack
  - `Surface` type with state management
  - `PushTransform()`, `PushOpacity()`, `PushBlendMode()`
  - `Pop()` to restore previous state
  - `Transform` API: Translate, Scale, Rotate, Skew
- **Fill/Stroke Separation** - Independent fill and stroke
  - `Fill` struct: Paint, Opacity, FillRule (NonZero, EvenOdd)
  - `Stroke` struct: Paint, Width, LineCap, LineJoin, Dash
  - `SetFill()`, `SetStroke()` on Surface
  - LineCap: Butt, Round, Square
  - LineJoin: Miter, Round, Bevel
- **Paint Interface** - Unified color/gradient abstraction
  - `RGB()`, `RGBA()`, `Hex()`, `GrayN()` convenience functions
  - Color, ColorRGBA, ColorCMYK implement Paint
  - Ready for Gradient integration
- **Path Builder API** - Full vector path support
  - `NewPath()` with fluent API
  - `MoveTo()`, `LineTo()`, `CubicTo()`, `QuadraticTo()`, `Close()`
  - Shape helpers: `AddRect()`, `AddRoundedRect()`, `AddCircle()`, `AddEllipse()`, `AddArc()`
  - `DrawPath()`, `FillPath()`, `StrokePath()` on Surface
  - QuadraticTo automatically converts to cubic (PDF spec)

#### Forms API (Interactive PDF Forms)
- **Form Reading** - Read interactive form fields from PDFs
  - `Document.GetFormFields()` - Get all form fields
  - `Document.GetFieldValue(name)` - Get specific field value
  - `Document.HasForm()` - Check if PDF has interactive form
  - `FormField` type with accessors: Name, Type, Value, Options, Flags
  - Support for Text, Button, Choice, Signature field types
- **Form Writing** - Fill form fields programmatically
  - `Appender.SetFieldValue(name, value)` - Set field value
  - `Appender.GetFieldValue(name)` - Get current/pending value
  - Type validation (string for text, bool/string for checkboxes)
  - Option validation for choice fields
- **Form Flattening** - Convert forms to static content
  - `Appender.FlattenForm()` - Flatten all fields
  - `Appender.FlattenFields(names...)` - Flatten specific fields
  - `Appender.CanFlattenForm()` - Check if flattening is possible
- **WASM/Byte API** - Generate PDFs in memory
  - `Creator.WriteTo(io.Writer)` - Write to any writer
  - `Creator.Bytes()` - Get PDF as byte slice
  - `NewPdfWriterFromWriter(io.Writer)` - Low-level writer

#### Advanced Graphics
- **Linear Gradients** - Axial shading (ShadingType 2)
  - `NewLinearGradient(x1, y1, x2, y2)` constructor
  - `AddColorStop()` for color transitions
  - ExtendStart/ExtendEnd flags
- **Radial Gradients** - Radial shading (ShadingType 3)
  - `NewRadialGradient(x0, y0, r0, x1, y1, r1)` constructor
  - Focal point support (inner/outer circle)
- **ClipPath Support** - Clipping path operations
  - `PushClipPath()` with NonZero and EvenOdd fill rules
  - Convenience methods: `PushClipRect`, `PushClipCircle`, `PushClipEllipse`
  - PDF 1.7 Spec Section 8.5.4 compliant

---

## Planned (v0.6.0+)
- Encrypted PDF reading (AES-128, RC4-128, user password)
- Digital signatures (sign and verify)
- PDF/A compliance

---

## [0.1.1] - 2026-01-30

### Added
- **Full Unicode Font Embedding** - Complete TrueType/OpenType infrastructure
  - Cyrillic, CJK (Chinese, Japanese, Korean), and special symbols support
  - TTF parser extensions: `post`, `OS/2`, `name` table parsing
  - FontDescriptor generator with all PDF metrics
  - ToUnicode CMap generation for text extraction
  - Font subsetting with deterministic naming (XXXXXX+FontName)
  - Type 0 Composite Font support for full Unicode range
- **Text Clipping** - Clip text to table cell boundaries
- **Enterprise Showcase** - Professional 7-page PDF brochure demonstrating all features

### Fixed
- **hhea Table Parsing** - Corrected numOfLongHorMetrics offset for proper glyph widths
- **Glyph Width Calculation** - Fixed empty GlyphWidths map issue
- **PostScriptName Parsing** - Fixed UTF-16BE decoding in `name` table (was causing garbled font names and rendering issues in PDF viewers)

### Planned
- Form filling (fill existing PDF forms)
- Form flattening (convert forms to static content)
- Digital signatures (sign and verify)
- PDF/A compliance (archival format)
- SVG import

---

## [0.1.0] - 2026-01-07

Initial public release of GxPDF - a modern, enterprise-grade PDF library for Go.

### Added

#### PDF Creation (Creator API)
- **Document Creation** - Create PDF documents from scratch
- **Text Rendering** - Add text with multiple fonts, sizes, and colors
- **Graphics** - Draw lines, rectangles, circles, polygons, ellipses, Bezier curves
- **Gradients** - Linear and radial gradient fills
- **Color Spaces** - RGB and CMYK color support
- **Tables** - Create tables with borders, backgrounds, and merged cells
- **Images** - Embed JPEG and PNG images with transparency support
- **Fonts** - Standard 14 PDF fonts + TTF/OTF font embedding
- **Chapters & TOC** - Document structure with auto-generated Table of Contents
- **Annotations** - Sticky notes, highlights, underlines, stamps
- **Interactive Forms (AcroForm)** - Text fields, checkboxes, radio buttons, dropdowns, list boxes
- **Encryption** - RC4 (40/128-bit) and AES (128/256-bit) encryption
- **Watermarks** - Text watermarks with rotation, opacity, and positioning
- **Bookmarks** - PDF outline/navigation structure
- **Page Operations** - Merge, split, rotate, append pages

#### PDF Reading & Extraction
- **PDF Parser** - Read PDF 1.0-2.0 files
  - Cross-reference table parsing (traditional and stream-based)
  - Object and stream parsing with caching
  - Indirect reference resolution
- **Text Extraction** - Extract text with X,Y positions
  - Unicode support (including Cyrillic)
  - Font decoding (CMap, Identity-H)
  - Content stream parsing
- **Table Extraction** - Industry-leading accuracy
  - 4-Pass Hybrid Detection Algorithm
  - Lattice mode (ruling lines) + Stream mode (whitespace analysis)
  - Multi-line cell support
  - 100% accuracy on real-world bank statements
- **Image Extraction** - Extract embedded images
- **Export Formats** - CSV, JSON, Excel

#### Infrastructure
- **Stream Decoders** - FlateDecode, ASCII85Decode, ASCIIHexDecode
- **Thread Safety** - Object cache with sync.RWMutex
- **DDD Architecture** - Domain-Driven Design with Rich Domain Model

### Architecture
- **Domain Layer** - Pure business logic with no external dependencies
- **Application Layer** - Use cases and service orchestration
- **Infrastructure Layer** - PDF parsing, encoding, file I/O
- **Public API** - Clean, intuitive API with functional options pattern

### Testing
- Comprehensive unit tests
- Integration tests with real PDF files
- Race detector clean
- golangci-lint with 15+ linters: 0 issues

### Documentation
- Full API documentation (godoc)
- Code examples for all features
- Architecture documentation
- Contributing guidelines
- Security policy

---

## Project Information

**Repository**: https://github.com/coregx/gxpdf

**License**: MIT

**Go Version**: 1.25+

---

[0.5.0]: https://github.com/coregx/gxpdf/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/coregx/gxpdf/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/coregx/gxpdf/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/coregx/gxpdf/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/coregx/gxpdf/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/coregx/gxpdf/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/coregx/gxpdf/releases/tag/v0.1.0
