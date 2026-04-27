package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ErrorHandler handles logging errors
func ErrorHandler(ctx *gin.Context) {
	ctx.Next()

	// don't log some system routes
	path := ctx.Request.URL.Path
	if path == "/api/healthcheck" || path == "/api/ready" {
		return
	}

	if ctx.Errors != nil {
		for _, err := range ctx.Errors {
			log.Error().
				Err(err).
				Str("path", ctx.Request.URL.Path).
				Str("authType", ctx.GetString("authType")).
				Str("patPreview", ctx.GetString("patPreview"))
		}
		// frequently causes json responses to be followed by a null
		// ctx.JSON(-1, ctx.Errors)
	}
}
