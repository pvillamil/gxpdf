package writer

import (
	"strings"
	"testing"

	"github.com/coregx/gxpdf/internal/encoding"
)

// TestContentStreamWriter_TextOperators tests text-related operators.
func TestContentStreamWriter_TextOperators(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*ContentStreamWriter)
		expected string
	}{
		{
			name: "BeginText and EndText",
			build: func(csw *ContentStreamWriter) {
				csw.BeginText()
				csw.EndText()
			},
			expected: "BT\nET\n",
		},
		{
			name: "SetFont",
			build: func(csw *ContentStreamWriter) {
				csw.SetFont("F1", 12.0)
			},
			expected: "/F1 12.00 Tf\n",
		},
		{
			name: "MoveTextPosition",
			build: func(csw *ContentStreamWriter) {
				csw.MoveTextPosition(100.0, 700.0)
			},
			expected: "100.00 700.00 Td\n",
		},
		{
			name: "MoveTextPositionSetLeading",
			build: func(csw *ContentStreamWriter) {
				csw.MoveTextPositionSetLeading(0.0, -14.0)
			},
			expected: "0.00 -14.00 TD\n",
		},
		{
			name: "SetTextMatrix",
			build: func(csw *ContentStreamWriter) {
				csw.SetTextMatrix(1.0, 0.0, 0.0, 1.0, 50.0, 750.0)
			},
			expected: "1.00 0.00 0.00 1.00 50.00 750.00 Tm\n",
		},
		{
			name: "ShowText",
			build: func(csw *ContentStreamWriter) {
				csw.ShowText("Hello World")
			},
			expected: "(Hello World) Tj\n",
		},
		{
			name: "ShowText with special characters",
			build: func(csw *ContentStreamWriter) {
				csw.ShowText("Text with (parentheses) and \\backslash")
			},
			expected: "(Text with \\(parentheses\\) and \\\\backslash) Tj\n",
		},
		{
			name: "ShowTextNextLine",
			build: func(csw *ContentStreamWriter) {
				csw.ShowTextNextLine("Next line")
			},
			expected: "(Next line) '\n",
		},
		{
			name: "SetLeading",
			build: func(csw *ContentStreamWriter) {
				csw.SetLeading(14.0)
			},
			expected: "14.00 TL\n",
		},
		{
			name: "MoveToNextLine",
			build: func(csw *ContentStreamWriter) {
				csw.MoveToNextLine()
			},
			expected: "T*\n",
		},
		{
			name: "Complete text example",
			build: func(csw *ContentStreamWriter) {
				csw.BeginText()
				csw.SetFont("F1", 12.0)
				csw.MoveTextPosition(100.0, 700.0)
				csw.ShowText("Hello World")
				csw.EndText()
			},
			expected: "BT\n/F1 12.00 Tf\n100.00 700.00 Td\n(Hello World) Tj\nET\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csw := NewContentStreamWriter()
			tt.build(csw)

			got := csw.String()
			if got != tt.expected {
				t.Errorf("ContentStreamWriter output mismatch\nGot:\n%s\nExpected:\n%s", got, tt.expected)
			}
		})
	}
}

