package signature

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SignatureInfo holds all information extracted from a PDF digital signature.
// It is returned by Verify and ParseSignatureInfo.
type SignatureInfo struct {
	// SubFilter is the CMS encoding format (e.g., "adbe.pkcs7.detached").
	SubFilter string

	// SignedBy is the Common Name from the signing certificate.
	SignedBy string

	// SignedAt is the signing time from the signingTime signed attribute.
	SignedAt time.Time

	// Reason is the /Reason field from the PDF signature dictionary.
	Reason string

	// Location is the /Location field from the PDF signature dictionary.
	Location string

	// Valid is true if the signature cryptographically verifies and the
	// byte range hash matches the messageDigest attribute.
	Valid bool

	// Err holds the verification error when Valid is false. Nil when Valid is true.
	Err error

	// Certificate is the signing certificate extracted from the CMS.
	Certificate *x509.Certificate

	// HasTimestamp is true if the signature contains an RFC 3161 timestamp token.
	HasTimestamp bool

	// byteRange is the [offset0, len0, offset2, len2] parsed from /ByteRange.
	byteRange [4]int64

	// Internal fields used during verification.
	rawSignature   []byte
	signedAttrsRaw []byte // re-encoded as SET for verification
	messageDigest  []byte
}

// Verify extracts and cryptographically verifies all digital signatures in pdfData.
// It returns a slice of SignatureInfo — one entry per signature dictionary found.
//
// Each entry's Valid field indicates whether the signature is intact. When Valid
// is false, Err contains the specific failure reason.
//
// Verify performs two checks per signature:
//  1. Byte range integrity: recomputes SHA-256 over the signed ranges and
//     compares against the messageDigest signed attribute
//  2. Cryptographic verification: verifies the CMS signature over the signed attributes
//
// Note: Verify does not validate certificate chains or revocation status.
// Use standard x509 pool verification for full PKI validation.
func Verify(pdfData []byte) ([]*SignatureInfo, error) {
	locs, err := findSignatureDicts(pdfData)
	if err != nil {
		return nil, fmt.Errorf("signature: find signature dicts: %w", err)
	}
	if len(locs) == 0 {
		return nil, nil
	}

	results := make([]*SignatureInfo, 0, len(locs))
	for _, loc := range locs {
		info, verifyErr := verifyOne(pdfData, loc)
		if verifyErr != nil {
			// Return a SignatureInfo with Valid=false rather than aborting the whole Verify.
			results = append(results, &SignatureInfo{Valid: false, Err: verifyErr})
			continue
		}
		results = append(results, info)
	}
	return results, nil
}

// sigDictLocation records where a signature dictionary was found in the PDF byte stream.
type sigDictLocation struct {
	// start and end are the byte offsets of the signature dictionary string.
	start, end int
	// raw is the text of the dictionary.
	raw string
}

// findSignatureDicts scans pdfData for /Type /Sig dictionaries and returns their locations.
func findSignatureDicts(pdfData []byte) ([]sigDictLocation, error) {
	s := string(pdfData)
	var locs []sigDictLocation

	// Find all occurrences of /Type /Sig — each is a signature dictionary.
	re := regexp.MustCompile(`<<[^<>]*?/Type\s*/Sig[^<>]*?(?:/ByteRange\s*\[\s*\d+\s+\d+\s+\d+\s+\d+\s*\])[^>]*?>>`)
	matches := re.FindAllStringIndex(s, -1)
	for _, m := range matches {
		locs = append(locs, sigDictLocation{start: m[0], end: m[1], raw: s[m[0]:m[1]]})
	}

	// Fallback: look for /ByteRange directly (handles multi-line dicts).
	if len(locs) == 0 {
		brRe := regexp.MustCompile(`/ByteRange\s*\[\s*\d+\s+\d+\s+\d+\s+\d+\s*\]`)
		brMatches := brRe.FindAllStringIndex(s, -1)
		for _, m := range brMatches {
			// Walk back to find << and forward to find >>.
			start := strings.LastIndex(s[:m[0]], "<<")
			end := strings.Index(s[m[1]:], ">>")
			if start < 0 || end < 0 {
				continue
			}
			end = m[1] + end + 2
			locs = append(locs, sigDictLocation{start: start, end: end, raw: s[start:end]})
		}
	}

	return locs, nil
}

