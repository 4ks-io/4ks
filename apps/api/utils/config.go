// Package utils provides utility functions for the application
package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/configor"
)

// SystemFlags is a struct for system flags
type SystemFlags struct {
	Debug         bool
	Development   bool
	JaegerEnabled bool
}

// SystemConfig contains global application modes.
type SystemConfig struct {
	GinMode     string `validate:"required,oneof=debug release test"`
	Development bool
}

// FeatureFlags contains operator-controlled feature switches.
type FeatureFlags struct {
	SwaggerEnabled bool
}

// RouteConfig contains API listener and file path settings.
type RouteConfig struct {
	Port              string `validate:"required,numeric"`
	SwaggerURLPrefix  string
	ReservedWordsFile string `validate:"required"`
	VersionFilePath   string
}

// CORSConfig contains the API's browser-facing CORS behavior.
type CORSConfig struct {
	// AllowedOrigins is an exact-match allowlist. Wildcards are rejected
	// because this API also allows credentialed browser requests.
	AllowedOrigins []string `validate:"required,min=1,dive,cors_origin"`
	// AllowedMethods is returned on preflight responses.
	AllowedMethods []string `validate:"required,min=1,dive,required"`
	// AllowedHeaders is returned on preflight responses.
	AllowedHeaders []string `validate:"required,min=1,dive,required"`
	// ExposedHeaders enumerates which response headers browsers may read.
	ExposedHeaders []string `validate:"required,min=1,dive,required"`
	// AllowCredentials enables cookies or authorization-bearing browser requests.
	AllowCredentials bool
	// MaxAge controls how long a browser may cache the preflight result.
	MaxAge time.Duration `validate:"gt=0"`
}

// ProxyConfig contains the upstream proxies we trust to supply forwarding headers.
type ProxyConfig struct {
	TrustedCIDRs []string `validate:"required,min=1,dive,cidr"`
}

// HTTPSecurityConfig groups the API's transport-layer security settings.
type HTTPSecurityConfig struct {
	CORS  CORSConfig
	Proxy ProxyConfig
}

// Auth0Config contains JWT validation settings.
type Auth0Config struct {
	Domain   string `validate:"required,hostname_rfc1123"`
	Audience string `validate:"required"`
}

// KitchenPassConfig contains AI Kitchen Pass secrets and URL settings.
type KitchenPassConfig struct {
	BaseURL          string `validate:"required,http_url"`
	DigestSecret     string `validate:"required,min=32"`
	EncryptionSecret string `validate:"required,min=32"`
}

// TypesenseConfig contains Typesense connectivity settings.
type TypesenseConfig struct {
	URL    string `validate:"required,http_url"`
	APIKey string `validate:"required"`
}

// StaticMediaConfig contains static media fallback settings.
type StaticMediaConfig struct {
	FallbackURL    string `validate:"required,http_url"`
	Bucket         string `validate:"required"`
	FallbackPrefix string `validate:"required"`
}

// RecipeRuntimeConfig contains recipe media and IAM settings.
type RecipeRuntimeConfig struct {
	DistributionBucket string `validate:"required"`
	UploadableBucket   string `validate:"required"`
	ServiceAccountName string `validate:"required,email"`
	ImageURL           string `validate:"required,http_url"`
}

// FirestoreConfig contains Firestore runtime settings.
type FirestoreConfig struct {
	ProjectID    string `validate:"required"`
	EmulatorHost string
}

// PubSubConfig contains Pub/Sub runtime settings.
type PubSubConfig struct {
	ProjectID    string `validate:"required"`
	EmulatorHost string
	FetcherTopic string `validate:"required"`
}

// FetcherConfig contains fetcher callback authentication settings.
type FetcherConfig struct {
	SharedSecret string `validate:"required,min=32"`
}

// TracingConfig contains tracing exporter settings.
type TracingConfig struct {
	Enabled            bool
	ExporterType       string `validate:"required,oneof=CONSOLE GOOGLE JAEGER"`
	JaegerEndpoint     string
	GoogleCloudProject string
	ServiceName        string `validate:"required"`
}

// RuntimeConfig contains all API runtime configuration loaded at startup.
type RuntimeConfig struct {
	Auth0       Auth0Config
	Features    FeatureFlags
	Fetcher     FetcherConfig
	Firestore   FirestoreConfig
	HTTP        HTTPSecurityConfig
	KitchenPass KitchenPassConfig
	PubSub      PubSubConfig
	Recipe      RecipeRuntimeConfig
	Routes      RouteConfig
	Static      StaticMediaConfig
	System      SystemConfig
	Tracing     TracingConfig
	Typesense   TypesenseConfig
}

