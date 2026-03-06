package api

import (
	"context"
	"net/http"

	"github.com/andikatampubolon10/hris-backend/pkg/database"
	"github.com/andikatampubolon10/hris-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

type UserRepository interface {
	LoginHandler(c *gin.Context)
	RegisterHandler(c *gin.Context)
}

// bookRepository holds shared resources like database and Redis client
type userRepository struct {
	DB  database.Database
	Ctx *context.Context
}

func NewUserRepository(db database.Database, ctx *context.Context) *userRepository {
	return &userRepository{
		DB:  db,
		Ctx: ctx,
	}
}

// @BasePath /api/v1

// LoginHandler godoc
// @Summary Authenticate a user
// @Schemes
// @Description Authenticates a user using username and password, returns a JWT token if successful
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    models.LoginRequest  true        "User login object"
// @Success 200 {object} models.Response "Use /api/v1/auth/login endpoint"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /login [post]
func (r *userRepository) LoginHandler(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}
	c.JSON(http.StatusNotImplemented, models.ErrorResponse("Deprecated endpoint", "Gunakan /api/v1/auth/login"))
}

// RegisterHandler godoc
// @Summary Register a new user
// @Schemes http
// @Description Registers a new user with the given username and password
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    models.RegisterRequest true "User registration object"
// @Success 201 {object} models.Response "Use /api/v1/auth/register endpoint"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /register [post]
func (r *userRepository) RegisterHandler(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}
	c.JSON(http.StatusNotImplemented, models.ErrorResponse("Deprecated endpoint", "Gunakan /api/v1/auth/register"))
}
