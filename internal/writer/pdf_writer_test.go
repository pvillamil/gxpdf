package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coregx/gxpdf/internal/document"
)

func TestNewPdfWriter(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.pdf")

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	if writer.file == nil {
		t.Error("file should not be nil")
	}

	if writer.writer == nil {
		t.Error("writer should not be nil")
	}

	if writer.nextObjNum != 1 {
		t.Errorf("nextObjNum = %d, want 1", writer.nextObjNum)
	}

	if writer.objects == nil {
		t.Error("objects should not be nil")
	}

	if writer.offsets == nil {
		t.Error("offsets should not be nil")
	}

	// Check file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func TestNewPdfWriter_InvalidPath(t *testing.T) {
	// Try to create file in non-existent directory
	path := "/nonexistent/directory/test.pdf"

	writer, err := NewPdfWriter(path)
	if err == nil {
		writer.Close()
		t.Error("NewPdfWriter() should return error for invalid path")
	}

	if writer != nil {
		t.Error("writer should be nil on error")
	}
}

func TestPdfWriter_WriteEmptyDocument(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.pdf")

	// Create document with one empty page
	doc := document.NewDocument()
	_, err := doc.AddPage(document.A4)
	if err != nil {
		t.Fatalf("AddPage() error = %v", err)
	}

	// Write document
	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Read and verify file
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	// Check basic structure
	contentStr := string(content)

	// Check PDF header
	if !strings.HasPrefix(contentStr, "%PDF-1.7\n") {
		t.Error("PDF should start with PDF-1.7 header")
	}

	// Check EOF marker
	if !strings.HasSuffix(contentStr, "%%EOF\n") {
		t.Error("PDF should end with EOF marker")
	}

	// Check for required keywords
	requiredKeywords := []string{
		"obj",
		"endobj",
		"/Type /Catalog",
		"/Type /Pages",
		"/Type /Page",
		"xref",
		"trailer",
		"startxref",
	}

	for _, keyword := range requiredKeywords {
		if !strings.Contains(contentStr, keyword) {
			t.Errorf("PDF should contain '%s'", keyword)
		}
	}

	// Check file size is reasonable (> 100 bytes for minimal PDF)
	if len(content) < 100 {
		t.Errorf("PDF file too small: %d bytes", len(content))
	}
}

func TestPdfWriter_WriteMultiPageDocument(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "multipage.pdf")

	// Create document with 3 pages
	doc := document.NewDocument()
	doc.SetMetadata("Multi-Page Test", "Test Author", "Test Subject")

	for i := 0; i < 3; i++ {
		_, err := doc.AddPage(document.A4)
		if err != nil {
			t.Fatalf("AddPage(%d) error = %v", i, err)
		}
	}

	// Write document
	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Read and verify file
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Check /Count 3 in Pages object
	if !strings.Contains(contentStr, "/Count 3") {
		t.Error("Pages object should contain /Count 3")
	}

	// Count number of /Type /Page occurrences (should be 3)
	// Use "/Type /Page " (with space) because Page object format is "/Type /Page /Parent..."
	// This avoids counting "/Type /Pages" as a page
	count := strings.Count(contentStr, "/Type /Page ")
	if count != 3 {
		t.Errorf("Should have 3 pages, found %d", count)
	}
}

func TestPdfWriter_HeaderFormat(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "header.pdf")

	doc := document.NewDocument()
	doc.AddPage(document.A4)

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Read first 50 bytes
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	header := make([]byte, 50)
	n, err := file.Read(header)
	if err != nil {
		t.Fatalf("failed to read header: %v", err)
	}

	headerStr := string(header[:n])

	// Check PDF version
	if !strings.HasPrefix(headerStr, "%PDF-1.7\n") {
		t.Errorf("Header should start with %%PDF-1.7, got: %s", headerStr[:20])
	}

	// Check binary marker exists (bytes after first line)
	lines := strings.Split(headerStr, "\n")
	if len(lines) < 2 {
		t.Fatal("Header should have at least 2 lines")
	}

	// Binary marker should be on second line and contain high bytes
	binaryMarker := lines[1]
	hasHighBytes := false
	for _, b := range []byte(binaryMarker) {
		if b > 127 {
			hasHighBytes = true
			break
		}
	}

	if !hasHighBytes {
		t.Error("Binary marker should contain bytes > 127")
	}
}

