package fetcher

import (
	"errors"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"
)

// RuntimeConfig contains the fetcher function's startup configuration.
type RuntimeConfig struct {
	Debug           bool
	APISharedSecret string `validate:"required,min=32"`
	APIEndpoint     string `validate:"required,http_url"`
	PubSubProjectID string `validate:"required"`
	PubSubTopicID   string `validate:"required"`
	Port            string `validate:"required,numeric"`
}

type rawRuntimeConfig struct {
	Debug           bool   `env:"DEBUG"`
	APISharedSecret string `required:"true" env:"API_FETCHER_PSK"`
	APIEndpoint     string `required:"true" env:"API_ENDPOINT_URL"`
	PubSubProjectID string `required:"true" env:"PUBSUB_PROJECT_ID"`
	PubSubTopicID   string `required:"true" env:"PUBSUB_TOPIC_ID"`
	Port            string `default:"5000" env:"PORT"`
}

// LoadRuntimeConfig loads and validates the fetcher runtime config once.
func LoadRuntimeConfig() (*RuntimeConfig, error) {
	raw := rawRuntimeConfig{}
	if err := configor.New(&configor.Config{ENVPrefix: ""}).Load(&raw); err != nil {
		return nil, err
	}

	cfg := &RuntimeConfig{
		Debug:           raw.Debug,
		APISharedSecret: raw.APISharedSecret,
		APIEndpoint:     raw.APIEndpoint,
		PubSubProjectID: raw.PubSubProjectID,
		PubSubTopicID:   raw.PubSubTopicID,
		Port:            raw.Port,
	}

	if err := validateRuntimeConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func init() {
	cfg := MustLoadRuntimeConfig()
	Register(cfg)
}

// Register registers the CloudEvent handler with the Functions Framework.
func Register(cfg RuntimeConfig) {
	functions.CloudEvent(
		fmt.Sprintf("projects/%s/topics/%s", cfg.PubSubProjectID, cfg.PubSubTopicID),
		newFetcherHandler(cfg),
	)
}

// MustLoadRuntimeConfig loads config and exits on validation failure.
func MustLoadRuntimeConfig() RuntimeConfig {
	cfg, err := LoadRuntimeConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid fetcher runtime configuration")
	}

	return *cfg
}

func validateRuntimeConfig(cfg *RuntimeConfig) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return normalizeValidationError(err)
	}

	return nil
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
		if field == "" {
			field = validationErr.Field()
		}

		switch validationErr.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", field))
		case "http_url":
			messages = append(messages, fmt.Sprintf("%s must be a valid HTTP URL", field))
		case "numeric":
			messages = append(messages, fmt.Sprintf("%s must be numeric", field))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s characters", field, validationErr.Param()))
		default:
			messages = append(messages, fmt.Sprintf("%s failed %s validation", field, validationErr.Tag()))
		}
	}

	return errors.New(strings.Join(messages, "; "))
}
