package writer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/coregx/gxpdf/internal/document"
	"github.com/coregx/gxpdf/internal/models/types"
)

func TestCreatePageTree_SinglePage(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	doc := document.NewDocument()
	_, err := doc.AddPage(document.A4)
	if err != nil {
		t.Fatalf("AddPage() error = %v", err)
	}

	objects, rootRef, err := w.createPageTree(doc)
	if err != nil {
		t.Fatalf("createPageTree() error = %v", err)
	}

	// Should have 2 objects: Pages root + 1 Page
	if len(objects) != 2 {
		t.Errorf("len(objects) = %d, want 2", len(objects))
	}

	// Check Pages root
	pagesRoot := objects[0]
	if pagesRoot.Number != rootRef {
		t.Errorf("Pages root number = %d, want %d", pagesRoot.Number, rootRef)
	}

	rootData := string(pagesRoot.Data)
	if !strings.Contains(rootData, "/Type /Pages") {
		t.Error("Pages root should contain /Type /Pages")
	}

	if !strings.Contains(rootData, "/Count 1") {
		t.Error("Pages root should contain /Count 1")
	}

	// Check individual page
	page := objects[1]
	pageData := string(page.Data)
	if !strings.Contains(pageData, "/Type /Page") {
		t.Error("Page should contain /Type /Page")
	}

	if !strings.Contains(pageData, "/MediaBox") {
		t.Error("Page should contain /MediaBox")
	}
}

func TestCreatePageTree_MultiplePages(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	doc := document.NewDocument()
	for i := 0; i < 5; i++ {
		_, err := doc.AddPage(document.A4)
		if err != nil {
			t.Fatalf("AddPage(%d) error = %v", i, err)
		}
	}

	objects, rootRef, err := w.createPageTree(doc)
	if err != nil {
		t.Fatalf("createPageTree() error = %v", err)
	}

	// Should have 6 objects: Pages root + 5 Pages
	if len(objects) != 6 {
		t.Errorf("len(objects) = %d, want 6", len(objects))
	}

	// Check Pages root
	pagesRoot := objects[0]
	rootData := string(pagesRoot.Data)

	if !strings.Contains(rootData, "/Count 5") {
		t.Errorf("Pages root should contain /Count 5, got: %s", rootData)
	}

	// Count /Kids array entries
	if !strings.Contains(rootData, "/Kids [") {
		t.Error("Pages root should contain /Kids array")
	}

	// Verify all pages have correct structure
	for i := 1; i < len(objects); i++ {
		pageData := string(objects[i].Data)
		if !strings.Contains(pageData, "/Type /Page") {
			t.Errorf("Page %d should contain /Type /Page", i-1)
		}

		if !strings.Contains(pageData, fmt.Sprintf("/Parent %d 0 R", rootRef)) {
			t.Errorf("Page %d should reference parent %d", i-1, rootRef)
		}
	}
}

func TestCreatePagesRoot(t *testing.T) {
	tests := []struct {
		name      string
		objNum    int
		pageRefs  []int
		count     int
		wantCount string
		wantKids  int
	}{
		{
			name:      "single page",
			objNum:    2,
			pageRefs:  []int{3},
			count:     1,
			wantCount: "/Count 1",
			wantKids:  1,
		},
		{
			name:      "three pages",
			objNum:    2,
			pageRefs:  []int{3, 4, 5},
			count:     3,
			wantCount: "/Count 3",
			wantKids:  3,
		},
		{
			name:      "empty pages (edge case)",
			objNum:    2,
			pageRefs:  []int{},
			count:     0,
			wantCount: "/Count 0",
			wantKids:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &PdfWriter{
				nextObjNum: 1,
			}

			obj := w.createPagesRoot(tt.objNum, tt.pageRefs, tt.count)

			if obj.Number != tt.objNum {
				t.Errorf("Object number = %d, want %d", obj.Number, tt.objNum)
			}

			data := string(obj.Data)

			if !strings.Contains(data, "/Type /Pages") {
				t.Error("Should contain /Type /Pages")
			}

			if !strings.Contains(data, tt.wantCount) {
				t.Errorf("Should contain '%s', got: %s", tt.wantCount, data)
			}

			if !strings.Contains(data, "/Kids [") {
				t.Error("Should contain /Kids array")
			}

			// Count references in Kids array
			for _, ref := range tt.pageRefs {
				refStr := fmt.Sprintf("%d 0 R", ref)
				if !strings.Contains(data, refStr) {
					t.Errorf("Kids should contain reference '%s'", refStr)
				}
			}
		})
	}
}

