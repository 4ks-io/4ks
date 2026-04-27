package middleware

import "github.com/gin-gonic/gin"

const (
	AuthTypeJWT = "jwt"
	AuthTypePAT = "pat"
)

type AuthIdentity struct {
	AuthID     string
	AuthType   string
	UserID     string
	Email      string
	PATDigest  string
	PATPreview string
}

func SetAuthIdentity(ctx *gin.Context, identity AuthIdentity) {
	if identity.AuthID != "" {
		ctx.Set("authID", identity.AuthID)
	}
	if identity.AuthType != "" {
		ctx.Set("authType", identity.AuthType)
	}
	if identity.UserID != "" {
		ctx.Set("id", identity.UserID)
	}
	if identity.Email != "" {
		ctx.Set("email", identity.Email)
	}
	if identity.PATDigest != "" {
		ctx.Set("patDigest", identity.PATDigest)
	}
	if identity.PATPreview != "" {
		ctx.Set("patPreview", identity.PATPreview)
	}
}