type rawHTTPSecurityConfig struct {
	CORSAllowedOrigins string `required:"true" env:"CORS_ALLOWED_ORIGINS"`
	TrustedProxyCIDRs  string `required:"true" env:"TRUSTED_PROXY_CIDRS"`
}

type rawRuntimeConfig struct {
	Auth0Audience       string `required:"true" env:"AUTH0_AUDIENCE"`
	Auth0Domain         string `required:"true" env:"AUTH0_DOMAIN"`
	AppBaseURL          string `required:"true" env:"APP_BASE_URL"`
	CORSAllowedOrigins  string `required:"true" env:"CORS_ALLOWED_ORIGINS"`
	Development         bool   `env:"IO_4KS_DEVELOPMENT"`
	DistributionBucket  string `required:"true" env:"DISTRIBUTION_BUCKET"`
	FetcherSharedSecret string `required:"true" env:"API_FETCHER_PSK"`
	FetcherTopic        string `default:"fetcher" env:"FETCHER_TOPIC_ID"`
	FirestoreEmulator   string `env:"FIRESTORE_EMULATOR_HOST"`
	FirestoreProjectID  string `required:"true" env:"FIRESTORE_PROJECT_ID"`
	GinMode             string `default:"release" env:"GIN_MODE"`
	GoogleCloudProject  string `env:"GOOGLE_CLOUD_PROJECT"`
	JaegerEndpoint      string `default:"http://jaeger:14268/api/traces" env:"OTEL_EXPORTER_JAEGER_ENDPOINT"`
	MediaFallbackURL    string `required:"true" env:"MEDIA_FALLBACK_URL"`
	MediaImageURL       string `required:"true" env:"MEDIA_IMAGE_URL"`
	PATDigestSecret     string `required:"true" env:"PAT_DIGEST_SECRET"`
	PATEncryptionSecret string `required:"true" env:"PAT_ENCRYPTION_SECRET"`
	Port                string `default:"5000" env:"PORT"`
	PubSubEmulator      string `env:"PUBSUB_EMULATOR_HOST"`
	PubSubProjectID     string `required:"true" env:"PUBSUB_PROJECT_ID"`
	ReservedWordsFile   string `default:"./reserved-words" env:"RESERVED_WORDS_FILE"`
	ServiceAccountEmail string `required:"true" env:"SERVICE_ACCOUNT_EMAIL"`
	StaticMediaBucket   string `required:"true" env:"STATIC_MEDIA_BUCKET"`
	StaticMediaPrefix   string `required:"true" env:"STATIC_MEDIA_FALLBACK_PREFIX"`
	SwaggerEnabled      bool   `env:"SWAGGER_ENABLED"`
	SwaggerURLPrefix    string `env:"SWAGGER_URL_PREFIX"`
	TracingEnabled      bool   `env:"JAEGER_ENABLED"`
	TracingExporterType string `default:"CONSOLE" env:"EXPORTER_TYPE"`
	TracingServiceName  string `default:"4ks-api" env:"OTEL_SERVICE_NAME"`
	TrustedProxyCIDRs   string `required:"true" env:"TRUSTED_PROXY_CIDRS"`
	TypesenseAPIKey     string `required:"true" env:"TYPESENSE_API_KEY"`
	TypesenseURL        string `required:"true" env:"TYPESENSE_URL"`
	UploadableBucket    string `required:"true" env:"UPLOADABLE_BUCKET"`
	VersionFilePath     string `env:"VERSION_FILE_PATH"`
}

// SystemFlags returns the legacy flags struct expected by existing services.
func (cfg RuntimeConfig) SystemFlags() SystemFlags {
	return SystemFlags{
		Debug:         strings.EqualFold(cfg.System.GinMode, "debug"),
		Development:   cfg.System.Development,
		JaegerEnabled: cfg.Tracing.Enabled,
	}
}

