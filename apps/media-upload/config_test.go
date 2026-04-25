package function

import "testing"

func TestLoadRuntimeConfig(t *testing.T) {
	t.Setenv("DISTRIBUTION_BUCKET", "distribution")
	t.Setenv("FIRESTORE_PROJECT_ID", "firestore-project")
	t.Setenv("IO_4KS_DEVELOPMENT", "true")
	t.Setenv("PORT", "8181")

	cfg, err := LoadRuntimeConfig()
	if err != nil {
		t.Fatalf("LoadRuntimeConfig returned error: %v", err)
	}

	if !cfg.Development || cfg.Port != "8181" || cfg.DistributionBucket != "distribution" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
