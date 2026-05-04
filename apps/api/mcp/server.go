package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"4ks/apps/api/app"
	"4ks/apps/api/dtos"
	"4ks/apps/api/middleware"
	usersvc "4ks/apps/api/services/user"
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
		mcpserver.WithInstructions("Tools for managing recipes on 4ks.io. Search before create. Fetch the current recipe before update. Do not delete recipes, change profiles, perform admin actions, or upload media."),
		mcpserver.WithToolCapabilities(true),
	)

	mcpSrv.AddTool(mcpsdk.NewTool("search_recipes",
		mcpsdk.WithDescription("Search the authenticated user's recipes by name and ingredients. Call before create_recipe to avoid duplicates."),
		mcpsdk.WithString("query", mcpsdk.Description("Search query.")),
		mcpsdk.WithNumber("limit", mcpsdk.Description("Maximum results to return."), mcpsdk.DefaultNumber(20)),
		mcpsdk.WithReadOnlyHintAnnotation(true),
		mcpsdk.WithDestructiveHintAnnotation(false),
	), s.logTool("search_recipes", s.handleSearchRecipes))

	mcpSrv.AddTool(mcpsdk.NewTool("list_recipes",
		mcpsdk.WithDescription("List recent public 4ks recipes. Prefer search_recipes before create_recipe to check the authenticated user's recipes for duplicates."),
		mcpsdk.WithNumber("limit", mcpsdk.Description("Maximum recipes to return."), mcpsdk.DefaultNumber(20)),
		mcpsdk.WithReadOnlyHintAnnotation(true),
		mcpsdk.WithDestructiveHintAnnotation(false),
	), s.logTool("list_recipes", s.handleListRecipes))

	mcpSrv.AddTool(mcpsdk.NewTool("get_recipe",
		mcpsdk.WithDescription("Fetch the current state of a recipe. Call before update_recipe."),
		mcpsdk.WithString("recipe_id", mcpsdk.Description("Recipe ID."), mcpsdk.Required()),
		mcpsdk.WithReadOnlyHintAnnotation(true),
		mcpsdk.WithDestructiveHintAnnotation(false),
	), s.logTool("get_recipe", s.handleGetRecipe))

	mcpSrv.AddTool(mcpsdk.NewTool("create_recipe",
		mcpsdk.WithDescription("Create a new recipe on 4ks.io for the authenticated user."),
		mcpsdk.WithString("name", mcpsdk.Description("Recipe title."), mcpsdk.Required()),
		mcpsdk.WithString("link", mcpsdk.Description("Source URL.")),
		mcpsdk.WithString("ingredients_json", mcpsdk.Description(`JSON array: [{"name":"flour","quantity":"2 cups"}]`)),
		mcpsdk.WithString("instructions_json", mcpsdk.Description(`JSON array: [{"name":"Step 1","text":"Mix dry ingredients"}]`)),
		mcpsdk.WithReadOnlyHintAnnotation(false),
		mcpsdk.WithDestructiveHintAnnotation(false),
		mcpsdk.WithOpenWorldHintAnnotation(false),
	), s.logTool("create_recipe", s.handleCreateRecipe))

	mcpSrv.AddTool(mcpsdk.NewTool("update_recipe",
		mcpsdk.WithDescription("Update the authenticated user's recipe. Only provided fields are changed; omitted fields keep their current values."),
		mcpsdk.WithString("recipe_id", mcpsdk.Description("Recipe ID."), mcpsdk.Required()),
		mcpsdk.WithString("name", mcpsdk.Description("Recipe title.")),
		mcpsdk.WithString("link", mcpsdk.Description("Source URL.")),
		mcpsdk.WithString("ingredients_json", mcpsdk.Description(`JSON array: [{"name":"flour","quantity":"2 cups"}]`)),
		mcpsdk.WithString("instructions_json", mcpsdk.Description(`JSON array: [{"name":"Step 1","text":"Mix dry ingredients"}]`)),
		mcpsdk.WithReadOnlyHintAnnotation(false),
		mcpsdk.WithDestructiveHintAnnotation(false),
		mcpsdk.WithOpenWorldHintAnnotation(false),
	), s.logTool("update_recipe", s.handleUpdateRecipe))

	mcpSrv.AddTool(mcpsdk.NewTool("list_recipe_forks",
		mcpsdk.WithDescription("List forks of a recipe."),
		mcpsdk.WithString("recipe_id", mcpsdk.Description("Recipe ID."), mcpsdk.Required()),
		mcpsdk.WithReadOnlyHintAnnotation(true),
		mcpsdk.WithDestructiveHintAnnotation(false),
	), s.logTool("list_recipe_forks", s.handleListRecipeForks))

	mcpSrv.AddTool(mcpsdk.NewTool("list_recipe_revisions",
		mcpsdk.WithDescription("List historical revisions of a recipe."),
		mcpsdk.WithString("recipe_id", mcpsdk.Description("Recipe ID."), mcpsdk.Required()),
		mcpsdk.WithReadOnlyHintAnnotation(true),
		mcpsdk.WithDestructiveHintAnnotation(false),
	), s.logTool("list_recipe_revisions", s.handleListRecipeRevisions))

	mcpSrv.AddTool(mcpsdk.NewTool("fork_recipe",
		mcpsdk.WithDescription("Fork a recipe into a new recipe owned by the authenticated user."),
		mcpsdk.WithString("recipe_id", mcpsdk.Description("Recipe ID."), mcpsdk.Required()),
		mcpsdk.WithReadOnlyHintAnnotation(false),
		mcpsdk.WithDestructiveHintAnnotation(false),
		mcpsdk.WithOpenWorldHintAnnotation(false),
	), s.logTool("fork_recipe", s.handleForkRecipe))

	mcpSrv.AddTool(mcpsdk.NewTool("fork_recipe_revision",
		mcpsdk.WithDescription("Fork a specific recipe revision into a new recipe owned by the authenticated user."),
		mcpsdk.WithString("revision_id", mcpsdk.Description("Recipe revision ID."), mcpsdk.Required()),
		mcpsdk.WithReadOnlyHintAnnotation(false),
		mcpsdk.WithDestructiveHintAnnotation(false),
		mcpsdk.WithOpenWorldHintAnnotation(false),
	), s.logTool("fork_recipe_revision", s.handleForkRecipeRevision))

	mcpSrv.AddTool(mcpsdk.NewTool("get_account_status",
		mcpsdk.WithDescription("Return the authenticated user's account status, username, and onboarding state. Read-only. No parameters."),
		mcpsdk.WithReadOnlyHintAnnotation(true),
		mcpsdk.WithDestructiveHintAnnotation(false),
	), s.logTool("get_account_status", s.handleGetAccountStatus))

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
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		log.Info().
			Str("route", route).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rec.status).
			Dur("duration", time.Since(start)).
			Msg("mcp response")
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
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

