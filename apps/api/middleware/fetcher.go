package middleware

import (
	"4ks/apps/api/utils"
	"4ks/libs/go/fetchauth"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	fetcherAuthClockSkew = 2 * time.Minute
	fetcherNonceTTL      = 5 * time.Minute
)

type fetcherNonceStore struct {
	mu     sync.Mutex
	values map[string]time.Time
}

func newFetcherNonceStore() *fetcherNonceStore {
	return &fetcherNonceStore{
		values: make(map[string]time.Time),
	}
}

func (s *fetcherNonceStore) Use(nonce string, now time.Time, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, expiry := range s.values {
		if !expiry.After(now) {
			delete(s.values, key)
		}
	}

	if _, exists := s.values[nonce]; exists {
		return false
	}

	s.values[nonce] = now.Add(ttl)
	return true
}

var replayProtection = newFetcherNonceStore()

// AuthorizeFetcher validates the request has been authorized to fetch.
func AuthorizeFetcher(cfg utils.FetcherConfig) gin.HandlerFunc {
	if cfg.SharedSecret == "" {
		log.Fatal().Msg("API_FETCHER_PSK required")
	}

	return authorizeFetcherWithSecret([]byte(cfg.SharedSecret), replayProtection, time.Now)
}

func authorizeFetcherWithSecret(secret []byte, nonces *fetcherNonceStore, nowFn func() time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		timestamp := c.GetHeader(fetchauth.HeaderTimestamp)
		nonce := c.GetHeader(fetchauth.HeaderNonce)
		bodyHash := c.GetHeader(fetchauth.HeaderBodyHash)
		signature := c.GetHeader(fetchauth.HeaderSignature)

		if timestamp == "" || nonce == "" || bodyHash == "" || signature == "" {
			abortFetcherAuth(c, errors.New("missing auth header"), "missing auth header")
			return
		}

		issuedAt, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			abortFetcherAuth(c, err, "malformed auth timestamp")
			return
		}

		now := nowFn().UTC()
		if issuedAt.Before(now.Add(-fetcherAuthClockSkew)) || issuedAt.After(now.Add(fetcherAuthClockSkew)) {
			abortFetcherAuth(c, errors.New("expired"), "expired auth timestamp")
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			abortFetcherAuth(c, err, "failed to read request body")
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		sum := sha256.Sum256(body)
		expectedBodyHash := hex.EncodeToString(sum[:])
		if !strings.EqualFold(expectedBodyHash, bodyHash) {
			abortFetcherAuth(c, errors.New("body hash mismatch"), "invalid body hash")
			return
		}

		if !nonces.Use(nonce, now, fetcherNonceTTL) {
			abortFetcherAuth(c, errors.New("nonce replay"), "replayed auth nonce")
			return
		}

		if err := fetchauth.Verify(
			secret,
			c.Request.Method,
			c.Request.Host,
			c.Request.URL.RequestURI(),
			bodyHash,
			timestamp,
			nonce,
			signature,
		); err != nil {
			abortFetcherAuth(c, err, "invalid auth signature")
			return
		}

		c.Next()
	}
}

func abortFetcherAuth(c *gin.Context, err error, msg string) {
	log.Error().Err(err).Msgf("authorization error: %s", msg)
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
}
