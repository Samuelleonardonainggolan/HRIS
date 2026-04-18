// pkg/models/break_time.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BreakTime struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id"` // string sesuai gambar
	Date      time.Time          `json:"date" bson:"date"`
	StartTime time.Time          `json:"start_time" bson:"start_time"`
	EndTime   time.Time          `json:"end_time" bson:"end_time"`
	Status    string             `json:"status" bson:"status"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type CreateBreakTimeRequest struct {
	UserID string `json:"user_id" binding:"required"`

	// date bisa "YYYY-MM-DD" (parsing di handler/service)
	Date string `json:"date" binding:"required"`

	// start_time/end_time bisa "HH:mm" atau ISO; parsing di handler/service
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`

	Status string `json:"status" binding:"required"`
}

type UpdateBreakTimeRequest struct {
	Date      *time.Time `json:"date,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Status    string     `json:"status,omitempty"`
}

type BreakTimeResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Date      time.Time `json:"date"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *BreakTime) ToResponse() BreakTimeResponse {
	return BreakTimeResponse{
		ID:        b.ID.Hex(),
		UserID:    b.UserID,
		Date:      b.Date,
		StartTime: b.StartTime,
		EndTime:   b.EndTime,
		Status:    b.Status,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}