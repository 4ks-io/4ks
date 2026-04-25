// Package controllers provides the controllers for the application
package controllers

import (
	recipeService "4ks/apps/api/services/recipe"
	"errors"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"4ks/apps/api/dtos"
	models "4ks/libs/go/models"
)

// BotCreateRecipe godoc
// @Schemes
// @Summary 		Bot Create a new Recipe
// @Description Bot Create a new Recipe
// @Tags 				Admin
// @Accept 			json
// @Produce 		json
// @Param       recipe   body  	   	dtos.CreateRecipe  true  "Recipe Data"
// @Success 		200 		 {object}		models.Recipe
// @Router		 	/api/_admin/recipes [post]
// @Security 		ApiKeyAuth
func (c *recipeController) BotCreateRecipe(ctx *gin.Context) {
	payload := dtos.CreateRecipe{}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	payload.Author = models.UserSummary{
		ID:          "bot",
		Username:    "4ks-bot",
		DisplayName: "4ks bot",
	}

	f, err := c.staticService.GetRandomFallbackImage(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get random fallback image")
	}
	u := c.staticService.GetRandomFallbackImageURL(f)
	payload.Banner = c.recipeService.CreateMockBanner(f, u)

	createdRecipe, err := c.recipeService.CreateRecipe(ctx, &payload)
	if err == recipeService.ErrUnableToCreateRecipe {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = c.searchService.UpsertSearchRecipeDocument(createdRecipe)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, createdRecipe)
}

// FetcherBotCreateRecipe godoc
// @Schemes
// @Summary 		Fetcher Bot Create a new Recipe
// @Description Fetcher Bot Create a new Recipe
// @Tags 				Admin
// @Accept 			json
// @Produce 		json
// @Param       data   body  	   	dtos.FetcherCreateRecipe  true  "Recipe Data"
// @Success 		200
// @Failure     400		"Invalid Request"
// @Failure     500		"Internal Error"
// @Router		 	/api/_fetcher/recipes [post]
// @Security 		ApiKeyAuth
func (c *recipeController) FetcherBotCreateRecipe(ctx *gin.Context) {
	payload := dtos.FetcherCreateRecipe{}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// handle user event
	if payload.UserID == "" || payload.UserEventID == uuid.Nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("unable to parse user event"))
		return
	}
	e := dtos.UpdateUserEvent{
		ID: payload.UserEventID,
	}

	// handle empty
	noTitle := payload.Recipe.Name == ""
	if noTitle && len(payload.Recipe.Ingredients) == 0 && len(payload.Recipe.Instructions) == 0 {
		err := errors.New("empty recipe")
		e.Status = models.UserEventErrorState
		e.Error = models.UserEventError{Message: err.Error()}
		c.userService.UpdateUserEventByUserIDEventID(ctx, payload.UserID, &e)
		log.Error().Err(err).Msg("failed to create recipe")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// set recipe author
	payload.Recipe.Author = models.UserSummary{
		ID:          "bot",
		Username:    "4ks-bot",
		DisplayName: "4ks bot",
	}

	// set recipe banner
	f, err := c.staticService.GetRandomFallbackImage(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get random fallback image")
	}
	u := c.staticService.GetRandomFallbackImageURL(f)
	payload.Recipe.Banner = c.recipeService.CreateMockBanner(f, u)

	// create recipe
	createdRecipe, err := c.recipeService.CreateRecipe(ctx, &payload.Recipe)
	if err == recipeService.ErrUnableToCreateRecipe {
		e.Status = models.UserEventErrorState
		e.Error = models.UserEventError{Message: err.Error()}
		c.userService.UpdateUserEventByUserIDEventID(ctx, payload.UserID, &e)
		log.Error().Err(err).Msg("failed to create recipe")
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	} else if err != nil {
		e.Status = models.UserEventErrorState
		e.Error = models.UserEventError{Message: err.Error()}
		c.userService.UpdateUserEventByUserIDEventID(ctx, payload.UserID, &e)
		log.Error().Err(err).Msg("failed to create recipe")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// set user event data
	e.Data = models.FetcherEventData{
		RecipeID:    createdRecipe.ID,
		RecipeTitle: createdRecipe.CurrentRevision.Name,
		URL:         payload.Recipe.Link,
	}

	// update search
	if err = c.searchService.UpsertSearchRecipeDocument(createdRecipe); err != nil {
		log.Error().Err(err).Msg("failed to update search document")
	}

	// update user event
	e.Status = models.UserEventReady
	if _, err = c.userService.UpdateUserEventByUserIDEventID(ctx, payload.UserID, &e); err != nil {
		e.Status = models.UserEventErrorState
		e.Error = models.UserEventError{Message: err.Error()}
		c.userService.UpdateUserEventByUserIDEventID(ctx, payload.UserID, &e)
		log.Error().Err(err).Msg("failed to create recipe")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}
