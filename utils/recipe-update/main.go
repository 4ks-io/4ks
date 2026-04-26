// Package main runs a one-off recipe data migration utility.
package main

import (
	"4ks/libs/go/models"
	utils "4ks/libs/go/utils"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

func main() {
	// context
	ctx := context.Background()
	log.Info().Msg("Starting media-update")

	// app config
	cfg := getConfig()
	utils.PrintStruct(cfg)

	// Initialize Firestore client.
	db, err := firestore.NewClientWithDatabase(ctx, cfg.ProjectID, cfg.FireDatabaseID)
	if err != nil {
		log.Fatal().Err(err).Caller().Msg("failed to create Firestore client")
	}
	defer db.Close()

	// track files and batches counts
	fileIdx := 0
	batchIdx := 0

	// build query
	var checkpoint string
	col := db.Collection(cfg.FireCollectionName)

Batch:
	for {
		log.Info().Int("batch", batchIdx).Int("files", fileIdx).Str("checkpoint", checkpoint).Msg("Processing batch")
		query := col.OrderBy("id", firestore.Asc).StartAfter(checkpoint).Limit(cfg.BatchSize)
		snaps, err := query.Documents(ctx).GetAll()
		if err != nil {
			log.Error().Err(err).Caller().Msg("failed to get documents")
			break Batch
		}
		if len(snaps) == 0 {
			log.Info().Msg("No more documents")
			break Batch // No more documents
		}

		// process batch
		for _, snap := range snaps {
			fileIdx++

			// Marshal the map to a JSON byte slice
			jsonBytes, err := json.Marshal(snap.Data())
			if err != nil {
				log.Fatal().Err(err).Caller().Msg("failed to marshal firestore document")
			}

			// Unmarshal the JSON data into the struct
			var img models.Recipe
			err = json.Unmarshal(jsonBytes, &img)
			if err != nil {
				log.Fatal().Err(err).Caller().Msg("failed to unmarshal firestore document")
			}

			// update url
			newVar := make([]models.RecipeMediaVariant, len(img.CurrentRevision.Banner))
			for i, v := range img.CurrentRevision.Banner {
				u := strings.Replace(v.URL, "storage.googleapis.com/media-read.4ks.io", "www.4ks.io/image", 1)
				u = strings.Replace(u, "storage.googleapis.com/static.4ks.io/fallback", "www.4ks.io/static/fallback", 1)
				u = strings.Replace(u, "storage.googleapis.com/static.4ks.io/static", "www.4ks.io/static", 1)
				if v.URL != u {
					fmt.Println(v.URL, " -> ", u)
					v.URL = u
				}
				newVar[i] = v
			}
			img.CurrentRevision.Banner = newVar

			// update firestore document
			if !cfg.DryRun {
				ref := col.Doc(snap.Ref.ID)
				_, err = ref.Set(ctx, &img)
				if err != nil {
					log.Fatal().Err(err).Caller().Msg("failed to set firestore document")
				}
			}

			// Limit file count
			if cfg.MaxFiles > 0 && fileIdx >= cfg.MaxFiles {
				log.Info().
					Int("files", fileIdx).
					Int("max", cfg.MaxFiles).
					Int("files", fileIdx).
					Msg("MAX FILES REACHED")
				// exit the Batch loop
				break Batch
			}

			if checkpoint == snap.Ref.ID {
				break Batch // No more documents
			}
			checkpoint = snap.Ref.ID
		}

	}
	log.Info().Msg("done")
}

// AppConfig is the configuration for the app
type AppConfig struct {
	Debug              bool
	DryRun             bool   `yaml:"dry_run"`
	ProjectID          string `yaml:"project_id"`
	FireDatabaseID     string `yaml:"fire_database_id"`
	FireCollectionName string `yaml:"fire_collection_name"`
	BatchSize          int    `yaml:"batch_size"`
	MaxFiles           int    `yaml:"max_files"`
}

func getConfig() AppConfig {
	dryRun := flag.Bool("dry-run", false, "Dry Run")
	debugFlag := flag.Bool("debug", false, "Debug logging level")
	configFilename := flag.String("config", "local.env.yaml", "Config File")
	flag.Parse()

	// Read config file
	configFile, err := os.ReadFile(*configFilename)
	if err != nil {
		panic(err)
	}

	// Unmarshal the JSON data into a Config struct
	var cfg AppConfig
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		panic(err)
	}

	if *debugFlag {
		cfg.Debug = *debugFlag
	}

	if *dryRun {
		cfg.DryRun = *dryRun
	}

	return cfg
}
