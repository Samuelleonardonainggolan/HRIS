// pkg/models/face_embedding.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FaceEmbedding struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID            primitive.ObjectID `json:"user_id" bson:"user_id"`
	FaceEmbedding     []float32          `json:"face_embedding" bson:"face_embedding"` // Gunakan []float32 konsisten
	FaceImageURL      string             `json:"face_image_url,omitempty" bson:"face_image_url,omitempty"`
	IsActive          bool               `json:"is_active" bson:"is_active"`
	IsFirstLogin      bool               `json:"is_first_login" bson:"is_first_login"`
	RegisteredAt      time.Time          `json:"registered_at" bson:"registered_at"`
	LastVerifiedAt    *time.Time         `json:"last_verified_at,omitempty" bson:"last_verified_at,omitempty"`
	VerificationCount int                `json:"verification_count" bson:"verification_count"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"` // Gunakan UpdatedAt, bukan LastUpdatedAt
}