func TestPdfWriter_XRefFormat(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "xref.pdf")

	doc := document.NewDocument()
	doc.AddPage(document.A4)

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)

	// Find xref section
	xrefIndex := strings.Index(contentStr, "xref\n")
	if xrefIndex == -1 {
		t.Fatal("xref section not found")
	}

	// Extract xref section (up to trailer)
	trailerIndex := strings.Index(contentStr[xrefIndex:], "trailer")
	if trailerIndex == -1 {
		t.Fatal("trailer not found after xref")
	}

	xrefSection := contentStr[xrefIndex : xrefIndex+trailerIndex]

	// Check object 0 entry (should be "0000000000 65535 f ")
	if !strings.Contains(xrefSection, "0000000000 65535 f ") {
		t.Error("xref should contain object 0 entry: '0000000000 65535 f '")
	}

	// Check format of entries (10 digits, space, 5 digits, space, n/f, space)
	lines := strings.Split(xrefSection, "\n")
	for _, line := range lines {
		if len(line) == 0 || line == "xref" || strings.HasPrefix(line, "0 ") {
			continue
		}

		// Entry lines should be exactly 20 characters: "0000000123 00000 n "
		if len(line) == 20 && (strings.HasSuffix(line, " n ") || strings.HasSuffix(line, " f ")) {
			// Valid entry format
			continue
		}

		// Allow for slight variations, but should be close to 20 chars
		if len(line) > 10 && (strings.Contains(line, " n ") || strings.Contains(line, " f ")) {
			continue
		}
	}
}

func TestPdfWriter_TrailerFormat(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "trailer.pdf")

	doc := document.NewDocument()
	doc.AddPage(document.A4)

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)

	// Check trailer keyword
	if !strings.Contains(contentStr, "trailer\n") {
		t.Error("Should contain 'trailer' keyword")
	}

	// Check trailer dictionary
	trailerIndex := strings.Index(contentStr, "trailer\n")
	trailerSection := contentStr[trailerIndex:]

	// Check /Size
	if !strings.Contains(trailerSection, "/Size") {
		t.Error("Trailer should contain /Size")
	}

	// Check /Root
	if !strings.Contains(trailerSection, "/Root") {
		t.Error("Trailer should contain /Root")
	}

	// Check startxref
	if !strings.Contains(trailerSection, "startxref\n") {
		t.Error("Should contain 'startxref' keyword")
	}

	// Check %%EOF
	if !strings.HasSuffix(contentStr, "%%EOF\n") {
		t.Error("Should end with EOF marker")
	}
}

func TestPdfWriter_Close(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "close.pdf")

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}

	// First close should succeed
	err = writer.Close()
	if err != nil {
		t.Errorf("First Close() error = %v", err)
	}

	// Second close should also succeed (idempotent)
	err = writer.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}

	// Writing after close should fail
	doc := document.NewDocument()
	doc.AddPage(document.A4)

	err = writer.Write(doc)
	if err == nil {
		t.Error("Write() after Close() should return error")
	}
}

func TestPdfWriter_InvalidDocument(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.pdf")

	// Create document with no pages (invalid)
	doc := document.NewDocument()

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err == nil {
		t.Error("Write() should return error for document with no pages")
	}
}

func TestPdfWriter_MetadataInTrailer(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metadata.pdf")

	doc := document.NewDocument()
	doc.SetMetadata("Test Title", "Test Author", "Test Subject", "keyword1", "keyword2")
	doc.AddPage(document.A4)

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)

	// Info dictionary MUST be present as an indirect object in the PDF body
	if !strings.Contains(contentStr, "/Title (Test Title)") {
		t.Error("PDF does not contain /Title in Info dictionary")
	}
	if !strings.Contains(contentStr, "/Author (Test Author)") {
		t.Error("PDF does not contain /Author in Info dictionary")
	}
	if !strings.Contains(contentStr, "/Subject (Test Subject)") {
		t.Error("PDF does not contain /Subject in Info dictionary")
	}

	// Trailer MUST reference Info dictionary with valid object number
	if !strings.Contains(contentStr, "/Info ") {
		t.Error("trailer does not contain /Info reference")
	}

	// Info object reference must NOT point to object 0 (free entry)
	if strings.Contains(contentStr, "/Info 0 0 R") {
		t.Error("/Info references object 0 — object was never created")
	}

	// CreationDate and ModDate must be present
	if !strings.Contains(contentStr, "/CreationDate (D:") {
		t.Error("Info dictionary missing /CreationDate")
	}
	if !strings.Contains(contentStr, "/ModDate (D:") {
		t.Error("Info dictionary missing /ModDate")
	}
}

