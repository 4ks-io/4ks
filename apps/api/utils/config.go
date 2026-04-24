// Package utils provides utility functions for the application
package utils

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// SystemFlags is a struct for system flags
type SystemFlags struct {
	Debug         bool
	Development   bool
	JaegerEnabled bool
}

// CORSConfig contains the API's browser-facing CORS behavior.
type CORSConfig struct {
	// AllowedOrigins is an exact-match allowlist. Wildcards are rejected
	// because this API also allows credentialed browser requests.
	AllowedOrigins []string
	// AllowedMethods is returned on preflight responses.
	AllowedMethods []string
	// AllowedHeaders is returned on preflight responses.
	AllowedHeaders []string
	// ExposedHeaders enumerates which response headers browsers may read.
	ExposedHeaders []string
	// AllowCredentials enables cookies or authorization-bearing browser requests.
	AllowCredentials bool
	// MaxAge controls how long a browser may cache the preflight result.
	MaxAge time.Duration
}

// ProxyConfig contains the upstream proxies we trust to supply forwarding headers.
type ProxyConfig struct {
	TrustedCIDRs []string
}

// HTTPSecurityConfig groups the API's transport-layer security settings.
type HTTPSecurityConfig struct {
	CORS  CORSConfig
	Proxy ProxyConfig
}

// GetStrEnvVar returns a string from an environment variable.
func GetStrEnvVar(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetBoolEnv returns a bool from an environment variable.
func GetBoolEnv(key string, fallback bool) bool {
	val := GetStrEnvVar(key, strconv.FormatBool(fallback))
	ret, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return ret
}

// GetEnvVarOrPanic returns an environment variable value or exits.
func GetEnvVarOrPanic(n string) string {
	v, ok := os.LookupEnv(n)
	if !ok || v == "" {
		panic(fmt.Sprintf("env var %s required", n))
	}
	return v
}

// GetCSVEnvVar returns a CSV environment variable as a trimmed string slice.
func GetCSVEnvVar(key string, fallback []string) []string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return append([]string(nil), fallback...)
	}

	// Empty segments are ignored so trailing commas do not silently create
	// invalid blank entries.
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}

	return values
}

// LoadHTTPSecurityConfig loads and validates HTTP-facing security settings.
func LoadHTTPSecurityConfig() (*HTTPSecurityConfig, error) {
	cfg := &HTTPSecurityConfig{
		CORS: CORSConfig{
			// These defaults reflect the API surface that browsers are expected to use.
			AllowedOrigins: GetCSVEnvVar("CORS_ALLOWED_ORIGINS", nil),
			AllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "HEAD", "OPTIONS"},
			AllowedHeaders: []string{
				"Origin",
				"Content-Type",
				"Content-Length",
				"Accept-Encoding",
				"X-CSRF-Token",
				"Authorization",
			},
			ExposedHeaders:   []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           24 * time.Hour,
		},
		Proxy: ProxyConfig{
			TrustedCIDRs: GetCSVEnvVar("TRUSTED_PROXY_CIDRS", nil),
		},
	}

	if len(cfg.CORS.AllowedOrigins) == 0 {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS must contain at least one origin")
	}

	for _, origin := range cfg.CORS.AllowedOrigins {
		// Browsers reject wildcard origins with credentials, so fail here instead
		// of serving a broken runtime configuration.
		if origin == "*" {
			return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS cannot contain * when credentials are enabled")
		}

		parsed, err := url.Parse(origin)
		if err != nil {
			return nil, fmt.Errorf("invalid CORS origin %q: %w", origin, err)
		}
		if parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("invalid CORS origin %q: expected scheme://host with no path, query, or fragment", origin)
		}
	}

	// Trusted proxies are required because Gin only uses forwarded headers when
	// the remote peer matches this allowlist.
	if len(cfg.Proxy.TrustedCIDRs) == 0 {
		return nil, fmt.Errorf("TRUSTED_PROXY_CIDRS must contain at least one CIDR")
	}

	for _, cidr := range cfg.Proxy.TrustedCIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", cidr, err)
		}
	}

	return cfg, nil
}
