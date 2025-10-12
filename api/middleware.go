package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/shama3541/code-execution-platform/token"
)

const (
	authHeaderKey  = "Authorization"
	authType       = "Bearer"
	authpayloadkey = "authpayloadkey"
)

func Middleware(maker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		Authorization := ctx.GetHeader(authHeaderKey)
		if len(Authorization) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Auhtorization header is empty",
			})
			return
		}
		authvalue := strings.Fields(Authorization)
		bearerToken := authvalue[1]
		response, err := maker.VerifyToken(bearerToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err})
			return
		}

		ctx.Set(authpayloadkey, response)
		ctx.Next()
	}
}
