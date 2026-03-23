package signature_test

import (
	"bytes"
	"crypto/x509"
	"encoding/asn1"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coregx/gxpdf/signature"
)

// TestNewTSAClient verifies the constructor sets the URL.
func TestNewTSAClient(t *testing.T) {
	c := signature.NewTSAClient("http://tsa.example.com")
	if c == nil {
		t.Fatal("NewTSAClient returned nil")
	}
	if c.URL != "http://tsa.example.com" {
		t.Errorf("URL = %q, want %q", c.URL, "http://tsa.example.com")
	}
}

// TestTSAClientEmptyURL verifies an error is returned for empty TSA URL.
func TestTSAClientEmptyURL(t *testing.T) {
	c := &signature.TSAClient{} // empty URL
	_, err := c.RequestTimestamp([]byte("test"))
	if err == nil {
		t.Fatal("expected error for empty TSA URL")
	}
}

// TestTSAClientHTTPError verifies the TSA client propagates HTTP errors.
func TestTSAClientHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := signature.NewTSAClient(srv.URL)
	c.HTTPClient = srv.Client()

	_, err := c.RequestTimestamp([]byte("digest"))
	if err == nil {
		t.Fatal("expected error for HTTP 503")
	}
}

// TestTSAClientBadResponse verifies the TSA client rejects malformed ASN.1.
func TestTSAClientBadResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/timestamp-reply")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid asn1"))
	}))
	defer srv.Close()

	c := signature.NewTSAClient(srv.URL)
	c.HTTPClient = srv.Client()

	_, err := c.RequestTimestamp([]byte("digest"))
	if err == nil {
		t.Fatal("expected error for malformed TSA response")
	}
}

// TestTSAClientRejectedStatus verifies status > 1 is an error.
func TestTSAClientRejectedStatus(t *testing.T) {
	// Build a TimeStampResp with status = 2 (rejection).
	type pkiStatusInfo struct {
		Status int
	}
	type timeStampResp struct {
		Status pkiStatusInfo
	}
	resp := timeStampResp{Status: pkiStatusInfo{Status: 2}}
	der, err := asn1.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal rejected response: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/timestamp-reply")
		_, _ = w.Write(der)
	}))
	defer srv.Close()

	c := signature.NewTSAClient(srv.URL)
	c.HTTPClient = srv.Client()

	_, err = c.RequestTimestamp([]byte("sig"))
	if err == nil {
		t.Fatal("expected error for TSA rejected status")
	}
}

// TestTSAClientGrantedWithMods verifies status = 1 (grantedWithMods) is accepted.
func TestTSAClientGrantedWithMods(t *testing.T) {
	// Build a TimeStampResp with status = 1 (grantedWithMods) and a dummy token.
	dummyToken := []byte{0x30, 0x05, 0x06, 0x03, 0x55, 0x04, 0x03} // dummy SEQUENCE

	type pkiStatusInfo struct {
		Status int
	}
	type timeStampResp struct {
		Status         pkiStatusInfo
		TimeStampToken asn1.RawValue `asn1:"optional"`
	}
	resp := timeStampResp{
		Status:         pkiStatusInfo{Status: 1},
		TimeStampToken: asn1.RawValue{FullBytes: dummyToken},
	}
	der, err := asn1.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal grantedWithMods response: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/timestamp-reply")
		_, _ = w.Write(der)
	}))
	defer srv.Close()

	c := signature.NewTSAClient(srv.URL)
	c.HTTPClient = srv.Client()

	token, err := c.RequestTimestamp([]byte("sig"))
	if err != nil {
		t.Fatalf("grantedWithMods should succeed: %v", err)
	}
	if !bytes.Equal(token, dummyToken) {
		t.Errorf("returned token does not match expected dummy token")
	}
}

// TestTSAClientSendsMimeType verifies the TSA request uses the correct Content-Type.
func TestTSAClientSendsMimeType(t *testing.T) {
	var receivedContentType string

	// Return a rejected status so the function returns an error without trying to parse a token.
	type pkiStatus struct{ Status int }
	type tsResp struct{ Status pkiStatus }
	rejDER, _ := asn1.Marshal(tsResp{Status: pkiStatus{Status: 2}})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		// Read and discard the body.
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write(rejDER)
	}))
	defer srv.Close()

	c := signature.NewTSAClient(srv.URL)
	c.HTTPClient = srv.Client()
	_, _ = c.RequestTimestamp([]byte("any"))

	if receivedContentType != "application/timestamp-query" {
		t.Errorf("Content-Type = %q, want %q", receivedContentType, "application/timestamp-query")
	}
}

// TestWithTimestampOption verifies the WithTimestamp option flows into the config
// and triggers a TSA request during signing. Uses a mock TSA server.
func TestWithTimestampOption(t *testing.T) {
	key, cert, err := signature.GenerateTestCertificate()
	if err != nil {
		t.Fatalf("GenerateTestCertificate: %v", err)
	}
	signer, err := signature.NewLocalSigner(key, []*x509.Certificate{cert})
	if err != nil {
		t.Fatalf("NewLocalSigner: %v", err)
	}

	// Count how many times the TSA is called.
	tsaCalled := 0

	// Build a valid grantedWithMods response with a dummy token.
	dummyToken := []byte{0x30, 0x02, 0x05, 0x00} // minimal SEQUENCE (NULL)
	type pkiStatus struct{ Status int }
	type tsResp struct {
		Status         pkiStatus
		TimeStampToken asn1.RawValue `asn1:"optional"`
	}
	validResp := tsResp{
		Status:         pkiStatus{Status: 0},
		TimeStampToken: asn1.RawValue{FullBytes: dummyToken},
	}
	validDER, _ := asn1.Marshal(validResp)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tsaCalled++
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write(validDER)
	}))
	defer srv.Close()

	// We can't inject the custom HTTP client into WithTimestamp's TSA call directly
	// since the TSA client uses http.DefaultClient by default. We need to use the
	// test server URL and temporarily override http.DefaultClient.
	origClient := http.DefaultClient
	http.DefaultClient = srv.Client()
	defer func() { http.DefaultClient = origClient }()

	signed, err := signature.SignDocument(makeMinimalPDF(), signer,
		signature.WithTimestamp(srv.URL),
	)
	if err != nil {
		t.Fatalf("SignDocument with timestamp: %v", err)
	}

	if tsaCalled == 0 {
		t.Error("TSA server was never called — WithTimestamp option not applied")
	}

	// Verify the signature still passes (even with dummy TSA token).
	infos, err := signature.Verify(signed)
	if err != nil {
		t.Fatalf("Verify timestamped: %v", err)
	}
	if len(infos) == 0 || !infos[0].Valid {
		t.Fatalf("timestamped signature not valid: %v", infos[0].Err)
	}

	// HasTimestamp should be true when a TSA token was embedded.
	if !infos[0].HasTimestamp {
		t.Error("HasTimestamp should be true after WithTimestamp signing")
	}
}