func (s *Server) logTool(tool string, next mcpserver.ToolHandlerFunc) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		start := time.Now()
		identity, _ := identityFromContext(ctx)

		log.Debug().
			Str("tool", tool).
			Str("userID", identity.UserID).
			Bool("hasEmailClaim", identity.Email != "").
			Strs("argumentKeys", argumentKeys(req)).
			Msg("mcp tool started")

		result, err := next(ctx, req)
		event := log.Info().
			Str("tool", tool).
			Str("userID", identity.UserID).
			Bool("hasEmailClaim", identity.Email != "").
			Dur("duration", time.Since(start))
		if result != nil {
			event.Bool("toolError", result.IsError)
		}

		if err != nil {
			event.Err(err).Msg("mcp tool failed")
			return result, err
		}
		if result != nil && result.IsError {
			event.Msg("mcp tool completed with user-visible error")
			return result, nil
		}
		event.Msg("mcp tool completed")
		return result, nil
	}
}

func (s *Server) handleSearchRecipes(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "search_recipes"
	user, err := s.resolveOrCreateMCPUser(ctx)
	if err != nil {
		return mcpToolError(tool, "authenticated_user", err)
	}
	author := toUserSummary(user)
	if s.services.Search == nil {
		return mcpToolErrorMessage(tool, "search_service", "search service is unavailable")
	}

	limit := clampLimit(req.GetInt("limit", 20))
	query := req.GetString("query", "")
	log.Debug().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("username", author.Username).
		Int("queryLength", len(query)).
		Int("limit", limit).
		Msg("searching MCP recipes")

	results, err := s.services.Search.SearchRecipesByAuthor(query, author.Username, limit)
	if err != nil {
		return mcpToolError(tool, "search_recipes_by_author", err)
	}
	log.Info().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("username", author.Username).
		Int("resultCount", len(results)).
		Msg("MCP recipe search completed")
	return mcpsdk.NewToolResultJSON(map[string]any{"recipes": results})
}

