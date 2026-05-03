package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	controllers "4ks/apps/api/controllers"
	"4ks/apps/api/dtos"
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	"4ks/apps/api/utils"
	"4ks/libs/go/models"

	"github.com/gin-gonic/gin"
)

type testUserController struct{}

func (testUserController) CreateUser(c *gin.Context)            { c.Status(http.StatusCreated) }
func (testUserController) HeadAuthenticatedUser(c *gin.Context) { c.Status(http.StatusOK) }
func (testUserController) GetAuthenticatedUser(c *gin.Context)  { c.Status(http.StatusOK) }
func (testUserController) GetUser(c *gin.Context)               { c.Status(http.StatusOK) }
func (testUserController) GetUsers(c *gin.Context)              { c.Status(http.StatusOK) }
func (testUserController) DeleteUser(c *gin.Context)            { c.Status(http.StatusOK) }
func (testUserController) UpdateUser(c *gin.Context)            { c.Status(http.StatusOK) }
func (testUserController) GetKitchenPass(c *gin.Context)        { c.Status(http.StatusOK) }
func (testUserController) CreateKitchenPass(c *gin.Context)     { c.Status(http.StatusOK) }
func (testUserController) DeleteKitchenPass(c *gin.Context)     { c.Status(http.StatusOK) }
func (testUserController) TestUsername(c *gin.Context)          { c.Status(http.StatusOK) }
func (testUserController) RemoveUserEvent(c *gin.Context)       { c.Status(http.StatusOK) }

type testRecipeController struct{}

func (testRecipeController) BotCreateRecipe(c *gin.Context)        { c.Status(http.StatusOK) }
func (testRecipeController) FetcherBotCreateRecipe(c *gin.Context) { c.Status(http.StatusOK) }
func (testRecipeController) CreateRecipe(c *gin.Context)           { c.Status(http.StatusCreated) }
func (testRecipeController) DeleteRecipe(c *gin.Context)           { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipe(c *gin.Context)              { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipes(c *gin.Context)             { c.Status(http.StatusOK) }
func (testRecipeController) SearchRecipes(c *gin.Context)          { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipesByUsername(c *gin.Context)   { c.Status(http.StatusOK) }
func (testRecipeController) UpdateRecipe(c *gin.Context)           { c.Status(http.StatusOK) }
func (testRecipeController) ForkRecipe(c *gin.Context)             { c.Status(http.StatusOK) }
func (testRecipeController) StarRecipe(c *gin.Context)             { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipeRevisions(c *gin.Context)     { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipeRevision(c *gin.Context)      { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipeForks(c *gin.Context)         { c.Status(http.StatusOK) }
func (testRecipeController) CreateRecipeMedia(c *gin.Context)      { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipeMedia(c *gin.Context)         { c.Status(http.StatusOK) }
func (testRecipeController) GetAdminRecipeMedias(c *gin.Context)   { c.Status(http.StatusOK) }
func (testRecipeController) FetchRecipe(c *gin.Context)            { c.Status(http.StatusOK) }
func (testRecipeController) ForkRecipeRevision(c *gin.Context)     { c.Status(http.StatusOK) }

type testSearchController struct{}

func (testSearchController) CreateSearchRecipeCollection(c *gin.Context) { c.Status(http.StatusOK) }

type testProber struct{}

func (testProber) Name() string                { return "ok" }
func (testProber) Probe(context.Context) error { return nil }

type stubKitchenPassService struct{}

func (stubKitchenPassService) GetStatus(context.Context, string) (*dtos.KitchenPassResponse, error) {
	return &dtos.KitchenPassResponse{Enabled: false}, nil
}

func (stubKitchenPassService) CreateOrRotate(context.Context, string) (*dtos.KitchenPassResponse, error) {
	return &dtos.KitchenPassResponse{Enabled: true}, nil
}

func (stubKitchenPassService) Revoke(context.Context, string) error { return nil }

func (stubKitchenPassService) ValidateToken(_ context.Context, token string) (*models.PersonalAccessToken, error) {
	if token == "4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789" {
		return &models.PersonalAccessToken{UserID: "user-1", TokenDigest: "digest-1", TokenPreview: "4ks_pass_abc...6789"}, nil
	}
	return nil, kitchenpasssvc.ErrKitchenPassNotFound
}

func (stubKitchenPassService) RecordUsage(context.Context, string, string) error { return nil }

func newTestControllers() *Controllers {
	return &Controllers{
		User:   testUserController{},
		Recipe: testRecipeController{},
		Search: testSearchController{},
		System: controllers.NewSystemController("test-version", controllers.SystemControllerDeps{
			DB:        testProber{},
			Search:    testProber{},
			Messaging: testProber{},
			Storage:   testProber{},
		}),
	}
}

func TestAppendRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRouter := func(development bool) *gin.Engine {
		router := gin.New()
		cfg := utils.MinimalRuntimeConfig()
		cfg.System.Development = development
		AppendRoutes(cfg, router, newTestControllers(), stubKitchenPassService{})
		return router
	}

	t.Run("readiness route is always exposed", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
		makeRouter(false).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("healthcheck is development only", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/healthcheck", nil)
		makeRouter(false).ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 outside development, got %d", rec.Code)
		}

		rec = httptest.NewRecorder()
		makeRouter(true).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 in development, got %d", rec.Code)
		}
	})

	t.Run("authenticated recipe writes are protected by jwt", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/recipes", nil)
		makeRouter(false).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("kitchen pass can access approved recipe routes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/recipes", nil)
		req.Header.Set("Authorization", "Bearer 4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789")
		makeRouter(false).ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("kitchen pass can access recipe search route", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/recipes/search?q=soup", nil)
		req.Header.Set("Authorization", "Bearer 4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789")
		makeRouter(false).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("kitchen pass is rejected on jwt-only routes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/user/", nil)
		req.Header.Set("Authorization", "Bearer 4ks_pass_abcdefghijklmnopqrstuvwxyz0123456789")
		makeRouter(false).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("fetcher route exists and rejects missing auth headers", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/_fetcher/recipes", nil)
		req.Host = "api.4ks.io"
		makeRouter(false).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}
