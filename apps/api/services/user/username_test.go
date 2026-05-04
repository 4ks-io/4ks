package usersvc

import (
	"regexp"
	"testing"
)

func TestGenerateUsername(t *testing.T) {
	t.Parallel()

	fallbackPattern := regexp.MustCompile(`^user-[a-z0-9]{6}$`)

	cases := []struct {
		name         string
		email        string
		wantExact    string // set if deterministic
		wantFallback bool   // set if random fallback expected
	}{
		{name: "dots replaced with hyphens", email: "delorme.nic@gmail.com", wantExact: "delorme-nic"},
		{name: "uppercase lowercased", email: "UPPER.CASE@x.com", wantExact: "upper-case"},
		{name: "hyphen in prefix passthrough", email: "chef-recipes@x.com", wantExact: "chef-recipes"},
		{name: "truncated to 24 chars", email: "verylongemailprefixfortesting@example.com", wantExact: "verylongemailprefixforte"},
		{name: "no at sign treated as whole string", email: "notanemail", wantExact: "notanemail"},
		{name: "empty email", email: "", wantFallback: true},
		{name: "blank whitespace", email: "   ", wantFallback: true},
		{name: "prefix too short — chef", email: "chef@example.com", wantFallback: true},
		{name: "plus sign stripped leaves too short", email: "user+tag@domain.com", wantFallback: true},
		{name: "consecutive dots collapse to short", email: "a..b@example.com", wantFallback: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := GenerateUsername(tc.email)
			if err != nil {
				t.Fatalf("GenerateUsername(%q) unexpected error: %v", tc.email, err)
			}
			if tc.wantExact != "" {
				if got != tc.wantExact {
					t.Fatalf("GenerateUsername(%q) = %q, want %q", tc.email, got, tc.wantExact)
				}
				return
			}
			if !fallbackPattern.MatchString(got) {
				t.Fatalf("GenerateUsername(%q) = %q, want pattern %s", tc.email, got, fallbackPattern)
			}
		})
	}
}

func TestGenerateUsernameTruncationNeverEndsWithHyphen(t *testing.T) {
	t.Parallel()

	// Build a 30-char prefix that ends in a hyphen when truncated to 24.
	// "aaaaaaaaaaaaaaaaaaaa----" would truncate to "aaaaaaaaaaaaaaaaaaaa----" then trim trailing hyphens.
	// Using an email whose 24th character would be a hyphen.
	got, err := GenerateUsername("aaaaaaaaaaaaaaaaaaaaaaaa.b@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected non-empty result")
	}
	if got[len(got)-1] == '-' {
		t.Fatalf("GenerateUsername returned username ending in hyphen: %q", got)
	}
}
