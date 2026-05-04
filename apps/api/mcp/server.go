package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"4ks/apps/api/app"
	"4ks/apps/api/dtos"
	"4ks/apps/api/middleware"
	"4ks/apps/api/utils"
	models "4ks/libs/go/models"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

const (
	shutdownTimeout              = 30 * time.Second
	internalBasePath             = "/api/mcp"
	publicDefaultBasePath        = "/mcp"
	protectedResourceMetadataURL = "/.well-known/oauth-protected-resource"
)

// Server owns the MCP SSE router and HTTP server lifecycle.
type Server struct {
	cfg        utils.MCPConfig
	auth0      utils.Auth0Config
	services   app.Services
	httpServer *http.Server
}

// New wires the MCP server. The listener is disabled unless cfg.MCP.Enabled is true.
func New(cfg *utils.RuntimeConfig, svc app.Services) *Server {
	if cfg == nil {
		cfg = utils.MinimalRuntimeConfig()
		cfg.MCP.Enabled = false
	}
	return &Server{
		cfg:      cfg.MCP,
		auth0:    cfg.Auth0,
		services: svc,
	}
}

// Start begins listening and blocks until ctx is cancelled, then gracefully shuts down.
func (s *Server) Start(ctx context.Context) error {
	if !s.cfg.Enabled {
		<-ctx.Done()
		return nil
	}

	s.httpServer = &http.Server{
		Addr:    "0.0.0.0:" + s.cfg.Port,
		Handler: s.handler(),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	log.Info().
		Str("baseURL", s.cfg.BaseURL).
		Str("audience", s.audience()).
		Str("port", s.cfg.Port).
		Msg("starting mcp server")

	errc := make(chan error, 1)
	go func() {
		err := s.httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errc <- nil
			return
		}
		errc <- err
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errc; err != nil {
		return err
	}

	log.Info().Msg("mcp server stopped")
	return nil
}

func (s *Server) handler() http.Handler {
	mcpSrv := mcpserver.NewMCPServer(
		"4ks-recipes",
		"0.1.0",
		mcpserver.WithInstructions("Tools for reading and creating recipes on 4ks.io. Call list_recipes before create_recipe to avoid duplicates."),
		mcpserver.WithToolCapabilities(true),
	)

	mcpSrv.AddTool(mcpsdk.NewTool("list_recipes",
		mcpsdk.WithDescription("List existing 4ks recipes. Call before create_recipe to check for duplicates."),
		mcpsdk.WithNumber("limit", mcpsdk.Description("Maximum recipes to return."), mcpsdk.DefaultNumber(20)),
		mcpsdk.WithReadOnlyHintAnnotation(true),
		mcpsdk.WithDestructiveHintAnnotation(false),
	), s.handleListRecipes)

	mcpSrv.AddTool(mcpsdk.NewTool("create_recipe",
		mcpsdk.WithDescription("Create a new recipe on 4ks.io for the authenticated user."),
		mcpsdk.WithString("name", mcpsdk.Description("Recipe title."), mcpsdk.Required()),
		mcpsdk.WithString("link", mcpsdk.Description("Source URL.")),
		mcpsdk.WithString("ingredients_json", mcpsdk.Description(`JSON array: [{"name":"flour","quantity":"2 cups"}]`)),
		mcpsdk.WithString("instructions_json", mcpsdk.Description(`JSON array: [{"name":"Step 1","text":"Mix dry ingredients"}]`)),
		mcpsdk.WithReadOnlyHintAnnotation(false),
		mcpsdk.WithDestructiveHintAnnotation(false),
		mcpsdk.WithOpenWorldHintAnnotation(false),
	), s.handleCreateRecipe)

	sse := mcpserver.NewSSEServer(
		mcpSrv,
		mcpserver.WithBaseURL(publicOrigin(s.cfg.BaseURL)),
		mcpserver.WithStaticBasePath(publicBasePath(s.cfg.BaseURL)),
		mcpserver.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			if claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims); ok {
				ctx = context.WithValue(ctx, jwtmiddleware.ContextKey{}, claims)
			}
			return ctx
		}),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+internalBasePath+protectedResourceMetadataURL, s.handleProtectedResource)
	mux.Handle(internalBasePath+"/sse", s.logRoute("mcp sse", s.jwtMiddleware()(sse.SSEHandler())))
	mux.Handle(internalBasePath+"/message", s.logRoute("mcp message", s.jwtMiddleware()(sse.MessageHandler())))
	return mux
}

func (s *Server) logRoute(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("route", route).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("baseURL", s.cfg.BaseURL).
			Msg("handling mcp route")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleProtectedResource(w http.ResponseWriter, r *http.Request) {
	log.Debug().
		Str("route", "mcp protected resource metadata").
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("baseURL", s.cfg.BaseURL).
		Msg("handling mcp route")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"resource":              strings.TrimRight(s.cfg.BaseURL, "/"),
		"authorization_servers": []string{"https://" + s.auth0.Domain},
		"scopes_supported":      []string{"openid", "profile", "email"},
	})
}

