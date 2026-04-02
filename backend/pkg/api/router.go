package api

import (
	"context"
	"os"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/cache"
	"github.com/andikatampubolon10/hris-backend/pkg/database"
	"github.com/andikatampubolon10/hris-backend/pkg/middleware"

	docs "github.com/andikatampubolon10/hris-backend/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"golang.org/x/time/rate"
)

func ContextMiddleware(bookRepository BookRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("appCtx", bookRepository)
		c.Next()
	}
}

func NewRouter(_logger interface{}, _mongoCollection interface{}, db database.Database, redisClient cache.Cache, ctx *context.Context) *gin.Engine {
	bookRepository := NewBookRepository(db, redisClient, ctx)
	userRepository := NewUserRepository(db, ctx)

	r := gin.Default()
	r.Use(ContextMiddleware(bookRepository))

	//r.Use(gin.Logger())
	r.Use(middleware.Logger())
	if gin.Mode() == gin.ReleaseMode {
		r.Use(middleware.Security())
		r.Use(middleware.Xss())
	}
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(rate.Every(1*time.Minute), 60)) // 60 requests per minute

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
		v1.GET("/", bookRepository.Healthcheck)
		v1.GET("/books", middleware.APIKeyAuth(), bookRepository.FindBooks)
		v1.POST("/books", middleware.APIKeyAuth(), middleware.AuthMiddleware(os.Getenv("JWT_SECRET")), bookRepository.CreateBook)
		v1.GET("/books/:id", middleware.APIKeyAuth(), bookRepository.FindBook)
		v1.PUT("/books/:id", middleware.APIKeyAuth(), bookRepository.UpdateBook)
		v1.DELETE("/books/:id", middleware.APIKeyAuth(), bookRepository.DeleteBook)

		v1.POST("/login", middleware.APIKeyAuth(), userRepository.LoginHandler)
		v1.POST("/register", middleware.APIKeyAuth(), userRepository.RegisterHandler)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return r
}
