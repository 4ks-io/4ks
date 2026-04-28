package controllers

import (
	"4ks/apps/api/dtos"
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	usersvc "4ks/apps/api/services/user"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// UserController is the interface for the user controller
type UserController interface {
	CreateUser(*gin.Context)
	HeadAuthenticatedUser(*gin.Context)
	GetAuthenticatedUser(*gin.Context)
	GetUser(*gin.Context)
	GetUsers(*gin.Context)
	DeleteUser(*gin.Context)
	UpdateUser(*gin.Context)
	GetKitchenPass(*gin.Context)
	CreateKitchenPass(*gin.Context)
	DeleteKitchenPass(*gin.Context)
	TestUsername(*gin.Context)
	RemoveUserEvent(*gin.Context)
}

type userController struct {
	usersvc        usersvc.Service
	kitchenPasssvc kitchenpasssvc.Service
}

// NewUserController creates a new user controller
func NewUserController(u usersvc.Service, kitchenPass kitchenpasssvc.Service) UserController {
	return &userController{
		usersvc:        u,
		kitchenPasssvc: kitchenPass,
	}
}

// RemoveUserEvent	godoc
// @Summary 		    Delete User
// @Description     Delete User
// @Tags 				    Users
// @Accept 			    json
// @Produce 		    json
// @Param           id 	path      string  true  "Event ID"
// @Success 		    200
// @Router 			    /api/user/events/{id}   [delete]
// @Security 		    ApiKeyAuth
func (c *userController) RemoveUserEvent(ctx *gin.Context) {
	userID := ctx.GetString("id")
	eventID := ctx.Param("id")

	// validate eventID
	log.Debug().Str("eventID", eventID).Msg("validating eventID")
	eid, err := uuid.Parse(eventID)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = c.usersvc.RemoveUserEventByUserIDEventID(ctx, userID, eid)
	if err == usersvc.ErrUserEventNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

// CreateUser   godoc
// @Schemes
// @Summary 		Create a new User
// @Description Create a new User
// @Tags 				Users
// @Accept 			json
// @Produce 		json
// @Param       user     body  	   dtos.CreateUser  true  "User Data"
// @Success 		200 		 {object} 	 models.User
// @Router		 	/api/user   [post]
// @Security 		ApiKeyAuth
func (c *userController) CreateUser(ctx *gin.Context) {
	userID := ctx.GetString("id")
	userEmail := ctx.GetString("email")

	payload := dtos.CreateUser{}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	createdUser, err := c.usersvc.CreateUser(ctx, userID, userEmail, &payload)
	if err == usersvc.ErrEmailInUse || err == usersvc.ErrUsernameInUse {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, createdUser)
}

// DeleteUser		godoc
// @Summary 		Delete User
// @Description Delete User
// @Tags 				Users
// @Accept 			json
// @Produce 		json
// @Param       userID 	path      string  true  "User ID"
// @Success 		200
// @Router 			/api/users/{userID}   [delete]
// @Security 		ApiKeyAuth
func (c *userController) DeleteUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	err := c.usersvc.DeleteUser(ctx, userID)

	if err == usersvc.ErrUserNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

// GetUser  		godoc
// @Schemes
// @Summary 	  Get a User (by ID)
// @Description Get a User (by ID)
// @Tags 		    Users
// @Accept 	   	json
// @Produce   	json
// @Param       id 	    path      string  true  "User ID"
// @Success 		200 		{object} 	models.User
// @Router 			/api/users/{id} [get]
// @Security 		ApiKeyAuth
func (c *userController) GetUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	user, err := c.usersvc.GetUserByID(ctx, userID)

	if err == usersvc.ErrUserNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetAuthenticatedUser	godoc
// @Schemes
// @Summary 	  Get Authenticated User
// @Description Get Authenticated User
// @Tags 		    Users
// @Accept 	   	json
// @Produce   	json
// @Success 		200 		{object} 	models.User
// @Router 			/api/user/ [get]
// @Security 		ApiKeyAuth
func (c *userController) GetAuthenticatedUser(ctx *gin.Context) {
	userID := ctx.GetString("id")
	user, err := c.usersvc.GetUserByID(ctx, userID)

	if err == usersvc.ErrUserNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// HeadAuthenticatedUser godoc
// @Schemes
// @Summary 	  Head Authenticated user
// @Description Head Authenticated user
// @Router 			/api/user/ [head]
// @Tags 		    Users
// @Produce   	json
// @Success 		200
// @Success 		204   "No Content"
// @Failure     400		"Invalid Request"
// @Failure     404		"Record Not Found"
// @Failure     500		"Internal Error"
// @Security 		ApiKeyAuth
func (c *userController) HeadAuthenticatedUser(ctx *gin.Context) {
	userID := ctx.GetString("id")

	if _, err := c.usersvc.GetUserByID(ctx, userID); err != nil {
		// handle user not found
		if err == usersvc.ErrUserNotFound {
			ctx.JSON(http.StatusNoContent, nil)
			return
		}
		// handle other errors
		log.Error().Err(err).Msg("GetUserByID")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// GetUsers	godoc
// @Schemes
// @Summary 	  Get All Users
// @Description Get All Users
// @Tags 		    Users
// @Accept 	   	json
// @Produce   	json
// @Success 		200 		{array} 	models.User
// @Router 			/api/users/	[get]
// @Security 		ApiKeyAuth
func (c *userController) GetUsers(ctx *gin.Context) {
	user, err := c.usersvc.GetAllUsers(ctx)

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// UpdateUser	  godoc
// @Summary 		Update User
// @Description Update User
// @Tags 				Users
// @Accept 			json
// @Produce 		json
// @Param       payload  body  	   dtos.UpdateUser  true  "User Data"
// @Success 		200 		 {object}  models.User
// @Router 			/api/user/   [patch]
// @Security 		ApiKeyAuth
func (c *userController) UpdateUser(ctx *gin.Context) {
	userID := ctx.GetString("id")

	payload := dtos.UpdateUser{}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	u, err := c.usersvc.UpdateUserByID(ctx, userID, &payload)
	if err == usersvc.ErrUserNotFound {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, u)
}

// GetKitchenPass godoc
// @Summary 		Get AI Kitchen Pass status
// @Description Returns whether the authenticated user has an active AI Kitchen Pass plus the current skill URL and copy text when enabled.
// @Tags 				Users
// @Produce 		json
// @Success 		200 {object} dtos.KitchenPassResponse
// @Router 			/api/user/kitchen-pass [get]
// @Security 		ApiKeyAuth
func (c *userController) GetKitchenPass(ctx *gin.Context) {
	status, err := c.kitchenPasssvc.GetStatus(ctx, ctx.GetString("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, status)
}

// CreateKitchenPass godoc
// @Summary 		Create or rotate AI Kitchen Pass
// @Description Creates an AI Kitchen Pass for the authenticated user, or rotates the existing pass and immediately invalidates the previous token.
// @Tags 				Users
// @Produce 		json
// @Success 		200 {object} dtos.KitchenPassResponse
// @Router 			/api/user/kitchen-pass [post]
// @Security 		ApiKeyAuth
func (c *userController) CreateKitchenPass(ctx *gin.Context) {
	status, err := c.kitchenPasssvc.CreateOrRotate(ctx, ctx.GetString("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, status)
}

// DeleteKitchenPass godoc
// @Summary 		Revoke AI Kitchen Pass
// @Description Revokes the authenticated user's AI Kitchen Pass and immediately disables further bearer-token access for that pass.
// @Tags 				Users
// @Produce 		json
// @Success 		200 {object} dtos.KitchenPassResponse
// @Router 			/api/user/kitchen-pass [delete]
// @Security 		ApiKeyAuth
func (c *userController) DeleteKitchenPass(ctx *gin.Context) {
	userID := ctx.GetString("id")
	if err := c.kitchenPasssvc.Revoke(ctx, userID); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	status, err := c.kitchenPasssvc.GetStatus(ctx, userID)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, status)
}

// TestUsername	godoc
// @Summary 		Returns username validity and availability
// @Description Returns username validity and availability
// @Tags 				Users
// @Accept 			json
// @Produce 		json
// @Param       username    body  	   dtos.TestUsernameRequest  true  "Username Data"
// @Success 		200 		 		{object} 	 dtos.TestUsernameResponse
// @Router 			/api/users/username   [post]
// @Security 		ApiKeyAuth
func (c *userController) TestUsername(ctx *gin.Context) {
	payload := dtos.TestUsernameRequest{}

	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp := dtos.TestUsernameResponse{}
	resp.Username = payload.Username
	resp.Available = false
	resp.Valid = false

	// isValid := c.usersvc.TestUsernameValid(payload.Username)
	// if !isValid {
	// 	resp.Message = "invalid"
	// 	resp.Valid = isValid
	// 	ctx.JSON(http.StatusOK, resp)
	// 	return
	// }

	// isFound, err := c.usersvc.TestUsernameExist(ctx, payload.Username)
	// resp.Valid = !isFound

	// if err == usersvc.ErrUsernameInUse {
	// 	resp.Msg = "unavailable"
	// } else if err != nil {
	// 	ctx.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }

	// ctx.JSON(http.StatusOK, resp)

	// handle empty
	if payload.Username == "" {
		resp.Message = "invalid" // tr@ck: i18n
		resp.Valid = false
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// validate
	if err := c.usersvc.TestName(ctx, payload.Username); err != nil {
		switch e := err; e {
		case usersvc.ErrInvalidUsername:
			resp.Message = "invalid"
			ctx.JSON(http.StatusOK, resp)
			return
		case usersvc.ErrReservedWord:
			resp.Message = "reserved"
			ctx.JSON(http.StatusOK, resp)
			return
		case usersvc.ErrUsernameInUse:
			resp.Valid = true
			resp.Message = "in use"
			ctx.JSON(http.StatusOK, resp)
			return
		default:
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	resp.Available = true
	resp.Valid = true

	ctx.JSON(http.StatusOK, resp)
}
