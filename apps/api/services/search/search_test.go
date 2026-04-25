package search

import (
	"4ks/libs/go/models"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/typesense/typesense-go/typesense"
)

func newTestSearchService(t *testing.T, handler func(http.ResponseWriter, *http.Request)) Service {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)

	client := typesense.NewClient(
		typesense.WithServer(server.URL),
		typesense.WithAPIKey("test-key"),
	)

	return New(client)
}

func TestUpsertSearchRecipeDocument(t *testing.T) {
	t.Parallel()

	recipe := &models.Recipe{
		ID: "recipe-1",
		Author: models.UserSummary{
			Username: "chef-user",
		},
		CurrentRevision: models.RecipeRevision{
			Name: "Soup",
			Ingredients: []models.Ingredient{
				{Name: "salt"},
				{Name: "pepper"},
			},
			Banner: []models.RecipeMediaVariant{
				{Alias: "sm", URL: "https://cdn.example/sm.jpg"},
				{Alias: "md", URL: "https://cdn.example/md.jpg"},
			},
		},
	}

	var method string
	var path string
	var payload map[string]any
	service := newTestSearchService(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		body, err := io.ReadAll(r.Body)
		if err == nil {
			_ = json.Unmarshal(body, &payload)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"recipe-1"}`))
	})

	if err := service.UpsertSearchRecipeDocument(recipe); err != nil {
		t.Fatalf("UpsertSearchRecipeDocument returned error: %v", err)
	}
	if method != http.MethodPost {
		t.Fatalf("expected POST, got %s", method)
	}
	if !strings.Contains(path, "/collections/recipes/documents") {
		t.Fatalf("unexpected path: %s", path)
	}
	if payload["author"] != "chef-user" || payload["name"] != "Soup" || payload["imageUrl"] != "https://cdn.example/md.jpg" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	ingredients, ok := payload["ingredients"].([]any)
	if !ok || len(ingredients) != 2 {
		t.Fatalf("unexpected ingredients payload: %#v", payload["ingredients"])
	}
}

func TestRemoveSearchRecipeDocument(t *testing.T) {
	t.Parallel()

	var method string
	var path string
	service := newTestSearchService(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"num_deleted":1}`))
	})

	if err := service.RemoveSearchRecipeDocument("recipe-1"); err != nil {
		t.Fatalf("RemoveSearchRecipeDocument returned error: %v", err)
	}
	if method != http.MethodDelete {
		t.Fatalf("expected DELETE, got %s", method)
	}
	if !strings.Contains(path, "/collections/recipes/documents/recipe-1") {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestCreateSearchRecipeCollection(t *testing.T) {
	t.Parallel()

	var method string
	var path string
	var bodyText string
	service := newTestSearchService(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		body, err := io.ReadAll(r.Body)
		if err == nil {
			bodyText = string(body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"name":"recipes","num_documents":0,"fields":[]}`))
	})

	if err := service.CreateSearchRecipeCollection(); err != nil {
		t.Fatalf("CreateSearchRecipeCollection returned error: %v", err)
	}
	if method != http.MethodPost {
		t.Fatalf("expected POST, got %s", method)
	}
	if !strings.Contains(path, "/collections") {
		t.Fatalf("unexpected path: %s", path)
	}
	if !strings.Contains(bodyText, `"name":"recipes"`) || !strings.Contains(bodyText, `"imageUrl"`) {
		t.Fatalf("unexpected schema body: %s", bodyText)
	}
}
