// Package models defines shared domain types used across 4ks services.
package models

import "github.com/google/uuid"

// FetcherRequest is a struct to hold the fetcher request data
type FetcherRequest struct {
	URL         string    `json:"url"`
	UserID      string    `json:"userId"`
	UserEventID uuid.UUID `json:"userEventId"`
}

// FetcherResponse holds the recipe data returned by the fetcher service.
type FetcherResponse struct {
	Name         string   `json:"name"`
	Link         string   `json:"link"`
	Instructions []string `json:"instructions"`
	Ingredients  []string `json:"ingredients"`
}

// FetcherEventData is the payload stored in a UserEvent after a recipe fetch completes.
type FetcherEventData struct {
	RecipeID    string `json:"recipeId"`
	RecipeTitle string `json:"recipeTitle"`
	URL         string `json:"url"`
}
