package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) MessageHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "able to hit to end point",
	})
}
