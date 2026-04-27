package controllers

import (
	kitchenpasssvc "4ks/apps/api/services/kitchenpass"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// KitchenPassController is the public AI Kitchen Pass skill controller.
type KitchenPassController interface {
	GetSkillPage(*gin.Context)
}

type kitchenPassController struct {
	kitchenPasssvc kitchenpasssvc.Service
}

// NewKitchenPassController creates a new skill page controller.
func NewKitchenPassController(service kitchenpasssvc.Service) KitchenPassController {
	return &kitchenPassController{kitchenPasssvc: service}
}

// GetSkillPage godoc
// @Summary 		Get AI Kitchen Pass skill page
// @Description Returns a markdown skill document for an active AI Kitchen Pass token.
// @Tags 				AI
// @Produce 		plain
// @Param       token path string true "AI Kitchen Pass token"
// @Success 		200 {string} string
// @Failure     404 {string} string
// @Router 			/ai/{token} [get]
func (c *kitchenPassController) GetSkillPage(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/markdown; charset=utf-8")
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Referrer-Policy", "no-referrer")
	ctx.Header("X-Robots-Tag", "noindex, nofollow")
	ctx.Header("Content-Security-Policy", "default-src 'none'")

	token := ctx.Param("token")
	if !kitchenpasssvc.IsKitchenPassToken(token) {
		ctx.String(http.StatusNotFound, "not found")
		return
	}

	if _, err := c.kitchenPasssvc.ValidateToken(ctx, token); err != nil {
		if errors.Is(err, kitchenpasssvc.ErrKitchenPassNotFound) || errors.Is(err, kitchenpasssvc.ErrInvalidKitchenPassToken) {
			ctx.String(http.StatusNotFound, "not found")
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.String(http.StatusOK, kitchenpasssvc.RenderSkillDocument(token))
}
