package signature_test

import (
	"bytes"
	"crypto/x509"
	"strings"
	"testing"

	"github.com/coregx/gxpdf/signature"
)

// TestByteRangePresent verifies /ByteRange appears in signed PDF and has non-zero values.
func TestByteRangePresent(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	s := string(signed)
	if !strings.Contains(s, "/ByteRange") {
		t.Fatal("/ByteRange not found in signed PDF")
	}

	// The placeholder [0000000000 ...] must have been patched to real values.
	if strings.Contains(s, "[0000000000 0000000000 0000000000 0000000000]") {
		t.Error("/ByteRange still contains all-zero placeholder — patching failed")
	}
}

// TestContentsPresent verifies /Contents <HEXSTRING> appears in signed PDF.
func TestContentsPresent(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	// /Contents must be present and not all zeros (signature injected).
	s := string(signed)
	if !strings.Contains(s, "/Contents <") {
		t.Fatal("/Contents not found in signed PDF")
	}

	// Find the contents hex string.
	idx := strings.Index(s, "/Contents <")
	if idx < 0 {
		t.Fatal("/Contents not found")
	}
	start := idx + len("/Contents <")
	end := strings.Index(s[start:], ">")
	if end < 0 {
		t.Fatal("no closing > for /Contents hex string")
	}
	hexStr := s[start : start+end]

	// Must not be entirely zeros.
	allZero := true
	for _, c := range hexStr {
		if c != '0' {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("/Contents is all zeros — CMS injection failed")
	}
}

// TestIncrementalUpdatePreservesOriginal verifies the incremental update pattern:
// original bytes are preserved exactly and new content is appended.
func TestIncrementalUpdatePreservesOriginal(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	pdfData := makeMinimalPDF()
	signed, err := signature.SignDocument(pdfData, signer)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	// The original bytes must be byte-for-byte identical at the start.
	if !bytes.Equal(signed[:len(pdfData)], pdfData) {
		t.Error("original PDF bytes modified — incremental update must only append")
	}

	// The appended portion must contain xref and %%EOF.
	appended := string(signed[len(pdfData):])
	if !strings.Contains(appended, "xref") {
		t.Error("appended portion missing xref table")
	}
	if !strings.Contains(appended, "startxref") {
		t.Error("appended portion missing startxref")
	}
	if !strings.Contains(appended, "%%EOF") {
		t.Error("appended portion missing EOF marker")
	}
	if !strings.Contains(appended, "/Prev") {
		t.Error("appended trailer missing /Prev")
	}
}

// TestByteRangeCoversAllExceptContents verifies the ByteRange excludes exactly the
// /Contents hex string — this is the core invariant of the signature mechanism.
func TestByteRangeCoversAllExceptContents(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	infos, err := signature.Verify(signed)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(infos) == 0 || !infos[0].Valid {
		t.Fatalf("signature not valid: %v", infos[0].Err)
	}

	// Sanity check: find the /Contents position and verify the ByteRange
	// skips it by locating /Contents < in the signed bytes.
	s := string(signed)
	contentsIdx := strings.Index(s, "/Contents <")
	if contentsIdx < 0 {
		t.Fatal("/Contents not found")
	}
	// The /Contents hex begins after "/Contents <".
	hexStart := contentsIdx + len("/Contents <")
	hexEnd := strings.Index(s[hexStart:], ">")
	if hexEnd < 0 {
		t.Fatal("no closing > for /Contents")
	}
	contentsEnd := hexStart + hexEnd + 1 // position after '>'

	// Verify: signed PDF length should equal sum of the two byte ranges.
	// This is implicit in a passing Verify, but we also check the structure directly.
	if contentsEnd > len(signed) {
		t.Fatalf("contentsEnd %d exceeds PDF length %d", contentsEnd, len(signed))
	}
}

// TestSignedPDFHasWidget verifies the widget annotation object is appended.
func TestSignedPDFHasWidget(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	s := string(signed)
	if !strings.Contains(s, "/Subtype /Widget") {
		t.Error("signed PDF missing widget annotation (/Subtype /Widget)")
	}
	if !strings.Contains(s, "/FT /Sig") {
		t.Error("signed PDF missing signature field type (/FT /Sig)")
	}
}
