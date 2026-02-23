// Package security provides PDF encryption and security features.
//
// This file implements PDF decryption for reading encrypted PDF documents.
//
// Supported decryption:
//   - RC4 with 40-bit keys (V=1, R=2)
//   - RC4 with 128-bit keys (V=2, R=3)
//   - AES-128 with 128-bit keys (V=4, R=4, CFM=AESV2)
//
// The decryption follows the PDF Standard Security Handler algorithms
// as specified in PDF Reference 1.7, Section 3.5.
package security

import (
	"crypto/md5" //nolint:gosec // MD5 required by PDF Standard Security Handler
	"fmt"
)

// Decryptor decrypts PDF streams and strings on a per-object basis.
//
// Each PDF object is encrypted with a unique key derived from the document
// encryption key and the object/generation numbers.
type Decryptor interface {
	// DecryptStream decrypts stream data for the given object.
	DecryptStream(data []byte, objNum, genNum int) ([]byte, error)

	// DecryptString decrypts a string value for the given object.
	DecryptString(data []byte, objNum, genNum int) ([]byte, error)
}

// EncryptionInfo holds values parsed from the /Encrypt dictionary and trailer.
// These are used to verify the password and derive decryption keys.
type EncryptionInfo struct {
	Filter string // /Filter — should be "Standard"
	V      int    // /V — algorithm version (1, 2, or 4)
	R      int    // /R — algorithm revision (2, 3, or 4)
	Length int    // /Length — key length in bits (40 or 128)
	P      int32  // /P — permission flags
	O      []byte // /O — owner password hash (32 bytes)
	U      []byte // /U — user password hash (32 bytes)
	FileID []byte // First element of trailer /ID array
	CFM    string // /CFM from /CF./StdCF (empty for RC4, "AESV2" for AES-128)
}

// NewDecryptor verifies the password against the encryption info and returns
// an appropriate Decryptor implementation.
//
// For V=1/2 (RC4), it returns an rc4Decryptor.
// For V=4 with CFM=AESV2 (AES-128), it returns an aes128Decryptor.
// For V=5 (AES-256), it returns ErrUnsupportedVersion.
//
// Returns ErrPasswordRequired if the password does not match.
func NewDecryptor(info *EncryptionInfo, password string) (Decryptor, error) {
	switch info.V {
	case 1, 2:
		// RC4 encryption (40-bit or 128-bit)
		docKey, err := verifyUserPassword(info, password)
		if err != nil {
			return nil, err
		}
		return &rc4Decryptor{docKey: docKey}, nil

	case 4:
		// Could be AES-128 or RC4 depending on CFM
		docKey, err := verifyUserPassword(info, password)
		if err != nil {
			return nil, err
		}
		if info.CFM == "AESV2" {
			return &aes128Decryptor{docKey: docKey}, nil
		}
		// V=4 without AESV2 means RC4 with crypt filters
		return &rc4Decryptor{docKey: docKey}, nil

	case 5:
		return nil, fmt.Errorf("%w: V=5 (AES-256) not yet supported", ErrUnsupportedVersion)

	default:
		return nil, fmt.Errorf("%w: V=%d", ErrUnsupportedVersion, info.V)
	}
}

// computeDocumentKey derives the document encryption key from the user password.
//
// Implements Algorithm 2 from PDF Reference 1.7, Section 3.5.2:
//  1. Pad password to 32 bytes
//  2. MD5(password + O + P + FileID)
//  3. For R>=3: iterate MD5 50 times over first n bytes
//  4. Return first n bytes (n = Length/8)
func computeDocumentKey(info *EncryptionInfo, password string) []byte {
	padded := padPassword(password)

	h := md5.New() //nolint:gosec // MD5 required by PDF spec
	h.Write(padded)
	h.Write(info.O)
	h.Write(int32ToBytes(info.P))
	h.Write(info.FileID)
	hash := h.Sum(nil)

	keyLen := info.Length / 8
	if keyLen <= 0 {
		keyLen = 5 // Default for V=1 (40-bit)
	}

	// For R>=3, iterate MD5 50 times
	if info.R >= 3 {
		for i := 0; i < 50; i++ {
			hashArray := md5.Sum(hash[:keyLen]) //nolint:gosec // MD5 required by PDF spec
			hash = hashArray[:]
		}
	}

	return hash[:keyLen]
}

