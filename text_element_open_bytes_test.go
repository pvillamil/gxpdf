package gxpdf

import (
	"os"
	"strings"
	"testing"
)

// ---------- TextElement struct ----------

func TestTextElement_Fields(t *testing.T) {
	e := TextElement{
		Text:     "Hello",
		X:        10.5,
		Y:        200.0,
		Width:    50.0,
		Height:   12.0,
		FontName: "/F1",
		FontSize: 12.0,
	}
	if e.Text != "Hello" {
		t.Errorf("Text = %q, want %q", e.Text, "Hello")
	}
	if e.X != 10.5 {
		t.Errorf("X = %v, want 10.5", e.X)
	}
	if e.FontSize != 12.0 {
		t.Errorf("FontSize = %v, want 12.0", e.FontSize)
	}
}

// ---------- Page.ExtractTextElements ----------

func TestPage_ExtractTextElements_ReturnsElements(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot open %s: %v", minimalPDF, err)
	}
	defer doc.Close()

	page := doc.Page(0)
	if page == nil {
		t.Fatal("Page(0) returned nil")
	}

	elements, err := page.ExtractTextElements()
	if err != nil {
		t.Fatalf("ExtractTextElements() error = %v", err)
	}
	// minimal.pdf is known to have at least some text
	t.Logf("ExtractTextElements() returned %d elements", len(elements))
}

func TestPage_ExtractTextElements_TextMatchesExtractText(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot open %s: %v", minimalPDF, err)
	}
	defer doc.Close()

	page := doc.Page(0)
	if page == nil {
		t.Fatal("Page(0) returned nil")
	}

	elements, err := page.ExtractTextElements()
	if err != nil {
		t.Fatalf("ExtractTextElements() error = %v", err)
	}

	// Build concatenated text from elements and compare with ExtractText
	var sb strings.Builder
	for _, e := range elements {
		sb.WriteString(e.Text)
		sb.WriteString(" ")
	}
	fromElements := strings.TrimSpace(sb.String())
	fromExtractText := strings.TrimSpace(page.ExtractText())

	if fromElements != fromExtractText {
		t.Errorf("Text from ExtractTextElements %q does not match ExtractText %q",
			fromElements, fromExtractText)
	}
}

func TestPage_ExtractTextElements_PositiveCoordinates(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot open %s: %v", minimalPDF, err)
	}
	defer doc.Close()

	page := doc.Page(0)
	elements, err := page.ExtractTextElements()
	if err != nil {
		t.Fatalf("ExtractTextElements() error = %v", err)
	}

	for i, e := range elements {
		if e.X < 0 {
			t.Errorf("elements[%d].X = %v, want >= 0", i, e.X)
		}
		if e.Y < 0 {
			t.Errorf("elements[%d].Y = %v, want >= 0", i, e.Y)
		}
		if e.Width < 0 {
			t.Errorf("elements[%d].Width = %v, want >= 0", i, e.Width)
		}
		if e.Height < 0 {
			t.Errorf("elements[%d].Height = %v, want >= 0", i, e.Height)
		}
	}
}

func TestPage_ExtractTextElements_FontSizePositive(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot open %s: %v", minimalPDF, err)
	}
	defer doc.Close()

	page := doc.Page(0)
	elements, err := page.ExtractTextElements()
	if err != nil {
		t.Fatalf("ExtractTextElements() error = %v", err)
	}

	for i, e := range elements {
		if e.FontSize < 0 {
			t.Errorf("elements[%d].FontSize = %v, want >= 0", i, e.FontSize)
		}
	}
}

// ---------- Document.ExtractTextElementsFromPage ----------

func TestDocument_ExtractTextElementsFromPage_ValidPage(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot open %s: %v", minimalPDF, err)
	}
	defer doc.Close()

	elements, err := doc.ExtractTextElementsFromPage(1)
	if err != nil {
		t.Fatalf("ExtractTextElementsFromPage(1) error = %v", err)
	}
	t.Logf("Page 1 has %d text elements", len(elements))
}

func TestDocument_ExtractTextElementsFromPage_InvalidPage_Zero(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot open %s: %v", minimalPDF, err)
	}
	defer doc.Close()

	_, err = doc.ExtractTextElementsFromPage(0)
	if err == nil {
		t.Error("ExtractTextElementsFromPage(0) expected error, got nil")
	}
}

func TestDocument_ExtractTextElementsFromPage_InvalidPage_TooHigh(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot open %s: %v", minimalPDF, err)
	}
	defer doc.Close()

	_, err = doc.ExtractTextElementsFromPage(9999)
	if err == nil {
		t.Error("ExtractTextElementsFromPage(9999) expected error, got nil")
	}
}

func TestDocument_ExtractTextElementsFromPage_MatchesPageExtractTextElements(t *testing.T) {
	doc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot open %s: %v", minimalPDF, err)
	}
	defer doc.Close()

	// Via Document convenience method (1-based)
	docElements, err := doc.ExtractTextElementsFromPage(1)
	if err != nil {
		t.Fatalf("ExtractTextElementsFromPage(1) error = %v", err)
	}

	// Via Page method (0-based)
	page := doc.Page(0)
	pageElements, err := page.ExtractTextElements()
	if err != nil {
		t.Fatalf("page.ExtractTextElements() error = %v", err)
	}

	if len(docElements) != len(pageElements) {
		t.Errorf("element count mismatch: Document=%d, Page=%d", len(docElements), len(pageElements))
	}

	for i := 0; i < len(docElements) && i < len(pageElements); i++ {
		if docElements[i].Text != pageElements[i].Text {
			t.Errorf("elements[%d].Text: Document=%q, Page=%q",
				i, docElements[i].Text, pageElements[i].Text)
		}
		if docElements[i].X != pageElements[i].X {
			t.Errorf("elements[%d].X: Document=%v, Page=%v",
				i, docElements[i].X, pageElements[i].X)
		}
	}
}

