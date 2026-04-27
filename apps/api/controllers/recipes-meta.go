package controllers

import (
	recipeService "4ks/apps/api/services/recipe"
	userService "4ks/apps/api/services/user"

	"net/http"

	"github.com/gin-gonic/gin"

	models "4ks/libs/go/models"
)

// ForkRecipe 	godoc
// @Summary 		Fork Recipe
// @Description Fork Recipe. This route accepts either an Auth0 JWT or an AI Kitchen Pass bearer token.
// @Tags 				Recipes
// @Accept 			json
// @Produce 		json
// @Param       recipeID 	path      	string  true  "Recipe ID"
// @Success 		200 		{object} 	models.Recipe
// @Router 			/api/recipes/{recipeID}/fork [post]
// @Security 		ApiKeyAuth
func (c *recipeController) ForkRecipe(ctx *gin.Context) {
	recipeID := ctx.Param("id")

	userID := ctx.GetString("id")
	author, err := c.userService.GetUserByID(ctx, userID)
	if err == userService.ErrUserNotFound {
		ctx.AbortWithError(http.StatusForbidden, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newRecipe, err := c.recipeService.ForkRecipeByID(ctx, recipeID, models.UserSummary{
		ID:          userID,
		Username:    author.Username,
		DisplayName: author.DisplayName,
	})
	if err != nil {
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

// StarRecipe		godoc
// @Summary 		Star Recipe
// @Description Star Recipe
// @Tags 				Recipes
// @Accept 			json
// @Produce 		json
// @Param       recipeID 	path      	string  true  "Recipe ID"
// @Success 		200
// @Router 			/api/recipes/{recipeID}/star [post]
// @Security 		ApiKeyAuth
func (c *recipeController) StarRecipe(ctx *gin.Context) {
	recipeID := ctx.Param("id")

	userID := ctx.GetString("id")
	author, err := c.userService.GetUserByID(ctx, userID)
	if err == userService.ErrUserNotFound {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	} else if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, err = c.recipeService.StarRecipeByID(ctx, recipeID, models.UserSummary{
		ID:          userID,
		Username:    author.Username,
		DisplayName: author.DisplayName,
	})

	if err == recipeService.ErrRecipeNotFound {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	} else if err == recipeService.ErrRecipeAlreadyStarred {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Recipe is already starred",
		})
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
