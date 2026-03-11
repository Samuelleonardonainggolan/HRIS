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
	faceHandler *handler.FaceHandler,
	attendanceHandler *handler.AttendanceHandler,
) {
	// CORS Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
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

			// Face Recognition
			protected.GET("/internal/face/health", faceHandler.Health)
			protected.POST("/admin/users/:id/register-face", middleware.ManagerHROnly(), faceHandler.RegisterFace)

			// PERBAIKAN: Hanya satu route untuk attendance/process
			// Gunakan attendanceHandler, hapus faceHandler.ProcessAttendance
			attendance := protected.Group("/attendance")
			{
				attendance.POST("/process", attendanceHandler.ProcessAttendance) // Hanya satu
				attendance.GET("/today", attendanceHandler.GetTodayAttendance)
				attendance.GET("/monthly", attendanceHandler.GetMonthlyAttendance)
			}

			// Departments (Manager HR Only)
			departments := protected.Group("/departments")
			departments.Use(middleware.ManagerHROnly())
			{
				departments.POST("", departmentHandler.CreateDepartment)
				departments.GET("", departmentHandler.GetAllDepartments)
				departments.GET("/:id", departmentHandler.GetDepartmentByID)
				departments.PUT("/:id", departmentHandler.UpdateDepartment)
				departments.DELETE("/:id", departmentHandler.DeleteDepartment)
			}
		}
	}
}
