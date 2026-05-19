// pkg/models/notification.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a notification in the system
type Notification struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	Title       string             `json:"title" bson:"title"`
	Message     string             `json:"message" bson:"message"`
	Type        string             `json:"type" bson:"type"` // e.g. "leave_request", "overtime", "assignment", "system"
	ReferenceID primitive.ObjectID `json:"reference_id,omitempty" bson:"reference_id,omitempty"`
	IsRead      bool               `json:"is_read" bson:"is_read"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateNotificationRequest represents request to create notification
type CreateNotificationRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Message     string `json:"message" binding:"required"`
	Type        string `json:"type" binding:"required"`
	ReferenceID string `json:"reference_id,omitempty"`
}

// UpdateNotificationRequest represents request to update notification status
type UpdateNotificationRequest struct {
	IsRead *bool `json:"is_read,omitempty" bson:"is_read,omitempty"`
}

// NotificationResponse represents notification response structure
type NotificationResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Type        string    `json:"type"`
	ReferenceID string    `json:"reference_id,omitempty"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts Notification to NotificationResponse
func (n *Notification) ToResponse() NotificationResponse {
	response := NotificationResponse{
		ID:        n.ID.Hex(),
		UserID:    n.UserID.Hex(),
		Title:     n.Title,
		Message:   n.Message,
		Type:      n.Type,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}

	if !n.ReferenceID.IsZero() {
		response.ReferenceID = n.ReferenceID.Hex()
	}

	return response
}
