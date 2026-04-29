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

type OvertimeRequestService interface {
	ListForManagerHR(ctx context.Context, status string, search string) ([]models.OvertimeApprovalResponse, error)
	GetForManagerHR(ctx context.Context, id string) (*models.OvertimeApprovalResponse, error)
	ApproveByManagerHR(ctx context.Context, id string, managerHRUserID string) (*models.OvertimeApprovalResponse, error)
	RejectByManagerHR(ctx context.Context, id string, managerHRUserID string, rejectionReason string) (*models.OvertimeApprovalResponse, error)
	ListForKepalaDepartemen(ctx context.Context, status string, search string, kepalaUserID string) ([]models.OvertimeApprovalResponse, error)
	GetForKepalaDepartemen(ctx context.Context, id string, kepalaUserID string) (*models.OvertimeApprovalResponse, error)
	ApproveByKepalaDepartemen(ctx context.Context, id string, kepalaUserID string) (*models.OvertimeApprovalResponse, error)
	RejectByKepalaDepartemen(ctx context.Context, id string, kepalaUserID string, rejectionReason string) (*models.OvertimeApprovalResponse, error)
}

type overtimeRequestService struct {
	overtimeRepo repository.OvertimeRequestRepository
	userRepo     repository.UserRepository
}

func NewOvertimeRequestService(
	overtimeRepo repository.OvertimeRequestRepository,
	userRepo repository.UserRepository,
) OvertimeRequestService {
	return &overtimeRequestService{
		overtimeRepo: overtimeRepo,
		userRepo:     userRepo,
	}
}

// ============================================================
// Manager HR
// ============================================================

func (s *overtimeRequestService) ListForManagerHR(ctx context.Context, status string, search string) ([]models.OvertimeApprovalResponse, error) {
	status = strings.TrimSpace(strings.ToUpper(status))

	filter := bson.M{}
	if status != "" && status != "ALL" {
		filter["final_status"] = status
	}

	requests, err := s.overtimeRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(requests))
	for _, r := range requests {
		userIDs = append(userIDs, r.UserID.Hex())
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
	result := make([]models.OvertimeApprovalResponse, 0, len(requests))
	for _, r := range requests {
		u, ok := userByID[r.UserID.Hex()]
		if q != "" {
			if !ok {
				continue
			}
			if !strings.Contains(strings.ToLower(u.FullName), q) && !strings.Contains(strings.ToLower(u.PayrollNumber), q) {
				continue
			}
		}
		result = append(result, s.toApprovalResponse(r, ok, &u))
	}

	return result, nil
}

func (s *overtimeRequestService) GetForManagerHR(ctx context.Context, id string) (*models.OvertimeApprovalResponse, error) {
	r, err := s.overtimeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.ToUpper(r.StatusKepalaDepartemen) == models.StatusPending {
		return nil, errors.New("pengajuan lembur belum diproses kepala departemen")
	}

	u, err := s.userRepo.FindByID(ctx, r.UserID.Hex())
	if err != nil {
		resp := models.OvertimeApprovalResponse{Overtime: r.ToResponse()}
		return &resp, nil
	}
	resp := s.toApprovalResponse(*r, true, u)
	return &resp, nil
}

func (s *overtimeRequestService) ApproveByManagerHR(ctx context.Context, id string, managerHRUserID string) (*models.OvertimeApprovalResponse, error) {
	return s.decideByManagerHR(ctx, id, managerHRUserID, models.StatusApproved, "")
}

func (s *overtimeRequestService) RejectByManagerHR(ctx context.Context, id string, managerHRUserID string, rejectionReason string) (*models.OvertimeApprovalResponse, error) {
	return s.decideByManagerHR(ctx, id, managerHRUserID, models.StatusRejected, rejectionReason)
}

func (s *overtimeRequestService) decideByManagerHR(ctx context.Context, id string, managerHRUserID string, status string, rejectionReason string) (*models.OvertimeApprovalResponse, error) {
	if status != models.StatusApproved && status != models.StatusRejected {
		return nil, errors.New("status tidak valid")
	}

	managerOID, err := primitive.ObjectIDFromHex(managerHRUserID)
	if err != nil {
		return nil, errors.New("manager HR user ID tidak valid")
	}

	current, err := s.overtimeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.ToUpper(current.StatusKepalaDepartemen) != models.StatusApproved {
		return nil, errors.New("pengajuan lembur belum disetujui kepala departemen")
	}
	if strings.ToUpper(current.StatusManagerHR) != models.StatusPending {
		return nil, errors.New("pengajuan lembur sudah diproses")
	}

	finalStatus := computeFinalStatus(strings.ToUpper(current.StatusKepalaDepartemen), status)
	updated, err := s.overtimeRepo.UpdateManagerHRDecision(ctx, id, managerOID, status, finalStatus, rejectionReason)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(ctx, updated.UserID.Hex())
	if err != nil {
		resp := models.OvertimeApprovalResponse{Overtime: updated.ToResponse()}
		return &resp, nil
	}
	resp := s.toApprovalResponse(*updated, true, u)
	return &resp, nil
}

// ============================================================
// Kepala Departemen
// ============================================================

