// pkg/models/face_embedding.go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// FaceEmbedding - Face recognition embedding data
type FaceEmbedding struct {
    ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    UserID          primitive.ObjectID `json:"user_id" bson:"user_id"`
    FaceEmbedding   []float64          `json:"face_embedding" bson:"face_embedding"` // Vector embedding (512 or 128 dimensions)
    FaceImageURL    string             `json:"face_image_url,omitempty" bson:"face_image_url,omitempty"` // Original face photo URL
    IsActive        bool               `json:"is_active" bson:"is_active"`
    RegisteredAt    time.Time          `json:"registered_at" bson:"registered_at"`
    LastUpdatedAt   time.Time          `json:"last_updated_at" bson:"last_updated_at"`
    CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// FaceEmbeddingWithUser - Face embedding with user details
type FaceEmbeddingWithUser struct {
    FaceEmbedding
    User *User `json:"user,omitempty" bson:"user,omitempty"`
}

// RegisterFaceRequest - Request to register face
type RegisterFaceRequest struct {
    UserID       string    `json:"user_id" binding:"required"`
    FaceEmbedding []float64 `json:"face_embedding" binding:"required"`
    FaceImageURL string    `json:"face_image_url,omitempty"`
}

// UpdateFaceRequest - Request to update face embedding
type UpdateFaceRequest struct {
    FaceEmbedding []float64 `json:"face_embedding" binding:"required"`
    FaceImageURL string    `json:"face_image_url,omitempty"`
}

// FaceRecognitionRequest - Request for face recognition/verification
type FaceRecognitionRequest struct {
    FaceEmbedding []float64 `json:"face_embedding" binding:"required"`
    Threshold     float64   `json:"threshold,omitempty"` // Similarity threshold (default: 0.6)
}

// FaceRecognitionResponse - Response after face recognition
type FaceRecognitionResponse struct {
    Success    bool    `json:"success"`
    UserID     string  `json:"user_id,omitempty"`
    Similarity float64 `json:"similarity,omitempty"`
    Message    string  `json:"message"`
}