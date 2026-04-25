package fetcher

import "testing"

func TestLoadRuntimeConfig(t *testing.T) {
	t.Setenv("DEBUG", "true")
	t.Setenv("API_FETCHER_PSK", "01234567890123456789012345678901")
	t.Setenv("API_ENDPOINT_URL", "https://api.4ks.io/api/_fetcher/recipes")
	t.Setenv("PUBSUB_PROJECT_ID", "test-project")
	t.Setenv("PUBSUB_TOPIC_ID", "fetcher")
	t.Setenv("PORT", "9191")

	cfg, err := LoadRuntimeConfig()
	if err != nil {
		t.Fatalf("LoadRuntimeConfig returned error: %v", err)
	}

	if !cfg.Debug || cfg.Port != "9191" || cfg.PubSubTopicID != "fetcher" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
