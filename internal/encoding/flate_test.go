package encoding

import (
	"bytes"
	"compress/zlib"
	"testing"
)

func TestNewFlateDecoder(t *testing.T) {
	d := NewFlateDecoder()
	if d == nil {
		t.Fatal("NewFlateDecoder returned nil")
	}
}

func TestFlateDecoder_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"single byte", []byte{0x42}},
		{"hello world", []byte("hello world")},
		{"binary data", []byte{0x00, 0xFF, 0x01, 0xFE, 0x80, 0x7F}},
		{"all zeros", make([]byte, 256)},
		{"all 0xFF", bytes.Repeat([]byte{0xFF}, 256)},
		{"repeated pattern", bytes.Repeat([]byte("abcdefgh"), 100)},
		{"random-ish", func() []byte {
			b := make([]byte, 1024)
			for i := range b {
				b[i] = byte(i * 7 % 256)
			}
			return b
		}()},
	}

	d := NewFlateDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := d.Encode(tt.input)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decoded, err := d.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("Round-trip mismatch: got %v, want %v", decoded, tt.input)
			}
		})
	}
}

func TestFlateDecoder_Decode_CorruptedData(t *testing.T) {
	d := NewFlateDecoder()

	tests := []struct {
		name  string
		input []byte
	}{
		{"not zlib at all", []byte("not zlib data")},
		{"truncated zlib header", []byte{0x78}},
		{"invalid zlib magic", []byte{0x00, 0x00, 0x00, 0x00, 0x00}},
		{"random bytes", []byte{0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Decode(tt.input)
			if err == nil {
				t.Errorf("Expected error for corrupted data %q, but got nil", tt.name)
			}
		})
	}
}

func TestFlateDecoder_Decode_ValidZlib(t *testing.T) {
	// Manually construct valid zlib data.
	input := []byte("test data for zlib decompression")

	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(input)
	w.Close()

	d := NewFlateDecoder()
	result, err := d.Decode(buf.Bytes())
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(result, input) {
		t.Errorf("Decode mismatch: got %q, want %q", result, input)
	}
}

func TestFlateDecoder_Encode_ProducesValidZlib(t *testing.T) {
	d := NewFlateDecoder()
	input := []byte("compress me!")

	encoded, err := d.Encode(input)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Verify it starts with zlib magic bytes (0x78 0x9C or 0x78 0x01 or 0x78 0xDA).
	if len(encoded) < 2 {
		t.Fatal("Encoded data too short to be valid zlib")
	}
	if encoded[0] != 0x78 {
		t.Errorf("Expected zlib magic first byte 0x78, got 0x%02X", encoded[0])
	}
}

func TestFlateDecoder_Decode_LargeData(t *testing.T) {
	// 1MB of compressible data.
	input := bytes.Repeat([]byte("AAABBBCCC"), 100_000/9+1)
	input = input[:100_000]

	d := NewFlateDecoder()
	encoded, err := d.Encode(input)
	if err != nil {
		t.Fatalf("Encode large data failed: %v", err)
	}

	// Compressed should be significantly smaller (high repetition).
	if len(encoded) >= len(input) {
		t.Logf("Note: encoded size %d >= input size %d (unexpected for repetitive data)", len(encoded), len(input))
	}

	decoded, err := d.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode large data failed: %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("Large data round-trip mismatch (lengths: got %d, want %d)", len(decoded), len(input))
	}
}

func BenchmarkFlateDecoder_Encode(b *testing.B) {
	d := NewFlateDecoder()
	input := bytes.Repeat([]byte("benchmark data for flate encode "), 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.Encode(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFlateDecoder_Decode(b *testing.B) {
	d := NewFlateDecoder()
	input := bytes.Repeat([]byte("benchmark data for flate decode "), 1000)
	encoded, _ := d.Encode(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}
