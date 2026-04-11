package service

import (
	"context"
	"errors"
	"strings"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PengajuanIzinCutiService interface {
	ListForManagerHR(ctx context.Context, status string, search string) ([]models.PengajuanIzinCutiApprovalResponse, error)
	GetForManagerHR(ctx context.Context, id string) (*models.PengajuanIzinCutiApprovalResponse, error)
	ApproveByManagerHR(ctx context.Context, id string, managerHRUserID string) (*models.PengajuanIzinCutiApprovalResponse, error)
	RejectByManagerHR(ctx context.Context, id string, managerHRUserID string) (*models.PengajuanIzinCutiApprovalResponse, error)
}

type pengajuanIzinCutiService struct {
	pengajuanRepo repository.PengajuanIzinCutiRepository
	userRepo      repository.UserRepository
}

func NewPengajuanIzinCutiService(pengajuanRepo repository.PengajuanIzinCutiRepository, userRepo repository.UserRepository) PengajuanIzinCutiService {
	return &pengajuanIzinCutiService{pengajuanRepo: pengajuanRepo, userRepo: userRepo}
}

func (s *pengajuanIzinCutiService) ListForManagerHR(ctx context.Context, status string, search string) ([]models.PengajuanIzinCutiApprovalResponse, error) {
	filter := bson.M{}
	status = strings.TrimSpace(strings.ToUpper(status))
	if status != "" && status != "ALL" {
		filter["status_manager_hr"] = status
	}

	pengajuans, err := s.pengajuanRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(pengajuans))
	for _, p := range pengajuans {
		userIDs = append(userIDs, p.UserID.Hex())
	}
	users, err := s.userRepo.FindByIDs(ctx, uniqueStrings(userIDs))
	if err != nil {
		return nil, err
	}
	userByID := make(map[string]models.User, len(users))
	for _, u := range users {
		userByID[u.ID.Hex()] = u
	}

	q := strings.ToLower(strings.TrimSpace(search))
	result := make([]models.PengajuanIzinCutiApprovalResponse, 0, len(pengajuans))
	for _, p := range pengajuans {
		u, ok := userByID[p.UserID.Hex()]
		if q != "" {
			if !ok {
				continue
			}
			if !strings.Contains(strings.ToLower(u.FullName), q) && !strings.Contains(strings.ToLower(u.PayrollNumber), q) {
				continue
			}
		}

		result = append(result, s.toApprovalResponse(p, ok, &u))
	}

	return result, nil
}

func (s *pengajuanIzinCutiService) GetForManagerHR(ctx context.Context, id string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	p, err := s.pengajuanRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(ctx, p.UserID.Hex())
	if err != nil {
		return &models.PengajuanIzinCutiApprovalResponse{Pengajuan: p.ToResponse()}, nil
	}
	resp := s.toApprovalResponse(*p, true, u)
	return &resp, nil
}

func (s *pengajuanIzinCutiService) ApproveByManagerHR(ctx context.Context, id string, managerHRUserID string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	return s.decideByManagerHR(ctx, id, managerHRUserID, models.StatusApproved)
}

func (s *pengajuanIzinCutiService) RejectByManagerHR(ctx context.Context, id string, managerHRUserID string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	return s.decideByManagerHR(ctx, id, managerHRUserID, models.StatusRejected)
}

func (s *pengajuanIzinCutiService) decideByManagerHR(ctx context.Context, id string, managerHRUserID string, statusManagerHR string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	if statusManagerHR != models.StatusApproved && statusManagerHR != models.StatusRejected {
		return nil, errors.New("status tidak valid")
	}

	managerOID, err := primitive.ObjectIDFromHex(managerHRUserID)
	if err != nil {
		return nil, errors.New("manager HR user ID tidak valid")
	}

	current, err := s.pengajuanRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.ToUpper(current.StatusManagerHR) != models.StatusPending {
		return nil, errors.New("pengajuan sudah diproses")
	}

	finalStatus := computeFinalStatus(strings.ToUpper(current.StatusKepalaDepartemen), statusManagerHR)
	updated, err := s.pengajuanRepo.UpdateManagerHRDecision(ctx, id, managerOID, statusManagerHR, finalStatus)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(ctx, updated.UserID.Hex())
	if err != nil {
		resp := models.PengajuanIzinCutiApprovalResponse{Pengajuan: updated.ToResponse()}
		return &resp, nil
	}
	resp := s.toApprovalResponse(*updated, true, u)
	return &resp, nil
}

func computeFinalStatus(statusKepalaDepartemen string, statusManagerHR string) string {
	if statusKepalaDepartemen == models.StatusRejected || statusManagerHR == models.StatusRejected {
		return models.StatusRejected
	}
	if statusKepalaDepartemen == models.StatusApproved && statusManagerHR == models.StatusApproved {
		return models.StatusApproved
	}
	return models.StatusPending
}

func (s *pengajuanIzinCutiService) toApprovalResponse(p models.LeaveRequest, hasUser bool, u *models.User) models.PengajuanIzinCutiApprovalResponse {
	resp := models.PengajuanIzinCutiApprovalResponse{Pengajuan: p.ToResponse()}
	if !hasUser || u == nil {
		return resp
	}
	emp := models.PengajuanIzinCutiApprovalEmployeeResponse{
		ID:             u.ID.Hex(),
		PayrollNumber:  u.PayrollNumber,
		FullName:       u.FullName,
		DepartmentName: u.DepartmentName,
		PositionName:   u.PositionName,
	}
	resp.Employee = &emp
	return resp
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		unique = append(unique, v)
	}
	return unique
}
