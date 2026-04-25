package utils

import (
	"testing"
)

// ---------------------------------------------------------------------------
// GetStrEnvVar
// ---------------------------------------------------------------------------

func TestGetStrEnvVarReturnsSetValue(t *testing.T) {
	t.Setenv("TEST_STR", "hello")
	if got := GetStrEnvVar("TEST_STR", "fallback"); got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestGetStrEnvVarReturnsFallbackWhenUnset(t *testing.T) {
	if got := GetStrEnvVar("DEFINITELY_NOT_SET_XYZ", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestGetStrEnvVarReturnsEmptyStringWhenSetToEmpty(t *testing.T) {
	t.Setenv("TEST_EMPTY", "")
	if got := GetStrEnvVar("TEST_EMPTY", "fallback"); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// GetBoolEnv
// ---------------------------------------------------------------------------

func TestGetBoolEnvReturnsTrueForTrueString(t *testing.T) {
	t.Setenv("TEST_BOOL", "true")
	if !GetBoolEnv("TEST_BOOL", false) {
		t.Fatal("expected true")
	}
}

func TestGetBoolEnvReturnsFalseForFalseString(t *testing.T) {
	t.Setenv("TEST_BOOL", "false")
	if GetBoolEnv("TEST_BOOL", true) {
		t.Fatal("expected false")
	}
}

func TestGetBoolEnvAccepts1And0(t *testing.T) {
	t.Setenv("TEST_BOOL_ONE", "1")
	if !GetBoolEnv("TEST_BOOL_ONE", false) {
		t.Fatal("expected 1 to parse as true")
	}
	t.Setenv("TEST_BOOL_ZERO", "0")
	if GetBoolEnv("TEST_BOOL_ZERO", true) {
		t.Fatal("expected 0 to parse as false")
	}
}

func TestGetBoolEnvReturnsFallbackWhenUnset(t *testing.T) {
	if GetBoolEnv("DEFINITELY_NOT_SET_BOOL_XYZ", true) != true {
		t.Fatal("expected fallback true")
	}
	if GetBoolEnv("DEFINITELY_NOT_SET_BOOL_XYZ", false) != false {
		t.Fatal("expected fallback false")
	}
}

func TestGetBoolEnvReturnsFallbackForInvalidValue(t *testing.T) {
	t.Setenv("TEST_BOOL_INVALID", "not-a-bool")
	if GetBoolEnv("TEST_BOOL_INVALID", true) != true {
		t.Fatal("expected fallback on invalid value")
	}
}

// ---------------------------------------------------------------------------
// GetEnvVarOrPanic
// ---------------------------------------------------------------------------

func TestGetEnvVarOrPanicReturnsValue(t *testing.T) {
	t.Setenv("TEST_REQUIRED", "secret")
	if got := GetEnvVarOrPanic("TEST_REQUIRED"); got != "secret" {
		t.Fatalf("expected %q, got %q", "secret", got)
	}
}

func TestGetEnvVarOrPanicPanicsWhenUnset(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unset env var")
		}
	}()
	GetEnvVarOrPanic("DEFINITELY_NOT_SET_REQUIRED_XYZ")
}

func TestGetEnvVarOrPanicPanicsWhenEmpty(t *testing.T) {
	t.Setenv("TEST_EMPTY_REQUIRED", "")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty env var")
		}
	}()
	GetEnvVarOrPanic("TEST_EMPTY_REQUIRED")
}

// ---------------------------------------------------------------------------
// GetCSVEnvVar
// ---------------------------------------------------------------------------

func TestGetCSVEnvVarReturnsFallbackWhenUnset(t *testing.T) {
	fallback := []string{"a", "b"}
	got := GetCSVEnvVar("DEFINITELY_NOT_SET_CSV_XYZ", fallback)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected fallback slice, got %v", got)
	}
}

func TestGetCSVEnvVarReturnsFallbackCopy(t *testing.T) {
	fallback := []string{"x"}
	got := GetCSVEnvVar("DEFINITELY_NOT_SET_CSV_XYZ", fallback)
	got[0] = "mutated"
	if fallback[0] != "x" {
		t.Fatal("GetCSVEnvVar must return a copy of the fallback, not the original")
	}
}

func TestGetCSVEnvVarSplitsOnComma(t *testing.T) {
	t.Setenv("TEST_CSV", "foo,bar,baz")
	got := GetCSVEnvVar("TEST_CSV", nil)
	if len(got) != 3 || got[0] != "foo" || got[1] != "bar" || got[2] != "baz" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestGetCSVEnvVarTrimsSpaces(t *testing.T) {
	t.Setenv("TEST_CSV_SPACES", "  foo , bar  , baz  ")
	got := GetCSVEnvVar("TEST_CSV_SPACES", nil)
	if len(got) != 3 || got[0] != "foo" || got[1] != "bar" || got[2] != "baz" {
		t.Fatalf("expected trimmed values, got %v", got)
	}
}

func TestGetCSVEnvVarIgnoresEmptySegments(t *testing.T) {
	t.Setenv("TEST_CSV_EMPTY", "foo,,bar,")
	got := GetCSVEnvVar("TEST_CSV_EMPTY", nil)
	if len(got) != 2 || got[0] != "foo" || got[1] != "bar" {
		t.Fatalf("expected empty segments to be dropped, got %v", got)
	}
}

func TestGetCSVEnvVarReturnsSingleValue(t *testing.T) {
	t.Setenv("TEST_CSV_ONE", "only")
	got := GetCSVEnvVar("TEST_CSV_ONE", nil)
	if len(got) != 1 || got[0] != "only" {
		t.Fatalf("expected [only], got %v", got)
	}
}

// ---------------------------------------------------------------------------
// LoadHTTPSecurityConfig
// ---------------------------------------------------------------------------

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
