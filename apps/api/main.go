// package main is the entrypoint for the api
package main

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "4ks/apps/api/docs"

	controllers "4ks/apps/api/controllers"
	middleware "4ks/apps/api/middleware"
	fetcherService "4ks/apps/api/services/fetcher"
	recipeService "4ks/apps/api/services/recipe"
	searchService "4ks/apps/api/services/search"
	staticService "4ks/apps/api/services/static"
	userService "4ks/apps/api/services/user"
	utils "4ks/apps/api/utils"
	tracing "4ks/libs/go/tracer"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/typesense/typesense-go/typesense"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Controllers contains the controllers
type Controllers struct {
	User   controllers.UserController
	Recipe controllers.RecipeController
	Search controllers.SearchController
	System controllers.SystemController
}

// getAPIVersion returns the api version stored in the VERSION file
func getAPIVersion(versionFilePath string) string {
	version := "0.0.0"
	if versionFilePath != "" {
		v, err := os.ReadFile(versionFilePath)
		if err != nil {
			panic(err)
		}
		version = strings.TrimSuffix(string(v), "\n")
	}
	return version
}

// configureLogging configures global logging based on the config file and flags.
func configureLogging() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	// Set log level
	zerolog.SetGlobalLevel(0)
	log.Logger = log.With().Caller().Logger()
	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// EnforceAuth enforces authentication
func EnforceAuth(authConfig utils.Auth0Config, r *gin.RouterGroup) {
	// JWT validation runs first so downstream middleware and handlers can rely on
	// the authenticated identity stored in Gin context.
	r.Use(adapter.Wrap(middleware.EnforceJWT(authConfig)))
	r.Use(middleware.AppendCustomClaims())
}

