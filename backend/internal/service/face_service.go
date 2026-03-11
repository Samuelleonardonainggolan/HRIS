// internal/service/face_service.go
package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/faceclient"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FaceService interface {
	Health(ctx context.Context) (bool, error)
	ExtractAndSaveEmbedding(ctx context.Context, userID string, photo []byte, filename string) error
	ProcessAttendance(ctx context.Context, userID string, latitude, longitude float64, recordType string, photo []byte, filename string) (*faceclient.AttendanceProcessResponse, error)
	VerifyFaceForAttendance(ctx context.Context, userID string, photo []byte, filename string) (bool, float64, error)
}

type faceService struct {
	userRepo          repository.UserRepository
	faceEmbeddingRepo repository.FaceEmbeddingRepository
	faceClient        *faceclient.Client
}

func NewFaceService(
	userRepo repository.UserRepository,
	faceEmbeddingRepo repository.FaceEmbeddingRepository,
	faceClient *faceclient.Client,
) FaceService {
	return &faceService{
		userRepo:          userRepo,
		faceEmbeddingRepo: faceEmbeddingRepo,
		faceClient:        faceClient,
	}
}

func (s *faceService) Health(ctx context.Context) (bool, error) {
	return s.faceClient.HealthCheck()
}

func (s *faceService) ExtractAndSaveEmbedding(ctx context.Context, userID string, photo []byte, filename string) error {
	// Verify user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Extract embedding from face recognition service
	embedding32, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		return err
	}

	if len(embedding32) == 0 {
		return errors.New("embedding kosong")
	}

	// Check if face embedding already exists
	existingEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	userOID, _ := primitive.ObjectIDFromHex(userID)

	if existingEmbedding != nil {
		// Update existing embedding
		existingEmbedding.FaceEmbedding = embedding32
		existingEmbedding.FaceImageURL = filename
		existingEmbedding.UpdatedAt = now
		existingEmbedding.IsFirstLogin = false
		return s.faceEmbeddingRepo.Update(ctx, existingEmbedding)
	} else {
		// Create new embedding
		newEmbedding := &models.FaceEmbedding{
			ID:                primitive.NewObjectID(),
			UserID:            userOID,
			FaceEmbedding:     embedding32,
			FaceImageURL:      filename,
			IsActive:          true,
			IsFirstLogin:      true,
			RegisteredAt:      now,
			VerificationCount: 0,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		return s.faceEmbeddingRepo.Create(ctx, newEmbedding)
	}
}

func (s *faceService) VerifyFaceForAttendance(ctx context.Context, userID string, photo []byte, filename string) (bool, float64, error) {
	// Get stored face embedding
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil || faceEmbedding == nil {
		return false, 0, errors.New("face not registered")
	}

	// Extract embedding from current photo
	currentEmbedding, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		return false, 0, err
	}

	// Calculate similarity
	similarity := s.cosineSimilarity(currentEmbedding, faceEmbedding.FaceEmbedding)
	matched := similarity >= 0.6

	if matched {
		// Update verification stats
		now := time.Now()
		faceEmbedding.LastVerifiedAt = &now
		faceEmbedding.VerificationCount++
		faceEmbedding.UpdatedAt = now
		s.faceEmbeddingRepo.Update(ctx, faceEmbedding)
	}

	return matched, similarity, nil
}

func (s *faceService) ProcessAttendance(
	ctx context.Context,
	userID string,
	latitude, longitude float64,
	recordType string,
	photo []byte,
	filename string,
) (*faceclient.AttendanceProcessResponse, error) {

	// First verify face
	matched, similarity, err := s.VerifyFaceForAttendance(ctx, userID, photo, filename)
	if err != nil {
		return nil, err
	}

	if !matched {
		return &faceclient.AttendanceProcessResponse{
			Approved: false,
			Message:  "Face does not match registered face",
			Face: &faceclient.FaceResult{
				Matched:    false,
				Similarity: similarity,
				Threshold:  0.6,
				Message:    "Face verification failed",
			},
		}, nil
	}

	// Get stored embedding
	faceEmbedding, _ := s.faceEmbeddingRepo.FindByUserID(ctx, userID)

	req := faceclient.ProcessAttendanceRequest{
		EmployeeID:      userID,
		StoredEmbedding: faceEmbedding.FaceEmbedding,
		Latitude:        latitude,
		Longitude:       longitude,
		RecordType:      recordType,
		Threshold:       0.6,
		RadiusM:         10000,
	}

	result, err := s.faceClient.ProcessAttendance(req, photo, filename)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *faceService) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