// MinimalRuntimeConfig returns a valid baseline configuration for tests.
func MinimalRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		System: SystemConfig{
			GinMode:     "release",
			Development: false,
		},
		Features: FeatureFlags{
			SwaggerEnabled: false,
		},
		Routes: RouteConfig{
			Port:              "5000",
			SwaggerURLPrefix:  "",
			ReservedWordsFile: "./reserved-words",
			VersionFilePath:   "",
		},
		HTTP: HTTPSecurityConfig{
			CORS: CORSConfig{
				AllowedOrigins:   []string{"https://www.4ks.io"},
				AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "HEAD", "OPTIONS"},
				AllowedHeaders:   []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
				ExposedHeaders:   []string{"Content-Length"},
				AllowCredentials: true,
				MaxAge:           24 * time.Hour,
			},
			Proxy: ProxyConfig{
				TrustedCIDRs: []string{"127.0.0.1/32"},
			},
		},
		Auth0: Auth0Config{
			Domain:   "example.auth0.com",
			Audience: "test",
		},
		KitchenPass: KitchenPassConfig{
			BaseURL:          "https://www.4ks.io",
			DigestSecret:     "01234567890123456789012345678901",
			EncryptionSecret: "abcdefghijklmnopqrstuvwxyz012345",
		},
		Typesense: TypesenseConfig{
			URL:    "http://typesense:8108",
			APIKey: "test-key",
		},
		Static: StaticMediaConfig{
			FallbackURL:    "https://media.4ks.io/fallback.jpg",
			Bucket:         "static-media",
			FallbackPrefix: "fallback",
		},
		Recipe: RecipeRuntimeConfig{
			DistributionBucket: "distribution",
			UploadableBucket:   "uploadable",
			ServiceAccountName: "svc@example.com",
			ImageURL:           "https://media.4ks.io",
		},
		Firestore: FirestoreConfig{
			ProjectID: "test-firestore",
		},
		PubSub: PubSubConfig{
			ProjectID:    "test-pubsub",
			FetcherTopic: "fetcher",
		},
		Fetcher: FetcherConfig{
			SharedSecret: "01234567890123456789012345678901",
		},
		Tracing: TracingConfig{
			Enabled:        false,
			ExporterType:   "CONSOLE",
			JaegerEndpoint: "http://jaeger:14268/api/traces",
			ServiceName:    "4ks-api",
		},
	}
}

// LoadRuntimeConfig loads and validates the full API runtime config once.
func LoadRuntimeConfig() (*RuntimeConfig, error) {
	raw := rawRuntimeConfig{}
	if err := configor.New(&configor.Config{ENVPrefix: ""}).Load(&raw); err != nil {
		return nil, err
	}

	return buildRuntimeConfig(raw)
}

// LoadHTTPSecurityConfig loads and validates HTTP-facing security settings.
func LoadHTTPSecurityConfig() (*HTTPSecurityConfig, error) {
	raw := rawHTTPSecurityConfig{}
	if err := configor.New(&configor.Config{ENVPrefix: ""}).Load(&raw); err != nil {
		return nil, err
	}

	return buildHTTPSecurityConfig(raw.CORSAllowedOrigins, raw.TrustedProxyCIDRs)
}

