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
            // Logout (All authenticated users)
            protected.POST("/logout", authHandler.Logout)

            // Profile endpoints (will be implemented later)
            // protected.GET("/profile", userHandler.GetProfile)
            // protected.PUT("/profile", userHandler.UpdateProfile)

            // ==================== MANAGER HR ONLY ====================
            // managerHR := protected.Group("/manager-hr")
            // managerHR.Use(middleware.ManagerHROnly())
            // {
            //     // User Management (Full Access) - To be implemented
            //     managerHR.POST("/users", userHandler.CreateUser)
            //     managerHR.GET("/users", userHandler.GetAllUsers)
            //     managerHR.GET("/users/:id", userHandler.GetUserByID)
            //     managerHR.PUT("/users/:id", userHandler.UpdateUser)
            //     managerHR.DELETE("/users/:id", userHandler.DeleteUser)
            //
            //     // Department Management - To be implemented
            //     managerHR.GET("/departments", departmentHandler.GetAll)
            //     managerHR.POST("/departments", departmentHandler.Create)
            //
            //     // Position Management - To be implemented
            //     managerHR.GET("/positions", positionHandler.GetAll)
            //     managerHR.GET("/departments/:id/positions", positionHandler.GetByDepartment)
            //
            //     // Reports (All Departments) - To be implemented
            //     managerHR.GET("/reports/attendance", reportHandler.GetAttendanceAll)
            //     managerHR.GET("/reports/employees", reportHandler.GetEmployeeAll)
            // }

            // ==================== MANAGER DEPARTEMEN ====================
            // managerDept := protected.Group("/manager-dept")
            // managerDept.Use(middleware.ManagerOnly())
            // managerDept.Use(middleware.DepartmentAccessMiddleware())
            // {
            //     // Team Management - To be implemented
            //     managerDept.GET("/team", userHandler.GetTeamMembers)
            //     managerDept.GET("/team/:id", userHandler.GetTeamMemberDetail)
            //
            //     // Attendance Management - To be implemented
            //     managerDept.GET("/attendance", attendanceHandler.GetDepartmentAttendance)
            //     managerDept.PUT("/attendance/:id/approve", attendanceHandler.Approve)
            //
            //     // Reports (Department Only) - To be implemented
            //     managerDept.GET("/reports/attendance", reportHandler.GetAttendanceDept)
            //     managerDept.GET("/reports/team", reportHandler.GetTeamReport)
            // }

            // ==================== ADMIN DEPARTEMEN ====================
            // adminDept := protected.Group("/admin-dept")
            // adminDept.Use(middleware.AdminAndManagerOnly())
            // adminDept.Use(middleware.DepartmentAccessMiddleware())
            // {
            //     // Create & Manage Staff - To be implemented
            //     adminDept.POST("/staff", userHandler.CreateStaff)
            //     adminDept.GET("/staff", userHandler.GetDepartmentStaff)
            //     adminDept.PUT("/staff/:id", userHandler.UpdateStaff)
            //
            //     // Attendance Management - To be implemented
            //     adminDept.GET("/attendance", attendanceHandler.GetDepartmentAttendance)
            //     adminDept.POST("/attendance", attendanceHandler.InputAttendance)
            // }

            // ==================== STAF ====================
            // staf := protected.Group("/staf")
            // {
            //     // Attendance - To be implemented
            //     staf.POST("/attendance/checkin", attendanceHandler.CheckIn)
            //     staf.POST("/attendance/checkout", attendanceHandler.CheckOut)
            //     staf.GET("/attendance/history", attendanceHandler.GetMyHistory)
            //     staf.GET("/attendance/today", attendanceHandler.GetMyToday)
            //
            //     // Leave Requests - To be implemented
            //     staf.POST("/leave-requests", leaveHandler.Create)
            //     staf.GET("/leave-requests", leaveHandler.GetMyRequests)
            //
            //     // Profile - To be implemented
            //     staf.GET("/profile", userHandler.GetMyProfile)
            //     staf.PUT("/profile", userHandler.UpdateMyProfile)
            // }
        }
    }
}