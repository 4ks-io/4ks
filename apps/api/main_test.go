package main

import (
	"os"
	"testing"
)

func TestGetAPIVersion(t *testing.T) {
	t.Run("defaults when version path is unset", func(t *testing.T) {
		if got := getAPIVersion(""); got != "0.0.0" {
			t.Fatalf("expected default version, got %q", got)
		}
	})

	t.Run("reads version file when configured", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "version")
		if err != nil {
			t.Fatalf("CreateTemp: %v", err)
		}
		if _, err := file.WriteString("1.2.3\n"); err != nil {
			t.Fatalf("WriteString: %v", err)
		}
		if err := file.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		if got := getAPIVersion(file.Name()); got != "1.2.3" {
			t.Fatalf("expected file-backed version, got %q", got)
		}
	})
}

func TestConfigureLogging(_ *testing.T) {
	configureLogging()
}

func TestReadWordsFromFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "words")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := file.WriteString("alpha\nbeta\n"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	words, err := ReadWordsFromFile(file.Name())
	if err != nil {
		t.Fatalf("ReadWordsFromFile: %v", err)
	}
	if len(words) != 2 || words[0] != "alpha" || words[1] != "beta" {
		t.Fatalf("unexpected words: %#v", words)
	}
}
