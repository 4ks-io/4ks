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
