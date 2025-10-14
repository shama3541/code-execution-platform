package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	database "github.com/shama3541/code-execution-platform/database/db"
	dockerpool "github.com/shama3541/code-execution-platform/docker-pool"
	dockerutil "github.com/shama3541/code-execution-platform/docker-util"
	"github.com/shama3541/code-execution-platform/token"
)

type Server struct {
	router     *gin.Engine
	Queries    *database.Queries
	TokenMaker token.Maker
	DockerCli  *dockerutil.DockerCLI
	Warmpool   *dockerpool.WarmPool
}

func NewServer(dbconn *sql.DB) *Server {
	tokenMaker := token.NewJwtMaker("123456")
	mydockercli, _ := dockerutil.NewDockerCLI()
	Warmpool := dockerpool.NewWarmPool(mydockercli.Client)
	return &Server{
		Queries:    database.New(dbconn),
		TokenMaker: tokenMaker,
		DockerCli:  mydockercli,
		Warmpool:   Warmpool,
	}
}

func (server *Server) CreateServer() *Server {
	server.router = gin.Default()
	server.router.GET("/messages")
	server.router.POST("/users", server.CreateUser)
	server.router.POST("login", server.LoginHandler)

	authroutes := server.router.Group("/").Use(Middleware(server.TokenMaker))
	authroutes.POST("/run", server.RunProgramHandler)

	return server
}

func (server *Server) StartServer(address string) error {
	return server.router.Run(address)
}
