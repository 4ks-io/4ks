package middleware

import (
	"context"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppendCustomClaims(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	claims := &validator.ValidatedClaims{
		RegisteredClaims: validator.RegisteredClaims{
			Subject: "auth0|abc123",
		},
		CustomClaims: &CustomClaims{
			ID:    "user-1",
			Email: "user@example.com",
		},
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/api/example", nil)
	req = req.WithContext(context.WithValue(req.Context(), jwtmiddleware.ContextKey{}, claims))
	ctx.Request = req

	called := false
	middleware := AppendCustomClaims()
	middleware(ctx)
	called = true

	if !called {
		t.Fatal("expected middleware to run")
	}
	if got := ctx.GetString("authID"); got != "auth0|abc123" {
		t.Fatalf("expected authID to be set, got %q", got)
	}
	if got := ctx.GetString("id"); got != "user-1" {
		t.Fatalf("expected id to be set, got %q", got)
	}
	if got := ctx.GetString("email"); got != "user@example.com" {
		t.Fatalf("expected email to be set, got %q", got)
	}
}

func TestExtractCustomClaimsFromClaims(t *testing.T) {
	t.Parallel()

	claims := &validator.ValidatedClaims{
		CustomClaims: &CustomClaims{
			ID:       "user-1",
			Email:    "user@example.com",
			Timezone: "Europe/London",
		},
	}

	custom := ExtractCustomClaimsFromClaims(claims)
	if custom.ID != "user-1" || custom.Email != "user@example.com" || custom.Timezone != "Europe/London" {
		t.Fatalf("unexpected custom claims: %+v", custom)
	}
}
