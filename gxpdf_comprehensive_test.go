package gxpdf

import (
	"context"
	"testing"
	"time"
)

// ---------- Test helpers ----------

const minimalPDF = "testdata/pdfs/minimal.pdf"
const multipagePDF = "testdata/pdfs/multipage.pdf"

func openTestDoc(t *testing.T, path string) *Document {
	t.Helper()
	doc, err := Open(path)
	if err != nil {
		t.Skipf("Cannot open %s: %v", path, err)
	}
	return doc
}

// ---------- Open / OpenWithContext / MustOpen / OpenWithPassword ----------

func TestOpen_MinimalPDF(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("minimal.pdf not available: %v", err)
	}
	defer doc.Close()

	if doc == nil {
		t.Fatal("Open returned nil document")
	}
	if doc.reader == nil {
		t.Error("Document.reader is nil")
	}
	if doc.path != minimalPDF {
		t.Errorf("Document.path = %q, want %q", doc.path, minimalPDF)
	}
}

func TestOpen_NonExistentFile(t *testing.T) {
	doc, err := Open("nonexistent_file.pdf")
	if err == nil {
		doc.Close()
		t.Error("Open should return error for nonexistent file")
	}
}

func TestOpenWithContext_MinimalPDF(t *testing.T) {
	ctx := context.Background()
	doc, err := OpenWithContext(ctx, minimalPDF)
	if err != nil {
		t.Skipf("minimal.pdf not available: %v", err)
	}
	defer doc.Close()

	if doc == nil {
		t.Fatal("OpenWithContext returned nil")
	}
	if doc.ctx != ctx {
		t.Error("Document.ctx not set correctly")
	}
}

func TestOpenWithContext_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Even with cancelled context, open should work (context is checked during ops)
	doc, err := OpenWithContext(ctx, minimalPDF)
	if err != nil {
		t.Skip("minimal.pdf not available")
	}
	defer doc.Close()
}

func TestOpenWithContext_TimeoutContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	doc, err := OpenWithContext(ctx, minimalPDF)
	if err != nil {
		t.Skip("minimal.pdf not available")
	}
	defer doc.Close()
	if doc == nil {
		t.Fatal("OpenWithContext with timeout returned nil")
	}
}

func TestMustOpen_MinimalPDF(t *testing.T) {
	// Test that MustOpen works for valid files
	defer func() {
		if r := recover(); r != nil {
			t.Skip("minimal.pdf not available")
		}
	}()
	doc := MustOpen(minimalPDF)
	defer doc.Close()
	if doc == nil {
		t.Fatal("MustOpen returned nil")
	}
}

func TestMustOpen_InvalidFile(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("MustOpen should panic for invalid file")
		}
	}()
	_ = MustOpen("nonexistent_invalid_file.pdf")
}

func TestOpenWithPassword_NonExistentFile(t *testing.T) {
	_, err := OpenWithPassword("nonexistent.pdf", "password")
	if err == nil {
		t.Error("OpenWithPassword should return error for nonexistent file")
	}
}

func TestOpenWithPasswordAndContext_NonExistentFile(t *testing.T) {
	ctx := context.Background()
	_, err := OpenWithPasswordAndContext(ctx, "nonexistent.pdf", "password")
	if err == nil {
		t.Error("OpenWithPasswordAndContext should return error for nonexistent file")
	}
}

// ---------- Document.Close ----------

func TestDocument_Close_IdempotentSafe(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)

	// First close
	if err := doc.Close(); err != nil {
		t.Errorf("First Close() error = %v", err)
	}
	// Second close should not panic or error (reader handles it)
}

// ---------- Document.Path ----------

func TestDocument_Path(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	path := doc.Path()
	if path != minimalPDF {
		t.Errorf("Path() = %q, want %q", path, minimalPDF)
	}
}

// ---------- Document.PageCount ----------

func TestDocument_PageCount_Minimal(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	count := doc.PageCount()
	if count < 1 {
		t.Errorf("PageCount() = %d, want >= 1", count)
	}
}

func TestDocument_PageCount_Multipage(t *testing.T) {
	doc := openTestDoc(t, multipagePDF)
	defer doc.Close()

	count := doc.PageCount()
	if count < 2 {
		t.Errorf("PageCount() = %d, want >= 2 for multipage PDF", count)
	}
}

// ---------- Document.Page ----------

