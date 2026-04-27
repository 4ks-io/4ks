package controllers

import (
	fetcherService "4ks/apps/api/services/fetcher"
	recipeService "4ks/apps/api/services/recipe"
	searchService "4ks/apps/api/services/search"
	staticService "4ks/apps/api/services/static"
	userService "4ks/apps/api/services/user"

	"github.com/gin-gonic/gin"
)

// RecipeController is the interface for the recipe controller
type RecipeController interface {
	BotCreateRecipe(*gin.Context)
	FetcherBotCreateRecipe(*gin.Context)
	CreateRecipe(*gin.Context)
	DeleteRecipe(*gin.Context)
	GetRecipe(*gin.Context)
	GetRecipes(*gin.Context)
	SearchRecipes(*gin.Context)
	GetRecipesByUsername(*gin.Context)
	UpdateRecipe(*gin.Context)
	ForkRecipe(*gin.Context)
	ForkRecipeRevision(*gin.Context)
	StarRecipe(*gin.Context)
	GetRecipeForks(*gin.Context)
	GetRecipeRevisions(*gin.Context)
	GetRecipeRevision(*gin.Context)
	CreateRecipeMedia(*gin.Context)
	GetRecipeMedia(*gin.Context)
	GetAdminRecipeMedias(*gin.Context)
	FetchRecipe(*gin.Context)
}

type recipeController struct {
	recipeService  recipeService.Service
	searchService  searchService.Service
	staticService  staticService.Service
	userService    userService.Service
	fetcherService fetcherService.Service
}

// NewRecipeController creates a new recipe controller
func NewRecipeController(u userService.Service, r recipeService.Service, search searchService.Service, static staticService.Service, fetcher fetcherService.Service) RecipeController {
	return &recipeController{
		userService:    u,
		recipeService:  r,
		searchService:  search,
		staticService:  static,
		fetcherService: fetcher,
	}
}
