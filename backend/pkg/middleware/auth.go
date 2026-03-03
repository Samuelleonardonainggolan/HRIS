// pkg/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"

    "github.com/andikatampubolon10/hris-backend/pkg/auth"
    "github.com/andikatampubolon10/hris-backend/pkg/models"
    "github.com/gin-gonic/gin"
)

// AuthMiddleware - JWT validation
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "No token provided"))
            c.Abort()
            return
        }

        tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

        claims, err := auth.ValidateToken(tokenString, jwtSecret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "Invalid token"))
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)
        c.Set("department", claims.Department)
        c.Next()
    }
}

// RoleMiddleware - Check if user has required role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "Role not found"))
            c.Abort()
            return
        }

        role := userRole.(string)
        for _, allowedRole := range allowedRoles {
            if role == allowedRole {
                c.Next()
                return
            }
        }

        c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "Insufficient permissions"))
        c.Abort()
    }
}

// ManagerHROnly - Only Manager HR can access
func ManagerHROnly() gin.HandlerFunc {
    return RoleMiddleware(models.RoleManagerHR)
}

// ManagerOnly - Manager HR or Manager Departemen
func ManagerOnly() gin.HandlerFunc {
    return RoleMiddleware(models.RoleManagerHR, models.RoleManagerDept)
}

// AdminAndManagerOnly - Admin Departemen, Manager Departemen, or Manager HR
func AdminAndManagerOnly() gin.HandlerFunc {
    return RoleMiddleware(models.RoleManagerHR, models.RoleManagerDept, models.RoleAdminDept)
}

// DepartmentAccessMiddleware - Check if user can access department data
func DepartmentAccessMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, _ := c.Get("role")
        userDept, _ := c.Get("department")

        // Manager HR can access all departments
        if userRole == models.RoleManagerHR {
            c.Next()
            return
        }

        // Get requested department from params or query
        requestedDept := c.Param("department")
        if requestedDept == "" {
            requestedDept = c.Query("department")
        }

        // If no department specified, allow (will be filtered in handler)
        if requestedDept == "" {
            c.Next()
            return
        }

        // Check if user's department matches requested department
        if userDept != requestedDept {
            c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "Cannot access other department data"))
            c.Abort()
            return
        }

        c.Next()
    }
}