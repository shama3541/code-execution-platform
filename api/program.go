package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProgramRequest struct {
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required,oneof=python goLang javascript"`
}

func (server *Server) RunProgramHandler(ctx *gin.Context) {
	var req ProgramRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Bad request",
		})
		return
	}
	stdout, stderr, err := server.DockerCli.RunCode(req.Language, req.Code)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "internal error while running code",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"output": stdout,
		"stderr": stderr,
	})

}
