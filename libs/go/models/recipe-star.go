package models

import "time"

// RecipeStar records that a user has starred a recipe.
type RecipeStar struct {
	User        UserSummary   `firestore:"user" json:"user"`
	Recipe      RecipeSummary `firestore:"recipe" json:"recipe"`
	CreatedDate time.Time     `firestore:"createdDate,omitempty" json:"createdDate"`
	UpdatedDate time.Time     `firestore:"updatedDate,omitempty" json:"updatedDate"`
}