func TestCreatePage(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
	}

	page := document.NewPage(0, document.A4)
	objNum := 3
	parentRef := 2

	obj := w.createPage(page, objNum, parentRef)

	if obj == nil {
		t.Fatal("createPage() returned nil")
	}

	if obj.Number != objNum {
		t.Errorf("Object number = %d, want %d", obj.Number, objNum)
	}

	data := string(obj.Data)

	// Check required page entries
	requiredEntries := []string{
		"/Type /Page",
		"/Parent 2 0 R",
		"/MediaBox",
		"/Resources",
	}

	for _, entry := range requiredEntries {
		if !strings.Contains(data, entry) {
			t.Errorf("Page should contain '%s', got: %s", entry, data)
		}
	}

	// Check MediaBox format [llx lly urx ury]
	if !strings.Contains(data, "[") || !strings.Contains(data, "]") {
		t.Error("MediaBox should be an array with brackets")
	}
}

func TestCreatePage_WithRotation(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
	}

	page := document.NewPage(0, document.A4)
	err := page.SetRotation(90)
	if err != nil {
		t.Fatalf("SetRotation() error = %v", err)
	}

	obj := w.createPage(page, 3, 2)
	data := string(obj.Data)

	if !strings.Contains(data, "/Rotate 90") {
		t.Errorf("Page should contain /Rotate 90, got: %s", data)
	}
}

func TestCreatePage_WithCropBox(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
	}

	page := document.NewPage(0, document.A4)

	// Set crop box (smaller than media box)
	cropBox, err := types.NewRectangle(50, 50, 545, 792)
	if err != nil {
		t.Fatalf("NewRectangle() error = %v", err)
	}

	err = page.SetCropBox(cropBox)
	if err != nil {
		t.Fatalf("SetCropBox() error = %v", err)
	}

	obj := w.createPage(page, 3, 2)
	data := string(obj.Data)

	if !strings.Contains(data, "/CropBox") {
		t.Errorf("Page should contain /CropBox, got: %s", data)
	}

	// Check CropBox has array format
	cropBoxIndex := strings.Index(data, "/CropBox")
	if cropBoxIndex == -1 {
		t.Fatal("CropBox not found")
	}

	// Extract substring after /CropBox
	cropBoxPart := data[cropBoxIndex:]
	if !strings.Contains(cropBoxPart, "[") || !strings.Contains(cropBoxPart, "]") {
		t.Error("CropBox should be an array")
	}
}

func TestCreatePage_DifferentSizes(t *testing.T) {
	tests := []struct {
		name string
		size document.PageSize
	}{
		{"A4", document.A4},
		{"Letter", document.Letter},
		{"Legal", document.Legal},
		{"A3", document.A3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &PdfWriter{
				nextObjNum: 1,
			}

			page := document.NewPage(0, tt.size)
			obj := w.createPage(page, 3, 2)

			if obj == nil {
				t.Fatalf("createPage() returned nil for %s", tt.name)
			}

			data := string(obj.Data)

			// Check MediaBox contains numbers
			if !strings.Contains(data, "/MediaBox [") {
				t.Errorf("Page should contain MediaBox array for %s", tt.name)
			}

			// Extract MediaBox values
			rect := tt.size.ToRectangle()
			llx, _ := rect.LowerLeft()
			urx, _ := rect.UpperRight()

			// Check coordinates are present (with some tolerance for formatting)
			mediaBoxPart := data[strings.Index(data, "/MediaBox"):]
			if !strings.Contains(mediaBoxPart, fmt.Sprintf("%.2f", llx)) {
				t.Errorf("MediaBox should contain llx coordinate %.2f", llx)
			}
			if !strings.Contains(mediaBoxPart, fmt.Sprintf("%.2f", urx)) {
				t.Errorf("MediaBox should contain urx coordinate %.2f", urx)
			}
		})
	}
}