func (s *Server) handleListRecipes(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "list_recipes"
	limit := clampLimit(req.GetInt("limit", 20))
	log.Debug().Str("tool", tool).Int("limit", limit).Msg("listing MCP recipes")

	recipes, err := s.services.Recipe.GetRecipes(ctx, limit)
	if err != nil {
		return mcpToolError(tool, "get_recipes", err)
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

	log.Info().Str("tool", tool).Int("resultCount", len(rows)).Msg("MCP recipe list completed")
	return mcpsdk.NewToolResultJSON(map[string]any{"recipes": rows})
}

func (s *Server) handleGetRecipe(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "get_recipe"
	recipeID, err := req.RequireString("recipe_id")
	if err != nil {
		return mcpToolError(tool, "parse_recipe_id", err)
	}
	log.Debug().Str("tool", tool).Str("recipeID", recipeID).Msg("fetching MCP recipe")
	recipe, err := s.services.Recipe.GetRecipeByID(ctx, recipeID)
	if err != nil {
		return mcpToolError(tool, "get_recipe_by_id", err)
	}
	log.Info().Str("tool", tool).Str("recipeID", recipe.ID).Msg("MCP recipe fetch completed")
	return mcpsdk.NewToolResultJSON(recipe)
}

func (s *Server) handleCreateRecipe(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "create_recipe"
	name, err := req.RequireString("name")
	if err != nil {
		return mcpToolError(tool, "parse_name", err)
	}

	user, err := s.resolveOrCreateMCPUser(ctx)
	if err != nil {
		return mcpToolError(tool, "authenticated_user", err)
	}
	author := toUserSummary(user)

	payload := dtos.CreateRecipe{
		Name:         name,
		Link:         req.GetString("link", ""),
		Ingredients:  parseIngredients(req.GetString("ingredients_json", "")),
		Instructions: parseInstructions(req.GetString("instructions_json", "")),
		Author:       author,
	}
	log.Debug().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("username", author.Username).
		Int("nameLength", len(name)).
		Int("ingredientCount", len(payload.Ingredients)).
		Int("instructionCount", len(payload.Instructions)).
		Bool("hasLink", payload.Link != "").
		Msg("creating MCP recipe")

	if s.services.Static != nil {
		filename, err := s.services.Static.GetRandomFallbackImage(ctx)
		if err == nil {
			url := s.services.Static.GetRandomFallbackImageURL(filename)
			payload.Banner = s.services.Recipe.CreateMockBanner(filename, url)
		} else {
			log.Warn().
				Err(err).
				Str("tool", tool).
				Str("userID", author.ID).
				Msg("failed to get random fallback image for MCP recipe")
		}
	}

	created, err := s.services.Recipe.CreateRecipe(ctx, &payload)
	if err != nil {
		return mcpToolError(tool, "create_recipe", err)
	}
	if s.services.Search != nil {
		if err := s.services.Search.UpsertSearchRecipeDocument(created); err != nil {
			return mcpToolError(tool, "upsert_search_document", err)
		}
	}

	log.Info().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("recipeID", created.ID).
		Msg("MCP recipe created")
	return mcpsdk.NewToolResultJSON(created)
}

func (s *Server) handleUpdateRecipe(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "update_recipe"
	recipeID, err := req.RequireString("recipe_id")
	if err != nil {
		return mcpToolError(tool, "parse_recipe_id", err)
	}
	user, err := s.resolveOrCreateMCPUser(ctx)
	if err != nil {
		return mcpToolError(tool, "authenticated_user", err)
	}
	author := toUserSummary(user)

	log.Debug().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("recipeID", recipeID).
		Strs("argumentKeys", argumentKeys(req)).
		Msg("fetching current recipe before MCP update")
	current, err := s.services.Recipe.GetRecipeByID(ctx, recipeID)
	if err != nil {
		return mcpToolError(tool, "get_current_recipe", err)
	}
	payload := dtos.UpdateRecipe{
		Name:         current.CurrentRevision.Name,
		Link:         current.CurrentRevision.Link,
		Ingredients:  current.CurrentRevision.Ingredients,
		Instructions: current.CurrentRevision.Instructions,
		Banner:       current.CurrentRevision.Banner,
		Author:       author,
	}

	if value, ok, err := stringArgument(req, "name"); err != nil {
		return mcpToolError(tool, "parse_name", err)
	} else if ok {
		payload.Name = value
	}
	if value, ok, err := stringArgument(req, "link"); err != nil {
		return mcpToolError(tool, "parse_link", err)
	} else if ok {
		payload.Link = value
	}
	if raw, ok, err := stringArgument(req, "ingredients_json"); err != nil {
		return mcpToolError(tool, "parse_ingredients_argument", err)
	} else if ok {
		ingredients, err := decodeIngredients(raw)
		if err != nil {
			return mcpToolError(tool, "decode_ingredients", err)
		}
		payload.Ingredients = ingredients
	}
	if raw, ok, err := stringArgument(req, "instructions_json"); err != nil {
		return mcpToolError(tool, "parse_instructions_argument", err)
	} else if ok {
		instructions, err := decodeInstructions(raw)
		if err != nil {
			return mcpToolError(tool, "decode_instructions", err)
		}
		payload.Instructions = instructions
	}

	log.Debug().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("recipeID", recipeID).
		Int("ingredientCount", len(payload.Ingredients)).
		Int("instructionCount", len(payload.Instructions)).
		Msg("updating MCP recipe")
	updated, err := s.services.Recipe.UpdateRecipeByID(ctx, recipeID, &payload)
	if err != nil {
		return mcpToolError(tool, "update_recipe_by_id", err)
	}
	if s.services.Search != nil {
		if err := s.services.Search.UpsertSearchRecipeDocument(updated); err != nil {
			return mcpToolError(tool, "upsert_search_document", err)
		}
	}

	log.Info().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("recipeID", updated.ID).
		Msg("MCP recipe updated")
	return mcpsdk.NewToolResultJSON(updated)
}

func (s *Server) handleListRecipeForks(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "list_recipe_forks"
	recipeID, err := req.RequireString("recipe_id")
	if err != nil {
		return mcpToolError(tool, "parse_recipe_id", err)
	}
	log.Debug().Str("tool", tool).Str("recipeID", recipeID).Msg("listing MCP recipe forks")
	forks, err := s.services.Recipe.GetRecipeForks(ctx, recipeID)
	if err != nil {
		return mcpToolError(tool, "get_recipe_forks", err)
	}
	log.Info().Str("tool", tool).Str("recipeID", recipeID).Int("resultCount", len(forks)).Msg("MCP recipe forks listed")
	return mcpsdk.NewToolResultJSON(map[string]any{"forks": forks})
}

func (s *Server) handleListRecipeRevisions(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "list_recipe_revisions"
	recipeID, err := req.RequireString("recipe_id")
	if err != nil {
		return mcpToolError(tool, "parse_recipe_id", err)
	}
	log.Debug().Str("tool", tool).Str("recipeID", recipeID).Msg("listing MCP recipe revisions")
	revisions, err := s.services.Recipe.GetRecipeRevisions(ctx, recipeID)
	if err != nil {
		return mcpToolError(tool, "get_recipe_revisions", err)
	}
	log.Info().Str("tool", tool).Str("recipeID", recipeID).Int("resultCount", len(revisions)).Msg("MCP recipe revisions listed")
	return mcpsdk.NewToolResultJSON(map[string]any{"revisions": revisions})
}

func (s *Server) handleForkRecipe(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "fork_recipe"
	recipeID, err := req.RequireString("recipe_id")
	if err != nil {
		return mcpToolError(tool, "parse_recipe_id", err)
	}
	user, err := s.resolveOrCreateMCPUser(ctx)
	if err != nil {
		return mcpToolError(tool, "authenticated_user", err)
	}
	author := toUserSummary(user)
	log.Debug().Str("tool", tool).Str("userID", author.ID).Str("recipeID", recipeID).Msg("forking MCP recipe")
	recipe, err := s.services.Recipe.ForkRecipeByID(ctx, recipeID, author)
	if err != nil {
		return mcpToolError(tool, "fork_recipe_by_id", err)
	}
	if s.services.Search != nil {
		if err := s.services.Search.UpsertSearchRecipeDocument(recipe); err != nil {
			return mcpToolError(tool, "upsert_search_document", err)
		}
	}
	log.Info().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("sourceRecipeID", recipeID).
		Str("recipeID", recipe.ID).
		Msg("MCP recipe forked")
	return mcpsdk.NewToolResultJSON(recipe)
}

func (s *Server) handleForkRecipeRevision(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "fork_recipe_revision"
	revisionID, err := req.RequireString("revision_id")
	if err != nil {
		return mcpToolError(tool, "parse_revision_id", err)
	}
	user, err := s.resolveOrCreateMCPUser(ctx)
	if err != nil {
		return mcpToolError(tool, "authenticated_user", err)
	}
	author := toUserSummary(user)
	log.Debug().Str("tool", tool).Str("userID", author.ID).Str("revisionID", revisionID).Msg("forking MCP recipe revision")
	recipe, err := s.services.Recipe.ForkRecipeByRevisionID(ctx, revisionID, author)
	if err != nil {
		return mcpToolError(tool, "fork_recipe_by_revision_id", err)
	}
	if s.services.Search != nil {
		if err := s.services.Search.UpsertSearchRecipeDocument(recipe); err != nil {
			return mcpToolError(tool, "upsert_search_document", err)
		}
	}
	log.Info().
		Str("tool", tool).
		Str("userID", author.ID).
		Str("sourceRevisionID", revisionID).
		Str("recipeID", recipe.ID).
		Msg("MCP recipe revision forked")
	return mcpsdk.NewToolResultJSON(recipe)
}

func (s *Server) handleGetAccountStatus(ctx context.Context, _ mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	const tool = "get_account_status"
	// Security: identity is derived from JWT claims only, not from any request
	// parameter. It is structurally impossible for one user to query another
	// user's account status via this tool.
	user, err := s.resolveOrCreateMCPUser(ctx)
	if err != nil {
		return mcpToolError(tool, "resolve_user", err)
	}
	settingsURL := strings.TrimRight(s.cfg.AppBaseURL, "/") + "/settings"
	return mcpsdk.NewToolResultJSON(map[string]any{
		"username":            user.Username,
		"onboarding_complete": !strings.HasPrefix(user.Username, "user-"),
		"first_login":         user.FirstLogin,
		"settings_url":        settingsURL,
	})
}

func toUserSummary(u *models.User) models.UserSummary {
	return models.UserSummary{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
	}
}

// resolveOrCreateMCPUser looks up the authenticated user, creating a new
// pending account if no matching record exists. It is the single entry point
// for all MCP tool handlers that require a resolved user.
func (s *Server) resolveOrCreateMCPUser(ctx context.Context) (*models.User, error) {
	identity, err := identityFromContext(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("MCP user resolution: missing identity")
		return nil, errors.New("authenticated user is required")
	}

	// Step 1: custom ID claim (set by Auth0 Post-Login Action for existing users).
	if identity.UserID != "" {
		if user, err := s.services.User.GetUserByID(ctx, identity.UserID); err == nil {
			return user, nil
		}
		log.Debug().Str("userID", identity.UserID).Msg("MCP user lookup by custom ID failed")
	}

	// Step 2: email — the canonical identity key.
	if identity.Email != "" {
		if user, err := s.services.User.GetUserByEmail(ctx, identity.Email); err == nil {
			log.Debug().Bool("hasEmailClaim", true).Msg("MCP user resolved by email")
			return user, nil
		}
		log.Debug().Bool("hasEmailClaim", true).Msg("MCP user lookup by email failed; will create account")
	}

	// Step 4: no existing user — create a pending account.
	return s.createPendingUser(ctx, identity)
}

// createPendingUser auto-registers a brand-new 4ks account from an MCP
// connection. It requires a verified email claim (enforced by Auth0 at the
// tenant level). The derived username is based on the email prefix; a random
// fallback is used only when the prefix cannot form a valid username.
func (s *Server) createPendingUser(ctx context.Context, identity mcpIdentity) (*models.User, error) {
	if identity.Email == "" {
		log.Error().Msg("createPendingUser: email claim missing; Auth0 should always populate this claim for verified users")
		return nil, errors.New("email claim required for account creation")
	}

	candidate, err := usersvc.GenerateUsername(identity.Email)
	if err != nil {
		return nil, fmt.Errorf("username generation failed: %w", err)
	}

	const maxAttempts = 20
	username := candidate
	for attempt := 0; attempt <= maxAttempts; attempt++ {
		if attempt > 0 {
			suffix := strconv.Itoa(attempt + 1)
			base := candidate
			if maxBase := 24 - 1 - len(suffix); len(base) > maxBase {
				base = strings.TrimRight(base[:maxBase], "-")
			}
			username = base + "-" + suffix
		}
		err := s.services.User.TestName(ctx, username)
		if err == nil {
			break
		}
		if errors.Is(err, usersvc.ErrUsernameInUse) || errors.Is(err, usersvc.ErrReservedWord) {
			if attempt == maxAttempts {
				return nil, fmt.Errorf("could not find an available username after %d attempts", maxAttempts+1)
			}
			continue
		}
		return nil, fmt.Errorf("username validation failed: %w", err)
	}

	user, err := s.services.User.CreateUserFromMCP(ctx, identity.Email, username)
	if err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}
	log.Info().
		Str("username", user.Username).
		Msg("MCP pending user created")
	return user, nil
}

