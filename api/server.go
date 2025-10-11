package api

import (
	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
}

func (server *Server) CreateServer() *Server {
	server.router = gin.Default()

	server.router.GET("/messages", server.MessageHandler)

	return server
}

func (server *Server) StartServer(address string) error {
	return server.router.Run(address)
}
