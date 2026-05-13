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
	jamKerjaHandler *handler.JamKerjaHandler,
	employeeBasicSalaryHandler *handler.EmployeeBasicSalaryHandler,
	faceEmbeddingApprovalHandler *handler.FaceEmbeddingApprovalHandler,
	overtimeRequestHandler *handler.OvertimeRequestHandler,
	assignmentHandler *handler.AssignmentHandler,
	reportHandler *handler.ReportHandler,
	sseHandler *handler.SSEHandler,
) {
	// ==================== CORS MIDDLEWARE ====================
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

	// ==================== HEALTH CHECK ====================
	router.GET("/health", healthHandler.HealthCheck)
	router.Static("/uploads", "./uploads")

	// ==================== API V1 ====================
	v1 := router.Group("/api/v1")
	{
		// -------------------- PUBLIC ROUTES --------------------
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// -------------------- PROTECTED ROUTES --------------------
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// LOGOUT
			protected.POST("/logout", authHandler.Logout)

			// PROFILE MANAGEMENT
			protected.GET("/profile", authHandler.GetProfile)
			protected.PUT("/profile", authHandler.UpdateProfile)
			protected.POST("/profile/change-password", authHandler.ChangePassword)

			// FACE RECOGNITION
			protected.GET("/internal/face/health", faceHandler.Health)
			protected.POST("/admin/users/:id/register-face", middleware.ManagerHROnly(), faceHandler.RegisterFace)
			protected.POST("/face/extract-embedding", faceHandler.ExtractEmbedding)
			protected.POST("/face/register", faceHandler.RegisterFace)

			// ATTENDANCE
			attendance := protected.Group("/attendance")
			{
				attendance.POST("/process", attendanceHandler.ProcessAttendance)
				attendance.POST("/break/start", attendanceHandler.StartBreak)
				attendance.POST("/break/end", attendanceHandler.EndBreak)
				attendance.GET("/today", attendanceHandler.GetTodayAttendance)
				attendance.GET("/monthly", attendanceHandler.GetMonthlyAttendance)
				attendance.GET("/schedule-info", attendanceHandler.GetScheduleInfo) // Informasi jadwal kerja
				attendance.GET("/records", middleware.ManagerHROnly(), attendanceHandler.GetManagerAttendanceRecords)
				attendance.GET("/records/export", middleware.ManagerHROnly(), attendanceHandler.ExportManagerAttendanceRecords)
				// ✅ Tambahan: Records untuk manager departemen (dari kode kedua)
				attendance.GET("/records/my-department", middleware.ManagerDepartemenOnly(), attendanceHandler.GetManagerDeptAttendanceRecords)
				attendance.GET("/records/my-department/export", middleware.ManagerDepartemenOnly(), attendanceHandler.ExportManagerDeptAttendanceRecords)
			}

			// PENGAJUAN IZIN / CUTI
			pengajuan := protected.Group("/pengajuan")
			{
				pengajuan.GET("/tipe", pengajuanHandler.GetTipePengajuan)
				pengajuan.GET("/leave-balance", pengajuanIzinCutiHandler.GetLeaveBalance)
				pengajuan.GET("", pengajuanHandler.GetMyPengajuan)
				pengajuan.POST("", pengajuanHandler.CreatePengajuan)
				// ✅ Tambahan: Update dan Cancel pengajuan (dari kode kedua)
				pengajuan.PUT("/:id", pengajuanHandler.UpdatePengajuan)
				pengajuan.DELETE("/:id", pengajuanHandler.CancelPengajuan)
			}

			// DEPARTMENTS (Manager HR Only)
			departments := protected.Group("/departments")
			departments.Use(middleware.ManagerHROnly())
			{
				departments.POST("", departmentHandler.CreateDepartment)
				departments.GET("", departmentHandler.GetAllDepartments)
				departments.GET("/:id", departmentHandler.GetDepartmentByID)
				departments.PUT("/:id", departmentHandler.UpdateDepartment)
				departments.DELETE("/:id", departmentHandler.DeleteDepartment)
			}

			// POSITIONS (Admin Only)
			positions := protected.Group("/positions")
			positions.Use(middleware.AdminOnly())
			{
				positions.GET("", positionHandler.GetAllPositions)
				positions.GET("/:id", positionHandler.GetPositionByID)
				positions.POST("", positionHandler.CreatePosition)
				positions.PUT("/:id", positionHandler.UpdatePosition)
				positions.DELETE("/:id", positionHandler.DeletePosition)
			}

			protected.GET("/employees/my-department", middleware.ManagerDepartemenOnly(), userHandler.GetEmployeesMyDepartment)
			protected.GET("/employees/search", userHandler.SearchEmployees)
			protected.GET("/payroll/next-number", middleware.ManagerHROnly(), userHandler.GetNextPayrollNumber)

			// EMPLOYEES (Admin Only)
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

			// JAM KERJA (Work Schedule)
			jamKerja := protected.Group("/jam-kerja")
			{
				// Get all work schedules (Admin/HR)
				jamKerja.GET("", jamKerjaHandler.GetAllJamKerja)

				// Get work schedules for manager department
				jamKerja.GET("/my-department", middleware.ManagerOnly(), jamKerjaHandler.GetJamKerjaMyDepartment)

				// Get work schedule by user ID
				jamKerja.GET("/user/:userId", jamKerjaHandler.GetJamKerjaByUserID)

				// Create work schedule (Manager/HR only)
				jamKerja.POST("", middleware.ManagerHROnly(), jamKerjaHandler.CreateJamKerja)

				// Update work schedule by user ID (Manager only)
				jamKerja.PUT("/user/:userId", middleware.ManagerOnly(), jamKerjaHandler.UpdateJamKerjaByUserID)

				// Get available employees without work schedule (Manager/HR only)
				jamKerja.GET("/available-employees", middleware.ManagerHROnly(), jamKerjaHandler.GetAvailableEmployees)
			}

			// FACE EMBEDDING APPROVAL (Manager HR Only) - READ ONLY
			faceEmbeddingApproval := protected.Group("/face-embeddings/detail")
			faceEmbeddingApproval.Use(middleware.ManagerHROnly())
			{
				faceEmbeddingApproval.GET("", faceEmbeddingApprovalHandler.List)
				faceEmbeddingApproval.GET("/:id", faceEmbeddingApprovalHandler.GetByID)
			}

			// GEOFENCING
			geofences := protected.Group("/geofences")
			{
				// Management routes (Manager/HR only)
				geofences.POST("", middleware.ManagerHROnly(), geofenceHandler.CreateGeofence)
				geofences.PUT("/:id", middleware.ManagerHROnly(), geofenceHandler.UpdateGeofence)
				geofences.DELETE("/:id", middleware.ManagerHROnly(), geofenceHandler.DeleteGeofence)

				// Public routes (all authenticated users)
				geofences.GET("", geofenceHandler.GetAllGeofences)
				geofences.GET("/active", geofenceHandler.GetActiveGeofences)
				geofences.GET("/:id", geofenceHandler.GetGeofenceByID)
			}

			// Employee Basic Salaries (Manager HR Only)
			basicSalaries := protected.Group("/employee-basic-salaries")
			basicSalaries.Use(middleware.ManagerHROnly())
			{
				basicSalaries.GET("", employeeBasicSalaryHandler.List)
				basicSalaries.POST("", employeeBasicSalaryHandler.Create)
				basicSalaries.GET("/available-employees", employeeBasicSalaryHandler.AvailableEmployees)
				basicSalaries.GET("/users/:userId/active", employeeBasicSalaryHandler.GetActiveByUser)
				basicSalaries.PATCH("/users/:userId/active", employeeBasicSalaryHandler.UpdateActiveByUser)
				basicSalaries.POST("/users/:userId/deactivate", employeeBasicSalaryHandler.DeactivateActiveByUser)
				basicSalaries.GET("/users/:userId/latest", employeeBasicSalaryHandler.GetLatestByUser)
				basicSalaries.PATCH("/:id", employeeBasicSalaryHandler.UpdateBySalaryID)
			}

			// Check user location against geofence
			protected.POST("/geofences/check", geofenceHandler.CheckUserInGeofence)

			// MY OVERTIME REQUEST (Employee)
			myOvertime := protected.Group("/my-overtime")
			{
				myOvertime.POST("", overtimeRequestHandler.Create)
				myOvertime.GET("", overtimeRequestHandler.GetMine)
				myOvertime.GET("/:id", overtimeRequestHandler.GetMineByID)
				myOvertime.PUT("/:id", overtimeRequestHandler.UpdateMine)
				myOvertime.DELETE("/:id", overtimeRequestHandler.DeleteMine)
				myOvertime.POST("/:id/agree", overtimeRequestHandler.AgreeOvertimeRequest)
				myOvertime.POST("/:id/reject", overtimeRequestHandler.RejectOvertimeRequest)
			}

			// OVERTIME REQUEST APPROVAL (Manager HR Only)
			overtimeRequests := protected.Group("/overtime-requests")
			overtimeRequests.Use(middleware.ManagerHROnly())
			{
				overtimeRequests.GET("", overtimeRequestHandler.ListForManagerHR)
				overtimeRequests.GET("/:id", overtimeRequestHandler.GetForManagerHR)
				overtimeRequests.POST("/:id/approve", overtimeRequestHandler.ApproveByManagerHR)
				overtimeRequests.POST("/:id/reject", overtimeRequestHandler.RejectByManagerHR)
				overtimeRequests.POST("/:id/publish-letter", overtimeRequestHandler.Publish)
			}

			// OVERTIME REQUEST APPROVAL (Kepala Departemen Only)
			deptOvertimeRequests := protected.Group("/dept-overtime-requests")
			deptOvertimeRequests.Use(middleware.ManagerDepartemenOnly())
			{
				deptOvertimeRequests.GET("", overtimeRequestHandler.ListForKepalaDepartemen)
				deptOvertimeRequests.GET("/:id", overtimeRequestHandler.GetForKepalaDepartemen)
				deptOvertimeRequests.POST("/:id/approve", overtimeRequestHandler.ApproveByKepalaDepartemen)
				deptOvertimeRequests.POST("/:id/reject", overtimeRequestHandler.RejectByKepalaDepartemen)
				deptOvertimeRequests.POST("", overtimeRequestHandler.Create)
				deptOvertimeRequests.PUT("/:id", overtimeRequestHandler.Update)
				deptOvertimeRequests.POST("/:id/submit", overtimeRequestHandler.Submit)
				deptOvertimeRequests.POST("/:id/publish", overtimeRequestHandler.Publish)
				deptOvertimeRequests.POST("/:id/employees/:user_id/publish", overtimeRequestHandler.PublishEmployee)
				deptOvertimeRequests.DELETE("/:id", overtimeRequestHandler.Delete)
			}

			// ASSIGNMENTS (Kepala Departemen Only)
			deptAssignments := protected.Group("/dept-assignments")
			deptAssignments.Use(middleware.ManagerDepartemenOnly())
			{
				deptAssignments.GET("", assignmentHandler.ListForManagerDepartemen)
				deptAssignments.GET("/:id", assignmentHandler.GetByID)
				deptAssignments.GET("/preview-schedule", assignmentHandler.PreviewOriginalSchedule)
				deptAssignments.POST("", assignmentHandler.Create)
				deptAssignments.PUT("/:id", assignmentHandler.Update)
				deptAssignments.DELETE("/:id", assignmentHandler.Delete)
			}

			// LEAVE REQUEST APPROVAL (Manager HR Only)
			leaveRequests := protected.Group("/leave-requests")
			leaveRequests.Use(middleware.ManagerHROnly())
			{
				leaveRequests.GET("", pengajuanIzinCutiHandler.ListForManagerHR)
				leaveRequests.GET("/:id", pengajuanIzinCutiHandler.GetForManagerHR)
				leaveRequests.POST("/:id/approve", pengajuanIzinCutiHandler.ApproveByManagerHR)
				leaveRequests.POST("/:id/reject", pengajuanIzinCutiHandler.RejectByManagerHR)
			}

			// LEAVE REQUEST APPROVAL (Kepala Departemen Only)
			deptLeaveRequests := protected.Group("/dept-leave-requests")
			deptLeaveRequests.Use(middleware.ManagerDepartemenOnly())
			{
				deptLeaveRequests.GET("", pengajuanIzinCutiHandler.ListForKepalaDepartemen)
				deptLeaveRequests.GET("/:id", pengajuanIzinCutiHandler.GetForKepalaDepartemen)
				deptLeaveRequests.POST("/:id/approve", pengajuanIzinCutiHandler.ApproveByKepalaDepartemen)
				deptLeaveRequests.POST("/:id/reject", pengajuanIzinCutiHandler.RejectByKepalaDepartemen)
			}

			// REPORTS
			reports := protected.Group("/reports")
			reports.Use(middleware.ManagerHROnly())
			{
				reports.GET("/attendance-activity", reportHandler.GetAttendanceActivityReport)
			}
		}

		// REALTIME (SSE) — endpoint untuk real-time push notification ke Flutter
		// Gunakan tanpa middleware AuthMiddleware karena token diambil dari query param
		// (EventSource tidak support custom headers)
		realtimeGroup := v1.Group("/realtime")
		{
			realtimeGroup.GET("/connect", sseHandler.Connect)
		}
	}
}
