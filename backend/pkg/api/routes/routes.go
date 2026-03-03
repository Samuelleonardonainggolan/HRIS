// pkg/api/routes/routes.go
package routes

import (
    "github.com/andikatampubolon10/hris-backend/internal/config"
    "github.com/andikatampubolon10/hris-backend/pkg/api/handler"
    "github.com/andikatampubolon10/hris-backend/pkg/middleware"
    "github.com/andikatampubolon10/hris-backend/pkg/models"
    "github.com/gin-gonic/gin"
)

func SetupRoutes(
    router *gin.Engine,
    cfg *config.Config,
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    healthHandler *handler.HealthHandler,
) {
    // Middleware global
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())

    // Health check (public)
    router.GET("/health", healthHandler.HealthCheck)

    // API v1
    v1 := router.Group("/api/v1")
    {
        // ==================== PUBLIC ROUTES ====================
        auth := v1.Group("/auth")
        {
            auth.POST("/login", authHandler.Login)
            // Register hanya bisa dilakukan oleh Manager HR (akan dibuat endpoint terpisah)
        }

        // ==================== PROTECTED ROUTES ====================
        protected := v1.Group("")
        protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
        {
            // Profile (All authenticated users)
            protected.GET("/profile", userHandler.GetProfile)
            protected.PUT("/profile", userHandler.UpdateProfile)
            protected.POST("/logout", authHandler.Logout)

            // ==================== MANAGER HR ONLY ====================
            managerHR := protected.Group("/manager-hr")
            managerHR.Use(middleware.ManagerHROnly())
            {
                // User Management (Full Access)
                managerHR.POST("/users", userHandler.CreateUser)
                managerHR.GET("/users", userHandler.GetAllUsers)
                managerHR.GET("/users/:id", userHandler.GetUserByID)
                managerHR.PUT("/users/:id", userHandler.UpdateUser)
                managerHR.DELETE("/users/:id", userHandler.DeleteUser)

                // Department Management
                managerHR.GET("/departments", userHandler.GetAllDepartments)
                managerHR.POST("/departments", userHandler.CreateDepartment)
                managerHR.PUT("/departments/:id", userHandler.UpdateDepartment)
                managerHR.DELETE("/departments/:id", userHandler.DeleteDepartment)

                // Reports (All Departments)
                managerHR.GET("/reports/attendance", userHandler.GetAttendanceReportAll)
                managerHR.GET("/reports/employees", userHandler.GetEmployeeReportAll)
            }

            // ==================== MANAGER DEPARTEMEN ====================
            managerDept := protected.Group("/manager-departemen")
            managerDept.Use(middleware.ManagerOnly())
            managerDept.Use(middleware.DepartmentAccessMiddleware())
            {
                // User Management (Department Only)
                managerDept.GET("/users", userHandler.GetUsersByDepartment)
                managerDept.PUT("/users/:id", userHandler.UpdateUserDepartment)

                // Attendance Management
                managerDept.GET("/attendance", userHandler.GetDepartmentAttendance)
                managerDept.PUT("/attendance/:id/approve", userHandler.ApproveAttendance)
                managerDept.PUT("/attendance/:id/reject", userHandler.RejectAttendance)

                // Reports (Department Only)
                managerDept.GET("/reports/attendance", userHandler.GetAttendanceReportDept)
                managerDept.GET("/reports/employees", userHandler.GetEmployeeReportDept)
            }

            // ==================== ADMIN DEPARTEMEN ====================
            adminDept := protected.Group("/admin-departemen")
            adminDept.Use(middleware.AdminAndManagerOnly())
            adminDept.Use(middleware.DepartmentAccessMiddleware())
            {
                // Create & Edit Staff
                adminDept.POST("/users/staf", userHandler.CreateStaf)
                adminDept.PUT("/users/staf/:id", userHandler.UpdateStaf)
                adminDept.GET("/users/staf", userHandler.GetStafByDepartment)

                // Attendance Input/View
                adminDept.GET("/attendance", userHandler.GetDepartmentAttendance)
                adminDept.POST("/attendance", userHandler.InputAttendance)
            }

            // ==================== STAF ====================
            staf := protected.Group("/staf")
            {
                // Attendance
                staf.POST("/attendance/checkin", userHandler.CheckIn)
                staf.POST("/attendance/checkout", userHandler.CheckOut)
                staf.GET("/attendance/history", userHandler.GetMyAttendance)
                staf.GET("/attendance/today", userHandler.GetTodayAttendance)
            }
        }
    }
}