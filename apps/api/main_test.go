package main

import (
	controllers "4ks/apps/api/controllers"
	"4ks/apps/api/utils"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
func (testUserController) TestUsername(c *gin.Context)          { c.Status(http.StatusOK) }
func (testUserController) RemoveUserEvent(c *gin.Context)       { c.Status(http.StatusOK) }

type testRecipeController struct{}

func (testRecipeController) BotCreateRecipe(c *gin.Context)        { c.Status(http.StatusOK) }
func (testRecipeController) FetcherBotCreateRecipe(c *gin.Context) { c.Status(http.StatusOK) }
func (testRecipeController) CreateRecipe(c *gin.Context)           { c.Status(http.StatusCreated) }
func (testRecipeController) DeleteRecipe(c *gin.Context)           { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipe(c *gin.Context)              { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipes(c *gin.Context)             { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipesByUsername(c *gin.Context)   { c.Status(http.StatusOK) }
func (testRecipeController) UpdateRecipe(c *gin.Context)           { c.Status(http.StatusOK) }
func (testRecipeController) ForkRecipe(c *gin.Context)             { c.Status(http.StatusOK) }
func (testRecipeController) StarRecipe(c *gin.Context)             { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipeRevisions(c *gin.Context)     { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipeRevision(c *gin.Context)      { c.Status(http.StatusOK) }
func (testRecipeController) CreateRecipeMedia(c *gin.Context)      { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipeMedia(c *gin.Context)         { c.Status(http.StatusOK) }
func (testRecipeController) GetAdminRecipeMedias(c *gin.Context)   { c.Status(http.StatusOK) }
func (testRecipeController) FetchRecipe(c *gin.Context)            { c.Status(http.StatusOK) }
func (testRecipeController) ForkRecipeRevision(c *gin.Context)     { c.Status(http.StatusOK) }
func (testRecipeController) GetRecipeForks(c *gin.Context)         { c.Status(http.StatusOK) }

type testSearchController struct{}

func (testSearchController) CreateSearchRecipeCollection(c *gin.Context) { c.Status(http.StatusOK) }

type testProber struct{}

func (testProber) Name() string                { return "ok" }
func (testProber) Probe(context.Context) error { return nil }

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

func TestGetAPIVersion(t *testing.T) {
	t.Run("defaults when version path is unset", func(t *testing.T) {
		if got := getAPIVersion(""); got != "0.0.0" {
			t.Fatalf("expected default version, got %q", got)
		}
	})

	t.Run("reads version file when configured", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "version")
		if err != nil {
			t.Fatalf("CreateTemp: %v", err)
		}
		if _, err := file.WriteString("1.2.3\n"); err != nil {
			t.Fatalf("WriteString: %v", err)
		}
		if err := file.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		if got := getAPIVersion(file.Name()); got != "1.2.3" {
			t.Fatalf("expected file-backed version, got %q", got)
		}
	})
}

func TestConfigureLogging(_ *testing.T) {
	configureLogging()
}

func TestReadWordsFromFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "words")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := file.WriteString("alpha\nbeta\n"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	words, err := ReadWordsFromFile(file.Name())
	if err != nil {
		t.Fatalf("ReadWordsFromFile: %v", err)
	}
	if len(words) != 2 || words[0] != "alpha" || words[1] != "beta" {
		t.Fatalf("unexpected words: %#v", words)
	}
}

func TestAppendRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	makeRouter := func(development bool) *gin.Engine {
		router := gin.New()
		cfg := utils.MinimalRuntimeConfig()
		cfg.System.Development = development
		AppendRoutes(cfg, router, newTestControllers())
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
		req := httptest.NewRequest(http.MethodPost, "/api/recipes/", nil)
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
