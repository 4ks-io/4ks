package controllers

import (
	recipeService "4ks/apps/api/services/recipe"
	userService "4ks/apps/api/services/user"

	"net/http"

	models "4ks/libs/go/models"

	"github.com/gin-gonic/gin"
)

// GetRecipeForks godoc
// @Summary 		Get direct forks for a Recipe
// @Description Get direct child forks for a Recipe
// @Tags 				Recipes
// @Accept 			json
// @Produce 		json
// @Param       recipeID 	path      	string  true  "Recipe ID"
// @Success 		200 		{array} 	models.Recipe
// @Router 			/api/recipes/{recipeID}/forks [get]
// @Security 		ApiKeyAuth
func (c *recipeController) GetRecipeForks(ctx *gin.Context) {
	recipeID := ctx.Param("id")
	recipeForks, err := c.recipeService.GetRecipeForks(ctx, recipeID)

	if err == recipeService.ErrRecipeNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, recipeForks)
}

// ForkRecipeRevision godoc
// @Summary 		Fork Recipe Revision
// @Description Fork a specific historical recipe revision
// @Tags 				Recipes
// @Accept 			json
// @Produce 		json
// @Param       revisionID 	path      	string  true  "Revision ID"
// @Success 		200 		{object} 	models.Recipe
// @Router 			/api/recipes/revisions/{revisionID}/fork [post]
// @Security 		ApiKeyAuth
func (c *recipeController) ForkRecipeRevision(ctx *gin.Context) {
	revisionID := ctx.Param("revisionID")

	userID := ctx.GetString("id")
	author, err := c.userService.GetUserByID(ctx, userID)
	if err == userService.ErrUserNotFound {
		ctx.AbortWithError(http.StatusForbidden, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newRecipe, err := c.recipeService.ForkRecipeByRevisionID(ctx, revisionID, models.UserSummary{
		ID:          userID,
		Username:    author.Username,
		DisplayName: author.DisplayName,
	})
	if err == recipeService.ErrRecipeRevisionNotFound || err == recipeService.ErrRecipeNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = c.searchService.UpsertSearchRecipeDocument(newRecipe)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, newRecipe)
}
