package document

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormField(t *testing.T) {
	rect := [4]float64{100, 700, 300, 720}
	f := NewFormField("Tx", "username", rect)

	require.NotNil(t, f)
	assert.Equal(t, "Tx", f.FieldType())
	assert.Equal(t, "username", f.Name())
	assert.Equal(t, rect, f.Rect())
	assert.Equal(t, "", f.Value())
	assert.Equal(t, "", f.DefaultValue())
	assert.Equal(t, 0, f.Flags())
	assert.Equal(t, 4, f.AnnotationFlags())
	assert.Equal(t, "/Helv 12 Tf 0 g", f.Appearance())
	assert.Nil(t, f.BorderColor())
	assert.Nil(t, f.FillColor())
	assert.Equal(t, 0, f.MaxLength())
	assert.Nil(t, f.Options())
}

func TestFormField_SettersGetters(t *testing.T) {
	f := NewFormField("Tx", "field1", [4]float64{0, 0, 200, 20})

	// AlternateText
	f.SetAlternateText("Enter username")
	assert.Equal(t, "Enter username", f.AlternateText())

	// Value
	f.SetValue("JohnDoe")
	assert.Equal(t, "JohnDoe", f.Value())

	// DefaultValue
	f.SetDefaultValue("DefaultUser")
	assert.Equal(t, "DefaultUser", f.DefaultValue())

	// Flags
	f.SetFlags(12)
	assert.Equal(t, 12, f.Flags())

	// AnnotationFlags
	f.SetAnnotationFlags(8)
	assert.Equal(t, 8, f.AnnotationFlags())

	// Appearance
	f.SetAppearance("/Courier 10 Tf 1 0 0 rg")
	assert.Equal(t, "/Courier 10 Tf 1 0 0 rg", f.Appearance())

	// MaxLength
	f.SetMaxLength(50)
	assert.Equal(t, 50, f.MaxLength())
}

func TestFormField_BorderAndFillColor(t *testing.T) {
	f := NewFormField("Tx", "colorfield", [4]float64{0, 0, 100, 20})

	assert.Nil(t, f.BorderColor())
	assert.Nil(t, f.FillColor())

	f.SetBorderColor(0.2, 0.4, 0.6)
	bc := f.BorderColor()
	require.NotNil(t, bc)
	assert.InDelta(t, 0.2, bc[0], 1e-9)
	assert.InDelta(t, 0.4, bc[1], 1e-9)
	assert.InDelta(t, 0.6, bc[2], 1e-9)

	f.SetFillColor(0.9, 0.8, 0.7)
	fc := f.FillColor()
	require.NotNil(t, fc)
	assert.InDelta(t, 0.9, fc[0], 1e-9)
	assert.InDelta(t, 0.8, fc[1], 1e-9)
	assert.InDelta(t, 0.7, fc[2], 1e-9)
}

func TestFormField_Options(t *testing.T) {
	f := NewFormField("Ch", "dropdown", [4]float64{0, 0, 100, 20})

	assert.Nil(t, f.Options())

	opts := []string{"Option A", "Option B", "Option C"}
	f.SetOptions(opts)

	result := f.Options()
	require.NotNil(t, result)
	assert.Equal(t, opts, result)

	// Ensure it's a copy — modifying result doesn't affect f.
	result[0] = "Modified"
	assert.Equal(t, "Option A", f.Options()[0])
}

func TestFormField_Validate_FieldTypes(t *testing.T) {
	tests := []struct {
		fieldType string
		wantError bool
	}{
		{"Tx", false},
		{"Btn", false},
		{"Ch", false},
		{"Sig", false},
		{"", true},
		{"tx", true},
		{"TEXT", true},
		{"Unknown", true},
	}

	for _, tt := range tests {
		t.Run("type_"+tt.fieldType, func(t *testing.T) {
			f := NewFormField(tt.fieldType, "name", [4]float64{0, 0, 100, 20})
			err := f.Validate()
			if tt.wantError {
				assert.Error(t, err, "fieldType=%q should fail", tt.fieldType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormField_Validate_EmptyName(t *testing.T) {
	f := NewFormField("Tx", "", [4]float64{0, 0, 100, 20})
	err := f.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field name cannot be empty")
}

func TestFormField_Validate_InvalidRect(t *testing.T) {
	tests := []struct {
		name string
		rect [4]float64
	}{
		{"x2 <= x1", [4]float64{100, 0, 50, 20}},
		{"y2 <= y1", [4]float64{0, 20, 100, 10}},
		{"x1==x2", [4]float64{50, 0, 50, 20}},
		{"y1==y2", [4]float64{0, 10, 100, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormField("Tx", "f", tt.rect)
			err := f.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid rectangle")
		})
	}
}

func TestFormField_Validate_MaxLength(t *testing.T) {
	f := NewFormField("Tx", "field", [4]float64{0, 0, 200, 20})
	f.SetMaxLength(5)

	// Value within limit.
	f.SetValue("hi")
	assert.NoError(t, f.Validate())

	// Value at limit.
	f.SetValue("hello")
	assert.NoError(t, f.Validate())

	// Value exceeds limit.
	f.SetValue("hello!")
	err := f.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value exceeds maximum length")
}

func TestFormField_Validate_MaxLength_NonTxField(t *testing.T) {
	// MaxLength should only apply to Tx fields.
	f := NewFormField("Btn", "btn", [4]float64{0, 0, 100, 20})
	f.SetMaxLength(3)
	f.SetValue("this is very long value that exceeds limit")
	// Button field should not check max length.
	assert.NoError(t, f.Validate())
}

func TestFormField_Validate_InvalidBorderColor(t *testing.T) {
	f := NewFormField("Tx", "field", [4]float64{0, 0, 100, 20})
	f.SetBorderColor(1.5, 0, 0) // Out of range.
	err := f.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "border")
}

func TestFormField_Validate_InvalidFillColor(t *testing.T) {
	f := NewFormField("Tx", "field", [4]float64{0, 0, 100, 20})
	f.SetFillColor(0, -0.1, 0) // Out of range.
	err := f.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fill")
}

func TestFormField_Validate_ValidColors(t *testing.T) {
	f := NewFormField("Tx", "field", [4]float64{0, 0, 100, 20})
	f.SetBorderColor(0, 0, 0)
	f.SetFillColor(1, 1, 1)
	assert.NoError(t, f.Validate())
}

func TestFormField_AllFieldTypes_Validate(t *testing.T) {
	// Each field type should validate without error when valid.
	for _, ft := range []string{"Tx", "Btn", "Ch", "Sig"} {
		f := NewFormField(ft, "f", [4]float64{0, 0, 100, 20})
		assert.NoError(t, f.Validate(), "field type %q should pass validation", ft)
	}
}
