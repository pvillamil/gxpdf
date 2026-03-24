package builder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coregx/gxpdf/builder"
)

// TestValue_Cm covers the Cm() constructor.
func TestValue_Cm(t *testing.T) {
	v := builder.Cm(2.54)
	// 2.54 cm == 1 inch == 72 points; just verify it round-trips without panic.
	_ = v
}

// TestValue_Pct covers the Pct() constructor.
func TestValue_Pct(t *testing.T) {
	v := builder.Pct(50)
	_ = v
}

// TestValue_Auto covers the Auto() constructor.
func TestValue_Auto(t *testing.T) {
	v := builder.Auto()
	_ = v
}

// TestWithDefaultFontFamily covers the WithDefaultFontFamily option.
func TestWithDefaultFontFamily(t *testing.T) {
	b := builder.NewBuilder(builder.WithDefaultFontFamily("Helvetica"))
	require.NotNil(t, b)
	pdfBytes, err := b.Build()
	require.NoError(t, err)
	assertValidPDF(t, pdfBytes)
}

// TestWithDefaultColor covers the WithDefaultColor option.
func TestWithDefaultColor(t *testing.T) {
	b := builder.NewBuilder(builder.WithDefaultColor(builder.Black))
	require.NotNil(t, b)
	pdfBytes, err := b.Build()
	require.NoError(t, err)
	assertValidPDF(t, pdfBytes)
}

// TestWithDefaultLineHeight covers the WithDefaultLineHeight option.
func TestWithDefaultLineHeight(t *testing.T) {
	b := builder.NewBuilder(builder.WithDefaultLineHeight(1.5))
	require.NotNil(t, b)
	pdfBytes, err := b.Build()
	require.NoError(t, err)
	assertValidPDF(t, pdfBytes)
}

// TestContainer_RichText covers the RichText container method including Span and Link.
func TestContainer_RichText(t *testing.T) {
	b := builder.NewBuilder()
	b.Page(func(p *builder.PageBuilder) {
		p.Content(func(c *builder.Container) {
			c.RichText(func(rt *builder.RichTextBuilder) {
				rt.Span("Hello ")
				rt.Span("bold", builder.Bold())
				rt.Link("GxPDF", "https://github.com/coregx/gxpdf")
			})
		})
	})

	pdfBytes, err := b.Build()
	require.NoError(t, err)
	assertValidPDF(t, pdfBytes)
}

// TestCellVAlignTop covers the CellVAlignTop option.
func TestCellVAlignTop(t *testing.T) {
	b := builder.NewBuilder()
	b.Page(func(p *builder.PageBuilder) {
		p.Content(func(c *builder.Container) {
			c.Table(func(tbl *builder.TableBuilder) {
				tbl.Columns(builder.Fr(1), builder.Fr(1))
				tbl.Header(func(r *builder.TableRowBuilder) {
					r.Cell(func(cb *builder.CellBuilder) {
						cb.Text("A")
					}, builder.CellVAlignTop())
					r.Cell(func(cb *builder.CellBuilder) {
						cb.Text("B")
					})
				})
			})
		})
	})

	pdfBytes, err := b.Build()
	require.NoError(t, err)
	assertValidPDF(t, pdfBytes)
}

// TestRowPadding covers the RowPadding row option.
func TestRowPadding(t *testing.T) {
	b := builder.NewBuilder()
	b.Page(func(p *builder.PageBuilder) {
		p.Content(func(c *builder.Container) {
			c.Row(func(r *builder.RowBuilder) {
				r.Col(12, func(cb *builder.ColBuilder) {
					cb.Text("padded")
				})
			}, builder.RowPadding(builder.Pt(5)))
		})
	})

	pdfBytes, err := b.Build()
	require.NoError(t, err)
	assertValidPDF(t, pdfBytes)
}

// TestBuilder_WriteTo covers the WriteTo(io.Writer) method.
func TestBuilder_WriteTo(t *testing.T) {
	b := builder.NewBuilder()
	b.Page(func(p *builder.PageBuilder) {
		p.Content(func(c *builder.Container) {
			c.Text("WriteTo test")
		})
	})

	var buf writerBuffer
	n, err := b.WriteTo(&buf)
	require.NoError(t, err)
	assert.Greater(t, n, int64(0))
	assert.True(t, len(buf.data) > 4)
	assert.Equal(t, "%PDF", string(buf.data[:4]))
}

// writerBuffer is a simple io.Writer for tests.
type writerBuffer struct {
	data []byte
}

func (w *writerBuffer) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}