func (s *Server) jwtMiddleware() func(http.Handler) http.Handler {
	auth0 := s.auth0
	auth0.Audience = s.audience()

	return middleware.EnforceJWTWithErrorHandler(auth0, func(w http.ResponseWriter, _ *http.Request, err error) {
		log.Error().Err(err).Msg("failed to validate MCP JWT")
		w.Header().Set("WWW-Authenticate", s.wwwAuthenticate("invalid_token", "JWT validation failed"))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (s *Server) audience() string {
	if s.cfg.Audience != "" {
		return s.cfg.Audience
	}
	return strings.TrimRight(s.cfg.BaseURL, "/")
}

func (s *Server) wwwAuthenticate(code string, description string) string {
	return fmt.Sprintf(
		`Bearer resource_metadata="%s%s", error="%s", error_description="%s"`,
		strings.TrimRight(s.cfg.BaseURL, "/"),
		protectedResourceMetadataURL,
		code,
		description,
	)
}

func publicOrigin(rawURL string) string {
	parsed, err := url.Parse(strings.TrimRight(rawURL, "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func publicBasePath(rawURL string) string {
	parsed, err := url.Parse(strings.TrimRight(rawURL, "/"))
	if err == nil && parsed.Path != "" && parsed.Path != "/" {
		return parsed.Path
	}
	return publicDefaultBasePath
}

func (s *Server) handleListRecipes(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	limit := req.GetInt("limit", 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	recipes, err := s.services.Recipe.GetRecipes(ctx, limit)
	if err != nil {
		return mcpsdk.NewToolResultError(err.Error()), nil
	}

	type row struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	rows := make([]row, 0, len(recipes))
	for _, recipe := range recipes {
		rows = append(rows, row{
			ID:   recipe.ID,
			Name: recipe.CurrentRevision.Name,
		})
	}

	return mcpsdk.NewToolResultJSON(map[string]any{"recipes": rows})
}

func (s *Server) handleCreateRecipe(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return mcpsdk.NewToolResultError("authenticated user is required"), nil
	}

	name, err := req.RequireString("name")
	if err != nil {
		return mcpsdk.NewToolResultError(err.Error()), nil
	}

	payload := dtos.CreateRecipe{
		Name:         name,
		Link:         req.GetString("link", ""),
		Ingredients:  parseIngredients(req.GetString("ingredients_json", "")),
		Instructions: parseInstructions(req.GetString("instructions_json", "")),
	}

	author, err := s.services.User.GetUserByID(ctx, userID)
	if err != nil {
		return mcpsdk.NewToolResultError(err.Error()), nil
	}
	payload.Author = models.UserSummary{
		ID:          userID,
		Username:    author.Username,
		DisplayName: author.DisplayName,
	}

	if s.services.Static != nil {
		filename, err := s.services.Static.GetRandomFallbackImage(ctx)
		if err == nil {
			url := s.services.Static.GetRandomFallbackImageURL(filename)
			payload.Banner = s.services.Recipe.CreateMockBanner(filename, url)
		} else {
			log.Error().Err(err).Msg("failed to get random fallback image for MCP recipe")
		}
	}

	created, err := s.services.Recipe.CreateRecipe(ctx, &payload)
	if err != nil {
		return mcpsdk.NewToolResultError(err.Error()), nil
	}
	if s.services.Search != nil {
		if err := s.services.Search.UpsertSearchRecipeDocument(created); err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
	}

	return mcpsdk.NewToolResultJSON(created)
}

func userIDFromContext(ctx context.Context) (string, error) {
	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok {
		return "", errors.New("missing JWT claims")
	}

	custom, ok := claims.CustomClaims.(*middleware.CustomClaims)
	if !ok || custom.ID == "" {
		return "", errors.New("missing JWT user ID claim")
	}

	return custom.ID, nil
}

func parseIngredients(raw string) []models.Ingredient {
	var input []struct {
		Name     string `json:"name"`
		Quantity string `json:"quantity"`
	}
	if raw == "" || json.Unmarshal([]byte(raw), &input) != nil {
		return nil
	}

	ingredients := make([]models.Ingredient, 0, len(input))
	for i, item := range input {
		ingredients = append(ingredients, models.Ingredient{
			ID:       i + 1,
			Name:     item.Name,
			Quantity: item.Quantity,
		})
	}
	return ingredients
}

func parseInstructions(raw string) []models.Instruction {
	var input []struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	if raw == "" || json.Unmarshal([]byte(raw), &input) != nil {
		return nil
	}

	instructions := make([]models.Instruction, 0, len(input))
	for i, item := range input {
		instructions = append(instructions, models.Instruction{
			ID:   i + 1,
			Name: item.Name,
			Text: item.Text,
		})
	}
	return instructions
}
