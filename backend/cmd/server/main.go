// cmd/server/main.go
package main

import (
    "log"

    "github.com/andikatampubolon10/hris-backend/internal/config"
    "github.com/andikatampubolon10/hris-backend/internal/service"
    "github.com/andikatampubolon10/hris-backend/pkg/api/handler"
    "github.com/andikatampubolon10/hris-backend/pkg/api/routes"
    "github.com/andikatampubolon10/hris-backend/pkg/database"
    "github.com/andikatampubolon10/hris-backend/pkg/database/repository"
    "github.com/gin-gonic/gin"
)

func main() {
    // Load configuration
    cfg := config.LoadConfig()

    // Connect to MongoDB
    mongodb, err := database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName)
    if err != nil {
        log.Fatal("Failed to connect to MongoDB:", err)
    }
    defer mongodb.Disconnect()

    log.Println("✅ Database connected successfully")

    // Initialize repositories
    userRepo := repository.NewUserRepository(mongodb.Database)

    // Initialize services
    authService := service.NewAuthService(
        userRepo,
        cfg.JWTSecret,
        cfg.JWTExpiry,
    )

    // Initialize handlers
    healthHandler := handler.NewHealthHandler()
    authHandler := handler.NewAuthHandler(authService)

    // Setup Gin router
    r := gin.Default()

    // Setup routes
    routes.SetupRoutes(r, cfg, authHandler, healthHandler)

    // Start server
    log.Printf("🚀 Server running on port %s", cfg.ServerPort)
    log.Printf("📍 Environment: %s", cfg.Environment)
    log.Printf("🔗 Health check: http://localhost:%s/health", cfg.ServerPort)
    log.Printf("🔗 API Base URL: http://localhost:%s/api/v1", cfg.ServerPort)
    log.Println("\n📋 Available Endpoints:")
    log.Println("   Public:")
    log.Println("     POST /api/v1/auth/login")
    log.Println("     POST /api/v1/auth/register")
    log.Println("     POST /api/v1/auth/refresh")
    log.Println("   Protected:")
    log.Println("     POST /api/v1/logout")
    log.Println("   Manager HR:")
    log.Println("     /api/v1/manager-hr/* (Coming soon)")
    log.Println("   Manager Dept:")
    log.Println("     /api/v1/manager-dept/* (Coming soon)")
    log.Println("   Admin Dept:")
    log.Println("     /api/v1/admin-dept/* (Coming soon)")
    log.Println("   Staff:")
    log.Println("     /api/v1/staf/* (Coming soon)")
    log.Println()

    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}