// AppendRoutes appends routes to the router
func AppendRoutes(cfg *utils.RuntimeConfig, r *gin.Engine, c *Controllers) {
	sysFlags := cfg.SystemFlags()

	// One shared store lets related routes reuse buckets consistently while
	// keeping the limiter implementation local to this process.
	rateLimitStore := middleware.NewLimiterStore()

	publicReadLimit := middleware.NewRateLimitMiddleware(rateLimitStore, middleware.RateLimitPolicy{
		Name: "public-read",
		// Public recipe reads are anonymous, so the limiter keys by resolved client IP.
		Rules: []middleware.RateLimitRule{
			middleware.QPSRule(5),
			middleware.QPMRule(120),
		},
		KeyFunc: middleware.RateLimitByIP,
	})
	authenticatedWriteLimit := middleware.NewRateLimitMiddleware(rateLimitStore, middleware.RateLimitPolicy{
		Name: "authenticated-write",
		// Authenticated writes are keyed by user ID first so one NATed IP does not
		// throttle unrelated users behind the same egress.
		Rules: []middleware.RateLimitRule{
			middleware.QPSRule(2),
			middleware.QPMRule(30),
		},
		KeyFunc: middleware.RateLimitByUserOrIP,
	})
	recipeFetchLimit := middleware.NewRateLimitMiddleware(rateLimitStore, middleware.RateLimitPolicy{
		Name: "recipe-fetch",
		// Fetch-by-URL fanout is expensive, so this policy stays tighter than the
		// generic write budget across both burst and sustained windows.
		Rules: []middleware.RateLimitRule{
			middleware.QPSRule(1),
			middleware.QPMRule(3),
		},
		KeyFunc: middleware.RateLimitByUserOrIP,
	})
	userCreateLimit := middleware.NewRateLimitMiddleware(rateLimitStore, middleware.RateLimitPolicy{
		Name: "user-create",
		// Account creation is low-volume by design and should resist scripted abuse.
		Rules: []middleware.RateLimitRule{
			middleware.QPSRule(1),
			middleware.QPMRule(3),
		},
		KeyFunc: middleware.RateLimitByUserOrIP,
	})
	usernameCheckLimit := middleware.NewRateLimitMiddleware(rateLimitStore, middleware.RateLimitPolicy{
		Name: "username-check",
		// Username checks are public-facing validation traffic, so they get their
		// own narrower bucket instead of sharing the generic write pool.
		Rules: []middleware.RateLimitRule{
			middleware.QPSRule(2),
			middleware.QPMRule(20),
		},
		KeyFunc: middleware.RateLimitByUserOrIP,
	})
	mediaInitLimit := middleware.NewRateLimitMiddleware(rateLimitStore, middleware.RateLimitPolicy{
		Name: "media-init",
		// Media initialization allocates upload metadata and signed URLs, so it is
		// rate-limited separately from other write endpoints.
		Rules: []middleware.RateLimitRule{
			middleware.QPSRule(1),
			middleware.QPMRule(6),
		},
		KeyFunc: middleware.RateLimitByUserOrIP,
	})

	// system
	r.GET("/api/ready", c.System.CheckReadiness)
	// /api/healthcheck is development-only; block this path at the GCP load balancer in production.
	if sysFlags.Development {
		r.GET("/api/healthcheck", c.System.Healthcheck)
	}

	// api
	api := r.Group("/api")
	{
		// otel
		if sysFlags.JaegerEnabled {
			api.Use(otelgin.Middleware("4ks-api"))
		}

		// swagger
		if value := cfg.Features.SwaggerEnabled; value {
			log.Info().Bool("enabled", value).Str("prefix", cfg.Routes.SwaggerURLPrefix).Msg("Swagger")
			swaggerURL := ginSwagger.URL(cfg.Routes.SwaggerURLPrefix + "/swagger/doc.json") // The url pointing to API definition
			api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, swaggerURL))
		}

		// speed up data seeding
		if sysFlags.Development {
			develop := api.Group("/_dev")
			{
				develop.POST("/recipes", c.Recipe.BotCreateRecipe)
				develop.POST("/init-search-collections", c.Search.CreateSearchRecipeCollection)
			}
		}

		// fetcher
		fetcher := api.Group("/_fetcher")
		{
			// uses custom encrypted timestamp validation auth shceme using X-4ks-Auth header and pre-shared secret
			fetcher.POST("/recipes", middleware.AuthorizeFetcher(cfg.Fetcher), c.Recipe.FetcherBotCreateRecipe)
		}

		// recipes
		recipes := api.Group("/recipes")
		{
			// Public recipe reads are limited separately from authenticated writes.
			recipes.GET("/:id", publicReadLimit, c.Recipe.GetRecipe)
			recipes.GET("/", publicReadLimit, c.Recipe.GetRecipes)
			recipes.GET("/:id/forks", publicReadLimit, c.Recipe.GetRecipeForks)
			recipes.GET("/:id/revisions", publicReadLimit, c.Recipe.GetRecipeRevisions)
			recipes.GET("/revisions/:revisionID", publicReadLimit, c.Recipe.GetRecipeRevision)
			recipes.GET("/:id/media", publicReadLimit, c.Recipe.GetRecipeMedia)
			recipes.GET("/author/:username", publicReadLimit, c.Recipe.GetRecipesByUsername)

			// authenticated routes below this line
			EnforceAuth(cfg.Auth0, recipes)

			recipes.POST("/", authenticatedWriteLimit, c.Recipe.CreateRecipe)
			recipes.POST("/fetch", recipeFetchLimit, c.Recipe.FetchRecipe)
			recipes.PATCH("/:id", authenticatedWriteLimit, c.Recipe.UpdateRecipe)
			recipes.POST("/:id/star", authenticatedWriteLimit, c.Recipe.StarRecipe)
			recipes.POST("/:id/fork", authenticatedWriteLimit, c.Recipe.ForkRecipe)
			recipes.POST("/revisions/:revisionID/fork", authenticatedWriteLimit, c.Recipe.ForkRecipeRevision)
			// Media initialization is its own abuse target because it creates a signed upload URL.
			recipes.POST("/:id/media", mediaInitLimit, c.Recipe.CreateRecipeMedia)
			recipes.DELETE("/:id", authenticatedWriteLimit, c.Recipe.DeleteRecipe)
		}

		// authenticated routes below this line
		EnforceAuth(cfg.Auth0, api)

		// user
		user := api.Group("/user")
		{
			user.HEAD("/", c.User.HeadAuthenticatedUser)
			user.GET("/", c.User.GetAuthenticatedUser)
			user.POST("/", userCreateLimit, c.User.CreateUser)
			user.PATCH("/", authenticatedWriteLimit, c.User.UpdateUser)
			user.DELETE("/events/:id", authenticatedWriteLimit, c.User.RemoveUserEvent)
		}

		// users
		users := api.Group("/users")
		{
			users.DELETE("/:id", authenticatedWriteLimit, middleware.Authorize("/users/*", "delete"), c.User.DeleteUser)
			users.POST("/username", usernameCheckLimit, c.User.TestUsername)
			users.POST("/", userCreateLimit, c.User.CreateUser)
			// users.GET("profile", c.User.GetAuthenticatedUser)
			// users.GET("exist", c.User.GetAuthenticatedUserExist)
			users.GET("", middleware.Authorize("/users/*", "list"), c.User.GetUsers)
			users.GET(":id", c.User.GetUser)
			users.PATCH(":id", authenticatedWriteLimit, c.User.UpdateUser)

			// users.GET("", middleware.Authorize("/users/*", "list"), c.User.GetUsers)
			// users.DELETE(":id", middleware.Authorize("/users/*", "delete"), c.User.DeleteUser)
			// users.POST("username", c.User.TestUsername)
			// users.GET(":id", c.User.GetUser)
		}

		// admin
		admin := api.Group("/_admin")
		{
			admin.POST("/recipes", middleware.Authorize("/recipes/*", "create"), c.Recipe.BotCreateRecipe)
			admin.GET("/recipes/:id/media", middleware.Authorize("/recipes/*", "list"), c.Recipe.GetAdminRecipeMedias)
			admin.POST("/init-search-collections", middleware.Authorize("/search/*", "create"), c.Search.CreateSearchRecipeCollection)
		}
	}
}