// ---------- OpenFromBytes ----------

func TestOpenFromBytes_MinimalPDF(t *testing.T) {
	data, err := os.ReadFile(minimalPDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", minimalPDF, err)
	}

	doc, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes() error = %v", err)
	}
	defer doc.Close()

	if doc == nil {
		t.Fatal("OpenFromBytes returned nil document")
	}
}

func TestOpenFromBytes_SameContentAsOpen(t *testing.T) {
	data, err := os.ReadFile(minimalPDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", minimalPDF, err)
	}

	// Open via file
	fileDoc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot Open %s: %v", minimalPDF, err)
	}
	defer fileDoc.Close()

	// Open via bytes
	bytesDoc, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes() error = %v", err)
	}
	defer bytesDoc.Close()

	// Page count must match
	if fileDoc.PageCount() != bytesDoc.PageCount() {
		t.Errorf("PageCount: file=%d, bytes=%d", fileDoc.PageCount(), bytesDoc.PageCount())
	}

	// Version must match
	if fileDoc.Version() != bytesDoc.Version() {
		t.Errorf("Version: file=%q, bytes=%q", fileDoc.Version(), bytesDoc.Version())
	}
}

func TestOpenFromBytes_TextMatchesOpen(t *testing.T) {
	data, err := os.ReadFile(minimalPDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", minimalPDF, err)
	}

	fileDoc, err := Open(minimalPDF)
	if err != nil {
		t.Skipf("cannot Open %s: %v", minimalPDF, err)
	}
	defer fileDoc.Close()

	bytesDoc, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes() error = %v", err)
	}
	defer bytesDoc.Close()

	// Compare text extraction on each page
	for i := 0; i < fileDoc.PageCount(); i++ {
		fileText := fileDoc.Page(i).ExtractText()
		bytesText := bytesDoc.Page(i).ExtractText()
		if fileText != bytesText {
			t.Errorf("page %d text mismatch:\n  file=%q\n  bytes=%q", i, fileText, bytesText)
		}
	}
}

func TestOpenFromBytes_EmptySlice(t *testing.T) {
	_, err := OpenFromBytes([]byte{})
	if err == nil {
		t.Error("OpenFromBytes(empty) expected error, got nil")
	}
}

func TestOpenFromBytes_NotAPDF(t *testing.T) {
	_, err := OpenFromBytes([]byte("this is not a PDF document"))
	if err == nil {
		t.Error("OpenFromBytes(invalid data) expected error, got nil")
	}
}

func TestOpenFromBytes_PathIsBytes(t *testing.T) {
	data, err := os.ReadFile(minimalPDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", minimalPDF, err)
	}

	doc, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes() error = %v", err)
	}
	defer doc.Close()

	// Path sentinel for in-memory documents
	if doc.Path() != pathFromBytes {
		t.Errorf("doc.Path() = %q, want %q", doc.Path(), pathFromBytes)
	}
}

func TestOpenFromBytes_CloseIsNoOp(t *testing.T) {
	data, err := os.ReadFile(minimalPDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", minimalPDF, err)
	}

	doc, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes() error = %v", err)
	}

	// Close multiple times must not panic or error
	if err := doc.Close(); err != nil {
		t.Errorf("first Close() error = %v", err)
	}
	if err := doc.Close(); err != nil {
		t.Errorf("second Close() error = %v", err)
	}
}

func TestOpenFromBytes_MultiPage(t *testing.T) {
	data, err := os.ReadFile(multipagePDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", multipagePDF, err)
	}

	doc, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes() multipage error = %v", err)
	}
	defer doc.Close()

	if doc.PageCount() < 2 {
		t.Skipf("multipage PDF has only %d pages", doc.PageCount())
	}
	t.Logf("OpenFromBytes multipage: %d pages", doc.PageCount())
}

// ---------- OpenFromBytesWithPassword ----------

func TestOpenFromBytesWithPassword_NonEncrypted(t *testing.T) {
	data, err := os.ReadFile(minimalPDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", minimalPDF, err)
	}

	// A non-encrypted PDF opened with any password should succeed
	// (the password is ignored for non-encrypted PDFs)
	doc, err := OpenFromBytesWithPassword(data, "anypassword")
	if err != nil {
		t.Fatalf("OpenFromBytesWithPassword(non-encrypted) error = %v", err)
	}
	defer doc.Close()

	if doc.PageCount() == 0 {
		t.Error("expected at least 1 page")
	}
}

func TestOpenFromBytesWithPassword_EmptyPassword(t *testing.T) {
	data, err := os.ReadFile(minimalPDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", minimalPDF, err)
	}

	doc, err := OpenFromBytesWithPassword(data, "")
	if err != nil {
		t.Fatalf("OpenFromBytesWithPassword(empty password) error = %v", err)
	}
	defer doc.Close()
}

func TestOpenFromBytesWithPassword_InvalidData(t *testing.T) {
	_, err := OpenFromBytesWithPassword([]byte("not a pdf"), "password")
	if err == nil {
		t.Error("OpenFromBytesWithPassword(invalid) expected error, got nil")
	}
}

// ---------- OpenFromBytesWithContext ----------

func TestOpenFromBytesWithContext_Success(t *testing.T) {
	data, err := os.ReadFile(minimalPDF)
	if err != nil {
		t.Skipf("cannot read %s: %v", minimalPDF, err)
	}

	// Uses background context (same as OpenFromBytes internally)
	doc, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes() error = %v", err)
	}
	defer doc.Close()

	if doc.PageCount() == 0 {
		t.Error("expected at least 1 page")
	}
}