// verifyOne extracts and verifies a single signature dictionary.
func verifyOne(pdfData []byte, loc sigDictLocation) (*SignatureInfo, error) {
	info := &SignatureInfo{}

	// Extract PDF-level metadata from the dictionary text.
	extractPDFFields(info, loc.raw)

	// Parse /ByteRange.
	br, err := parseByteRangeStr(loc.raw)
	if err != nil {
		return nil, fmt.Errorf("signature: parse /ByteRange: %w", err)
	}
	info.byteRange = br

	// Extract /Contents hex string.
	cmsData, err := extractContentsHex(loc.raw)
	if err != nil {
		return nil, fmt.Errorf("signature: extract /Contents: %w", err)
	}

	// Parse CMS structure.
	cms, err := parseCMSStructure(cmsData)
	if err != nil {
		return nil, fmt.Errorf("signature: parse CMS: %w", err)
	}
	info.rawSignature = cms.rawSignature
	info.signedAttrsRaw = cms.signedAttrsSet
	info.messageDigest = cms.messageDigest
	info.Certificate = cms.certificate
	info.HasTimestamp = cms.hasTimestamp

	if info.Certificate != nil {
		info.SignedBy = info.Certificate.Subject.CommonName
	}
	info.SignedAt = cms.signingTime

	// Verify byte range hash.
	if err := verifyByteRangeHash(pdfData, info); err != nil {
		info.Valid = false
		info.Err = err
		return info, nil
	}

	// Cryptographic signature verification.
	if err := verifyCMSSignature(info); err != nil {
		info.Valid = false
		info.Err = err
		return info, nil
	}

	info.Valid = true
	return info, nil
}

// verifyByteRangeHash recomputes the SHA-256 hash over the signed byte ranges
// and compares it against the messageDigest from the CMS signed attributes.
func verifyByteRangeHash(pdfData []byte, info *SignatureInfo) error {
	hash, err := computeByteRangeHash(pdfData, info.byteRange)
	if err != nil {
		return fmt.Errorf("compute byte range hash: %w", err)
	}
	if len(info.messageDigest) == 0 {
		return fmt.Errorf("no messageDigest in signed attributes")
	}
	if !constantTimeEqual(hash, info.messageDigest) {
		return fmt.Errorf("byte range hash does not match messageDigest: document may have been modified")
	}
	return nil
}

// verifyCMSSignature cryptographically verifies the CMS signature using the
// certificate's public key. The signed attributes are re-encoded as a SET before hashing.
func verifyCMSSignature(info *SignatureInfo) error {
	if info.Certificate == nil {
		return fmt.Errorf("no certificate in CMS")
	}
	if len(info.signedAttrsRaw) == 0 {
		return fmt.Errorf("no signed attributes")
	}
	if len(info.rawSignature) == 0 {
		return fmt.Errorf("no signature value")
	}

	// Hash the signed attributes (must be SET-tagged 0x31).
	attrHash := crypto.SHA256.New()
	attrHash.Write(info.signedAttrsRaw)
	digest := attrHash.Sum(nil)

	switch key := info.Certificate.PublicKey.(type) {
	case *rsa.PublicKey:
		return rsa.VerifyPKCS1v15(key, crypto.SHA256, digest, info.rawSignature)
	case *ecdsa.PublicKey:
		if !ecdsa.VerifyASN1(key, digest, info.rawSignature) {
			return fmt.Errorf("ECDSA signature verification failed")
		}
		return nil
	default:
		return fmt.Errorf("unsupported public key type %T", key)
	}
}

// --- CMS parsing ---

// cmsParseResult holds the structural fields extracted from a CMS SignedData.
type cmsParseResult struct {
	certificate    *x509.Certificate
	rawSignature   []byte
	signedAttrsSet []byte // re-encoded as SET (0x31) for verification
	messageDigest  []byte
	signingTime    time.Time
	hasTimestamp   bool
}