// @title 4ks API
// @version 2.0
// @description This is the 4ks api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// context
	var ctx = context.Background()
	configureLogging()

	cfg, err := utils.LoadRuntimeConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid runtime configuration")
	}

	sysFlags := cfg.SystemFlags()

	// reserved words
	reservedWords, err := ReadWordsFromFile(cfg.Routes.ReservedWordsFile)
	if err != nil {
		panic(err)
	}

	// jaeger
	if cfg.Tracing.Enabled {
		log.Info().Bool("enabled", cfg.Tracing.Enabled).Str("exporter", cfg.Tracing.ExporterType).Msg("Tracing")
		tp := tracing.InitTracerProvider(tracing.Config{
			ExporterType:       cfg.Tracing.ExporterType,
			JaegerEndpoint:     cfg.Tracing.JaegerEndpoint,
			GoogleCloudProject: cfg.Tracing.GoogleCloudProject,
			ServiceName:        cfg.Tracing.ServiceName,
		})
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Error().Err(err).Msg("Error shutting down tracer provider")
			}
		}()
	}

	// storage
	store, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create storage client")
	}
	defer store.Close()

	// firestore
	if cfg.Firestore.EmulatorHost != "" {
		log.Info().Msgf("Using Firestore Emulator: '%s'", cfg.Firestore.EmulatorHost)
	}
	var fire, _ = firestore.NewClient(ctx, cfg.Firestore.ProjectID)
	defer fire.Close()

	// pubsub
	if cfg.PubSub.EmulatorHost != "" {
		log.Info().Msgf("Using PubSub Emulator: '%s'", cfg.PubSub.EmulatorHost)
	}
	// create pubsub client
	psub, err := pubsub.NewClient(ctx, cfg.PubSub.ProjectID)
	if err != nil {
		log.Fatal().Err(err).Str("project", cfg.PubSub.ProjectID).Msg("failed to create pubsub client")
	}
	defer psub.Close()
	log.Debug().Str("project", cfg.PubSub.ProjectID).Msg("pubsub client created")

	// pubsub options
	reso := fetcherService.FetcherOpts{
		ProjectID: cfg.PubSub.ProjectID,
		TopicID:   cfg.PubSub.FetcherTopic,
	}

	// typesense
	ts := typesense.NewClient(typesense.WithServer(cfg.Typesense.URL), typesense.WithAPIKey(cfg.Typesense.APIKey))

	// services
	static := staticService.New(store, cfg.Static.FallbackURL, cfg.Static.Bucket, cfg.Static.FallbackPrefix)
	search := searchService.New(ts)

	v := validator.New()
	user := userService.New(&sysFlags, fire, v, &reservedWords)
	recipe := recipeService.New(&sysFlags, store, fire, v, &recipeService.RecipeServiceConfig{
		DistributionBucket: cfg.Recipe.DistributionBucket,
		UploadableBucket:   cfg.Recipe.UploadableBucket,
		ServiceAccountName: cfg.Recipe.ServiceAccountName,
		ImageURL:           cfg.Recipe.ImageURL,
	})
	fetcher := fetcherService.New(ctx, &sysFlags, psub, reso, user, recipe, search, static)

	// controllers
	c := &Controllers{
		System: controllers.NewSystemController(
			getAPIVersion(cfg.Routes.VersionFilePath),
			controllers.SystemControllerDeps{
				DB:        controllers.NewFirestoreProber(fire),
				Search:    controllers.NewTypesenseProber(ts),
				Messaging: controllers.NewPubSubProber(psub, reso.TopicID),
				Storage:   controllers.NewStorageProber(store, cfg.Recipe.DistributionBucket),
			},
		),
		Recipe: controllers.NewRecipeController(user, recipe, search, static, fetcher),
		User:   controllers.NewUserController(user),
		Search: controllers.NewSearchController(search),
	}

	// gin and middleware
	router := gin.New()
	router.Use(gin.Recovery())
	// Trust forwarding headers only from the explicitly configured proxy layer.
	if err := router.SetTrustedProxies(cfg.HTTP.Proxy.TrustedCIDRs); err != nil {
		log.Fatal().Err(err).Msg("failed to configure trusted proxies")
	}
	router.Use(middleware.ErrorHandler)
	router.Use(middleware.CorsMiddleware(cfg.HTTP.CORS))
	if sysFlags.Debug {
		router.Use(middleware.DefaultStructuredLogger())
	}

	// metrics
	prom := ginprometheus.NewPrometheus("gin")
	prom.Use(router)

	AppendRoutes(cfg, router, c)

	addr := "0.0.0.0:" + cfg.Routes.Port
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Error starting http server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down server")
}

// ReadWordsFromFile reads words from a file
func ReadWordsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}
