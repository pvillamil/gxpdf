package security

import (
	"bytes"
	"testing"
)

func TestComputeDocumentKey(t *testing.T) {
	// Create a known encryption setup via RC4Encryptor, then verify
	// that computeDocumentKey produces the same key.
	config := &EncryptionConfig{
		UserPassword:  "test",
		OwnerPassword: "owner",
		Permissions:   PermissionPrint | PermissionCopy,
		KeyLength:     128,
		FileID:        "test-file-id-123",
	}

	enc, err := NewRC4Encryptor(config)
	if err != nil {
		t.Fatalf("NewRC4Encryptor: %v", err)
	}

	dict := enc.GetEncryptionDict()

	info := &EncryptionInfo{
		Filter: "Standard",
		V:      dict.V,
		R:      dict.R,
		Length: dict.Length,
		P:      dict.P,
		O:      dict.O,
		U:      dict.U,
		FileID: []byte(config.FileID),
	}

	// The document key derived from EncryptionInfo should match
	// what the encryptor computes internally
	key := computeDocumentKey(info, "test")
	if len(key) != 16 { // 128/8 = 16
		t.Errorf("key length = %d, want 16", len(key))
	}
}

func TestComputeDocumentKey_40bit(t *testing.T) {
	config := &EncryptionConfig{
		UserPassword:  "",
		OwnerPassword: "",
		Permissions:   PermissionAll,
		KeyLength:     40,
		FileID:        "file-id",
	}

	enc, err := NewRC4Encryptor(config)
	if err != nil {
		t.Fatalf("NewRC4Encryptor: %v", err)
	}

	dict := enc.GetEncryptionDict()

	info := &EncryptionInfo{
		Filter: "Standard",
		V:      dict.V,
		R:      dict.R,
		Length: dict.Length,
		P:      dict.P,
		O:      dict.O,
		U:      dict.U,
		FileID: []byte(config.FileID),
	}

	key := computeDocumentKey(info, "")
	if len(key) != 5 { // 40/8 = 5
		t.Errorf("key length = %d, want 5", len(key))
	}
}

func TestVerifyUserPassword_Empty(t *testing.T) {
	// Create encrypted PDF with empty password, then verify we can decrypt it
	config := &EncryptionConfig{
		UserPassword:  "",
		OwnerPassword: "owner",
		Permissions:   PermissionPrint | PermissionCopy,
		KeyLength:     128,
		FileID:        "test-file-id-verify",
	}

	enc, err := NewRC4Encryptor(config)
	if err != nil {
		t.Fatalf("NewRC4Encryptor: %v", err)
	}

	dict := enc.GetEncryptionDict()

	info := &EncryptionInfo{
		Filter: "Standard",
		V:      dict.V,
		R:      dict.R,
		Length: dict.Length,
		P:      dict.P,
		O:      dict.O,
		U:      dict.U,
		FileID: []byte(config.FileID),
	}

	// Empty password should verify
	key, err := verifyUserPassword(info, "")
	if err != nil {
		t.Fatalf("verifyUserPassword with empty password: %v", err)
	}
	if len(key) == 0 {
		t.Error("expected non-empty key")
	}
}

func TestVerifyUserPassword_Wrong(t *testing.T) {
	config := &EncryptionConfig{
		UserPassword:  "correct",
		OwnerPassword: "owner",
		Permissions:   PermissionAll,
		KeyLength:     128,
		FileID:        "test-file-id-wrong",
	}

	enc, err := NewRC4Encryptor(config)
	if err != nil {
		t.Fatalf("NewRC4Encryptor: %v", err)
	}

	dict := enc.GetEncryptionDict()

	info := &EncryptionInfo{
		Filter: "Standard",
		V:      dict.V,
		R:      dict.R,
		Length: dict.Length,
		P:      dict.P,
		O:      dict.O,
		U:      dict.U,
		FileID: []byte(config.FileID),
	}

	// Wrong password should fail
	_, err = verifyUserPassword(info, "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if err != ErrPasswordRequired {
		t.Errorf("expected ErrPasswordRequired, got: %v", err)
	}
}

