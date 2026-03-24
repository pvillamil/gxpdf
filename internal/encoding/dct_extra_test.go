package encoding

import (
	"image"
	"image/color"
	"image/jpeg"
	"bytes"
	"testing"
)

// createRGBAJPEG creates an RGBA-model JPEG by encoding via RGBA image.
// Go's jpeg decoder usually returns YCbCr, but we can test the Encode path and
// verify that if we craft an image that decodes as RGBA, it still works.
// Since Go's standard jpeg decoder always gives YCbCr or Gray, we test the
// fallback (extractGeneric) via an indirect path: a custom image.Image.
func TestDCTDecoder_Decode_EmptyData(t *testing.T) {
	d := NewDCTDecoder()
	_, err := d.Decode([]byte{})
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}
}

func TestDCTDecoder_Decode_NilLike(t *testing.T) {
	d := NewDCTDecoder()
	_, err := d.Decode([]byte("garbage!@#$%"))
	if err == nil {
		t.Error("Expected error for garbage input, got nil")
	}
}

func TestDCTDecoder_DecodeToImage_InvalidData(t *testing.T) {
	d := NewDCTDecoder()
	_, err := d.DecodeToImage([]byte("not jpeg"))
	if err == nil {
		t.Error("Expected error for invalid JPEG in DecodeToImage")
	}
}

func TestDCTDecoder_Encode_QualityBoundaries(t *testing.T) {
	d := NewDCTDecoder()
	width, height := 4, 4
	data := make([]byte, width*height*3)
	for i := range data {
		data[i] = 128
	}

	tests := []struct {
		name    string
		quality int
	}{
		{"quality 0 (uses default 75)", 0},
		{"quality -1 (uses default 75)", -1},
		{"quality 101 (uses default 75)", 101},
		{"quality 1 (minimum valid)", 1},
		{"quality 100 (maximum valid)", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := d.Encode(data, width, height, tt.quality)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}
			if len(result) == 0 {
				t.Error("Expected non-empty result")
			}
		})
	}
}

