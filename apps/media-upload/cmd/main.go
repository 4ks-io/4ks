package main

import (
	"log"

	function "4ks.io/media-upload"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	cfg := function.MustLoadRuntimeConfig()

	if err := funcframework.Start(cfg.Port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
