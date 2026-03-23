package signature

import (
	"bytes"
	"crypto"
	"encoding/asn1"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// TSAClient is an RFC 3161 Time-Stamp Authority client.
// It constructs a TimeStampReq, sends it to the TSA URL, and returns the
// DER-encoded timestamp token (a ContentInfo wrapping a TSTInfo structure).
//
// The timestamp token is embedded as an unsigned attribute in the CMS SignerInfo,
// elevating the signature from PAdES B-B to PAdES B-T level.
type TSAClient struct {
	// URL is the TSA HTTP endpoint (e.g., "http://timestamp.digicert.com").
	URL string

	// HTTPClient is used for the HTTP POST. If nil, http.DefaultClient is used.
	HTTPClient *http.Client
}

// NewTSAClient creates a TSA client for the given endpoint URL.
func NewTSAClient(url string) *TSAClient {
	return &TSAClient{URL: url}
}

// RequestTimestamp sends an RFC 3161 timestamp request for the given signature bytes.
// It computes SHA-256 of the signature, builds a TimeStampReq, sends it to the TSA,
// and returns the raw DER-encoded timestamp token (ContentInfo).
//
// The returned token is stored as an unsigned attribute in the CMS SignerInfo.
func (c *TSAClient) RequestTimestamp(signature []byte) ([]byte, error) {
	if c.URL == "" {
		return nil, errors.New("signature: TSA URL is empty")
	}

	h := crypto.SHA256.New()
	h.Write(signature)
	digest := h.Sum(nil)

	reqDER, err := buildTimestampReq(digest, crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("signature: build TSA request: %w", err)
	}

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Post(c.URL, "application/timestamp-query", bytes.NewReader(reqDER))
	if err != nil {
		return nil, fmt.Errorf("signature: TSA request to %s: %w", c.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signature: TSA returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("signature: read TSA response: %w", err)
	}

	return parseTimestampResp(body)
}

// --- ASN.1 structures for RFC 3161 ---

// tsaTimeStampReq is the ASN.1 TimeStampReq (RFC 3161 Section 2.4.1).
type tsaTimeStampReq struct {
	Version        int
	MessageImprint tsaMessageImprint
	CertReq        bool `asn1:"optional"`
}

// tsaMessageImprint carries the hash algorithm and digest being timestamped.
type tsaMessageImprint struct {
	HashAlgorithm algorithmIdentifier
	HashedMessage []byte
}

// tsaTimeStampResp is the ASN.1 TimeStampResp (RFC 3161 Section 2.4.2).
type tsaTimeStampResp struct {
	Status         tsaPKIStatusInfo
	TimeStampToken asn1.RawValue `asn1:"optional"`
}

// tsaPKIStatusInfo carries the TSA response status code.
type tsaPKIStatusInfo struct {
	// Status: 0 = granted, 1 = grantedWithMods, 2+ = rejection/waiting/revocation.
	Status int
}

// buildTimestampReq creates a DER-encoded RFC 3161 TimeStampReq for the given digest.
func buildTimestampReq(digest []byte, hashFunc crypto.Hash) ([]byte, error) {
	var hashOID asn1.ObjectIdentifier
	switch hashFunc {
	case crypto.SHA256:
		hashOID = oidSHA256
	case crypto.SHA384:
		hashOID = oidSHA384
	case crypto.SHA512:
		hashOID = oidSHA512
	default:
		return nil, fmt.Errorf("signature: unsupported hash function %v for TSA request", hashFunc)
	}

	req := tsaTimeStampReq{
		Version: 1,
		MessageImprint: tsaMessageImprint{
			HashAlgorithm: algorithmIdentifier{Algorithm: hashOID},
			HashedMessage: digest,
		},
		CertReq: true,
	}
	return asn1.Marshal(req)
}

// parseTimestampResp parses a DER-encoded TimeStampResp and returns the timestamp token.
// Status values 0 (granted) and 1 (grantedWithMods) are accepted; all others are errors.
func parseTimestampResp(data []byte) ([]byte, error) {
	var resp tsaTimeStampResp
	if _, err := asn1.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("signature: unmarshal TSA response: %w", err)
	}

	// Status 0 = granted, 1 = grantedWithMods.
	if resp.Status.Status > 1 {
		return nil, fmt.Errorf("signature: TSA rejected request with status %d", resp.Status.Status)
	}

	if len(resp.TimeStampToken.FullBytes) == 0 {
		return nil, errors.New("signature: TSA response contains no timestamp token")
	}

	return resp.TimeStampToken.FullBytes, nil
}
