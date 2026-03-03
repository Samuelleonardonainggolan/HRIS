// pkg/api/handler/health_handler.go
package handler

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
    return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "ok",
        "message":   "HRIS Backend is running",
        "timestamp": time.Now().Format(time.RFC3339),
        "version":   "1.0.0",
    })
}