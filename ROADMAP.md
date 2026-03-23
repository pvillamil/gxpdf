# GxPDF Roadmap

Strategic development plan for the GxPDF PDF library.

**Current Version**: See [CHANGELOG.md](CHANGELOG.md)

## Version History

### v0.3.0 "Parser Hardening"

**Released**: February 2026

Major parser robustness improvements and rendering fixes:

#### New: Logging Package
- slog-based configurable logging (silent by default)
- Error visibility for convenience methods (ExtractText, etc.)

#### Image & Watermark Rendering (Writer)
- Complete image XObject rendering (JPEG + PNG + alpha)
- Watermark rendering with rotation and opacity
- Fixes #36: DrawImage/DrawImageFit now work correctly

#### Error Propagation
- Public API methods properly return/log errors instead of silently failing
- ExtractTextFromPage returns actual errors

#### Parser Hardening (11 community PRs by @mikeschinkel)
- Leading whitespace, CR line endings, trailing garbage after %%EOF
- CMap uint16 infinite loop fix (DoS vulnerability)
- PNG predictor support for xref streams (all 5 filter types)
- Progressive xref stream buffer, /W [0 0 0] support
- Off-by-one xref object recovery with lenient parsing

### v0.2.0 "Graphics Revolution"

**Released**: February 2026

Major graphics and forms capabilities:

#### Skia-like Graphics API (for GoGPU/gg integration)
- Alpha channel support with transparency
- Push/Pop graphics state stack
- Fill/Stroke separation with Paint interface
- Path Builder API (MoveTo, LineTo, CubicTo, etc.)
- Linear and Radial gradients
- ClipPath support

#### Forms API
- Form field reading (GetFormFields, GetFieldValue)
- Form field writing (SetFieldValue with validation)
- Form flattening (FlattenForm, FlattenFields)

#### Platform Support
- WASM API (WriteTo, Bytes for in-memory generation)

### v0.1.1

**Released**: January 2026

Unicode font embedding infrastructure:
- Full Unicode support (Cyrillic, CJK, symbols)
- TrueType font subsetting with ToUnicode CMap
- Type 0 Composite Font for full Unicode range
- Enterprise showcase PDF demonstrating all features
- Fixed PostScriptName parsing for proper font rendering

### v0.1.0

**Released**: January 2026

Full-featured PDF library with:
- PDF creation (Creator API)
- PDF reading and parsing
- Text and table extraction
- Multiple export formats
- DDD architecture

### v0.4.0 "Creator API"

**Released**: February 2026

Page sizes, custom dimensions, landscape orientation, and text rotation:

#### 35+ Built-in Page Sizes
- ISO A/B/C, ANSI, photo, book, JIS, envelopes, slides
- Map-based architecture for maintainability

#### Custom Page Dimensions (#41)
- `NewPageWithDimensions(widthPt, heightPt)` for arbitrary sizes
- Unit conversion helpers: `InchesToPoints`, `MMToPoints`, `CMToPoints` + reverse

#### Landscape Orientation (#41)
- `NewPageWithSize(size, Landscape)` — industry-standard swapped-MediaBox

#### Text Rotation (#42)
- `AddTextRotated` / `AddTextColorRotated` — standard 14 + custom TTF/OTF fonts
- Angle normalization to [0, 360)

### v0.5.0 "Opacity & Bezier"

**Released**: February 2026

#### Shape Opacity Fix (#47)
- Fixed 3-layer pipeline gap: opacity now propagated through writer

#### Quadratic Bezier Curves (#45)
- `DrawQuadBezierCurve` with `QuadBezierSegment` struct
- Exact degree elevation to cubic (not approximation)

#### Text Opacity (#46)
- `AddTextColorAlpha`, `AddTextColorRotatedAlpha` + custom font variants
- ExtGState transparency via `/ca` and `/CA` keys

### v0.6.0 "Encrypted Reading & Gradients"

**Released**: February 2026

