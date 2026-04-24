package fetcher

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

var (
	errEmptyURL             = errors.New("url is required")
	errInvalidURL           = errors.New("url must be an absolute https URL")
	errUnsupportedScheme    = errors.New("url scheme must be https")
	errURLHasCredentials    = errors.New("url must not contain embedded credentials")
	errBlockedHostname      = errors.New("url host is not allowed")
	errBlockedIPAddress     = errors.New("url resolves to a blocked address")
	errIPLiteralNotAllowed  = errors.New("ip literal targets are not allowed")
	errHostResolutionFailed = errors.New("url host could not be resolved")
)

type validatedURL struct {
	URL        *url.URL
	Normalized string
	Hostname   string
	ResolvedIP []net.IPAddr
}

func validateFetchURL(ctx context.Context, raw string) (*validatedURL, error) {
	u, normalized, err := normalizeFetchURL(raw)
	if err != nil {
		return nil, err
	}

	validated := &validatedURL{
		URL:        u,
		Normalized: normalized,
		Hostname:   u.Hostname(),
	}

	if isBlockedHostname(validated.Hostname) {
		return nil, fmt.Errorf("%w: %s", errBlockedHostname, validated.Hostname)
	}
	if isIPLiteral(validated.Hostname) {
		return nil, fmt.Errorf("%w: %s", errIPLiteralNotAllowed, validated.Hostname)
	}

	resolveCtx, cancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer cancel()

	ipAddrs, err := net.DefaultResolver.LookupIPAddr(resolveCtx, validated.Hostname)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errHostResolutionFailed, validated.Hostname)
	}
	if len(ipAddrs) == 0 {
		return nil, fmt.Errorf("%w: %s", errHostResolutionFailed, validated.Hostname)
	}
	if err := validateResolvedIPs(ipAddrs); err != nil {
		return nil, err
	}

	validated.ResolvedIP = ipAddrs
	return validated, nil
}

func normalizeFetchURL(raw string) (*url.URL, string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, "", errEmptyURL
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", errInvalidURL, err)
	}
	if u == nil || !u.IsAbs() || u.Host == "" || u.Opaque != "" {
		return nil, "", errInvalidURL
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return nil, "", errUnsupportedScheme
	}
	if u.User != nil {
		return nil, "", errURLHasCredentials
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return nil, "", errInvalidURL
	}

	if port := u.Port(); port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else {
		u.Host = host
	}
	u.Scheme = "https"

	return u, u.String(), nil
}

func isBlockedHostname(host string) bool {
	normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if normalized == "" {
		return true
	}

	return normalized == "localhost" || strings.HasSuffix(normalized, ".localhost")
}

func isIPLiteral(host string) bool {
	_, err := netip.ParseAddr(strings.Trim(host, "[]"))
	return err == nil
}

func validateResolvedIPs(ipAddrs []net.IPAddr) error {
	for _, ipAddr := range ipAddrs {
		if isBlockedIP(ipAddr.IP) {
			return fmt.Errorf("%w: %s", errBlockedIPAddress, ipAddr.IP.String())
		}
	}
	return nil
}

func isBlockedIP(ip net.IP) bool {
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