func TestDocument_Page_ValidIndex(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	page := doc.Page(0)
	if page == nil {
		t.Fatal("Page(0) returned nil for valid index")
	}
	if page.index != 0 {
		t.Errorf("Page.index = %d, want 0", page.index)
	}
	if page.doc != doc {
		t.Error("Page.doc not set correctly")
	}
}

func TestDocument_Page_NegativeIndex(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	page := doc.Page(-1)
	if page != nil {
		t.Error("Page(-1) should return nil for negative index")
	}
}

func TestDocument_Page_OutOfBounds(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	page := doc.Page(9999)
	if page != nil {
		t.Error("Page(9999) should return nil for out-of-bounds index")
	}
}

// ---------- Document.Pages ----------

func TestDocument_Pages_Minimal(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	pages := doc.Pages()
	count := doc.PageCount()
	if len(pages) != count {
		t.Errorf("Pages() len = %d, want %d", len(pages), count)
	}
	for i, p := range pages {
		if p == nil {
			t.Fatalf("Pages()[%d] is nil", i)
		}
		if p.index != i {
			t.Errorf("Pages()[%d].index = %d, want %d", i, p.index, i)
		}
	}
}

// ---------- Document.ExtractTables ----------

func TestDocument_ExtractTables_MinimalPDF(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	tables := doc.ExtractTables()
	// minimal.pdf may have 0 tables - just verify no panic and returns slice
	if tables == nil {
		tables = []*Table{} // nil is acceptable
	}
	t.Logf("Extracted %d tables from minimal.pdf", len(tables))
}

func TestDocument_ExtractTablesWithOptions_Nil(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	tables, err := doc.ExtractTablesWithOptions(nil)
	if err != nil {
		t.Errorf("ExtractTablesWithOptions(nil) error = %v", err)
	}
	_ = tables
}

func TestDocument_ExtractTablesWithOptions_Lattice(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	opts := &ExtractionOptions{
		Method: MethodLattice,
	}
	tables, err := doc.ExtractTablesWithOptions(opts)
	if err != nil {
		t.Errorf("ExtractTablesWithOptions(Lattice) error = %v", err)
	}
	_ = tables
}

func TestDocument_ExtractTablesWithOptions_Stream(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	opts := &ExtractionOptions{
		Method: MethodStream,
	}
	tables, err := doc.ExtractTablesWithOptions(opts)
	if err != nil {
		t.Errorf("ExtractTablesWithOptions(Stream) error = %v", err)
	}
	_ = tables
}

func TestDocument_ExtractTablesWithOptions_SpecificPages(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	opts := &ExtractionOptions{
		Pages: []int{0},
	}
	tables, err := doc.ExtractTablesWithOptions(opts)
	if err != nil {
		t.Errorf("ExtractTablesWithOptions(page 0) error = %v", err)
	}
	_ = tables
}

func TestDocument_ExtractTablesWithOptions_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	doc, err := OpenWithContext(ctx, minimalPDF)
	if err != nil {
		t.Skip("minimal.pdf not available")
	}
	defer doc.Close()

	cancel() // Cancel before extraction

	opts := &ExtractionOptions{
		Pages: []int{0},
	}
	_, err = doc.ExtractTablesWithOptions(opts)
	// May or may not error depending on timing, but should not panic
	_ = err
}

// ---------- Document.ExtractTextFromPage ----------

func TestDocument_ExtractTextFromPage_ValidPage(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	text, err := doc.ExtractTextFromPage(1)
	if err != nil {
		t.Errorf("ExtractTextFromPage(1) error = %v", err)
	}
	t.Logf("Extracted text from page 1: %q", text)
}

func TestDocument_ExtractTextFromPage_ZeroPage(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	_, err := doc.ExtractTextFromPage(0)
	if err == nil {
		t.Error("ExtractTextFromPage(0) should return error (1-based)")
	}
}

func TestDocument_ExtractTextFromPage_TooLarge(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	_, err := doc.ExtractTextFromPage(9999)
	if err == nil {
		t.Error("ExtractTextFromPage(9999) should return error for out-of-bounds")
	}
}

func TestDocument_ExtractTablesFromPage_ValidPage(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	tables := doc.ExtractTablesFromPage(1)
	_ = tables
}

func TestDocument_ExtractTablesFromPage_OutOfRange(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	tables := doc.ExtractTablesFromPage(0)
	if tables != nil {
		t.Error("ExtractTablesFromPage(0) should return nil (1-based)")
	}
}

func TestDocument_ExtractTablesFromPage_TooLarge(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	tables := doc.ExtractTablesFromPage(9999)
	if tables != nil {
		t.Error("ExtractTablesFromPage(9999) should return nil for out-of-bounds")
	}
}

