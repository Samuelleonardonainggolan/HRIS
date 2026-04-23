// pkg/models/face_registration_request.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FaceRegistrationRequest represents face registration approval request
type FaceRegistrationRequest struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID          primitive.ObjectID `json:"user_id" bson:"user_id"`
	FaceEmbeddingID primitive.ObjectID `json:"face_embedding_id,omitempty" bson:"face_embedding_id,omitempty"`
	Status          string             `json:"status" bson:"status"`
	SubmittedAt     time.Time          `json:"submitted_at" bson:"submitted_at"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateFaceRegistrationRequestRequest represents request to create face registration request
type CreateFaceRegistrationRequestRequest struct {
	UserID          string `json:"user_id" binding:"required"`
	FaceEmbeddingID string `json:"face_embedding_id,omitempty"`
	Status          string `json:"status,omitempty"`
	SubmittedAt     string `json:"submitted_at,omitempty"` // optional, format YYYY-MM-DD or ISO; can be set by server
}

// UpdateFaceRegistrationRequestRequest represents request to update face registration request
type UpdateFaceRegistrationRequestRequest struct {
	FaceEmbeddingID *string `json:"face_embedding_id,omitempty"`
	Status          *string `json:"status,omitempty"`
	SubmittedAt     *string `json:"submitted_at,omitempty"`
}

// FaceRegistrationRequestResponse represents response
type FaceRegistrationRequestResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	FaceEmbeddingID string    `json:"face_embedding_id,omitempty"`
	Status          string    `json:"status"`
	SubmittedAt     time.Time `json:"submitted_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ToResponse converts FaceRegistrationRequest to FaceRegistrationRequestResponse
func (r *FaceRegistrationRequest) ToResponse() FaceRegistrationRequestResponse {
	resp := FaceRegistrationRequestResponse{
		ID:          r.ID.Hex(),
		UserID:      r.UserID.Hex(),
		Status:      r.Status,
		SubmittedAt: r.SubmittedAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if !r.FaceEmbeddingID.IsZero() {
		resp.FaceEmbeddingID = r.FaceEmbeddingID.Hex()
	}

	return resp
}