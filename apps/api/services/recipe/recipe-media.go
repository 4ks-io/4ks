package recipesvc

import (
	"4ks/apps/api/middleware"
	"4ks/apps/api/utils"
	models "4ks/libs/go/models"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	firestore "cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s recipeService) CreateRecipeMedia(ctx context.Context, mp *utils.MediaProps, recipeID string, userID string, wg *sync.WaitGroup) (*models.RecipeMedia, error) {
	defer wg.Done()

	recipeDoc, err := s.recipeCollection.Doc(recipeID).Get(ctx)
	if err != nil {
		return nil, ErrRecipeNotFound
	}

	recipe := new(models.Recipe)
	recipeDoc.DataTo(recipe)

	e, err := middleware.EnforceContributor(userID, recipe.Contributors)
	if err != nil {
		return nil, ErrUnableToCreateRecipeMedia
	} else if !e {
		return nil, ErrUnauthorized
	}

	newRecipeMediaDoc := s.recipeMediasCollection.NewDoc()
	timestamp := time.Now().UTC()

	a := []models.RecipeMediaVariant{}
	a = append(a, models.RecipeMediaVariant{
		MaxWidth: 256,
		URL:      fmt.Sprintf("%s/%s", s.imageURL, mp.Basename+"_256"+mp.Extension),
		Filename: mp.Basename + "_256" + mp.Extension,
		Alias:    "sm",
	})
	a = append(a, models.RecipeMediaVariant{
		MaxWidth: 1024,
		URL:      fmt.Sprintf("%s/%s", s.imageURL, mp.Basename+"_1024"+mp.Extension),
		Filename: mp.Basename + "_1024" + mp.Extension,
		Alias:    "md",
	})

	// https://github.com/4ks-io/4ks-monorepo/blob/f4f12c2f7eb4c6dc671b6b58dcafbeaf5702eeb8/apps/media-upload/function.go
	recipeMedia := &models.RecipeMedia{
		ID:           newRecipeMediaDoc.ID,
		Variants:     a,
		ContentType:  mp.ContentType,
		RecipeID:     recipe.ID,
		RootRecipeID: recipe.Root,
		OwnerID:      userID,
		Status:       models.MediaStatusRequested,
		BestUse:      models.MediaBestUseGeneral,
		Source:       models.MediaSourceUpload,
		CreatedDate:  timestamp,
		UpdatedDate:  timestamp,
	}

	_, err = s.recipeMediasCollection.Doc(mp.Basename).Create(ctx, recipeMedia)
	if err != nil {
		return nil, err
	}

	return recipeMedia, nil
}

func (s recipeService) CreateRecipeMediaSignedURL(_ context.Context, mp *utils.MediaProps, wg *sync.WaitGroup) (string, error) {
	defer wg.Done()

	// https://pkg.go.dev/cloud.google.com/go/storage#SignedURLOptions
	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		GoogleAccessID: s.serviceAccountName,
		Method:         "PUT",
		Expires:        time.Now().Add(expirationMinutes * time.Minute),
		// ContentType:    mp.ct,
	}

	filename := "image/" + mp.Basename + mp.Extension
	url, err := s.store.Bucket(s.uploadableBucket).SignedURL(filename, opts)
	if err != nil {
		return "", fmt.Errorf("Bucket(%q). SignedURL: %v", s.uploadableBucket, err)
	}

	return url, nil
}

// GetRecipeMedia retreives recipe media
func (s recipeService) GetRecipeMedia(ctx context.Context, recipeID string) ([]*models.RecipeMedia, error) {
	var status [2]int
	status[0] = int(models.MediaStatusReady)

	// workaround to see images locally; upload-media status update callback only works in hosted firestore
	if s.sysFlags.Development {
		status[1] = int(models.MediaStatusRequested)
	}

	recipeMediasDocs, err := s.recipeMediasCollection.
		Where("rootRecipeId", "==", recipeID).
		Where("status", "in", status).
		OrderBy("createdDate", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		log.Error().Err(err).Msg("failed to get recipe media")
		return nil, err
	}

	numberOfMedias := len(recipeMediasDocs)
	if numberOfMedias == 0 {
		recipeMedias := make([]*models.RecipeMedia, 0)
		return recipeMedias, nil
	}

	recipeMedias := make([]*models.RecipeMedia, numberOfMedias)
	for i, ds := range recipeMediasDocs {
		recipeMedia := new(models.RecipeMedia)
		ds.DataTo(recipeMedia)
		recipeMedias[i] = recipeMedia
	}
	return recipeMedias, nil
}

func (s recipeService) GetAdminRecipeMedias(ctx context.Context, recipeID string) ([]*models.RecipeMedia, error) {
	recipeMediasDocs, err := s.recipeMediasCollection.Where("rootRecipeId", "==", recipeID).Documents(ctx).GetAll()

	if err != nil {
		return nil, err
	}

	numberOfMedias := len(recipeMediasDocs)
	if numberOfMedias == 0 {
		recipeMedias := make([]*models.RecipeMedia, 0)
		return recipeMedias, nil
	}

	recipeMedias := make([]*models.RecipeMedia, numberOfMedias)
	for i, ds := range recipeMediasDocs {
		recipeMedia := new(models.RecipeMedia)
		ds.DataTo(recipeMedia)
		recipeMedias[i] = recipeMedia
	}

	return recipeMedias, nil
}

