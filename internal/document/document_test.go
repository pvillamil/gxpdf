package document

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument()

	assert.NotEmpty(t, doc.id, "document should have an ID")
	assert.Equal(t, "1.7", doc.version.String(), "should default to PDF 1.7")
	assert.Equal(t, "gxpdf", doc.creator)
	assert.Equal(t, "gxpdf (github.com/coregx/gxpdf)", doc.producer)
	assert.NotZero(t, doc.creationDate)
	assert.NotZero(t, doc.modDate)
	assert.Empty(t, doc.pages, "new document should have no pages")
}

func TestDocument_AddPage(t *testing.T) {
	doc := NewDocument()

	page1, err := doc.AddPage(A4)
	require.NoError(t, err)
	assert.NotNil(t, page1)
	assert.Equal(t, 0, page1.Number(), "first page should be number 0")
	assert.Equal(t, 1, doc.PageCount())

	page2, err := doc.AddPage(Letter)
	require.NoError(t, err)
	assert.NotNil(t, page2)
	assert.Equal(t, 1, page2.Number(), "second page should be number 1")
	assert.Equal(t, 2, doc.PageCount())
}

func TestDocument_AddPageWithRect(t *testing.T) {
	doc := NewDocument()

	// Custom 6×9 inch page (432×648 pt)
	rect := CustomPageSize(432, 648)
	page, err := doc.AddPageWithRect(rect)
	require.NoError(t, err)
	require.NotNil(t, page)

	assert.Equal(t, 432.0, page.Width(), "width must match custom rect")
	assert.Equal(t, 648.0, page.Height(), "height must match custom rect")
	assert.Equal(t, 1, doc.PageCount(), "page count incremented")
	assert.Equal(t, 0, page.Number(), "first page is number 0")

	// Second page with different dimensions
	rect2 := CustomPageSize(842, 595)
	page2, err := doc.AddPageWithRect(rect2)
	require.NoError(t, err)
	assert.Equal(t, 842.0, page2.Width())
	assert.Equal(t, 595.0, page2.Height())
	assert.Equal(t, 2, doc.PageCount())

	// ModDate should be updated
	assert.False(t, doc.ModificationDate().IsZero(), "modDate should be set")
}

func TestDocument_InsertPage(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Document) // Add initial pages
		index     int
		wantError bool
	}{
		{
			name:      "insert at beginning",
			setup:     func(d *Document) { d.AddPage(A4); d.AddPage(A4) },
			index:     0,
			wantError: false,
		},
		{
			name:      "insert in middle",
			setup:     func(d *Document) { d.AddPage(A4); d.AddPage(A4) },
			index:     1,
			wantError: false,
		},
		{
			name:      "insert at end",
			setup:     func(d *Document) { d.AddPage(A4); d.AddPage(A4) },
			index:     2,
			wantError: false,
		},
		{
			name:      "insert in empty document",
			setup:     func(d *Document) {},
			index:     0,
			wantError: false,
		},
		{
			name:      "negative index",
			setup:     func(d *Document) { d.AddPage(A4) },
			index:     -1,
			wantError: true,
		},
		{
			name:      "index too large",
			setup:     func(d *Document) { d.AddPage(A4) },
			index:     2,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument()
			tt.setup(doc)
			initialCount := doc.PageCount()

			page, err := doc.InsertPage(tt.index, A4)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, page)
				assert.Equal(t, initialCount, doc.PageCount(), "page count should not change on error")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, page)
				assert.Equal(t, tt.index, page.Number(), "inserted page should have correct number")
				assert.Equal(t, initialCount+1, doc.PageCount())

				// Verify page numbering is correct
				for i := 0; i < doc.PageCount(); i++ {
					p, _ := doc.Page(i)
					assert.Equal(t, i, p.Number(), "page %d should have number %d", i, i)
				}
			}
		})
	}
}