func (s *overtimeRequestService) ListForKepalaDepartemen(ctx context.Context, status string, search string, kepalaUserID string) ([]models.OvertimeApprovalResponse, error) {
	kepala, err := s.userRepo.FindByID(ctx, kepalaUserID)
	if err != nil {
		return nil, errors.New("kepala departemen tidak ditemukan")
	}
	if kepala.DepartmentID.IsZero() {
		return nil, errors.New("departemen kepala departemen tidak valid")
	}

	deptUsers, err := s.userRepo.FindByDepartment(ctx, kepala.DepartmentID.Hex())
	if err != nil {
		return nil, err
	}
	if len(deptUsers) == 0 {
		return []models.OvertimeApprovalResponse{}, nil
	}	

	userOIDs := make([]primitive.ObjectID, 0, len(deptUsers))
	userByID := make(map[string]models.User, len(deptUsers))
	for _, u := range deptUsers {
		userOIDs = append(userOIDs, u.ID)
		userByID[u.ID.Hex()] = u
	}

	filter := bson.M{"user_id": bson.M{"$in": userOIDs}}
	status = strings.TrimSpace(strings.ToUpper(status))
	if status != "" && status != "ALL" {
		filter["status_kepala_departemen"] = status
	}

	requests, err := s.overtimeRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(strings.TrimSpace(search))
	result := make([]models.OvertimeApprovalResponse, 0, len(requests))
	for _, r := range requests {
		u, ok := userByID[r.UserID.Hex()]
		if q != "" {
			if !ok {
				continue
			}
			if !strings.Contains(strings.ToLower(u.FullName), q) && !strings.Contains(strings.ToLower(u.PayrollNumber), q) {
				continue
			}
		}
		result = append(result, s.toApprovalResponse(r, ok, &u))
	}

	return result, nil
}

func (s *overtimeRequestService) GetForKepalaDepartemen(ctx context.Context, id string, kepalaUserID string) (*models.OvertimeApprovalResponse, error) {
	kepala, err := s.userRepo.FindByID(ctx, kepalaUserID)
	if err != nil {
		return nil, errors.New("kepala departemen tidak ditemukan")
	}
	if kepala.DepartmentID.IsZero() {
		return nil, errors.New("departemen kepala departemen tidak valid")
	}

	r, err := s.overtimeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(ctx, r.UserID.Hex())
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	if u.DepartmentID != kepala.DepartmentID {
		return nil, errors.New("akses ditolak")
	}
	resp := s.toApprovalResponse(*r, true, u)
	return &resp, nil
}

func (s *overtimeRequestService) ApproveByKepalaDepartemen(ctx context.Context, id string, kepalaUserID string) (*models.OvertimeApprovalResponse, error) {
	return s.decideByKepalaDepartemen(ctx, id, kepalaUserID, models.StatusApproved, "")
}

func (s *overtimeRequestService) RejectByKepalaDepartemen(ctx context.Context, id string, kepalaUserID string, rejectionReason string) (*models.OvertimeApprovalResponse, error) {
	return s.decideByKepalaDepartemen(ctx, id, kepalaUserID, models.StatusRejected, rejectionReason)
}

func (s *overtimeRequestService) decideByKepalaDepartemen(ctx context.Context, id string, kepalaUserID string, status string, rejectionReason string) (*models.OvertimeApprovalResponse, error) {
	if status != models.StatusApproved && status != models.StatusRejected {
		return nil, errors.New("status tidak valid")
	}

	kepalaOID, err := primitive.ObjectIDFromHex(kepalaUserID)
	if err != nil {
		return nil, errors.New("kepala departemen user ID tidak valid")
	}
	kepala, err := s.userRepo.FindByID(ctx, kepalaUserID)
	if err != nil {
		return nil, errors.New("kepala departemen tidak ditemukan")
	}
	if kepala.DepartmentID.IsZero() {
		return nil, errors.New("departemen kepala departemen tidak valid")
	}

	current, err := s.overtimeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.ToUpper(current.StatusKepalaDepartemen) != models.StatusPending {
		return nil, errors.New("pengajuan lembur sudah diproses")
	}

	requester, err := s.userRepo.FindByID(ctx, current.UserID.Hex())
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	if requester.DepartmentID != kepala.DepartmentID {
		return nil, errors.New("akses ditolak")
	}

	finalStatus := computeFinalStatus(status, strings.ToUpper(current.StatusManagerHR))
	updated, err := s.overtimeRepo.UpdateKepalaDepartemenDecision(ctx, id, kepalaOID, status, finalStatus, rejectionReason)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(ctx, updated.UserID.Hex())
	if err != nil {
		resp := models.OvertimeApprovalResponse{Overtime: updated.ToResponse()}
		return &resp, nil
	}
	resp := s.toApprovalResponse(*updated, true, u)
	return &resp, nil
}

// ============================================================
// Helper
// ============================================================

func (s *overtimeRequestService) toApprovalResponse(r models.OvertimeRequest, hasUser bool, u *models.User) models.OvertimeApprovalResponse {
	resp := models.OvertimeApprovalResponse{Overtime: r.ToResponse()}
	if !hasUser || u == nil {
		return resp
	}
	emp := models.OvertimeApprovalEmployeeResponse{
		ID:             u.ID.Hex(),
		PayrollNumber:  u.PayrollNumber,
		FullName:       u.FullName,
		DepartmentName: u.DepartmentName,
		PositionName:   u.PositionName,
	}
	resp.Employee = &emp
	return resp
}
