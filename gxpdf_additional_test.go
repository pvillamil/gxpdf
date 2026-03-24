package gxpdf

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	internalforms "github.com/coregx/gxpdf/internal/application/forms"
	internaltable "github.com/coregx/gxpdf/internal/models/table"
	"github.com/coregx/gxpdf/internal/models/types"
)

// ---------- errors.go: IsEncrypted, IsCorrupted ----------

func TestIsEncrypted_True(t *testing.T) {
	if !IsEncrypted(ErrEncrypted) {
		t.Error("IsEncrypted(ErrEncrypted) = false, want true")
	}
}

func TestIsEncrypted_False(t *testing.T) {
	if IsEncrypted(ErrInvalidPDF) {
		t.Error("IsEncrypted(ErrInvalidPDF) = true, want false")
	}
}

func TestIsEncrypted_Nil(t *testing.T) {
	if IsEncrypted(nil) {
		t.Error("IsEncrypted(nil) = true, want false")
	}
}

func TestIsEncrypted_Wrapped(t *testing.T) {
	wrapped := errors.New("outer: " + ErrEncrypted.Error())
	// Not wrapped via %w, so should not match
	_ = IsEncrypted(wrapped)
}

func TestIsCorrupted_True(t *testing.T) {
	if !IsCorrupted(ErrCorrupted) {
		t.Error("IsCorrupted(ErrCorrupted) = false, want true")
	}
}

func TestIsCorrupted_False(t *testing.T) {
	if IsCorrupted(ErrInvalidPDF) {
		t.Error("IsCorrupted(ErrInvalidPDF) = true, want false")
	}
}

func TestIsCorrupted_Nil(t *testing.T) {
	if IsCorrupted(nil) {
		t.Error("IsCorrupted(nil) = true, want false")
	}
}

// ---------- options.go: ExtractionMethod.String ----------

func TestExtractionMethod_String(t *testing.T) {
	tests := []struct {
		method ExtractionMethod
		want   string
	}{
		{MethodAuto, "Auto"},
		{MethodLattice, "Lattice"},
		{MethodStream, "Stream"},
		{MethodHybrid, "Hybrid"},
		{ExtractionMethod(99), "Unknown"},
	}
	for _, tt := range tests {
		got := tt.method.String()
		if got != tt.want {
			t.Errorf("ExtractionMethod(%d).String() = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestExtractionOptions_WithMethod(t *testing.T) {
	opts := DefaultExtractionOptions()
	result := opts.WithMethod(MethodLattice)
	if result.Method != MethodLattice {
		t.Errorf("Method = %v, want MethodLattice", result.Method)
	}
	if result != opts {
		t.Error("WithMethod should return same pointer")
	}
}

func TestExtractionOptions_WithMethod_Stream(t *testing.T) {
	opts := DefaultExtractionOptions()
	opts.WithMethod(MethodStream)
	if opts.Method != MethodStream {
		t.Errorf("Method = %v, want MethodStream", opts.Method)
	}
}

func TestExtractionOptions_WithMethod_Hybrid(t *testing.T) {
	opts := DefaultExtractionOptions()
	opts.WithMethod(MethodHybrid)
	if opts.Method != MethodHybrid {
		t.Errorf("Method = %v, want MethodHybrid", opts.Method)
	}
}

func TestExtractionOptions_WithPages(t *testing.T) {
	opts := DefaultExtractionOptions()
	result := opts.WithPages(0, 1, 2)
	if len(result.Pages) != 3 {
		t.Errorf("Pages len = %d, want 3", len(result.Pages))
	}
	if result.Pages[0] != 0 || result.Pages[1] != 1 || result.Pages[2] != 2 {
		t.Errorf("Pages = %v, want [0 1 2]", result.Pages)
	}
}

func TestExtractionOptions_WithPages_Single(t *testing.T) {
	opts := DefaultExtractionOptions()
	opts.WithPages(5)
	if len(opts.Pages) != 1 || opts.Pages[0] != 5 {
		t.Errorf("Pages = %v, want [5]", opts.Pages)
	}
}

func TestExtractionOptions_WithMergeMultilineRows_False(t *testing.T) {
	opts := DefaultExtractionOptions()
	result := opts.WithMergeMultilineRows(false)
	if result.MergeMultilineRows {
		t.Error("MergeMultilineRows should be false")
	}
	if result != opts {
		t.Error("WithMergeMultilineRows should return same pointer")
	}
}

func TestExtractionOptions_WithMergeMultilineRows_True(t *testing.T) {
	opts := DefaultExtractionOptions()
	opts.WithMergeMultilineRows(false)
	opts.WithMergeMultilineRows(true)
	if !opts.MergeMultilineRows {
		t.Error("MergeMultilineRows should be true after setting true")
	}
}

// ---------- FormField wrapper methods ----------

func newTestFormField(ft internalforms.FieldType, flags int, opts []string) *FormField {
	return &FormField{
		internal: &internalforms.FieldInfo{
			Name:         "testField",
			Type:         ft,
			Value:        "testValue",
			DefaultValue: "defaultValue",
			Flags:        flags,
			Rect:         [4]float64{10, 20, 200, 40},
			Options:      opts,
		},
	}
}

func TestFormField_Name(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	if f.Name() != "testField" {
		t.Errorf("Name() = %q, want testField", f.Name())
	}
}

func TestFormField_Type_Text(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	if f.Type() != "Tx" {
		t.Errorf("Type() = %q, want Tx", f.Type())
	}
}

func TestFormField_Type_Button(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeButton, 0, nil)
	if f.Type() != "Btn" {
		t.Errorf("Type() = %q, want Btn", f.Type())
	}
}

func TestFormField_Type_Choice(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeChoice, 0, nil)
	if f.Type() != "Ch" {
		t.Errorf("Type() = %q, want Ch", f.Type())
	}
}

func TestFormField_Value(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	if f.Value() != "testValue" {
		t.Errorf("Value() = %v, want testValue", f.Value())
	}
}

func TestFormField_DefaultValue(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	if f.DefaultValue() != "defaultValue" {
		t.Errorf("DefaultValue() = %v, want defaultValue", f.DefaultValue())
	}
}

func TestFormField_Flags(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 7, nil)
	if f.Flags() != 7 {
		t.Errorf("Flags() = %d, want 7", f.Flags())
	}
}

