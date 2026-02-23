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
