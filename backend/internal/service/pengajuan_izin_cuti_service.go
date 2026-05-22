package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PengajuanIzinCutiService interface {
	ListForManagerHR(ctx context.Context, status string, search string) ([]models.PengajuanIzinCutiApprovalResponse, error)
	GetForManagerHR(ctx context.Context, id string) (*models.PengajuanIzinCutiApprovalResponse, error)
	ApproveByManagerHR(ctx context.Context, id string, managerHRUserID string) (*models.PengajuanIzinCutiApprovalResponse, error)
	RejectByManagerHR(ctx context.Context, id string, managerHRUserID string, rejectionReason string) (*models.PengajuanIzinCutiApprovalResponse, error)
	ListForKepalaDepartemen(ctx context.Context, status string, search string, kepalaDepartemenUserID string) ([]models.PengajuanIzinCutiApprovalResponse, error)
	GetForKepalaDepartemen(ctx context.Context, id string, kepalaDepartemenUserID string) (*models.PengajuanIzinCutiApprovalResponse, error)
	ApproveByKepalaDepartemen(ctx context.Context, id string, kepalaDepartemenUserID string) (*models.PengajuanIzinCutiApprovalResponse, error)
	RejectByKepalaDepartemen(ctx context.Context, id string, kepalaDepartemenUserID string, rejectionReason string) (*models.PengajuanIzinCutiApprovalResponse, error)
	GetLeaveBalance(ctx context.Context, userID string) (*models.LeaveBalanceResponse, error)
	SetWSHub(hub *WSHub) // inject WSHub untuk real-time broadcast
	SetNotificationService(service NotificationService)
}

type pengajuanIzinCutiService struct {
	pengajuanRepo repository.PengajuanIzinCutiRepository
	userRepo      repository.UserRepository
	db            *mongo.Database
	wsHub         *WSHub // untuk broadcast real-time events
	notificationService NotificationService
}

func NewPengajuanIzinCutiService(pengajuanRepo repository.PengajuanIzinCutiRepository, userRepo repository.UserRepository, db *mongo.Database) PengajuanIzinCutiService {
	return &pengajuanIzinCutiService{pengajuanRepo: pengajuanRepo, userRepo: userRepo, db: db, wsHub: nil}
}

// SetWSHub mengatur WebSocket hub untuk broadcast real-time events.
func (s *pengajuanIzinCutiService) SetWSHub(hub *WSHub) {
	s.wsHub = hub
}

func (s *pengajuanIzinCutiService) SetNotificationService(service NotificationService) {
	s.notificationService = service
}

