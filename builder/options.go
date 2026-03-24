package builder

import (
	"os"

	"github.com/coregx/gxpdf/creator"
	"github.com/coregx/gxpdf/layout"
)

// Option is a functional option that configures the document-level Builder.
type Option func(*config)

// config holds the document-level configuration assembled from Option values.
type config struct {
	pageSize     layout.Size
	margins      layout.Edges
	fonts        map[string]*creator.CustomFont
	defaultStyle layout.Style
	meta         metadata
}

// metadata holds PDF document metadata.
type metadata struct {
	title  string
	author string
}

// defaultConfig returns a config with sensible defaults:
// A4 page size, 20mm margins, Helvetica 12pt, black text.
func defaultConfig() config {
	return config{
		pageSize: layout.PageA4,
		margins: layout.Edges{
			Top:    layout.Mm(20),
			Right:  layout.Mm(15),
			Bottom: layout.Mm(20),
			Left:   layout.Mm(15),
		},
		fonts:        make(map[string]*creator.CustomFont),
		defaultStyle: layout.DefaultStyle(),
	}
}

// WithPageSize sets the default page size for all pages in the document.
//
// Example:
//
//	builder.NewBuilder(builder.WithPageSize(builder.Letter))
func WithPageSize(size Size) Option {
	return func(c *config) {
		c.pageSize = size.toLayout()
	}
}

// WithMargins sets the default page margins using Value units (pt, mm, cm, in).
//
// Example:
//
//	builder.NewBuilder(builder.WithMargins(
//	    builder.Mm(20), builder.Mm(15), builder.Mm(20), builder.Mm(15),
//	))
func WithMargins(top, right, bottom, left Value) Option {
	return func(c *config) {
		c.margins = layout.Edges{
			Top:    top.toLayout(),
			Right:  right.toLayout(),
			Bottom: bottom.toLayout(),
			Left:   left.toLayout(),
		}
	}
}

// WithFont registers a TrueType font from raw bytes under the given family name.
// Once registered, the family name can be used with FontFamily() text option.
//
// Example:
//
//	interTTF, _ := os.ReadFile("fonts/Inter-Regular.ttf")
//	builder.NewBuilder(builder.WithFont("Inter", interTTF))
func WithFont(family string, data []byte) Option {
	return func(c *config) {
		if len(data) == 0 {
			return
		}
		// Write to a temporary file because creator.LoadFont expects a path.
		f, err := os.CreateTemp("", "gxpdf-font-*.ttf")
		if err != nil {
			return
		}
		defer func() { _ = os.Remove(f.Name()) }()
		if _, err := f.Write(data); err != nil {
			_ = f.Close()
			return
		}
		_ = f.Close()

		font, err := creator.LoadFont(f.Name())
		if err != nil {
			return
		}
		c.fonts[family] = font
	}
}

// WithFontFile registers a TrueType font from a file path under the given family name.
//
// Example:
//
//	builder.NewBuilder(builder.WithFontFile("Inter", "fonts/Inter-Regular.ttf"))
func WithFontFile(family string, path string) Option {
	return func(c *config) {
		font, err := creator.LoadFont(path)
		if err != nil {
			// Errors are deferred — the builder will surface this on Build().
			return
		}
		c.fonts[family] = font
	}
}

// WithDefaultFontSize sets the default font size for all text elements.
func WithDefaultFontSize(size float64) Option {
	return func(c *config) {
		c.defaultStyle.FontSize = size
	}
}

// WithDefaultFontFamily sets the default font family for all text elements.
// The family must have been registered via WithFont or WithFontFile.
func WithDefaultFontFamily(family string) Option {
	return func(c *config) {
		c.defaultStyle.Font.Family = family
	}
}

// WithDefaultColor sets the default text color for all text elements.
func WithDefaultColor(color Color) Option {
	return func(c *config) {
		c.defaultStyle.Color = color.toLayout()
	}
}

// WithDefaultLineHeight sets the default line height multiplier for all text elements.
func WithDefaultLineHeight(multiplier float64) Option {
	return func(c *config) {
		c.defaultStyle.LineHeight = multiplier
	}
}

// WithTitle sets the PDF document title metadata.
func WithTitle(title string) Option {
	return func(c *config) {
		c.meta.title = title
	}
}

// WithAuthor sets the PDF document author metadata.
func WithAuthor(author string) Option {
	return func(c *config) {
		c.meta.author = author
	}
}

// pageDef is the internal representation of a page built by PageBuilder.
// It is later converted to a layout.PageDef for the paginator.
type pageDef struct {
	size    *layout.Size
	margins *layout.Edges
	header  []layout.Element
	footer  []layout.Element
	content []layout.Element
}

// resolvedSize returns the effective page size: the page-specific size if set,
// otherwise the document-level default.
func (pd *pageDef) resolvedSize(def layout.Size) layout.Size {
	if pd.size != nil {
		return *pd.size
	}
	return def
}

// resolvedMargins returns the effective margins: the page-specific margins if
// set, otherwise the document-level defaults.
func (pd *pageDef) resolvedMargins(def layout.Edges) layout.Edges {
	if pd.margins != nil {
		return *pd.margins
	}
	return def
}

// toLayoutPageDef converts the internal pageDef to a layout.PageDef.
func (pd *pageDef) toLayoutPageDef(defSize layout.Size, defMargins layout.Edges) *layout.PageDef {
	return &layout.PageDef{
		Size:    pd.resolvedSize(defSize),
		Margins: pd.resolvedMargins(defMargins),
		Header:  pd.header,
		Footer:  pd.footer,
		Content: pd.content,
	}
}

// validateFontFile validates that the given path can be loaded as a font.
// Returns an error suitable for accumulation in the Builder error list.
