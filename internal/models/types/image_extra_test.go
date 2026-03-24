package types

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"bytes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createSmallJPEG returns a minimal JPEG byte slice for testing.
func createSmallJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}))
	return buf.Bytes()
}

// TestImage_Data covers the Data() copy method.
func TestImage_Data(t *testing.T) {
	rawData := []byte{1, 2, 3, 4, 5, 6}
	img, err := NewImage(rawData, 2, 1, "DeviceRGB", 8, "/FlateDecode")
	require.NoError(t, err)

	got := img.Data()
	assert.Equal(t, rawData, got)

	// Ensure it's a copy (mutating the result does not affect the image).
	got[0] = 99
	second := img.Data()
	assert.Equal(t, byte(1), second[0], "Data() should return a copy, not a reference")
}

// TestImage_SaveToFile_JPEG covers the SaveToFile path for JPEG images.
func TestImage_SaveToFile_JPEG(t *testing.T) {
	jpegData := createSmallJPEG(t)
	img, err := NewImage(jpegData, 4, 4, "DeviceRGB", 8, "/DCTDecode")
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "out.jpg")

	err = img.SaveToFile(path)
	require.NoError(t, err)

	info, statErr := os.Stat(path)
	require.NoError(t, statErr)
	assert.Greater(t, info.Size(), int64(0))
}