// parseCMSStructure parses a DER-encoded CMS ContentInfo / SignedData and extracts
// fields needed for verification and display.
func parseCMSStructure(cmsData []byte) (*cmsParseResult, error) {
	// Parse ContentInfo wrapper.
	type cmsContentInfo struct {
		ContentType asn1.ObjectIdentifier
		Content     asn1.RawValue `asn1:"explicit,tag:0"`
	}
	var ci cmsContentInfo
	if _, err := asn1.Unmarshal(cmsData, &ci); err != nil {
		return nil, fmt.Errorf("unmarshal ContentInfo: %w", err)
	}

	// Parse SignedData.
	type cmsSignedData struct {
		Version          int
		DigestAlgorithms asn1.RawValue `asn1:"set"`
		EncapContentInfo asn1.RawValue
		Certificates     asn1.RawValue `asn1:"optional,tag:0"`
		SignerInfos      asn1.RawValue `asn1:"set"`
	}
	var sd cmsSignedData
	if _, err := asn1.Unmarshal(ci.Content.Bytes, &sd); err != nil {
		return nil, fmt.Errorf("unmarshal SignedData: %w", err)
	}

	result := &cmsParseResult{}

	// Parse the first certificate from IMPLICIT [0] SET OF Certificate.
	// When multiple certificates are concatenated (cert chain), ParseCertificate
	// would fail with "trailing data". Use ParseCertificates which handles the
	// concatenated DER format and returns all certs; we use the first (signer cert).
	if len(sd.Certificates.Bytes) > 0 {
		certs, err := x509.ParseCertificates(sd.Certificates.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse certificates: %w", err)
		}
		if len(certs) > 0 {
			result.certificate = certs[0]
		}
	}

	// Extract the first SignerInfo from the SET.
	siBytes, err := firstSetElement(sd.SignerInfos.FullBytes)
	if err != nil {
		return nil, fmt.Errorf("extract SignerInfo: %w", err)
	}

	// Parse SignerInfo.
	type cmsSignerInfo struct {
		Version         int
		SID             asn1.RawValue
		DigestAlgorithm asn1.RawValue
		SignedAttrs     asn1.RawValue `asn1:"optional,tag:0"`
		SignatureAlg    asn1.RawValue
		Signature       []byte
		UnsignedAttrs   asn1.RawValue `asn1:"optional,tag:1"`
	}
	var si cmsSignerInfo
	if _, err := asn1.Unmarshal(siBytes, &si); err != nil {
		return nil, fmt.Errorf("unmarshal SignerInfo: %w", err)
	}

	result.rawSignature = si.Signature

	// Re-encode signed attributes as SET (0x31) for verification.
	// The SignerInfo stores them as IMPLICIT [0], but verification requires SET tag.
	if len(si.SignedAttrs.Bytes) > 0 {
		result.signedAttrsSet = reEncodeAsSet(si.SignedAttrs.Bytes)
		result.messageDigest = extractMessageDigest(si.SignedAttrs.Bytes)
		result.signingTime = extractSigningTime(si.SignedAttrs.Bytes)
	}

	// Detect timestamp token in unsigned attributes.
	if len(si.UnsignedAttrs.Bytes) > 0 {
		result.hasTimestamp = hasTimestampToken(si.UnsignedAttrs.Bytes)
	}

	return result, nil
}

// reEncodeAsSet wraps the inner content bytes of signed attributes with SET tag 0x31.
// The IMPLICIT [0] tag in the SignerInfo must be converted back to SET for hashing.
func reEncodeAsSet(innerBytes []byte) []byte {
	return marshalSet(innerBytes)
}

// firstSetElement extracts the first element from a DER-encoded SET.
func firstSetElement(setDER []byte) ([]byte, error) {
	var raw asn1.RawValue
	if _, err := asn1.Unmarshal(setDER, &raw); err != nil {
		return nil, err
	}
	return raw.Bytes, nil
}

// extractMessageDigest walks the signed attributes bytes to find id-messageDigest
// (OID 1.2.840.113549.1.9.4) and returns its OCTET STRING value.
func extractMessageDigest(attrsBytes []byte) []byte {
	rest := attrsBytes
	for len(rest) > 0 {
		var attr struct {
			Type   asn1.ObjectIdentifier
			Values asn1.RawValue `asn1:"set"`
		}
		var err error
		rest, err = asn1.Unmarshal(rest, &attr)
		if err != nil {
			break
		}
		if attr.Type.Equal(oidMessageDigest) {
			var digest []byte
			if _, err := asn1.Unmarshal(attr.Values.Bytes, &digest); err == nil {
				return digest
			}
		}
	}
	return nil
}

