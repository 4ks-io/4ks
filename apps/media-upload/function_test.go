package function

import "testing"

func TestParseUploadObjectName(t *testing.T) {
	t.Parallel()

	id, props, err := parseUploadObjectName("image/abc123.jpg")
	if err != nil {
		t.Fatalf("parseUploadObjectName returned error: %v", err)
	}
	if id != "abc123" {
		t.Fatalf("expected id abc123, got %q", id)
	}
	if props.Basename != "image/abc123" {
		t.Fatalf("expected basename image/abc123, got %q", props.Basename)
	}
	if props.Extension != ".jpg" {
		t.Fatalf("expected extension .jpg, got %q", props.Extension)
	}
}

func TestParseUploadObjectNameRejectsMalformedPath(t *testing.T) {
	t.Parallel()

	id, _, err := parseUploadObjectName("abc123.jpg")
	if err == nil {
		t.Fatal("expected error for malformed object name")
	}
	if id != "" {
		t.Fatalf("expected empty candidate id, got %q", id)
	}
}
