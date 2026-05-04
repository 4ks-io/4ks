package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"4ks/apps/api/app"
	"4ks/apps/api/dtos"
	"4ks/apps/api/middleware"
	usersvc "4ks/apps/api/services/user"
	"4ks/apps/api/utils"
	models "4ks/libs/go/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/google/uuid"
	mcpsdk "github.com/mark3labs/mcp-go/mcp"
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

func TestResolveOrCreateMCPUserFallsBackToEmail(t *testing.T) {
	claims := &validator.ValidatedClaims{
		CustomClaims: &middleware.CustomClaims{
			ID:    "auth0-claim-id",
			Email: "chef@example.com",
		},
	}
	ctx := context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, claims)

	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(context.Context, string) (*models.User, error) {
				return nil, errors.New("user not found")
			},
			getUserByEmailFn: func(_ context.Context, email string) (*models.User, error) {
				if email != "chef@example.com" {
					t.Fatalf("unexpected email lookup: %q", email)
				}
				return &models.User{ID: "firestore-user-id", Username: "chef", DisplayName: "Chef"}, nil
			},
		},
	})

	user, err := server.resolveOrCreateMCPUser(ctx)
	if err != nil {
		t.Fatalf("resolveOrCreateMCPUser returned error: %v", err)
	}
	if user.ID != "firestore-user-id" || user.Username != "chef" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

// --- get_account_status tests ---

func makeClaimsCtx(id, email string) context.Context {
	claims := &validator.ValidatedClaims{
		CustomClaims: &middleware.CustomClaims{ID: id, Email: email},
	}
	return context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, claims)
}

func TestGetAccountStatus_AuthenticatedUser(t *testing.T) {
	ctx := makeClaimsCtx("uid-1", "delorme.nic@gmail.com")
	cfg := utils.MinimalRuntimeConfig()
	cfg.MCP.AppBaseURL = "https://www.4ks.io"

	server := New(cfg, app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(_ context.Context, id string) (*models.User, error) {
				return &models.User{ID: id, Username: "delorme-nic", FirstLogin: false}, nil
			},
		},
	})

	result, err := server.handleGetAccountStatus(ctx, mcpsdk.CallToolRequest{})
	if err != nil || result == nil || result.IsError {
		t.Fatalf("expected success, got err=%v result=%v", err, result)
	}

	raw := extractJSON(t, result)
	if raw["username"] != "delorme-nic" {
		t.Errorf("username = %v, want delorme-nic", raw["username"])
	}
	if raw["onboarding_complete"] != true {
		t.Errorf("onboarding_complete = %v, want true", raw["onboarding_complete"])
	}
	if raw["first_login"] != false {
		t.Errorf("first_login = %v, want false", raw["first_login"])
	}
	if su, _ := raw["settings_url"].(string); su != "https://www.4ks.io/settings" {
		t.Errorf("settings_url = %q, want https://www.4ks.io/settings", su)
	}
}

func TestGetAccountStatus_NoJWTClaims_ReturnsError(t *testing.T) {
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{},
	})
	result, err := server.handleGetAccountStatus(context.Background(), mcpsdk.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result for missing JWT")
	}
}

func TestGetAccountStatus_UserNotFound_ReturnsError(t *testing.T) {
	ctx := makeClaimsCtx("uid-1", "")
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(context.Context, string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			// email is empty so createPendingUser will fail too
		},
	})
	result, err := server.handleGetAccountStatus(ctx, mcpsdk.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("expected error result when user cannot be resolved")
	}
}

func TestGetAccountStatus_FirstLoginTrue_Reflected(t *testing.T) {
	ctx := makeClaimsCtx("uid-1", "chef-user@example.com")
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(_ context.Context, id string) (*models.User, error) {
				return &models.User{ID: id, Username: "chef-user", FirstLogin: true}, nil
			},
		},
	})
	result, err := server.handleGetAccountStatus(ctx, mcpsdk.CallToolRequest{})
	if err != nil || result == nil || result.IsError {
		t.Fatalf("expected success: err=%v", err)
	}
	raw := extractJSON(t, result)
	if raw["first_login"] != true {
		t.Errorf("first_login = %v, want true", raw["first_login"])
	}
}

