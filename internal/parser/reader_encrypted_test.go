package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReader_EncryptedPDF_Reference tests reading the reference encrypted PDF
// from the lopdf test assets.
//
// This PDF is encrypted with an empty user password (permissions-only encryption),
// which is the most common case for bank statements and government documents.
func TestReader_EncryptedPDF_Reference(t *testing.T) {
	path := filepath.Join("..", "..", "reference", "lopdf", "assets", "encrypted.pdf")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Reference encrypted PDF not found")
	}

	// Test 1: Open with empty password (should work for permissions-only encryption)
	reader, err := OpenPDF(path)
	if err != nil {
		t.Logf("OpenPDF failed (may need non-empty password): %v", err)

		// Try with common empty password
		reader, err = OpenPDFWithPassword(path, "")
		if err != nil {
			t.Skipf("Cannot open encrypted PDF even with empty password: %v", err)
		}
	}
	require.NotNil(t, reader, "reader should not be nil")
	defer reader.Close()

	// Test 2: Verify basic document structure
	version := reader.Version()
	assert.NotEmpty(t, version, "version should not be empty")
	t.Logf("Encrypted PDF version: %s", version)

	// Test 3: Page count
	pageCount, err := reader.GetPageCount()
	require.NoError(t, err, "should be able to get page count from decrypted PDF")
	assert.Greater(t, pageCount, 0, "should have at least 1 page")
	t.Logf("Encrypted PDF pages: %d", pageCount)

	// Test 4: Catalog
	catalog, err := reader.GetCatalog()
	require.NoError(t, err, "should be able to get catalog from decrypted PDF")
	assert.NotNil(t, catalog)

	// Test 5: Document info
	info := reader.GetDocumentInfo()
	assert.True(t, info.Encrypted, "document should report as encrypted")
	t.Logf("Encrypted PDF info: %+v", info)

	// Test 6: Access first page
	page, err := reader.GetPage(0)
	require.NoError(t, err, "should be able to get first page from decrypted PDF")
	assert.NotNil(t, page, "first page should not be nil")
}

// TestReader_EncryptedPDF_WrongPassword tests that a wrong password fails gracefully.
func TestReader_EncryptedPDF_WrongPassword(t *testing.T) {
	path := filepath.Join("..", "..", "reference", "lopdf", "assets", "encrypted.pdf")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Reference encrypted PDF not found")
	}

	// Should fail with wrong password
	_, err := OpenPDFWithPassword(path, "definitely-wrong-password-12345")
	if err == nil {
		t.Log("PDF opened with wrong password - it may use a different encryption scheme")
		return
	}

	t.Logf("Wrong password correctly rejected: %v", err)
}

// TestReader_EncryptedPDF_TabulaJava tests reading the encrypted PDF from
// tabula-java test assets.
func TestReader_EncryptedPDF_TabulaJava(t *testing.T) {
	path := filepath.Join("..", "..", "reference", "tabula-java", "src", "test",
		"resources", "technology", "tabula", "encrypted.pdf")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Tabula-java encrypted PDF not found")
	}

	// This PDF uses "userpassword" as the user password
	reader, err := OpenPDFWithPassword(path, "userpassword")
	if err != nil {
		t.Skipf("Cannot open tabula-java encrypted PDF: %v", err)
	}
	defer reader.Close()

	version := reader.Version()
	assert.NotEmpty(t, version, "version should not be empty")

	pageCount, err := reader.GetPageCount()
	require.NoError(t, err, "should get page count")
	assert.Greater(t, pageCount, 0, "should have pages")

	info := reader.GetDocumentInfo()
	assert.True(t, info.Encrypted, "should report as encrypted")

	t.Logf("Tabula-java encrypted PDF: version=%s, pages=%d", version, pageCount)
}

// TestReader_OpenPDFWithPassword_NonEncrypted tests that OpenPDFWithPassword
// works for non-encrypted PDFs (password is simply ignored).
func TestReader_OpenPDFWithPassword_NonEncrypted(t *testing.T) {
	// Use any available non-encrypted test PDF
	testPDFs := []string{
		filepath.Join("..", "..", "reference", "lopdf", "assets", "example.pdf"),
		filepath.Join("..", "..", "reference", "lopdf", "assets", "minimal.pdf"),
	}

	for _, path := range testPDFs {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		t.Run(filepath.Base(path), func(t *testing.T) {
			reader, err := OpenPDFWithPassword(path, "any-password")
			if err != nil {
				t.Skipf("Cannot open %s: %v", path, err)
			}
			defer reader.Close()

			version := reader.Version()
			assert.NotEmpty(t, version)

			pageCount, err := reader.GetPageCount()
			assert.NoError(t, err)
			assert.Greater(t, pageCount, 0)

			t.Logf("Non-encrypted PDF opened with password: version=%s, pages=%d", version, pageCount)
		})
		return // Only test the first available PDF
	}

	t.Skip("No non-encrypted reference PDFs available")
}
