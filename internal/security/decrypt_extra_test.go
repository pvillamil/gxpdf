package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAES128Decryptor_DecryptString_Empty tests the empty data fast path.
func TestAES128Decryptor_DecryptString_Empty(t *testing.T) {
	dec := &aes128Decryptor{docKey: make([]byte, 16)}

	result, err := dec.DecryptString([]byte{}, 1, 0)
	require.NoError(t, err)
	assert.Empty(t, result)
}

// TestAES128Decryptor_DecryptString_InvalidData verifies error on malformed input.
func TestAES128Decryptor_DecryptString_InvalidData(t *testing.T) {
	dec := &aes128Decryptor{docKey: make([]byte, 16)}

	// Data too short to contain a valid IV (16 bytes) + ciphertext will error.
	shortData := []byte{0x01, 0x02, 0x03}
	_, err := dec.DecryptString(shortData, 1, 0)
	assert.Error(t, err)
}

// TestAES128Decryptor_DecryptString_ValidBlock exercises the non-empty code path
// with a properly encrypted block.
func TestAES128Decryptor_DecryptString_ValidBlock(t *testing.T) {
	// We use a known 16-byte document key.
	docKey := []byte("0123456789abcdef")
	dec := &aes128Decryptor{docKey: docKey}

	// Build the per-object key the same way the decryptor does.
	objKey := objectKey(docKey, 1, 0, true)

	// Encrypt a block using the per-object key.
	plaintext := []byte("0123456789abcdef") // 16 bytes
	encrypted, err := encryptAES(objKey, plaintext)
	require.NoError(t, err)

	// Now decrypt using DecryptString which should derive the same per-object key.
	decrypted, err := dec.DecryptString(encrypted, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