func TestGetAccountStatus_FirstLoginNotFlipped(t *testing.T) {
	ctx := makeClaimsCtx("uid-1", "chef-user@example.com")
	updateCalled := false
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(_ context.Context, id string) (*models.User, error) {
				return &models.User{ID: id, Username: "chef-user", FirstLogin: false}, nil
			},
			updateUserByIDFn: func(context.Context, string, *dtos.UpdateUser) (*models.User, error) {
				updateCalled = true
				return nil, nil
			},
		},
	})
	if _, err := server.handleGetAccountStatus(ctx, mcpsdk.CallToolRequest{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updateCalled {
		t.Fatal("get_account_status must not flip first_login via UpdateUserByID")
	}
}

func TestGetAccountStatus_AutoGeneratedUsername_OnboardingIncomplete(t *testing.T) {
	ctx := makeClaimsCtx("uid-1", "chef@example.com")
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(_ context.Context, id string) (*models.User, error) {
				return &models.User{ID: id, Username: "user-ab1234", FirstLogin: true}, nil
			},
		},
	})
	result, err := server.handleGetAccountStatus(ctx, mcpsdk.CallToolRequest{})
	if err != nil || result == nil || result.IsError {
		t.Fatalf("expected success: err=%v", err)
	}
	raw := extractJSON(t, result)
	if raw["onboarding_complete"] != false {
		t.Errorf("onboarding_complete = %v, want false for auto-generated username", raw["onboarding_complete"])
	}
}

func TestGetAccountStatus_SettingsURLUsesAppBaseURL(t *testing.T) {
	ctx := makeClaimsCtx("uid-1", "chef-user@example.com")
	cfg := utils.MinimalRuntimeConfig()
	cfg.MCP.AppBaseURL = "https://custom.4ks.io"
	server := New(cfg, app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(_ context.Context, id string) (*models.User, error) {
				return &models.User{ID: id, Username: "chef-user"}, nil
			},
		},
	})
	result, err := server.handleGetAccountStatus(ctx, mcpsdk.CallToolRequest{})
	if err != nil || result == nil || result.IsError {
		t.Fatalf("expected success: err=%v", err)
	}
	raw := extractJSON(t, result)
	if su, _ := raw["settings_url"].(string); su != "https://custom.4ks.io/settings" {
		t.Errorf("settings_url = %q, want https://custom.4ks.io/settings", su)
	}
}

// --- createPendingUser suffix loop tests ---

func TestCreatePendingUser_SuffixAppendedWhenBaseTaken(t *testing.T) {
	ctx := makeClaimsCtx("auth0|sub-1", "delorme.nic@gmail.com")
	attempts := 0
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(context.Context, string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			testNameFn: func(_ context.Context, name string) error {
				attempts++
				if name == "delorme-nic" {
					return usersvc.ErrUsernameInUse
				}
				return nil
			},
			createUserFromMCPFn: func(_ context.Context, _, username string) (*models.User, error) {
				return &models.User{ID: "test-user-id", Username: username}, nil
			},
		},
	})

	user, err := server.resolveOrCreateMCPUser(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "delorme-nic-2" {
		t.Errorf("username = %q, want delorme-nic-2", user.Username)
	}
	if attempts < 2 {
		t.Errorf("expected at least 2 TestName calls, got %d", attempts)
	}
}

func TestCreatePendingUser_ReservedWordTriggersSuffix(t *testing.T) {
	ctx := makeClaimsCtx("auth0|sub-2", "delorme.nic@gmail.com")
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(context.Context, string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			testNameFn: func(_ context.Context, name string) error {
				if name == "delorme-nic" {
					return usersvc.ErrReservedWord
				}
				return nil
			},
			createUserFromMCPFn: func(_ context.Context, _, username string) (*models.User, error) {
				return &models.User{ID: "test-user-id", Username: username}, nil
			},
		},
	})

	user, err := server.resolveOrCreateMCPUser(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "delorme-nic-2" {
		t.Errorf("username = %q, want delorme-nic-2", user.Username)
	}
}

