package signature

import (
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"fmt"
	"time"
)

// buildCMS constructs a DER-encoded CMS SignedData structure suitable for embedding
// in a PDF signature dictionary's /Contents field.
//
// Parameters:
//   - digest: SHA-256 (or other) hash of the PDF byte ranges being signed
//   - signer: provides the signing operation, algorithm, and certificate chain
//   - signingTime: UTC time to embed in the signingTime signed attribute
//   - tsaToken: optional DER-encoded RFC 3161 timestamp token; nil for B-B level
//
// The result is a detached CMS signature (SubFilter: adbe.pkcs7.detached).
func buildCMS(digest []byte, signer Signer, signingTime time.Time, tsaToken []byte) ([]byte, error) {
	certs := signer.CertificateChain()
	if len(certs) == 0 {
		return nil, errors.New("signature: no certificates in signer chain")
	}
	algo := signer.Algorithm()
	signingCert := certs[0]

	// Build signed attributes (contentType, messageDigest, signingTime, ESS sigCertV2).
	signedAttrs, err := buildSignedAttributes(digest, signingTime, signingCert, algo)
	if err != nil {
		return nil, fmt.Errorf("signature: build signed attributes: %w", err)
	}

	// DER-encode signed attributes as SET OF for hashing.
	// Critical: marshalAttributes produces a SET-tagged (0x31) encoding.
	signedAttrBytes, err := marshalAttributes(signedAttrs)
	if err != nil {
		return nil, fmt.Errorf("signature: marshal signed attributes: %w", err)
	}

	// Hash the signed attributes using the configured algorithm.
	h := algo.HashFunc().New()
	h.Write(signedAttrBytes)
	attrDigest := h.Sum(nil)

	// Sign the attribute digest.
	sig, err := signer.Sign(attrDigest)
	if err != nil {
		return nil, fmt.Errorf("signature: sign: %w", err)
	}

	// Encode certificates (signing cert first, then intermediates).
	var certsDER []byte
	for _, cert := range certs {
		certsDER = append(certsDER, cert.Raw...)
	}

	// Build digest algorithm SET.
	digestAlg := algorithmIdentifier{Algorithm: algo.DigestOID()}
	digestAlgDER, err := asn1.Marshal(digestAlg)
	if err != nil {
		return nil, fmt.Errorf("signature: marshal digest algorithm: %w", err)
	}
	digestAlgSet := marshalSet(digestAlgDER)

	// Build SignerInfo.
	siDER, err := marshalSignerInfo(signingCert, algo, signedAttrBytes, sig, tsaToken)
	if err != nil {
		return nil, fmt.Errorf("signature: marshal signer info: %w", err)
	}
	siSet := marshalSet(siDER)

	// Build SignedData.
	sd := signedData{
		Version:          1,
		DigestAlgorithms: asn1.RawValue{FullBytes: digestAlgSet},
		EncapContentInfo: encapContentInfo{ContentType: oidData},
		Certificates: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      certsDER,
		},
		SignerInfos: asn1.RawValue{FullBytes: siSet},
	}
	sdDER, err := asn1.Marshal(sd)
	if err != nil {
		return nil, fmt.Errorf("signature: marshal SignedData: %w", err)
	}

	// Wrap in ContentInfo.
	ci := contentInfo{
		ContentType: oidSignedData,
		Content: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      sdDER,
		},
	}
	result, err := asn1.Marshal(ci)
	if err != nil {
		return nil, fmt.Errorf("signature: marshal ContentInfo: %w", err)
	}
	return result, nil
}

// buildSignedAttributes constructs the CMS signed attributes required for PAdES B-B:
//   - id-contentType: id-data
//   - id-messageDigest: SHA-256 hash of the PDF byte ranges
//   - id-signingTime: signing timestamp
//   - id-aa-signingCertificateV2: ESS certificate hash (required for PAdES B-B)
func buildSignedAttributes(digest []byte, signingTime time.Time, cert *x509.Certificate, algo Algorithm) ([]attribute, error) {
	// Content type: id-data.
	contentTypeVal, err := asn1.Marshal(oidData)
	if err != nil {
		return nil, fmt.Errorf("signature: marshal content type: %w", err)
	}

	// Message digest: SHA-256 hash of the byte ranges.
	digestVal, err := asn1.Marshal(asn1.RawValue{
		Class: asn1.ClassUniversal,
		Tag:   asn1.TagOctetString,
		Bytes: digest,
	})
	if err != nil {
		return nil, fmt.Errorf("signature: marshal message digest: %w", err)
	}

	// Signing time.
	timeVal, err := asn1.Marshal(signingTime.UTC())
	if err != nil {
		return nil, fmt.Errorf("signature: marshal signing time: %w", err)
	}

	//nolint:prealloc // capacity known, literal initialization is clearer
	attrs := []attribute{
		{
			Type:   oidContentType,
			Values: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: contentTypeVal},
		},
		{
			Type:   oidMessageDigest,
			Values: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: digestVal},
		},
		{
			Type:   oidSigningTime,
			Values: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: timeVal},
		},
	}

	// ESS signing-certificate-v2 (OID 1.2.840.113549.1.9.16.2.47).
	// Required for PAdES B-B compliance. Contains SHA-256 hash of the signing certificate.
	certHash := hashData(algo.HashFunc(), cert.Raw)
	essCert := signingCertificateV2{
		Certs: []essCertIDv2{
			{
				HashAlgorithm: algorithmIdentifier{Algorithm: algo.DigestOID()},
				CertHash:      certHash,
			},
		},
	}
	essDER, err := asn1.Marshal(essCert)
	if err != nil {
		return nil, fmt.Errorf("signature: marshal ESS signing-certificate-v2: %w", err)
	}
	attrs = append(attrs, attribute{
		Type:   oidSigningCertificateV2,
		Values: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: essDER},
	})

	return attrs, nil
}

