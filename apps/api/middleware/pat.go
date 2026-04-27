package middleware

import (
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	"4ks/apps/api/utils"
	"4ks/libs/go/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func writePATUnauthorized(w http.ResponseWriter, err error) {
	log.Error().Err(err).Msg("failed to validate kitchen pass")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"message":"Failed to validate Kitchen Pass."}`))
}

func applyPATIdentity(ctx *gin.Context, record *models.PersonalAccessToken) {
	SetAuthIdentity(ctx, AuthIdentity{
		AuthID:     fmt.Sprintf("kitchen-pass:%s", record.UserID),
		AuthType:   AuthTypePAT,
		UserID:     record.UserID,
		PATDigest:  record.TokenDigest,
		PATPreview: record.TokenPreview,
	})
}

func validateJWTToken(ctx *gin.Context, jwtValidator *validator.Validator, token string) error {
	claimsAny, err := jwtValidator.ValidateToken(ctx.Request.Context(), token)
	if err != nil {
		return err
	}

	claims, ok := claimsAny.(*validator.ValidatedClaims)
	if !ok {
		return errors.New("unexpected validated claims type")
	}

	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), jwtmiddleware.ContextKey{}, claims))
	appendJWTClaims(ctx, claims)
	return nil
}

// RequirePAT validates a Kitchen Pass bearer token and stores its owner identity in Gin context.
func RequirePAT(service kitchenpasssvc.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := extractBearerToken(ctx.GetHeader("Authorization"))
		if err != nil || !kitchenpasssvc.IsKitchenPassToken(token) {
			writePATUnauthorized(ctx.Writer, kitchenpasssvc.ErrInvalidKitchenPassToken)
			ctx.Abort()
			return
		}

		record, err := service.ValidateToken(ctx.Request.Context(), token)
		if err != nil {
			writePATUnauthorized(ctx.Writer, err)
			ctx.Abort()
			return
		}

		applyPATIdentity(ctx, record)
		ctx.Next()

		recordKitchenPassUsage(ctx, service)
	}
}

// RequireJWTOrPAT validates either a Kitchen Pass token or an Auth0 JWT.
func RequireJWTOrPAT(cfg utils.Auth0Config, service kitchenpasssvc.Service) gin.HandlerFunc {
	jwtValidator := buildJWTValidator(cfg)

	return func(ctx *gin.Context) {
		token, err := extractBearerToken(ctx.GetHeader("Authorization"))
		if err != nil {
			writeJWTUnauthorized(ctx.Writer, err)
			ctx.Abort()
			return
		}

		if kitchenpasssvc.IsKitchenPassToken(token) {
			record, err := service.ValidateToken(ctx.Request.Context(), token)
			if err != nil {
				writePATUnauthorized(ctx.Writer, err)
				ctx.Abort()
				return
			}

			applyPATIdentity(ctx, record)
			ctx.Next()
			recordKitchenPassUsage(ctx, service)
			return
		}

		if err := validateJWTToken(ctx, jwtValidator, token); err != nil {
			writeJWTUnauthorized(ctx.Writer, err)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func recordKitchenPassUsage(ctx *gin.Context, service kitchenpasssvc.Service) {
	if ctx.GetString("authType") != AuthTypePAT {
		return
	}

	tokenDigest := ctx.GetString("patDigest")
	if tokenDigest == "" {
		return
	}

	action := kitchenPassActionLabel(ctx)
	if action == "" {
		return
	}

	if err := service.RecordUsage(ctx.Request.Context(), tokenDigest, action); err != nil {
		log.Warn().
			Err(err).
			Str("auth_type", AuthTypePAT).
			Str("pat_preview", ctx.GetString("patPreview")).
			Str("action", action).
			Msg("failed to record kitchen pass usage")
	}
}

// kitchenPassActionLabel keeps the user-visible activity labels stable.
func kitchenPassActionLabel(ctx *gin.Context) string {
	path := normalizeRoutePath(ctx.FullPath())

	switch {
	case ctx.Request.Method == http.MethodGet && path == "/api/recipes/search":
		return "searched recipes"
	case ctx.Request.Method == http.MethodPost && path == "/api/recipes":
		return "created recipe"
	case ctx.Request.Method == http.MethodPatch && path == "/api/recipes/:id":
		return "updated recipe"
	case ctx.Request.Method == http.MethodPost && path == "/api/recipes/:id/fork":
		return "forked recipe"
	case ctx.Request.Method == http.MethodPost && path == "/api/recipes/revisions/:revisionID/fork":
		return "forked recipe revision"
	case ctx.Request.Method == http.MethodGet && path == "/api/recipes/:id/forks":
		return "viewed recipe forks"
	case ctx.Request.Method == http.MethodGet && path == "/api/recipes/:id/revisions":
		return "viewed recipe revisions"
	case ctx.Request.Method == http.MethodGet && path == "/api/recipes/revisions/:revisionID":
		return "viewed recipe revision"
	}

	if path == "" {
		return ""
	}

	return fmt.Sprintf("used %s %s", strings.ToUpper(ctx.Request.Method), path)
}

func normalizeRoutePath(path string) string {
	if path == "" || path == "/" {
		return path
	}

	return strings.TrimSuffix(path, "/")
}
