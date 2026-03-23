package signature_test

import (
	"crypto/x509"
	"testing"

	"github.com/coregx/gxpdf/signature"
)

func TestTamperDetection(t *testing.T) {
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

	// Verify the original is valid.
	infos, err := signature.Verify(signed)
	if err != nil {
		t.Fatalf("Verify original: %v", err)
	}
	if len(infos) == 0 || !infos[0].Valid {
		t.Fatal("original signature should be valid")
	}

	// Tamper with a byte in the first signed byte range.
	offset := infos[0].FindSignedByteOffset()
	if offset < 0 {
		t.Skip("could not find a safe tamper offset")
	}
	tampered := signature.TamperByte(signed, offset)

	// Tampered PDF should fail verification.
	tamperedInfos, err := signature.Verify(tampered)
	if err != nil {
		t.Fatalf("Verify tampered: %v", err)
	}
	if len(tamperedInfos) == 0 {
		t.Fatal("Verify tampered returned no signatures")
	}
	if tamperedInfos[0].Valid {
		t.Error("tampered PDF signature should NOT be valid")
	}
}

func TestTamperDetectionECDSA(t *testing.T) {
	key, cert, err := signature.GenerateTestECCertificate()
	if err != nil {
		t.Fatalf("GenerateTestECCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer)
	if err != nil {
		t.Fatalf("SignDocument ECDSA: %v", err)
	}

	infos, err := signature.Verify(signed)
	if err != nil {
		t.Fatalf("Verify ECDSA: %v", err)
	}
	if len(infos) == 0 || !infos[0].Valid {
		t.Fatal("ECDSA original signature should be valid")
	}

	offset := infos[0].FindSignedByteOffset()
	if offset < 0 {
		t.Skip("could not find a safe tamper offset")
	}
	tampered := signature.TamperByte(signed, offset)

	tamperedInfos, err := signature.Verify(tampered)
	if err != nil {
		t.Fatalf("Verify ECDSA tampered: %v", err)
	}
	if len(tamperedInfos) == 0 {
		t.Fatal("Verify ECDSA tampered returned no signatures")
	}
	if tamperedInfos[0].Valid {
		t.Error("ECDSA tampered PDF signature should NOT be valid")
	}
}

func TestSignatureInfoFields(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer,
		signature.WithReason("Field Test"),
		signature.WithLocation("Moscow"),
		signature.WithContactInfo("test@example.com"),
	)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	infos, err := signature.Verify(signed)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(infos) == 0 {
		t.Fatal("no signatures found")
	}
	info := infos[0]

	if !info.Valid {
		t.Fatalf("signature not valid: %v", info.Err)
	}
	if info.SubFilter != "adbe.pkcs7.detached" {
		t.Errorf("SubFilter = %q, want %q", info.SubFilter, "adbe.pkcs7.detached")
	}
	if info.Reason != "Field Test" {
		t.Errorf("Reason = %q, want %q", info.Reason, "Field Test")
	}
	if info.Location != "Moscow" {
		t.Errorf("Location = %q, want %q", info.Location, "Moscow")
	}
	if info.SignedBy != "Test RSA Signer" {
		t.Errorf("SignedBy = %q, want %q", info.SignedBy, "Test RSA Signer")
	}
	if info.Certificate == nil {
		t.Fatal("Certificate is nil")
	}
	if info.HasTimestamp {
		t.Error("HasTimestamp should be false without TSA")
	}
}

func TestVerifyNoSignature(t *testing.T) {
	infos, err := signature.Verify(makeMinimalPDF())
	if err != nil {
		t.Fatalf("Verify unsigned PDF: unexpected error: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("expected 0 signatures in unsigned PDF, got %d", len(infos))
	}
}

func TestParseSignatureInfo(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	signed, err := signature.SignDocument(makeMinimalPDF(), signer,
		signature.WithReason("Parse Test"),
	)
	if err != nil {
		t.Fatalf("SignDocument: %v", err)
	}

	info, err := signature.ParseSignatureInfo(signed)
	if err != nil {
		t.Fatalf("ParseSignatureInfo: %v", err)
	}

	if info.Reason != "Parse Test" {
		t.Errorf("Reason = %q, want %q", info.Reason, "Parse Test")
	}
	if info.Certificate == nil {
		t.Error("Certificate is nil")
	}
}

func TestTamperByteEdgeCases(t *testing.T) {
	data := []byte("hello world")

	// Negative offset — returns unchanged.
	result := signature.TamperByte(data, -1)
	if string(result) != "hello world" {
		t.Error("negative offset should return unchanged data")
	}

	// Out-of-bounds offset — returns unchanged.
	result = signature.TamperByte(data, len(data))
	if string(result) != "hello world" {
		t.Error("out-of-bounds offset should return unchanged data")
	}

	// Valid offset — byte is XORed.
	result = signature.TamperByte(data, 0)
	if result[0] == data[0] {
		t.Error("TamperByte should change the byte at offset 0")
	}
	// Applying twice returns original.
	double := signature.TamperByte(result, 0)
	if double[0] != data[0] {
		t.Error("double TamperByte should restore original byte")
	}
}

func TestSignatureInfoIsHelpers(t *testing.T) {
	// RSA signer.
	rsaKey, rsaCert, _ := signature.GenerateTestCertificate()
	rsaSigner, _ := signature.NewLocalSigner(rsaKey, []*x509.Certificate{rsaCert})
	rsaSigned, _ := signature.SignDocument(makeMinimalPDF(), rsaSigner)
	rsaInfos, _ := signature.Verify(rsaSigned)
	if len(rsaInfos) == 0 || rsaInfos[0].Certificate == nil {
		t.Fatal("RSA sign/verify produced no info")
	}
	rsaInfo := rsaInfos[0]
	if !rsaInfo.IsRSA() {
		t.Error("IsRSA should be true for RSA signer")
	}
	if rsaInfo.IsECDSA() {
		t.Error("IsECDSA should be false for RSA signer")
	}
	if rsaInfo.RSAKeySize() != 2048 {
		t.Errorf("RSAKeySize = %d, want 2048", rsaInfo.RSAKeySize())
	}
	if rsaInfo.ECDSACurve() != "" {
		t.Errorf("ECDSACurve should be empty for RSA, got %q", rsaInfo.ECDSACurve())
	}

	// ECDSA signer.
	ecKey, ecCert, _ := signature.GenerateTestECCertificate()
	ecSigner, _ := signature.NewLocalSigner(ecKey, []*x509.Certificate{ecCert})
	ecSigned, _ := signature.SignDocument(makeMinimalPDF(), ecSigner)
	ecInfos, _ := signature.Verify(ecSigned)
	if len(ecInfos) == 0 || ecInfos[0].Certificate == nil {
		t.Fatal("ECDSA sign/verify produced no info")
	}
	ecInfo := ecInfos[0]
	if ecInfo.IsRSA() {
		t.Error("IsRSA should be false for ECDSA signer")
	}
	if !ecInfo.IsECDSA() {
		t.Error("IsECDSA should be true for ECDSA signer")
	}
	if ecInfo.RSAKeySize() != 0 {
		t.Errorf("RSAKeySize should be 0 for ECDSA, got %d", ecInfo.RSAKeySize())
	}
	if !ecInfo.IsECDSAP256() {
		t.Error("IsECDSAP256 should be true for P-256 key")
	}
}