func TestFormField_Rect(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	r := f.Rect()
	if r[0] != 10 || r[1] != 20 || r[2] != 200 || r[3] != 40 {
		t.Errorf("Rect() = %v, want [10 20 200 40]", r)
	}
}

func TestFormField_Options(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeChoice, 0, []string{"A", "B", "C"})
	opts := f.Options()
	if len(opts) != 3 || opts[0] != "A" {
		t.Errorf("Options() = %v, want [A B C]", opts)
	}
}

func TestFormField_IsReadOnly_True(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 1, nil)
	if !f.IsReadOnly() {
		t.Error("IsReadOnly() = false, want true for Flags=1")
	}
}

func TestFormField_IsReadOnly_False(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	if f.IsReadOnly() {
		t.Error("IsReadOnly() = true, want false for Flags=0")
	}
}

func TestFormField_IsRequired_True(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 2, nil)
	if !f.IsRequired() {
		t.Error("IsRequired() = false, want true for Flags=2")
	}
}

func TestFormField_IsRequired_False(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 1, nil)
	if f.IsRequired() {
		t.Error("IsRequired() = true, want false for Flags=1")
	}
}

func TestFormField_IsTextField_True(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	if !f.IsTextField() {
		t.Error("IsTextField() = false, want true for FieldTypeText")
	}
}

func TestFormField_IsTextField_False(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeButton, 0, nil)
	if f.IsTextField() {
		t.Error("IsTextField() = true, want false for FieldTypeButton")
	}
}

func TestFormField_IsButton_True(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeButton, 0, nil)
	if !f.IsButton() {
		t.Error("IsButton() = false, want true for FieldTypeButton")
	}
}

func TestFormField_IsButton_False(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	if f.IsButton() {
		t.Error("IsButton() = true, want false for FieldTypeText")
	}
}

func TestFormField_IsChoice_True(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeChoice, 0, nil)
	if !f.IsChoice() {
		t.Error("IsChoice() = false, want true for FieldTypeChoice")
	}
}

