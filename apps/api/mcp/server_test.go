package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"4ks/apps/api/app"
	"4ks/apps/api/utils"
)

func TestStartReturnsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() {
		done <- New(nil, app.Services{}).Start(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

func TestProtectedResourceMetadata(t *testing.T) {
	cfg := utils.MinimalRuntimeConfig()
	cfg.MCP.Enabled = true
	cfg.MCP.BaseURL = "https://example.ngrok.app/mcp"
	cfg.MCP.Audience = "https://example.ngrok.app/mcp"
	cfg.Auth0.Domain = "tenant.auth0.com"

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/.well-known/oauth-protected-resource", nil)
	rec := httptest.NewRecorder()

	New(cfg, app.Services{}).handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Resource             string   `json:"resource"`
		AuthorizationServers []string `json:"authorization_servers"`
		ScopesSupported      []string `json:"scopes_supported"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body.Resource != "https://example.ngrok.app/mcp" {
		t.Fatalf("unexpected resource: %q", body.Resource)
	}
	if len(body.AuthorizationServers) != 1 || body.AuthorizationServers[0] != "https://tenant.auth0.com" {
		t.Fatalf("unexpected authorization servers: %#v", body.AuthorizationServers)
	}
}

func TestMCPRoutesAreMountedUnderAPIMCP(t *testing.T) {
	cfg := utils.MinimalRuntimeConfig()
	cfg.MCP.Enabled = true
	cfg.MCP.BaseURL = "https://example.ngrok.app/mcp"
	cfg.Auth0.Domain = "tenant.auth0.com"

	handler := New(cfg, app.Services{}).handler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected legacy /sse to be 404, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/mcp/sse", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected /api/mcp/sse to require auth, got %d", rec.Code)
	}
	want := `Bearer resource_metadata="https://example.ngrok.app/mcp/.well-known/oauth-protected-resource", error="invalid_token", error_description="JWT validation failed"`
	if got := rec.Header().Get("WWW-Authenticate"); got != want {
		t.Fatalf("unexpected WWW-Authenticate header:\nwant %q\n got %q", want, got)
	}
}

func TestMCPAudienceFallsBackToBaseURL(t *testing.T) {
	cfg := utils.MinimalRuntimeConfig()
	cfg.MCP.Enabled = true
	cfg.MCP.BaseURL = "https://example.ngrok.app/mcp"
	cfg.MCP.Audience = ""

	server := New(cfg, app.Services{})
	if got := server.audience(); got != "https://example.ngrok.app/mcp" {
		t.Fatalf("unexpected fallback audience: %q", got)
	}
}

func TestPublicURLParts(t *testing.T) {
	if got := publicOrigin("https://example.ngrok.app/mcp"); got != "https://example.ngrok.app" {
		t.Fatalf("unexpected public origin: %q", got)
	}
	if got := publicBasePath("https://example.ngrok.app/mcp"); got != "/mcp" {
		t.Fatalf("unexpected public base path: %q", got)
	}
	if got := publicBasePath("https://example.ngrok.app"); got != "/mcp" {
		t.Fatalf("unexpected default public base path: %q", got)
	}
}

func TestParseRecipeArguments(t *testing.T) {
	ingredients := parseIngredients(`[{"name":"flour","quantity":"2 cups"},{"name":"salt","quantity":"1 tsp"}]`)
	if len(ingredients) != 2 {
		t.Fatalf("expected 2 ingredients, got %d", len(ingredients))
	}
	if ingredients[0].ID != 1 || ingredients[0].Name != "flour" || ingredients[1].ID != 2 {
		t.Fatalf("unexpected ingredients: %#v", ingredients)
	}

	instructions := parseInstructions(`[{"name":"Step 1","text":"Mix"},{"name":"Step 2","text":"Bake"}]`)
	if len(instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(instructions))
	}
	if instructions[0].ID != 1 || instructions[0].Text != "Mix" || instructions[1].ID != 2 {
		t.Fatalf("unexpected instructions: %#v", instructions)
	}
}
