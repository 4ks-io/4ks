package utils

import (
	"testing"
)

// ---------------------------------------------------------------------------
// LoadHTTPSecurityConfig
// ---------------------------------------------------------------------------

func TestLoadHTTPSecurityConfig(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://www.4ks.io, https://dev.4ks.io")
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8,127.0.0.1/32")

	cfg, err := LoadHTTPSecurityConfig()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Fatalf("expected 2 allowed origins, got %d", len(cfg.CORS.AllowedOrigins))
	}
	if len(cfg.Proxy.TrustedCIDRs) != 2 {
		t.Fatalf("expected 2 trusted proxy CIDRs, got %d", len(cfg.Proxy.TrustedCIDRs))
	}
}

func TestLoadHTTPSecurityConfigRejectsWildcardOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")

	if _, err := LoadHTTPSecurityConfig(); err == nil {
		t.Fatal("expected wildcard origin to be rejected")
	}
}

func TestLoadHTTPSecurityConfigRejectsInvalidProxyCIDR(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://www.4ks.io")
	t.Setenv("TRUSTED_PROXY_CIDRS", "not-a-cidr")

	if _, err := LoadHTTPSecurityConfig(); err == nil {
		t.Fatal("expected invalid proxy cidr to be rejected")
	}
}

func TestMinimalRuntimeConfigProvidesValidTestDefaults(t *testing.T) {
	cfg := MinimalRuntimeConfig()

	if cfg.Auth0.Domain == "" || cfg.Fetcher.SharedSecret == "" || cfg.Routes.Port == "" {
		t.Fatalf("expected minimal config to populate required fields, got %+v", cfg)
	}
}

func TestLoadRuntimeConfig(t *testing.T) {
	t.Setenv("GIN_MODE", "debug")
	t.Setenv("IO_4KS_DEVELOPMENT", "true")
	t.Setenv("SWAGGER_ENABLED", "true")
	t.Setenv("SWAGGER_URL_PREFIX", "/api")
	t.Setenv("PORT", "9000")
	t.Setenv("RESERVED_WORDS_FILE", "/tmp/words")
	t.Setenv("VERSION_FILE_PATH", "/tmp/version")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://www.4ks.io, https://dev.4ks.io")
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8,127.0.0.1/32")
	t.Setenv("AUTH0_DOMAIN", "example.auth0.com")
	t.Setenv("AUTH0_AUDIENCE", "test-audience")
	t.Setenv("APP_BASE_URL", "https://www.4ks.io")
	t.Setenv("TYPESENSE_URL", "http://typesense:8108")
	t.Setenv("TYPESENSE_API_KEY", "typesense-key")
	t.Setenv("MEDIA_FALLBACK_URL", "https://media.4ks.io/fallback.jpg")
	t.Setenv("STATIC_MEDIA_BUCKET", "static-media")
	t.Setenv("STATIC_MEDIA_FALLBACK_PREFIX", "fallback")
	t.Setenv("DISTRIBUTION_BUCKET", "distribution")
	t.Setenv("UPLOADABLE_BUCKET", "uploadable")
	t.Setenv("SERVICE_ACCOUNT_EMAIL", "svc@example.com")
	t.Setenv("MEDIA_IMAGE_URL", "https://media.4ks.io")
	t.Setenv("FIRESTORE_PROJECT_ID", "firestore-project")
	t.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
	t.Setenv("PUBSUB_PROJECT_ID", "pubsub-project")
	t.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	t.Setenv("FETCHER_TOPIC_ID", "fetcher-custom")
	t.Setenv("API_FETCHER_PSK", "01234567890123456789012345678901")
	t.Setenv("PAT_DIGEST_SECRET", "01234567890123456789012345678901")
	t.Setenv("PAT_ENCRYPTION_SECRET", "abcdefghijklmnopqrstuvwxyz012345")
	t.Setenv("JAEGER_ENABLED", "true")
	t.Setenv("EXPORTER_TYPE", "jaeger")
	t.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", "http://jaeger:14268/api/traces")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "gcp-project")
	t.Setenv("OTEL_SERVICE_NAME", "4ks-api-test")

	cfg, err := LoadRuntimeConfig()
	if err != nil {
		t.Fatalf("expected runtime config to load, got error: %v", err)
	}

	if !cfg.System.Development || cfg.System.GinMode != "debug" {
		t.Fatalf("unexpected system config: %+v", cfg.System)
	}
	if !cfg.Features.SwaggerEnabled || cfg.Routes.SwaggerURLPrefix != "/api" || cfg.Routes.Port != "9000" {
		t.Fatalf("unexpected route config: %+v %+v", cfg.Features, cfg.Routes)
	}
	if cfg.KitchenPass.BaseURL != "https://www.4ks.io" {
		t.Fatalf("unexpected kitchen pass config: %+v", cfg.KitchenPass)
	}
	if cfg.PubSub.FetcherTopic != "fetcher-custom" || !cfg.Tracing.Enabled || cfg.Tracing.ExporterType != "JAEGER" {
		t.Fatalf("unexpected typed runtime config: %+v %+v", cfg.PubSub, cfg.Tracing)
	}
}