func TestCreateExtGStateObjects(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	resources := NewResourceDictionary()

	// Register two ExtGState entries (placeholders with objNum=0).
	name1, created1 := resources.GetOrCreateExtGState(0.5)
	if !created1 {
		t.Fatal("expected first GetOrCreateExtGState to return needsCreation=true")
	}
	name2, created2 := resources.GetOrCreateExtGState(0.3)
	if !created2 {
		t.Fatal("expected second GetOrCreateExtGState to return needsCreation=true")
	}

	// Before creating objects, both should have placeholder objNum=0.
	if resources.GetExtGStateObjNum(name1) != 0 {
		t.Errorf("expected objNum 0 for %s before creation, got %d", name1, resources.GetExtGStateObjNum(name1))
	}
	if resources.GetExtGStateObjNum(name2) != 0 {
		t.Errorf("expected objNum 0 for %s before creation, got %d", name2, resources.GetExtGStateObjNum(name2))
	}

	// Create ExtGState objects.
	objects := w.createExtGStateObjects(resources)

	// Should have created 2 objects.
	if len(objects) != 2 {
		t.Fatalf("expected 2 ExtGState objects, got %d", len(objects))
	}

	// All object numbers must be > 0.
	for _, obj := range objects {
		if obj.Number <= 0 {
			t.Errorf("ExtGState object has invalid number %d", obj.Number)
		}
		data := string(obj.Data)
		if !strings.Contains(data, "/Type /ExtGState") {
			t.Errorf("ExtGState object missing /Type /ExtGState: %s", data)
		}
		if !strings.Contains(data, "/ca ") || !strings.Contains(data, "/CA ") {
			t.Errorf("ExtGState object missing /ca or /CA: %s", data)
		}
	}

	// After creation, resource dictionary should have real object numbers.
	if resources.GetExtGStateObjNum(name1) == 0 {
		t.Errorf("expected objNum > 0 for %s after creation", name1)
	}
	if resources.GetExtGStateObjNum(name2) == 0 {
		t.Errorf("expected objNum > 0 for %s after creation", name2)
	}

	// Resource dictionary bytes should NOT contain "0 0 R" for any GS entry.
	resStr := resources.String()
	if strings.Contains(resStr, " 0 0 R") {
		t.Errorf("resource dictionary still contains '0 0 R' placeholder: %s", resStr)
	}
}

func TestCreateExtGStateObjects_Empty(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	resources := NewResourceDictionary()

	// No ExtGState entries — should return nil.
	objects := w.createExtGStateObjects(resources)
	if objects != nil {
		t.Errorf("expected nil for empty ExtGState, got %d objects", len(objects))
	}
}

func TestCreateShadingObjects_Linear(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	resources := NewResourceDictionary()
	grad := &GradientOp{
		Type: GradientTypeLinear,
		X1:   0, Y1: 0, X2: 200, Y2: 0,
		ExtendStart: true,
		ExtendEnd:   true,
		ColorStops: []ColorStopOp{
			{Position: 0, Color: RGB{R: 1, G: 0, B: 0}},
			{Position: 1, Color: RGB{R: 0, G: 0, B: 1}},
		},
	}
	shName := resources.AddShading(grad)

	objects := w.createShadingObjects(resources)

	// 2 stops → 1 Type 2 function + 1 Shading dict = 2 objects
	if len(objects) != 2 {
		t.Fatalf("expected 2 objects (function + shading), got %d", len(objects))
	}

	// Verify function object (Type 2).
	funcData := string(objects[0].Data)
	if !strings.Contains(funcData, "/FunctionType 2") {
		t.Errorf("function object missing /FunctionType 2: %s", funcData)
	}
	if !strings.Contains(funcData, "/C0 [1.0000 0.0000 0.0000]") {
		t.Errorf("function object missing /C0 for red: %s", funcData)
	}
	if !strings.Contains(funcData, "/C1 [0.0000 0.0000 1.0000]") {
		t.Errorf("function object missing /C1 for blue: %s", funcData)
	}

	// Verify shading dict.
	shadingData := string(objects[1].Data)
	if !strings.Contains(shadingData, "/ShadingType 2") {
		t.Errorf("shading dict missing /ShadingType 2: %s", shadingData)
	}
	if !strings.Contains(shadingData, "/ColorSpace /DeviceRGB") {
		t.Errorf("shading dict missing /ColorSpace /DeviceRGB: %s", shadingData)
	}
	if !strings.Contains(shadingData, "/Coords [0.00 0.00 200.00 0.00]") {
		t.Errorf("shading dict missing /Coords: %s", shadingData)
	}
	if !strings.Contains(shadingData, "/Extend [true true]") {
		t.Errorf("shading dict missing /Extend: %s", shadingData)
	}

	// Resource dict should have real object number.
	resStr := resources.String()
	if strings.Contains(resStr, fmt.Sprintf("/%s 0 0 R", shName)) {
		t.Errorf("resource dict still has placeholder for %s: %s", shName, resStr)
	}
}

