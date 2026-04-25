// package function is the entrypoint for the Cloud Function that validates and resizes images
package function

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	firestore "cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/cloudevents/sdk-go/v2/event"
)

func updateRecipeMedia(cfg RuntimeConfig, c *firestore.CollectionRef, ctx context.Context, id string) func(MediaStatus) {
	return func(s MediaStatus) {
		if cfg.Development {
			log.Printf("mock action: update status %s (%d): %s", id, s, cfg.FirestoreProjectID)
			return
		}
		_, err := c.Doc(id).Update(ctx, []firestore.Update{
			{
				Path:  "updatedDate",
				Value: time.Now().UTC(),
			},
			{
				Path:  "status",
				Value: s,
			},
		})
		if err != nil {
			log.Fatalf("error updating recipe-medias %s", err)
		}
		log.Printf("update status %s (%d): %s", id, s, cfg.FirestoreProjectID)
	}
}

// newUploadImageHandler creates size variants of an uploaded image.
func newUploadImageHandler(cfg RuntimeConfig) func(context.Context, event.Event) error {
	return func(ctx context.Context, e event.Event) error {
		var s, _ = firestore.NewClient(ctx, cfg.FirestoreProjectID)
		var c = s.Collection("recipe-medias")

		// init
		var data StorageObjectData
		if err := e.DataAs(&data); err != nil {
			return fmt.Errorf("event.DataAs: %v", err)
		}

		id, f, err := parseUploadObjectName(data.Name)
		if err != nil {
			if id != "" {
				updateRecipeMedia(cfg, c, ctx, id)(MediaStatusErrorUnknown)
			}
			log.Printf("upload: rejecting malformed object name %q: %v", data.Name, err)
			return err
		}

		log.Printf("Processing (%s) gs://%s/%s", id, data.Bucket, data.Name)

		// update status
		var up = updateRecipeMedia(cfg, c, ctx, id)
		up(MediaStatusProcessing)

		// storage client
		client, err := storage.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("upload: failed to construct a client, error = %v", err)
		}
		defer client.Close()

		// src ObjectHandle
		src := client.Bucket(data.Bucket).Object(data.Name)
		// dst BucketHandle
		dstbkt := client.Bucket(cfg.DistributionBucket)

		attrs, err := src.Attrs(ctx)
		if err != nil {
			up(MediaStatusErrorMissingAttr)
			return fmt.Errorf("upload: failed to get object attributes %q: %w", data.Name, err)
		}

		// verify and validate src content-type and content (image vision)
		if err, status := validate(ctx, src, attrs); err != nil {
			up(status)
			return err
		}

		// read and decode src image
		rc, err := src.NewReader(ctx)
		if err != nil {
			up(MediaStatusErrorFailedResize)
			return fmt.Errorf("unable to read file %s in %s (%v)", data.Name, cfg.DistributionBucket, err)
		}
		i, ifmt, err := decodeImageForProcessing(rc, attrs.ContentType)
		rc.Close()
		if err != nil {
			up(MediaStatusErrorFailedResize)
			return fmt.Errorf("unable to decode image: %v", err)
		}

		// create variants
		variants := []int{256, 1024}
		var wg sync.WaitGroup
		for _, s := range variants {
			wg.Add(1)
			o, err := createVariant(ctx, dstbkt, i, ifmt, f, s)
			wg.Done()
			if err != nil {
				up(MediaStatusErrorFailedVariant)
				return fmt.Errorf("failed to create %s variant %d: %v", o, s, err)
			}
		}

		up(MediaStatusReady)

		// delete src file
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if err := src.Delete(ctx); err != nil {
			return fmt.Errorf("Object(%q).Delete: %v", data.Name, err)
		}

		// terminate ctx
		if err := client.Close(); err != nil {
			return fmt.Errorf("client.Close: %v", err)
		}

		return nil
	}
}

func parseUploadObjectName(name string) (string, FileProps, error) {
	ext := filepath.Ext(name)
	if ext == "" {
		return "", FileProps{}, errors.New("object name is missing a file extension")
	}

	base := strings.TrimSuffix(name, ext)
	parts := strings.Split(base, "/")
	if len(parts) != 2 || parts[0] != "image" || parts[1] == "" {
		candidateID := ""
		if len(parts) >= 2 {
			candidateID = parts[1]
		}
		return candidateID, FileProps{}, fmt.Errorf("object name must match image/<id>%s", ext)
	}

	return parts[1], FileProps{Extension: ext, Basename: base}, nil
}

func decodeImageForProcessing(r io.Reader, contentType string) (image.Image, string, error) {
	slurp, format, err := loadValidatedImageBytes(r, contentType)
	if err != nil {
		return nil, "", err
	}

	i, ifmt, err := image.Decode(bytes.NewReader(slurp))
	if err != nil {
		return nil, "", err
	}
	if !isMIMETypeCompatible(contentType, ifmt) {
		return nil, "", fmt.Errorf("upload: decoded image format %q does not match content type %q", ifmt, contentType)
	}

	return i, format, nil
}
