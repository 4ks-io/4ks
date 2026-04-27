package middleware

import (
	"4ks/apps/api/utils"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope       string `json:"scope"`
	Email       string `json:"https://4ks.io/email"`
	ID          string `json:"https://4ks.io/id"`
	Timezone    string `json:"https://4ks.io/timeZone"`
	CountryCode string `json:"https://4ks.io/countryCode"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(_ context.Context) error {
	return nil
}

// ExtractClaimsFromRequest extracts the claims from the request
func ExtractClaimsFromRequest(request *http.Request) validator.ValidatedClaims {
	return *request.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
}

// ExtractCustomClaimsFromClaims extracts the custom claims from the claims
func ExtractCustomClaimsFromClaims(claims *validator.ValidatedClaims) CustomClaims {
	return *claims.CustomClaims.(*CustomClaims)
}

func appendJWTClaims(ctx *gin.Context, claims *validator.ValidatedClaims) {
	customClaims := ExtractCustomClaimsFromClaims(claims)
	SetAuthIdentity(ctx, AuthIdentity{
		AuthID:   claims.RegisteredClaims.Subject,
		AuthType: AuthTypeJWT,
		UserID:   customClaims.ID,
		Email:    customClaims.Email,
	})
}

// AppendCustomClaims is a middleware that will append custom claims to the context.
func AppendCustomClaims() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ExtractClaimsFromRequest(ctx.Request)
		appendJWTClaims(ctx, &claims)
		ctx.Next()
	}
}

func buildJWTValidator(cfg utils.Auth0Config) *validator.Validator {
	issuerURL, err := url.Parse("https://" + cfg.Domain + "/")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse the issuer url")
		panic(err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{cfg.Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set up the jwt validator")
	}

	return jwtValidator
}

func writeJWTUnauthorized(w http.ResponseWriter, err error) {
	log.Error().Err(err).Msg("failed to validate JWT")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"message":"Failed to validate JWT."}`))
}

func extractBearerToken(value string) (string, error) {
	const prefix = "Bearer "
	if value == "" || !strings.HasPrefix(value, prefix) {
		return "", http.ErrNoCookie
	}

	token := strings.TrimSpace(strings.TrimPrefix(value, prefix))
	if token == "" {
		return "", http.ErrNoCookie
	}

	return token, nil
}

// RequireJWT validates a bearer JWT and stores the resolved identity in Gin context.
func RequireJWT(cfg utils.Auth0Config) gin.HandlerFunc {
	jwtValidator := buildJWTValidator(cfg)

	return func(ctx *gin.Context) {
		token, err := extractBearerToken(ctx.GetHeader("Authorization"))
		if err != nil {
			writeJWTUnauthorized(ctx.Writer, err)
			ctx.Abort()
			return
		}

		claimsAny, err := jwtValidator.ValidateToken(ctx.Request.Context(), token)
		if err != nil {
			writeJWTUnauthorized(ctx.Writer, err)
			ctx.Abort()
			return
		}

		claims, ok := claimsAny.(*validator.ValidatedClaims)
		if !ok {
			writeJWTUnauthorized(ctx.Writer, errors.New("unexpected validated claims type"))
			ctx.Abort()
			return
		}

		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), jwtmiddleware.ContextKey{}, claims))
		appendJWTClaims(ctx, claims)
		ctx.Next()
	}
}

// EnforceJWT is a net/http middleware wrapper retained for compatibility with existing callers.
func EnforceJWT(cfg utils.Auth0Config) func(next http.Handler) http.Handler {
	jwtValidator := buildJWTValidator(cfg)

	errorHandler := func(w http.ResponseWriter, _ *http.Request, err error) {
		writeJWTUnauthorized(w, err)
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}
