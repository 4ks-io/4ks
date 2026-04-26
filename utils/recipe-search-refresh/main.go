// Package main refreshes the Typesense recipe search index from Firestore.
package main

import (
	"4ks/apps/api/dtos"
	"4ks/libs/go/models"
	"context"
	"encoding/json"
	"flag"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/rs/zerolog/log"
	"github.com/typesense/typesense-go/typesense"
	"gopkg.in/yaml.v2"
)

func main() {
	// context
	ctx := context.Background()
	log.Info().Msg("Starting media-update")

	// app config
	cfg := getConfig()
	// utils.PrintStruct(cfg)

	// typesense client
	ts := typesense.NewClient(typesense.WithServer(cfg.TypesenseURL), typesense.WithAPIKey(cfg.TypesenseKey))

	// firestore client
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
			var r models.Recipe
			err = json.Unmarshal(jsonBytes, &r)
			if err != nil {
				log.Fatal().Err(err).Caller().Msg("failed to unmarshal firestore document")
			}

			// upsert typesense
			ing := []string{}
			for _, v := range r.CurrentRevision.Ingredients {
				ing = append(ing, v.Name)
			}

			var banner string
			for _, b := range r.CurrentRevision.Banner {
				if b.Alias == "md" {
					banner = b.URL
				}
			}

			document := dtos.CreateSearchRecipe{
				ID:          r.ID,
				Author:      r.Author.Username,
				Name:        r.CurrentRevision.Name,
				Ingredients: ing,
				ImageURL:    banner,
			}

			log.Info().Str("id", document.ID).Str("title", r.CurrentRevision.Name).Msg("upsert")

			if !cfg.DryRun {
				_, err = ts.Collection("recipes").Documents().Upsert(document)
				if err != nil {
					log.Error().Err(err).Caller().Msg("upsert failed")
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
	TypesenseURL       string `yaml:"typesense_url"`
	TypesenseKey       string `yaml:"typesense_key"`
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
