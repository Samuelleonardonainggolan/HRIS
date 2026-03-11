// pkg/api/handler/geofence_handler.go
package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GeofenceHandler struct {
	service service.GeofenceService
}

func NewGeofenceHandler(service service.GeofenceService) *GeofenceHandler {
	return &GeofenceHandler{
		service: service,
	}
}

// CreateGeofence creates a new geofence
func (h *GeofenceHandler) CreateGeofence(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req models.CreateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get userID (as string for service)
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Validate userID format
	_, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID format",
		})
		return
	}

	userName := getUserName(c)

	fmt.Printf("📍 Creating geofence - UserID: %s, UserName: %s, Location: %s\n", 
		userIDStr, userName, req.Name)

	// ✅ Call service with string ID
	geofence, err := h.service.CreateGeofence(ctx, req, userIDStr, userName)
	if err != nil {
		fmt.Printf("❌ Failed to create geofence: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	fmt.Printf("✅ Geofence created: %s (ID: %s)\n", geofence.Name, geofence.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Geofence created successfully",
		"data":    geofence,
	})
}

// GetAllGeofences retrieves all geofences
func (h *GeofenceHandler) GetAllGeofences(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	geofences, err := h.service.GetAllGeofences(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	fmt.Printf("📋 Retrieved %d geofences\n", len(geofences))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Geofences retrieved successfully",
		"data":    geofences,
	})
}

// GetGeofenceByID retrieves a specific geofence
func (h *GeofenceHandler) GetGeofenceByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id := c.Param("id")

	// Validate ID format
	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid geofence ID format",
		})
		return
	}

	// ✅ Pass string ID to service
	geofence, err := h.service.GetGeofenceByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Geofence retrieved successfully",
		"data":    geofence,
	})
}

// UpdateGeofence updates an existing geofence
func (h *GeofenceHandler) UpdateGeofence(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id := c.Param("id")

	// Validate ID format
	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid geofence ID format",
		})
		return
	}

	var req models.UpdateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	fmt.Printf("✏️  Updating geofence %s\n", id)

	// ✅ Pass string ID and UpdateGeofenceRequest
	geofence, err := h.service.UpdateGeofence(ctx, id, req)
	if err != nil {
		fmt.Printf("❌ Failed to update geofence: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	fmt.Printf("✅ Geofence updated: %s\n", geofence.Name)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Geofence updated successfully",
		"data":    geofence,
	})
}

// DeleteGeofence deletes a geofence
func (h *GeofenceHandler) DeleteGeofence(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id := c.Param("id")

	// Validate ID format
	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid geofence ID format",
		})
		return
	}

	// Get geofence name before deleting (for logging)
	geofence, err := h.service.GetGeofenceByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	fmt.Printf("🗑️  Deleting geofence: %s (ID: %s)\n", geofence.Name, id)

	// ✅ Pass string ID
	if err := h.service.DeleteGeofence(ctx, id); err != nil {
		fmt.Printf("❌ Failed to delete geofence: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	fmt.Printf("✅ Geofence deleted: %s\n", geofence.Name)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Geofence deleted successfully",
		"data": gin.H{
			"id":   id,
			"name": geofence.Name,
		},
	})
}

// CheckUserInGeofence checks if user is within any active geofence
func (h *GeofenceHandler) CheckUserInGeofence(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var reqBody struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		UserID    string  `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if reqBody.Latitude < -90 || reqBody.Latitude > 90 || reqBody.Longitude < -180 || reqBody.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid coordinates",
		})
		return
	}

	// If no user ID provided, use authenticated user
	if reqBody.UserID == "" {
		reqBody.UserID = c.GetString("userID")
	}

	fmt.Printf("📍 Checking location: lat=%f, lng=%f, userID=%s\n", 
		reqBody.Latitude, reqBody.Longitude, reqBody.UserID)

	// ✅ Create CheckLocationRequest
	checkReq := models.CheckLocationRequest{
		Latitude:  reqBody.Latitude,
		Longitude: reqBody.Longitude,
		UserID:    reqBody.UserID,
	}

	result, err := h.service.CheckLocation(ctx, checkReq)
	if err != nil {
		fmt.Printf("❌ Failed to check location: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if result.IsWithinGeofence {
		fmt.Printf("✅ User inside geofence: %s (distance: %.2fm)\n", 
			result.Geofence.Name, result.Distance)
	} else {
		fmt.Printf("📏 User outside all geofences (distance: %.2fm)\n", result.Distance)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetActiveGeofences retrieves only active geofences
func (h *GeofenceHandler) GetActiveGeofences(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	geofences, err := h.service.GetActiveGeofences(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	fmt.Printf("📋 Retrieved %d active geofences\n", len(geofences))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Active geofences retrieved successfully",
		"data":    geofences,
		"total":   len(geofences),
	})
}

// Helper functions
func getUserName(c *gin.Context) string {
	if name := c.GetString("userName"); name != "" {
		return name
	}
	if name := c.GetString("full_name"); name != "" {
		return name
	}
	if email := c.GetString("email"); email != "" {
		return email
	}
	return "Unknown User"
}