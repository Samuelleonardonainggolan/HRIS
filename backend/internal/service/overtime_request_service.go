package service

import (
	"context"
	"errors"
	"strings"
	"time"

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

	// Employee
	CreateOvertimeRequest(ctx context.Context, req models.CreateOvertimeRequestRequest) (*models.OvertimeRequestResponse, error)
	GetMyOvertimeRequests(ctx context.Context, userID string) ([]models.OvertimeRequestResponse, error)
	GetMyOvertimeRequestByID(ctx context.Context, userID string, id string) (*models.OvertimeRequestResponse, error)
	UpdateMyOvertimeRequest(ctx context.Context, userID string, id string, req models.UpdateOvertimeRequestRequest) (*models.OvertimeRequestResponse, error)
	DeleteMyOvertimeRequest(ctx context.Context, userID string, id string) error
	SetWSHub(hub *WSHub)
}

type overtimeRequestService struct {
	overtimeRepo repository.OvertimeRequestRepository
	userRepo     repository.UserRepository
	wsHub        *WSHub
}

func NewOvertimeRequestService(
	overtimeRepo repository.OvertimeRequestRepository,
	userRepo repository.UserRepository,
) OvertimeRequestService {
	return &overtimeRequestService{
		overtimeRepo: overtimeRepo,
		userRepo:     userRepo,
		wsHub:        nil,
	}
}

func (s *overtimeRequestService) SetWSHub(hub *WSHub) {
	s.wsHub = hub
}

// ============================================================
// Employee
// ============================================================

func (s *overtimeRequestService) CreateOvertimeRequest(ctx context.Context, req models.CreateOvertimeRequestRequest) (*models.OvertimeRequestResponse, error) {
	userOID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	loc := time.FixedZone("WIB", 7*3600)

	date, err := time.ParseInLocation("2006-01-02", req.Date, loc)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}

	// Cek apakah sudah ada pengajuan lembur di tanggal yang sama (kecuali yang dibatalkan/ditolak)
	existingRequests, err := s.overtimeRepo.Find(ctx, bson.M{
		"user_id": userOID,
		"date":    date,
	})
	if err == nil {
		for _, r := range existingRequests {
			if r.FinalStatus != "CANCELLED" && r.FinalStatus != "REJECTED" {
				return nil, errors.New("pengajuan lembur untuk tanggal tersebut sudah ada")
			}
		}
	}

	// Assuming StartTime and EndTime are HH:MM
	var startTime, endTime time.Time
	if req.StartTime != "" {
		st, err := time.ParseInLocation("15:04", req.StartTime, loc)
		if err == nil {
			startTime = time.Date(date.Year(), date.Month(), date.Day(), st.Hour(), st.Minute(), 0, 0, loc)
		}
	}
	if req.EndTime != "" {
		et, err := time.ParseInLocation("15:04", req.EndTime, loc)
		if err == nil {
			endTime = time.Date(date.Year(), date.Month(), date.Day(), et.Hour(), et.Minute(), 0, 0, loc)
		}
	}

	overtimeReq := &models.OvertimeRequest{
		UserID:                 userOID,
		Date:                   date,
		StartTime:              startTime,
		EndTime:                endTime,
		Reason:                 req.Reason,
		Total:                  req.Total,
		StatusKepalaDepartemen: models.StatusPending,
		StatusManagerHR:        models.StatusPending,
		FinalStatus:            models.StatusPending,
	}

	created, err := s.overtimeRepo.Create(ctx, overtimeReq)
	if err != nil {
		return nil, err
	}

	resp := created.ToResponse()

	if s.wsHub != nil {
		// Broadcast ke user
		s.wsHub.BroadcastToUser(req.UserID, WSEventLeaveUpdated, map[string]any{
			"action":  "overtime_create",
			"message": "Pengajuan lembur berhasil dibuat",
		})
		// Broadcast ke semua (untuk manager)
		s.wsHub.BroadcastToAll(WSEventLeaveUpdated, map[string]any{
			"action": "new_overtime_request",
		})
	}

	return &resp, nil
}

func (s *overtimeRequestService) GetMyOvertimeRequests(ctx context.Context, userID string) ([]models.OvertimeRequestResponse, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	filter := bson.M{"user_id": userOID}
	requests, err := s.overtimeRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []models.OvertimeRequestResponse
	for _, r := range requests {
		results = append(results, r.ToResponse())
	}
	if results == nil {
		results = []models.OvertimeRequestResponse{}
	}

	return results, nil
}

