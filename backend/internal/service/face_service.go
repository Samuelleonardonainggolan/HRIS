// internal/service/face_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/faceclient"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/andikatampubolon10/hris-backend/pkg/storage"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ─── Interface ────────────────────────────────────────────────────────────────

type FaceService interface {
	Health(ctx context.Context) (bool, error)
	ExtractAndSaveEmbedding(ctx context.Context, userID string, photo []byte, filename string) error
	ProcessAttendance(ctx context.Context, userID string, latitude, longitude float64, recordType string, photo []byte, filename string) (*faceclient.AttendanceProcessResponse, error)
	ExtractEmbeddingOnly(ctx context.Context, userID string, photo []byte, filename string) ([]float32, error)
}

// ─── Struct ───────────────────────────────────────────────────────────────────

type faceService struct {
	userRepo          repository.UserRepository
	faceEmbeddingRepo repository.FaceEmbeddingRepository
	faceClient        *faceclient.Client
	publicBaseURL     string
	faceImageDir      string
	supabaseUploader  *storage.SupabaseUploader
}

func NewFaceService(
	userRepo repository.UserRepository,
	faceEmbeddingRepo repository.FaceEmbeddingRepository,
	faceClient *faceclient.Client,
	publicBaseURL string,
	faceImageDir string,
	supabaseUploader *storage.SupabaseUploader,
) FaceService {
	baseURL := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	imageDir := strings.TrimSpace(faceImageDir)
	if imageDir == "" {
		imageDir = filepath.Join("uploads", "face")
	}

	return &faceService{
		userRepo:          userRepo,
		faceEmbeddingRepo: faceEmbeddingRepo,
		faceClient:        faceClient,
		publicBaseURL:     baseURL,
		faceImageDir:      imageDir,
		supabaseUploader:  supabaseUploader,
	}
}

// ─── Health ───────────────────────────────────────────────────────────────────

func (s *faceService) Health(ctx context.Context) (bool, error) {
	return s.faceClient.HealthCheck()
}

// ─── ExtractAndSaveEmbedding ──────────────────────────────────────────────────
// Dipanggil saat registrasi wajah pegawai pertama kali.

func (s *faceService) ExtractAndSaveEmbedding(ctx context.Context, userID string, photo []byte, filename string) error {
	log.Printf("[FaceService] ExtractAndSaveEmbedding for user: %s", userID)

	// Verify user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Ekstrak embedding dari Face Client (FastAPI)
	embedding32, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		log.Printf("[FaceService] ExtractEmbedding error: %v", err)
		return fmt.Errorf("gagal mengekstrak wajah: %w", err)
	}

	if len(embedding32) == 0 {
		return errors.New("tidak ada wajah terdeteksi dalam foto")
	}

	// ✅ FIX: FastAPI mengembalikan 512-dim embedding, bukan 128
	// Validasi: harus 128 atau 512
	if len(embedding32) != 128 && len(embedding32) != 512 {
		return fmt.Errorf("dimensi embedding tidak valid: %d (expected 128 or 512)", len(embedding32))
	}

	log.Printf("[FaceService] Embedding berhasil diekstrak, dimension: %d", len(embedding32))

	// Check apakah embedding sudah ada untuk user ini
	existingEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	now := time.Now()
	userOID, _ := primitive.ObjectIDFromHex(userID)
	faceImageURL, err := s.saveFaceImageAndBuildURL(photo, filename)
	if err != nil {
		return fmt.Errorf("gagal menyimpan foto wajah: %w", err)
	}

	if existingEmbedding != nil {
		// Update embedding yang sudah ada
		existingEmbedding.FaceEmbedding = embedding32
		existingEmbedding.FaceImageURL = faceImageURL
		existingEmbedding.UpdatedAt = now
		existingEmbedding.IsFirstLogin = false
		log.Printf("[FaceService] Updating existing embedding for user: %s", userID)
		return s.faceEmbeddingRepo.Update(ctx, existingEmbedding)
	}

	// Buat embedding baru
	newEmbedding := &models.FaceEmbedding{
		ID:                primitive.NewObjectID(),
		UserID:            userOID,
		FaceEmbedding:     embedding32,
		FaceImageURL:      faceImageURL,
		IsActive:          true,
		IsFirstLogin:      true,
		RegisteredAt:      now,
		VerificationCount: 0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	log.Printf("[FaceService] Creating new embedding for user: %s", userID)
	return s.faceEmbeddingRepo.Create(ctx, newEmbedding)
}

