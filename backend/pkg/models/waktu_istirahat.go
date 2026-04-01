// pkg/models/waktu_istirahat.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WaktuIstirahat struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       string             `json:"user_id" bson:"user_id"`
	WaktuMulai   time.Time          `json:"waktu_mulai" bson:"waktu_mulai"`
	WaktuSelesai time.Time          `json:"waktu_selesai" bson:"waktu_selesai"`
	Status       string             `json:"status" bson:"status"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}