func (s *overtimeRequestService) GetMyOvertimeRequestByID(ctx context.Context, userID string, id string) (*models.OvertimeRequestResponse, error) {
	r, err := s.overtimeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if r.UserID.Hex() != userID {
		return nil, errors.New("akses ditolak")
	}

	resp := r.ToResponse()
	return &resp, nil
}

func (s *overtimeRequestService) UpdateMyOvertimeRequest(ctx context.Context, userID string, id string, req models.UpdateOvertimeRequestRequest) (*models.OvertimeRequestResponse, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	loc := time.FixedZone("WIB", 7*3600)
	updates := bson.M{}

	if req.Date != nil {
		date, err := time.ParseInLocation("2006-01-02", *req.Date, loc)
		if err != nil {
			return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
		}
		updates["date"] = date
	}

	// Assuming StartTime and EndTime are HH:MM
	if req.StartTime != nil {
		st, err := time.ParseInLocation("15:04", *req.StartTime, loc)
		if err == nil {
			// We need date to construct time. If Date is not updated, we should ideally fetch the current date of the request.
			// But for simplicity, we can fetch the request first.
			r, err := s.overtimeRepo.FindByID(ctx, id)
			if err == nil {
				d := r.Date
				if req.Date != nil {
					d, _ = time.ParseInLocation("2006-01-02", *req.Date, loc)
				}
				updates["start_time"] = time.Date(d.Year(), d.Month(), d.Day(), st.Hour(), st.Minute(), 0, 0, loc)
			}
		}
	}

	if req.EndTime != nil {
		et, err := time.ParseInLocation("15:04", *req.EndTime, loc)
		if err == nil {
			r, err := s.overtimeRepo.FindByID(ctx, id)
			if err == nil {
				d := r.Date
				if req.Date != nil {
					d, _ = time.ParseInLocation("2006-01-02", *req.Date, loc)
				}
				updates["end_time"] = time.Date(d.Year(), d.Month(), d.Day(), et.Hour(), et.Minute(), 0, 0, loc)
			}
		}
	}

	if req.Reason != nil {
		updates["reason"] = *req.Reason
	}
	if req.Total != nil {
		updates["total"] = *req.Total
	}

	if len(updates) == 0 {
		return nil, errors.New("tidak ada data yang diperbarui")
	}

	updated, err := s.overtimeRepo.Update(ctx, id, userOID, updates)
	if err != nil {
		return nil, err
	}

	resp := updated.ToResponse()

	if s.wsHub != nil {
		s.wsHub.BroadcastToUser(userID, WSEventLeaveUpdated, map[string]any{
			"action":  "overtime_update",
			"message": "Pengajuan lembur berhasil diperbarui",
		})
	}

	return &resp, nil
}

func (s *overtimeRequestService) DeleteMyOvertimeRequest(ctx context.Context, userID string, id string) error {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("user ID tidak valid")
	}

	err = s.overtimeRepo.Delete(ctx, id, userOID)
	if err == nil && s.wsHub != nil {
		s.wsHub.BroadcastToUser(userID, WSEventLeaveUpdated, map[string]any{
			"action":  "overtime_delete",
			"message": "Pengajuan lembur berhasil dihapus",
		})
	}
	return err
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

	if s.wsHub != nil {
		s.wsHub.BroadcastToUser(updated.UserID.Hex(), WSEventLeaveUpdated, map[string]any{
			"action":  "overtime_status_updated",
			"status":  updated.FinalStatus,
			"message": "Pengajuan lembur Anda telah di-review Manager HR",
		})
		s.wsHub.BroadcastToAll(WSEventLeaveUpdated, map[string]any{
			"action": "overtime_processed",
		})
	}

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

	if s.wsHub != nil {
		s.wsHub.BroadcastToUser(updated.UserID.Hex(), WSEventLeaveUpdated, map[string]any{
			"action":  "overtime_status_updated",
			"status":  updated.StatusKepalaDepartemen,
			"message": "Pengajuan lembur Anda telah di-review Kepala Departemen",
		})
		s.wsHub.BroadcastToAll(WSEventLeaveUpdated, map[string]any{
			"action": "overtime_processed",
		})
	}

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
