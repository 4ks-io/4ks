package fetchurl

import (
	"context"
	"net"
	"testing"
)

type staticResolver struct {
	ipAddrs []net.IPAddr
	err     error
}

func (s staticResolver) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	return s.ipAddrs, s.err
}

func TestNormalizeRejectsCredentials(t *testing.T) {
	_, _, err := Normalize("https://user:pass@example.com/recipe")
	if err == nil {
		t.Fatal("expected credentials to be rejected")
	}
}

func TestValidateRejectsIPLiteral(t *testing.T) {
	_, err := ValidateWithResolver(context.Background(), "https://127.0.0.1/recipe", staticResolver{})
	if err == nil {
		t.Fatal("expected ip literal target to be rejected")
	}
}

func TestValidateRejectsBlockedDNSAnswer(t *testing.T) {
	_, err := ValidateWithResolver(context.Background(), "https://example.com/recipe", staticResolver{
		ipAddrs: []net.IPAddr{{IP: net.ParseIP("169.254.169.254")}},
	})
	if err == nil {
		t.Fatal("expected blocked dns answer to be rejected")
	}
}

func TestValidateAcceptsPublicHTTPSURL(t *testing.T) {
	validated, err := ValidateWithResolver(context.Background(), "https://example.com/recipe?a=1", staticResolver{
		ipAddrs: []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}},
	})
	if err != nil {
		t.Fatalf("expected valid url, got %v", err)
	}
	if validated.Normalized != "https://example.com/recipe?a=1" {
		t.Fatalf("unexpected normalized url: %s", validated.Normalized)
	}
}