func (s *pengajuanIzinCutiService) ListForManagerHR(ctx context.Context, status string, search string) ([]models.PengajuanIzinCutiApprovalResponse, error) {
	status = strings.TrimSpace(strings.ToUpper(status))

	filter := bson.M{}
	if status != "" && status != "ALL" {
		filter["final_status"] = status
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
	if strings.ToUpper(p.StatusKepalaDepartemen) == models.StatusPending {
		return nil, errors.New("pengajuan belum diproses kepala departemen")
	}

	u, err := s.userRepo.FindByID(ctx, p.UserID.Hex())
	if err != nil {
		return &models.PengajuanIzinCutiApprovalResponse{Pengajuan: p.ToResponse()}, nil
	}
	resp := s.toApprovalResponse(*p, true, u)
	return &resp, nil
}

func (s *pengajuanIzinCutiService) ApproveByManagerHR(ctx context.Context, id string, managerHRUserID string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	return s.decideByManagerHR(ctx, id, managerHRUserID, models.StatusApproved, "")
}

func (s *pengajuanIzinCutiService) RejectByManagerHR(ctx context.Context, id string, managerHRUserID string, rejectionReason string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	return s.decideByManagerHR(ctx, id, managerHRUserID, models.StatusRejected, rejectionReason)
}

func (s *pengajuanIzinCutiService) ListForKepalaDepartemen(ctx context.Context, status string, search string, kepalaDepartemenUserID string) ([]models.PengajuanIzinCutiApprovalResponse, error) {
	kepala, err := s.userRepo.FindByID(ctx, kepalaDepartemenUserID)
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
		return []models.PengajuanIzinCutiApprovalResponse{}, nil
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

	pengajuans, err := s.pengajuanRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
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

func (s *pengajuanIzinCutiService) GetForKepalaDepartemen(ctx context.Context, id string, kepalaDepartemenUserID string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	kepala, err := s.userRepo.FindByID(ctx, kepalaDepartemenUserID)
	if err != nil {
		return nil, errors.New("kepala departemen tidak ditemukan")
	}
	if kepala.DepartmentID.IsZero() {
		return nil, errors.New("departemen kepala departemen tidak valid")
	}

	p, err := s.pengajuanRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(ctx, p.UserID.Hex())
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	if u.DepartmentID != kepala.DepartmentID {
		return nil, errors.New("akses ditolak")
	}
	resp := s.toApprovalResponse(*p, true, u)
	return &resp, nil
}

func (s *pengajuanIzinCutiService) ApproveByKepalaDepartemen(ctx context.Context, id string, kepalaDepartemenUserID string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	return s.decideByKepalaDepartemen(ctx, id, kepalaDepartemenUserID, models.StatusApproved, "")
}

func (s *pengajuanIzinCutiService) RejectByKepalaDepartemen(ctx context.Context, id string, kepalaDepartemenUserID string, rejectionReason string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	return s.decideByKepalaDepartemen(ctx, id, kepalaDepartemenUserID, models.StatusRejected, rejectionReason)
}

func (s *pengajuanIzinCutiService) decideByKepalaDepartemen(ctx context.Context, id string, kepalaDepartemenUserID string, statusKepalaDepartemen string, rejectionReason string) (*models.PengajuanIzinCutiApprovalResponse, error) {
	if statusKepalaDepartemen != models.StatusApproved && statusKepalaDepartemen != models.StatusRejected {
		return nil, errors.New("status tidak valid")
	}

	kepalaOID, err := primitive.ObjectIDFromHex(kepalaDepartemenUserID)
	if err != nil {
		return nil, errors.New("kepala departemen user ID tidak valid")
	}
	kepala, err := s.userRepo.FindByID(ctx, kepalaDepartemenUserID)
	if err != nil {
		return nil, errors.New("kepala departemen tidak ditemukan")
	}
	if kepala.DepartmentID.IsZero() {
		return nil, errors.New("departemen kepala departemen tidak valid")
	}

	current, err := s.pengajuanRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.ToUpper(current.StatusKepalaDepartemen) != models.StatusPending {
		return nil, errors.New("pengajuan sudah diproses")
	}

	requester, err := s.userRepo.FindByID(ctx, current.UserID.Hex())
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	if requester.DepartmentID != kepala.DepartmentID {
		return nil, errors.New("akses ditolak")
	}

	finalStatus := computeFinalStatus(statusKepalaDepartemen, strings.ToUpper(current.StatusManagerHR))
	updated, err := s.pengajuanRepo.UpdateKepalaDepartemenDecision(ctx, id, kepalaOID, statusKepalaDepartemen, finalStatus, rejectionReason)
	if err != nil {
		return nil, err
	}

	// Kirim Notifikasi ke Karyawan
	if s.notificationService != nil {
		statusText := "DISETUJUI"
		if statusKepalaDepartemen == models.StatusRejected {
			statusText = "DITOLAK"
		}
		
		message := fmt.Sprintf("Pengajuan %s Anda telah %s oleh Kepala Departemen.", updated.TypeName, statusText)
		if rejectionReason != "" {
			message += fmt.Sprintf(" Alasan: %s", rejectionReason)
		}
		
		_, _ = s.notificationService.CreateNotification(ctx, models.CreateNotificationRequest{
			UserID:      updated.UserID.Hex(),
			SenderID:    kepalaDepartemenUserID,
			Title:       "Update Status Pengajuan",
			Message:     message,
			Type:        "leave_request",
			ReferenceID: updated.ID.Hex(),
		})
	}

	u, err := s.userRepo.FindByID(ctx, updated.UserID.Hex())
	if err != nil {
		resp := models.PengajuanIzinCutiApprovalResponse{Pengajuan: updated.ToResponse()}
		return &resp, nil
	}
	resp := s.toApprovalResponse(*updated, true, u)

	// Broadcast real-time event
	if s.wsHub != nil {
		s.wsHub.BroadcastToUser(updated.UserID.Hex(), WSEventLeaveUpdated, map[string]any{
			"action":  "status_updated",
			"status":  updated.StatusKepalaDepartemen,
			"message": "Pengajuan Anda telah di-review Kepala Departemen",
		})
		// Notifikasi ke manager lain (agar list terupdate)
		s.wsHub.BroadcastToAll(WSEventLeaveUpdated, map[string]any{
			"action": "leave_request_processed",
		})
	}

	return &resp, nil
}

func (s *pengajuanIzinCutiService) decideByManagerHR(ctx context.Context, id string, managerHRUserID string, statusManagerHR string, rejectionReason string) (*models.PengajuanIzinCutiApprovalResponse, error) {
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
	if strings.ToUpper(current.StatusKepalaDepartemen) != models.StatusApproved {
		return nil, errors.New("pengajuan belum disetujui kepala departemen")
	}
	if strings.ToUpper(current.StatusManagerHR) != models.StatusPending {
		return nil, errors.New("pengajuan sudah diproses")
	}

	finalStatus := computeFinalStatus(strings.ToUpper(current.StatusKepalaDepartemen), statusManagerHR)
	updated, err := s.pengajuanRepo.UpdateManagerHRDecision(ctx, id, managerOID, statusManagerHR, finalStatus, rejectionReason)
	if err != nil {
		return nil, err
	}

	// Kirim Notifikasi ke Karyawan
	if s.notificationService != nil {
		statusText := "DISETUJUI"
		if statusManagerHR == models.StatusRejected {
			statusText = "DITOLAK"
		}
		
		message := fmt.Sprintf("Pengajuan %s Anda telah %s oleh Manager HR.", updated.TypeName, statusText)
		if rejectionReason != "" {
			message += fmt.Sprintf(" Alasan: %s", rejectionReason)
		}
		
		_, _ = s.notificationService.CreateNotification(ctx, models.CreateNotificationRequest{
			UserID:      updated.UserID.Hex(),
			SenderID:    managerHRUserID,
			Title:       "Update Status Pengajuan",
			Message:     message,
			Type:        "leave_request",
			ReferenceID: updated.ID.Hex(),
		})
	}

	if strings.ToUpper(updated.FinalStatus) == models.StatusApproved && !updated.UserID.IsZero() {
		// Hanya kurangi kuota jika tipe pengajuan memang memotong kuota (kategori Cuti).
		// Pengajuan Izin tidak memotong kuota cuti.
		consumesQuota, err := requestTypeConsumesQuota(ctx, s.db, updated.RequestTypeID)
		if err != nil {
			return nil, err
		}
		if consumesQuota {
			leaveBalance, err := syncLeaveBalanceForYear(ctx, s.db, updated.UserID, updated.StartDate.Year())
			if err != nil {
				return nil, err
			}
			if updated.DaysTotal > leaveBalance.RemainingKuota {
				return nil, errors.New("sisa kuota cuti tidak mencukupi")
			}
			if updated.LeaveBalanceID == nil || updated.LeaveBalanceID.IsZero() {
				_, err = s.db.Collection("leave_request").UpdateOne(
					ctx,
					bson.M{"_id": updated.ID},
					bson.M{"$set": bson.M{"leave_balance_id": leaveBalance.ID, "updated_at": time.Now()}},
				)
				if err != nil {
					return nil, err
				}
				updated.LeaveBalanceID = &leaveBalance.ID
			}
		}
	}

	u, err := s.userRepo.FindByID(ctx, updated.UserID.Hex())
	if err != nil {
		resp := models.PengajuanIzinCutiApprovalResponse{Pengajuan: updated.ToResponse()}
		return &resp, nil
	}
	resp := s.toApprovalResponse(*updated, true, u)

	// Broadcast real-time event
	if s.wsHub != nil {
		s.wsHub.BroadcastToUser(updated.UserID.Hex(), WSEventLeaveUpdated, map[string]any{
			"action":  "status_updated",
			"status":  updated.FinalStatus,
			"message": "Pengajuan Anda telah di-review Manager HR",
		})
		s.wsHub.BroadcastToUser(updated.UserID.Hex(), WSEventStatsUpdated, map[string]any{
			"reason": "leave_approved",
		})
		// Notifikasi ke manager lain (agar list terupdate)
		s.wsHub.BroadcastToAll(WSEventLeaveUpdated, map[string]any{
			"action": "leave_request_processed",
		})
	}

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
		Avatar:         u.Avatar,
		DepartmentName: u.DepartmentName,
		PositionName:   u.PositionName,
	}
	resp.Employee = &emp
	return resp
}

func (s *pengajuanIzinCutiService) GetLeaveBalance(ctx context.Context, userID string) (*models.LeaveBalanceResponse, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	balance, err := syncLeaveBalanceForYear(ctx, s.db, userOID, time.Now().Year())
	if err != nil {
		return nil, err
	}

	resp := balance.ToResponse()
	return &resp, nil
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
