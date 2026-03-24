package forms

import (
	"testing"

	"github.com/coregx/gxpdf/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openFormReader opens the generated form PDF as a parser.Reader.
// Skips the test if formPDFPath is not available.
func openFormReader(t *testing.T) *parser.Reader {
	t.Helper()
	if formPDFPath == "" {
		t.Skip("form PDF not generated; skipping test requiring real parser.Reader")
	}
	r, err := parser.OpenPDF(formPDFPath)
	require.NoError(t, err, "opening form PDF")
	t.Cleanup(func() { _ = r.Close() })
	return r
}

// ============================================================================
// Writer.ValidateFieldValue — requires real parser.Reader to find field
// ============================================================================

func TestWriter_ValidateFieldValue_ViaRealReader_Valid(t *testing.T) {
	r := openFormReader(t)
	w := NewWriter(r)
	// "firstName" is a text field in the generated PDF.
	err := w.ValidateFieldValue("firstName", "Alice")
	assert.NoError(t, err)
}

func TestWriter_ValidateFieldValue_ViaRealReader_WrongType(t *testing.T) {
	r := openFormReader(t)
	w := NewWriter(r)
	// Passing an integer to a text field should error.
	err := w.ValidateFieldValue("firstName", 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires string value")
}

func TestWriter_ValidateFieldValue_ViaRealReader_FieldNotFound(t *testing.T) {
	r := openFormReader(t)
	w := NewWriter(r)
	err := w.ValidateFieldValue("nonExistentField", "value")
	assert.Error(t, err)
}

// ============================================================================
// Writer.SetFieldValue — requires real parser.Reader
// ============================================================================

func TestWriter_SetFieldValue_ViaRealReader_Success(t *testing.T) {
	r := openFormReader(t)
	w := NewWriter(r)

	err := w.SetFieldValue("firstName", "Bob")
	require.NoError(t, err)
	assert.True(t, w.HasUpdates())
	assert.Equal(t, "Bob", w.GetUpdates()["firstName"])
}

// ============================================================================
// Writer.GetFieldsToUpdate — with a real update
// ============================================================================

func TestWriter_GetFieldsToUpdate_WithUpdate(t *testing.T) {
	r := openFormReader(t)
	w := NewWriter(r)

	err := w.SetFieldValue("firstName", "Charlie")
	require.NoError(t, err)

	fields, err := w.GetFieldsToUpdate()
	require.NoError(t, err)
	require.Len(t, fields, 1)
	assert.Equal(t, "firstName", fields[0].Name)
}

// ============================================================================
// Writer.ApplyUpdatesToDict — button field with bool true (covers setValueInDict bool)
// ============================================================================

func TestWriter_ApplyUpdatesToDict_ButtonBoolTrue(t *testing.T) {
	w := NewWriter(nil)
	w.updates["checkField"] = true

	field := &FieldInfo{Name: "checkField", Type: FieldTypeButton}
	dict := parser.NewDictionary()

	result := w.ApplyUpdatesToDict(field, dict)
	require.NotNil(t, result)
	assert.NotNil(t, result.Get("V"))
	assert.NotNil(t, result.Get("AS"))
}

// ============================================================================
// Writer.ApplyUpdatesToDict — button field with bool false
// ============================================================================

func TestWriter_ApplyUpdatesToDict_ButtonBoolFalse(t *testing.T) {
	w := NewWriter(nil)
	w.updates["checkField"] = false

	field := &FieldInfo{Name: "checkField", Type: FieldTypeButton}
	dict := parser.NewDictionary()

	result := w.ApplyUpdatesToDict(field, dict)
	require.NotNil(t, result)
	assert.NotNil(t, result.Get("V"))
}

// ============================================================================
// Writer.ApplyUpdatesToDict — button field with string value
// ============================================================================

func TestWriter_ApplyUpdatesToDict_ButtonString(t *testing.T) {
	w := NewWriter(nil)
	w.updates["radioField"] = "Option1"

	field := &FieldInfo{Name: "radioField", Type: FieldTypeButton}
	dict := parser.NewDictionary()

	result := w.ApplyUpdatesToDict(field, dict)
	require.NotNil(t, result)
	assert.NotNil(t, result.Get("V"))
	assert.NotNil(t, result.Get("AS"))
}

// ============================================================================
// Writer.ApplyUpdatesToDict — choice field with []string
// ============================================================================

func TestWriter_ApplyUpdatesToDict_ChoiceMultiSelect(t *testing.T) {
	w := NewWriter(nil)
	w.updates["listBox"] = []string{"OptionA", "OptionB"}

	field := &FieldInfo{Name: "listBox", Type: FieldTypeChoice}
	dict := parser.NewDictionary()

	result := w.ApplyUpdatesToDict(field, dict)
	require.NotNil(t, result)
	assert.NotNil(t, result.Get("V"))
}

// ============================================================================
// Writer.ValidateFieldValue — signature and unknown type (via validateFieldValueType)
// ============================================================================

func TestValidateFieldValueType_SignatureType(t *testing.T) {
	field := &FieldInfo{Name: "sig", Type: FieldTypeSignature}
	err := validateFieldValueType(field, "anything")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature field")
}

func TestValidateFieldValueType_UnknownType_AllowsAny(t *testing.T) {
	field := &FieldInfo{Name: "unknown", Type: FieldType("UnknownCustomType")}
	err := validateFieldValueType(field, 99999)
	assert.NoError(t, err)
}