func TestDocument_RemovePage(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Document) // Add initial pages
		index     int
		wantError bool
	}{
		{
			name:      "remove first page",
			setup:     func(d *Document) { d.AddPage(A4); d.AddPage(A4); d.AddPage(A4) },
			index:     0,
			wantError: false,
		},
		{
			name:      "remove middle page",
			setup:     func(d *Document) { d.AddPage(A4); d.AddPage(A4); d.AddPage(A4) },
			index:     1,
			wantError: false,
		},
		{
			name:      "remove last page",
			setup:     func(d *Document) { d.AddPage(A4); d.AddPage(A4); d.AddPage(A4) },
			index:     2,
			wantError: false,
		},
		{
			name:      "negative index",
			setup:     func(d *Document) { d.AddPage(A4) },
			index:     -1,
			wantError: true,
		},
		{
			name:      "index too large",
			setup:     func(d *Document) { d.AddPage(A4) },
			index:     1,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument()
			tt.setup(doc)
			initialCount := doc.PageCount()

			err := doc.RemovePage(tt.index)

			if tt.wantError {
				assert.Error(t, err)
				assert.Equal(t, initialCount, doc.PageCount(), "page count should not change on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, initialCount-1, doc.PageCount())

				// Verify page numbering is correct
				for i := 0; i < doc.PageCount(); i++ {
					p, _ := doc.Page(i)
					assert.Equal(t, i, p.Number(), "page %d should have number %d", i, i)
				}
			}
		})
	}
}

func TestDocument_Page(t *testing.T) {
	doc := NewDocument()
	doc.AddPage(A4)
	doc.AddPage(Letter)
	doc.AddPage(Legal)

	tests := []struct {
		name      string
		index     int
		wantError bool
	}{
		{name: "first page", index: 0, wantError: false},
		{name: "middle page", index: 1, wantError: false},
		{name: "last page", index: 2, wantError: false},
		{name: "negative index", index: -1, wantError: true},
		{name: "index too large", index: 3, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := doc.Page(tt.index)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, page)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, page)
				assert.Equal(t, tt.index, page.Number())
			}
		})
	}
}

func TestDocument_Pages(t *testing.T) {
	doc := NewDocument()
	doc.AddPage(A4)
	doc.AddPage(Letter)
	doc.AddPage(Legal)

	pages := doc.Pages()
	assert.Len(t, pages, 3)

	// Verify it's a copy (modifying returned slice doesn't affect document)
	pages[0] = nil
	p, _ := doc.Page(0)
	assert.NotNil(t, p, "original page should not be affected")
}

func TestDocument_SetMetadata(t *testing.T) {
	doc := NewDocument()

	// Set all metadata
	doc.SetMetadata("Test Document", "John Doe", "Testing", "test", "pdf", "unit")

	assert.Equal(t, "Test Document", doc.Title())
	assert.Equal(t, "John Doe", doc.Author())
	assert.Equal(t, "Testing", doc.Subject())
	assert.Equal(t, []string{"test", "pdf", "unit"}, doc.Keywords())

	// Set partial metadata (should not overwrite existing)
	doc.SetMetadata("New Title", "", "")
	assert.Equal(t, "New Title", doc.Title())
	assert.Equal(t, "John Doe", doc.Author(), "author should remain unchanged")
	assert.Equal(t, "Testing", doc.Subject(), "subject should remain unchanged")
}

func TestDocument_ModificationDate(t *testing.T) {
	doc := NewDocument()
	initialModDate := doc.ModificationDate()

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Modify document
	doc.AddPage(A4)
	newModDate := doc.ModificationDate()

	assert.True(t, newModDate.After(initialModDate), "modification date should be updated")
}

func TestDocument_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Document
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid document",
			setup: func() *Document {
				doc := NewDocument()
				doc.AddPage(A4)
				return doc
			},
			wantError: false,
		},
		{
			name: "empty document",
			setup: func() *Document {
				return NewDocument()
			},
			wantError: true,
			errorMsg:  "document has no pages",
		},
		{
			name: "document with nil page",
			setup: func() *Document {
				doc := NewDocument()
				doc.pages = append(doc.pages, nil)
				return doc
			},
			wantError: true,
			errorMsg:  "page 0 is nil",
		},
		{
			name: "document with invalid page",
			setup: func() *Document {
				doc := NewDocument()
				page := NewPage(0, A4)
				page.rotation = 45 // Invalid rotation
				doc.pages = append(doc.pages, page)
				return doc
			},
			wantError: true,
			errorMsg:  "page 0 validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.setup()
			err := doc.Validate()

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocument_Getters(t *testing.T) {
	doc := NewDocument()

	// Test all getters return expected default values
	assert.Equal(t, "1.7", doc.Version().String())
	assert.Equal(t, "gxpdf", doc.Creator())
	assert.Equal(t, "gxpdf (github.com/coregx/gxpdf)", doc.Producer())
	assert.NotZero(t, doc.CreationDate())
	assert.NotZero(t, doc.ModificationDate())
}
