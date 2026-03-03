// cmd/server/main.go
package main

import (
    "log"

    "github.com/andikatampubolon10/hris-backend/internal/config"
    "github.com/andikatampubolon10/hris-backend/internal/service"
    "github.com/andikatampubolon10/hris-backend/pkg/api/handler"
    "github.com/andikatampubolon10/hris-backend/pkg/database"
    "github.com/andikatampubolon10/hris-backend/pkg/database/repository"
    "github.com/andikatampubolon10/hris-backend/pkg/middleware"
    "github.com/gin-gonic/gin"
)

func main() {
    // Load configuration
    cfg := config.LoadConfig()

    // Setup MongoDB
    mongodb, err := database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName)
    if err != nil {
        log.Fatal("Failed to connect to MongoDB:", err)
    }
    defer mongodb.Disconnect()

    log.Println("✅ Database connected successfully")

    // Initialize repositories
    userRepo := repository.NewUserRepository(mongodb.Database)

    // Initialize services
    authService := service.NewAuthService(userRepo, cfg)

    // Initialize handlers
    authHandler := handler.NewAuthHandler(authService)
    healthHandler := handler.NewHealthHandler()

    // Setup Gin
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.Default()

    // Global middleware
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())

    // Health check
    router.GET("/health", healthHandler.HealthCheck)

    // API v1
    v1 := router.Group("/api/v1")
    {
        // Public routes
        auth := v1.Group("/auth")
        {
            auth.POST("/login", authHandler.Login)
            auth.POST("/refresh", authHandler.RefreshToken)
        }

        // Protected routes
        protected := v1.Group("")
        protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
        {
            protected.POST("/logout", authHandler.Logout)
        }
    }

    // Start server
    port := cfg.ServerPort
    log.Printf("🚀 Server running on port %s", port)
    log.Printf("📍 Environment: %s", cfg.Environment)
    log.Printf("🔗 Health check: http://localhost:%s/health", port)
    log.Printf("🔗 API Base URL: http://localhost:%s/api/v1", port)

    if err := router.Run(":" + port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}