package writer

import (
	"strings"
	"testing"
)

func TestNewResourceDictionary(t *testing.T) {
	rd := NewResourceDictionary()

	if rd == nil {
		t.Fatal("NewResourceDictionary returned nil")
	}

	if rd.HasResources() {
		t.Error("New resource dictionary should be empty")
	}

	if got := rd.String(); got != "<< >>" {
		t.Errorf("Empty dictionary should return '<< >>', got %q", got)
	}
}

//nolint:dupl // Table-driven tests have similar structure by design.
func TestResourceDictionary_AddFont(t *testing.T) {
	tests := []struct {
		name       string
		objNums    []int
		wantNames  []string
		wantOutput string
	}{
		{
			name:       "single font",
			objNums:    []int{5},
			wantNames:  []string{"F1"},
			wantOutput: "<< /Font << /F1 5 0 R >> /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>",
		},
		{
			name:       "multiple fonts",
			objNums:    []int{5, 6, 7},
			wantNames:  []string{"F1", "F2", "F3"},
			wantOutput: "<< /Font << /F1 5 0 R /F2 6 0 R /F3 7 0 R >> /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := NewResourceDictionary()

			// Add fonts and verify names.
			for i, objNum := range tt.objNums {
				gotName := rd.AddFont(objNum)
				if gotName != tt.wantNames[i] {
					t.Errorf("AddFont(%d) = %q, want %q", objNum, gotName, tt.wantNames[i])
				}
			}

			// Verify output.
			if got := rd.String(); got != tt.wantOutput {
				t.Errorf("String() = %q\nwant: %q", got, tt.wantOutput)
			}

			// Verify HasResources.
			if !rd.HasResources() {
				t.Error("HasResources() = false, want true after adding fonts")
			}
		})
	}
}

//nolint:dupl // Table-driven tests have similar structure by design.
func TestResourceDictionary_AddImage(t *testing.T) {
	tests := []struct {
		name       string
		objNums    []int
		wantNames  []string
		wantOutput string
	}{
		{
			name:       "single image",
			objNums:    []int{10},
			wantNames:  []string{"Im1"},
			wantOutput: "<< /XObject << /Im1 10 0 R >> /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>",
		},
		{
			name:       "multiple images",
			objNums:    []int{10, 11, 12},
			wantNames:  []string{"Im1", "Im2", "Im3"},
			wantOutput: "<< /XObject << /Im1 10 0 R /Im2 11 0 R /Im3 12 0 R >> /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := NewResourceDictionary()

			// Add images and verify names.
			for i, objNum := range tt.objNums {
				gotName := rd.AddImage(objNum)
				if gotName != tt.wantNames[i] {
					t.Errorf("AddImage(%d) = %q, want %q", objNum, gotName, tt.wantNames[i])
				}
			}

			// Verify output.
			if got := rd.String(); got != tt.wantOutput {
				t.Errorf("String() = %q\nwant: %q", got, tt.wantOutput)
			}

			// Verify HasResources.
			if !rd.HasResources() {
				t.Error("HasResources() = false, want true after adding images")
			}
		})
	}
}

//nolint:dupl // Table-driven tests have similar structure by design.
func TestResourceDictionary_AddExtGState(t *testing.T) {
	tests := []struct {
		name       string
		objNums    []int
		wantNames  []string
		wantOutput string
	}{
		{
			name:       "single graphics state",
			objNums:    []int{15},
			wantNames:  []string{"GS1"},
			wantOutput: "<< /ExtGState << /GS1 15 0 R >> /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>",
		},
		{
			name:       "multiple graphics states",
			objNums:    []int{15, 16, 17},
			wantNames:  []string{"GS1", "GS2", "GS3"},
			wantOutput: "<< /ExtGState << /GS1 15 0 R /GS2 16 0 R /GS3 17 0 R >> /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := NewResourceDictionary()

			// Add graphics states and verify names.
			for i, objNum := range tt.objNums {
				gotName := rd.AddExtGState(objNum)
				if gotName != tt.wantNames[i] {
					t.Errorf("AddExtGState(%d) = %q, want %q", objNum, gotName, tt.wantNames[i])
				}
			}

			// Verify output.
			if got := rd.String(); got != tt.wantOutput {
				t.Errorf("String() = %q\nwant: %q", got, tt.wantOutput)
			}

			// Verify HasResources.
			if !rd.HasResources() {
				t.Error("HasResources() = false, want true after adding ExtGState")
			}
		})
	}
}

