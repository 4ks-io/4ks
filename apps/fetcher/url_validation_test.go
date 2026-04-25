package fetcher

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
)

func TestNormalizeFetchURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		wantURL string
		wantErr error
	}{
		{name: "normalizes host case", input: "  HTTPS://Example.COM/Recipe  ", wantURL: "https://example.com/Recipe"},
		{name: "preserves port", input: "https://Example.COM:8443/Recipe", wantURL: "https://example.com:8443/Recipe"},
		{name: "empty", input: "", wantErr: errEmptyURL},
		{name: "http scheme", input: "http://example.com", wantErr: errUnsupportedScheme},
		{name: "credentials", input: "https://user:pass@example.com", wantErr: errURLHasCredentials},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, normalized, err := normalizeFetchURL(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeFetchURL returned error: %v", err)
			}
			if normalized != tc.wantURL {
				t.Fatalf("normalized url = %q, want %q", normalized, tc.wantURL)
			}
		})
	}
}

func TestHostnameAndIPBlocking(t *testing.T) {
	t.Parallel()

	if !isBlockedHostname("localhost") || !isBlockedHostname("api.localhost") {
		t.Fatal("expected localhost hostnames to be blocked")
	}
	if isBlockedHostname("example.com") {
		t.Fatal("did not expect example.com to be blocked")
	}
	if !isIPLiteral("127.0.0.1") || !isIPLiteral("[::1]") {
		t.Fatal("expected IP literals to be detected")
	}
	if isIPLiteral("example.com") {
		t.Fatal("did not expect hostname to be treated as IP literal")
	}

	privateIP := net.ParseIP("10.0.0.1")
	publicIP := net.ParseIP("8.8.8.8")
	if !isBlockedIP(privateIP) {
		t.Fatal("expected private IP to be blocked")
	}
	if isBlockedIP(publicIP) {
		t.Fatal("did not expect public IP to be blocked")
	}
}

func TestValidateResolvedIPs(t *testing.T) {
	t.Parallel()

	err := validateResolvedIPs([]net.IPAddr{{IP: net.ParseIP("127.0.0.1")}})
	if !errors.Is(err, errBlockedIPAddress) {
		t.Fatalf("expected blocked IP error, got %v", err)
	}

	err = validateResolvedIPs([]net.IPAddr{{IP: net.ParseIP("8.8.8.8")}})
	if err != nil {
		t.Fatalf("expected public IP to pass, got %v", err)
	}
}

func TestMustPrefixes(t *testing.T) {
	t.Parallel()

	prefixes := mustPrefixes("10.0.0.0/8", "192.168.0.0/16")
	if len(prefixes) != 2 {
		t.Fatalf("expected two prefixes, got %d", len(prefixes))
	}
	if !prefixes[0].Contains(netip.MustParseAddr("10.0.0.1")) {
		t.Fatal("expected first prefix to contain 10.0.0.1")
	}
}

func TestValidateFetchURLRejectsBadInputs(t *testing.T) {
	t.Parallel()

	if _, err := validateFetchURL(context.Background(), "https://localhost/recipe"); !errors.Is(err, errBlockedHostname) {
		t.Fatalf("expected blocked hostname error, got %v", err)
	}
	if _, err := validateFetchURL(context.Background(), "https://127.0.0.1/recipe"); !errors.Is(err, errIPLiteralNotAllowed) {
		t.Fatalf("expected IP literal error, got %v", err)
	}
}
