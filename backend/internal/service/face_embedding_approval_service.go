// internal/service/face_embedding_approval_service.go
package service

import (
	"context"
	"strings"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FaceEmbeddingApprovalService interface {
	List(ctx context.Context, q string, department string, active *bool) ([]models.FaceEmbeddingApprovalItem, error)
	GetByID(ctx context.Context, id string) (*models.FaceEmbeddingApprovalItem, error)
}

type faceEmbeddingApprovalService struct {
	faceRepo repository.FaceEmbeddingRepository
	userRepo repository.UserRepository
}

func NewFaceEmbeddingApprovalService(
	faceRepo repository.FaceEmbeddingRepository,
	userRepo repository.UserRepository,
) FaceEmbeddingApprovalService {
	return &faceEmbeddingApprovalService{
		faceRepo: faceRepo,
		userRepo: userRepo,
	}
}

func (s *faceEmbeddingApprovalService) List(ctx context.Context, q string, department string, active *bool) ([]models.FaceEmbeddingApprovalItem, error) {
	filter := bson.M{
		// hanya yang ada gambar untuk ditampilkan di UI
		"face_image_url": bson.M{"$ne": ""},
	}
	if active != nil {
		filter["is_active"] = *active
	}

	opts := options.Find().SetSort(bson.D{{Key: "registered_at", Value: -1}})
	embs, err := s.faceRepo.FindAll(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	qq := strings.ToLower(strings.TrimSpace(q))
	dept := strings.TrimSpace(department)

	out := make([]models.FaceEmbeddingApprovalItem, 0, len(embs))

	for _, e := range embs {
		user, err := s.userRepo.FindByID(ctx, e.UserID.Hex())
		if err != nil || user == nil {
			continue
		}

		// filter department
		if dept != "" && dept != "Semua Departemen" {
			if user.DepartmentName != dept {
				continue
			}
		}

		// q search: name/payroll/dept/position/email
		if qq != "" {
			blob := strings.ToLower(
				user.FullName + " " +
					user.PayrollNumber + " " +
					user.DepartmentName + " " +
					user.PositionName + " " +
					user.Email,
			)
			if !strings.Contains(blob, qq) {
				continue
			}
		}

		out = append(out, models.FaceEmbeddingApprovalItem{
			ID:           e.ID.Hex(),
			UserID:       e.UserID.Hex(),
			FaceImageURL: e.FaceImageURL,

			FullName:       user.FullName,
			PayrollNumber:  user.PayrollNumber,
			DepartmentName: user.DepartmentName,
			PositionName:   user.PositionName,
			Email:          user.Email,

			RegisteredAt: e.RegisteredAt,
			UpdatedAt:    e.UpdatedAt,

			IsActive:     e.IsActive,
			IsFirstLogin: e.IsFirstLogin,
		})
	}

	return out, nil
}

func (s *faceEmbeddingApprovalService) GetByID(ctx context.Context, id string) (*models.FaceEmbeddingApprovalItem, error) {
	e, err := s.faceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user, _ := s.userRepo.FindByID(ctx, e.UserID.Hex())

	item := &models.FaceEmbeddingApprovalItem{
		ID:           e.ID.Hex(),
		UserID:       e.UserID.Hex(),
		FaceImageURL: e.FaceImageURL,

		RegisteredAt: e.RegisteredAt,
		UpdatedAt:    e.UpdatedAt,

		IsActive:     e.IsActive,
		IsFirstLogin: e.IsFirstLogin,
	}

	if user != nil {
		item.FullName = user.FullName
		item.PayrollNumber = user.PayrollNumber
		item.DepartmentName = user.DepartmentName
		item.PositionName = user.PositionName
		item.Email = user.Email
	}

	return item, nil
}