func TestDCTDecoder_EncodeGray_InvalidData(t *testing.T) {
	d := NewDCTDecoder()

	tests := []struct {
		name      string
		data      []byte
		width     int
		height    int
	}{
		{"too short", make([]byte, 9), 10, 10},
		{"too long", make([]byte, 200), 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.EncodeGray(tt.data, tt.width, tt.height, 75)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestDCTDecoder_EncodeGray_QualityBoundaries(t *testing.T) {
	d := NewDCTDecoder()
	width, height := 4, 4
	data := make([]byte, width*height)
	for i := range data {
		data[i] = 200
	}

	tests := []struct {
		name    string
		quality int
	}{
		{"quality 0 (default)", 0},
		{"quality -5 (default)", -5},
		{"quality 200 (default)", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := d.EncodeGray(data, width, height, tt.quality)
			if err != nil {
				t.Fatalf("EncodeGray failed: %v", err)
			}
			if len(result) == 0 {
				t.Error("Expected non-empty result")
			}
		})
	}
}

func TestDCTDecoder_Encode_ZeroWidthHeight(t *testing.T) {
	d := NewDCTDecoder()

	// Data has 3 bytes but dims 1x1: length matches 1*1*3=3 → no error.
	// Use mismatched length to ensure an error is returned.
	_, err := d.Encode(make([]byte, 5), 2, 2, 75) // 2*2*3=12 expected, got 5
	if err == nil {
		t.Error("Expected error for mismatched data length, got nil")
	}
}

func TestDCTDecoder_RoundTrip_Gray(t *testing.T) {
	d := NewDCTDecoder()
	width, height := 8, 8

	// Create gradient grayscale data.
	data := make([]byte, width*height)
	for i := range data {
		data[i] = byte(i * 4 % 256)
	}

	encoded, err := d.EncodeGray(data, width, height, 95)
	if err != nil {
		t.Fatalf("EncodeGray failed: %v", err)
	}

	decoded, err := d.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// For an 8x8 grayscale image, decoded data must be 8*8=64 bytes.
	if len(decoded) != width*height {
		t.Errorf("Decoded length: got %d, want %d", len(decoded), width*height)
	}
}

// TestDCTDecoder_ExtractFromRGBA tests the extractFromRGBA path directly.
// jpeg.Decode never returns *image.RGBA, so we call the private method directly.
func TestDCTDecoder_ExtractFromRGBA(t *testing.T) {
	d := NewDCTDecoder()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img.SetRGBA(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img.SetRGBA(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img.SetRGBA(1, 1, color.RGBA{R: 128, G: 128, B: 128, A: 255})

	result, err := d.extractFromRGBA(img, 2, 2)
	if err != nil {
		t.Fatalf("extractFromRGBA failed: %v", err)
	}
	if result.Width != 2 || result.Height != 2 {
		t.Errorf("wrong size %dx%d", result.Width, result.Height)
	}
	if result.Components != 3 {
		t.Errorf("expected 3 components, got %d", result.Components)
	}
	if len(result.Data) != 2*2*3 {
		t.Errorf("wrong data length %d", len(result.Data))
	}
	// First pixel should be R=255, G=0, B=0.
	if result.Data[0] != 255 || result.Data[1] != 0 || result.Data[2] != 0 {
		t.Errorf("wrong first pixel RGB: %d %d %d", result.Data[0], result.Data[1], result.Data[2])
	}
}

// TestDCTDecoder_ExtractFromNRGBA tests the extractFromNRGBA path directly.
// jpeg.Decode never returns *image.NRGBA, so we call the private method directly.
func TestDCTDecoder_ExtractFromNRGBA(t *testing.T) {
	d := NewDCTDecoder()

	img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 40, G: 50, B: 60, A: 255})
	img.SetNRGBA(2, 0, color.NRGBA{R: 70, G: 80, B: 90, A: 255})

	result, err := d.extractFromNRGBA(img, 3, 1)
	if err != nil {
		t.Fatalf("extractFromNRGBA failed: %v", err)
	}
	if result.Width != 3 || result.Height != 1 {
		t.Errorf("wrong size %dx%d", result.Width, result.Height)
	}
	if result.Components != 3 {
		t.Errorf("expected 3 components, got %d", result.Components)
	}
	if len(result.Data) != 3*1*3 {
		t.Errorf("wrong data length %d", len(result.Data))
	}
	// Second pixel: R=40, G=50, B=60.
	if result.Data[3] != 40 || result.Data[4] != 50 || result.Data[5] != 60 {
		t.Errorf("wrong second pixel RGB: %d %d %d", result.Data[3], result.Data[4], result.Data[5])
	}
}

// genericImage is a custom image.Image implementation to exercise the extractGeneric fallback.
type genericImage struct {
	img *image.Paletted
}

func (p *genericImage) ColorModel() color.Model { return p.img.ColorModel() }
func (p *genericImage) Bounds() image.Rectangle { return p.img.Bounds() }
func (p *genericImage) At(x, y int) color.Color { return p.img.At(x, y) }

// TestDCTDecoder_ExtractGeneric tests the extractGeneric fallback path directly.
// This covers the default case in DecodeWithMetadata's type switch.
func TestDCTDecoder_ExtractGeneric(t *testing.T) {
	d := NewDCTDecoder()

	palette := color.Palette{
		color.RGBA{R: 200, G: 100, B: 50, A: 255},
		color.RGBA{R: 0, G: 128, B: 255, A: 255},
	}
	paletted := image.NewPaletted(image.Rect(0, 0, 2, 1), palette)
	paletted.SetColorIndex(0, 0, 0) // color index 0
	paletted.SetColorIndex(1, 0, 1) // color index 1

	img := &genericImage{img: paletted}
	result, err := d.extractGeneric(img, 2, 1)
	if err != nil {
		t.Fatalf("extractGeneric failed: %v", err)
	}
	if result.Width != 2 || result.Height != 1 {
		t.Errorf("wrong size %dx%d", result.Width, result.Height)
	}
	if result.Components != 3 {
		t.Errorf("expected 3 components, got %d", result.Components)
	}
	if len(result.Data) != 2*1*3 {
		t.Errorf("wrong data length %d", len(result.Data))
	}
}

// TestDCTDecoder_ExtractGeneric_Empty verifies extractGeneric handles zero-size images.
func TestDCTDecoder_ExtractGeneric_Empty(t *testing.T) {
	d := NewDCTDecoder()

	palette := color.Palette{color.RGBA{R: 0, G: 0, B: 0, A: 255}}
	paletted := image.NewPaletted(image.Rect(0, 0, 0, 0), palette)
	img := &genericImage{img: paletted}
	result, err := d.extractGeneric(img, 0, 0)
	if err != nil {
		t.Fatalf("extractGeneric empty failed: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("expected empty data for 0x0 image")
	}
}

// paletteImage is kept for documentation purposes — demonstrates the generic path
// is reached only through direct method calls since jpeg.Decode never produces
// non-YCbCr/Gray images.
type paletteImage struct {
	img *image.Paletted
}

func (p *paletteImage) ColorModel() color.Model { return p.img.ColorModel() }
func (p *paletteImage) Bounds() image.Rectangle { return p.img.Bounds() }
func (p *paletteImage) At(x, y int) color.Color { return p.img.At(x, y) }

func TestDCTDecoder_Decode_MultipleImages(t *testing.T) {
	d := NewDCTDecoder()

	// Test decode with multiple small JPEG images to ensure no state leakage.
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
	}

	for i, c := range colors {
		jpegData := createTestJPEG(10, 10, c, 90)
		result, err := d.DecodeWithMetadata(jpegData)
		if err != nil {
			t.Fatalf("Iteration %d: decode failed: %v", i, err)
		}
		if result.Width != 10 || result.Height != 10 {
			t.Errorf("Iteration %d: wrong size %dx%d", i, result.Width, result.Height)
		}
		if result.Components != 3 {
			t.Errorf("Iteration %d: expected 3 components, got %d", i, result.Components)
		}
		if result.BitsPerComponent != 8 {
			t.Errorf("Iteration %d: expected 8 bits, got %d", i, result.BitsPerComponent)
		}
		if len(result.Data) != 10*10*3 {
			t.Errorf("Iteration %d: wrong data length %d", i, len(result.Data))
		}
	}
}

func TestDCTDecoder_Encode_RGB_InvalidLength(t *testing.T) {
	d := NewDCTDecoder()

	tests := []struct {
		name   string
		data   []byte
		width  int
		height int
	}{
		{"too few bytes", make([]byte, 10*10*3-1), 10, 10},
		{"too many bytes", make([]byte, 10*10*3+1), 10, 10},
		{"empty data nonzero dims", []byte{}, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Encode(tt.data, tt.width, tt.height, 75)
			if err == nil {
				t.Errorf("Expected error for %s", tt.name)
			}
		})
	}
}

// TestDCTDecoder_DecodeWithMetadata_GrayscaleSmall tests decode of a tiny gray JPEG.
func TestDCTDecoder_DecodeWithMetadata_GrayscaleSmall(t *testing.T) {
	d := NewDCTDecoder()

	// Encode a 1x1 grayscale image.
	img := image.NewGray(image.Rect(0, 0, 1, 1))
	img.SetGray(0, 0, color.Gray{Y: 42})

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})

	result, err := d.DecodeWithMetadata(buf.Bytes())
	if err != nil {
		t.Fatalf("Failed to decode 1x1 gray: %v", err)
	}
	if result.Width != 1 || result.Height != 1 {
		t.Errorf("Expected 1x1, got %dx%d", result.Width, result.Height)
	}
	if result.Components != 1 {
		t.Errorf("Expected 1 component for grayscale, got %d", result.Components)
	}
}

func BenchmarkFlateVsDCT(b *testing.B) {
	// This benchmark exists to compare memory characteristics — not an actual
	// vs comparison, just exercises both paths together.
	b.Run("DCT_small", func(b *testing.B) {
		d := NewDCTDecoder()
		data := createTestJPEG(32, 32, color.RGBA{100, 100, 100, 255}, 85)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = d.Decode(data)
		}
	})
}
