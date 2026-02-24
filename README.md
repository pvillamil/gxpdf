# GxPDF - Enterprise-Grade PDF Library for Go

![GxPDF Cover](assets/gh_cover.png)

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/coregx/gxpdf/actions/workflows/test.yml/badge.svg)](https://github.com/coregx/gxpdf/actions/workflows/test.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/coregx/gxpdf)](https://pkg.go.dev/github.com/coregx/gxpdf)
[![Go Report Card](https://goreportcard.com/badge/github.com/coregx/gxpdf)](https://goreportcard.com/report/github.com/coregx/gxpdf)

**GxPDF** is a modern, high-performance PDF library for Go, built with clean architecture and Go 1.25+ best practices.

**[View Example PDF](assets/gxpdf_enterprise_brochure.pdf)** - See what GxPDF can create!

## Key Features

### PDF Creation (Creator API)
- **Text & Typography** - Rich text with multiple fonts, styles, and colors
- **Graphics** - Lines, rectangles, circles, polygons, ellipses, cubic and quadratic Bezier curves
- **Gradients** - Linear and radial gradient fills with full PDF Shading (multi-stop support)
- **Color Spaces** - RGB and CMYK support
- **Tables** - Complex tables with merged cells, borders, backgrounds
- **Images** - JPEG and PNG with transparency support
- **Fonts** - Standard 14 PDF fonts + TTF/OTF embedding with full Unicode support (Cyrillic, CJK, symbols)
- **Document Structure** - Chapters, auto-generated Table of Contents
- **Annotations** - Sticky notes, highlights, underlines, stamps
- **Interactive Forms** - Text fields, checkboxes, radio buttons, dropdowns
- **Security** - RC4 (40/128-bit) and AES (128/256-bit) encryption
- **Watermarks** - Text watermarks with rotation and opacity
- **Page Sizes** - 35+ built-in sizes (ISO A/B/C, ANSI, photo, book, JIS, envelopes, slides)
- **Custom Dimensions** - Arbitrary page sizes with unit conversion helpers (inches, mm, cm)
- **Landscape/Portrait** - True landscape via `NewPageWithSize(A4, Landscape)` — industry-standard swapped-MediaBox
- **Text Rotation** - Rotated text at any angle via `Tm` operator (standard 14 + custom TTF fonts)
- **Opacity/Transparency** - Text and shape opacity via ExtGState (`AddTextColorAlpha`, shape `Opacity` option)
- **Page Operations** - Merge, split, rotate, append

### PDF Reading & Extraction
- **Encrypted PDF Reading** - Open password-protected PDFs (RC4-40/128, AES-128)
- **Table Extraction** - Industry-leading accuracy (100% on bank statements)
- **Text Extraction** - Full text with positions and Unicode support
- **Image Extraction** - Extract embedded images
- **Export Formats** - CSV, JSON, Excel

## Installation

```bash
go get github.com/coregx/gxpdf
```

**Requirements**: Go 1.25 or later

## Quick Start

### Creating a PDF Document

```go
package main

import (
    "log"
    "github.com/coregx/gxpdf/creator"
)

func main() {
    c := creator.New()
    c.SetTitle("My Document")
    c.SetAuthor("GxPDF")

    page, _ := c.NewPage()

    // Add text
    page.AddText("Hello, GxPDF!", 100, 750, creator.HelveticaBold, 24)
    page.AddText("Professional PDF creation in Go", 100, 720, creator.Helvetica, 12)

    // Draw shapes
    page.DrawRectangle(100, 600, 200, 100, &creator.RectangleOptions{
        FillColor:   &creator.Blue,
        StrokeColor: &creator.Black,
        StrokeWidth: 2,
    })

    if err := c.WriteToFile("output.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

### Unicode Text with Custom Fonts

```go
c := creator.New()
page, _ := c.NewPage()

// Load custom font with Unicode support
font, _ := c.LoadFont("/path/to/arial.ttf")

// Cyrillic text
page.AddTextCustomFont("Привет, мир!", 100, 700, font, 18)

// CJK text (requires appropriate font like Malgun Gothic)
cjkFont, _ := c.LoadFont("/path/to/malgun.ttf")
page.AddTextCustomFont("你好世界 • 안녕하세요", 100, 670, cjkFont, 16)

c.WriteToFile("unicode.pdf")
```

### Creating Encrypted PDFs

```go
c := creator.New()

c.SetEncryption(creator.EncryptionOptions{
    UserPassword:  "user123",
    OwnerPassword: "owner456",
    Permissions:   creator.PermissionPrint | creator.PermissionCopy,
    Algorithm:     creator.EncryptionAES256,
})

page, _ := c.NewPage()
page.AddText("This document is encrypted!", 100, 750, creator.Helvetica, 14)
c.WriteToFile("encrypted.pdf")
```

### Creating Documents with Chapters and TOC

```go
c := creator.New()
c.EnableTOC()

ch1 := creator.NewChapter("Introduction")
ch1.Add(creator.NewParagraph("Introduction content..."))

ch1_1 := ch1.NewSubChapter("Background")
ch1_1.Add(creator.NewParagraph("Background information..."))

ch2 := creator.NewChapter("Methods")
ch2.Add(creator.NewParagraph("Methods description..."))

c.AddChapter(ch1)
c.AddChapter(ch2)

c.WriteToFile("document_with_toc.pdf")
```

### Interactive Forms (AcroForm)

```go
import "github.com/coregx/gxpdf/creator/forms"

c := creator.New()
page, _ := c.NewPage()

// Text field
nameField := forms.NewTextField("name", 100, 700, 200, 20)
nameField.SetLabel("Full Name:")
nameField.SetRequired(true)
page.AddField(nameField)

// Checkbox
agreeBox := forms.NewCheckbox("agree", 100, 660, 15, 15)
agreeBox.SetLabel("I agree to the terms")
page.AddField(agreeBox)

// Dropdown
countryDropdown := forms.NewDropdown("country", 100, 620, 150, 20)
countryDropdown.AddOption("us", "United States")
countryDropdown.AddOption("uk", "United Kingdom")
page.AddField(countryDropdown)

c.WriteToFile("form.pdf")
```

### Reading and Filling Forms

```go
// Read form fields from existing PDF
doc, _ := gxpdf.Open("form.pdf")
defer doc.Close()

// Check if document has a form
if doc.HasForm() {
    fields, _ := doc.GetFormFields()
    for _, f := range fields {
        fmt.Printf("%s (%s): %v\n", f.Name(), f.Type(), f.Value())
    }
}

// Fill form fields
app, _ := creator.NewAppender("form.pdf")
defer app.Close()

app.SetFieldValue("name", "John Doe")
app.SetFieldValue("email", "john@example.com")
app.SetFieldValue("agree", true)  // Checkbox
app.SetFieldValue("country", "USA")  // Dropdown

app.WriteToFile("filled_form.pdf")
```

### Flattening Forms

```go
// Convert form fields to static content (non-editable)
app, _ := creator.NewAppender("filled_form.pdf")
defer app.Close()

app.FlattenForm()  // Flatten all fields
// Or: app.FlattenFields("signature", "date")  // Specific fields

app.WriteToFile("flattened.pdf")
```

### Extracting Text from PDFs

```go
doc, _ := gxpdf.Open("document.pdf")
defer doc.Close()

// Extract text from a specific page (1-based)
text, err := doc.ExtractTextFromPage(1)
if err != nil {
    log.Fatal(err)
}
fmt.Println(text)

// Or iterate all pages
for _, page := range doc.Pages() {
    fmt.Println(page.ExtractText())
}
```

### Extracting Tables from PDFs

```go
doc, _ := gxpdf.Open("bank_statement.pdf")
defer doc.Close()

tables := doc.ExtractTables()
for _, table := range tables {
    fmt.Printf("Table: %d rows x %d cols\n", table.RowCount(), table.ColumnCount())

    // Export to CSV
    csv, _ := table.ToCSV()
    fmt.Println(csv)
}
```

### Reading Encrypted PDFs

```go
// PDFs with empty user password (permissions-only) open transparently
doc, _ := gxpdf.Open("bank_statement_encrypted.pdf")
defer doc.Close()

// PDFs requiring a password
doc, err := gxpdf.OpenWithPassword("protected.pdf", "secret")
if errors.Is(err, gxpdf.ErrPasswordRequired) {
    log.Fatal("Wrong password")
}
defer doc.Close()

fmt.Printf("Pages: %d\n", doc.PageCount())
```

## Package Structure

```
github.com/coregx/gxpdf
├── gxpdf.go          # Main entry point
├── export/           # Export formats (CSV, JSON, Excel)
├── creator/          # PDF creation API
│   └── forms/        # Interactive form fields
├── logging/          # Configurable debug logging
└── internal/         # Private implementation
    ├── application/  # Use cases (extraction, reading)
    └── infrastructure/ # PDF parsing, encoding, writing
```

## Documentation

- **[API Reference](https://pkg.go.dev/github.com/coregx/gxpdf)** - Full API documentation
- **[Examples](examples/)** - Code examples for all features
- **[Enterprise Brochure (PDF)](assets/gxpdf_enterprise_brochure.pdf)** - Sample PDF created with GxPDF
- **[Architecture](docs/ARCHITECTURE.md)** - DDD architecture overview
- **[Contributing](CONTRIBUTING.md)** - Contribution guidelines
- **[Security](SECURITY.md)** - Security policy

## Testing

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
```

### Debug Logging

Debug output is disabled by default. To enable it, configure a logger via the `logging` package:

```go
import (
    "log/slog"
    "os"
    "github.com/coregx/gxpdf/logging"
)

logging.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelWarn,
})))
```

Convenience methods like `ExtractText()`, `ExtractTables()`, and `GetImages()` log errors via slog instead of returning them. Enable logging to see why these methods return empty results.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for the full development plan and version history.

## License

GxPDF is released under the **MIT License**. See [LICENSE](LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/coregx/gxpdf/issues)
- **Discussions**: [GitHub Discussions](https://github.com/coregx/gxpdf/discussions)

---

**Built with Go 1.25+ and Domain-Driven Design**
