package dtos

// CreateSearchRecipe godoc
type CreateSearchRecipe struct {
	ID          string   `json:"id"`
	Author      string   `json:"author"`
	Name        string   `json:"name"`
	ImageURL    string   `json:"imageUrl"`
	Ingredients []string `json:"ingredients"`
}

// SearchRecipesResponse godoc
type SearchRecipesResponse struct {
	Data []*CreateSearchRecipe `json:"data"`
}