type mcpIdentity struct {
	UserID string // https://4ks.io/id — Firestore doc ID; absent for brand-new users
	Email  string // https://4ks.io/email — canonical identity key
}

func identityFromContext(ctx context.Context) (mcpIdentity, error) {
	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok {
		return mcpIdentity{}, errors.New("missing JWT claims")
	}

	custom, ok := claims.CustomClaims.(*middleware.CustomClaims)
	if !ok {
		return mcpIdentity{}, errors.New("missing JWT custom claims")
	}

	// custom.ID (https://4ks.io/id) is absent for users who have not yet
	// created a 4ks account — resolveOrCreateMCPUser handles that case.
	return mcpIdentity{
		UserID: custom.ID,
		Email:  custom.Email,
	}, nil
}

func userIDFromContext(ctx context.Context) (string, error) {
	identity, err := identityFromContext(ctx)
	if err != nil {
		return "", err
	}
	return identity.UserID, nil
}

func clampLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 20
	}
	return limit
}

func stringArgument(req mcpsdk.CallToolRequest, key string) (string, bool, error) {
	args := req.GetArguments()
	val, ok := args[key]
	if !ok {
		return "", false, nil
	}
	str, ok := val.(string)
	if !ok {
		return "", true, fmt.Errorf("argument %q is not a string", key)
	}
	return str, true, nil
}

