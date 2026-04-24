package fetchauth

import (
	"testing"
	"time"
)

func TestBuildAndVerifyHeaders(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	body := []byte(`{"ok":true}`)
	now := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)

	headers := BuildHeaders(secret, "POST", "api.4ks.io", "/api/_fetcher/recipes", body, now, "nonce")

	if headers.BodyHash != HashBody(body) {
		t.Fatalf("unexpected body hash: %s", headers.BodyHash)
	}

	if err := Verify(
		secret,
		"POST",
		"api.4ks.io",
		"/api/_fetcher/recipes",
		headers.BodyHash,
		headers.Timestamp,
		headers.Nonce,
		headers.Signature,
	); err != nil {
		t.Fatalf("expected signature to verify: %v", err)
	}
}

func TestVerifyRejectsTampering(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	body := []byte(`{"ok":true}`)
	now := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)

	headers := BuildHeaders(secret, "POST", "api.4ks.io", "/api/_fetcher/recipes", body, now, "nonce")

	if err := Verify(
		secret,
		"POST",
		"api.4ks.io",
		"/api/_fetcher/recipes",
		HashBody([]byte(`{"ok":false}`)),
		headers.Timestamp,
		headers.Nonce,
		headers.Signature,
	); err == nil {
		t.Fatal("expected tampered body hash to fail verification")
	}
}
