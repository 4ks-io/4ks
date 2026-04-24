package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// RateLimitKeyFunc extracts the identity used for a route's rate-limit bucket.
type RateLimitKeyFunc func(*gin.Context) string

// RateLimitRule defines one budget window, such as QPS, QPM, or QPH.
type RateLimitRule struct {
	Name     string
	Requests int
	Window   time.Duration
}

// RateLimitPolicy describes a route policy that can enforce several windows at once.
type RateLimitPolicy struct {
	Name    string
	Rules   []RateLimitRule
	KeyFunc RateLimitKeyFunc
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type limiterStore struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
}

// NewLimiterStore creates a reusable in-memory limiter store.
//
// This store is process-local. In Cloud Run, each instance keeps its own
// buckets, so limits are enforced per live instance rather than globally across
// all instances. When the service scales down to zero, all buckets disappear.
func NewLimiterStore() *limiterStore {
	return &limiterStore{
		entries: make(map[string]*limiterEntry),
	}
}

// get returns the limiter for a specific policy rule and key, creating it on first use.
func (s *limiterStore) get(key string, rule RateLimitRule) *rate.Limiter {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for existingKey, entry := range s.entries {
		// Expire cold buckets so the in-memory store does not grow without bound.
		if now.Sub(entry.lastSeen) > rule.Window*3 {
			delete(s.entries, existingKey)
		}
	}

	if entry, ok := s.entries[key]; ok {
		entry.lastSeen = now
		return entry.limiter
	}

	limiter := rate.NewLimiter(rate.Every(rule.Window/time.Duration(rule.Requests)), rule.Requests)
	s.entries[key] = &limiterEntry{
		limiter:  limiter,
		lastSeen: now,
	}
	return limiter
}

// NewRateLimitMiddleware creates a gin middleware for a route policy that can
// enforce multiple budgets, for example QPS + QPM + QPH, using the same key.
func NewRateLimitMiddleware(store *limiterStore, policy RateLimitPolicy) gin.HandlerFunc {
	if store == nil {
		panic("rate limiter store is required")
	}
	if len(policy.Rules) == 0 {
		panic("rate limit policy must contain at least one rule")
	}
	if policy.KeyFunc == nil {
		panic("rate limit key func is required")
	}
	for _, rule := range policy.Rules {
		if rule.Requests <= 0 {
			panic("rate limit requests must be positive")
		}
		if rule.Window <= 0 {
			panic("rate limit window must be positive")
		}
	}

	return func(c *gin.Context) {
		key := policy.KeyFunc(c)
		if key == "" {
			key = "anonymous"
		}

		for _, rule := range policy.Rules {
			limiter := store.get(fmt.Sprintf("%s:%s:%s", policy.Name, rule.Name, key), rule)
			if limiter.Allow() {
				continue
			}

			retryAfter := int(rule.Window.Seconds())
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			log.Warn().
				Str("policy", policy.Name).
				Str("rule", rule.Name).
				Str("key", key).
				Str("clientIP", c.ClientIP()).
				Str("path", c.Request.URL.Path).
				Int("requests", rule.Requests).
				Dur("window", rule.Window).
				Msg("rate limit exceeded")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "rate limit exceeded",
				"policy":  policy.Name,
				"rule":    rule.Name,
			})
			return
		}

		c.Next()
	}
}

// QPSRule is a convenience helper for a per-second budget.
func QPSRule(requests int) RateLimitRule {
	return RateLimitRule{Name: "qps", Requests: requests, Window: time.Second}
}

// QPMRule is a convenience helper for a per-minute budget.
func QPMRule(requests int) RateLimitRule {
	return RateLimitRule{Name: "qpm", Requests: requests, Window: time.Minute}
}

// QPHRule is a convenience helper for a per-hour budget.
func QPHRule(requests int) RateLimitRule {
	return RateLimitRule{Name: "qph", Requests: requests, Window: time.Hour}
}

// WindowRule is a convenience helper for nonstandard windows used by a policy.
func WindowRule(name string, requests int, window time.Duration) RateLimitRule {
	return RateLimitRule{Name: name, Requests: requests, Window: window}
}

// RateLimitByIP keys requests by the resolved client IP.
func RateLimitByIP(c *gin.Context) string {
	return c.ClientIP()
}

// RateLimitByUserOrIP keys requests by user ID and falls back to client IP.
func RateLimitByUserOrIP(c *gin.Context) string {
	if userID := c.GetString("id"); userID != "" {
		return userID
	}
	return c.ClientIP()
}
