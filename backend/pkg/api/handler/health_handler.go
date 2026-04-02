// pkg/api/handler/health_handler.go
package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/pkg/database"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	mongodb *database.MongoDB
}

func NewHealthHandler(mongodb *database.MongoDB) *HealthHandler {
	return &HealthHandler{
		mongodb: mongodb,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Check MongoDB connection
	if err := h.mongodb.Database.Client().Ping(c.Request.Context(), nil); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"message": "Database connection failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"message":  "Service is running",
		"database": "connected",
	})
}
