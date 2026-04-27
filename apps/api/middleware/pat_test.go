package middleware

import (
	"4ks/apps/api/dtos"
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	"4ks/apps/api/utils"
	"4ks/libs/go/models"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubKitchenPassService struct {
	validateTokenFn func(context.Context, string) (*models.PersonalAccessToken, error)
	recordUsageFn   func(context.Context, string, string) error
}

func (s stubKitchenPassService) GetStatus(context.Context, string) (*dtos.KitchenPassResponse, error) {
	return nil, nil
}

func (s stubKitchenPassService) CreateOrRotate(context.Context, string) (*dtos.KitchenPassResponse, error) {
	return nil, nil
}

func (s stubKitchenPassService) Revoke(context.Context, string) error { return nil }

func (s stubKitchenPassService) ValidateToken(ctx context.Context, token string) (*models.PersonalAccessToken, error) {
	return s.validateTokenFn(ctx, token)
}

func (s stubKitchenPassService) RecordUsage(ctx context.Context, tokenDigest string, action string) error {
	if s.recordUsageFn != nil {
		return s.recordUsageFn(ctx, tokenDigest, action)
	}
	return nil
}

func TestRequirePAT(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequirePAT(stubKitchenPassService{
		validateTokenFn: func(_ context.Context, token string) (*models.PersonalAccessToken, error) {
			if token != "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789" {
				t.Fatalf("unexpected token %q", token)
			}
			return &models.PersonalAccessToken{UserID: "user-1", TokenDigest: "digest-1", TokenPreview: "4ks_pass_abc...6789"}, nil
		},
	}))
	router.GET("/api/example", func(c *gin.Context) {
		if got := c.GetString("id"); got != "user-1" {
			t.Fatalf("expected user ID in context, got %q", got)
		}
		if got := c.GetString("authType"); got != AuthTypePAT {
			t.Fatalf("expected authType pat, got %q", got)
		}
		if got := c.GetString("patDigest"); got != "digest-1" {
			t.Fatalf("expected patDigest in context, got %q", got)
		}
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/example", nil)
	req.Header.Set("Authorization", "Bearer 4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequirePATRecordsUsage(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	var gotDigest string
	var gotAction string

	router.Use(RequirePAT(stubKitchenPassService{
		validateTokenFn: func(_ context.Context, _ string) (*models.PersonalAccessToken, error) {
			return &models.PersonalAccessToken{
				UserID:       "user-1",
				TokenDigest:  "digest-1",
				TokenPreview: "4ks_pass_abc...6789",
			}, nil
		},
		recordUsageFn: func(_ context.Context, tokenDigest string, action string) error {
			gotDigest = tokenDigest
			gotAction = action
			return nil
		},
	}))
	router.GET("/api/recipes/search", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/recipes/search?q=soup", nil)
	req.Header.Set("Authorization", "Bearer 4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotDigest != "digest-1" {
		t.Fatalf("expected usage digest digest-1, got %q", gotDigest)
	}
	if gotAction != "searched recipes" {
		t.Fatalf("expected search action label, got %q", gotAction)
	}
}

func TestRequireJWTOrPAT(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := utils.MinimalRuntimeConfig()
	router.Use(RequireJWTOrPAT(cfg.Auth0, stubKitchenPassService{
		validateTokenFn: func(_ context.Context, _ string) (*models.PersonalAccessToken, error) {
			return nil, kitchenpasssvc.ErrKitchenPassNotFound
		},
	}))
	router.GET("/api/example", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/example", nil)
	req.Header.Set("Authorization", "Bearer 4ks_pass_bad")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
