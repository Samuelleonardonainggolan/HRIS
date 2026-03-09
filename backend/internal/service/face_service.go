// internal/service/face_service.go
package service

import (
	"context"
	"errors"
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

	// Extract embedding from face recognition service (returns []float32)
	embedding32, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		return err
	}

	if len(embedding32) == 0 {
		return errors.New("embedding kosong")
	}

	// ✅ Convert []float32 to []float64 for database storage
	embedding64 := make([]float64, len(embedding32))
	for i, v := range embedding32 {
		embedding64[i] = float64(v)
	}

	// Check if face embedding already exists
	existingEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if existingEmbedding != nil {
		// Update existing embedding
		if err := s.faceEmbeddingRepo.Update(ctx, userID, embedding64); err != nil {
			return err
		}
	} else {
		// Create new embedding
		userOID, _ := primitive.ObjectIDFromHex(userID)
		newEmbedding := &models.FaceEmbedding{
			UserID:        userOID,
			FaceEmbedding: embedding64, // ✅ Use []float64
			FaceImageURL:  "",
			IsActive:      true,
			RegisteredAt:  time.Now(),
			LastUpdatedAt: time.Now(),
		}

		if err := s.faceEmbeddingRepo.Create(ctx, newEmbedding); err != nil {
			return err
		}
	}

	return nil
}

func (s *faceService) ProcessAttendance(
	ctx context.Context,
	userID string,
	latitude, longitude float64,
	recordType string,
	photo []byte,
	filename string,
) (*faceclient.AttendanceProcessResponse, error) {
	// Verify user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Get face embedding from database ([]float64)
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if faceEmbedding == nil || len(faceEmbedding.FaceEmbedding) == 0 {
		return nil, errors.New("embedding wajah belum terdaftar")
	}

	if !faceEmbedding.IsActive {
		return nil, errors.New("face embedding is inactive")
	}

	// ✅ Convert []float64 to []float32 for face recognition API
	embedding32 := make([]float32, len(faceEmbedding.FaceEmbedding))
	for i, v := range faceEmbedding.FaceEmbedding {
		embedding32[i] = float32(v)
	}

	// Process attendance with face recognition
	req := faceclient.ProcessAttendanceRequest{
		EmployeeID:      userID,
		StoredEmbedding: embedding32, // ✅ Use []float32
		Latitude:        latitude,
		Longitude:       longitude,
		RecordType:      recordType,
	}

	res, err := s.faceClient.ProcessAttendance(req, photo, filename)
	if err != nil {
		return nil, err
	}

	// TODO: Save attendance record to database
	_ = user // Will be used when creating attendance record
	_ = time.Now()

	return res, nil
}