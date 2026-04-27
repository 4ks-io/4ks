package controllers

import (
	recipeService "4ks/apps/api/services/recipe"

	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRecipeRevisions godoc
// @Summary 		Get all revisions for a Recipe
// @Description Get all revisions for a Recipe. This route is part of the AI Kitchen Pass workflow surface.
// @Tags 				Recipes
// @Accept 			json
// @Produce 		json
// @Param       recipeID 	path      	string  true  "Recipe ID"
// @Success 		200 		{array} 	models.RecipeRevision
// @Router 			/api/recipes/{recipeID}/revisions [get]
// @Security 		ApiKeyAuth
func (c *recipeController) GetRecipeRevisions(ctx *gin.Context) {
	recipeID := ctx.Param("id")
	recipeRevisions, err := c.recipeService.GetRecipeRevisions(ctx, recipeID)

	if err == recipeService.ErrRecipeNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, recipeRevisions)
}

// GetRecipeRevision godoc
// @Summary 		Get a Recipe Revision
// @Description Get a Revision By ID
// @Tags 				Recipes
// @Accept 			json
// @Produce 		json
// @Param       revisionID 	path      	string  true  "Revision ID"
// @Success 		200 		{object} 	models.RecipeRevision
// @Router 			/api/recipes/revisions/{revisionID} [get]
// @Security 		ApiKeyAuth
func (c *recipeController) GetRecipeRevision(ctx *gin.Context) {
	revisionID := ctx.Param("revisionID")
	recipeRevision, err := c.recipeService.GetRecipeRevisionByID(ctx, revisionID)

	if err == recipeService.ErrRecipeRevisionNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, recipeRevision)
}
