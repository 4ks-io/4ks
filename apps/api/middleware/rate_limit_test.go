package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddlewareByIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := NewLimiterStore()
	router := gin.New()
	if err := router.SetTrustedProxies([]string{"10.0.0.0/8"}); err != nil {
		t.Fatalf("failed to set trusted proxies: %v", err)
	}
	router.Use(NewRateLimitMiddleware(store, RateLimitPolicy{
		Name: "test-ip",
		Rules: []RateLimitRule{
			QPSRule(10),
			QPMRule(1),
			QPHRule(100),
		},
		KeyFunc: RateLimitByIP,
	}))
	router.GET("/api/example", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	makeRequest := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/example", nil)
		req.RemoteAddr = "10.0.0.5:1234"
		req.Header.Set("X-Forwarded-For", "203.0.113.10")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	if rec := makeRequest(); rec.Code != http.StatusOK {
		t.Fatalf("expected first request to pass, got %d", rec.Code)
	}
	if rec := makeRequest(); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be limited, got %d", rec.Code)
	} else {
		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["rule"] != "qpm" {
			t.Fatalf("expected qpm rule to trip, got %q", body["rule"])
		}
	}
}
