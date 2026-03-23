// Package signature provides PDF digital signature support.
//
// It implements PAdES B-B (basic CMS signature) and B-T (with RFC 3161 timestamp)
// using only the Go standard library. The package supports RSA and ECDSA keys.
//
// Basic usage:
//
//	signer, err := signature.NewLocalSigner(privateKey, []*x509.Certificate{cert})
//	signed, err := signature.SignDocument(pdfBytes, signer,
//	    signature.WithReason("Approved"),
//	    signature.WithLocation("New York"),
//	)
//
//	info, err := signature.Verify(signed)
package signature

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
)

// Signer performs the cryptographic signing operation for PDF signatures.
// Implementations may use local keys, HSMs, or KMS services.
type Signer interface {
	// Sign signs the given digest and returns the raw signature bytes.
	// The digest is computed over the DER-encoded signed attributes (SET-tagged).
	Sign(digest []byte) ([]byte, error)

	// Algorithm returns the signature algorithm used by this signer.
	Algorithm() Algorithm

	// CertificateChain returns the signer's certificate chain.
	// The first certificate must be the signing certificate.
	// Subsequent certificates are intermediate CA certificates.
	CertificateChain() []*x509.Certificate
}

// LocalSigner signs using a local private key (RSA or ECDSA).
// It implements the Signer interface and is suitable for software-based signing.
type LocalSigner struct {
	key   crypto.Signer
	certs []*x509.Certificate
	algo  Algorithm
}

// NewLocalSigner creates a LocalSigner from a local private key and certificate chain.
// The algorithm is auto-detected from the key type: RSA keys use SHA256WithRSA,
// ECDSA keys use SHA256WithECDSA. Use SetAlgorithm to override.
//
// The certs slice must contain at least the signing certificate as the first element.
// Intermediate certificates should follow in chain order.
func NewLocalSigner(key crypto.Signer, certs []*x509.Certificate) (*LocalSigner, error) {
	if key == nil {
		return nil, errors.New("signature: key must not be nil")
	}
	if len(certs) == 0 {
		return nil, errors.New("signature: at least one certificate is required")
	}

	var algo Algorithm
	switch key.(type) {
	case *rsa.PrivateKey:
		algo = SHA256WithRSA
	case *ecdsa.PrivateKey:
		algo = SHA256WithECDSA
	default:
		return nil, fmt.Errorf("signature: unsupported key type %T; use *rsa.PrivateKey or *ecdsa.PrivateKey", key)
	}

	return &LocalSigner{key: key, certs: certs, algo: algo}, nil
}

// SetAlgorithm overrides the auto-detected algorithm.
// Use this to select SHA-384 or SHA-512 variants.
func (s *LocalSigner) SetAlgorithm(algo Algorithm) {
	s.algo = algo
}

// Sign signs the given digest using the local private key.
// The digest must be pre-hashed with the algorithm returned by Algorithm().
func (s *LocalSigner) Sign(digest []byte) ([]byte, error) {
	return s.key.Sign(rand.Reader, digest, s.algo.HashFunc())
}

// Algorithm returns the signature algorithm configured for this signer.
func (s *LocalSigner) Algorithm() Algorithm { return s.algo }

// CertificateChain returns the certificate chain.
// The first certificate is the signing certificate.
func (s *LocalSigner) CertificateChain() []*x509.Certificate { return s.certs }
