// cmd/server/main.go (FINAL VERSION - GABUNGAN)
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
	"github.com/andikatampubolon10/hris-backend/pkg/storage"
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

	// ==================== Initialize Repositories ====================
	userRepo := repository.NewUserRepository(mongodb.Database)
	departmentRepo := repository.NewDepartmentRepository(mongodb.Database)
	positionRepo := repository.NewPositionRepository(mongodb.Database)
	faceEmbeddingRepo := repository.NewFaceEmbeddingRepository(mongodb.Database)
	attendanceRepo := repository.NewAttendanceRepository(mongodb.Database)
	breakTimeRepo := repository.NewBreakTimeRepository(mongodb.Database)
	geofenceRepo := repository.NewGeofenceRepository(mongodb.Database)
	pengajuanIzinCutiRepo := repository.NewPengajuanIzinCutiRepository(mongodb.Database)
	jamKerjaRepo := repository.NewJamKerjaRepository(mongodb.Database) // ✅ Dari kode kedua
	employeeBasicSalaryRepo := repository.NewEmployeeBasicSalaryRepository(mongodb.Database)

	log.Println("📦 Repositories initialized")

	// ==================== Initialize External Clients ====================
	timeout, err := time.ParseDuration(cfg.FaceHTTPTimeout)
	if err != nil {
		timeout = 30 * time.Second
	}
	faceClient := faceclient.New(cfg.FaceServiceURL, cfg.FaceAPIKey, timeout)

	log.Println("🔌 External clients initialized")

	// ==================== Initialize Services ====================
	jwtExpiryStr := strconv.Itoa(cfg.JWTExpiry)
	// Initialize Supabase uploader if configured
	var supabaseUploader *storage.SupabaseUploader
	storageKey := cfg.SupabaseServiceRoleKey
	if storageKey == "" {
		storageKey = cfg.SupabaseAPIKey
	}

	if cfg.SupabaseURL != "" && storageKey != "" {
		supabaseUploader = storage.NewSupabaseUploader(cfg.SupabaseURL, storageKey, cfg.SupabaseBucket)
		log.Println("☁️  Supabase uploader initialized")
	} else {
		log.Println("⚠️  Supabase not configured, using local file storage")
	}

	authService := service.NewAuthService(userRepo, faceEmbeddingRepo, cfg.JWTSecret, jwtExpiryStr)
	userService := service.NewUserService(userRepo, departmentRepo, positionRepo)
	departmentService := service.NewDepartmentService(departmentRepo, userRepo)
	positionService := service.NewPositionService(positionRepo)

	// FaceService dengan parameter lengkap (dari kode kedua)
	faceService := service.NewFaceService(userRepo, faceEmbeddingRepo, faceClient, cfg.PublicBaseURL, cfg.FaceImageDir, supabaseUploader)

	// ✅ AttendanceService dengan jamKerjaRepo (dari kode pertama)
	attendanceService := service.NewAttendanceService(attendanceRepo, breakTimeRepo, userRepo, faceEmbeddingRepo, jamKerjaRepo, geofenceRepo, faceClient)

	// PengajuanService dengan konfigurasi lengkap (dari kode kedua)
	var pengajuanService service.PengajuanService
	if supabaseUploader != nil {
		pengajuanService = service.NewPengajuanServiceWithSupabase(mongodb.Database, supabaseUploader)
	} else {
		pengajuanService = service.NewPengajuanServiceWithConfig(mongodb.Database, cfg.PublicBaseURL, cfg.PengajuanDocDir)
	}

	geofenceService := service.NewGeofenceService(geofenceRepo, userRepo)
	pengajuanIzinCutiService := service.NewPengajuanIzinCutiService(pengajuanIzinCutiRepo, userRepo, mongodb.Database)
	jamKerjaService := service.NewJamKerjaService(jamKerjaRepo, userRepo) // ✅ Dari kode kedua
	employeeBasicSalaryService := service.NewEmployeeBasicSalaryService(employeeBasicSalaryRepo, userRepo)
	faceEmbeddingApprovalService := service.NewFaceEmbeddingApprovalService(faceEmbeddingRepo, userRepo)
	log.Println("⚙️  Services initialized")

	// ==================== Initialize Handlers ====================
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler(mongodb)
	departmentHandler := handler.NewDepartmentHandler(departmentService)
	positionHandler := handler.NewPositionHandler(positionService)
	faceHandler := handler.NewFaceHandler(faceService)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService, faceService)
	geofenceHandler := handler.NewGeofenceHandler(geofenceService)
	pengajuanIzinCutiHandler := handler.NewPengajuanIzinCutiHandler(pengajuanIzinCutiService)
	pengajuanHandler := handler.NewPengajuanHandler(pengajuanService)
	jamKerjaHandler := handler.NewJamKerjaHandler(jamKerjaService) // ✅ Dari kode kedua
	employeeBasicSalaryHandler := handler.NewEmployeeBasicSalaryHandler(employeeBasicSalaryService)
	faceEmbeddingApprovalHandler := handler.NewFaceEmbeddingApprovalHandler(faceEmbeddingApprovalService)
	log.Println("🎯 Handlers initialized")

	// ==================== Setup Gin ====================
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// ==================== Setup Routes ====================
	routes.SetupRoutes(
		router,
		cfg,
		authHandler,
		healthHandler,
		departmentHandler,
		positionHandler,
		faceHandler,
		userHandler,
		attendanceHandler,
		geofenceHandler,
		pengajuanIzinCutiHandler,
		pengajuanHandler,
		jamKerjaHandler, // ✅ Dari kode kedua
		employeeBasicSalaryHandler,
		faceEmbeddingApprovalHandler,
	)

	log.Println("🛣️  Routes configured")

	// ==================== Start Server ====================
	port := cfg.ServerPort
	log.Println("================================================")
	log.Printf("🚀 Server running on port %s", port)
	log.Printf("📍 Environment: %s", cfg.Environment)
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Printf("🔗 API Base URL: http://localhost:%s/api/v1", port)
	log.Println("================================================")
	log.Println("📋 Available endpoints:")
	log.Println("   Auth:")
	log.Println("     POST   /api/v1/auth/login")
	log.Println("     POST   /api/v1/auth/register")
	log.Println("   Departments:")
	log.Println("     GET    /api/v1/departments")
	log.Println("     POST   /api/v1/departments")
	log.Println("   Employees:")
	log.Println("     GET    /api/v1/employees")
	log.Println("     POST   /api/v1/employees")
	log.Println("   Attendance:")
	log.Println("     POST   /api/v1/attendance/process")
	log.Println("     GET    /api/v1/attendance/today")
	log.Println("     GET    /api/v1/attendance/monthly")       // ✅ Dari kode pertama
	log.Println("     GET    /api/v1/attendance/schedule-info") // ✅ Dari kode pertama (via routes)
	log.Println("   Jam Kerja (Work Schedule):")                // ✅ Dari kode kedua
	log.Println("     GET    /api/v1/jam-kerja")
	log.Println("     GET    /api/v1/jam-kerja/my-department")
	log.Println("     GET    /api/v1/jam-kerja/user/:userId")
	log.Println("     POST   /api/v1/jam-kerja")
	log.Println("     PUT    /api/v1/jam-kerja/user/:userId")
	log.Println("   Geofencing:")
	log.Println("     GET    /api/v1/geofences")
	log.Println("     POST   /api/v1/geofences")
	log.Println("     POST   /api/v1/geofences/check")
	log.Println("   Pengajuan Izin/Cuti:")
	log.Println("     GET    /api/v1/pengajuan")
	log.Println("     GET    /api/v1/pengajuan/tipe")
	log.Println("     POST   /api/v1/pengajuan")
	log.Println("================================================")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