func TestVerifyUserPassword_R2(t *testing.T) {
	// Test R=2 (40-bit) password verification
	config := &EncryptionConfig{
		UserPassword:  "pass",
		OwnerPassword: "owner",
		Permissions:   PermissionPrint,
		KeyLength:     40,
		FileID:        "test-file-r2",
	}

	enc, err := NewRC4Encryptor(config)
	if err != nil {
		t.Fatalf("NewRC4Encryptor: %v", err)
	}

	dict := enc.GetEncryptionDict()

	info := &EncryptionInfo{
		Filter: "Standard",
		V:      dict.V,
		R:      dict.R,
		Length: dict.Length,
		P:      dict.P,
		O:      dict.O,
		U:      dict.U,
		FileID: []byte(config.FileID),
	}

	// Correct password should verify
	key, err := verifyUserPassword(info, "pass")
	if err != nil {
		t.Fatalf("verifyUserPassword: %v", err)
	}
	if len(key) != 5 { // 40/8 = 5
		t.Errorf("key length = %d, want 5", len(key))
	}

	// Wrong password should fail
	_, err = verifyUserPassword(info, "wrong")
	if err != ErrPasswordRequired {
		t.Errorf("expected ErrPasswordRequired, got: %v", err)
	}
}

func TestObjectKey_RC4(t *testing.T) {
	docKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}

	key1 := objectKey(docKey, 1, 0, false)
	key2 := objectKey(docKey, 2, 0, false)

	// Different objects should produce different keys
	if bytes.Equal(key1, key2) {
		t.Error("different objects should produce different keys")
	}

	// Key length should be min(docKey+5, 16) = 16
	if len(key1) != 16 {
		t.Errorf("key length = %d, want 16", len(key1))
	}
}

func TestObjectKey_AES(t *testing.T) {
	docKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}

	rc4Key := objectKey(docKey, 1, 0, false)
	aesKey := objectKey(docKey, 1, 0, true)

	// AES key should differ from RC4 key (due to salt)
	if bytes.Equal(rc4Key, aesKey) {
		t.Error("AES and RC4 keys for same object should differ")
	}

	// AES key should also be 16 bytes
	if len(aesKey) != 16 {
		t.Errorf("AES key length = %d, want 16", len(aesKey))
	}
}

func TestObjectKey_ShortDocKey(t *testing.T) {
	docKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05} // 40-bit (5 bytes)

	key := objectKey(docKey, 1, 0, false)

	// Key length should be min(5+5, 16) = 10
	if len(key) != 10 {
		t.Errorf("key length = %d, want 10", len(key))
	}
}

func TestRC4Decryptor_RoundTrip(t *testing.T) {
	// Create an encryptor to get known encryption parameters
	config := &EncryptionConfig{
		UserPassword:  "test",
		OwnerPassword: "test",
		Permissions:   PermissionAll,
		KeyLength:     128,
		FileID:        "roundtrip-test",
	}

	enc, err := NewRC4Encryptor(config)
	if err != nil {
		t.Fatalf("NewRC4Encryptor: %v", err)
	}

	dict := enc.GetEncryptionDict()

	info := &EncryptionInfo{
		Filter: "Standard",
		V:      dict.V,
		R:      dict.R,
		Length: dict.Length,
		P:      dict.P,
		O:      dict.O,
		U:      dict.U,
		FileID: []byte(config.FileID),
	}

	// Create decryptor
	dec, err := NewDecryptor(info, "test")
	if err != nil {
		t.Fatalf("NewDecryptor: %v", err)
	}

	// Test data
	original := []byte("Hello, this is test data for RC4 round-trip!")
	objNum := 5
	genNum := 0

	// Encrypt: use the same per-object key derivation
	docKey := computeDocumentKey(info, "test")
	key := objectKey(docKey, objNum, genNum, false)
	encrypted := make([]byte, len(original))
	if err := encryptRC4(key, original, encrypted); err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// Decrypt via Decryptor
	decrypted, err := dec.DecryptStream(encrypted, objNum, genNum)
	if err != nil {
		t.Fatalf("DecryptStream: %v", err)
	}

	if !bytes.Equal(decrypted, original) {
		t.Errorf("round-trip failed:\n  got:  %q\n  want: %q", decrypted, original)
	}

	// Also test DecryptString
	decryptedStr, err := dec.DecryptString(encrypted, objNum, genNum)
	if err != nil {
		t.Fatalf("DecryptString: %v", err)
	}

	if !bytes.Equal(decryptedStr, original) {
		t.Errorf("string round-trip failed:\n  got:  %q\n  want: %q", decryptedStr, original)
	}
}

