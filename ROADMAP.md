# GxPDF Roadmap

Strategic development plan for the GxPDF PDF library.

**Current Version**: v0.3.0 "Parser Hardening"

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

## Current Development

### v0.4.0 "Creator API"

**Status**: In Development

#### 35+ Built-in Page Sizes
- **ISO A series** (A0–A8), **B series** (B0–B6), **C/DL envelopes**
- **ANSI engineering** (C, D, E), **US sizes** (Letter, Legal, Tabloid, Executive, Half Letter)
- **Photo** (4×6, 5×7, 8×10), **Book publishing** (Digest, US Trade Book)
- **Presentation slides** (16:9, 4:3) — PowerPoint/Keynote defaults
- **JIS B series** (B4, B5), **US #10 envelope**
- Map-based architecture for maintainability

#### Custom Page Dimensions (#41)
- `NewPageWithDimensions(widthPt, heightPt)` for arbitrary sizes
- Unit conversion helpers: `InchesToPoints`, `MMToPoints`, `CMToPoints` + reverse

#### Landscape Orientation (#41)
- `NewPageWithSize(size, Landscape)` — industry-standard approach
- True landscape via swapped MediaBox (no `/Rotate`)

#### Text Rotation (#42)
- `AddTextRotated` / `AddTextColorRotated` — standard 14 fonts
- `AddTextCustomFontRotated` / `AddTextCustomFontColorRotated` — TTF/OTF fonts
- Uses PDF `Tm` operator per ISO 32000 §9.4.2

## Planned Features

### v0.5.0 - Encryption Reading & Digital Signatures

**Priority**: P2

#### Encrypted PDF Reading
- **Standard Security Handler** - V4/R4 AES-128 with empty password
- **Key Derivation** - MD5/SHA-256 per PDF spec revision
- **Stream/String Decryption** - AES-CBC before decompression

#### Digital Signatures
- **Sign PDFs** - Apply digital signatures with PKCS#12 certificates
- **Verify Signatures** - Validate existing signatures
- **Visible/Invisible** - Both signature types

#### Developer Experience
- **Fluent Text API** - Chainable text rendering
- **Paragraph Support** - Multi-line text with wrapping
- **Y-Cursor** - Automatic vertical positioning
- **Simple Table API** - Easy table creation

### v0.6.0 - PDF/A & Advanced Features

- **PDF/A-1b** - Basic archival compliance
- **PDF/A-2b** - Extended archival compliance
- **SVG Import** - Convert SVG to PDF graphics
- **Invoice Template** - Pre-built invoice generation
- **Chart Integration** - Embed charts in PDFs

### v0.7.0 - Rendering & Optimization

- **PDF Render** - Render PDF pages to images
- **Barcode Generation** - QR codes, Code128, etc.
- **Font Subsetting Optimization** - Reduce file size
- **Linearization** - Fast web view support

### v1.0.0 - Stable Release

- API stability guarantee
- Performance optimization
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
| Encrypted PDF Reading | Planned | v0.5.0 |
| Digital Signatures | Planned | v0.5.0 |
| Fluent Text API | Planned | v0.5.0 |
| PDF/A Compliance | Planned | v0.6.0 |
| PDF Render to Image | Planned | v0.7.0 |

## Backlog (11 tasks)

| ID | Feature | Priority | Status | Description |
|----|---------|----------|--------|-------------|
| feat-067 | Custom Page Dimensions | **P1** | **Done** | Arbitrary page sizes (#41) |
| feat-068 | Text Rotation | **P1** | **Done** | Rotated text via Tm operator (#42) |
| feat-069 | Paper Sizes Expansion | **P1** | **Done** | 35+ built-in page sizes |
| feat-042 | Encrypted PDF Reading | **P2** | Backlog | AES-128 with empty password |
| feat-037 | Digital Signatures | **P2** | Backlog | Sign and verify PDFs |
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

GxPDF uses Domain-Driven Design (DDD):

```
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