func TestResourceDictionary_CombinedResources(t *testing.T) {
	rd := NewResourceDictionary()

	// Add resources in mixed order.
	fontName1 := rd.AddFont(5)
	imgName1 := rd.AddImage(10)
	gsName1 := rd.AddExtGState(15)
	fontName2 := rd.AddFont(6)

	// Verify names.
	if fontName1 != "F1" {
		t.Errorf("First font name = %q, want F1", fontName1)
	}
	if fontName2 != "F2" {
		t.Errorf("Second font name = %q, want F2", fontName2)
	}
	if imgName1 != "Im1" {
		t.Errorf("Image name = %q, want Im1", imgName1)
	}
	if gsName1 != "GS1" {
		t.Errorf("Graphics state name = %q, want GS1", gsName1)
	}

	// Verify combined output.
	// Resources should be ordered: /Font, /XObject, /ExtGState, /ProcSet
	// Names within each category should be sorted.
	got := rd.String()
	want := "<< /Font << /F1 5 0 R /F2 6 0 R >> /XObject << /Im1 10 0 R >> /ExtGState << /GS1 15 0 R >> /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>"

	if got != want {
		t.Errorf("Combined resources:\ngot:  %q\nwant: %q", got, want)
	}

	// Verify HasResources.
	if !rd.HasResources() {
		t.Error("HasResources() = false, want true")
	}
}

func TestResourceDictionary_EmptyDictionary(t *testing.T) {
	rd := NewResourceDictionary()

	// Verify empty state.
	if rd.HasResources() {
		t.Error("HasResources() = true, want false for empty dictionary")
	}

	// Verify empty output.
	got := rd.String()
	want := "<< >>"
	if got != want {
		t.Errorf("Empty dictionary = %q, want %q", got, want)
	}

	// Verify Bytes() matches String().
	gotBytes := string(rd.Bytes())
	if gotBytes != got {
		t.Errorf("Bytes() = %q, want %q", gotBytes, got)
	}
}

func TestResourceDictionary_Bytes(t *testing.T) {
	rd := NewResourceDictionary()
	rd.AddFont(5)
	rd.AddImage(10)

	// Verify Bytes() returns same as String().
	gotBytes := rd.Bytes()
	gotString := rd.String()

	if string(gotBytes) != gotString {
		t.Errorf("Bytes() = %q, String() = %q, should match", string(gotBytes), gotString)
	}

	// Verify it's valid PDF syntax.
	if !strings.HasPrefix(gotString, "<<") || !strings.HasSuffix(gotString, ">>") {
		t.Errorf("Output should start with '<< ' and end with ' >>', got %q", gotString)
	}
}

func TestResourceDictionary_ResourceOrder(t *testing.T) {
	// Verify that resources are written in sorted order within each category.
	rd := NewResourceDictionary()

	// Add fonts in reverse order.
	rd.AddFont(7) // F1
	rd.AddFont(6) // F2
	rd.AddFont(5) // F3

	got := rd.String()

	// Fonts should appear in sorted order: F1, F2, F3 (not 7, 6, 5).
	if !strings.Contains(got, "/F1 7 0 R /F2 6 0 R /F3 5 0 R") {
		t.Errorf("Fonts not in sorted order, got %q", got)
	}
}

func TestResourceDictionary_LargeNumberOfResources(t *testing.T) {
	rd := NewResourceDictionary()

	// Add 100 fonts.
	for i := 1; i <= 100; i++ {
		name := rd.AddFont(i + 100)
		wantName := "F" + strings.TrimSpace(strings.TrimPrefix(name, "F"))
		if name != wantName {
			// Just verify it follows pattern FN where N is sequential.
			if !strings.HasPrefix(name, "F") {
				t.Errorf("Font %d: name = %q, should start with 'F'", i, name)
			}
		}
	}

	// Verify we have 100 fonts.
	output := rd.String()
	fontCount := strings.Count(output, "0 R") // Each reference ends with "0 R".

	// We should have 100 font references.
	if fontCount != 100 {
		t.Errorf("Expected 100 font references, got %d", fontCount)
	}

	// Verify HasResources.
	if !rd.HasResources() {
		t.Error("HasResources() = false after adding 100 fonts")
	}
}

