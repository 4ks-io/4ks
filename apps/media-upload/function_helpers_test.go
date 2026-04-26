package function

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

func TestParseUploadObjectNameValidPath(t *testing.T) {
	t.Parallel()

	id, props, err := parseUploadObjectName("image/media-1.png")
	if err != nil {
		t.Fatalf("parseUploadObjectName returned error: %v", err)
	}
	if id != "media-1" {
		t.Fatalf("expected id media-1, got %q", id)
	}
	if props.Extension != ".png" || props.Basename != "image/media-1" {
		t.Fatalf("unexpected file props: %+v", props)
	}
}

func TestGetBoolEnv(t *testing.T) {
	cfg := RuntimeConfig{Development: true}
	if !cfg.Development {
		t.Fatal("expected development flag to stay true")
	}
}

func TestGetFilenameDetails(t *testing.T) {
	t.Parallel()

	props := getFilenameDetails("image/media-1.jpeg")
	if props.Extension != ".jpeg" || props.Basename != "image/media-1" {
		t.Fatalf("unexpected file props: %+v", props)
	}
}

func TestValidateImageConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  image.Config
		want string
	}{
		{name: "zero dimensions", cfg: image.Config{Width: 0, Height: 1}, want: "invalid image dimensions"},
		{name: "width too large", cfg: image.Config{Width: maxDecodedWidth + 1, Height: 1}, want: "dimensions exceed limit"},
		{name: "pixel count too large", cfg: image.Config{Width: 8000, Height: 6000}, want: "pixel count exceeds limit"},
		{name: "valid dimensions", cfg: image.Config{Width: 400, Height: 300}, want: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateImageConfig(tc.cfg)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestMIMEHelpers(t *testing.T) {
	t.Parallel()

	if !isSupportedMIMEType("image/png") {
		t.Fatal("expected png to be supported")
	}
	if isSupportedMIMEType("image/webp") {
		t.Fatal("did not expect webp to be supported")
	}
	if !isMIMETypeCompatible("image/jpeg", "jpeg") {
		t.Fatal("expected jpeg content type and format to be compatible")
	}
	if isMIMETypeCompatible("image/png", "jpeg") {
		t.Fatal("did not expect mismatched content type and format to be compatible")
	}
}

func TestDecodeImageForProcessingRejectsMismatchedMIME(t *testing.T) {
	t.Parallel()

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.NRGBA{G: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode returned error: %v", err)
	}

	_, _, err := decodeImageForProcessing(bytes.NewReader(buf.Bytes()), "image/jpeg")
	if err == nil || !strings.Contains(err.Error(), "does not match content type") {
		t.Fatalf("expected MIME mismatch error, got %v", err)
	}
}

func TestDecodeImageForProcessingAcceptsJPEG(t *testing.T) {
	t.Parallel()

	img := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	img.Set(1, 1, color.NRGBA{R: 255, G: 100, A: 255})

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("jpeg.Encode returned error: %v", err)
	}

	decoded, format, err := decodeImageForProcessing(bytes.NewReader(buf.Bytes()), "image/jpeg")
	if err != nil {
		t.Fatalf("decodeImageForProcessing returned error: %v", err)
	}
	if format != "jpeg" {
		t.Fatalf("expected jpeg format, got %q", format)
	}
	if decoded.Bounds().Dx() != 3 || decoded.Bounds().Dy() != 2 {
		t.Fatalf("unexpected decoded bounds: %v", decoded.Bounds())
	}
}

func TestParseUploadObjectNameRejectsMissingExtension(t *testing.T) {
	t.Parallel()

	if _, _, err := parseUploadObjectName("image/media-1"); err == nil || !strings.Contains(err.Error(), "missing a file extension") {
		t.Fatalf("expected missing extension error, got %v", err)
	}
}

func TestUpdateRecipeMediaDevelopmentNoop(t *testing.T) {
	t.Parallel()

	updater := updateRecipeMedia(context.Background(), RuntimeConfig{Development: true, FirestoreProjectID: "test"}, nil, "media-1")
	updater(MediaStatusReady)
}
