// pkg/models/face_embedding.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FaceEmbedding represents face recognition data
type FaceEmbedding struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID        primitive.ObjectID `json:"user_id" bson:"user_id"`
	FaceEmbedding []float64          `json:"face_embedding" bson:"face_embedding"`
	FaceImageURL  string             `json:"face_image_url,omitempty" bson:"face_image_url,omitempty"`
	IsActive      bool               `json:"is_active" bson:"is_active"`
	RegisteredAt  time.Time          `json:"registered_at" bson:"registered_at"`
	LastUpdatedAt time.Time          `json:"last_updated_at" bson:"last_updated_at"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
}

// FaceEmbeddingResponse represents face embedding response
type FaceEmbeddingResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	FaceImageURL  string    `json:"face_image_url,omitempty"`
	IsActive      bool      `json:"is_active"`
	RegisteredAt  time.Time `json:"registered_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponse converts FaceEmbedding to FaceEmbeddingResponse
func (f *FaceEmbedding) ToResponse() FaceEmbeddingResponse {
	return FaceEmbeddingResponse{
		ID:            f.ID.Hex(),
		UserID:        f.UserID.Hex(),
		FaceImageURL:  f.FaceImageURL,
		IsActive:      f.IsActive,
		RegisteredAt:  f.RegisteredAt,
		LastUpdatedAt: f.LastUpdatedAt,
		CreatedAt:     f.CreatedAt,
		UpdatedAt:     f.UpdatedAt,
	}
}