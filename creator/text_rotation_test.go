package creator

import (
	"bytes"
	"testing"

	"github.com/coregx/gxpdf/internal/writer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddTextRotated_ZeroRotation verifies that 0° rotation behaves identically
// to a normal AddText call — the TextOperation is stored with Rotation == 0.
func TestAddTextRotated_ZeroRotation(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextRotated("Hello", 100, 700, Helvetica, 12, 0)
	require.NoError(t, err)

	ops := page.TextOperations()
	require.Len(t, ops, 1)
	assert.Equal(t, 0.0, ops[0].Rotation)
	assert.Equal(t, "Hello", ops[0].Text)
	assert.Equal(t, 100.0, ops[0].X)
	assert.Equal(t, 700.0, ops[0].Y)
}

// TestAddTextRotated_90Degrees verifies that 90° rotation is stored correctly.
func TestAddTextRotated_90Degrees(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextRotated("Sideways", 50, 400, Helvetica, 14, 90)
	require.NoError(t, err)

	ops := page.TextOperations()
	require.Len(t, ops, 1)
	assert.Equal(t, 90.0, ops[0].Rotation)
}

// TestAddTextRotated_NegativeRotation verifies that negative (clockwise) rotation works.
func TestAddTextRotated_NegativeRotation(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextRotated("Tilted", 100, 500, Helvetica, 12, -45)
	require.NoError(t, err)

	ops := page.TextOperations()
	require.Len(t, ops, 1)
	assert.Equal(t, -45.0, ops[0].Rotation)
}

// TestAddTextColorRotated_StoresColorCorrectly verifies color is stored alongside rotation.
func TestAddTextColorRotated_StoresColorCorrectly(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextColorRotated("Draft", 300, 400, HelveticaBold, 48, Red, 45)
	require.NoError(t, err)

	ops := page.TextOperations()
	require.Len(t, ops, 1)
	op := ops[0]
	assert.Equal(t, 45.0, op.Rotation)
	assert.Equal(t, 1.0, op.Color.R)
	assert.Equal(t, 0.0, op.Color.G)
	assert.Equal(t, 0.0, op.Color.B)
}

// TestAddTextRotated_ValidationErrors verifies that validation (font size > 0, color range)
// still applies for rotated text methods.
func TestAddTextRotated_ValidationErrors(t *testing.T) {
	t.Run("zero font size rejected", func(t *testing.T) {
		c := New()
		page, err := c.NewPage()
		require.NoError(t, err)

		err = page.AddTextRotated("text", 100, 700, Helvetica, 0, 45)
		assert.Error(t, err)
	})

	t.Run("negative font size rejected", func(t *testing.T) {
		c := New()
		page, err := c.NewPage()
		require.NoError(t, err)

		err = page.AddTextRotated("text", 100, 700, Helvetica, -1, 45)
		assert.Error(t, err)
	})

	t.Run("color component > 1 rejected", func(t *testing.T) {
		c := New()
		page, err := c.NewPage()
		require.NoError(t, err)

		err = page.AddTextColorRotated("text", 100, 700, Helvetica, 12, Color{R: 2.0}, 45)
		assert.Error(t, err)
	})

	t.Run("negative color component rejected", func(t *testing.T) {
		c := New()
		page, err := c.NewPage()
		require.NoError(t, err)

		err = page.AddTextColorRotated("text", 100, 700, Helvetica, 12, Color{R: -0.1}, 45)
		assert.Error(t, err)
	})
}

// TestAddTextRotated_UsesBlackByDefault verifies that AddTextRotated (without explicit color)
// produces black text — same as AddText.
func TestAddTextRotated_UsesBlackByDefault(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddTextRotated("Black", 100, 700, Helvetica, 12, 30)
	require.NoError(t, err)

	ops := page.TextOperations()
	require.Len(t, ops, 1)
	assert.Equal(t, 0.0, ops[0].Color.R)
	assert.Equal(t, 0.0, ops[0].Color.G)
	assert.Equal(t, 0.0, ops[0].Color.B)
}

// TestTextOperation_RotationField_OnStruct verifies the Rotation field is
// present on the TextOperation struct (structural test).
func TestTextOperation_RotationField_OnStruct(t *testing.T) {
	op := TextOperation{
		Text:     "Test",
		X:        100,
		Y:        200,
		Font:     Helvetica,
		Size:     12,
		Color:    Black,
		Rotation: 45.0,
	}
	assert.Equal(t, 45.0, op.Rotation)
}

// TestTextOp_RotationField_OnWriterStruct verifies the Rotation field is
// present on the writer.TextOp struct (structural test for the infrastructure layer).
func TestTextOp_RotationField_OnWriterStruct(t *testing.T) {
	op := writer.TextOp{
		Text:     "Test",
		X:        100,
		Y:        200,
		Font:     "Helvetica",
		Size:     12,
		Rotation: 90.0,
	}
	assert.Equal(t, 90.0, op.Rotation)
}

// TestConvertTextOps_RotationPassedThrough verifies that the convertTextOps
// function (inside creator) correctly propagates the Rotation field from
// TextOperation to writer.TextOp.
func TestConvertTextOps_RotationPassedThrough(t *testing.T) {
	ops := []TextOperation{
		{Text: "Normal", X: 100, Y: 700, Font: Helvetica, Size: 12, Color: Black, Rotation: 0},
		{Text: "Rotated", X: 100, Y: 600, Font: Helvetica, Size: 12, Color: Black, Rotation: 90},
		{Text: "Angled", X: 100, Y: 500, Font: Helvetica, Size: 12, Color: Black, Rotation: -45},
	}

	writerOps := convertTextOps(ops)
	require.Len(t, writerOps, 3)

	assert.Equal(t, 0.0, writerOps[0].Rotation, "no rotation")
	assert.Equal(t, 90.0, writerOps[1].Rotation, "90 degree rotation")
	assert.Equal(t, -45.0, writerOps[2].Rotation, "negative rotation")
}

// TestGenerateContentStream_RotationMatrix verifies that the content stream
// generator uses SetTextMatrix (Tm) for rotated text and MoveTextPosition (Td)
// for non-rotated text. The Tm operator must produce correct cosine/sine values.
func TestGenerateContentStream_RotationMatrix(t *testing.T) {
	t.Run("zero rotation uses Td (MoveTextPosition)", func(t *testing.T) {
		textOps := []writer.TextOp{
			{Text: "Normal", X: 100, Y: 700, Font: "Helvetica", Size: 12, Rotation: 0},
		}

		content, _, err := writer.GenerateContentStream(textOps)
		require.NoError(t, err)

		// Should contain Td (text displacement) operator, not Tm.
		contentStr := string(content)
		assert.Contains(t, contentStr, "Td", "zero rotation should use Td operator")
		assert.NotContains(t, contentStr, " Tm", "zero rotation should not use Tm operator")
	})

	t.Run("90 degree rotation produces correct Tm matrix", func(t *testing.T) {
		textOps := []writer.TextOp{
			{Text: "Sideways", X: 100, Y: 400, Font: "Helvetica", Size: 12, Rotation: 90},
		}

		content, _, err := writer.GenerateContentStream(textOps)
		require.NoError(t, err)

		// For 90° CCW: cos(90°) ≈ 0, sin(90°) = 1
		// Matrix: [0.00 1.00 -1.00 0.00 100.00 400.00] Tm
		contentStr := string(content)
		assert.Contains(t, contentStr, "Tm", "90 degree rotation should use Tm operator")
		assert.Contains(t, contentStr, "0.00 1.00 -1.00 0.00 100.00 400.00 Tm",
			"90° matrix should be [cos sin -sin cos x y] = [0 1 -1 0 100 400]")
	})

	t.Run("45 degree rotation produces correct matrix values", func(t *testing.T) {
		textOps := []writer.TextOp{
			{Text: "Diagonal", X: 200, Y: 300, Font: "Helvetica", Size: 12, Rotation: 45},
		}

		content, _, err := writer.GenerateContentStream(textOps)
		require.NoError(t, err)

		// For 45° CCW: cos(45°) = sin(45°) ≈ 0.71
		contentStr := string(content)
		assert.Contains(t, contentStr, "Tm", "45 degree rotation should use Tm operator")
		assert.Contains(t, contentStr, "0.71 0.71 -0.71 0.71 200.00 300.00 Tm",
			"45° matrix should be [cos sin -sin cos x y] ≈ [0.71 0.71 -0.71 0.71 200 300]")
	})

	t.Run("negative rotation (clockwise) also uses Tm", func(t *testing.T) {
		textOps := []writer.TextOp{
			{Text: "Clockwise", X: 100, Y: 500, Font: "Helvetica", Size: 12, Rotation: -30},
		}

		content, _, err := writer.GenerateContentStream(textOps)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "Tm", "negative rotation should use Tm operator")
	})
}

