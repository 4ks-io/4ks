// Package fetchurl provides URL validation and normalization with SSRF protection.
package fetchurl

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const dnsLookupTimeout = 5 * time.Second

// Sentinel errors returned by Validate and related functions.
var (
	ErrEmptyURL             = errors.New("url is required")
	ErrInvalidURL           = errors.New("url must be an absolute https URL")
	ErrUnsupportedScheme    = errors.New("url scheme must be https")
	ErrURLHasCredentials    = errors.New("url must not contain embedded credentials")
	ErrBlockedHostname      = errors.New("url host is not allowed")
	ErrBlockedIPAddress     = errors.New("url resolves to a blocked address")
	ErrIPLiteralNotAllowed  = errors.New("ip literal targets are not allowed")
	ErrHostResolutionFailed = errors.New("url host could not be resolved")
)

// ValidatedURL holds the parsed and DNS-resolved result of a successful URL validation.
type ValidatedURL struct {
	URL        *url.URL
	Normalized string
	Hostname   string
	ResolvedIP []net.IPAddr
}

// Resolver is the DNS lookup interface used by ValidateWithResolver.
type Resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// Validate normalizes raw, checks it against the block list, and resolves its IP addresses.
func Validate(ctx context.Context, raw string) (*ValidatedURL, error) {
	return ValidateWithResolver(ctx, raw, net.DefaultResolver)
}

// ValidateWithResolver is like Validate but accepts a custom DNS resolver for testing.
func ValidateWithResolver(ctx context.Context, raw string, resolver Resolver) (*ValidatedURL, error) {
	u, normalized, err := Normalize(raw)
	if err != nil {
		return nil, err
	}

	validated := &ValidatedURL{
		URL:        u,
		Normalized: normalized,
		Hostname:   u.Hostname(),
	}

	if IsBlockedHostname(validated.Hostname) {
		return nil, fmt.Errorf("%w: %s", ErrBlockedHostname, validated.Hostname)
	}
	if IsIPLiteral(validated.Hostname) {
		return nil, fmt.Errorf("%w: %s", ErrIPLiteralNotAllowed, validated.Hostname)
	}

	resolveCtx, cancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer cancel()

	ipAddrs, err := resolver.LookupIPAddr(resolveCtx, validated.Hostname)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHostResolutionFailed, validated.Hostname)
	}
	if len(ipAddrs) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrHostResolutionFailed, validated.Hostname)
	}
	if err := ValidateResolvedIPs(ipAddrs); err != nil {
		return nil, err
	}

	validated.ResolvedIP = ipAddrs
	return validated, nil
}

// Normalize parses, validates scheme/host, and returns a canonical URL string.
func Normalize(raw string) (*url.URL, string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, "", ErrEmptyURL
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u == nil || !u.IsAbs() || u.Host == "" || u.Opaque != "" {
		return nil, "", ErrInvalidURL
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return nil, "", ErrUnsupportedScheme
	}
	if u.User != nil {
		return nil, "", ErrURLHasCredentials
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return nil, "", ErrInvalidURL
	}

	if port := u.Port(); port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else {
		u.Host = host
	}
	u.Scheme = "https"

	return u, u.String(), nil
}

// IsBlockedHostname reports whether host is a reserved or loopback name that must not be fetched.
func IsBlockedHostname(host string) bool {
	normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if normalized == "" {
		return true
	}

	if normalized == "localhost" || strings.HasSuffix(normalized, ".localhost") {
		return true
	}

	return false
}

// IsIPLiteral reports whether host is a bare IP address or bracketed IPv6 literal.
func IsIPLiteral(host string) bool {
	_, err := netip.ParseAddr(strings.Trim(host, "[]"))
	return err == nil
}

// ValidateResolvedIPs returns an error if any address in ipAddrs is blocked.
func ValidateResolvedIPs(ipAddrs []net.IPAddr) error {
	for _, ipAddr := range ipAddrs {
		if IsBlockedIP(ipAddr.IP) {
			return fmt.Errorf("%w: %s", ErrBlockedIPAddress, ipAddr.IP.String())
		}
	}
	return nil
}

// IsBlockedIP reports whether ip falls in a loopback, private, or otherwise disallowed range.
func IsBlockedIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return true
	}

	if addr.IsLoopback() || addr.IsMulticast() || addr.IsUnspecified() {
		return true
	}
	if addr.IsLinkLocalMulticast() || addr.IsLinkLocalUnicast() || addr.IsPrivate() {
		return true
	}
	if !addr.IsGlobalUnicast() {
		return true
	}

	for _, prefix := range blockedPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}

var blockedPrefixes = mustPrefixes(
	"0.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/128",
	"::1/128",
	"fe80::/10",
	"fc00::/7",
	"ff00::/8",
	"2001:db8::/32",
)

func mustPrefixes(raw ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(raw))
	for _, item := range raw {
		prefixes = append(prefixes, netip.MustParsePrefix(item))
	}
	return prefixes
}
