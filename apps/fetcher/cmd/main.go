// Package main is the main application for the nlp-worker local cmd
package main

import (
	"log"

	fetcher "4ks.io/fetcher"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	cfg := fetcher.MustLoadRuntimeConfig()
	fetcher.Register(cfg)

	if err := funcframework.Start(cfg.Port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