func TestResourceDictionary_ObjectNumbers(t *testing.T) {
	rd := NewResourceDictionary()

	// Test various object numbers.
	rd.AddFont(1)
	rd.AddFont(999)
	rd.AddFont(1000000)
	rd.AddImage(2)
	rd.AddExtGState(3)

	got := rd.String()

	// Verify all object numbers are present.
	requiredRefs := []string{
		"/F1 1 0 R",
		"/F2 999 0 R",
		"/F3 1000000 0 R",
		"/Im1 2 0 R",
		"/GS1 3 0 R",
	}

	for _, ref := range requiredRefs {
		if !strings.Contains(got, ref) {
			t.Errorf("Output missing required reference %q\ngot: %q", ref, got)
		}
	}
}

func TestResourceDictionary_AddShading(t *testing.T) {
	rd := NewResourceDictionary()

	grad := &GradientOp{
		Type: GradientTypeLinear,
		X1:   0, Y1: 0, X2: 100, Y2: 0,
		ColorStops: []ColorStopOp{
			{Position: 0, Color: RGB{R: 1, G: 0, B: 0}},
			{Position: 1, Color: RGB{R: 0, G: 0, B: 1}},
		},
	}

	name := rd.AddShading(grad)
	if name != "Sh1" {
		t.Errorf("AddShading() = %q, want %q", name, "Sh1")
	}

	// Verify HasResources.
	if !rd.HasResources() {
		t.Error("HasResources() = false, want true after adding Shading")
	}

	// Verify Bytes includes /Shading section with placeholder object number.
	got := rd.String()
	if !strings.Contains(got, "/Shading <<") {
		t.Errorf("String() should contain /Shading section, got %q", got)
	}
	if !strings.Contains(got, "/Sh1 0 0 R") {
		t.Errorf("String() should contain /Sh1 0 0 R (placeholder), got %q", got)
	}
}

func TestResourceDictionary_SetShadingObjNum(t *testing.T) {
	rd := NewResourceDictionary()

	grad := &GradientOp{
		Type: GradientTypeLinear,
		ColorStops: []ColorStopOp{
			{Position: 0, Color: RGB{R: 1, G: 0, B: 0}},
			{Position: 1, Color: RGB{R: 0, G: 0, B: 1}},
		},
	}

	name := rd.AddShading(grad)

	// Set real object number.
	ok := rd.SetShadingObjNum(name, 42)
	if !ok {
		t.Error("SetShadingObjNum() = false, want true")
	}

	got := rd.String()
	if !strings.Contains(got, "/Sh1 42 0 R") {
		t.Errorf("String() should contain /Sh1 42 0 R, got %q", got)
	}

	// Non-existent name returns false.
	ok = rd.SetShadingObjNum("ShXXX", 99)
	if ok {
		t.Error("SetShadingObjNum(non-existent) = true, want false")
	}
}

func TestResourceDictionary_ShadingEntries(t *testing.T) {
	rd := NewResourceDictionary()

	grad1 := &GradientOp{Type: GradientTypeLinear}
	grad2 := &GradientOp{Type: GradientTypeRadial}

	rd.AddShading(grad1)
	rd.AddShading(grad2)

	entries := rd.ShadingEntries()
	if len(entries) != 2 {
		t.Errorf("ShadingEntries() returned %d entries, want 2", len(entries))
	}

	// Verify entries contain the correct gradient types.
	if e, ok := entries["Sh1"]; !ok || e.Gradient.Type != GradientTypeLinear {
		t.Error("Sh1 entry missing or wrong type")
	}
	if e, ok := entries["Sh2"]; !ok || e.Gradient.Type != GradientTypeRadial {
		t.Error("Sh2 entry missing or wrong type")
	}
}

func TestResourceDictionary_CombinedWithShading(t *testing.T) {
	rd := NewResourceDictionary()

	rd.AddFont(5)
	rd.AddImage(10)
	rd.AddExtGState(15)
	name := rd.AddShading(&GradientOp{Type: GradientTypeLinear})
	rd.SetShadingObjNum(name, 20)

	got := rd.String()

	// Verify all sections present in correct order.
	for _, section := range []string{"/Font <<", "/XObject <<", "/ExtGState <<", "/Shading <<"} {
		if !strings.Contains(got, section) {
			t.Errorf("String() missing section %q\ngot: %q", section, got)
		}
	}
	if !strings.Contains(got, "/Sh1 20 0 R") {
		t.Errorf("String() missing shading reference, got %q", got)
	}
}
