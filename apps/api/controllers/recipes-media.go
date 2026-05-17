package controllers

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"net/http"

	"github.com/gin-gonic/gin"

	"4ks/apps/api/dtos"
	recipeService "4ks/apps/api/services/recipe"
	userService "4ks/apps/api/services/user"
	"4ks/apps/api/utils"

	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

func getMediaContentType(ext string) (string, error) {
	switch ext {
	case ".png":
		return "image/png", nil
	case ".jpeg", ".jpg":
		return "image/jpeg", nil
	case ".gif":
		return "image/gif", nil
	}
	return "", fmt.Errorf("invalid File Type %s", ext)
}

// CreateRecipeMedia   godoc
// @Schemes
// @Summary 		Create a new Media SignedURL
// @Description Create a new Media SignedURL
// @Tags 				Recipes
// @Accept 			json
// @Produce 		json
// @Param       recipeID 			 path      	 string  true  "Recipe ID"
// @Param       payload 	     body      	 dtos.CreateRecipeMedia  true  "Payload"
// @Success 		200 		       {object} 	 models.CreateRecipeMedia
// @Router		 	/api/recipes/{recipeID}/media  [post]
// @Security 		ApiKeyAuth
func (c *recipeController) CreateRecipeMedia(ctx *gin.Context) {
	recipeID := ctx.Param("id")

	// The JWT claim ID (https://4ks.io/id) may not match the Firestore document ID
	// for users whose claim was issued before or after account creation. Resolve the
	// real Firestore user ID using the same ID→email fallback as GetAuthenticatedUser.
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
	userID := resolvedUser.ID

	payload := dtos.CreateRecipeMedia{}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// compute and validate file extention/content-type
	ext := filepath.Ext(payload.Filename)
	ct, err := getMediaContentType(ext)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	mp := utils.MediaProps{
		ContentType: ct,
		Extension:   ext,
		Basename:    xid.New().String(),
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// &mediaId, &mediaEmail, &payload
	signedURL, err1 := c.recipeService.CreateRecipeMediaSignedURL(ctx, &mp, &wg)
	recipeMedia, err2 := c.recipeService.CreateRecipeMedia(ctx, &mp, recipeID, userID, &wg)

	if err1 != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err1)
		return
	} else if err2 != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err2)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"recipeMedia": recipeMedia,
		"signedURL":   signedURL,
	})
}

// GetRecipeMedia godoc
// @Summary 		Get all medias for a Recipe
// @Description Get all medias for a Recipe
// @Tags 				Recipes
// @Accept 			json
// @Produce 		json
// @Param       recipeID 	path      	string  true  "Recipe ID"
// @Success 		200 		  {object} 	  dtos.GetRecipeMediaResponse
// @Router 			/api/recipes/{recipeID}/media [get]
// @Security 		ApiKeyAuth
func (c *recipeController) GetRecipeMedia(ctx *gin.Context) {
	// binding
	recipeID := ctx.Param("id")

	// data
	recipeMedias, err := c.recipeService.GetRecipeMedia(ctx, recipeID)
	if err != nil {
		log.Error().Err(err).
			Caller().
			Str("recipeID", recipeID).
			Msg("client: could not create request")

		if err == recipeService.ErrRecipeNotFound {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		} else if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, dtos.GetRecipeMediaResponse{
		Data: recipeMedias,
	})
}

// GetAdminRecipeMedias godoc
// @Summary 		Get all medias for a Recipe
// @Description Get all medias for a Recipe
// @Tags 				Admin
// @Accept 			json
// @Produce 		json
// @Param       recipeID 	path      	string  true  "Recipe ID"
// @Success 		200 		  {array} 	  models.RecipeMedia
// @Router 			/api/_admin/recipes/{recipeID}/media [get]
// @Security 		ApiKeyAuth
func (c *recipeController) GetAdminRecipeMedias(ctx *gin.Context) {
	recipeID := ctx.Param("id")
	recipeMedias, err := c.recipeService.GetAdminRecipeMedias(ctx, recipeID)

	if err == recipeService.ErrRecipeNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, recipeMedias)
}
