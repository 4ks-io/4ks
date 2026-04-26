package models

import "time"

// RecipeMetadata holds aggregate counters for a recipe.
type RecipeMetadata struct {
	Stars int32 `firestore:"stars" json:"stars"`
	Forks int32 `firestore:"forks" json:"forks"`
}

// Recipe is the top-level Firestore document for a recipe.
type Recipe struct {
	ID              string         `firestore:"id" json:"id"`
	Author          UserSummary    `firestore:"author" json:"author"`
	Contributors    []UserSummary  `firestore:"contributors" json:"contributors"`
	CurrentRevision RecipeRevision `firestore:"currentRevision" json:"currentRevision"`
	Metadata        RecipeMetadata `firestore:"metadata" json:"metadata"`
	Root            string         `firestore:"root" json:"root"`
	Branch          string         `firestore:"branch" json:"branch"`
	CreatedDate     time.Time      `firestore:"createdDate" json:"createdDate"`
	UpdatedDate     time.Time      `firestore:"updatedDate" json:"updatedDate"`
}

// RecipeSummary is a compact reference to a recipe used in nested documents.
type RecipeSummary struct {
	ID   string `firestore:"id" json:"id"`
	Name string `firestore:"name" json:"name"`
}