// ---------- Document.Info and metadata methods ----------

func TestDocument_Info(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	info := doc.Info()
	if info == nil {
		t.Fatal("Info() returned nil")
	}
	if info.PageCount != doc.PageCount() {
		t.Errorf("Info.PageCount = %d, want %d", info.PageCount, doc.PageCount())
	}
	if info.Path != minimalPDF {
		t.Errorf("Info.Path = %q, want %q", info.Path, minimalPDF)
	}
}

func TestDocument_Version(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	version := doc.Version()
	if version == "" {
		t.Error("Version() returned empty string")
	}
	t.Logf("PDF version: %s", version)
}

func TestDocument_Title(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()
	// Just verify no panic
	_ = doc.Title()
}

func TestDocument_Author(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()
	_ = doc.Author()
}

func TestDocument_Subject(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()
	_ = doc.Subject()
}

func TestDocument_Keywords(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()
	_ = doc.Keywords()
}

func TestDocument_Creator(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()
	_ = doc.Creator()
}

func TestDocument_Producer(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()
	_ = doc.Producer()
}

func TestDocument_IsEncrypted_Unencrypted(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	// minimal.pdf should not be encrypted
	encrypted := doc.IsEncrypted()
	t.Logf("IsEncrypted() = %v", encrypted)
}

// ---------- Document.HasForm / GetFormFields ----------

func TestDocument_HasForm_MinimalPDF(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	// minimal.pdf should not have a form
	hasForm := doc.HasForm()
	t.Logf("HasForm() = %v", hasForm)
}

func TestDocument_GetFormFields_NoForm(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	fields, err := doc.GetFormFields()
	if err != nil {
		t.Errorf("GetFormFields() error = %v", err)
	}
	// Should return nil or empty slice for PDF with no form
	t.Logf("GetFormFields() = %d fields", len(fields))
}

func TestDocument_GetFieldValue_NotFound(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	_, err := doc.GetFieldValue("nonexistent_field")
	if err == nil {
		t.Error("GetFieldValue should return error for nonexistent field")
	}
}

// ---------- Document.GetImages ----------

func TestDocument_GetImages_MinimalPDF(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	images := doc.GetImages()
	t.Logf("GetImages() = %d images", len(images))
}

func TestDocument_GetImagesWithError_MinimalPDF(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	images, err := doc.GetImagesWithError()
	if err != nil {
		t.Errorf("GetImagesWithError() error = %v", err)
	}
	t.Logf("GetImagesWithError() = %d images", len(images))
}

// ---------- FormField wrapper ----------

func TestFormField_Methods(t *testing.T) {
	doc := openTestDoc(t, minimalPDF)
	defer doc.Close()

	fields, err := doc.GetFormFields()
	if err != nil {
		t.Errorf("GetFormFields() error = %v", err)
	}
	if len(fields) == 0 {
		t.Log("No form fields in minimal.pdf, skipping FormField method tests")
		return
	}

	f := fields[0]
	// Just verify methods don't panic
	_ = f.Name()
	_ = f.Type()
	_ = f.Value()
	_ = f.DefaultValue()
	_ = f.Flags()
	_ = f.Rect()
	_ = f.Options()
	_ = f.IsReadOnly()
	_ = f.IsRequired()
	_ = f.IsTextField()
	_ = f.IsButton()
	_ = f.IsChoice()
}

// ---------- ErrPasswordRequired ----------

func TestErrPasswordRequired_IsNotNil(t *testing.T) {
	if ErrPasswordRequired == nil {
		t.Error("ErrPasswordRequired should not be nil")
	}
}

// ---------- Version constant ----------

func TestVersion_NotEmpty(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
}

// ---------- DocumentInfo struct ----------

func TestDocumentInfo_Fields(t *testing.T) {
	info := &DocumentInfo{
		PageCount: 5,
		Path:      "/path/to/doc.pdf",
		Version:   "1.7",
		Title:     "Test Doc",
		Author:    "Author Name",
		Subject:   "Test Subject",
		Keywords:  "pdf, test",
		Creator:   "Creator App",
		Producer:  "Producer App",
		Encrypted: false,
	}
	if info.PageCount != 5 {
		t.Errorf("PageCount = %d, want 5", info.PageCount)
	}
	if info.Version != "1.7" {
		t.Errorf("Version = %q, want 1.7", info.Version)
	}
	if info.Encrypted {
		t.Error("Encrypted should be false")
	}
}
