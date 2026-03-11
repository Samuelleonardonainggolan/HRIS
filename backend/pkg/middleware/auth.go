// pkg/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/andikatampubolon10/hris-backend/pkg/auth"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "Missing authorization header"))
			c.Abort()
			return
		}

		// Log untuk debugging
		println("[AuthMiddleware] Authorization header:", authHeader)

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "Invalid authorization format"))
			c.Abort()
			return
		}

		token := tokenParts[1]
		println("[AuthMiddleware] Token extracted, length:", len(token))

		// Validate token
		claims, err := auth.ValidateToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "Invalid or expired token"))
			c.Abort()
			return
		}

		println("[AuthMiddleware] Token valid for user:", claims.UserID)

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("userDepartment", claims.Department)

		c.Next()
	}
}

// ManagerHROnly restricts access to Manager HR only
func ManagerHROnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "User role not found"))
			c.Abort()
			return
		}

		if role != models.RoleManagerHR {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "Access denied: Manager HR only"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ManagerOnly restricts access to Manager HR and Manager Departemen
func ManagerOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "User role not found"))
			c.Abort()
			return
		}

		// ✅ Fixed: Use correct role constants
		if role != models.RoleManagerHR && role != models.RoleManagerDepartemen {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "Access denied: Manager only"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly restricts access to Admins (Manager + Admin Departemen)
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "User role not found"))
			c.Abort()
			return
		}

		// ✅ Fixed: Use correct role constants
		allowedRoles := []string{
			models.RoleManagerHR,
			models.RoleManagerDepartemen,
			models.RoleAdminDepartemen,
		}

		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", "Access denied: Admin only"))
			c.Abort()
			return
		}

		c.Next()
	}
}
