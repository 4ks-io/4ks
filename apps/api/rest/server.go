package rest

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"4ks/apps/api/app"
	controllers "4ks/apps/api/controllers"
	_ "4ks/apps/api/docs"
	middleware "4ks/apps/api/middleware"
	utils "4ks/apps/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

const shutdownTimeout = 30 * time.Second

// Deps contains REST dependencies that are not part of the shared service bundle.
type Deps struct {
	Version string
	System  controllers.SystemControllerDeps
}

// Controllers contains the controllers used by REST routes.
type Controllers struct {
	User   controllers.UserController
	Recipe controllers.RecipeController
	Search controllers.SearchController
	System controllers.SystemController
}

// Server owns the REST router and HTTP server lifecycle.
type Server struct {
	httpServer *http.Server
}

// New wires the Gin router, middleware, routes, and http.Server.
func New(cfg *utils.RuntimeConfig, svc app.Services, deps Deps) (*Server, error) {
	sysFlags := cfg.SystemFlags()

	c := &Controllers{
		System: controllers.NewSystemController(deps.Version, deps.System),
		Recipe: controllers.NewRecipeController(svc.User, svc.Recipe, svc.Search, svc.Static, svc.Fetcher),
		User:   controllers.NewUserController(svc.User, svc.KitchenPass),
		Search: controllers.NewSearchController(svc.Search),
	}

	router := gin.New()
	router.Use(gin.Recovery())
	// Trust forwarding headers only from the explicitly configured proxy layer.
	if err := router.SetTrustedProxies(cfg.HTTP.Proxy.TrustedCIDRs); err != nil {
		return nil, err
	}
	router.Use(middleware.ErrorHandler)
	router.Use(middleware.CorsMiddleware(cfg.HTTP.CORS))
	if sysFlags.Debug {
		router.Use(middleware.DefaultStructuredLogger())
	}

	prom := ginprometheus.NewPrometheus("gin")
	prom.Use(router)

	AppendRoutes(cfg, router, c, svc.KitchenPass)

	return &Server{
		httpServer: &http.Server{
			Addr:    "0.0.0.0:" + cfg.Routes.Port,
			Handler: router,
		},
	}, nil
}

// Start begins listening and blocks until ctx is cancelled, then gracefully shuts down.
func (s *Server) Start(ctx context.Context) error {
	s.httpServer.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

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

	log.Info().Msg("rest server stopped")
	return nil
}
