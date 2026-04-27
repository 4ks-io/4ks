package middleware

import (
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	"4ks/apps/api/utils"
	"context"
	"errors"
	"fmt"
	"net/http"

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

func applyPATIdentity(ctx *gin.Context, userID string) {
	SetAuthIdentity(ctx, AuthIdentity{
		AuthID:   fmt.Sprintf("kitchen-pass:%s", userID),
		AuthType: AuthTypePAT,
		UserID:   userID,
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

		applyPATIdentity(ctx, record.UserID)
		ctx.Next()
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

			applyPATIdentity(ctx, record.UserID)
			ctx.Next()
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
