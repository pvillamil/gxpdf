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

		// Header: company name on the left, document label on the right.
		p.Header(func(h *builder.Container) {
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
			// Separator line under the header.
			h.Spacer(builder.Mm(2))
			h.Line(builder.LineColor(builder.Navy), builder.LineWidth(1.5))
			h.Spacer(builder.Mm(3))
		})

		// Footer: page number centered, confidentiality notice on right.
		p.Footer(func(f *builder.Container) {
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
		})

		// Body content.
		p.Content(func(c *builder.Container) {

			// --- Title block ---
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

			// --- KPI summary row: 3 equal columns ---
			c.Text("Key Performance Indicators",
				builder.Bold(),
				builder.FontSize(12),
				builder.TextColor(builder.Navy),
			)
			c.Spacer(builder.Mm(3))

			// Column header row with navy background.
			c.Row(func(r *builder.RowBuilder) {
				r.Col(5, func(col *builder.ColBuilder) {
					col.Text("Metric",
						builder.Bold(),
						builder.FontSize(9),
						builder.TextColor(builder.White),
					)
				})
				r.Col(3, func(col *builder.ColBuilder) {
					col.Text("Q1 2026",
						builder.Bold(),
						builder.FontSize(9),
						builder.TextColor(builder.White),
						builder.AlignCenter(),
					)
				})
				r.Col(4, func(col *builder.ColBuilder) {
					col.Text("vs Q1 2025",
						builder.Bold(),
						builder.FontSize(9),
						builder.TextColor(builder.White),
						builder.AlignRight(),
					)
				})
			}, builder.RowBg(builder.Navy), builder.RowPadding(builder.Pt(4)))

			c.Spacer(builder.Mm(1))

			// Data rows — alternating backgrounds for readability.
			kpiRows := []struct {
				metric, current, delta string
				highlight              bool
			}{
				{"Revenue", "$4.2M", "+18%", false},
				{"Gross Margin", "62.4%", "+3.1pp", true},
				{"New Customers", "1,247", "+31%", false},
				{"Churn Rate", "2.1%", "-0.4pp", true},
				{"NPS Score", "74", "+6pts", false},
			}

			tablePad := builder.RowPadding(builder.Pt(4))
			for _, row := range kpiRows {
				bg := builder.White
				if row.highlight {
					bg = builder.LightGray
				}

				r := row // capture loop variable
				c.Row(func(rb *builder.RowBuilder) {
					rb.Col(5, func(col *builder.ColBuilder) {
						col.Text(r.metric, builder.FontSize(10))
					})
					rb.Col(3, func(col *builder.ColBuilder) {
						col.Text(r.current,
							builder.FontSize(10),
							builder.AlignCenter(),
						)
					})
					rb.Col(4, func(col *builder.ColBuilder) {
						col.Text(r.delta,
							builder.FontSize(10),
							builder.TextColor(builder.Hex("#1B7B34")), // dark green
							builder.AlignRight(),
						)
					})
				}, builder.RowBg(bg), tablePad)
			}

			c.Spacer(builder.Mm(8))

			// --- Regional breakdown: asymmetric 8/4 split ---
			c.KeepTogether(func(inner *builder.Container) {
				inner.Text("Regional Breakdown",
					builder.Bold(),
					builder.FontSize(12),
					builder.TextColor(builder.Navy),
				)
				inner.Spacer(builder.Mm(3))
				inner.Row(func(r *builder.RowBuilder) {
					// Left: narrative text (6 columns).
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
					// Gap column for visual separation.
					r.Col(1, func(col *builder.ColBuilder) {})
					// Right: region stats (5 columns).
					r.Col(5, func(col *builder.ColBuilder) {
						col.Row(func(rr *builder.RowBuilder) {
							rr.Col(4, func(cc *builder.ColBuilder) {
								cc.Text("Region", builder.Bold(), builder.FontSize(9), builder.TextColor(builder.Navy))
							})
							rr.Col(4, func(cc *builder.ColBuilder) {
								cc.Text("Share", builder.Bold(), builder.FontSize(9), builder.TextColor(builder.Navy), builder.AlignRight())
							})
							rr.Col(4, func(cc *builder.ColBuilder) {
								cc.Text("YoY", builder.Bold(), builder.FontSize(9), builder.TextColor(builder.Navy), builder.AlignRight())
							})
						})
						col.Spacer(builder.Mm(2))
						col.Line(builder.LineColor(builder.LightGray), builder.LineWidth(0.5))
						col.Spacer(builder.Mm(2))
						regionData := []struct{ name, share, yoy string }{
							{"NA", "54%", "+18%"},
							{"EMEA", "27%", "+34%"},
							{"APAC", "19%", "+22%"},
						}
						for _, rd := range regionData {
							d := rd
							col.Row(func(rr *builder.RowBuilder) {
								rr.Col(4, func(cc *builder.ColBuilder) {
									cc.Text(d.name, builder.FontSize(9), builder.TextColor(builder.DarkGray))
								})
								rr.Col(4, func(cc *builder.ColBuilder) {
									cc.Text(d.share, builder.FontSize(9), builder.AlignRight())
								})
								rr.Col(4, func(cc *builder.ColBuilder) {
									cc.Text(d.yoy, builder.FontSize(9), builder.TextColor(builder.Hex("#1B7B34")), builder.AlignRight())
								})
							})
							col.Spacer(builder.Mm(1))
						}
					})
				})
			})

			c.Spacer(builder.Mm(8))

			// --- Management commentary ---
			c.KeepTogether(func(inner *builder.Container) {
				inner.Text("Management Commentary",
					builder.Bold(),
					builder.FontSize(12),
					builder.TextColor(builder.Navy),
				)
				inner.Spacer(builder.Mm(3))
				inner.Text(
					"Q1 2026 exceeded expectations across all primary metrics. "+
						"The strategic shift toward enterprise contracts, initiated in "+
						"H2 2025, is delivering measurable results with average contract "+
						"value increasing 41% to $18,400. Customer acquisition costs "+
						"decreased 12% due to improved marketing efficiency and stronger "+
						"inbound demand from our content and community programs.",
					builder.FontSize(10),
					builder.LineHeight(1.5),
					builder.AlignJustify(),
				)
				inner.Spacer(builder.Mm(3))
				inner.Text(
					"Looking ahead to Q2 2026, the pipeline stands at $9.7M with "+
						"an estimated close rate of 38%. We are on track to meet full-year "+
						"targets and will continue to invest in product-led growth initiatives.",
					builder.FontSize(10),
					builder.LineHeight(1.5),
					builder.AlignJustify(),
				)
			})

			// Explicit page break — the next section starts on page 2.
			c.PageBreak()

			// --- Page 2: Financial Detail ---

			c.Text("Financial Detail",
				builder.Bold(),
				builder.FontSize(16),
				builder.TextColor(builder.Navy),
			)
			c.Spacer(builder.Mm(2))
			c.Line(builder.LineColor(builder.Hex("#CCCCCC")), builder.LineWidth(0.5))
			c.Spacer(builder.Mm(5))

			// Quarterly revenue table with 6 columns.
			c.Text("Quarterly Revenue by Product Line (USD thousands)",
				builder.Bold(),
				builder.FontSize(11),
				builder.TextColor(builder.Navy),
			)
			c.Spacer(builder.Mm(3))

			// Table header.
			revPad := builder.RowPadding(builder.Pt(4))

			// Revenue table header: 3+2+2+2+2+1 = 12
			c.Row(func(r *builder.RowBuilder) {
				headers := []string{"Product Line", "Q1", "Q2", "Q3", "Q4", "Annual"}
				spans := []int{3, 2, 2, 2, 2, 1}
				aligns := []builder.TextOption{
					builder.AlignLeft(),
					builder.AlignRight(),
					builder.AlignRight(),
					builder.AlignRight(),
					builder.AlignRight(),
					builder.AlignRight(),
				}
				for i, h := range headers {
					hCopy := h
					aCopy := aligns[i]
					r.Col(spans[i], func(col *builder.ColBuilder) {
						col.Text(hCopy,
							builder.Bold(),
							builder.FontSize(9),
							builder.TextColor(builder.White),
							aCopy,
						)
					})
				}
			}, builder.RowBg(builder.Navy), revPad)

			// Revenue data rows.
			revenueRows := []struct {
				name   string
				q1, q2, q3, q4, annual string
				even   bool
			}{
				{"Enterprise SaaS", "1,840", "1,920", "2,100", "2,340", "8,200", false},
				{"SMB Platform", "1,050", "1,080", "1,140", "1,210", "4,480", true},
				{"Professional Svcs", "780", "820", "890", "940", "3,430", false},
				{"Marketplace", "530", "610", "670", "720", "2,530", true},
				{"Total", "4,200", "4,430", "4,800", "5,210", "18,640", false},
			}

			for _, row := range revenueRows {
				bg := builder.White
				if row.even {
					bg = builder.LightGray
				}

				isTotalRow := row.name == "Total"
				boldOpt := builder.FontSize(10)
				if isTotalRow {
					boldOpt = builder.Bold()
				}

				r := row
				c.Row(func(rb *builder.RowBuilder) {
					rb.Col(3, func(col *builder.ColBuilder) {
						col.Text(r.name, boldOpt, builder.FontSize(10))
					})
					for _, val := range []string{r.q1, r.q2, r.q3, r.q4} {
						v := val
						rb.Col(2, func(col *builder.ColBuilder) {
							col.Text(v, builder.FontSize(10), builder.AlignRight(), boldOpt)
						})
					}
					annual := r.annual
					rb.Col(1, func(col *builder.ColBuilder) {
						col.Text(annual, builder.FontSize(10), builder.AlignRight(), boldOpt)
					})
				}, builder.RowBg(bg), revPad)
			}

			c.Spacer(builder.Mm(8))

			// --- Footnotes / disclosure block ---
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
		})
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
	fmt.Println("  - 12-column grid layout (Row/Col with spans 2/4/6/8)")
	fmt.Println("  - Text styling: Bold, FontSize, TextColor, AlignCenter/Right")
	fmt.Println("  - Header/Footer zones with PageNumber placeholders")
	fmt.Println("  - Horizontal separator lines with custom colors")
	fmt.Println("  - Row background colors (Navy header, LightGray zebra stripes)")
	fmt.Println("  - Spacer for vertical rhythm")
	fmt.Println("  - KeepTogether to prevent section splits")
	fmt.Println("  - Explicit PageBreak for multi-page documents")
	fmt.Println("  - Hex() colors alongside predefined color constants")
}