func TestCreatePendingUser_AllSuffixesExhaustedReturnsError(t *testing.T) {
	ctx := makeClaimsCtx("auth0|sub-3", "delorme.nic@gmail.com")
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(context.Context, string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			testNameFn: func(context.Context, string) error {
				return usersvc.ErrUsernameInUse
			},
		},
	})

	_, err := server.resolveOrCreateMCPUser(ctx)
	if err == nil {
		t.Fatal("expected error when all suffix attempts are exhausted")
	}
}

func TestCreatePendingUser_EmptyEmailReturnsError(t *testing.T) {
	ctx := makeClaimsCtx("auth0|sub-4", "") // empty email
	server := New(utils.MinimalRuntimeConfig(), app.Services{
		User: mcpTestUserService{
			getUserByIDFn: func(context.Context, string) (*models.User, error) {
				return nil, errors.New("not found")
			},
		},
	})

	_, err := server.resolveOrCreateMCPUser(ctx)
	if err == nil {
		t.Fatal("expected error when email claim is missing")
	}
}

// extractJSON decodes the first text content of an MCP result into a map.
func extractJSON(t *testing.T, result *mcpsdk.CallToolResult) map[string]any {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	raw, err := json.Marshal(result.Content[0])
	if err != nil {
		t.Fatalf("marshal content: %v", err)
	}
	var outer struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &outer); err != nil {
		t.Fatalf("unmarshal outer: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(outer.Text), &m); err != nil {
		t.Fatalf("unmarshal JSON text: %v", err)
	}
	return m
}

// --- stub service ---

type mcpTestUserService struct {
	getUserByIDFn           func(context.Context, string) (*models.User, error)
	getUserByEmailFn        func(context.Context, string) (*models.User, error)
	createUserFromMCPFn func(context.Context, string, string) (*models.User, error)
	updateUserByIDFn        func(context.Context, string, *dtos.UpdateUser) (*models.User, error)
	testNameFn              func(context.Context, string) error
}

func (s mcpTestUserService) GetAllUsers(context.Context) ([]*models.User, error) {
	return nil, nil
}

func (s mcpTestUserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if s.getUserByIDFn != nil {
		return s.getUserByIDFn(ctx, id)
	}
	return nil, errors.New("user not found")
}

func (s mcpTestUserService) GetUserByUsername(context.Context, string) (*models.User, error) {
	return nil, nil
}

func (s mcpTestUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if s.getUserByEmailFn != nil {
		return s.getUserByEmailFn(ctx, email)
	}
	return nil, errors.New("user not found")
}

func (s mcpTestUserService) CreateUser(context.Context, string, string, *dtos.CreateUser) (*models.User, error) {
	return nil, nil
}

func (s mcpTestUserService) CreateUserFromMCP(ctx context.Context, email, username string) (*models.User, error) {
	if s.createUserFromMCPFn != nil {
		return s.createUserFromMCPFn(ctx, email, username)
	}
	return nil, errors.New("CreateUserFromMCP not configured")
}

func (s mcpTestUserService) UpdateUserByID(ctx context.Context, id string, u *dtos.UpdateUser) (*models.User, error) {
	if s.updateUserByIDFn != nil {
		return s.updateUserByIDFn(ctx, id, u)
	}
	return nil, nil
}

func (s mcpTestUserService) DeleteUser(context.Context, string) error {
	return nil
}

func (s mcpTestUserService) CreateUserEventByUserID(context.Context, string, *dtos.CreateUserEvent) (*models.UserEvent, error) {
	return nil, nil
}

func (s mcpTestUserService) UpdateUserEventByUserIDEventID(context.Context, string, *dtos.UpdateUserEvent) (*models.UserEvent, error) {
	return nil, nil
}

func (s mcpTestUserService) RemoveUserEventByUserIDEventID(context.Context, string, uuid.UUID) error {
	return nil
}

func (s mcpTestUserService) TestName(ctx context.Context, name string) error {
	if s.testNameFn != nil {
		return s.testNameFn(ctx, name)
	}
	return nil
}

func (s mcpTestUserService) TestValidName(string) bool {
	return true
}

func (s mcpTestUserService) TestReservedWord(string) bool {
	return false
}

func (s mcpTestUserService) TestAvailableName(context.Context, string) (bool, error) {
	return true, nil
}