// TestContentStreamWriter_GraphicsOperators tests graphics path operators.
func TestContentStreamWriter_GraphicsOperators(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*ContentStreamWriter)
		expected string
	}{
		{
			name: "MoveTo",
			build: func(csw *ContentStreamWriter) {
				csw.MoveTo(100.0, 100.0)
			},
			expected: "100.00 100.00 m\n",
		},
		{
			name: "LineTo",
			build: func(csw *ContentStreamWriter) {
				csw.LineTo(200.0, 200.0)
			},
			expected: "200.00 200.00 l\n",
		},
		{
			name: "CurveTo",
			build: func(csw *ContentStreamWriter) {
				csw.CurveTo(100.0, 150.0, 150.0, 200.0, 200.0, 200.0)
			},
			expected: "100.00 150.00 150.00 200.00 200.00 200.00 c\n",
		},
		{
			name: "Rectangle",
			build: func(csw *ContentStreamWriter) {
				csw.Rectangle(50.0, 50.0, 200.0, 100.0)
			},
			expected: "50.00 50.00 200.00 100.00 re\n",
		},
		{
			name: "ClosePath",
			build: func(csw *ContentStreamWriter) {
				csw.ClosePath()
			},
			expected: "h\n",
		},
		{
			name: "Stroke",
			build: func(csw *ContentStreamWriter) {
				csw.Stroke()
			},
			expected: "S\n",
		},
		{
			name: "CloseAndStroke",
			build: func(csw *ContentStreamWriter) {
				csw.CloseAndStroke()
			},
			expected: "s\n",
		},
		{
			name: "Fill",
			build: func(csw *ContentStreamWriter) {
				csw.Fill()
			},
			expected: "f\n",
		},
		{
			name: "FillEvenOdd",
			build: func(csw *ContentStreamWriter) {
				csw.FillEvenOdd()
			},
			expected: "f*\n",
		},
		{
			name: "FillAndStroke",
			build: func(csw *ContentStreamWriter) {
				csw.FillAndStroke()
			},
			expected: "B\n",
		},
		{
			name: "FillAndStrokeEvenOdd",
			build: func(csw *ContentStreamWriter) {
				csw.FillAndStrokeEvenOdd()
			},
			expected: "B*\n",
		},
		{
			name: "EndPath",
			build: func(csw *ContentStreamWriter) {
				csw.EndPath()
			},
			expected: "n\n",
		},
		{
			name: "Complete line example",
			build: func(csw *ContentStreamWriter) {
				csw.MoveTo(100.0, 100.0)
				csw.LineTo(200.0, 200.0)
				csw.Stroke()
			},
			expected: "100.00 100.00 m\n200.00 200.00 l\nS\n",
		},
		{
			name: "Complete rectangle example",
			build: func(csw *ContentStreamWriter) {
				csw.Rectangle(50.0, 50.0, 200.0, 100.0)
				csw.Fill()
			},
			expected: "50.00 50.00 200.00 100.00 re\nf\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csw := NewContentStreamWriter()
			tt.build(csw)

			got := csw.String()
			if got != tt.expected {
				t.Errorf("ContentStreamWriter output mismatch\nGot:\n%s\nExpected:\n%s", got, tt.expected)
			}
		})
	}
}

// TestContentStreamWriter_GraphicsStateOperators tests graphics state operators.
func TestContentStreamWriter_GraphicsStateOperators(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*ContentStreamWriter)
		expected string
	}{
		{
			name: "SaveState and RestoreState",
			build: func(csw *ContentStreamWriter) {
				csw.SaveState()
				csw.RestoreState()
			},
			expected: "q\nQ\n",
		},
		{
			name: "ConcatMatrix",
			build: func(csw *ContentStreamWriter) {
				csw.ConcatMatrix(2.0, 0.0, 0.0, 2.0, 0.0, 0.0)
			},
			expected: "2.00 0.00 0.00 2.00 0.00 0.00 cm\n",
		},
		{
			name: "SetLineWidth",
			build: func(csw *ContentStreamWriter) {
				csw.SetLineWidth(2.5)
			},
			expected: "2.50 w\n",
		},
		{
			name: "SetLineCap",
			build: func(csw *ContentStreamWriter) {
				csw.SetLineCap(1)
			},
			expected: "1 J\n",
		},
		{
			name: "SetLineJoin",
			build: func(csw *ContentStreamWriter) {
				csw.SetLineJoin(2)
			},
			expected: "2 j\n",
		},
		{
			name: "SetMiterLimit",
			build: func(csw *ContentStreamWriter) {
				csw.SetMiterLimit(10.0)
			},
			expected: "10.00 M\n",
		},
		{
			name: "SetDashPattern empty",
			build: func(csw *ContentStreamWriter) {
				csw.SetDashPattern([]float64{}, 0.0)
			},
			expected: "[] 0.00 d\n",
		},
		{
			name: "SetDashPattern dashed",
			build: func(csw *ContentStreamWriter) {
				csw.SetDashPattern([]float64{3.0, 1.0}, 0.0)
			},
			expected: "[3.00 1.00] 0.00 d\n",
		},
		{
			name: "SetStrokeColorRGB",
			build: func(csw *ContentStreamWriter) {
				csw.SetStrokeColorRGB(1.0, 0.0, 0.0)
			},
			expected: "1.00 0.00 0.00 RG\n",
		},
		{
			name: "SetFillColorRGB",
			build: func(csw *ContentStreamWriter) {
				csw.SetFillColorRGB(0.0, 1.0, 0.0)
			},
			expected: "0.00 1.00 0.00 rg\n",
		},
		{
			name: "SetStrokeColorGray",
			build: func(csw *ContentStreamWriter) {
				csw.SetStrokeColorGray(0.5)
			},
			expected: "0.50 G\n",
		},
		{
			name: "SetFillColorGray",
			build: func(csw *ContentStreamWriter) {
				csw.SetFillColorGray(0.75)
			},
			expected: "0.75 g\n",
		},
		{
			name: "SetStrokeColorCMYK",
			build: func(csw *ContentStreamWriter) {
				csw.SetStrokeColorCMYK(0.0, 1.0, 1.0, 0.0)
			},
			expected: "0.00 1.00 1.00 0.00 K\n",
		},
		{
			name: "SetFillColorCMYK",
			build: func(csw *ContentStreamWriter) {
				csw.SetFillColorCMYK(1.0, 0.0, 0.0, 0.0)
			},
			expected: "1.00 0.00 0.00 0.00 k\n",
		},
		{
			name: "Complete state example",
			build: func(csw *ContentStreamWriter) {
				csw.SaveState()
				csw.SetStrokeColorRGB(1.0, 0.0, 0.0)
				csw.SetLineWidth(2.0)
				csw.Rectangle(100.0, 100.0, 50.0, 50.0)
				csw.Stroke()
				csw.RestoreState()
			},
			expected: "q\n1.00 0.00 0.00 RG\n2.00 w\n100.00 100.00 50.00 50.00 re\nS\nQ\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csw := NewContentStreamWriter()
			tt.build(csw)

			got := csw.String()
			if got != tt.expected {
				t.Errorf("ContentStreamWriter output mismatch\nGot:\n%s\nExpected:\n%s", got, tt.expected)
			}
		})
	}
}

