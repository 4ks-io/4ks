package middleware

import (
	"4ks/libs/go/models"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestCasbinAuthorizationHelpers(t *testing.T) {
	t.Parallel()

	ok, err := Enforce("46a9ae64525d761613dd5cb865618526e5ee0c6c36cfd519716530e5f2694c75", "/users/*", "delete")
	if err != nil {
		t.Fatalf("Enforce returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected admin hash to be allowed to delete users")
	}

	ok, err = Enforce("anonymous", "/users/*", "delete")
	if err != nil {
		t.Fatalf("Enforce returned error: %v", err)
	}
	if ok {
		t.Fatal("expected anonymous subject to be denied")
	}

	ok, err = EnforceAuthor("user-1", models.UserSummary{ID: "user-1"})
	if err != nil {
		t.Fatalf("EnforceAuthor returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected subject to match author ID")
	}

	ok, err = EnforceContributor("user-2", []models.UserSummary{{ID: "user-1"}, {ID: "user-2"}})
	if err != nil {
		t.Fatalf("EnforceContributor returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected contributor list to include subject")
	}

	ids := getIds([]models.UserSummary{{ID: "a"}, {ID: "b"}})
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("unexpected id list: %#v", ids)
	}
}

func TestAuthorizeMiddleware(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("id", "anonymous")
		c.Next()
	})
	router.GET("/admin", Authorize("/users/*", "delete"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRateLimitHelpers(t *testing.T) {
	t.Parallel()

	if WindowRule("burst", 3, 2*time.Second) != (RateLimitRule{Name: "burst", Requests: 3, Window: 2 * time.Second}) {
		t.Fatal("expected WindowRule to preserve inputs")
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/api/example", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	ctx.Request = req

	if got := RateLimitByUserOrIP(ctx); got != "192.0.2.10" {
		t.Fatalf("expected remote IP fallback, got %q", got)
	}

	ctx.Set("id", "user-1")
	if got := RateLimitByUserOrIP(ctx); got != "user-1" {
		t.Fatalf("expected user ID key, got %q", got)
	}
}

func TestErrorAndLoggingMiddleware(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("error handler skips readiness logs", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/ready", nil)
		ctx.Error(errors.New("boom"))
		ErrorHandler(ctx)
	})

	t.Run("structured logger writes request metadata", func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)
		router := gin.New()
		router.Use(StructuredLogger(&logger))
		router.GET("/api/example", func(c *gin.Context) {
			c.String(http.StatusCreated, "ok")
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/example?q=1", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		var payload map[string]any
		if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal log: %v", err)
		}
		if payload["path"] != "/api/example?q=1" || payload["method"] != http.MethodGet {
			t.Fatalf("unexpected log payload: %+v", payload)
		}
	})
}

func TestCustomClaimsValidate(t *testing.T) {
	t.Parallel()

	if err := (CustomClaims{}).Validate(nil); err != nil {
		t.Fatalf("expected Validate to return nil, got %v", err)
	}
}
