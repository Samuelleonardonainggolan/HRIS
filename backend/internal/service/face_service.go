package service

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/faceclient"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
)

type FaceService interface {
	Health(ctx context.Context) (bool, error)
	ExtractAndSaveEmbedding(ctx context.Context, userID string, photo []byte, filename string) error
	ProcessAttendance(ctx context.Context, userID string, latitude, longitude float64, recordType string, photo []byte, filename string) (*faceclient.AttendanceProcessResponse, error)
}

type faceService struct {
	userRepo   repository.UserRepository
	faceClient *faceclient.Client
}

func NewFaceService(userRepo repository.UserRepository, faceClient *faceclient.Client) FaceService {
	return &faceService{
		userRepo:   userRepo,
		faceClient: faceClient,
	}
}

func (s *faceService) Health(ctx context.Context) (bool, error) {
	return s.faceClient.HealthCheck()
}

func (s *faceService) ExtractAndSaveEmbedding(ctx context.Context, userID string, photo []byte, filename string) error {
	embedding, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		return err
	}
	if len(embedding) == 0 {
		return errors.New("embedding kosong")
	}
	if err := s.userRepo.UpdateFaceEmbedding(ctx, userID, embedding); err != nil {
		return err
	}
	return nil
}

func (s *faceService) ProcessAttendance(ctx context.Context, userID string, latitude, longitude float64, recordType string, photo []byte, filename string) (*faceclient.AttendanceProcessResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || len(user.FaceEmbedding) == 0 {
		return nil, errors.New("embedding wajah belum terdaftar")
	}
	req := faceclient.ProcessAttendanceRequest{
		EmployeeID:      userID,
		StoredEmbedding: user.FaceEmbedding,
		Latitude:        latitude,
		Longitude:       longitude,
		RecordType:      recordType,
	}
	res, err := s.faceClient.ProcessAttendance(req, photo, filename)
	if err != nil {
		return nil, err
	}
	_ = time.Now() // reserved for future attendance creation timestamps
	return res, nil
}