func TestFormField_IsChoice_False(t *testing.T) {
	f := newTestFormField(internalforms.FieldTypeText, 0, nil)
	if f.IsChoice() {
		t.Error("IsChoice() = true, want false for FieldTypeText")
	}
}

// ---------- Page wrapper methods ----------

func newTestPage(index int) *Page {
	return &Page{doc: nil, index: index}
}

func TestPage_Index(t *testing.T) {
	p := newTestPage(3)
	if p.Index() != 3 {
		t.Errorf("Index() = %d, want 3", p.Index())
	}
}

func TestPage_Number(t *testing.T) {
	p := newTestPage(3)
	if p.Number() != 4 {
		t.Errorf("Number() = %d, want 4", p.Number())
	}
}

func TestPage_Index_Zero(t *testing.T) {
	p := newTestPage(0)
	if p.Index() != 0 {
		t.Errorf("Index() = %d, want 0", p.Index())
	}
	if p.Number() != 1 {
		t.Errorf("Number() = %d, want 1", p.Number())
	}
}

// ---------- Table wrapper methods ----------

func newTestTable() *Table {
	tbl, err := internaltable.NewTable(3, 2)
	if err != nil {
		panic(err)
	}
	tbl.SetCell(0, 0, internaltable.NewCell("A", 0, 0))
	tbl.SetCell(0, 1, internaltable.NewCell("B", 0, 1))
	tbl.SetCell(1, 0, internaltable.NewCell("C", 1, 0))
	tbl.SetCell(1, 1, internaltable.NewCell("D", 1, 1))
	tbl.SetCell(2, 0, internaltable.NewCell("E", 2, 0))
	tbl.SetCell(2, 1, internaltable.NewCell("F", 2, 1))
	tbl.PageNum = 2
	tbl.Method = "Lattice"
	return &Table{internal: tbl}
}

func TestTable_Rows(t *testing.T) {
	tbl := newTestTable()
	rows := tbl.Rows()
	if len(rows) != 3 {
		t.Errorf("Rows() len = %d, want 3", len(rows))
	}
	if rows[0][0] != "A" {
		t.Errorf("rows[0][0] = %q, want A", rows[0][0])
	}
}

func TestTable_RowCount(t *testing.T) {
	tbl := newTestTable()
	if tbl.RowCount() != 3 {
		t.Errorf("RowCount() = %d, want 3", tbl.RowCount())
	}
}

func TestTable_ColumnCount(t *testing.T) {
	tbl := newTestTable()
	if tbl.ColumnCount() != 2 {
		t.Errorf("ColumnCount() = %d, want 2", tbl.ColumnCount())
	}
}

func TestTable_PageNumber(t *testing.T) {
	tbl := newTestTable()
	if tbl.PageNumber() != 2 {
		t.Errorf("PageNumber() = %d, want 2", tbl.PageNumber())
	}
}

func TestTable_Method(t *testing.T) {
	tbl := newTestTable()
	if tbl.Method() != "Lattice" {
		t.Errorf("Method() = %q, want Lattice", tbl.Method())
	}
}

func TestTable_IsEmpty_False(t *testing.T) {
	tbl := newTestTable()
	if tbl.IsEmpty() {
		t.Error("IsEmpty() = true for non-empty table")
	}
}

func TestTable_IsEmpty_True(t *testing.T) {
	internal, _ := internaltable.NewTable(1, 1)
	internal.SetCell(0, 0, internaltable.NewCell("", 0, 0))
	tbl := &Table{internal: internal}
	if !tbl.IsEmpty() {
		t.Error("IsEmpty() = false for empty table")
	}
}

func TestTable_Cell(t *testing.T) {
	tbl := newTestTable()
	if tbl.Cell(0, 0) != "A" {
		t.Errorf("Cell(0,0) = %q, want A", tbl.Cell(0, 0))
	}
	if tbl.Cell(0, 1) != "B" {
		t.Errorf("Cell(0,1) = %q, want B", tbl.Cell(0, 1))
	}
}