// TestContentStreamWriter_ApplyShading tests the sh operator.
func TestContentStreamWriter_ApplyShading(t *testing.T) {
	csw := NewContentStreamWriter()
	csw.ApplyShading("Sh1")

	got := csw.String()
	expected := "/Sh1 sh\n"
	if got != expected {
		t.Errorf("ApplyShading output = %q, want %q", got, expected)
	}
}

// TestContentStreamWriter_ClipAndShade tests the clip+shade pattern for gradient fills.
func TestContentStreamWriter_ClipAndShade(t *testing.T) {
	csw := NewContentStreamWriter()

	csw.SaveState()
	csw.Rectangle(50, 620, 200, 60)
	csw.Clip()
	csw.EndPath()
	csw.ApplyShading("Sh1")
	csw.RestoreState()

	got := csw.String()
	for _, op := range []string{"q\n", "50.00 620.00 200.00 60.00 re\n", "W\n", "n\n", "/Sh1 sh\n", "Q\n"} {
		if !strings.Contains(got, op) {
			t.Errorf("Clip+shade pattern missing %q\nGot:\n%s", op, got)
		}
	}
}

// TestContentStreamWriter_CombinedOperations tests complex combined operations.
func TestContentStreamWriter_CombinedOperations(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*ContentStreamWriter)
		contains []string // Verify key parts are present
	}{
		{
			name: "Text and graphics combined",
			build: func(csw *ContentStreamWriter) {
				// Draw rectangle
				csw.SaveState()
				csw.SetStrokeColorRGB(0.0, 0.0, 1.0)
				csw.Rectangle(50.0, 50.0, 200.0, 100.0)
				csw.Stroke()
				csw.RestoreState()

				// Add text
				csw.BeginText()
				csw.SetFont("Helvetica", 14.0)
				csw.MoveTextPosition(60.0, 90.0)
				csw.ShowText("Inside Box")
				csw.EndText()
			},
			contains: []string{
				"q\n",
				"0.00 0.00 1.00 RG\n",
				"50.00 50.00 200.00 100.00 re\n",
				"S\n",
				"Q\n",
				"BT\n",
				"/Helvetica 14.00 Tf\n",
				"60.00 90.00 Td\n",
				"(Inside Box) Tj\n",
				"ET\n",
			},
		},
		{
			name: "Multiple text lines",
			build: func(csw *ContentStreamWriter) {
				csw.BeginText()
				csw.SetFont("Times-Roman", 12.0)
				csw.SetLeading(14.0)
				csw.MoveTextPosition(50.0, 750.0)
				csw.ShowText("Line 1")
				csw.MoveToNextLine()
				csw.ShowText("Line 2")
				csw.MoveToNextLine()
				csw.ShowText("Line 3")
				csw.EndText()
			},
			contains: []string{
				"BT\n",
				"/Times-Roman 12.00 Tf\n",
				"14.00 TL\n",
				"50.00 750.00 Td\n",
				"(Line 1) Tj\n",
				"T*\n",
				"(Line 2) Tj\n",
				"(Line 3) Tj\n",
				"ET\n",
			},
		},
		{
			name: "Complex path",
			build: func(csw *ContentStreamWriter) {
				csw.SaveState()
				csw.SetFillColorRGB(0.8, 0.8, 1.0)
				csw.SetStrokeColorRGB(0.0, 0.0, 0.5)
				csw.SetLineWidth(1.5)

				csw.MoveTo(100.0, 100.0)
				csw.LineTo(200.0, 100.0)
				csw.LineTo(200.0, 200.0)
				csw.LineTo(100.0, 200.0)
				csw.ClosePath()
				csw.FillAndStroke()

				csw.RestoreState()
			},
			contains: []string{
				"q\n",
				"0.80 0.80 1.00 rg\n",
				"0.00 0.00 0.50 RG\n",
				"1.50 w\n",
				"100.00 100.00 m\n",
				"200.00 100.00 l\n",
				"200.00 200.00 l\n",
				"100.00 200.00 l\n",
				"h\n",
				"B\n",
				"Q\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csw := NewContentStreamWriter()
			tt.build(csw)

			got := csw.String()
			for _, expected := range tt.contains {
				if !strings.Contains(got, expected) {
					t.Errorf("ContentStreamWriter output missing expected substring:\n%s\n\nFull output:\n%s", expected, got)
				}
			}
		})
	}
}

