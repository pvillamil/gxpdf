package creator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coregx/gxpdf/creator/forms"
)

// TestMapFontNameToPDF verifies font name mapping for PDF appearance strings.
func TestMapFontNameToPDF(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Helvetica", "Helv"},
		{"Courier", "Cour"},
		{"Times-Roman", "TiRo"},
		{"Times", "TiRo"},
		{"ArialMT", "Aria"}, // >4 chars: take first 4
		{"Abc", "Abc"},      // <=4 chars: return as-is
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapFontNameToPDF(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestBuildAppearanceString verifies the /DA string format.
func TestBuildAppearanceString(t *testing.T) {
	s := buildAppearanceString("Helvetica", 12.0, [3]float64{0, 0, 0})
	assert.True(t, strings.HasPrefix(s, "/Helv "), "should start with mapped font name")
	assert.Contains(t, s, "12.00 Tf")
	assert.Contains(t, s, "0.000 0.000 0.000 rg")
}

func TestBuildAppearanceString_CustomColor(t *testing.T) {
	s := buildAppearanceString("Courier", 10.5, [3]float64{1, 0, 0.5})
	assert.Contains(t, s, "/Cour")
	assert.Contains(t, s, "10.50 Tf")
	assert.Contains(t, s, "1.000 0.000 0.500 rg")
}

// TestConvertTextFieldToDomain tests the internal conversion function.
func TestConvertTextFieldToDomain(t *testing.T) {
	tf := forms.NewTextField("username", 100, 700, 200, 20)
	tf.SetValue("Alice")

	field, err := convertTextFieldToDomain(tf)
	require.NoError(t, err)
	require.NotNil(t, field)
	assert.Equal(t, "Tx", field.FieldType())
	assert.Equal(t, "username", field.Name())
}

// TestConvertTextFieldToDomain_WithBorderAndFill exercises the color branches.
func TestConvertTextFieldToDomain_WithBorderAndFill(t *testing.T) {
	tf := forms.NewTextField("email", 50, 600, 150, 20)

	r, g, b := 0.0, 0.0, 0.0
	require.NoError(t, tf.SetBorderColor(&r, &g, &b))

	fr, fg, fb := 1.0, 1.0, 1.0
	require.NoError(t, tf.SetFillColor(&fr, &fg, &fb))

	field, err := convertTextFieldToDomain(tf)
	require.NoError(t, err)
	require.NotNil(t, field)
}

// TestConvertTextFieldToDomain_InvalidField verifies validation error propagation.
func TestConvertTextFieldToDomain_InvalidField(t *testing.T) {
	// NewTextField with empty name should fail Validate().
	tf := forms.NewTextField("", 100, 700, 200, 20)

	_, err := convertTextFieldToDomain(tf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation")
}

// TestConvertFieldToDomain_UnsupportedType verifies unsupported type error.
func TestConvertFieldToDomain_UnsupportedType(t *testing.T) {
	_, err := convertFieldToDomain("not a form field")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedFieldType)
}
