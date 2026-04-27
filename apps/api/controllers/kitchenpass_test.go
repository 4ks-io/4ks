package controllers

import (
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	"context"
	"encoding/json"
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

	newRequest := func(token string, accept string, service stubKitchenPassService) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		req := httptest.NewRequest(http.MethodGet, "/ai/"+token, nil)
		if accept != "" {
			req.Header.Set("Accept", accept)
		}
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "token", Value: token}}

		NewKitchenPassController(service).GetSkillPage(ctx)
		return rec
	}

	t.Run("renders markdown for active token", func(t *testing.T) {
		t.Parallel()

		token := "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789"
		rec := newRequest(token, "", stubKitchenPassService{
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
		if !strings.Contains(body, "If a recipe title matches exactly, present the existing recipe before creating a new one.") {
			t.Fatalf("expected duplicate detection guidance, got %q", body)
		}
		if !strings.Contains(body, "`429`: transient. Retry with exponential backoff starting at 2 seconds.") {
			t.Fatalf("expected retry guidance in body, got %q", body)
		}
	})

	t.Run("renders json for application json accept header", func(t *testing.T) {
		t.Parallel()

		token := "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789"
		rec := newRequest(token, "application/json", stubKitchenPassService{
			validateTokenFn: func(_ context.Context, got string) (*models.PersonalAccessToken, error) {
				return &models.PersonalAccessToken{UserID: "user-1"}, nil
			},
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Fatalf("unexpected content type %q", got)
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if payload["apiBaseUrl"] != "https://api.4ks.io" {
			t.Fatalf("unexpected api base url: %+v", payload)
		}
		rules, ok := payload["decisionRules"].([]any)
		if !ok || len(rules) < 2 {
			t.Fatalf("expected decision rules, got %+v", payload["decisionRules"])
		}
		foundExactMatchRule := false
		for _, rule := range rules {
			if rule == "If a recipe title matches exactly, present the existing recipe before creating a new one." {
				foundExactMatchRule = true
			}
		}
		if !foundExactMatchRule {
			t.Fatalf("expected exact-match duplicate rule, got %+v", rules)
		}
		errors, ok := payload["errorGuidance"].([]any)
		if !ok || len(errors) == 0 {
			t.Fatalf("expected error guidance, got %+v", payload["errorGuidance"])
		}
	})

	t.Run("renders openapi json for openapi accept header", func(t *testing.T) {
		t.Parallel()

		token := "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789"
		rec := newRequest(token, "application/openapi+json", stubKitchenPassService{
			validateTokenFn: func(_ context.Context, got string) (*models.PersonalAccessToken, error) {
				return &models.PersonalAccessToken{UserID: "user-1"}, nil
			},
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/openapi+json; charset=utf-8" {
			t.Fatalf("unexpected content type %q", got)
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if payload["openapi"] != "3.1.0" {
			t.Fatalf("unexpected openapi version: %+v", payload)
		}
		paths, ok := payload["paths"].(map[string]any)
		if !ok {
			t.Fatalf("expected paths object, got %+v", payload["paths"])
		}
		if _, ok := paths["/api/recipes/search"]; !ok {
			t.Fatalf("expected search path, got %+v", paths)
		}
		guidance, ok := payload["x-4ks-guidance"].(map[string]any)
		if !ok {
			t.Fatalf("expected x-4ks-guidance, got %+v", payload)
		}
		if _, ok := guidance["errorGuidance"]; !ok {
			t.Fatalf("expected error guidance in openapi extension, got %+v", guidance)
		}
	})

	t.Run("invalid or unknown tokens return not found", func(t *testing.T) {
		t.Parallel()

		rec := newRequest("bad-token", "", stubKitchenPassService{})
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for malformed token, got %d", rec.Code)
		}

		token := "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789"
		rec = newRequest(token, "", stubKitchenPassService{
			validateTokenFn: func(context.Context, string) (*models.PersonalAccessToken, error) {
				return nil, kitchenpasssvc.ErrKitchenPassNotFound
			},
		})
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for missing token, got %d", rec.Code)
		}
	})
}
