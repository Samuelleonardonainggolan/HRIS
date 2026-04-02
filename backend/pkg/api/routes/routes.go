// pkg/api/routes/routes.go
package routes

import (
	"github.com/andikatampubolon10/hris-backend/internal/config"
	"github.com/andikatampubolon10/hris-backend/pkg/api/handler"
	"github.com/andikatampubolon10/hris-backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	healthHandler *handler.HealthHandler,
	departmentHandler *handler.DepartmentHandler,
	positionHandler *handler.PositionHandler,
	faceHandler *handler.FaceHandler,
	userHandler *handler.UserHandler,
	attendanceHandler *handler.AttendanceHandler,
	geofenceHandler *handler.GeofenceHandler,
	pengajuanIzinCutiHandler *handler.PengajuanIzinCutiHandler,
	pengajuanHandler *handler.PengajuanHandler,
) {
	// CORS Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check (public)
	router.GET("/health", healthHandler.HealthCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// ==================== PUBLIC ROUTES ====================
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// ==================== PROTECTED ROUTES ====================
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Logout
			protected.POST("/logout", authHandler.Logout)

			// ==================== PROFILE ====================
			protected.GET("/profile", authHandler.GetProfile)
			protected.PUT("/profile", authHandler.UpdateProfile)
			protected.POST("/profile/change-password", authHandler.ChangePassword)

			// ==================== FACE RECOGNITION ====================
			protected.GET("/internal/face/health", faceHandler.Health)
			protected.POST("/admin/users/:id/register-face", middleware.ManagerHROnly(), faceHandler.RegisterFace)
			protected.POST("/face/extract-embedding", faceHandler.ExtractEmbedding)
			protected.POST("/face/register", faceHandler.RegisterFace)

			// ==================== ATTENDANCE ====================
			attendance := protected.Group("/attendance")
			{
				attendance.POST("/process", attendanceHandler.ProcessAttendance)
				attendance.GET("/today", attendanceHandler.GetTodayAttendance)
				attendance.GET("/monthly", attendanceHandler.GetMonthlyAttendance)
			}

			// ==================== PENGAJUAN IZIN / CUTI ====================
			// ✅ Route baru — sebelumnya menyebabkan 404 di Flutter
			pengajuan := protected.Group("/pengajuan")
			{
				pengajuan.GET("/tipe", pengajuanHandler.GetTipePengajuan) // GET /api/v1/pengajuan/tipe
				pengajuan.GET("", pengajuanHandler.GetMyPengajuan)        // GET /api/v1/pengajuan
				pengajuan.POST("", pengajuanHandler.CreatePengajuan)      // POST /api/v1/pengajuan
			}

			// ==================== DEPARTMENTS (Manager HR Only) ====================
			departments := protected.Group("/departments")
			departments.Use(middleware.ManagerHROnly())
			{
				departments.POST("", departmentHandler.CreateDepartment)
				departments.GET("", departmentHandler.GetAllDepartments)
				departments.GET("/:id", departmentHandler.GetDepartmentByID)
				departments.PUT("/:id", departmentHandler.UpdateDepartment)
				departments.DELETE("/:id", departmentHandler.DeleteDepartment)
			}

			// ==================== POSITIONS (Admin Only) ====================
			positions := protected.Group("/positions")
			positions.Use(middleware.AdminOnly())
			{
				positions.GET("", positionHandler.GetAllPositions)
				positions.GET("/:id", positionHandler.GetPositionByID)
			}

			// ==================== EMPLOYEES (Admin Only) ====================
			employees := protected.Group("/employees")
			employees.Use(middleware.AdminOnly())
			{
				employees.POST("", userHandler.CreateEmployee)
				employees.GET("", userHandler.GetAllEmployees)
				employees.GET("/template", userHandler.DownloadEmployeeTemplate)
				employees.POST("/import", userHandler.ImportEmployees)
				employees.GET("/:id", userHandler.GetEmployeeByID)
				employees.PUT("/:id", userHandler.UpdateEmployee)
				employees.DELETE("/:id", userHandler.DeleteEmployee)
			}

			// ==================== GEOFENCING ====================
			geofences := protected.Group("/geofences")
			{
				geofences.POST("", middleware.ManagerHROnly(), geofenceHandler.CreateGeofence)
				geofences.PUT("/:id", middleware.ManagerHROnly(), geofenceHandler.UpdateGeofence)
				geofences.DELETE("/:id", middleware.ManagerHROnly(), geofenceHandler.DeleteGeofence)

				geofences.GET("", geofenceHandler.GetAllGeofences)
				geofences.GET("/active", geofenceHandler.GetActiveGeofences)
				geofences.GET("/:id", geofenceHandler.GetGeofenceByID)
			}

			protected.POST("/geofences/check", geofenceHandler.CheckUserInGeofence)

			// ==================== LEAVE REQUEST APPROVAL (Manager HR Only) ====================
			leaveRequests := protected.Group("/leave-requests")
			leaveRequests.Use(middleware.ManagerHROnly())
			{
				leaveRequests.GET("", pengajuanIzinCutiHandler.ListForManagerHR)
				leaveRequests.GET("/:id", pengajuanIzinCutiHandler.GetForManagerHR)
				leaveRequests.POST("/:id/approve", pengajuanIzinCutiHandler.ApproveByManagerHR)
				leaveRequests.POST("/:id/reject", pengajuanIzinCutiHandler.RejectByManagerHR)
			}

			// ==================== WORK SCHEDULES (Manager Departemen Only) ====================
			workSchedules := protected.Group("/work-schedules")
			workSchedules.Use(middleware.ManagerDepartmentOnly()) // jika belum ada, sementara gunakan middleware.ManagerOnly()
			{
				workSchedules.GET("", workScheduleHandler.ListForManagerDepartment)
				workSchedules.PUT("/:userId", workScheduleHandler.UpsertForManagerDepartment)
			}
		}
	}
}