#### Encrypted PDF Reading (#34)
- Standard Security Handler: RC4-40, RC4-128, AES-128
- `OpenWithPassword()` / `Open()` with transparent empty-password handling
- `ErrPasswordRequired` sentinel error

#### Full Gradient Rendering (#57)
- PDF Shading Type 2 (axial) + Type 3 (radial), multi-stop
- Clip+Shade technique on all shape types

#### ExtGState Object Creation Fix (#46, #47)
- Shape and text opacity now produce valid PDF output

## Current Development

### Unreleased (on main)

- **Arc Drawing** (#59) — elliptical/circular arcs with wedge/chord fill modes
- **Declarative Builder API** — QuestPDF-inspired layout with 12-col grid, auto-pagination

### Planned

#### v0.7.0 - "Builder & Signatures"

**In Progress**

- **Declarative Builder API** (done) — QuestPDF-inspired layout with 12-col grid, auto-pagination
  - `layout/` pure computation engine (81% coverage)
  - `builder/` user-facing API (81% coverage)
  - Own types (Value, Color, Size) — no layout/ import leak
  - Font measurement bridge (Standard 14 + TTF)
- **Tables with ColSpan/RowSpan** (planned) — enterprise-grade tables
- **Rich Text** (planned) — mixed-style inline text
- **Digital Signatures** (planned) — CMS/PKCS#7 + PAdES
- **Test Coverage Push** (planned) — target 80%+ project-wide

#### v0.8.0 - Advanced Features

- **PDF/A Compliance** - Archival format support
- **SVG Import** - Convert SVG to PDF graphics
- **PDF Render** - Render PDF pages to images
- **Barcode Generation** - QR codes, Code128, etc.
- **HTML to PDF** - Render WYSIWYG HTML into PDF (may be separate library)

#### v1.0.0 - Stable Release

- API stability guarantee
- Test coverage 80%+
- Comprehensive documentation
- Security audit

## Feature Status

| Feature | Status | Version |
|---------|--------|---------|
| PDF Creation | Done | v0.1.0 |
| Text Rendering | Done | v0.1.0 |
| Graphics (shapes, curves) | Done | v0.1.0 |
| Tables | Done | v0.1.0 |
| Images (JPEG, PNG) | Done | v0.1.0 |
| Fonts (Standard 14 + TTF) | Done | v0.1.0 |
| Unicode Font Embedding | Done | v0.1.1 |
| Chapters & TOC | Done | v0.1.0 |
| Annotations | Done | v0.1.0 |
| Interactive Forms | Done | v0.1.0 |
| Encryption (RC4, AES) | Done | v0.1.0 |
| Watermarks | Done | v0.1.0 |
| PDF Reading | Done | v0.1.0 |
| Text Extraction | Done | v0.1.0 |
| Table Extraction | Done | v0.1.0 |
| Export (CSV, JSON, Excel) | Done | v0.1.0 |
| Skia-like Graphics API | Done | v0.2.0 |
| Linear/Radial Gradients | Done | v0.2.0 |
| ClipPath Support | Done | v0.2.0 |
| Form Reading | Done | v0.2.0 |
| Form Filling | Done | v0.2.0 |
| Form Flattening | Done | v0.2.0 |
| WASM API | Done | v0.2.0 |
| Logging (slog) | Done | v0.3.0 |
| Image XObject Rendering | Done | v0.3.0 |
| Watermark Rendering | Done | v0.3.0 |
| Error Propagation | Done | v0.3.0 |
| Parser Hardening | Done | v0.3.0 |
| 35+ Page Sizes | Done | v0.4.0 |
| Custom Page Dimensions | Done | v0.4.0 |
| Landscape/Portrait | Done | v0.4.0 |
| Text Rotation | Done | v0.4.0 |
| Shape Opacity | Done | v0.5.0 |
| Quadratic Bezier Curves | Done | v0.5.0 |
| Text Opacity | Done | v0.5.0 |
| Gradient Rendering (PDF Shading) | Done | v0.6.0 |
| Encrypted PDF Reading | Done | v0.6.0 |
| Arc Drawing (elliptical/circular) | Done | unreleased |
| Declarative Builder API | Done | unreleased |
| Tables with ColSpan/RowSpan | Planned | v0.7.0 |
| Rich Text (mixed inline styles) | Planned | v0.7.0 |
| Digital Signatures | Planned | v0.7.0 |
| HTML to PDF | Planned | v0.8.0 |
| PDF/A Compliance | Planned | v0.8.0 |
| PDF Render to Image | Planned | v0.8.0 |
| SVG Import | Planned | v0.8.0 |
| Barcode / QR Code | Planned | v0.8.0 |

## Backlog

| ID | Feature | Priority | Status | Description |
|----|---------|----------|--------|-------------|
| feat-067 | Custom Page Dimensions | **P1** | **Done** | Arbitrary page sizes (#41) |
| feat-068 | Text Rotation | **P1** | **Done** | Rotated text via Tm operator (#42) |
| feat-069 | Paper Sizes Expansion | **P1** | **Done** | 35+ built-in page sizes |
| fix-006 | Shape Opacity | **P1** | **Done** | Pipeline gap fix (#47) |
| feat-074 | Quadratic Bezier | **P2** | **Done** | Degree elevation to cubic (#45) |
| feat-075 | Text Opacity | **P2** | **Done** | ExtGState transparency (#46) |
| feat-078 | Gradient Rendering | **P1** | **Done** | Full PDF Shading (Type 2/3) for gradients (#57) |
| feat-042 | Encrypted PDF Reading | **P1** | **Done** | RC4/AES-128 with password support |
| feat-079 | Arc Drawing | **P2** | **Done** | Elliptical/circular arcs with wedge/chord (#59) |
| feat-076 | Declarative Builder API | **P1** | **Done** | QuestPDF-inspired layout, 12-col grid, auto-pagination |
| feat-077 | Tables ColSpan/RowSpan | **P1** | Backlog | Enterprise-grade table layout in builder |
| feat-081 | Rich Text | **P1** | Backlog | Mixed-style inline text elements |
| feat-037 | Digital Signatures | **P1** | Backlog | Sign and verify PDFs (CMS/PKCS#7 + PAdES) |
| feat-080 | HTML to PDF | **P2** | Backlog | Render WYSIWYG HTML via GxPDF |
| feat-062 | Fluent Text API | P3 | Backlog | Chainable text methods |
| feat-063 | Paragraph | P3 | Backlog | Multi-line text container |
| feat-064 | Y-Cursor | P3 | Backlog | Auto vertical positioning |
| feat-065 | Simple Table API | P3 | Backlog | Easy table creation |
| feat-066 | Shape Builders | P3 | Backlog | Fluent shape construction |
| feat-036 | SVG Import | P3 | Backlog | Vector graphics import |
| feat-039 | Invoice Template | P3 | Backlog | Pre-built invoice |
| feat-040 | Chart Integration | P3 | Backlog | Embed charts |
| feat-041 | PDF Render | P3 | Backlog | Render to images |

## Architecture

GxPDF uses Domain-Driven Design (DDD) with a three-layer generation stack:

```
builder/        — User-facing declarative API
layout/         — Pure computation layout engine (zero PDF deps)
creator/        — Low-level PDF primitives

internal/
├── domain/         # Pure business logic
├── application/    # Use cases
└── infrastructure/ # PDF parsing, encoding
```

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Contributing

We welcome contributions! Priority areas:

- **Documentation** - Examples, tutorials
- **Tests** - Increase coverage
- **Performance** - Optimization
- **Features** - See planned features above

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Timeline

No fixed timelines. Features are released when ready and tested.

Priorities are based on:
1. User demand (GitHub issues)
2. Technical dependencies
3. Maintainer availability

## Feedback

Feature requests and feedback welcome:

- **GitHub Issues**: https://github.com/coregx/gxpdf/issues
- **Discussions**: https://github.com/coregx/gxpdf/discussions

---

*This roadmap is updated as priorities evolve.*
