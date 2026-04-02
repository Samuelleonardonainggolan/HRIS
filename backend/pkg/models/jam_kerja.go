// pkg/models/jam_kerja.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JamKerja struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	HariKerja    []string           `json:"hari_kerja" bson:"hari_kerja"` // ["Senin",...]
	WaktuMulai   time.Time          `json:"waktu_mulai" bson:"waktu_mulai"`
	WaktuSelesai time.Time          `json:"waktu_selesai" bson:"waktu_selesai"`
	Aktif        bool               `json:"aktif" bson:"aktif"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}