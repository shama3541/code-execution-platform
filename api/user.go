package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	database "github.com/shama3541/code-execution-platform/database/db"
	"github.com/shama3541/code-execution-platform/util"
)

type UserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

func (server *Server) CreateUser(ctx *gin.Context) {
	var req UserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Bad request, invalid params",
			"error":   err.Error(),
		})
		return
	}
	Hashedpassword, _ := util.HashPassword(req.Password)
	args := database.CreateUserParams{
		Username: req.Username,
		Password: Hashedpassword,
		Email:    req.Email,
		FullName: req.FullName,
	}

	response, err := server.Queries.CreateUser(ctx, args)
	if err != nil {
		if pqerror, ok := err.(*pq.Error); ok {
			switch pqerror.Code.Name() {
			case "unique_code_violation":
				ctx.JSON(http.StatusConflict, gin.H{
					"message": "User already exists",
				})
			}
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal error",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)

}

type Loginrequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (server *Server) LoginHandler(ctx *gin.Context) {
	var req Loginrequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Bad request: invalid params",
			"error":   err.Error(),
		})
		return
	}

	response, err := server.Queries.FindUserByName(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "User not found , please register first",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
			"error":   err.Error(),
		})
		return
	}
	ok := util.VerifyPassword(response.Password, req.Password)
	if ok != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "incorrect password",
		})
		return
	}

	duration, err := time.ParseDuration("15m")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to parse token duration",
			"error":   err.Error(),
		})
		return
	}

	token, _ := server.TokenMaker.CreateToken(response.Username, duration)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "login was successful",
		"token":   token,
	})

}
