// cmd/server/main.go
package main

import (
	"log"
	"strconv"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/config"
	"github.com/andikatampubolon10/hris-backend/internal/faceclient"
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

	// Setup MongoDB
	mongodb, err := database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongodb.Disconnect()

	log.Println("✅ Database connected successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(mongodb.Database)
	departmentRepo := repository.NewDepartmentRepository(mongodb.Database)
	faceEmbeddingRepo := repository.NewFaceEmbeddingRepository(mongodb.Database)
	attendanceRepo := repository.NewAttendanceRepository(mongodb.Database)

	// Initialize external clients
	timeout, err := time.ParseDuration(cfg.FaceHTTPTimeout)
	if err != nil {
		timeout = 30 * time.Second
	}
	faceClient := faceclient.New(cfg.FaceServiceURL, cfg.FaceAPIKey, timeout)

	// --- PERBAIKAN DI SINI ---
	// Konversi cfg.JWTExpiry dari int ke string karena fungsi NewAuthService mengharapkan string
	jwtExpiryStr := strconv.Itoa(cfg.JWTExpiry)
	log.Printf("🔑 JWT Expiry (original int): %d, (converted to string): %s", cfg.JWTExpiry, jwtExpiryStr)

	// Initialize services dengan parameter yang sudah benar
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, jwtExpiryStr)
	departmentService := service.NewDepartmentService(departmentRepo, userRepo)
	faceService := service.NewFaceService(userRepo, faceEmbeddingRepo, faceClient)
	attendanceService := service.NewAttendanceService(attendanceRepo, userRepo, faceEmbeddingRepo, faceClient)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler(mongodb)
	departmentHandler := handler.NewDepartmentHandler(departmentService)
	faceHandler := handler.NewFaceHandler(faceService)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService, faceService)

	// Setup Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup all routes
	routes.SetupRoutes(router, cfg, authHandler, healthHandler, departmentHandler, faceHandler, attendanceHandler)

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
