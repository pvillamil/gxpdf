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

			// --- KPI Table using the real Table API ---
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
				// 5/12 metric, 3/12 current, 4/12 delta — using Fr proportions.
				t.Columns(builder.Fr(5), builder.Fr(3), builder.Fr(4))

				// Repeating header row: navy background, white text.
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

				// Body rows with zebra stripes.
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
							c.Text(r.delta,
								builder.FontSize(10),
								builder.TextColor(builder.Hex("#1B7B34")),
								builder.AlignRight(),
							)
						}, builder.CellPadding(builder.Pt(4)))
					}, rowBgForIndex(rowIdx))
				}
			})

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
					// Right: regional stats as a mini table.
					r.Col(5, func(col *builder.ColBuilder) {
						col.Table(func(t *builder.TableBuilder) {
							t.Columns(builder.Fr(4), builder.Fr(4), builder.Fr(4))

							t.Header(func(h *builder.TableRowBuilder) {
								h.Cell(func(c *builder.CellBuilder) {
									c.Text("Region", builder.Bold(), builder.FontSize(9))
								})
								h.Cell(func(c *builder.CellBuilder) {
									c.Text("Share", builder.Bold(), builder.FontSize(9), builder.AlignRight())
								})
								h.Cell(func(c *builder.CellBuilder) {
									c.Text("YoY", builder.Bold(), builder.FontSize(9), builder.AlignRight())
								})
							}, builder.TableRowBg(builder.Hex("#E8EDF4")))

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

			// --- Quarterly revenue table using the real Table API ---
			c.Text("Quarterly Revenue by Product Line (USD thousands)",
				builder.Bold(),
				builder.FontSize(11),
				builder.TextColor(builder.Navy),
			)
			c.Spacer(builder.Mm(3))

			revenueRows := []struct {
				name               string
				q1, q2, q3, q4    string
				annual             string
				isTotal            bool
			}{
				{"Enterprise SaaS", "1,840", "1,920", "2,100", "2,340", "8,200", false},
				{"SMB Platform", "1,050", "1,080", "1,140", "1,210", "4,480", false},
				{"Professional Svcs", "780", "820", "890", "940", "3,430", false},
				{"Marketplace", "530", "610", "670", "720", "2,530", false},
				{"Total", "4,200", "4,430", "4,800", "5,210", "18,640", true},
			}

			c.Table(func(t *builder.TableBuilder) {
				// 3/12 name, 2/12 each quarter, 1/12 annual
				t.Columns(
					builder.Fr(3),
					builder.Fr(2), builder.Fr(2), builder.Fr(2), builder.Fr(2),
					builder.Fr(1),
				)

				// Repeating header.
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

				// Body rows.
				for i, row := range revenueRows {
					r := row
					rowIdx := i
					if r.isTotal {
						// Total row: spans product name as a summary.
						t.Row(func(rb *builder.TableRowBuilder) {
							rb.Cell(func(c *builder.CellBuilder) {
								c.Text(r.name, builder.Bold(), builder.FontSize(10))
							}, builder.CellPadding(builder.Pt(4)))
							for _, val := range []string{r.q1, r.q2, r.q3, r.q4} {
								v := val
								rb.Cell(func(c *builder.CellBuilder) {
									c.Text(v, builder.Bold(), builder.FontSize(10), builder.AlignRight())
								}, builder.CellPadding(builder.Pt(4)))
							}
							annual := r.annual
							rb.Cell(func(c *builder.CellBuilder) {
								c.Text(annual, builder.Bold(), builder.FontSize(10), builder.AlignRight())
							}, builder.CellPadding(builder.Pt(4)))
						}, builder.TableRowBg(builder.Hex("#E8EDF4")))
					} else {
						t.Row(func(rb *builder.TableRowBuilder) {
							rb.Cell(func(c *builder.CellBuilder) {
								c.Text(r.name, builder.FontSize(10))
							}, builder.CellPadding(builder.Pt(4)))
							for _, val := range []string{r.q1, r.q2, r.q3, r.q4} {
								v := val
								rb.Cell(func(c *builder.CellBuilder) {
									c.Text(v, builder.FontSize(10), builder.AlignRight())
								}, builder.CellPadding(builder.Pt(4)))
							}
							annual := r.annual
							rb.Cell(func(c *builder.CellBuilder) {
								c.Text(annual, builder.FontSize(10), builder.AlignRight())
							}, builder.CellPadding(builder.Pt(4)))
						}, rowBgForIndex(rowIdx))
					}
				}

				// Footer: ColSpan for the "Summary" label + total annual.
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
}

// rowBgForIndex returns a TableRowBg option for zebra-striped rows.
// Even-indexed rows are white, odd-indexed rows are light gray.
func rowBgForIndex(i int) builder.TableRowOption {
	if i%2 == 1 {
		return builder.TableRowBg(builder.LightGray)
	}
	return builder.TableRowBg(builder.White)
}
