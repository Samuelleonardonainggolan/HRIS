// pkg/models/waktu_istirahat.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BreakTime struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id"`
	Date      time.Time          `json:"date" bson:"date"`
	StartTime time.Time          `json:"start_time" bson:"start_time"`
	EndTime   time.Time          `json:"end_time" bson:"end_time"`
	Status    string             `json:"status" bson:"status"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}