// extractSigningTime walks the signed attributes bytes to find id-signingTime
// (OID 1.2.840.113549.1.9.5) and parses it as a time.Time.
func extractSigningTime(attrsBytes []byte) time.Time {
	rest := attrsBytes
	for len(rest) > 0 {
		var attr struct {
			Type   asn1.ObjectIdentifier
			Values asn1.RawValue `asn1:"set"`
		}
		var err error
		rest, err = asn1.Unmarshal(rest, &attr)
		if err != nil {
			break
		}
		if attr.Type.Equal(oidSigningTime) {
			var t time.Time
			if _, err := asn1.Unmarshal(attr.Values.Bytes, &t); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// hasTimestampToken walks the unsigned attributes bytes to detect the presence of
// id-aa-timeStampToken (OID 1.2.840.113549.1.9.16.2.14).
func hasTimestampToken(attrsBytes []byte) bool {
	rest := attrsBytes
	for len(rest) > 0 {
		var attr struct {
			Type   asn1.ObjectIdentifier
			Values asn1.RawValue `asn1:"set"`
		}
		var err error
		rest, err = asn1.Unmarshal(rest, &attr)
		if err != nil {
			break
		}
		if attr.Type.Equal(oidTimeStampToken) {
			return true
		}
	}
	return false
}

// --- PDF-level parsing helpers ---

var reByteRange = regexp.MustCompile(`/ByteRange\s*\[\s*(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s*\]`)

// parseByteRangeStr extracts the four ByteRange integers from a PDF dictionary string.
func parseByteRangeStr(s string) ([4]int64, error) {
	m := reByteRange.FindStringSubmatch(s)
	if m == nil {
		return [4]int64{}, fmt.Errorf("/ByteRange not found")
	}
	var br [4]int64
	for i := 0; i < 4; i++ {
		v, err := strconv.ParseInt(m[i+1], 10, 64)
		if err != nil {
			return [4]int64{}, fmt.Errorf("invalid ByteRange value %q: %w", m[i+1], err)
		}
		br[i] = v
	}
	return br, nil
}

// extractContentsHex finds the /Contents <HEXSTRING> field, decodes the full hex string,
// then determines the actual DER length from the ASN.1 ContentInfo header so that
// trailing zero-padding from the placeholder is excluded without stripping real zero bytes.
func extractContentsHex(s string) ([]byte, error) {
	re := regexp.MustCompile(`/Contents\s*<([0-9A-Fa-f]+)>`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		return nil, fmt.Errorf("/Contents hex string not found")
	}
	hexStr := m[1]
	if len(hexStr) == 0 {
		return nil, fmt.Errorf("/Contents is empty")
	}
	// Ensure even length for hex decoding.
	if len(hexStr)%2 != 0 {
		hexStr += "0"
	}
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("decode /Contents hex: %w", err)
	}

	// Determine the true CMS length from the DER TLV header of the outer SEQUENCE.
	// This avoids stripping real trailing zero bytes from the CMS structure.
	cmsLen, err := derContentLength(raw)
	if err != nil {
		// Fallback: trim trailing zeros (less safe but maintains compatibility).
		trimmed := strings.TrimRight(hexStr, "0")
		if len(trimmed) == 0 {
			return nil, fmt.Errorf("/Contents appears to be all zeros — document not yet signed")
		}
		if len(trimmed)%2 != 0 {
			trimmed += "0"
		}
		return hex.DecodeString(trimmed)
	}
	if cmsLen > len(raw) {
		return nil, fmt.Errorf("DER length %d exceeds /Contents size %d", cmsLen, len(raw))
	}
	return raw[:cmsLen], nil
}

// derContentLength reads the DER tag+length from the start of data and returns
// the total encoded length (tag + length octets + value octets).
// This is used to determine the exact CMS byte count, excluding trailing zero padding.
func derContentLength(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("DER data too short")
	}
	// Skip tag byte.
	pos := 1
	b := data[pos]
	var valueLen int
	if b&0x80 == 0 {
		// Short form.
		valueLen = int(b)
		pos++
	} else {
		numBytes := int(b & 0x7F)
		if numBytes == 0 || numBytes > 4 {
			return 0, fmt.Errorf("DER: unsupported length encoding (numBytes=%d)", numBytes)
		}
		pos++
		if pos+numBytes > len(data) {
			return 0, fmt.Errorf("DER: length octets exceed data")
		}
		for i := 0; i < numBytes; i++ {
			valueLen = (valueLen << 8) | int(data[pos+i])
		}
		pos += numBytes
	}
	total := pos + valueLen
	if total > len(data) {
		return 0, fmt.Errorf("DER: total length %d exceeds data length %d", total, len(data))
	}
	return total, nil
}

// extractPDFFields populates the string metadata fields of info from the
// signature dictionary text.
func extractPDFFields(info *SignatureInfo, dictText string) {
	if m := regexp.MustCompile(`/SubFilter\s*/(\S+)`).FindStringSubmatch(dictText); m != nil {
		info.SubFilter = m[1]
	}
	if m := regexp.MustCompile(`/Reason\s*\(([^)]*)\)`).FindStringSubmatch(dictText); m != nil {
		info.Reason = m[1]
	}
	if m := regexp.MustCompile(`/Location\s*\(([^)]*)\)`).FindStringSubmatch(dictText); m != nil {
		info.Location = m[1]
	}
}

// constantTimeEqual compares two byte slices in constant time to prevent timing attacks.
func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

// --- Convenience helpers on SignatureInfo ---

// IsRSA reports whether the signing certificate uses an RSA public key.
func (info *SignatureInfo) IsRSA() bool {
	if info.Certificate == nil {
		return false
	}
	_, ok := info.Certificate.PublicKey.(*rsa.PublicKey)
	return ok
}

// IsECDSA reports whether the signing certificate uses an ECDSA public key.
func (info *SignatureInfo) IsECDSA() bool {
	if info.Certificate == nil {
		return false
	}
	_, ok := info.Certificate.PublicKey.(*ecdsa.PublicKey)
	return ok
}

// RSAKeySize returns the RSA key size in bits, or 0 if not RSA.
func (info *SignatureInfo) RSAKeySize() int {
	if info.Certificate == nil {
		return 0
	}
	if key, ok := info.Certificate.PublicKey.(*rsa.PublicKey); ok {
		return key.N.BitLen()
	}
	return 0
}

// ECDSACurve returns the ECDSA curve name (e.g., "P-256"), or "" if not ECDSA.
func (info *SignatureInfo) ECDSACurve() string {
	if info.Certificate == nil {
		return ""
	}
	if key, ok := info.Certificate.PublicKey.(*ecdsa.PublicKey); ok {
		return key.Curve.Params().Name
	}
	return ""
}

// IsECDSAP256 reports whether the ECDSA key uses the P-256 curve.
func (info *SignatureInfo) IsECDSAP256() bool {
	if info.Certificate == nil {
		return false
	}
	key, ok := info.Certificate.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false
	}
	return key.Curve == elliptic.P256()
}

