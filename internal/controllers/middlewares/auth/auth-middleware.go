package authmiddleware

import (
	"github.com/gin-gonic/gin"
	tokenauth "github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth"
	parseerrors "github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/parse-errors"
	"net/http"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware(tokenMaker tokenauth.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationCookie, err := ctx.Cookie("auth-financial-chat")
		if err != nil {
			ctx.Redirect(http.StatusTemporaryRedirect, "/users/login")
			return
		}

		payload, err := tokenMaker.VerifyToken(authorizationCookie)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, parseerrors.ErrorResponse(err))
			return
		}

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}