// TestAddTextRotated_WritesValidPDF is an end-to-end test verifying that a
// document containing rotated text can be written to a valid PDF byte stream.
func TestAddTextRotated_WritesValidPDF(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		rotation float64
	}{
		{"normal text (0°)", "Normal", 0},
		{"rotated 90°", "Sideways", 90},
		{"rotated 45°", "Diagonal", 45},
		{"rotated 180°", "Upside down", 180},
		{"rotated -45°", "Clockwise", -45},
		{"rotated 270°", "Other way", 270},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			page, err := c.NewPage()
			require.NoError(t, err)

			err = page.AddTextRotated(tt.text, 200, 400, Helvetica, 14, tt.rotation)
			require.NoError(t, err)

			pdfBytes, err := c.Bytes()
			require.NoError(t, err)
			require.NotEmpty(t, pdfBytes)

			assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")), "must be valid PDF")
			assert.True(t, bytes.HasSuffix(bytes.TrimSpace(pdfBytes), []byte("%%EOF")))
		})
	}
}

// TestAddTextRotated_MultipleRotationsOnOnePage verifies that multiple rotated
// text items can be added to a single page without conflicts.
func TestAddTextRotated_MultipleRotationsOnOnePage(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	rotations := []float64{0, 30, 45, 60, 90, 120, 180, 270, -45}
	for _, rot := range rotations {
		err = page.AddTextRotated("Text", 300, 400, Helvetica, 12, rot)
		require.NoError(t, err, "rotation %.0f should not error", rot)
	}

	ops := page.TextOperations()
	assert.Len(t, ops, len(rotations))

	pdfBytes, err := c.Bytes()
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")))
}

// TestAddText_RegressionNoRotation verifies that existing AddText (no rotation)
// still works correctly after the Rotation field was added to TextOperation.
// This is a regression test to ensure backward compatibility.
func TestAddText_RegressionNoRotation(t *testing.T) {
	c := New()
	page, err := c.NewPage()
	require.NoError(t, err)

	err = page.AddText("Standard text", 100, 700, Helvetica, 12)
	require.NoError(t, err)

	ops := page.TextOperations()
	require.Len(t, ops, 1)
	assert.Equal(t, 0.0, ops[0].Rotation, "AddText must produce zero rotation")

	// Verify PDF is still valid.
	pdfBytes, err := c.Bytes()
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")))
}
