package usersvc

import (
	"crypto/rand"
	"io"
	"regexp"
	"strings"
)

const fallbackAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

var (
	reStripInvalid    = regexp.MustCompile(`[^a-z0-9-]+`)
	reCollapseHyphens = regexp.MustCompile(`-{2,}`)
)

// GenerateUsername derives a username candidate from an email address.
// It normalizes the email prefix and returns a valid-format username or a
// random "user-XXXXXX" fallback. It does NOT check uniqueness or reserved
// words — callers must use TestName and a suffix loop.
func GenerateUsername(email string) (string, error) {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return randomFallback()
	}

	prefix := trimmed
	if i := strings.IndexByte(trimmed, '@'); i >= 0 {
		prefix = trimmed[:i]
	}
	if prefix == "" {
		return randomFallback()
	}

	normalized := strings.ToLower(prefix)
	normalized = strings.ReplaceAll(normalized, ".", "-")
	normalized = reStripInvalid.ReplaceAllString(normalized, "")
	normalized = reCollapseHyphens.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")

	if len(normalized) < 8 {
		return randomFallback()
	}
	if len(normalized) > 24 {
		normalized = strings.TrimRight(normalized[:24], "-")
	}

	return normalized, nil
}

// randomFallback returns "user-" followed by 6 cryptographically random
// lowercase alphanumeric characters.
func randomFallback() (string, error) {
	buf := make([]byte, 6)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	result := make([]byte, 6)
	for i, b := range buf {
		result[i] = fallbackAlphabet[int(b)%len(fallbackAlphabet)]
	}
	return "user-" + string(result), nil
}
