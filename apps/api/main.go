// package main is the entrypoint for the api
package main

import (
	"context"
	"os/signal"
	"syscall"

	"4ks/apps/api/mcp"
	"4ks/apps/api/rest"
	utils "4ks/apps/api/utils"
	tracing "4ks/libs/go/tracer"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// @title 4ks API
// @version 2.0
// @description 4ks recipe API.
// @description
// @description Authentication uses `Authorization: Bearer <token>`.
// @description Most authenticated routes expect an Auth0 JWT.
// @description Approved recipe routes documented as AI Kitchen Pass compatible also accept a Kitchen Pass bearer token.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	configureLogging()

	cfg, err := utils.LoadRuntimeConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid runtime configuration")
	}

	reservedWords, err := ReadWordsFromFile(cfg.Routes.ReservedWordsFile)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read reserved words")
	}

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

	wiring, err := buildRuntimeWiring(rootCtx, cfg, reservedWords)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to build runtime wiring")
	}
	defer wiring.cleanup()

	restSrv, err := rest.New(cfg, wiring.services, wiring.restDeps)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create rest server")
	}
	mcpSrv := mcp.New(cfg, wiring.services)

	g, ctx := errgroup.WithContext(rootCtx)
	g.Go(func() error { return restSrv.Start(ctx) })
	g.Go(func() error { return mcpSrv.Start(ctx) })
	if err := g.Wait(); err != nil {
		log.Fatal().Err(err).Msg("api server stopped with error")
	}
}
