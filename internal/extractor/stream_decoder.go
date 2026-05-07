package extractor

import (
	"compress/zlib"
	"fmt"
	"io"

	"github.com/coregx/gxpdf/internal/parser"
)

// decodeStreamData decodes a PDF stream based on its Filter entry.
//
// This is the package-level shared implementation used by both
// TextExtractor and FontExtractor. It handles the most common filter:
// FlateDecode (zlib/deflate).
//
// Returns raw stream bytes when no filter is present.
// Returns raw bytes for unsupported filters so callers can decide.
func decodeStreamData(stream *parser.Stream) ([]byte, error) {
	filterObj := stream.Dictionary().Get("Filter")
	if filterObj == nil {
		// No filter — return raw content directly.
		return stream.Content(), nil
	}

	filterName := extractFirstFilterName(filterObj)

	switch filterName {
	case filterFlateDecode:
		return decodeFlateDecode(stream.Content())
	case "":
		return stream.Content(), nil
	default:
		// Return raw bytes for unsupported filters; callers decide.
		return stream.Content(), nil
	}
}

// extractFirstFilterName extracts a single filter name from a /Filter entry.
//
// /Filter can be either a Name (/FlateDecode) or an Array ([/FlateDecode]).
// For multi-filter chains we return only the first element — the callers
// in this codebase only ever encounter single-filter streams for fonts.
func extractFirstFilterName(filterObj parser.PdfObject) string {
	if name, ok := filterObj.(*parser.Name); ok {
		return name.Value()
	}
	if arr, ok := filterObj.(*parser.Array); ok && arr.Len() > 0 {
		if name, ok := arr.Get(0).(*parser.Name); ok {
			return name.Value()
		}
	}
	return ""
}

// decodeFlateDecode decompresses zlib/deflate-compressed data.
//
// This is the canonical FlateDecode implementation in the extractor package.
// Both TextExtractor.decodeFlateDecode and FontExtractor delegate here so
// the logic lives in one place.
func decodeFlateDecode(data []byte) ([]byte, error) {
	rc, err := zlib.NewReader(newBytesRC(data))
	if err != nil {
		return nil, fmt.Errorf("zlib reader: %w", err)
	}
	defer func() { _ = rc.Close() }()

	decoded, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("FlateDecode: %w", err)
	}
	return decoded, nil
}

// newBytesRC wraps a byte slice in an io.ReadCloser for zlib.NewReader.
func newBytesRC(data []byte) io.ReadCloser {
	return &sharedBytesRC{data: data}
}

type sharedBytesRC struct {
	data []byte
	pos  int
}

func (b *sharedBytesRC) Read(p []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *sharedBytesRC) Close() error { return nil }
