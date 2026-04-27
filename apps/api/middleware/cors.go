package middleware

import (
	"4ks/apps/api/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsMiddleware adds config-driven CORS headers to the response.
func CorsMiddleware(cfg utils.CORSConfig) gin.HandlerFunc {
	// Exact-match lookups keep behavior predictable across environments and
	// avoid accidental subdomain wildcarding.
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}

	// Precompute static header payloads once when the middleware is built.
	allowMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowHeaders := strings.Join(cfg.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(cfg.ExposedHeaders, ", ")
	maxAge := strconv.FormatInt(int64(cfg.MaxAge.Seconds()), 10)

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/ai/") {
			c.Writer.Header().Add("Vary", "Access-Control-Request-Method")
			c.Writer.Header().Add("Vary", "Access-Control-Request-Headers")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.Header("Access-Control-Max-Age", maxAge)

			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}

			c.Next()
			return
		}

		// Responses vary by origin and by preflight request headers, so advertise
		// that variance explicitly for downstream caches.
		c.Writer.Header().Add("Vary", "Origin")
		c.Writer.Header().Add("Vary", "Access-Control-Request-Method")
		c.Writer.Header().Add("Vary", "Access-Control-Request-Headers")

		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				// Reflect only allowlisted origins so credentialed requests remain valid.
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", allowMethods)
				c.Header("Access-Control-Allow-Headers", allowHeaders)
				c.Header("Access-Control-Expose-Headers", exposeHeaders)
				c.Header("Access-Control-Max-Age", maxAge)
				if cfg.AllowCredentials {
					c.Header("Access-Control-Allow-Credentials", "true")
				}
			}
		}

		if c.Request.Method == http.MethodOptions {
			// CORS preflights do not need to hit application handlers once the
			// browser has the policy response.
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