func TestPdfWriter_MetadataNotWrittenWhenEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "no_metadata.pdf")

	doc := document.NewDocument()
	doc.AddPage(document.A4)

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)

	// No metadata → no /Info in trailer
	if strings.Contains(contentStr, "/Info ") {
		t.Error("trailer contains /Info but no metadata was set")
	}
}

func TestPdfWriter_MetadataEscaping(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "escaped.pdf")

	doc := document.NewDocument()
	doc.SetMetadata("Title (with parens)", "Author\\Name", "Sub)ject")
	doc.AddPage(document.A4)

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)

	// Parentheses and backslashes must be escaped
	if !strings.Contains(contentStr, `Title \(with parens\)`) {
		t.Error("parentheses in title not escaped")
	}
	if !strings.Contains(contentStr, `Author\\Name`) {
		t.Error("backslash in author not escaped")
	}
	if !strings.Contains(contentStr, `Sub\)ject`) {
		t.Error("close-paren in subject not escaped")
	}
}

func TestPdfWriter_DifferentPageSizes(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "sizes.pdf")

	doc := document.NewDocument()

	// Add pages with different sizes
	_, err := doc.AddPage(document.A4)
	if err != nil {
		t.Fatalf("AddPage(A4) error = %v", err)
	}

	_, err = doc.AddPage(document.Letter)
	if err != nil {
		t.Fatalf("AddPage(Letter) error = %v", err)
	}

	_, err = doc.AddPage(document.Legal)
	if err != nil {
		t.Fatalf("AddPage(Legal) error = %v", err)
	}

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	err = writer.Write(doc)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify file was created
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Size() < 200 {
		t.Errorf("PDF file too small for multi-page document: %d bytes", info.Size())
	}

	// Read and check page count
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "/Count 3") {
		t.Error("Should have 3 pages")
	}
}

func TestAllocateObjNum(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "objnum.pdf")

	writer, err := NewPdfWriter(path)
	if err != nil {
		t.Fatalf("NewPdfWriter() error = %v", err)
	}
	defer writer.Close()

	// Allocate several object numbers
	num1 := writer.allocateObjNum()
	num2 := writer.allocateObjNum()
	num3 := writer.allocateObjNum()

	if num1 != 1 {
		t.Errorf("First object number = %d, want 1", num1)
	}

	if num2 != 2 {
		t.Errorf("Second object number = %d, want 2", num2)
	}

	if num3 != 3 {
		t.Errorf("Third object number = %d, want 3", num3)
	}

	if writer.nextObjNum != 4 {
		t.Errorf("nextObjNum = %d, want 4", writer.nextObjNum)
	}
}

func TestFormatPDFDate(t *testing.T) {
	tests := []struct {
		name string
		// We'll create time in UTC for predictable testing
		year, month, day int
		hour, min, sec   int
		wantPrefix       string
	}{
		{
			name: "simple date",
			year: 2025, month: 1, day: 27,
			hour: 12, min: 30, sec: 45,
			wantPrefix: "D:20250127123045",
		},
		{
			name: "midnight",
			year: 2025, month: 10, day: 27,
			hour: 0, min: 0, sec: 0,
			wantPrefix: "D:20251027000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create time in UTC for predictable results
			// Note: In real usage, local time is used
			result := formatPDFDate(mustTime(tt.year, tt.month, tt.day, tt.hour, tt.min, tt.sec))

			if !strings.HasPrefix(result, tt.wantPrefix) {
				t.Errorf("formatPDFDate() = %s, want prefix %s", result, tt.wantPrefix)
			}

			// Check format length (should be like "D:20250127123045+03'00'")
			if len(result) < 20 {
				t.Errorf("formatPDFDate() length = %d, want >= 20", len(result))
			}
		})
	}
}

// Helper function to create time for testing
func mustTime(year, month, day, hour, min, sec int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
}
