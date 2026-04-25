package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readinessTimeout = 3 * time.Second

// Prober checks whether a single dependency is reachable.
type Prober interface {
	Name() string
	Probe(ctx context.Context) error
}

// SystemControllerDeps holds the dependency probers used by the rich healthcheck.
type SystemControllerDeps struct {
	DB        Prober // firestore
	Search    Prober // typesense
	Messaging Prober // pubsub
	Storage   Prober // gcs
}

// SystemController is the interface for the systemController
type SystemController interface {
	CheckReadiness(*gin.Context)
	Healthcheck(*gin.Context)
}

type systemController struct {
	version string
	deps    SystemControllerDeps
}

// NewSystemController creates a new systemController.
func NewSystemController(version string, deps SystemControllerDeps) SystemController {
	return &systemController{version: version, deps: deps}
}

// CheckReadiness godoc
//
//	@Summary 		 Checks Readiness
//	@Description Shallow liveness probe. Always returns 200; use /api/healthcheck for dependency status.
//	@Tags 				System
//	@Produce 		  json
//	@Success 		  200 		 {object} map[string]string
//	@Router       /api/ready [get]
func (s *systemController) CheckReadiness(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "Ok"})
}

// Healthcheck godoc
//
//	@Summary 		 Healthcheck
//	@Description Reports version and downstream dependency status. Development only.
//	@Tags 			 System
//	@Produce 		 json
//	@Success 		 200 		 {object} map[string]interface{}
//	@Router      /api/healthcheck [get]
func (s *systemController) Healthcheck(ctx *gin.Context) {
	probe, cancel := context.WithTimeout(ctx.Request.Context(), readinessTimeout)
	defer cancel()

	status := func(p Prober) string {
		if err := p.Probe(probe); err != nil {
			return err.Error()
		}
		return "ok"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"version": s.version,
		"database": gin.H{
			"provider": "firestore",
			"status":   status(s.deps.DB),
		},
		"storage": gin.H{
			"provider": "gcs",
			"status":   status(s.deps.Storage),
		},
		"services": gin.H{
			"search":    status(s.deps.Search),
			"messaging": status(s.deps.Messaging),
		},
	})
}
