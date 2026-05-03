package main

import (
	"bufio"
	"context"
	"os"
	"strings"

	"4ks/apps/api/app"
	controllers "4ks/apps/api/controllers"
	"4ks/apps/api/rest"
	fetcherService "4ks/apps/api/services/fetcher"
	kitchenPassService "4ks/apps/api/services/kitchenpass"
	recipeService "4ks/apps/api/services/recipe"
	searchService "4ks/apps/api/services/search"
	staticService "4ks/apps/api/services/static"
	userService "4ks/apps/api/services/user"
	utils "4ks/apps/api/utils"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/storage"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/typesense/typesense-go/typesense"
)

type runtimeWiring struct {
	services app.Services
	restDeps rest.Deps
	cleanup  func()
}

func buildRuntimeWiring(ctx context.Context, cfg *utils.RuntimeConfig, reservedWords []string) (runtimeWiring, error) {
	sysFlags := cfg.SystemFlags()

	store, err := storage.NewClient(ctx)
	if err != nil {
		return runtimeWiring{}, err
	}

	if cfg.Firestore.EmulatorHost != "" {
		log.Info().Msgf("Using Firestore Emulator: '%s'", cfg.Firestore.EmulatorHost)
	}
	fire, err := firestore.NewClient(ctx, cfg.Firestore.ProjectID)
	if err != nil {
		_ = store.Close()
		return runtimeWiring{}, err
	}

	if cfg.PubSub.EmulatorHost != "" {
		log.Info().Msgf("Using PubSub Emulator: '%s'", cfg.PubSub.EmulatorHost)
	}
	psub, err := pubsub.NewClient(ctx, cfg.PubSub.ProjectID)
	if err != nil {
		_ = fire.Close()
		_ = store.Close()
		return runtimeWiring{}, err
	}

	feso := fetcherService.FetcherOpts{ProjectID: cfg.PubSub.ProjectID, TopicID: cfg.PubSub.FetcherTopic}
	ts := typesense.NewClient(typesense.WithServer(cfg.Typesense.URL), typesense.WithAPIKey(cfg.Typesense.APIKey))
	v := validator.New()

	static := staticService.New(store, cfg.Static.FallbackURL, cfg.Static.Bucket, cfg.Static.FallbackPrefix)
	search := searchService.New(ts)
	kitchenPass := kitchenPassService.New(fire, kitchenPassService.Config{
		BaseURL:          cfg.KitchenPass.BaseURL,
		DigestSecret:     cfg.KitchenPass.DigestSecret,
		EncryptionSecret: cfg.KitchenPass.EncryptionSecret,
	})
	user := userService.New(&sysFlags, fire, v, &reservedWords)
	recipe := recipeService.New(&sysFlags, store, fire, v, &recipeService.RecipeServiceConfig{
		DistributionBucket: cfg.Recipe.DistributionBucket,
		UploadableBucket:   cfg.Recipe.UploadableBucket,
		ServiceAccountName: cfg.Recipe.ServiceAccountName,
		ImageURL:           cfg.Recipe.ImageURL,
	})
	fetcher := fetcherService.New(ctx, &sysFlags, psub, feso, user, recipe, search, static)

	return runtimeWiring{
		services: app.Services{
			User:        user,
			Recipe:      recipe,
			Search:      search,
			Static:      static,
			Fetcher:     fetcher,
			KitchenPass: kitchenPass,
		},
		restDeps: rest.Deps{
			Version: getAPIVersion(cfg.Routes.VersionFilePath),
			System: controllers.SystemControllerDeps{
				DB:        controllers.NewFirestoreProber(fire),
				Search:    controllers.NewTypesenseProber(ts),
				Messaging: controllers.NewPubSubProber(psub, feso.TopicID),
				Storage:   controllers.NewStorageProber(store, cfg.Recipe.DistributionBucket),
			},
		},
		cleanup: func() {
			_ = psub.Close()
			_ = fire.Close()
			_ = store.Close()
		},
	}, nil
}

// getAPIVersion returns the api version stored in the VERSION file.
func getAPIVersion(versionFilePath string) string {
	version := "0.0.0"
	if versionFilePath != "" {
		v, err := os.ReadFile(versionFilePath)
		if err != nil {
			panic(err)
		}
		version = strings.TrimSuffix(string(v), "\n")
	}
	return version
}

// configureLogging configures global logging based on the config file and flags.
func configureLogging() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	// Set log level
	zerolog.SetGlobalLevel(0)
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// ReadWordsFromFile reads words from a file.
func ReadWordsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}