// TestContentStreamWriter_Utilities tests utility methods.
func TestContentStreamWriter_Utilities(t *testing.T) {
	t.Run("Bytes returns correct data", func(t *testing.T) {
		csw := NewContentStreamWriter()
		csw.BeginText()
		csw.EndText()

		bytes := csw.Bytes()
		expected := "BT\nET\n"

		if string(bytes) != expected {
			t.Errorf("Bytes() = %q, want %q", string(bytes), expected)
		}
	})

	t.Run("Len returns correct length", func(t *testing.T) {
		csw := NewContentStreamWriter()
		csw.ShowText("Hello")

		if csw.Len() == 0 {
			t.Error("Len() should be > 0 after writing content")
		}

		expected := len("(Hello) Tj\n")
		if csw.Len() != expected {
			t.Errorf("Len() = %d, want %d", csw.Len(), expected)
		}
	})

	t.Run("Reset clears buffer", func(t *testing.T) {
		csw := NewContentStreamWriter()
		csw.ShowText("Hello")

		if csw.Len() == 0 {
			t.Error("Len() should be > 0 before reset")
		}

		csw.Reset()

		if csw.Len() != 0 {
			t.Errorf("Len() after Reset() = %d, want 0", csw.Len())
		}
	})

	t.Run("String returns same as Bytes", func(t *testing.T) {
		csw := NewContentStreamWriter()
		csw.MoveTo(100.0, 200.0)

		str := csw.String()
		bytes := string(csw.Bytes())

		if str != bytes {
			t.Errorf("String() != string(Bytes())\nString(): %q\nBytes():  %q", str, bytes)
		}
	})
}

// TestContentStreamWriter_Compression tests Flate compression.
func TestContentStreamWriter_Compression(t *testing.T) {
	csw := NewContentStreamWriter()
	csw.BeginText()
	csw.SetFont("Helvetica", 12.0)
	csw.MoveTextPosition(100.0, 700.0)
	csw.ShowText("Hello World")
	csw.EndText()

	// Get uncompressed content.
	uncompressed := csw.Bytes()

	// Compress.
	compressed, err := csw.Compress()
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	// Compressed should be smaller (usually).
	// For very small content, it might be larger due to zlib headers.
	if len(compressed) == 0 {
		t.Error("Compressed content is empty")
	}

	// Verify we can decompress back.
	decoder := encoding.NewFlateDecoder()
	decompressed, err := decoder.Decode(compressed)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if string(decompressed) != string(uncompressed) {
		t.Errorf("Decompressed content doesn't match original\nOriginal:     %q\nDecompressed: %q",
			string(uncompressed), string(decompressed))
	}
}

