package signature

import (
	"fmt"
	"time"
)

// Option configures the signing process.
type Option func(*signConfig)

// signConfig holds the optional signing parameters.
type signConfig struct {
	reason      string
	location    string
	contactInfo string
	tsaURL      string
	signTime    time.Time
}

// WithReason sets the /Reason field in the PDF signature dictionary.
// This describes why the document was signed (e.g., "Approved", "Author").
func WithReason(reason string) Option {
	return func(cfg *signConfig) { cfg.reason = reason }
}

// WithLocation sets the /Location field in the PDF signature dictionary.
// This identifies the geographic location where the signature was applied.
func WithLocation(location string) Option {
	return func(cfg *signConfig) { cfg.location = location }
}

// WithContactInfo sets the /ContactInfo field in the PDF signature dictionary.
// This provides contact information for the signer (e.g., email address).
func WithContactInfo(info string) Option {
	return func(cfg *signConfig) { cfg.contactInfo = info }
}

// WithTimestamp enables PAdES B-T by requesting an RFC 3161 timestamp token from
// the specified TSA URL (e.g., "http://timestamp.digicert.com" or "http://freetsa.org/tsr").
// The token is embedded as an unsigned attribute in the CMS SignerInfo.
// Without this option, only PAdES B-B (basic signature) is produced.
func WithTimestamp(tsaURL string) Option {
	return func(cfg *signConfig) { cfg.tsaURL = tsaURL }
}

// withSignTime overrides the signing time. Exported for testing only via the
// unexported function; in production, time.Now() is always used.

// SignDocument signs pdfData with the given Signer and returns the signed PDF bytes.
//
// The signing process:
//  1. Parse the existing PDF trailer to determine the next object number
//  2. Append an incremental update containing the signature dictionary (with placeholders),
//     a widget annotation referencing it, and a new xref + trailer with /Prev
//  3. Patch /ByteRange with the actual byte offsets
//  4. Compute SHA-256 over the two signed byte ranges (everything except /Contents)
//  5. Build a CMS/PKCS#7 SignedData structure (with optional RFC 3161 timestamp)
//  6. Hex-encode the CMS and patch it into the /Contents placeholder
//  7. Return the complete signed PDF bytes
//
// The original pdfData bytes are never modified; a new allocation is returned.
//
// Default level: PAdES B-B. Add WithTimestamp to get PAdES B-T.
//
// Example:
//
//	key, cert, _ := signature.GenerateTestCertificate()
//	signer, _ := signature.NewLocalSigner(key, []*x509.Certificate{cert})
//	signed, err := signature.SignDocument(pdfBytes, signer,
//	    signature.WithReason("Approved"),
//	    signature.WithLocation("New York"),
//	)
func SignDocument(pdfData []byte, signer Signer, opts ...Option) ([]byte, error) {
	cfg := &signConfig{
		signTime: time.Now().UTC(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Step 1+2: Build PDF with incremental update and placeholder signature dict.
	sr, err := buildSignedPDF(pdfData, cfg)
	if err != nil {
		return nil, fmt.Errorf("signature: build signed PDF: %w", err)
	}

	// Step 4: Compute SHA-256 hash over the signed byte ranges.
	digest, err := computeByteRangeHash(sr.pdf, sr.byteRange)
	if err != nil {
		return nil, fmt.Errorf("signature: compute byte range hash: %w", err)
	}

	// Step 5a: Optionally fetch RFC 3161 timestamp token (B-T level).
	var tsaToken []byte
	if cfg.tsaURL != "" {
		// We need the signature bytes first to timestamp them.
		// For B-T: build CMS without timestamp first to get sig bytes, then timestamp.
		// Per RFC 3161 / PAdES B-T: the timestamp covers the signature value bytes.
		// We pass nil tsaToken for the first CMS build to get the raw signature value,
		// then request a timestamp on it, then rebuild CMS with the token included.
		cmsPrelim, err := buildCMS(digest, signer, cfg.signTime, nil)
		if err != nil {
			return nil, fmt.Errorf("signature: build preliminary CMS for timestamp: %w", err)
		}

		// Extract the signature value from the preliminary CMS to timestamp it.
		sigBytes, err := extractSignatureValue(cmsPrelim)
		if err != nil {
			return nil, fmt.Errorf("signature: extract signature value for TSA: %w", err)
		}

		tsa := NewTSAClient(cfg.tsaURL)
		tsaToken, err = tsa.RequestTimestamp(sigBytes)
		if err != nil {
			return nil, fmt.Errorf("signature: request timestamp: %w", err)
		}
	}

	// Step 5b: Build the final CMS SignedData (with timestamp token if present).
	cms, err := buildCMS(digest, signer, cfg.signTime, tsaToken)
	if err != nil {
		return nil, fmt.Errorf("signature: build CMS: %w", err)
	}

	// Step 6: Inject hex-encoded CMS into /Contents placeholder.
	signed, err := injectSignature(sr.pdf, sr.contentsOffset, sr.contentsLength, cms)
	if err != nil {
		return nil, fmt.Errorf("signature: inject signature: %w", err)
	}

	return signed, nil
}

// extractSignatureValue parses a DER-encoded CMS ContentInfo and returns the
// raw signature bytes from the first SignerInfo. Used for B-T timestamping:
// we build a preliminary CMS (without TSA token), extract the signature value,
// then request a timestamp on it before building the final CMS.
func extractSignatureValue(cmsData []byte) ([]byte, error) {
	info, err := parseCMSStructure(cmsData)
	if err != nil {
		return nil, fmt.Errorf("extract signature value: %w", err)
	}
	if len(info.rawSignature) == 0 {
		return nil, fmt.Errorf("no signature value in CMS")
	}
	return info.rawSignature, nil
}
