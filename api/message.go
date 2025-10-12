package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) MessageHandler(ctx *gin.Context) {

	response, _, _ := server.DockerCli.RunPython()
	ctx.JSON(http.StatusOK, gin.H{
		"message": response,
	})

}
