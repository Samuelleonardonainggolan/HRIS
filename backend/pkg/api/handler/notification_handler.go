// pkg/api/handler/notification_handler.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(service service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

// GetNotifications - GET /api/v1/notifications
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	list, err := h.service.GetNotificationsByUserID(c.Request.Context(), userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil notifikasi: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Notifikasi berhasil diambil",
		"data":    list,
	})
}

// GetUnreadCount - GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	count, err := h.service.GetUnreadCount(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil jumlah unread: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Jumlah unread notifikasi berhasil diambil",
		"data":    gin.H{"unread_count": count},
	})
}

// MarkAsRead - PATCH /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID notifikasi wajib diisi",
		})
		return
	}

	err := h.service.MarkAsRead(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menandai notifikasi dibaca: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Notifikasi berhasil ditandai telah dibaca",
	})
}

// MarkAllAsRead - POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	err := h.service.MarkAllAsRead(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menandai semua notifikasi dibaca: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Semua notifikasi berhasil ditandai telah dibaca",
	})
}
