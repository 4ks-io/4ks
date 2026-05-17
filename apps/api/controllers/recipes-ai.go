package controllers

import (
	models "4ks/libs/go/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	recipeService "4ks/apps/api/services/recipe"
	userService "4ks/apps/api/services/user"
)

type generateAIImageRequest struct {
	Prompt string `json:"prompt"`
}

// GenerateRecipeAIImage godoc
// @Summary      Generate an AI image for a recipe
// @Description  Generates an image via OpenAI and attaches it as a recipe media record
// @Tags         Recipes
// @Accept       json
// @Produce      json
// @Param        recipeID  path      string                   true  "Recipe ID"
// @Param        payload   body      generateAIImageRequest   false "Optional prompt"
// @Success      200       {object}  models.RecipeMedia
// @Router       /api/recipes/{recipeID}/ai-image [post]
// @Security     ApiKeyAuth
func (c *recipeController) GenerateRecipeAIImage(ctx *gin.Context) {
	if c.imageGenService == nil {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "image generation is not configured"})
		return
	}

	recipeID := ctx.Param("id")

	claimUserID := ctx.GetString("id")
	email := strings.ToLower(ctx.GetString("email"))
	resolvedUser, err := c.userService.GetUserByID(ctx, claimUserID)
	if err != nil {
		if err != userService.ErrUserNotFound || email == "" {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		resolvedUser, err = c.userService.GetUserByEmail(ctx, email)
		if err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}
	}
	if resolvedUser == nil {
		ctx.AbortWithError(http.StatusUnauthorized, userService.ErrUserNotFound)
		return
	}

	var payload generateAIImageRequest
	// body is optional — ignore bind errors
	_ = ctx.ShouldBindJSON(&payload)

	prompt := payload.Prompt
	if prompt == "" {
		recipe, err := c.recipeService.GetRecipeByID(ctx, recipeID)
		if err != nil {
			if err == recipeService.ErrRecipeNotFound {
				ctx.AbortWithError(http.StatusNotFound, err)
			} else {
				ctx.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}
		prompt = "A delicious recipe photo for " + recipe.CurrentRevision.Name
	}

	media, err := c.recipeService.ReserveRecipeAIImageMedia(ctx, recipeID, resolvedUser.ID)
	if err != nil {
		if err == recipeService.ErrRecipeNotFound {
			ctx.AbortWithError(http.StatusNotFound, err)
		} else if err == recipeService.ErrUnauthorized {
			ctx.AbortWithError(http.StatusForbidden, err)
		} else if err == recipeService.ErrRecipeAIImageAlreadyExists {
			ctx.AbortWithError(http.StatusConflict, err)
		} else {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	imgBytes, err := c.imageGenService.GenerateImage(ctx, prompt)
	if err != nil {
		_ = c.recipeService.UpdateRecipeMediaStatus(ctx, media.ID, models.MediaStatusErrorUnknown)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := c.recipeService.WriteRecipeMediaBytes(ctx, media, imgBytes); err != nil {
		_ = c.recipeService.UpdateRecipeMediaStatus(ctx, media.ID, models.MediaStatusErrorUnknown)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"recipeMedia": media})
}
