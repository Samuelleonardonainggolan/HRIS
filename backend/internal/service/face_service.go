// internal/service/face_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/faceclient"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FaceService interface {
	Health(ctx context.Context) (bool, error)
	ExtractAndSaveEmbedding(ctx context.Context, userID string, photo []byte, filename string) error
	ProcessAttendance(ctx context.Context, userID string, latitude, longitude float64, recordType string, photo []byte, filename string) (*faceclient.AttendanceProcessResponse, error)
	ExtractEmbeddingOnly(ctx context.Context, userID string, photo []byte, filename string) ([]float32, error) // TAMBAHKAN INI
}

// Implementasi method ExtractEmbeddingOnly
func (s *faceService) ExtractEmbeddingOnly(ctx context.Context, userID string, photo []byte, filename string) ([]float32, error) {
	log.Printf("[FaceService] ExtractEmbeddingOnly for user: %s", userID)

	// Extract embedding from face client
	embedding, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		log.Printf("[FaceService] ExtractEmbedding error: %v", err)
		return nil, fmt.Errorf("extract embedding failed: %w", err)
	}

	if len(embedding) == 0 {
		log.Printf("[FaceService] Empty embedding received")
		return nil, errors.New("embedding kosong")
	}

	log.Printf("[FaceService] Embedding extracted, length: %d", len(embedding))
	return embedding, nil
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
	log.Printf("[FaceService] ExtractAndSaveEmbedding for user: %s", userID)

	// Verify user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// 🔥 EKSTRAK EMBEDDING REAL DARI FACE CLIENT
	embedding32, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		log.Printf("[FaceService] ExtractEmbedding error: %v", err)
		return fmt.Errorf("gagal mengekstrak wajah: %w", err)
	}

	if len(embedding32) == 0 {
		return errors.New("tidak ada wajah terdeteksi dalam foto")
	}

	// Validasi dimensi embedding (harus 128 atau 512)
	if len(embedding32) != 128 && len(embedding32) != 512 {
		return errors.New("dimensi embedding tidak valid")
	}

	log.Printf("[FaceService] Embedding berhasil diekstrak, length: %d", len(embedding32))

	// Check if face already exists for this user
	existingEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil && err != mongo.ErrNoDocuments {
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

// func (s *faceService) ExtractEmbeddingOnly(ctx context.Context, userID string, photo []byte, filename string) ([]float32, error) {
// 	log.Printf("[FaceService] ExtractEmbeddingOnly for user: %s", userID)

// 	// Extract embedding from face client
// 	embedding, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
// 	if err != nil {
// 		log.Printf("[FaceService] ExtractEmbedding error: %v", err)
// 		return nil, fmt.Errorf("extract embedding failed: %w", err)
// 	}

// 	if len(embedding) == 0 {
// 		log.Printf("[FaceService] Empty embedding received")
// 		return nil, errors.New("embedding kosong")
// 	}

// 	log.Printf("[FaceService] Embedding extracted, length: %d", len(embedding))
// 	return embedding, nil
// }
