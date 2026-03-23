package builder

import "github.com/coregx/gxpdf/layout"

// RichTextBuilder accumulates styled text spans that together form a single
// mixed-style paragraph. It is passed to the callback in [Container.RichText].
//
// Usage:
//
//	c.RichText(func(rt *builder.RichTextBuilder) {
//	    rt.Span("Normal text ")
//	    rt.Span("bold text ", builder.Bold())
//	    rt.Span("red italic", builder.Italic(), builder.TextColor(builder.Red))
//	})
type RichTextBuilder struct {
	b         *Builder
	baseStyle layout.Style
	fragments []layout.RichTextFragment
}

// Span adds a styled text fragment to the paragraph. The opts are applied on
// top of the base style inherited from the [Container.RichText] call.
//
// Example:
//
//	rt.Span("Important: ", builder.Bold(), builder.TextColor(builder.Red))
func (rt *RichTextBuilder) Span(text string, opts ...TextOption) {
	style := applyTextOptions(rt.baseStyle, opts)
	rt.fragments = append(rt.fragments, layout.RichTextFragment{
		Text:  text,
		Style: style,
	})
}

// Link adds a hyperlink span. The text is rendered with underline and the
// link color (blue by default) unless overridden by opts. The url parameter
// sets the hyperlink target.
//
// Example:
//
//	rt.Link("GxPDF repository", "https://github.com/coregx/gxpdf")
func (rt *RichTextBuilder) Link(text, url string, opts ...TextOption) {
	// Default link style: blue + underline.
	linkBase := rt.baseStyle
	linkBase.Color = layout.RGB(0.071, 0.357, 0.722) // #1259B8 — accessible link blue
	linkBase.Underline = true
	style := applyTextOptions(linkBase, opts)
	rt.fragments = append(rt.fragments, layout.RichTextFragment{
		Text:  text,
		Style: style,
		URL:   url,
	})
}

// build returns the layout.RichText element accumulated by this builder.
// Called internally by Container.RichText after the callback returns.
func (rt *RichTextBuilder) build(align layout.Align, lineHeight float64) *layout.RichText {
	return &layout.RichText{
		Fragments:  rt.fragments,
		Align:      align,
		LineHeight: lineHeight,
		Fonts:      rt.b.fontResolver(),
	}
}
