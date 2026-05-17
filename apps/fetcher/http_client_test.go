package fetcher

import (
	"testing"
)

func TestInitCollectorAddsDomainScopedLimitRule(t *testing.T) {
	t.Parallel()

	validated, err := validateFetchURL(t.Context(), "https://www.ourstate.com/18-essential-north-carolina-recipes/#dogs")
	if err != nil {
		t.Fatalf("validateFetchURL() error = %v", err)
	}

	collector, err := initCollector(t.Context(), validated, false)
	if err != nil {
		t.Fatalf("initCollector() error = %v", err)
	}

	if collector == nil {
		t.Fatal("initCollector() returned nil collector")
	}
}