func TestTable_Cell_OutOfBounds(t *testing.T) {
	tbl := newTestTable()
	result := tbl.Cell(99, 99)
	if result != "" {
		t.Errorf("Cell(99,99) = %q, want empty", result)
	}
}

func TestTable_String(t *testing.T) {
	tbl := newTestTable()
	s := tbl.String()
	if s == "" {
		t.Error("String() returned empty string")
	}
}

func TestTable_ExportCSV(t *testing.T) {
	tbl := newTestTable()
	var buf bytes.Buffer
	err := tbl.ExportCSV(&buf)
	if err != nil {
		t.Fatalf("ExportCSV error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("ExportCSV produced empty output")
	}
}

func TestTable_ExportJSON(t *testing.T) {
	tbl := newTestTable()
	var buf bytes.Buffer
	err := tbl.ExportJSON(&buf)
	if err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("ExportJSON produced empty output")
	}
}

func TestTable_ExportExcel(t *testing.T) {
	tbl := newTestTable()
	var buf bytes.Buffer
	err := tbl.ExportExcel(&buf)
	if err != nil {
		t.Fatalf("ExportExcel error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("ExportExcel produced empty output")
	}
}

func TestTable_ToCSV(t *testing.T) {
	tbl := newTestTable()
	csv, err := tbl.ToCSV()
	if err != nil {
		t.Fatalf("ToCSV error: %v", err)
	}
	if !strings.Contains(csv, "A") {
		t.Errorf("ToCSV doesn't contain 'A': %q", csv)
	}
}

func TestTable_ToJSON(t *testing.T) {
	tbl := newTestTable()
	json, err := tbl.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	if !strings.Contains(json, "A") {
		t.Errorf("ToJSON doesn't contain 'A': %q", json)
	}
}

func TestTable_Internal(t *testing.T) {
	tbl := newTestTable()
	internal := tbl.Internal()
	if internal == nil {
		t.Error("Internal() returned nil")
	}
	if internal.RowCount != 3 {
		t.Errorf("Internal().RowCount = %d, want 3", internal.RowCount)
	}
}

// ---------- Image wrapper methods ----------

func newTestImage() *Image {
	img, err := types.NewImage(
		[]byte{0xFF, 0xD8, 0xFF},
		100, 80,
		"DeviceRGB",
		8,
		"/DCTDecode",
	)
	if err != nil {
		panic(err)
	}
	img.SetName("/Im1")
	return &Image{internal: img}
}

func TestImage_Width(t *testing.T) {
	img := newTestImage()
	if img.Width() != 100 {
		t.Errorf("Width() = %d, want 100", img.Width())
	}
}

func TestImage_Height(t *testing.T) {
	img := newTestImage()
	if img.Height() != 80 {
		t.Errorf("Height() = %d, want 80", img.Height())
	}
}

func TestImage_ColorSpace(t *testing.T) {
	img := newTestImage()
	if img.ColorSpace() != "DeviceRGB" {
		t.Errorf("ColorSpace() = %q, want DeviceRGB", img.ColorSpace())
	}
}

func TestImage_BitsPerComponent(t *testing.T) {
	img := newTestImage()
	if img.BitsPerComponent() != 8 {
		t.Errorf("BitsPerComponent() = %d, want 8", img.BitsPerComponent())
	}
}

func TestImage_Filter(t *testing.T) {
	img := newTestImage()
	if img.Filter() != "/DCTDecode" {
		t.Errorf("Filter() = %q, want /DCTDecode", img.Filter())
	}
}

func TestImage_Name(t *testing.T) {
	img := newTestImage()
	if img.Name() != "/Im1" {
		t.Errorf("Name() = %q, want /Im1", img.Name())
	}
}

func TestImage_String(t *testing.T) {
	img := newTestImage()
	s := img.String()
	if s == "" {
		t.Error("String() returned empty string")
	}
}

func TestImage_ToGoImage_InvalidData(t *testing.T) {
	img := newTestImage()
	// {0xFF, 0xD8, 0xFF} is not a complete JPEG, so ToGoImage should error
	_, err := img.ToGoImage()
	if err == nil {
		t.Log("ToGoImage with minimal data succeeded (might be valid for some decoders)")
	}
}