// TamperByte returns a copy of pdfData with the byte at offset XORed with 0xFF.
// Used in tests to simulate document tampering within the signed byte ranges.
func TamperByte(pdfData []byte, offset int) []byte {
	if offset < 0 || offset >= len(pdfData) {
		return pdfData
	}
	out := make([]byte, len(pdfData))
	copy(out, pdfData)
	out[offset] ^= 0xFF
	return out
}

// FindSignedByteOffset returns a byte offset that lies within the first signed byte
// range — safe to tamper with for testing tamper detection. Returns -1 if byteRange is empty.
func (info *SignatureInfo) FindSignedByteOffset() int {
	if info.byteRange[1] < 20 {
		return -1
	}
	// Use a position in the middle of the first range, avoiding the PDF header.
	mid := int(info.byteRange[1] / 2)
	if mid < 10 {
		mid = 10
	}
	return mid
}

// ParseSignatureInfo is a lower-level alternative to Verify that extracts CMS data
// and metadata without running cryptographic verification. It returns a SignatureInfo
// with Valid always false (verification is left to the caller).
//
// Use Verify for complete integrity checking.
func ParseSignatureInfo(pdfData []byte) (*SignatureInfo, error) {
	s := string(pdfData)

	// Find the first signature dictionary.
	brIdx := reByteRange.FindStringIndex(s)
	if brIdx == nil {
		return nil, fmt.Errorf("signature: no /ByteRange found in PDF")
	}
	start := strings.LastIndex(s[:brIdx[0]], "<<")
	end := strings.Index(s[brIdx[1]:], ">>")
	if start < 0 || end < 0 {
		return nil, fmt.Errorf("signature: cannot locate signature dictionary boundaries")
	}
	dictText := s[start : brIdx[1]+end+2]

	locs := []sigDictLocation{{start: start, end: brIdx[1] + end + 2, raw: dictText}}

	if len(locs) == 0 {
		return nil, fmt.Errorf("signature: no signature dictionary found")
	}

	info := &SignatureInfo{}
	extractPDFFields(info, locs[0].raw)

	br, err := parseByteRangeStr(locs[0].raw)
	if err != nil {
		return nil, fmt.Errorf("signature: parse /ByteRange: %w", err)
	}
	info.byteRange = br

	cmsData, err := extractContentsHex(locs[0].raw)
	if err != nil {
		return nil, fmt.Errorf("signature: extract /Contents: %w", err)
	}

	cms, err := parseCMSStructure(cmsData)
	if err != nil {
		return nil, fmt.Errorf("signature: parse CMS: %w", err)
	}
	info.rawSignature = cms.rawSignature
	info.signedAttrsRaw = cms.signedAttrsSet
	info.messageDigest = cms.messageDigest
	info.Certificate = cms.certificate
	info.HasTimestamp = cms.hasTimestamp
	info.SignedAt = cms.signingTime
	if info.Certificate != nil {
		info.SignedBy = info.Certificate.Subject.CommonName
	}

	return info, nil
}
