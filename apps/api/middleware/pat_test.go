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

func TestRequirePAT(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequirePAT(stubKitchenPassService{
		validateTokenFn: func(_ context.Context, token string) (*models.PersonalAccessToken, error) {
			if token != "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789" {
				t.Fatalf("unexpected token %q", token)
			}
			return &models.PersonalAccessToken{UserID: "user-1"}, nil
		},
	}))
	router.GET("/api/example", func(c *gin.Context) {
		if got := c.GetString("id"); got != "user-1" {
			t.Fatalf("expected user ID in context, got %q", got)
		}
		if got := c.GetString("authType"); got != AuthTypePAT {
			t.Fatalf("expected authType pat, got %q", got)
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

func TestRequireJWTOrPAT(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := utils.MinimalRuntimeConfig()
	router.Use(RequireJWTOrPAT(cfg.Auth0, stubKitchenPassService{
		validateTokenFn: func(_ context.Context, token string) (*models.PersonalAccessToken, error) {
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
