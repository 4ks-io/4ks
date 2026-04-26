package models

import "time"

// MediaBestUse describes the intended display context for a media asset.
type MediaBestUse int

// MediaBestUse constants.
const (
	MediaBestUseGeneral     MediaBestUse = 0
	MediaBestUseIngredient  MediaBestUse = 1
	MediaBestUseInstruction MediaBestUse = 2
)

// RecipeMediaVariant holds a single resized variant of a recipe media asset.
type RecipeMediaVariant struct {
	MaxWidth int    `firestore:"maxWidth" json:"maxWidth"`
	URL      string `firestore:"url" json:"url"`
	Filename string `firestore:"filename" json:"filename"`
	Alias    string `firestore:"alias" json:"alias"`
}

// RecipeMedia represents a media asset attached to a recipe.
type RecipeMedia struct {
	ID           string               `firestore:"id" json:"id"`
	Variants     []RecipeMediaVariant `firestore:"variants" json:"variants"`
	ContentType  string               `firestore:"contentType" json:"contentType"`
	RecipeID     string               `firestore:"recipeId" json:"recipeId"`
	RootRecipeID string               `firestore:"rootRecipeId" json:"rootRecipeId"`
	OwnerID      string               `firestore:"ownerId" json:"ownerId"`
	Status       MediaStatus          `firestore:"status" json:"status"`
	BestUse      MediaBestUse         `firestore:"bestUse" json:"bestUse"`
	CreatedDate  time.Time            `firestore:"createdDate" json:"createdDate"`
	UpdatedDate  time.Time            `firestore:"updatedDate" json:"updatedDate"`
}

// CreateRecipeMedia bundles a new RecipeMedia document with its upload signed URL.
type CreateRecipeMedia struct {
	RecipeMedia RecipeMedia `json:"recipeMedia"`
	SignedURL   string      `json:"signedURL"`
}
