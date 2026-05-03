package rest

import (
	middleware "4ks/apps/api/middleware"
	kitchenPassService "4ks/apps/api/services/kitchenpass"
	utils "4ks/apps/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// EnforceAuth enforces authentication.
func EnforceAuth(authConfig utils.Auth0Config, r *gin.RouterGroup) {
	r.Use(middleware.RequireJWT(authConfig))
}

// AppendRoutes appends routes to the router.
func AppendRoutes(cfg *utils.RuntimeConfig, r *gin.Engine, c *Controllers, kitchenPass kitchenPassService.Service) {
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
		// throttle unrelated users behind the same egress. PAT traffic uses the
		// token digest instead so AI traffic does not share JWT buckets.
		Rules: []middleware.RateLimitRule{
			middleware.QPSRule(2),
			middleware.QPMRule(30),
		},
		KeyFunc: middleware.RateLimitByAuthOrIP,
	})
	authenticatedRecipeSearchLimit := middleware.NewRateLimitMiddleware(rateLimitStore, middleware.RateLimitPolicy{
		Name: "authenticated-recipe-search",
		Rules: []middleware.RateLimitRule{
			middleware.QPSRule(5),
			middleware.QPMRule(120),
		},
		KeyFunc: middleware.RateLimitByAuthOrIP,
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
			recipes.GET("", publicReadLimit, c.Recipe.GetRecipes)
			recipes.GET("/search", middleware.RequireJWTOrPAT(cfg.Auth0, kitchenPass), authenticatedRecipeSearchLimit, c.Recipe.SearchRecipes)
			recipes.GET("/:id", publicReadLimit, c.Recipe.GetRecipe)
			recipes.GET("/:id/forks", publicReadLimit, c.Recipe.GetRecipeForks)
			recipes.GET("/:id/revisions", publicReadLimit, c.Recipe.GetRecipeRevisions)
			recipes.GET("/revisions/:revisionID", publicReadLimit, c.Recipe.GetRecipeRevision)
			recipes.GET("/:id/media", publicReadLimit, c.Recipe.GetRecipeMedia)
			recipes.GET("/author/:username", publicReadLimit, c.Recipe.GetRecipesByUsername)

			recipesJWTOrPAT := recipes.Group("")
			recipesJWTOrPAT.Use(middleware.RequireJWTOrPAT(cfg.Auth0, kitchenPass))
			recipesJWTOrPAT.POST("", authenticatedWriteLimit, c.Recipe.CreateRecipe)
			recipesJWTOrPAT.PATCH("/:id", authenticatedWriteLimit, c.Recipe.UpdateRecipe)
			recipesJWTOrPAT.POST("/:id/fork", authenticatedWriteLimit, c.Recipe.ForkRecipe)
			recipesJWTOrPAT.POST("/revisions/:revisionID/fork", authenticatedWriteLimit, c.Recipe.ForkRecipeRevision)

			recipesJWTOnly := recipes.Group("")
			EnforceAuth(cfg.Auth0, recipesJWTOnly)
			recipesJWTOnly.POST("/fetch", recipeFetchLimit, c.Recipe.FetchRecipe)
			recipesJWTOnly.POST("/:id/star", authenticatedWriteLimit, c.Recipe.StarRecipe)
			// Media initialization is its own abuse target because it creates a signed upload URL.
			recipesJWTOnly.POST("/:id/media", mediaInitLimit, c.Recipe.CreateRecipeMedia)
			recipesJWTOnly.DELETE("/:id", authenticatedWriteLimit, c.Recipe.DeleteRecipe)
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
			user.GET("/kitchen-pass", c.User.GetKitchenPass)
			user.POST("/kitchen-pass", authenticatedWriteLimit, c.User.CreateKitchenPass)
			user.DELETE("/kitchen-pass", c.User.DeleteKitchenPass)
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
