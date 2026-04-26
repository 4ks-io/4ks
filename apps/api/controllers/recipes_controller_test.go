package controllers

import (
	"4ks/apps/api/dtos"
	recipesvc "4ks/apps/api/services/recipe"
	usersvc "4ks/apps/api/services/user"
	"4ks/apps/api/utils"
	models "4ks/libs/go/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func performRecipeControllerRequest(t *testing.T, handler gin.HandlerFunc, method string, target string, body []byte, setup func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	if setup != nil {
		setup(ctx)
	}

	handler(ctx)
	return rec
}

func recipeFixture() *models.Recipe {
	return &models.Recipe{
		ID: "recipe-1",
		Author: models.UserSummary{
			ID:          "user-1",
			Username:    "chef-user",
			DisplayName: "Chef User",
		},
		CurrentRevision: models.RecipeRevision{
			Name: "Soup",
			Ingredients: []models.Ingredient{
				{Name: "salt"},
				{Name: "pepper"},
			},
			Banner: []models.RecipeMediaVariant{
				{Alias: "md", URL: "https://cdn.example/banner.jpg", Filename: "banner.jpg"},
			},
		},
	}
}

func TestRecipeControllerCreateRecipe(t *testing.T) {
	t.Parallel()

	t.Run("builds author and fallback banner before create", func(t *testing.T) {
		t.Parallel()

		var created *dtos.CreateRecipe
		controller := NewRecipeController(
			stubUserService{
				getUserByIDFn: func(context.Context, string) (*models.User, error) {
					return &models.User{ID: "user-1", Username: "chef-user", DisplayName: "Chef User"}, nil
				},
			},
			stubRecipeService{
				createMockBannerFn: func(filename string, url string) []models.RecipeMediaVariant {
					if filename != "fallback.jpg" || url != "https://cdn.example/fallback.jpg" {
						t.Fatalf("unexpected fallback banner inputs: %q %q", filename, url)
					}
					return []models.RecipeMediaVariant{{Alias: "md", URL: url, Filename: filename}}
				},
				createRecipeFn: func(_ context.Context, payload *dtos.CreateRecipe) (*models.Recipe, error) {
					created = payload
					return recipeFixture(), nil
				},
			},
			stubSearchService{
				upsertSearchRecipeDocumentFn: func(recipe *models.Recipe) error {
					if recipe.ID != "recipe-1" {
						t.Fatalf("unexpected recipe passed to search index: %+v", recipe)
					}
					return nil
				},
			},
			stubStaticService{
				getRandomFallbackImageFn: func(context.Context) (string, error) { return "fallback.jpg", nil },
				getRandomFallbackImageURLFn: func(_ string) string {
					return "https://cdn.example/fallback.jpg"
				},
			},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.CreateRecipe, http.MethodPost, "/api/recipes", []byte(`{"name":"Soup"}`), func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if created == nil {
			t.Fatal("expected recipe to be passed to service")
		}
		if created.Author.Username != "chef-user" || len(created.Banner) != 1 {
			t.Fatalf("expected author and banner to be populated, got %+v", created)
		}
	})

	t.Run("search index errors fail the request", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{
				getUserByIDFn: func(context.Context, string) (*models.User, error) {
					return &models.User{ID: "user-1", Username: "chef-user", DisplayName: "Chef User"}, nil
				},
			},
			stubRecipeService{
				createMockBannerFn: func(string, string) []models.RecipeMediaVariant { return nil },
				createRecipeFn:     func(context.Context, *dtos.CreateRecipe) (*models.Recipe, error) { return recipeFixture(), nil },
			},
			stubSearchService{
				upsertSearchRecipeDocumentFn: func(*models.Recipe) error { return errors.New("search down") },
			},
			stubStaticService{
				getRandomFallbackImageFn:    func(context.Context) (string, error) { return "fallback.jpg", nil },
				getRandomFallbackImageURLFn: func(string) string { return "https://cdn.example/fallback.jpg" },
			},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.CreateRecipe, http.MethodPost, "/api/recipes", []byte(`{"name":"Soup"}`), func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestRecipeControllerCreateRecipeMedia(t *testing.T) {
	t.Parallel()

	t.Run("invalid extension returns internal error", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(stubUserService{}, stubRecipeService{}, stubSearchService{}, stubStaticService{}, stubFetcherService{})
		rec := performRecipeControllerRequest(t, controller.CreateRecipeMedia, http.MethodPost, "/api/recipes/recipe-1/media", []byte(`{"filename":"notes.txt"}`), func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("creates signed url and media metadata", func(t *testing.T) {
		t.Parallel()

		var signedURLCalled bool
		var mediaCalled bool
		controller := NewRecipeController(
			stubUserService{},
			stubRecipeService{
				createRecipeMediaSignedURLFn: func(_ context.Context, mp *utils.MediaProps, _ *sync.WaitGroup) (string, error) {
					signedURLCalled = true
					if mp.ContentType != "image/png" || mp.Extension != ".png" || mp.Basename == "" {
						t.Fatalf("unexpected media props for signed URL: %+v", mp)
					}
					return "https://signed.example/upload", nil
				},
				createRecipeMediaFn: func(_ context.Context, mp *utils.MediaProps, recipeID string, userID string, _ *sync.WaitGroup) (*models.RecipeMedia, error) {
					mediaCalled = true
					if recipeID != "recipe-1" || userID != "user-1" {
						t.Fatalf("unexpected ids: %q %q", recipeID, userID)
					}
					return &models.RecipeMedia{ID: "media-1", RecipeID: recipeID, OwnerID: userID, ContentType: mp.ContentType}, nil
				},
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.CreateRecipeMedia, http.MethodPost, "/api/recipes/recipe-1/media", []byte(`{"filename":"banner.png"}`), func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !signedURLCalled || !mediaCalled {
			t.Fatal("expected both signed URL and media creation to run")
		}

		var payload map[string]json.RawMessage
		if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if _, ok := payload["signedURL"]; !ok {
			t.Fatalf("expected signedURL field, got %v", payload)
		}
	})
}

func TestRecipeControllerGetRecipesByUsername(t *testing.T) {
	t.Parallel()

	t.Run("bot username bypasses user lookup", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{
				getUserByUsernameFn: func(context.Context, string) (*models.User, error) {
					t.Fatal("bot requests should not look up users")
					return nil, nil
				},
			},
			stubRecipeService{
				getRecipesByUserIDFn: func(_ context.Context, userID string, limit int) ([]*models.Recipe, error) {
					if userID != "bot" || limit != 40 {
						t.Fatalf("unexpected getRecipesByUserID inputs: %q %d", userID, limit)
					}
					return []*models.Recipe{recipeFixture()}, nil
				},
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.GetRecipesByUsername, http.MethodGet, "/api/recipes/author/4ks-bot", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "username", Value: "4ks-bot"}}
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("missing author returns not found", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{
				getUserByUsernameFn: func(context.Context, string) (*models.User, error) {
					return nil, usersvc.ErrUserNotFound
				},
			},
			stubRecipeService{},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.GetRecipesByUsername, http.MethodGet, "/api/recipes/author/missing", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "username", Value: "missing"}}
		})

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}

func TestRecipeControllerStarRecipe(t *testing.T) {
	t.Parallel()

	t.Run("already starred returns bad request payload", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{
				getUserByIDFn: func(context.Context, string) (*models.User, error) {
					return &models.User{ID: "user-1", Username: "chef-user", DisplayName: "Chef User"}, nil
				},
			},
			stubRecipeService{
				starRecipeByIDFn: func(context.Context, string, models.UserSummary) (bool, error) {
					return false, recipesvc.ErrRecipeAlreadyStarred
				},
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.StarRecipe, http.MethodPost, "/api/recipes/recipe-1/star", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}

		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body["message"] != "Recipe is already starred" {
			t.Fatalf("unexpected body: %+v", body)
		}
	})
}

func TestRecipeControllerFetchRecipe(t *testing.T) {
	t.Parallel()

	t.Run("rejects invalid fetch url before side effects", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{
				getUserByIDFn: func(context.Context, string) (*models.User, error) {
					return &models.User{ID: "user-1"}, nil
				},
				createUserEventByUserIDFn: func(context.Context, string, *dtos.CreateUserEvent) (*models.UserEvent, error) {
					t.Fatal("should not create an event for invalid URL")
					return nil, nil
				},
			},
			stubRecipeService{},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.FetchRecipe, http.MethodPost, "/api/recipes/fetch", []byte(`{"url":"http://localhost/test"}`), func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("creates a processing event and dispatches fetch", func(t *testing.T) {
		t.Parallel()

		eventID := uuid.New()
		var sent *models.FetcherRequest
		controller := NewRecipeController(
			stubUserService{
				getUserByIDFn: func(context.Context, string) (*models.User, error) {
					return &models.User{ID: "user-1"}, nil
				},
				createUserEventByUserIDFn: func(_ context.Context, userID string, payload *dtos.CreateUserEvent) (*models.UserEvent, error) {
					if userID != "user-1" || payload.Type != models.UserEventTypeFetchRecipe || payload.Status != models.UserEventProcessing {
						t.Fatalf("unexpected event payload: %q %+v", userID, payload)
					}
					data, ok := payload.Data.(models.FetcherEventData)
					if !ok || data.URL == "" {
						t.Fatalf("expected fetcher event data, got %#v", payload.Data)
					}
					return &models.UserEvent{ID: eventID, Type: payload.Type, Status: payload.Status, Data: payload.Data}, nil
				},
			},
			stubRecipeService{},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{
				sendFn: func(_ context.Context, req *models.FetcherRequest) (string, error) {
					sent = req
					return "msg-1", nil
				},
			},
		)

		rec := performRecipeControllerRequest(t, controller.FetchRecipe, http.MethodPost, "/api/recipes/fetch", []byte(`{"url":"https://example.com/recipe"}`), func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if sent == nil || sent.UserID != "user-1" || sent.UserEventID != eventID || sent.URL != "https://example.com/recipe" {
			t.Fatalf("unexpected fetch dispatch payload: %+v", sent)
		}
	})
}

func TestSearchControllerCreateCollection(t *testing.T) {
	t.Parallel()

	t.Run("success returns ok", func(t *testing.T) {
		t.Parallel()

		controller := NewSearchController(stubSearchService{
			createSearchRecipeCollectionFn: func() error { return nil },
		})
		rec := performRecipeControllerRequest(t, controller.CreateSearchRecipeCollection, http.MethodPost, "/api/_admin/init-search-collections", nil, nil)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("search failures return internal error", func(t *testing.T) {
		t.Parallel()

		controller := NewSearchController(stubSearchService{
			createSearchRecipeCollectionFn: func() error { return errors.New("typesense down") },
		})
		rec := performRecipeControllerRequest(t, controller.CreateSearchRecipeCollection, http.MethodPost, "/api/_admin/init-search-collections", nil, nil)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestRecipeControllerReadHandlers(t *testing.T) {
	t.Parallel()

	t.Run("get recipe returns wrapped payload", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{},
			stubRecipeService{
				getRecipeByIDFn: func(context.Context, string) (*models.Recipe, error) { return recipeFixture(), nil },
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.GetRecipe, http.MethodGet, "/api/recipes/recipe-1", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var payload dtos.GetRecipeResponse
		if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if payload.Data == nil || payload.Data.ID != "recipe-1" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
	})

	t.Run("get recipe not found maps to 404", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{},
			stubRecipeService{
				getRecipeByIDFn: func(context.Context, string) (*models.Recipe, error) { return nil, recipesvc.ErrRecipeNotFound },
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.GetRecipe, http.MethodGet, "/api/recipes/recipe-1", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
		})

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("get recipe revisions not found maps to 404", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{},
			stubRecipeService{
				getRecipeRevisionsFn: func(context.Context, string) ([]*models.RecipeRevision, error) {
					return nil, recipesvc.ErrRecipeNotFound
				},
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.GetRecipeRevisions, http.MethodGet, "/api/recipes/recipe-1/revisions", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
		})

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("get recipe revision not found maps to 404", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{},
			stubRecipeService{
				getRecipeRevisionByIDFn: func(context.Context, string) (*models.RecipeRevision, error) {
					return nil, recipesvc.ErrRecipeRevisionNotFound
				},
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.GetRecipeRevision, http.MethodGet, "/api/recipes/revisions/rev-1", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "revisionID", Value: "rev-1"}}
		})

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("admin recipe medias use admin service path", func(t *testing.T) {
		t.Parallel()

		calledAdmin := false
		controller := NewRecipeController(
			stubUserService{},
			stubRecipeService{
				getRecipeMediaFn: func(context.Context, string) ([]*models.RecipeMedia, error) {
					t.Fatal("expected admin endpoint to avoid public media accessor")
					return nil, nil
				},
				getAdminRecipeMediasFn: func(_ context.Context, recipeID string) ([]*models.RecipeMedia, error) {
					calledAdmin = true
					return []*models.RecipeMedia{{ID: "media-1", RecipeID: recipeID}}, nil
				},
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.GetAdminRecipeMedias, http.MethodGet, "/api/_admin/recipes/recipe-1/media", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !calledAdmin {
			t.Fatal("expected admin media service to be called")
		}
	})
}

func TestRecipeControllerWriteHandlers(t *testing.T) {
	t.Parallel()

	t.Run("update recipe unauthorized maps to 401", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{
				getUserByIDFn: func(context.Context, string) (*models.User, error) {
					return &models.User{ID: "user-1", Username: "chef-user", DisplayName: "Chef User"}, nil
				},
			},
			stubRecipeService{
				updateRecipeByIDFn: func(context.Context, string, *dtos.UpdateRecipe) (*models.Recipe, error) {
					return nil, recipesvc.ErrUnauthorized
				},
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.UpdateRecipe, http.MethodPatch, "/api/recipes/recipe-1", []byte(`{"name":"Soup"}`), func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("delete recipe unauthorized maps to 401", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{},
			stubRecipeService{
				deleteRecipeFn: func(context.Context, string, string) error { return recipesvc.ErrUnauthorized },
			},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.DeleteRecipe, http.MethodDelete, "/api/recipes/recipe-1", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("fork recipe search update failure bubbles up", func(t *testing.T) {
		t.Parallel()

		controller := NewRecipeController(
			stubUserService{
				getUserByIDFn: func(context.Context, string) (*models.User, error) {
					return &models.User{ID: "user-1", Username: "chef-user", DisplayName: "Chef User"}, nil
				},
			},
			stubRecipeService{
				forkRecipeByIDFn: func(context.Context, string, models.UserSummary) (*models.Recipe, error) {
					return recipeFixture(), nil
				},
			},
			stubSearchService{
				upsertSearchRecipeDocumentFn: func(*models.Recipe) error { return errors.New("search unavailable") },
			},
			stubStaticService{},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.ForkRecipe, http.MethodPost, "/api/recipes/recipe-1/fork", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "recipe-1"}}
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestBotControllers(t *testing.T) {
	t.Parallel()

	t.Run("bot create recipe uses bot author", func(t *testing.T) {
		t.Parallel()

		var created *dtos.CreateRecipe
		controller := NewRecipeController(
			stubUserService{},
			stubRecipeService{
				createMockBannerFn: func(string, string) []models.RecipeMediaVariant { return []models.RecipeMediaVariant{{Alias: "md"}} },
				createRecipeFn: func(_ context.Context, payload *dtos.CreateRecipe) (*models.Recipe, error) {
					created = payload
					return recipeFixture(), nil
				},
			},
			stubSearchService{},
			stubStaticService{
				getRandomFallbackImageFn:    func(context.Context) (string, error) { return "fallback.jpg", nil },
				getRandomFallbackImageURLFn: func(string) string { return "https://cdn.example/fallback.jpg" },
			},
			stubFetcherService{},
		)

		rec := performRecipeControllerRequest(t, controller.BotCreateRecipe, http.MethodPost, "/api/_admin/recipes", []byte(`{"name":"Soup"}`), nil)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if created == nil || created.Author.ID != "bot" || created.Author.Username != "4ks-bot" {
			t.Fatalf("expected bot author, got %+v", created)
		}
	})

	t.Run("fetcher bot rejects empty recipe and marks event errored", func(t *testing.T) {
		t.Parallel()

		var updatePayload *dtos.UpdateUserEvent
		controller := NewRecipeController(
			stubUserService{
				updateUserEventByUserIDEventFn: func(_ context.Context, _ string, payload *dtos.UpdateUserEvent) (*models.UserEvent, error) {
					updatePayload = payload
					return &models.UserEvent{ID: payload.ID, Status: payload.Status, Error: payload.Error}, nil
				},
			},
			stubRecipeService{},
			stubSearchService{},
			stubStaticService{},
			stubFetcherService{},
		)

		eventID := uuid.New()
		body := `{"userId":"user-1","userEventId":"` + eventID.String() + `","recipe":{"name":"","ingredients":[],"instructions":[]}}`
		rec := performRecipeControllerRequest(t, controller.FetcherBotCreateRecipe, http.MethodPost, "/api/_fetcher/recipes", []byte(body), nil)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
		if updatePayload == nil || updatePayload.Status != models.UserEventErrorState {
			t.Fatalf("expected event update to mark error, got %+v", updatePayload)
		}
	})

	t.Run("fetcher bot returns 500 when ready-state event update fails", func(t *testing.T) {
		t.Parallel()

		eventID := uuid.New()
		updateCalls := 0
		controller := NewRecipeController(
			stubUserService{
				updateUserEventByUserIDEventFn: func(_ context.Context, _ string, payload *dtos.UpdateUserEvent) (*models.UserEvent, error) {
					updateCalls++
					if updateCalls == 1 && payload.Status == models.UserEventReady {
						return nil, errors.New("write failed")
					}
					return &models.UserEvent{ID: payload.ID, Status: payload.Status, Error: payload.Error}, nil
				},
			},
			stubRecipeService{
				createMockBannerFn: func(string, string) []models.RecipeMediaVariant { return []models.RecipeMediaVariant{{Alias: "md"}} },
				createRecipeFn: func(context.Context, *dtos.CreateRecipe) (*models.Recipe, error) {
					return recipeFixture(), nil
				},
			},
			stubSearchService{},
			stubStaticService{
				getRandomFallbackImageFn:    func(context.Context) (string, error) { return "fallback.jpg", nil },
				getRandomFallbackImageURLFn: func(string) string { return "https://cdn.example/fallback.jpg" },
			},
			stubFetcherService{},
		)

		body := `{"userId":"user-1","userEventId":"` + eventID.String() + `","recipe":{"name":"Soup","link":"https://example.com","ingredients":[{"name":"salt"}]}}`
		rec := performRecipeControllerRequest(t, controller.FetcherBotCreateRecipe, http.MethodPost, "/api/_fetcher/recipes", []byte(body), nil)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
		if updateCalls < 2 {
			t.Fatalf("expected retry path to mark event errored after ready update failure, got %d calls", updateCalls)
		}
	})
}
