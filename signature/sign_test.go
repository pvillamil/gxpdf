package signature_test

import (
	"bytes"
	"crypto/x509"
	"strings"
	"testing"
	"time"

	"github.com/coregx/gxpdf/signature"
)

// minimalPDF is a tiny but valid PDF that can be signed.
// It contains a minimal header, single empty page, and proper trailer.
const minimalPDF = `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj

2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj

3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj

xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n

trailer
<< /Size 4 /Root 1 0 R >>
startxref
190
%%EOF
`

// makeMinimalPDF returns the minimal PDF as a byte slice.
func makeMinimalPDF() []byte {
	return []byte(minimalPDF)
}

func TestSignDocumentRSA(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}

	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	pdfData := makeMinimalPDF()
	signed, err := signature.SignDocument(pdfData, signer,
		signature.WithReason("Test RSA Signature"),
		signature.WithLocation("Test Location"),
	)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	// Signed PDF must still be a valid PDF.
	if !bytes.HasPrefix(signed, []byte("%PDF-")) {
		t.Error("signed PDF does not start with %PDF-")
	}

	// Must contain the incremental update markers.
	s := string(signed)
	if !strings.Contains(s, "/Type /Sig") {
		t.Error("signed PDF missing /Type /Sig")
	}
	if !strings.Contains(s, "/SubFilter /adbe.pkcs7.detached") {
		t.Error("signed PDF missing /SubFilter /adbe.pkcs7.detached")
	}
	if !strings.Contains(s, "/ByteRange") {
		t.Error("signed PDF missing /ByteRange")
	}
	if !strings.Contains(s, "/Contents") {
		t.Error("signed PDF missing /Contents")
	}
	if !strings.Contains(s, "/Reason (Test RSA Signature)") {
		t.Error("signed PDF missing /Reason")
	}
	if !strings.Contains(s, "/Location (Test Location)") {
		t.Error("signed PDF missing /Location")
	}

	// Signed PDF must be larger than original (incremental update appended).
	if len(signed) <= len(pdfData) {
		t.Errorf("signed PDF (%d bytes) not larger than original (%d bytes)", len(signed), len(pdfData))
	}

	// Original PDF bytes must be unchanged (incremental update only appends).
	if !bytes.Equal(signed[:len(pdfData)], pdfData) {
		t.Error("original PDF bytes were modified — incremental update violated")
	}
}

func TestSignDocumentECDSA(t *testing.T) {
	key, cert, err := signature.GenerateTestECCertificate()
	if err != nil {
		t.Fatalf("GenerateTestECCertificate: %v", err)
	}

	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	pdfData := makeMinimalPDF()
	signed, err := signature.SignDocument(pdfData, signer,
		signature.WithReason("Test ECDSA Signature"),
	)
	if err != nil {
		t.Fatalf("SignDocument ECDSA: %v", err)
	}

	if !bytes.HasPrefix(signed, []byte("%PDF-")) {
		t.Error("ECDSA signed PDF does not start with %PDF-")
	}
}

func TestSignDocumentContactInfo(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer,
		signature.WithContactInfo("signer@example.com"),
	)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	if !strings.Contains(string(signed), "/ContactInfo (signer@example.com)") {
		t.Error("signed PDF missing /ContactInfo")
	}
}

func TestSignDocumentNoOptions(t *testing.T) {
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
		t.Fatalf("SignDocument (no options): %v", err)
	}
	if !bytes.HasPrefix(signed, []byte("%PDF-")) {
		t.Error("signed PDF does not start with %PDF-")
	}
}

func TestSignRoundTripRSA(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer,
		signature.WithReason("Round-trip RSA"),
	)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	infos, err := signature.Verify(signed)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(infos) == 0 {
		t.Fatal("Verify returned no signatures")
	}

	info := infos[0]
	if !info.Valid {
		t.Fatalf("signature not valid: %v", info.Err)
	}
	if info.SignedBy != "Test RSA Signer" {
		t.Errorf("SignedBy = %q, want %q", info.SignedBy, "Test RSA Signer")
	}
	if info.Reason != "Round-trip RSA" {
		t.Errorf("Reason = %q, want %q", info.Reason, "Round-trip RSA")
	}
	if !info.IsRSA() {
		t.Error("expected RSA signature")
	}
	if info.RSAKeySize() != 2048 {
		t.Errorf("RSAKeySize = %d, want 2048", info.RSAKeySize())
	}
}