// marshalAttributes DER-encodes a list of CMS attributes as a SET OF.
// The result is tagged with SET (0x31) so it can be used directly for hashing.
// This is the encoding required when computing the signature over signed attributes.
func marshalAttributes(attrs []attribute) ([]byte, error) {
	var attrsDER []byte
	for _, attr := range attrs {
		b, err := asn1.Marshal(attr)
		if err != nil {
			return nil, fmt.Errorf("signature: marshal attribute: %w", err)
		}
		attrsDER = append(attrsDER, b...)
	}
	return marshalSet(attrsDER), nil
}

// marshalSignerInfo builds and DER-encodes the CMS SignerInfo structure.
//
// The signedAttrsSet parameter must be the SET-tagged encoding of signed attributes
// (as produced by marshalAttributes). The inner bytes are extracted and re-encoded
// with the IMPLICIT [0] context tag as required by the CMS SignerInfo structure.
//
// If tsaToken is non-nil, it is included as an unsigned attribute (id-aa-timeStampToken).
func marshalSignerInfo(cert *x509.Certificate, algo Algorithm, signedAttrsSet, sig, tsaToken []byte) ([]byte, error) {
	// Use raw issuer bytes from certificate for IssuerAndSerialNumber.
	issuerDER := cert.RawIssuer

	serialDER, err := asn1.Marshal(cert.SerialNumber)
	if err != nil {
		return nil, fmt.Errorf("signature: marshal serial number: %w", err)
	}

	// Strip the outer SET tag from signedAttrsSet to get inner content bytes.
	// The SignerInfo signed attributes field uses IMPLICIT [0] tag (not SET).
	innerAttrs := stripTag(signedAttrsSet)

	si := signerInfo{
		Version: 1,
		SID: issuerAndSerialNumber{
			Issuer:       asn1.RawValue{FullBytes: issuerDER},
			SerialNumber: asn1.RawValue{FullBytes: serialDER},
		},
		DigestAlgorithm: algorithmIdentifier{Algorithm: algo.DigestOID()},
		SignedAttrs: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      innerAttrs,
		},
		SignatureAlgorithm: algorithmIdentifier{Algorithm: algo.SignatureOID()},
		Signature:          sig,
	}

	// Add timestamp token as unsigned attribute (B-T level).
	if len(tsaToken) > 0 {
		tsaAttr := attribute{
			Type:   oidTimeStampToken,
			Values: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: tsaToken},
		}
		tsaDER, err := asn1.Marshal(tsaAttr)
		if err != nil {
			return nil, fmt.Errorf("signature: marshal timestamp attribute: %w", err)
		}
		si.UnsignedAttrs = asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        1,
			IsCompound: true,
			Bytes:      tsaDER,
		}
	}

	return asn1.Marshal(si)
}

// marshalSet wraps DER content bytes in an ASN.1 SET tag (0x31).
func marshalSet(content []byte) []byte {
	return marshalTLV(0x31, content)
}

// marshalTLV creates a DER TLV (tag-length-value) encoding.
// Handles both short-form (< 128 bytes) and long-form length encoding.
func marshalTLV(tag byte, content []byte) []byte {
	l := len(content)
	if l < 128 {
		result := make([]byte, 2+l)
		result[0] = tag
		result[1] = byte(l)
		copy(result[2:], content)
		return result
	}
	// Long-form length encoding.
	var lenBytes []byte
	n := l
	for n > 0 {
		lenBytes = append([]byte{byte(n & 0xFF)}, lenBytes...)
		n >>= 8
	}
	result := make([]byte, 2+len(lenBytes)+l)
	result[0] = tag
	result[1] = byte(0x80 | len(lenBytes))
	copy(result[2:], lenBytes)
	copy(result[2+len(lenBytes):], content)
	return result
}

// stripTag strips the outer ASN.1 tag and length bytes, returning only the value content.
// Used to extract inner bytes from a SET-tagged encoding for re-tagging.
func stripTag(der []byte) []byte {
	if len(der) < 2 {
		return der
	}
	pos := 1 // skip tag byte
	if der[pos]&0x80 == 0 {
		// Short-form length.
		pos++
	} else {
		// Long-form length: skip 1 byte + numBytes length octets.
		numBytes := int(der[pos] & 0x7F)
		pos += 1 + numBytes
	}
	if pos >= len(der) {
		return der
	}
	return der[pos:]
}

// hashData computes the hash of data using the specified crypto.Hash algorithm.
func hashData(h crypto.Hash, data []byte) []byte {
	hasher := h.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}
