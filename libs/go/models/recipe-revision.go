package models

import "time"

// Image is a reference to an image asset with a URL.
type Image struct {
	ID  string `firestore:"id" json:"id"`
	URL string `firestore:"url" json:"url"`
}

// Instruction is a single step in a recipe.
type Instruction struct {
	ID   int    `firestore:"id" json:"id"`
	Type string `firestore:"type" json:"type"`
	Name string `firestore:"name" json:"name"`
	Text string `firestore:"text" json:"text"`
}

// Ingredient is a single ingredient entry in a recipe revision.
type Ingredient struct {
	ID       int    `firestore:"id" json:"id"`
	Type     string `firestore:"type" json:"type"`
	Name     string `firestore:"name" json:"name"`
	Quantity string `firestore:"quantity" json:"quantity"`
}

// RecipeRevision is a versioned snapshot of a recipe's content.
type RecipeRevision struct {
	ID           string               `firestore:"id" json:"id"`
	Name         string               `firestore:"name" json:"name"`
	Link         string               `firestore:"link" json:"link"`
	RecipeID     string               `firestore:"recipeId" json:"recipeId"`
	Author       UserSummary          `firestore:"author" json:"author"`
	Instructions []Instruction        `firestore:"instructions" json:"instructions"`
	Ingredients  []Ingredient         `firestore:"ingredients" json:"ingredients"`
	Banner       []RecipeMediaVariant `firestore:"banner" json:"banner"`
	CreatedDate  time.Time            `firestore:"createdDate" json:"createdDate"`
	UpdatedDate  time.Time            `firestore:"updatedDate" json:"updatedDate"`
}