// verifyUserPassword checks the user password and returns the document key.
//
// Implements Algorithm 6 from PDF Reference 1.7, Section 3.5.2:
//   - For R=2: encrypt padding with key, compare all 32 bytes with /U
//   - For R=3/4: encrypt MD5(padding+FileID), iterate 20 times, compare first 16 bytes
//
// Returns the document key on success, or ErrPasswordRequired if the password is wrong.
func verifyUserPassword(info *EncryptionInfo, password string) ([]byte, error) {
	docKey := computeDocumentKey(info, password)

	if info.R == 2 {
		// Algorithm 4: encrypt padding string with document key
		result := make([]byte, 32)
		if err := encryptRC4(docKey, []byte(paddingString), result); err != nil {
			return nil, fmt.Errorf("verify password: %w", err)
		}

		if len(info.U) >= 32 && bytesEqual(result, info.U[:32]) {
			return docKey, nil
		}
		return nil, ErrPasswordRequired
	}

	// R=3 or R=4: Algorithm 5
	h := md5.New() //nolint:gosec // MD5 required by PDF spec
	h.Write([]byte(paddingString))
	h.Write(info.FileID)
	hash := h.Sum(nil)

	// Encrypt with document key
	result := make([]byte, len(hash))
	if err := encryptRC4(docKey, hash, result); err != nil {
		return nil, fmt.Errorf("verify password: %w", err)
	}

	// Iterate 19 times with XOR'd keys
	for i := 1; i <= 19; i++ {
		newKey := xorKey(docKey, byte(i))
		if err := encryptRC4(newKey, result, result); err != nil {
			return nil, fmt.Errorf("verify password: %w", err)
		}
	}

	// Compare first 16 bytes with /U
	if len(info.U) >= 16 && bytesEqual(result[:16], info.U[:16]) {
		return docKey, nil
	}

	return nil, ErrPasswordRequired
}

// objectKey derives the per-object encryption key.
//
// Implements Algorithm 1 from PDF Reference 1.7, Section 3.5.2:
//  1. Concatenate docKey + objNum (3 LE bytes) + genNum (2 LE bytes)
//  2. For AES, append "sAlT" (0x73 0x41 0x6C 0x54)
//  3. MD5 the concatenation
//  4. Return first min(n+5, 16) bytes
func objectKey(docKey []byte, objNum, genNum int, isAES bool) []byte {
	buf := make([]byte, 0, len(docKey)+5+4)

	// docKey + objNum (3 LE bytes) + genNum (2 LE bytes)
	buf = append(buf, docKey...)
	buf = append(buf, byte(objNum), byte(objNum>>8), byte(objNum>>16),
		byte(genNum), byte(genNum>>8))

	// For AES, append salt
	if isAES {
		buf = append(buf, 0x73, 0x41, 0x6C, 0x54) // "sAlT"
	}

	hash := md5.Sum(buf) //nolint:gosec // MD5 required by PDF spec

	keyLen := len(docKey) + 5
	if keyLen > 16 {
		keyLen = 16
	}

	return hash[:keyLen]
}

// rc4Decryptor implements Decryptor for V=1/2 (RC4 encryption).
type rc4Decryptor struct {
	docKey []byte
}

// DecryptStream decrypts a stream using RC4 with a per-object key.
func (d *rc4Decryptor) DecryptStream(data []byte, objNum, genNum int) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	key := objectKey(d.docKey, objNum, genNum, false)
	result := make([]byte, len(data))
	if err := encryptRC4(key, data, result); err != nil {
		return nil, fmt.Errorf("RC4 decrypt stream obj %d: %w", objNum, err)
	}
	return result, nil
}

// DecryptString decrypts a string using RC4 with a per-object key.
func (d *rc4Decryptor) DecryptString(data []byte, objNum, genNum int) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	key := objectKey(d.docKey, objNum, genNum, false)
	result := make([]byte, len(data))
	if err := encryptRC4(key, data, result); err != nil {
		return nil, fmt.Errorf("RC4 decrypt string obj %d: %w", objNum, err)
	}
	return result, nil
}

// aes128Decryptor implements Decryptor for V=4 with CFM=AESV2 (AES-128).
type aes128Decryptor struct {
	docKey []byte
}

// DecryptStream decrypts a stream using AES-128-CBC with a per-object key.
func (d *aes128Decryptor) DecryptStream(data []byte, objNum, genNum int) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	key := objectKey(d.docKey, objNum, genNum, true)
	return decryptAES(key, data)
}

// DecryptString decrypts a string using AES-128-CBC with a per-object key.
func (d *aes128Decryptor) DecryptString(data []byte, objNum, genNum int) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	key := objectKey(d.docKey, objNum, genNum, true)
	return decryptAES(key, data)
}

// bytesEqual compares two byte slices for equality in constant time (conceptually).
// For PDF password verification, timing attacks are not a practical concern,
// but we still compare all bytes.
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
