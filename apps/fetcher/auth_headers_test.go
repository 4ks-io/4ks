package fetcher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewNonce(t *testing.T) {
	t.Parallel()

	nonce, err := newNonce()
	if err != nil {
		t.Fatalf("newNonce returned error: %v", err)
	}
	if len(nonce) != 32 {
		t.Fatalf("expected 32-char nonce, got %q", nonce)
	}
	if nonce2, err := newNonce(); err != nil || nonce2 == nonce {
		t.Fatalf("expected distinct nonce, got %q %v", nonce2, err)
	}
}

func TestHashBodyAndSign(t *testing.T) {
	t.Parallel()

	body := []byte(`{"ok":true}`)
	gotHash := hashBody(body)
	sum := sha256.Sum256(body)
	wantHash := hex.EncodeToString(sum[:])
	if gotHash != wantHash {
		t.Fatalf("hashBody mismatch: got %q want %q", gotHash, wantHash)
	}

	secret := []byte("01234567890123456789012345678901")
	payload := strings.Join([]string{
		"POST",
		"api.4ks.io",
		"/api/_fetcher/recipes",
		strings.ToLower(gotHash),
		"2026-04-25T12:00:00Z",
		"nonce-1",
	}, "\n")
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(payload))
	wantSig := hex.EncodeToString(mac.Sum(nil))

	gotSig := sign(secret, "post", "API.4KS.IO", "/api/_fetcher/recipes", gotHash, "2026-04-25T12:00:00Z", "nonce-1")
	if gotSig != wantSig {
		t.Fatalf("sign mismatch: got %q want %q", gotSig, wantSig)
	}
}

func TestBuildAndApplySignatureHeaders(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	body := []byte(`{"ok":true}`)
	headers := buildSignatureHeaders([]byte("01234567890123456789012345678901"), http.MethodPost, "api.4ks.io", "/api/_fetcher/recipes", body, now, "nonce-1")

	if headers.Timestamp != "2026-04-25T12:00:00Z" || headers.Nonce != "nonce-1" || headers.BodyHash == "" || headers.Signature == "" {
		t.Fatalf("unexpected headers: %+v", headers)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.4ks.io/api/_fetcher/recipes", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	applySignatureHeaders(req, headers)

	if req.Header.Get(headerTimestamp) != headers.Timestamp ||
		req.Header.Get(headerNonce) != headers.Nonce ||
		req.Header.Get(headerBodyHash) != headers.BodyHash ||
		req.Header.Get(headerSignature) != headers.Signature {
		t.Fatalf("expected request headers to match signature headers: %+v", req.Header)
	}
}
