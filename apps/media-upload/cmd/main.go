package main

import (
	"log"

	// Blank-import the function package so the init() runs
	function "4ks.io/media-upload"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

// export FUNCTION_TARGET=UploadImage

func main() {
	cfg := function.MustLoadRuntimeConfig()
	function.Register(cfg)

	if err := funcframework.Start(cfg.Port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
