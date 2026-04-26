package function

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"

	"cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/v2/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
)

var errRetryable = errors.New("upload: retryable error")

const (
	maxUploadBytes   int64 = 6 * 1024 * 1024
	maxDecodedWidth        = 8192
	maxDecodedHeight       = 8192
	maxDecodedPixels int64 = 40 * 1024 * 1024
)

func validate(ctx context.Context, obj *storage.ObjectHandle, attrs *storage.ObjectAttrs) (MediaStatus, error) {
	if err := validateObjectSize(attrs.Size); err != nil {
		return MediaStatusErrorSize, err
	}
	if err := validateMIMEType(ctx, attrs.ContentType, obj); err != nil {
		return MediaStatusErrorInvalidMIMEType, err
	}
	// Validates obj by calling Vision API.
	return validateByVisionAPI(ctx, obj)
}

func validateObjectSize(size int64) error {
	if size > maxUploadBytes {
		return fmt.Errorf("upload: image file is too large, got = %d", size)
	}
	return nil
}

func validateMIMEType(ctx context.Context, contentType string, obj *storage.ObjectHandle) error {
	r, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("upload: failed to open new file %q : %w",
			obj.ObjectName(), errRetryable)
	}
	defer r.Close()
	_, _, err = loadValidatedImageBytes(r, contentType)
	return err
}

func loadValidatedImageBytes(r io.Reader, contentType string) ([]byte, string, error) {
	if !isSupportedMIMEType(contentType) {
		return nil, "", fmt.Errorf("upload: unsupported MIME type, got = %q", contentType)
	}

	slurp, err := io.ReadAll(io.LimitReader(r, maxUploadBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("upload: failed to read image: %w", err)
	}
	if err := validateObjectSize(int64(len(slurp))); err != nil {
		return nil, "", err
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(slurp))
	if err != nil {
		return nil, "", fmt.Errorf("upload: failed to inspect image: %w", err)
	}
	if !isMIMETypeCompatible(contentType, format) {
		return nil, "", fmt.Errorf("upload: decoded image format %q does not match content type %q", format, contentType)
	}
	if err := validateImageConfig(cfg); err != nil {
		return nil, "", err
	}

	return slurp, format, nil
}

func validateImageConfig(cfg image.Config) error {
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return fmt.Errorf("upload: invalid image dimensions %dx%d", cfg.Width, cfg.Height)
	}
	if cfg.Width > maxDecodedWidth || cfg.Height > maxDecodedHeight {
		return fmt.Errorf("upload: image dimensions exceed limit, got = %dx%d", cfg.Width, cfg.Height)
	}
	if int64(cfg.Width)*int64(cfg.Height) > maxDecodedPixels {
		return fmt.Errorf("upload: image pixel count exceeds limit, got = %d", int64(cfg.Width)*int64(cfg.Height))
	}
	return nil
}

func isSupportedMIMEType(contentType string) bool {
	switch contentType {
	case "image/png", "image/jpeg", "image/jpg", "image/gif":
		return true
	default:
		return false
	}
}

func isMIMETypeCompatible(contentType string, format string) bool {
	switch contentType {
	case "image/png":
		return format == "png"
	case "image/jpeg", "image/jpg":
		return format == "jpeg"
	case "image/gif":
		return format == "gif"
	default:
		return false
	}
}

// validateByVisionAPI uses Safe Search Detection provided by Cloud Vision API.
// See more details: https://cloud.google.com/vision/docs/detecting-safe-search
func validateByVisionAPI(ctx context.Context, obj *storage.ObjectHandle) (MediaStatus, error) {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return MediaStatusErrorVision, fmt.Errorf(
			"upload: failed to create a ImageAnnotator client, error = %v : %w",
			err,
			errRetryable,
		)
	}
	defer client.Close()

	resp, err := client.BatchAnnotateImages(ctx, &visionpb.BatchAnnotateImagesRequest{
		Requests: []*visionpb.AnnotateImageRequest{
			{
				Image: &visionpb.Image{
					Source: &visionpb.ImageSource{
						ImageUri: fmt.Sprintf("gs://%s/%s", obj.BucketName(), obj.ObjectName()),
					},
				},
				Features: []*visionpb.Feature{
					{
						Type: visionpb.Feature_SAFE_SEARCH_DETECTION,
					},
				},
			},
		},
	})
	if err != nil {
		return MediaStatusErrorSafeSearch, fmt.Errorf(
			"upload: failed to detect safe search, error = %v : %w",
			err,
			errRetryable,
		)
	}
	if len(resp.GetResponses()) != 1 || resp.GetResponses()[0].GetSafeSearchAnnotation() == nil {
		return MediaStatusErrorSafeSearch, errors.New("upload: safe search response missing annotation")
	}

	ssa := resp.GetResponses()[0].GetSafeSearchAnnotation()
	// Returns an unretryable error if there is any possibility of inappropriate image.
	if ssa.Adult >= visionpb.Likelihood_POSSIBLE {
		return MediaStatusErrorInappropriateAdult, errors.New("upload: exceeds the prescribed likelihood (adult)")
	}
	if ssa.Medical >= visionpb.Likelihood_POSSIBLE {
		return MediaStatusErrorInappropriateMedical, errors.New("upload: exceeds the prescribed likelihood (medical)")
	}
	if ssa.Violence >= visionpb.Likelihood_POSSIBLE {
		return MediaStatusErrorInappropriateViolence, errors.New("upload: exceeds the prescribed likelihood (violence)")
	}
	if ssa.Racy >= visionpb.Likelihood_POSSIBLE {
		return MediaStatusErrorInappropriateRacy, errors.New("upload: exceeds the prescribed likelihood")
	}
	return MediaStatusProcessing, nil
}
