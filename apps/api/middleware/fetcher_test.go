package middleware

import (
	"4ks/libs/go/fetchauth"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAuthorizeFetcher(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)

	newRouter := func() *gin.Engine {
		router := gin.New()
		router.POST("/api/_fetcher/recipes", authorizeFetcherWithSecret(secret, newFetcherNonceStore(), func() time.Time { return now }), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		return router
	}

	body, err := json.Marshal(map[string]string{"ok": "true"})
	if err != nil {
		t.Fatal(err)
	}

	makeRequest := func(headers map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/_fetcher/recipes", bytes.NewReader(body))
		req.Host = "api.4ks.io"
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		rec := httptest.NewRecorder()
		newRouter().ServeHTTP(rec, req)
		return rec
	}

	t.Run("missing headers", func(t *testing.T) {
		rec := makeRequest(nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("malformed timestamp", func(t *testing.T) {
		rec := makeRequest(map[string]string{
			fetchauth.HeaderTimestamp: "not-a-time",
			fetchauth.HeaderNonce:     "nonce",
			fetchauth.HeaderBodyHash:  fetchauth.HashBody(body),
			fetchauth.HeaderSignature: "deadbeef",
		})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("expired timestamp", func(t *testing.T) {
		headers := fetchauth.BuildHeaders(secret, http.MethodPost, "api.4ks.io", "/api/_fetcher/recipes", body, now.Add(-10*time.Minute), "nonce-expired")
		rec := makeRequest(map[string]string{
			fetchauth.HeaderTimestamp: headers.Timestamp,
			fetchauth.HeaderNonce:     headers.Nonce,
			fetchauth.HeaderBodyHash:  headers.BodyHash,
			fetchauth.HeaderSignature: headers.Signature,
		})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("valid headers", func(t *testing.T) {
		headers := fetchauth.BuildHeaders(secret, http.MethodPost, "api.4ks.io", "/api/_fetcher/recipes", body, now, "nonce-valid")
		rec := makeRequest(map[string]string{
			fetchauth.HeaderTimestamp: headers.Timestamp,
			fetchauth.HeaderNonce:     headers.Nonce,
			fetchauth.HeaderBodyHash:  headers.BodyHash,
			fetchauth.HeaderSignature: headers.Signature,
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("replayed nonce is rejected", func(t *testing.T) {
		router := gin.New()
		store := newFetcherNonceStore()
		router.POST("/api/_fetcher/recipes", authorizeFetcherWithSecret(secret, store, func() time.Time { return now }), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		headers := fetchauth.BuildHeaders(secret, http.MethodPost, "api.4ks.io", "/api/_fetcher/recipes", body, now, "nonce-replayed")
		req := func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/_fetcher/recipes", bytes.NewReader(body))
			request.Host = "api.4ks.io"
			request.Header.Set(fetchauth.HeaderTimestamp, headers.Timestamp)
			request.Header.Set(fetchauth.HeaderNonce, headers.Nonce)
			request.Header.Set(fetchauth.HeaderBodyHash, headers.BodyHash)
			request.Header.Set(fetchauth.HeaderSignature, headers.Signature)
			return request
		}

		first := httptest.NewRecorder()
		router.ServeHTTP(first, req())
		if first.Code != http.StatusOK {
			t.Fatalf("expected first request to succeed, got %d", first.Code)
		}

		second := httptest.NewRecorder()
		router.ServeHTTP(second, req())
		if second.Code != http.StatusUnauthorized {
			t.Fatalf("expected replay to be rejected, got %d", second.Code)
		}
	})

	t.Run("body hash mismatch is rejected", func(t *testing.T) {
		headers := fetchauth.BuildHeaders(secret, http.MethodPost, "api.4ks.io", "/api/_fetcher/recipes", body, now, "nonce-hash")
		rec := makeRequest(map[string]string{
			fetchauth.HeaderTimestamp: headers.Timestamp,
			fetchauth.HeaderNonce:     headers.Nonce,
			fetchauth.HeaderBodyHash:  strings.Repeat("a", len(headers.BodyHash)),
			fetchauth.HeaderSignature: headers.Signature,
		})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}
