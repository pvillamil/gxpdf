// Package main demonstrates the GxPDF Builder API for declarative PDF generation.
//
// The Builder API provides a high-level, 12-column grid layout system for
// constructing professional documents. It sits on top of the layout engine
// and the creator package, connecting them through a clean declarative DSL.
//
// This example generates a two-page business report that demonstrates:
//   - Document-level configuration (page size, margins, metadata)
//   - Page header and footer zones with page numbers
//   - Horizontal separator lines with custom colors
//   - 12-column grid layout via Row/Col
//   - Text styling: Bold, FontSize, TextColor, AlignCenter, AlignRight
//   - Vertical spacing with Spacer
//   - Row background colors for visual grouping
//   - Explicit PageBreak to start a new page
//   - KeepTogether to prevent section splits across pages
//   - Real Table API with header repeat, zebra stripes, ColSpan, cell padding
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/coregx/gxpdf/builder"
)

func main() {
	// --- Document configuration ---
	// WithMargins: top, right, bottom, left
	b := builder.NewBuilder(
		builder.WithPageSize(builder.A4),
		builder.WithMargins(
			builder.Mm(18), builder.Mm(18), builder.Mm(18), builder.Mm(18),
		),
		builder.WithDefaultFontSize(10),
		builder.WithTitle("Q1 2026 Business Report"),
		builder.WithAuthor("GxPDF Builder API"),
	)

	// --- Page 1: Executive Summary ---
	b.Page(func(p *builder.PageBuilder) {
		p.Header(buildPageHeader)
		p.Footer(buildPageFooter)
		p.Content(buildPageContent)
	})

	// --- Generate the PDF ---
	outPath := "D:/projects/gopdf/tmp/builder_example.pdf"
	if err := b.BuildToFile(outPath); err != nil {
		log.Fatalf("BuildToFile failed: %v", err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		log.Fatalf("stat failed: %v", err)
	}

	fmt.Printf("PDF written to: %s\n", outPath)
	fmt.Printf("File size:      %d bytes (%.1f KB)\n", info.Size(), float64(info.Size())/1024)
	fmt.Println()
	fmt.Println("Demonstrates:")
	fmt.Println("  - Real Table API: t.Columns(), t.Header(), t.Row(), t.Footer()")
	fmt.Println("  - Navy header with white text via CellTextColor inheritance")
	fmt.Println("  - Zebra stripe rows via TableRowBg")
	fmt.Println("  - Cell padding via CellPadding()")
	fmt.Println("  - ColSpan(5) in footer row")
	fmt.Println("  - Nested table in a Col for regional breakdown")
	fmt.Println("  - Header repeat on overflow pages (multi-page table)")
	fmt.Println("  - 12-column grid layout (Row/Col)")
	fmt.Println("  - Text styling: Bold, FontSize, TextColor, AlignCenter/Right")
	fmt.Println("  - Header/Footer zones with PageNumber placeholders")
	fmt.Println("  - KeepTogether to prevent section splits")
	fmt.Println("  - Explicit PageBreak for multi-page documents")
	fmt.Println("  - RichText: mixed-style inline text with bold/italic/color spans")
}

// buildPageHeader renders the company name and document label in the header zone.
func buildPageHeader(h *builder.Container) {
	h.Row(func(r *builder.RowBuilder) {
		r.Col(7, func(c *builder.ColBuilder) {
			c.Text("ACME Corporation",
				builder.Bold(),
				builder.FontSize(13),
				builder.TextColor(builder.Navy),
			)
		})
		r.Col(5, func(c *builder.ColBuilder) {
			c.Text("Q1 2026 Business Report",
				builder.FontSize(9),
				builder.TextColor(builder.Gray),
				builder.AlignRight(),
			)
		})
	})
	h.Spacer(builder.Mm(2))
	h.Line(builder.LineColor(builder.Navy), builder.LineWidth(1.5))
	h.Spacer(builder.Mm(3))
}

// buildPageFooter renders the confidentiality notice and page number in the footer zone.
func buildPageFooter(f *builder.Container) {
	f.Line(builder.LineColor(builder.LightGray), builder.LineWidth(0.5))
	f.Spacer(builder.Mm(2))
	f.Row(func(r *builder.RowBuilder) {
		r.Col(6, func(c *builder.ColBuilder) {
			c.Text("Confidential - Internal Use Only",
				builder.FontSize(7),
				builder.TextColor(builder.Gray),
			)
		})
		r.Col(6, func(c *builder.ColBuilder) {
			c.PageNumber(
				"Page "+builder.PageNum+" of "+builder.TotalPages,
				builder.FontSize(8),
				builder.TextColor(builder.DarkGray),
				builder.AlignRight(),
			)
		})
	})
}

// buildPageContent assembles all body sections for the two-page report.
func buildPageContent(c *builder.Container) {
	buildTitleBlock(c)
	buildKPISection(c)
	c.Spacer(builder.Mm(8))
	buildRegionalSection(c)
	c.Spacer(builder.Mm(8))
	buildManagementCommentary(c)

	// Explicit page break — the next section starts on page 2.
	c.PageBreak()
	buildFinancialDetail(c)
}

// buildTitleBlock renders the executive summary title and subtitle.
func buildTitleBlock(c *builder.Container) {
	c.Text("Executive Summary",
		builder.Bold(),
		builder.FontSize(20),
		builder.TextColor(builder.Navy),
		builder.AlignCenter(),
	)
	c.Spacer(builder.Mm(2))
	c.Text("First Quarter - January through March 2026",
		builder.FontSize(10),
		builder.TextColor(builder.DarkGray),
		builder.AlignCenter(),
	)
	c.Spacer(builder.Mm(6))
	c.Line(builder.LineColor(builder.Hex("#CCCCCC")), builder.LineWidth(0.5))
	c.Spacer(builder.Mm(6))
}

// buildKPISection renders the Key Performance Indicators table.
func buildKPISection(c *builder.Container) {
	c.Text("Key Performance Indicators",
		builder.Bold(),
		builder.FontSize(12),
		builder.TextColor(builder.Navy),
	)
	c.Spacer(builder.Mm(3))

	kpiRows := []struct {
		metric, current, delta string
	}{
		{"Revenue", "$4.2M", "+18%"},
		{"Gross Margin", "62.4%", "+3.1pp"},
		{"New Customers", "1,247", "+31%"},
		{"Churn Rate", "2.1%", "-0.4pp"},
		{"NPS Score", "74", "+6pts"},
	}

	c.Table(func(t *builder.TableBuilder) {
		t.Columns(builder.Fr(5), builder.Fr(3), builder.Fr(4))
		w := builder.TextColor(builder.White)
		t.Header(func(h *builder.TableRowBuilder) {
			h.Cell(func(c *builder.CellBuilder) {
				c.Text("Metric", builder.Bold(), builder.FontSize(9), w)
			}, builder.CellPadding(builder.Pt(4)))
			h.Cell(func(c *builder.CellBuilder) {
				c.Text("Q1 2026", builder.Bold(), builder.FontSize(9), builder.AlignCenter(), w)
			}, builder.CellPadding(builder.Pt(4)))
			h.Cell(func(c *builder.CellBuilder) {
				c.Text("vs Q1 2025", builder.Bold(), builder.FontSize(9), builder.AlignRight(), w)
			}, builder.CellPadding(builder.Pt(4)))
		}, builder.TableRowBg(builder.Navy))

		for i, row := range kpiRows {
			r := row
			rowIdx := i
			t.Row(func(rb *builder.TableRowBuilder) {
				rb.Cell(func(c *builder.CellBuilder) {
					c.Text(r.metric, builder.FontSize(10))
				}, builder.CellPadding(builder.Pt(4)))
				rb.Cell(func(c *builder.CellBuilder) {
					c.Text(r.current, builder.FontSize(10), builder.AlignCenter())
				}, builder.CellPadding(builder.Pt(4)))
				rb.Cell(func(c *builder.CellBuilder) {
					c.Text(r.delta, builder.FontSize(10),
						builder.TextColor(builder.Hex("#1B7B34")),
						builder.AlignRight(),
					)
				}, builder.CellPadding(builder.Pt(4)))
			}, rowBgForIndex(rowIdx))
		}
	})
}

// buildRegionalSection renders the regional breakdown row with narrative text
// and a nested mini-table.
func buildRegionalSection(c *builder.Container) {
	c.KeepTogether(func(inner *builder.Container) {
		inner.Text("Regional Breakdown",
			builder.Bold(),
			builder.FontSize(12),
			builder.TextColor(builder.Navy),
		)
		inner.Spacer(builder.Mm(3))
		inner.Row(func(r *builder.RowBuilder) {
			r.Col(6, func(col *builder.ColBuilder) {
				col.Text(
					"North America remains our strongest market, contributing "+
						"54% of total revenue. EMEA showed the highest growth "+
						"rate at 34% year-over-year, driven by new enterprise "+
						"contracts in Germany and France. APAC grew 22% despite "+
						"macroeconomic headwinds, with Japan and Australia "+
						"leading performance.",
					builder.FontSize(10),
					builder.LineHeight(1.5),
					builder.AlignJustify(),
				)
			})
			r.Col(1, func(col *builder.ColBuilder) {})
			r.Col(5, func(col *builder.ColBuilder) {
				buildRegionalTable(col)
			})
		})
	})
}

// buildRegionalTable renders the regional stats mini-table inside a column.
func buildRegionalTable(col *builder.ColBuilder) {
	col.Table(func(t *builder.TableBuilder) {
		t.Columns(builder.Fr(4), builder.Fr(4), builder.Fr(4))
		t.Header(func(h *builder.TableRowBuilder) {
			h.Cell(func(c *builder.CellBuilder) {
				c.Text("Region", builder.Bold(), builder.FontSize(9), builder.TextColor(builder.White))
			}, builder.CellPadding(builder.Pt(8)))
			h.Cell(func(c *builder.CellBuilder) {
				c.Text("Share", builder.Bold(), builder.FontSize(9), builder.TextColor(builder.White), builder.AlignRight())
			}, builder.CellPadding(builder.Pt(8)))
			h.Cell(func(c *builder.CellBuilder) {
				c.Text("YoY", builder.Bold(), builder.FontSize(9), builder.TextColor(builder.White), builder.AlignRight())
			}, builder.CellPadding(builder.Pt(8)))
		}, builder.TableRowBg(builder.Navy))

		regionData := []struct{ name, share, yoy string }{
			{"NA", "54%", "+18%"},
			{"EMEA", "27%", "+34%"},
			{"APAC", "19%", "+22%"},
		}
		for _, rd := range regionData {
			d := rd
			t.Row(func(r *builder.TableRowBuilder) {
				r.Cell(func(c *builder.CellBuilder) {
					c.Text(d.name, builder.FontSize(9), builder.TextColor(builder.DarkGray))
				}, builder.CellPadding(builder.Pt(3)))
				r.Cell(func(c *builder.CellBuilder) {
					c.Text(d.share, builder.FontSize(9), builder.AlignRight())
				}, builder.CellPadding(builder.Pt(3)))
				r.Cell(func(c *builder.CellBuilder) {
					c.Text(d.yoy, builder.FontSize(9),
						builder.TextColor(builder.Hex("#1B7B34")),
						builder.AlignRight())
				}, builder.CellPadding(builder.Pt(3)))
			})
		}
	})
}

// buildManagementCommentary renders the management commentary section with
// RichText for inline bold/italic/color mixing.
func buildManagementCommentary(c *builder.Container) {
	c.KeepTogether(func(inner *builder.Container) {
		inner.Text("Management Commentary",
			builder.Bold(),
			builder.FontSize(12),
			builder.TextColor(builder.Navy),
		)
		inner.Spacer(builder.Mm(3))

		inner.RichText(func(rt *builder.RichTextBuilder) {
			rt.Span("Q1 2026 exceeded expectations across all primary metrics. " +
				"The strategic shift toward enterprise contracts, initiated in " +
				"H2 2025, is delivering measurable results with average contract " +
				"value increasing ")
			rt.Span("41%", builder.Bold(), builder.TextColor(builder.Hex("#1B7B34")))
			rt.Span(" to ")
			rt.Span("$18,400", builder.Bold(), builder.TextColor(builder.Hex("#1B7B34")))
			rt.Span(". Customer acquisition costs decreased ")
			rt.Span("12%", builder.Bold(), builder.TextColor(builder.Hex("#1B7B34")))
			rt.Span(" due to improved marketing efficiency and stronger " +
				"inbound demand from our content and community programs.")
		}, builder.FontSize(10), builder.LineHeight(1.5), builder.AlignJustify())

		inner.Spacer(builder.Mm(3))

		inner.RichText(func(rt *builder.RichTextBuilder) {
			rt.Span("Looking ahead to Q2 2026, the pipeline stands at ")
			rt.Span("$9.7M", builder.Bold())
			rt.Span(" with an estimated close rate of ")
			rt.Span("38%", builder.Bold())
			rt.Span(". We are on track to meet full-year targets and will " +
				"continue to invest in ")
			rt.Span("product-led growth", builder.Italic())
			rt.Span(" initiatives.")
		}, builder.FontSize(10), builder.LineHeight(1.5), builder.AlignJustify())
	})
}

// buildFinancialDetail renders page 2: the quarterly revenue table and footnotes.
func buildFinancialDetail(c *builder.Container) {
	c.Text("Financial Detail",
		builder.Bold(),
		builder.FontSize(16),
		builder.TextColor(builder.Navy),
	)
	c.Spacer(builder.Mm(2))
	c.Line(builder.LineColor(builder.Hex("#CCCCCC")), builder.LineWidth(0.5))
	c.Spacer(builder.Mm(5))

	buildRevenueTable(c)

	c.Spacer(builder.Mm(8))
	buildFootnotes(c)
}

// buildRevenueTable renders the quarterly revenue by product line table.
func buildRevenueTable(c *builder.Container) {
	c.Text("Quarterly Revenue by Product Line (USD thousands)",
		builder.Bold(),
		builder.FontSize(11),
		builder.TextColor(builder.Navy),
	)
	c.Spacer(builder.Mm(3))

	type revenueRow struct {
		name           string
		q1, q2, q3, q4 string
		annual         string
		isTotal        bool
	}
	revenueRows := []revenueRow{
		{"Enterprise SaaS", "1,840", "1,920", "2,100", "2,340", "8,200", false},
		{"SMB Platform", "1,050", "1,080", "1,140", "1,210", "4,480", false},
		{"Professional Svcs", "780", "820", "890", "940", "3,430", false},
		{"Marketplace", "530", "610", "670", "720", "2,530", false},
		{"Total", "4,200", "4,430", "4,800", "5,210", "18,640", true},
	}

	c.Table(func(t *builder.TableBuilder) {
		t.Columns(
			builder.Fr(3),
			builder.Fr(2), builder.Fr(2), builder.Fr(2), builder.Fr(2),
			builder.Fr(1),
		)

		wh := builder.TextColor(builder.White)
		t.Header(func(h *builder.TableRowBuilder) {
			headers := []string{"Product Line", "Q1", "Q2", "Q3", "Q4", "Annual"}
			aligns := []builder.TextOption{
				builder.AlignLeft(),
				builder.AlignRight(), builder.AlignRight(),
				builder.AlignRight(), builder.AlignRight(),
				builder.AlignRight(),
			}
			for i, hdr := range headers {
				hText := hdr
				hAlign := aligns[i]
				h.Cell(func(c *builder.CellBuilder) {
					c.Text(hText, builder.Bold(), builder.FontSize(9), hAlign, wh)
				}, builder.CellPadding(builder.Pt(4)))
			}
		}, builder.TableRowBg(builder.Navy))

		for i, row := range revenueRows {
			r := row
			rowIdx := i
			addRevenueRow(t, r.name, r.q1, r.q2, r.q3, r.q4, r.annual, r.isTotal, rowIdx)
		}

		t.Footer(func(f *builder.TableRowBuilder) {
			f.Cell(func(c *builder.CellBuilder) {
				c.Text("FY 2026 projection: $21.5M",
					builder.Bold(),
					builder.FontSize(9),
					builder.TextColor(builder.DarkGray),
				)
			}, builder.ColSpan(5), builder.CellPadding(builder.Pt(4)))
			f.Cell(func(c *builder.CellBuilder) {
				c.Text("21,500",
					builder.Bold(),
					builder.FontSize(9),
					builder.AlignRight(),
				)
			}, builder.CellPadding(builder.Pt(4)))
		}, builder.TableRowBg(builder.LightGray))
	})
}

// addRevenueRow appends one body row to the revenue table.
func addRevenueRow(t *builder.TableBuilder, name, q1, q2, q3, q4, annual string, isTotal bool, rowIdx int) {
	quarters := []string{q1, q2, q3, q4}
	if isTotal {
		t.Row(func(rb *builder.TableRowBuilder) {
			rb.Cell(func(c *builder.CellBuilder) {
				c.Text(name, builder.Bold(), builder.FontSize(10))
			}, builder.CellPadding(builder.Pt(4)))
			for _, val := range quarters {
				v := val
				rb.Cell(func(c *builder.CellBuilder) {
					c.Text(v, builder.Bold(), builder.FontSize(10), builder.AlignRight())
				}, builder.CellPadding(builder.Pt(4)))
			}
			ann := annual
			rb.Cell(func(c *builder.CellBuilder) {
				c.Text(ann, builder.Bold(), builder.FontSize(10), builder.AlignRight())
			}, builder.CellPadding(builder.Pt(4)))
		}, builder.TableRowBg(builder.Hex("#E8EDF4")))
	} else {
		t.Row(func(rb *builder.TableRowBuilder) {
			rb.Cell(func(c *builder.CellBuilder) {
				c.Text(name, builder.FontSize(10))
			}, builder.CellPadding(builder.Pt(4)))
			for _, val := range quarters {
				v := val
				rb.Cell(func(c *builder.CellBuilder) {
					c.Text(v, builder.FontSize(10), builder.AlignRight())
				}, builder.CellPadding(builder.Pt(4)))
			}
			ann := annual
			rb.Cell(func(c *builder.CellBuilder) {
				c.Text(ann, builder.FontSize(10), builder.AlignRight())
			}, builder.CellPadding(builder.Pt(4)))
		}, rowBgForIndex(rowIdx))
	}
}

// buildFootnotes renders the disclosure footnotes at the bottom of page 2.
func buildFootnotes(c *builder.Container) {
	c.Line(builder.LineColor(builder.LightGray), builder.LineWidth(0.5))
	c.Spacer(builder.Mm(3))
	c.Text("Notes",
		builder.Bold(),
		builder.FontSize(9),
		builder.TextColor(builder.DarkGray),
	)
	c.Spacer(builder.Mm(2))
	c.Text(
		"All figures are unaudited and presented in USD thousands unless otherwise noted. "+
			"Year-over-year comparisons are based on restated Q1 2025 figures. "+
			"This document is intended for internal use only and may not be distributed "+
			"without prior written approval from the Chief Financial Officer.",
		builder.FontSize(8),
		builder.TextColor(builder.Gray),
		builder.LineHeight(1.4),
		builder.AlignJustify(),
	)
}

// rowBgForIndex returns a TableRowBg option for zebra-striped rows.
// Even-indexed rows are white, odd-indexed rows are light gray.
func rowBgForIndex(i int) builder.TableRowOption {
	if i%2 == 1 {
		return builder.TableRowBg(builder.LightGray)
	}
	return builder.TableRowBg(builder.White)
}