func argumentKeys(req mcpsdk.CallToolRequest) []string {
	args := req.GetArguments()
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func mcpToolError(tool string, stage string, err error) (*mcpsdk.CallToolResult, error) {
	log.Warn().
		Err(err).
		Str("tool", tool).
		Str("stage", stage).
		Msg("MCP tool returning user-visible error")
	return mcpsdk.NewToolResultError(err.Error()), nil
}

func mcpToolErrorMessage(tool string, stage string, message string) (*mcpsdk.CallToolResult, error) {
	log.Warn().
		Str("tool", tool).
		Str("stage", stage).
		Str("error", message).
		Msg("MCP tool returning user-visible error")
	return mcpsdk.NewToolResultError(message), nil
}

func parseIngredients(raw string) []models.Ingredient {
	ingredients, err := decodeIngredients(raw)
	if err != nil {
		return nil
	}
	return ingredients
}

func decodeIngredients(raw string) ([]models.Ingredient, error) {
	var input []struct {
		Name     string `json:"name"`
		Quantity string `json:"quantity"`
	}
	if raw == "" {
		return nil, nil
	}
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return nil, fmt.Errorf("ingredients_json must be a JSON array")
	}

	ingredients := make([]models.Ingredient, 0, len(input))
	for i, item := range input {
		ingredients = append(ingredients, models.Ingredient{
			ID:       i + 1,
			Name:     item.Name,
			Quantity: item.Quantity,
		})
	}
	return ingredients, nil
}

func parseInstructions(raw string) []models.Instruction {
	instructions, err := decodeInstructions(raw)
	if err != nil {
		return nil
	}
	return instructions
}

func decodeInstructions(raw string) ([]models.Instruction, error) {
	var input []struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	if raw == "" {
		return nil, nil
	}
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return nil, fmt.Errorf("instructions_json must be a JSON array")
	}

	instructions := make([]models.Instruction, 0, len(input))
	for i, item := range input {
		instructions = append(instructions, models.Instruction{
			ID:   i + 1,
			Name: item.Name,
			Text: item.Text,
		})
	}
	return instructions, nil
}
