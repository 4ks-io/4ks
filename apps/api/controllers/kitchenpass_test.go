package controllers

import (
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"4ks/libs/go/models"
	"github.com/gin-gonic/gin"
)

func TestKitchenPassControllerGetSkillPage(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	newRequest := func(token string, service stubKitchenPassService) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/ai/"+token, nil)
		ctx.Params = gin.Params{{Key: "token", Value: token}}

		NewKitchenPassController(service).GetSkillPage(ctx)
		return rec
	}

	t.Run("renders markdown for active token", func(t *testing.T) {
		t.Parallel()

		token := "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789"
		rec := newRequest(token, stubKitchenPassService{
			validateTokenFn: func(_ context.Context, got string) (*models.PersonalAccessToken, error) {
				if got != token {
					t.Fatalf("unexpected token %q", got)
				}
				return &models.PersonalAccessToken{UserID: "user-1"}, nil
			},
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "text/markdown; charset=utf-8" {
			t.Fatalf("unexpected content type %q", got)
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Fatalf("unexpected cache control %q", got)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Authorization: Bearer "+token) {
			t.Fatalf("expected skill body to embed token, got %q", body)
		}
		if !strings.Contains(body, "GET https://api.4ks.io/api/recipes/search?q=chicken+soup") {
			t.Fatalf("expected search example in body, got %q", body)
		}
	})

	t.Run("invalid or unknown tokens return not found", func(t *testing.T) {
		t.Parallel()

		rec := newRequest("bad-token", stubKitchenPassService{})
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for malformed token, got %d", rec.Code)
		}

		token := "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789"
		rec = newRequest(token, stubKitchenPassService{
			validateTokenFn: func(context.Context, string) (*models.PersonalAccessToken, error) {
				return nil, kitchenpasssvc.ErrKitchenPassNotFound
			},
		})
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for missing token, got %d", rec.Code)
		}
	})
}