func (s *faceService) saveFaceImageAndBuildURL(photo []byte, originalFilename string) (string, error) {
	// Upload ke Supabase Cloud Storage
	if s.supabaseUploader != nil {
		return s.supabaseUploader.UploadFile(photo, originalFilename, "face")
	}

	// Fallback ke local storage jika Supabase tidak dikonfigurasi
	if err := os.MkdirAll(s.faceImageDir, 0o755); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		ext = ".jpg"
	}

	storedName := fmt.Sprintf("face_%d%s", time.Now().UnixMilli(), ext)
	filePath := filepath.Join(s.faceImageDir, storedName)
	if err := os.WriteFile(filePath, photo, 0o644); err != nil {
		return "", err
	}

	urlPath := strings.ReplaceAll(filepath.ToSlash(filePath), "\\", "/")
	urlPath = strings.TrimPrefix(urlPath, "./")

	return fmt.Sprintf("%s/%s", s.publicBaseURL, urlPath), nil
}

// ─── ExtractEmbeddingOnly ─────────────────────────────────────────────────────
// Hanya ekstrak embedding tanpa menyimpan ke DB.
// Digunakan oleh handler untuk endpoint /face/extract-embedding.

func (s *faceService) ExtractEmbeddingOnly(ctx context.Context, userID string, photo []byte, filename string) ([]float32, error) {
	log.Printf("[FaceService] ExtractEmbeddingOnly for user: %s", userID)

	embedding, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		log.Printf("[FaceService] ExtractEmbedding error: %v", err)
		return nil, fmt.Errorf("extract embedding failed: %w", err)
	}

	if len(embedding) == 0 {
		log.Printf("[FaceService] Empty embedding received")
		return nil, errors.New("embedding kosong - tidak ada wajah terdeteksi")
	}

	log.Printf("[FaceService] Embedding extracted, dimension: %d", len(embedding))
	return embedding, nil
}

// ─── VerifyFaceForAttendance ──────────────────────────────────────────────────
// Internal method - bandingkan foto vs embedding tersimpan.

func (s *faceService) verifyFaceForAttendance(ctx context.Context, userID string, photo []byte, filename string) (bool, float64, error) {
	// Get stored face embedding
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil || faceEmbedding == nil {
		return false, 0, errors.New("wajah belum terdaftar. Silakan daftarkan wajah Anda")
	}

	// Extract embedding dari foto saat ini
	currentEmbedding, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		return false, 0, fmt.Errorf("gagal memproses foto: %w", err)
	}

	// Calculate similarity
	similarity := s.cosineSimilarity(currentEmbedding, faceEmbedding.FaceEmbedding)

	// ✅ Threshold sesuai FastAPI: SIMILARITY_THRESHOLD = 0.75
	const threshold = 0.75
	matched := similarity >= threshold

	if matched {
		// Update verification stats
		now := time.Now()
		faceEmbedding.LastVerifiedAt = &now
		faceEmbedding.VerificationCount++
		faceEmbedding.UpdatedAt = now
		_ = s.faceEmbeddingRepo.Update(ctx, faceEmbedding)
	}

	return matched, similarity, nil
}

// ─── ProcessAttendance ────────────────────────────────────────────────────────
// Digunakan oleh FaceHandler.ProcessAttendance (route /face/process).
// Berbeda dengan AttendanceHandler.ProcessAttendance (/attendance/process).

func (s *faceService) ProcessAttendance(
	ctx context.Context,
	userID string,
	latitude, longitude float64,
	recordType string,
	photo []byte,
	filename string,
) (*faceclient.AttendanceProcessResponse, error) {

	// Verifikasi wajah terlebih dahulu
	matched, similarity, err := s.verifyFaceForAttendance(ctx, userID, photo, filename)
	if err != nil {
		return nil, err
	}

	if !matched {
		return &faceclient.AttendanceProcessResponse{
			Approved: false,
			Message:  fmt.Sprintf("Wajah tidak cocok (similarity: %.1f%%)", similarity*100),
			Face: &faceclient.FaceResult{
				Matched:    false,
				Similarity: similarity,
				Threshold:  0.75,
				Message:    "Face verification failed",
			},
		}, nil
	}

	// Get stored embedding untuk dikirim ke FastAPI
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("gagal mengambil data wajah")
	}

	req := faceclient.ProcessAttendanceRequest{
		EmployeeID:      userID,
		StoredEmbedding: faceEmbedding.FaceEmbedding,
		Latitude:        latitude,
		Longitude:       longitude,
		RecordType:      recordType,
		Threshold:       0.75,
		RadiusM:         100, // ✅ Sesuai FastAPI GEOFENCE_RADIUS_M
	}

	result, err := s.faceClient.ProcessAttendance(req, photo, filename)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ─── Cosine Similarity ────────────────────────────────────────────────────────

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
