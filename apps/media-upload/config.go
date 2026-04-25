package function

import (
	"errors"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"
)

// RuntimeConfig contains media-upload startup configuration.
type RuntimeConfig struct {
	DistributionBucket string `validate:"required"`
	FirestoreProjectID string `validate:"required"`
	Development        bool
	Port               string `validate:"required,numeric"`
}

type rawRuntimeConfig struct {
	DistributionBucket string `required:"true" env:"DISTRIBUTION_BUCKET"`
	FirestoreProjectID string `required:"true" env:"FIRESTORE_PROJECT_ID"`
	Development        bool   `env:"IO_4KS_DEVELOPMENT"`
	Port               string `default:"8080" env:"PORT"`
}

// LoadRuntimeConfig loads and validates the function config once.
func LoadRuntimeConfig() (*RuntimeConfig, error) {
	raw := rawRuntimeConfig{}
	if err := configor.New(&configor.Config{ENVPrefix: ""}).Load(&raw); err != nil {
		return nil, err
	}

	cfg := &RuntimeConfig{
		DistributionBucket: raw.DistributionBucket,
		FirestoreProjectID: raw.FirestoreProjectID,
		Development:        raw.Development,
		Port:               raw.Port,
	}

	if err := validateRuntimeConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Register registers the upload handler with the Functions Framework.
func Register(cfg RuntimeConfig) {
	functions.CloudEvent("UploadImage", newUploadImageHandler(cfg))
}

// MustLoadRuntimeConfig loads config and exits on validation failure.
func MustLoadRuntimeConfig() RuntimeConfig {
	cfg, err := LoadRuntimeConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid media-upload runtime configuration")
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
		case "numeric":
			messages = append(messages, fmt.Sprintf("%s must be numeric", field))
		default:
			messages = append(messages, fmt.Sprintf("%s failed %s validation", field, validationErr.Tag()))
		}
	}

	return errors.New(strings.Join(messages, "; "))
}