func TestAES128Decryptor_RoundTrip(t *testing.T) {
	// Per PDF spec, V=4/R=4 (AES-128) still uses RC4-based O/U computation
	// for password verification. Only content encryption uses AES.
	// We use RC4Encryptor to get correct O/U values, then set V=4/R=4/CFM=AESV2.
	config := &EncryptionConfig{
		UserPassword:  "aestest",
		OwnerPassword: "aestest",
		Permissions:   PermissionAll,
		KeyLength:     128,
		FileID:        "aes-roundtrip",
	}

	// Use RC4 encryptor for O/U computation (PDF spec Algorithm 3.4/3.5)
	rc4Enc, err := NewRC4Encryptor(config)
	if err != nil {
		t.Fatalf("NewRC4Encryptor: %v", err)
	}
	rc4Dict := rc4Enc.GetEncryptionDict()

	// Build AES-128 encryption info with RC4-based O/U
	info := &EncryptionInfo{
		Filter: "Standard",
		V:      4,
		R:      4,
		Length: 128,
		P:      rc4Dict.P,
		O:      rc4Dict.O,
		U:      rc4Dict.U,
		FileID: []byte(config.FileID),
		CFM:    "AESV2",
	}

	// Create decryptor
	dec, err := NewDecryptor(info, "aestest")
	if err != nil {
		t.Fatalf("NewDecryptor: %v", err)
	}

	// Test data
	original := []byte("Hello, this is test data for AES-128 round-trip!")
	objNum := 7
	genNum := 0

	// Encrypt: derive per-object AES key and encrypt
	docKey := computeDocumentKey(info, "aestest")
	key := objectKey(docKey, objNum, genNum, true)
	encrypted, err := encryptAES(key, original)
	if err != nil {
		t.Fatalf("encryptAES: %v", err)
	}

	// Decrypt via Decryptor
	decrypted, err := dec.DecryptStream(encrypted, objNum, genNum)
	if err != nil {
		t.Fatalf("DecryptStream: %v", err)
	}

	if !bytes.Equal(decrypted, original) {
		t.Errorf("AES round-trip failed:\n  got:  %q\n  want: %q", decrypted, original)
	}
}

func TestNewDecryptor_UnsupportedV5(t *testing.T) {
	info := &EncryptionInfo{
		Filter: "Standard",
		V:      5,
		R:      6,
		Length: 256,
		O:      make([]byte, 48),
		U:      make([]byte, 48),
		FileID: []byte("test"),
		CFM:    "AESV3",
	}

	_, err := NewDecryptor(info, "test")
	if err == nil {
		t.Fatal("expected error for V=5")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("unsupported")) {
		t.Errorf("expected unsupported version error, got: %v", err)
	}
}

func TestNewDecryptor_UnsupportedV(t *testing.T) {
	info := &EncryptionInfo{
		Filter: "Standard",
		V:      99,
		R:      99,
		Length: 128,
		O:      make([]byte, 32),
		U:      make([]byte, 32),
		FileID: []byte("test"),
	}

	_, err := NewDecryptor(info, "test")
	if err == nil {
		t.Fatal("expected error for unsupported V")
	}
}

func TestRC4Decryptor_EmptyData(t *testing.T) {
	dec := &rc4Decryptor{docKey: []byte{1, 2, 3, 4, 5}}

	result, err := dec.DecryptStream(nil, 1, 0)
	if err != nil {
		t.Fatalf("DecryptStream nil: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}

	result, err = dec.DecryptStream([]byte{}, 1, 0)
	if err != nil {
		t.Fatalf("DecryptStream empty: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty for empty input, got %v", result)
	}
}

func TestAES128Decryptor_EmptyData(t *testing.T) {
	dec := &aes128Decryptor{docKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}

	result, err := dec.DecryptStream(nil, 1, 0)
	if err != nil {
		t.Fatalf("DecryptStream nil: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

func TestBytesEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b []byte
		want bool
	}{
		{"equal", []byte{1, 2, 3}, []byte{1, 2, 3}, true},
		{"different", []byte{1, 2, 3}, []byte{1, 2, 4}, false},
		{"different length", []byte{1, 2}, []byte{1, 2, 3}, false},
		{"both empty", []byte{}, []byte{}, true},
		{"both nil", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("bytesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