// TestEscapePDFString tests PDF string escaping integration.
// Comprehensive tests are in string_escape_test.go.
func TestEscapePDFString_Integration(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello", "Hello"},
		{"Text with (parentheses)", "Text with \\(parentheses\\)"},
		{"Backslash: \\", "Backslash: \\\\"},
		{"Both \\ and ()", "Both \\\\ and \\(\\)"},
		{"Newline\nHere", "Newline\\nHere"},
		{"Return\rHere", "Return\\rHere"},
		{"Tab\tHere", "Tab\\tHere"},
		{"All special: \\ ( ) \n \r \t", "All special: \\\\ \\( \\) \\n \\r \\t"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := EscapePDFString(tt.input)
			if got != tt.expected {
				t.Errorf("EscapePDFString(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestContentStreamWriter_EmptyStream tests empty stream behavior.
func TestContentStreamWriter_EmptyStream(t *testing.T) {
	csw := NewContentStreamWriter()

	if csw.Len() != 0 {
		t.Errorf("Empty stream Len() = %d, want 0", csw.Len())
	}

	if csw.String() != "" {
		t.Errorf("Empty stream String() = %q, want empty string", csw.String())
	}

	bytes := csw.Bytes()
	if len(bytes) != 0 {
		t.Errorf("Empty stream Bytes() length = %d, want 0", len(bytes))
	}
}

// TestContentStreamWriter_LargeContent tests handling of large content streams.
func TestContentStreamWriter_LargeContent(t *testing.T) {
	csw := NewContentStreamWriter()

	// Generate large content stream.
	csw.BeginText()
	csw.SetFont("Courier", 10.0)
	csw.SetLeading(12.0)

	for i := 0; i < 1000; i++ {
		csw.MoveTextPosition(50.0, float64(750-i*12))
		csw.ShowText("This is line number " + string(rune(i)))
	}

	csw.EndText()

	// Verify content was accumulated.
	if csw.Len() == 0 {
		t.Error("Large content stream is empty")
	}

	// Verify compression works on large content.
	compressed, err := csw.Compress()
	if err != nil {
		t.Fatalf("Compress() error on large content = %v", err)
	}

	// Compressed should be significantly smaller.
	compressionRatio := float64(len(compressed)) / float64(csw.Len())
	if compressionRatio > 0.9 {
		t.Logf("Warning: Compression ratio %.2f is not very good (expected < 0.9)", compressionRatio)
	}
}

// BenchmarkContentStreamWriter_SimpleText benchmarks simple text operations.
func BenchmarkContentStreamWriter_SimpleText(b *testing.B) {
	for i := 0; i < b.N; i++ {
		csw := NewContentStreamWriter()
		csw.BeginText()
		csw.SetFont("Helvetica", 12.0)
		csw.MoveTextPosition(100.0, 700.0)
		csw.ShowText("Hello World")
		csw.EndText()
		_ = csw.Bytes()
	}
}

// BenchmarkContentStreamWriter_Graphics benchmarks graphics operations.
func BenchmarkContentStreamWriter_Graphics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		csw := NewContentStreamWriter()
		csw.SaveState()
		csw.SetStrokeColorRGB(1.0, 0.0, 0.0)
		csw.SetLineWidth(2.0)
		csw.Rectangle(100.0, 100.0, 200.0, 100.0)
		csw.Stroke()
		csw.RestoreState()
		_ = csw.Bytes()
	}
}

// BenchmarkContentStreamWriter_Compression benchmarks compression.
func BenchmarkContentStreamWriter_Compression(b *testing.B) {
	// Create content once.
	csw := NewContentStreamWriter()
	csw.BeginText()
	csw.SetFont("Helvetica", 12.0)
	for i := 0; i < 100; i++ {
		csw.MoveTextPosition(50.0, float64(750-i*12))
		csw.ShowText("This is a line of text to compress")
	}
	csw.EndText()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := csw.Compress()
		if err != nil {
			b.Fatal(err)
		}
	}
}