func buildRuntimeConfig(raw rawRuntimeConfig) (*RuntimeConfig, error) {
	httpConfig, err := buildHTTPSecurityConfig(raw.CORSAllowedOrigins, raw.TrustedProxyCIDRs)
	if err != nil {
		return nil, err
	}

	cfg := &RuntimeConfig{
		System: SystemConfig{
			GinMode:     raw.GinMode,
			Development: raw.Development,
		},
		Features: FeatureFlags{
			SwaggerEnabled: raw.SwaggerEnabled,
		},
		Routes: RouteConfig{
			Port:              raw.Port,
			SwaggerURLPrefix:  raw.SwaggerURLPrefix,
			ReservedWordsFile: raw.ReservedWordsFile,
			VersionFilePath:   raw.VersionFilePath,
		},
		HTTP: *httpConfig,
		Auth0: Auth0Config{
			Domain:   raw.Auth0Domain,
			Audience: raw.Auth0Audience,
		},
		KitchenPass: KitchenPassConfig{
			BaseURL:          raw.AppBaseURL,
			DigestSecret:     raw.PATDigestSecret,
			EncryptionSecret: raw.PATEncryptionSecret,
		},
		Typesense: TypesenseConfig{
			URL:    raw.TypesenseURL,
			APIKey: raw.TypesenseAPIKey,
		},
		Static: StaticMediaConfig{
			FallbackURL:    raw.MediaFallbackURL,
			Bucket:         raw.StaticMediaBucket,
			FallbackPrefix: raw.StaticMediaPrefix,
		},
		Recipe: RecipeRuntimeConfig{
			DistributionBucket: raw.DistributionBucket,
			UploadableBucket:   raw.UploadableBucket,
			ServiceAccountName: raw.ServiceAccountEmail,
			ImageURL:           raw.MediaImageURL,
		},
		Firestore: FirestoreConfig{
			ProjectID:    raw.FirestoreProjectID,
			EmulatorHost: raw.FirestoreEmulator,
		},
		PubSub: PubSubConfig{
			ProjectID:    raw.PubSubProjectID,
			EmulatorHost: raw.PubSubEmulator,
			FetcherTopic: raw.FetcherTopic,
		},
		Fetcher: FetcherConfig{
			SharedSecret: raw.FetcherSharedSecret,
		},
		Tracing: TracingConfig{
			Enabled:            raw.TracingEnabled,
			ExporterType:       strings.ToUpper(raw.TracingExporterType),
			JaegerEndpoint:     raw.JaegerEndpoint,
			GoogleCloudProject: raw.GoogleCloudProject,
			ServiceName:        raw.TracingServiceName,
		},
	}

	if cfg.Routes.Port == "" {
		return nil, fmt.Errorf("PORT must not be empty")
	}
	if err := validateRuntimeConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func buildHTTPSecurityConfig(corsAllowedOrigins string, trustedProxyCIDRs string) (*HTTPSecurityConfig, error) {
	cfg := &HTTPSecurityConfig{
		CORS: CORSConfig{
			// These defaults reflect the API surface that browsers are expected to use.
			AllowedOrigins: parseCSVValues(corsAllowedOrigins),
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
			TrustedCIDRs: parseCSVValues(trustedProxyCIDRs),
		},
	}

	if err := validateHTTPSecurityConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseCSVValues(value string) []string {
	if value == "" {
		return nil
	}

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

func validateRuntimeConfig(cfg *RuntimeConfig) error {
	validate := newConfigValidator()
	if err := validate.Struct(cfg); err != nil {
		return normalizeValidationError(err)
	}

	switch cfg.Tracing.ExporterType {
	case "GOOGLE":
		if cfg.Tracing.GoogleCloudProject == "" {
			return fmt.Errorf("Tracing.GoogleCloudProject is required when Tracing.ExporterType=GOOGLE")
		}
	case "JAEGER":
		if cfg.Tracing.JaegerEndpoint == "" {
			return fmt.Errorf("Tracing.JaegerEndpoint is required when Tracing.ExporterType=JAEGER")
		}
		if err := validate.Var(cfg.Tracing.JaegerEndpoint, "http_url"); err != nil {
			return fmt.Errorf("Tracing.JaegerEndpoint must be a valid HTTP URL")
		}
	}

	return nil
}

func validateHTTPSecurityConfig(cfg *HTTPSecurityConfig) error {
	validate := newConfigValidator()
	if err := validate.Struct(cfg); err != nil {
		return normalizeValidationError(err)
	}

	if cfg.CORS.AllowCredentials {
		for _, origin := range cfg.CORS.AllowedOrigins {
			if origin == "*" {
				return fmt.Errorf("HTTP.CORS.AllowedOrigins cannot contain * when credentials are enabled")
			}
		}
	}

	return nil
}

func newConfigValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("cors_origin", validateCORSOrigin)
	return validate
}

func validateCORSOrigin(fl validator.FieldLevel) bool {
	origin := fl.Field().String()
	if origin == "*" {
		return false
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}

	host := parsed.Hostname()
	if host == "" {
		return false
	}

	if parsed.Port() != "" {
		if _, err := strconv.Atoi(parsed.Port()); err != nil {
			return false
		}
	}

	return strings.Contains(host, ".") || net.ParseIP(host) != nil || host == "localhost"
}

func normalizeValidationError(err error) error {
	var invalidValidation *validator.InvalidValidationError
	if errors.As(err, &invalidValidation) {
		return err
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return err
	}

	messages := make([]string, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		field := strings.TrimPrefix(validationErr.Namespace(), "RuntimeConfig.")
		field = strings.TrimPrefix(field, "HTTPSecurityConfig.")
		if field == "" {
			field = validationErr.Field()
		}

		switch validationErr.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", field))
		case "oneof":
			messages = append(messages, fmt.Sprintf("%s must be one of [%s]", field, validationErr.Param()))
		case "numeric":
			messages = append(messages, fmt.Sprintf("%s must be numeric", field))
		case "hostname_rfc1123":
			messages = append(messages, fmt.Sprintf("%s must be a valid hostname", field))
		case "http_url":
			messages = append(messages, fmt.Sprintf("%s must be a valid HTTP URL", field))
		case "email":
			messages = append(messages, fmt.Sprintf("%s must be a valid email address", field))
		case "cidr":
			messages = append(messages, fmt.Sprintf("%s must be a valid CIDR", field))
		case "min":
			if validationErr.Kind() == reflect.String {
				messages = append(messages, fmt.Sprintf("%s must be at least %s characters", field, validationErr.Param()))
			} else {
				messages = append(messages, fmt.Sprintf("%s must contain at least %s item(s)", field, validationErr.Param()))
			}
		case "gt":
			messages = append(messages, fmt.Sprintf("%s must be greater than %s", field, validationErr.Param()))
		case "cors_origin":
			messages = append(messages, fmt.Sprintf("%s must contain exact origins in scheme://host form with no path, query, or fragment", field))
		default:
			messages = append(messages, fmt.Sprintf("%s failed %s validation", field, validationErr.Tag()))
		}
	}

	return errors.New(strings.Join(messages, "; "))
}
