package utils

import "testing"

func TestLoadHTTPSecurityConfig(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://www.4ks.io, https://dev.4ks.io")
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8,127.0.0.1/32")

	cfg, err := LoadHTTPSecurityConfig()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Fatalf("expected 2 allowed origins, got %d", len(cfg.CORS.AllowedOrigins))
	}
	if len(cfg.Proxy.TrustedCIDRs) != 2 {
		t.Fatalf("expected 2 trusted proxy CIDRs, got %d", len(cfg.Proxy.TrustedCIDRs))
	}
}

func TestLoadHTTPSecurityConfigRejectsWildcardOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")

	if _, err := LoadHTTPSecurityConfig(); err == nil {
		t.Fatal("expected wildcard origin to be rejected")
	}
}

func TestLoadHTTPSecurityConfigRejectsInvalidProxyCIDR(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://www.4ks.io")
	t.Setenv("TRUSTED_PROXY_CIDRS", "not-a-cidr")

	if _, err := LoadHTTPSecurityConfig(); err == nil {
		t.Fatal("expected invalid proxy cidr to be rejected")
	}
}