func TestSignRoundTripECDSA(t *testing.T) {
	key, cert, err := signature.GenerateTestECCertificate()
	if err != nil {
		t.Fatalf("GenerateTestECCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer,
		signature.WithReason("Round-trip ECDSA"),
		signature.WithLocation("Berlin"),
	)
	if err != nil {
		t.Fatalf("SignDocument ECDSA: %v", err)
	}

	infos, err := signature.Verify(signed)
	if err != nil {
		t.Fatalf("Verify ECDSA: %v", err)
	}
	if len(infos) == 0 {
		t.Fatal("Verify returned no signatures")
	}

	info := infos[0]
	if !info.Valid {
		t.Fatalf("ECDSA signature not valid: %v", info.Err)
	}
	if !info.IsECDSA() {
		t.Error("expected ECDSA signature")
	}
	if !info.IsECDSAP256() {
		t.Error("expected P-256 curve")
	}
	if info.ECDSACurve() != "P-256" {
		t.Errorf("ECDSACurve = %q, want %q", info.ECDSACurve(), "P-256")
	}
	if info.Location != "Berlin" {
		t.Errorf("Location = %q, want %q", info.Location, "Berlin")
	}
}

func TestSignRoundTripSigningTime(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	before := time.Now().UTC().Truncate(time.Second)
	signed, err := signature.SignDocument(makeMinimalPDF(), signer)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}
	after := time.Now().UTC()

	infos, err := signature.Verify(signed)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(infos) == 0 || !infos[0].Valid {
		t.Fatal("signature invalid after round-trip")
	}

	signedAt := infos[0].SignedAt
	if signedAt.IsZero() {
		t.Fatal("SignedAt is zero")
	}
	if signedAt.Before(before) || signedAt.After(after) {
		t.Errorf("SignedAt %v not in range [%v, %v]", signedAt, before, after)
	}
}

func TestVerifyEmptyInput(t *testing.T) {
	infos, err := signature.Verify([]byte("%PDF-1.4\n%%EOF\n"))
	if err != nil {
		t.Fatalf("Verify empty: unexpected error: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("expected 0 signatures, got %d", len(infos))
	}
}

func TestSignDocumentInvalidPDF(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	_, err = signature.SignDocument([]byte("not a pdf"), signer)
	if err == nil {
		t.Fatal("expected error for non-PDF input")
	}
}

func TestNewLocalSignerNilKey(t *testing.T) {
	_, err := signature.NewLocalSigner(nil, []*x509.Certificate{})
	if err == nil {
		t.Fatal("expected error for nil key")
	}
}

func TestNewLocalSignerNoCerts(t *testing.T) {
	key, _, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	_, err = signature.NewLocalSigner(key, nil)
	if err == nil {
		t.Fatal("expected error for nil certs")
	}
}

func TestLocalSignerAlgorithmOverride(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	// Default is SHA256WithRSA for RSA key.
	if signer.Algorithm() != signature.SHA256WithRSA {
		t.Errorf("default algorithm = %d, want SHA256WithRSA", signer.Algorithm())
	}

	signer.SetAlgorithm(signature.SHA384WithRSA)
	if signer.Algorithm() != signature.SHA384WithRSA {
		t.Errorf("overridden algorithm = %d, want SHA384WithRSA", signer.Algorithm())
	}
}

func TestAlgorithmProperties(t *testing.T) {
	tests := []struct {
		algo       signature.Algorithm
		wantDigest string
		wantSig    string
	}{
		{signature.SHA256WithRSA, "2.16.840.1.101.3.4.2.1", "1.2.840.113549.1.1.11"},
		{signature.SHA384WithRSA, "2.16.840.1.101.3.4.2.2", "1.2.840.113549.1.1.12"},
		{signature.SHA512WithRSA, "2.16.840.1.101.3.4.2.3", "1.2.840.113549.1.1.13"},
		{signature.SHA256WithECDSA, "2.16.840.1.101.3.4.2.1", "1.2.840.10045.4.3.2"},
		{signature.SHA384WithECDSA, "2.16.840.1.101.3.4.2.2", "1.2.840.10045.4.3.3"},
		{signature.SHA512WithECDSA, "2.16.840.1.101.3.4.2.3", "1.2.840.10045.4.3.4"},
	}
	for _, tt := range tests {
		digestOID := tt.algo.DigestOID().String()
		if digestOID != tt.wantDigest {
			t.Errorf("algo %d DigestOID = %s, want %s", tt.algo, digestOID, tt.wantDigest)
		}
		sigOID := tt.algo.SignatureOID().String()
		if sigOID != tt.wantSig {
			t.Errorf("algo %d SignatureOID = %s, want %s", tt.algo, sigOID, tt.wantSig)
		}
	}
}
