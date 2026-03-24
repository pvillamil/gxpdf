package reader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/coregx/gxpdf/internal/reader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPDFPath returns the path to the minimal test PDF.
func testPDFPath(t *testing.T) string {
	t.Helper()
	// Walk up from internal/reader to project root then into testdata.
	path := filepath.Join("..", "..", "testdata", "pdfs", "minimal.pdf")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("testdata PDF not found: %v", err)
	}
	return path
}

func TestNewPdfReader_Success(t *testing.T) {
	path := testPDFPath(t)
	r, err := reader.NewPdfReader(path)
	require.NoError(t, err)
	require.NotNil(t, r)
	defer func() { assert.NoError(t, r.Close()) }()
}

func TestNewPdfReader_NotFound(t *testing.T) {
	_, err := reader.NewPdfReader("/nonexistent/file.pdf")
	assert.Error(t, err)
}

func TestNewPdfReader_InvalidFile(t *testing.T) {
	f, err := os.CreateTemp("", "notapdf-*.pdf")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("this is not a pdf")
	require.NoError(t, err)
	f.Close()

	_, err = reader.NewPdfReader(f.Name())
	assert.Error(t, err)
}

func TestPdfReader_PageCount(t *testing.T) {
	path := testPDFPath(t)
	r, err := reader.NewPdfReader(path)
	require.NoError(t, err)
	defer r.Close()

	count := r.PageCount()
	assert.Greater(t, count, 0)
}

func TestPdfReader_Version(t *testing.T) {
	path := testPDFPath(t)
	r, err := reader.NewPdfReader(path)
	require.NoError(t, err)
	defer r.Close()

	v := r.Version()
	assert.NotEmpty(t, v)
}

func TestPdfReader_GetPage_FirstPage(t *testing.T) {
	path := testPDFPath(t)
	r, err := reader.NewPdfReader(path)
	require.NoError(t, err)
	defer r.Close()

	page, err := r.GetPage(0)
	require.NoError(t, err)
	assert.NotNil(t, page)
}

func TestPdfReader_GetParserReader(t *testing.T) {
	path := testPDFPath(t)
	r, err := reader.NewPdfReader(path)
	require.NoError(t, err)
	defer r.Close()

	pr := r.GetParserReader()
	assert.NotNil(t, pr)
}

func TestNewPdfReaderWithPassword_UnencryptedPDF(t *testing.T) {
	// An unencrypted PDF should still be openable with an empty password.
	path := testPDFPath(t)
	r, err := reader.NewPdfReaderWithPassword(path, "")
	require.NoError(t, err)
	require.NotNil(t, r)
	defer r.Close()

	assert.Greater(t, r.PageCount(), 0)
}

func TestNewPdfReaderWithPassword_NotFound(t *testing.T) {
	_, err := reader.NewPdfReaderWithPassword("/no/such.pdf", "pass")
	assert.Error(t, err)
}
