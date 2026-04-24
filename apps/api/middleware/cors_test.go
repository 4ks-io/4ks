package middleware

import (
	"4ks/apps/api/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestCorsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := utils.CORSConfig{
		AllowedOrigins:   []string{"https://www.4ks.io", "https://dev.4ks.io"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}

	newRouter := func() *gin.Engine {
		router := gin.New()
		router.Use(CorsMiddleware(cfg))
		router.GET("/api/example", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		router.OPTIONS("/api/example", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		return router
	}

	t.Run("allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/example", nil)
		req.Header.Set("Origin", "https://www.4ks.io")
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://www.4ks.io" {
			t.Fatalf("expected allowed origin header, got %q", got)
		}
	})

	t.Run("denied origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/example", nil)
		req.Header.Set("Origin", "https://evil.example")
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("expected no allow-origin header, got %q", got)
		}
	})

	t.Run("preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/example", nil)
		req.Header.Set("Origin", "https://dev.4ks.io")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
			t.Fatal("expected allow-methods header to be set")
		}
	})

	t.Run("credentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/example", nil)
		req.Header.Set("Origin", "https://www.4ks.io")
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Fatalf("expected credentials header, got %q", got)
		}
	})
}
