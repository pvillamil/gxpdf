package signature

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

// GenerateTestCertificate generates a self-signed RSA-2048 certificate and key pair
// suitable for testing PDF signatures. The certificate has KeyUsageDigitalSignature.
//
// Do not use in production. Test certificates are not trusted by any root store.
func GenerateTestCertificate() (crypto.Signer, *x509.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	cert, err := selfSignedCert(key, &key.PublicKey, "Test RSA Signer", big.NewInt(1))
	if err != nil {
		return nil, nil, err
	}
	return key, cert, nil
}

// GenerateTestECCertificate generates a self-signed ECDSA P-256 certificate and key pair
// suitable for testing PDF signatures. The certificate has KeyUsageDigitalSignature.
//
// Do not use in production. Test certificates are not trusted by any root store.
func GenerateTestECCertificate() (crypto.Signer, *x509.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	cert, err := selfSignedCert(key, key.Public(), "Test ECDSA Signer", big.NewInt(2))
	if err != nil {
		return nil, nil, err
	}
	return key, cert, nil
}

// selfSignedCert creates a self-signed DER certificate and parses it.
func selfSignedCert(signer crypto.Signer, pub crypto.PublicKey, cn string, serial *big.Int) (*x509.Certificate, error) {
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   cn,
			Organization: []string{"GxPDF Test"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, pub, signer)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDER)
}