func isBlockingAIImageStatus(s models.MediaStatus) bool {
	return s == models.MediaStatusRequested ||
		s == models.MediaStatusProcessing ||
		s == models.MediaStatusReady
}

func aiRecipeMediaID(rootRecipeID string) string {
	return "ai-" + rootRecipeID
}

func recipeRootID(recipe *models.Recipe) string {
	if recipe.Root != "" {
		return recipe.Root
	}
	return recipe.ID
}

func (s recipeService) ReserveRecipeAIImageMedia(
	ctx context.Context,
	recipeID string,
	userID string,
) (*models.RecipeMedia, error) {
	var reserved *models.RecipeMedia

	err := s.fire.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		recipeDocRef := s.recipeCollection.Doc(recipeID)
		recipeDoc, err := tx.Get(recipeDocRef)
		if err != nil {
			return ErrRecipeNotFound
		}

		recipe := new(models.Recipe)
		if err := recipeDoc.DataTo(recipe); err != nil {
			return err
		}

		e, err := middleware.EnforceContributor(userID, recipe.Contributors)
		if err != nil {
			return ErrUnableToCreateRecipeMedia
		} else if !e {
			return ErrUnauthorized
		}

		rootRecipeID := recipeRootID(recipe)
		mediaID := aiRecipeMediaID(rootRecipeID)
		mediaDocRef := s.recipeMediasCollection.Doc(mediaID)
		existingDoc, err := tx.Get(mediaDocRef)
		if err == nil && existingDoc.Exists() {
			existing := new(models.RecipeMedia)
			if err := existingDoc.DataTo(existing); err != nil {
				return err
			}
			if existing.Source == models.MediaSourceAI && isBlockingAIImageStatus(existing.Status) {
				return ErrRecipeAIImageAlreadyExists
			}
		} else if err != nil && status.Code(err) != codes.NotFound {
			return err
		}

		ext := ".png"
		ct := "image/png"
		timestamp := time.Now().UTC()
		variants := []models.RecipeMediaVariant{
			{
				MaxWidth: 256,
				URL:      fmt.Sprintf("%s/%s", s.imageURL, mediaID+"_256"+ext),
				Filename: mediaID + "_256" + ext,
				Alias:    "sm",
			},
			{
				MaxWidth: 1024,
				URL:      fmt.Sprintf("%s/%s", s.imageURL, mediaID+"_1024"+ext),
				Filename: mediaID + "_1024" + ext,
				Alias:    "md",
			},
		}

		reserved = &models.RecipeMedia{
			ID:           mediaID,
			Variants:     variants,
			ContentType:  ct,
			RecipeID:     recipe.ID,
			RootRecipeID: rootRecipeID,
			OwnerID:      userID,
			Status:       models.MediaStatusRequested,
			BestUse:      models.MediaBestUseGeneral,
			Source:       models.MediaSourceAI,
			CreatedDate:  timestamp,
			UpdatedDate:  timestamp,
		}

		return tx.Set(mediaDocRef, reserved)
	})
	if err != nil {
		return nil, err
	}

	return reserved, nil
}

func (s recipeService) UpdateRecipeMediaStatus(ctx context.Context, mediaID string, mediaStatus models.MediaStatus) error {
	_, err := s.recipeMediasCollection.Doc(mediaID).Update(ctx, []firestore.Update{
		{Path: "updatedDate", Value: time.Now().UTC()},
		{Path: "status", Value: mediaStatus},
	})
	return err
}

func (s recipeService) WriteRecipeMediaBytes(
	ctx context.Context,
	media *models.RecipeMedia,
	data []byte,
) error {
	objectName := "image/" + media.ID + ".png"
	wc := s.store.Bucket(s.uploadableBucket).Object(objectName).NewWriter(ctx)
	wc.ContentType = media.ContentType
	if _, err := wc.Write(data); err != nil {
		_ = wc.Close()
		return fmt.Errorf("GCS write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("GCS close: %w", err)
	}

	return nil
}

// CreateRecipeMediaFromBytes uploads AI-generated image bytes directly to GCS
// and creates the corresponding Firestore RecipeMedia record.
// Firestore is written first so the media-upload Cloud Function finds the doc
// when the GCS write triggers it.
func (s recipeService) CreateRecipeMediaFromBytes(
	ctx context.Context,
	data []byte,
	recipeID string,
	userID string,
) (*models.RecipeMedia, error) {
	recipeMedia, err := s.ReserveRecipeAIImageMedia(ctx, recipeID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.WriteRecipeMediaBytes(ctx, recipeMedia, data); err != nil {
		_ = s.UpdateRecipeMediaStatus(ctx, recipeMedia.ID, models.MediaStatusErrorUnknown)
		return nil, err
	}

	return recipeMedia, nil
}