func TestCreateShadingObjects_Radial(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	resources := NewResourceDictionary()
	grad := &GradientOp{
		Type: GradientTypeRadial,
		X0:   100, Y0: 100, R0: 0,
		X1: 100, Y1: 100, R1: 50,
		ExtendStart: true,
		ExtendEnd:   true,
		ColorStops: []ColorStopOp{
			{Position: 0, Color: RGB{R: 1, G: 1, B: 1}},
			{Position: 1, Color: RGB{R: 0, G: 0, B: 1}},
		},
	}
	resources.AddShading(grad)

	objects := w.createShadingObjects(resources)
	if len(objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(objects))
	}

	shadingData := string(objects[1].Data)
	if !strings.Contains(shadingData, "/ShadingType 3") {
		t.Errorf("radial shading dict missing /ShadingType 3: %s", shadingData)
	}
	if !strings.Contains(shadingData, "/Coords [100.00 100.00 0.00 100.00 100.00 50.00]") {
		t.Errorf("radial shading dict missing /Coords: %s", shadingData)
	}
}

func TestCreateShadingObjects_MultiStop(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	resources := NewResourceDictionary()
	grad := &GradientOp{
		Type: GradientTypeLinear,
		X1:   0, Y1: 0, X2: 300, Y2: 0,
		ExtendStart: true,
		ExtendEnd:   true,
		ColorStops: []ColorStopOp{
			{Position: 0.0, Color: RGB{R: 1, G: 0, B: 0}},   // Red
			{Position: 0.5, Color: RGB{R: 1, G: 1, B: 0}},   // Yellow
			{Position: 1.0, Color: RGB{R: 0, G: 0.5, B: 0}}, // Green
		},
	}
	resources.AddShading(grad)

	objects := w.createShadingObjects(resources)

	// 3 stops → 2 Type 2 functions + 1 Type 3 stitch + 1 Shading dict = 4 objects
	if len(objects) != 4 {
		t.Fatalf("expected 4 objects (2 funcs + stitch + shading), got %d", len(objects))
	}

	// Verify first Type 2 function (red → yellow).
	func0Data := string(objects[0].Data)
	if !strings.Contains(func0Data, "/FunctionType 2") {
		t.Errorf("first function missing /FunctionType 2: %s", func0Data)
	}

	// Verify second Type 2 function (yellow → green).
	func1Data := string(objects[1].Data)
	if !strings.Contains(func1Data, "/FunctionType 2") {
		t.Errorf("second function missing /FunctionType 2: %s", func1Data)
	}

	// Verify Type 3 stitching function.
	stitchData := string(objects[2].Data)
	if !strings.Contains(stitchData, "/FunctionType 3") {
		t.Errorf("stitching function missing /FunctionType 3: %s", stitchData)
	}
	if !strings.Contains(stitchData, "/Bounds [0.5000]") {
		t.Errorf("stitching function missing /Bounds: %s", stitchData)
	}
	if !strings.Contains(stitchData, "/Encode [0 1 0 1]") {
		t.Errorf("stitching function missing /Encode: %s", stitchData)
	}

	// Verify shading dict references the stitch function.
	shadingData := string(objects[3].Data)
	if !strings.Contains(shadingData, "/ShadingType 2") {
		t.Errorf("shading dict missing /ShadingType 2: %s", shadingData)
	}
	stitchObjNum := objects[2].Number
	if !strings.Contains(shadingData, fmt.Sprintf("/Function %d 0 R", stitchObjNum)) {
		t.Errorf("shading dict should reference stitch function %d: %s", stitchObjNum, shadingData)
	}
}

func TestCreateShadingObjects_Empty(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	resources := NewResourceDictionary()
	objects := w.createShadingObjects(resources)
	if objects != nil {
		t.Errorf("expected nil for empty shadings, got %d objects", len(objects))
	}
}

func TestCreatePageTree_EmptyDocument(t *testing.T) {
	w := &PdfWriter{
		nextObjNum: 1,
		objects:    make([]*IndirectObject, 0),
		offsets:    make(map[int]int64),
	}

	// Create document with no pages
	doc := document.NewDocument()

	// Empty document should create valid page tree with /Count 0
	// This is valid PDF behavior - documents can have 0 pages
	objects, _, err := w.createPageTree(doc)
	if err != nil {
		t.Errorf("createPageTree() unexpected error for empty document: %v", err)
	}

	// Should have at least the Pages root object
	if len(objects) < 1 {
		t.Error("createPageTree() should return at least the Pages root object")
	}
}
