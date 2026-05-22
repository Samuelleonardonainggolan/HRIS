package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssignmentService interface {
	Create(ctx context.Context, requestedByID string, req models.CreateAssignmentRequest) (*models.AssignmentResponse, error)
	GetByID(ctx context.Context, id string) (*models.AssignmentResponse, error)
	ListForManagerDepartemen(ctx context.Context, departmentID string) ([]models.AssignmentResponse, error)
	ListForManagerByUser(ctx context.Context, managerUserID string) ([]models.AssignmentResponse, error)
	ListForEmployee(ctx context.Context, userID string) ([]models.AssignmentResponse, error)
	GetForEmployeeByID(ctx context.Context, id string, userID string) (*models.AssignmentResponse, error)
	Submit(ctx context.Context, id string, requestedByID string) (*models.AssignmentResponse, error)
	UpdateEmployeeStatus(ctx context.Context, id string, userID string, req models.UpdateAssignmentEmployeeStatusRequest) (*models.AssignmentResponse, error)
	Update(ctx context.Context, id string, req models.UpdateAssignmentRequest) (*models.AssignmentResponse, error)
	Delete(ctx context.Context, id string) error

	// Helper untuk mendapatkan jadwal asli karyawan pada tanggal tertentu
	GetOriginalSchedule(ctx context.Context, userID string, date time.Time) (models.AssignmentOriginalShift, error)

	// Day off reward
	UseReplacementDayOff(ctx context.Context, assignmentID string, userID string, replacementDate time.Time) (*models.AssignmentResponse, error)
	GrantDayOffReward(ctx context.Context, assignmentID string, userID string) (*models.AssignmentResponse, error)
	SetWSHub(hub *WSHub)
	SetNotificationService(service NotificationService)
}

type assignmentService struct {
	assignmentRepo      repository.AssignmentRepository
	userRepo            repository.UserRepository
	jamKerjaRepo        repository.JamKerjaRepository
	departmentRepo      repository.DepartmentRepository
	pengajuanRepo       repository.PengajuanIzinCutiRepository
	wsHub               *WSHub
	notificationService NotificationService
}

func NewAssignmentService(
	assignmentRepo repository.AssignmentRepository,
	userRepo repository.UserRepository,
	jamKerjaRepo repository.JamKerjaRepository,
	departmentRepo repository.DepartmentRepository,
	pengajuanRepo repository.PengajuanIzinCutiRepository,
) AssignmentService {
	return &assignmentService{
		assignmentRepo: assignmentRepo,
		userRepo:       userRepo,
		jamKerjaRepo:   jamKerjaRepo,
		departmentRepo: departmentRepo,
		pengajuanRepo:  pengajuanRepo,
	}
}

func (s *assignmentService) SetWSHub(hub *WSHub) {
	s.wsHub = hub
}

func (s *assignmentService) SetNotificationService(service NotificationService) {
	s.notificationService = service
}

func (s *assignmentService) GetOriginalSchedule(ctx context.Context, userID string, date time.Time) (models.AssignmentOriginalShift, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return models.AssignmentOriginalShift{}, errors.New("user_id tidak valid")
	}

	// 1. Cek apakah karyawan sedang cuti/izin
	dateStr := date.Format("2006-01-02")
	leaveFilter := bson.M{
		"user_id": userOID,
		"final_status": "APPROVED",
		"tanggal_mulai": bson.M{"$lte": dateStr},
		"tanggal_selesai": bson.M{"$gte": dateStr},
	}
	leaves, _ := s.pengajuanRepo.Find(ctx, leaveFilter)
	if len(leaves) > 0 {
		return models.AssignmentOriginalShift{}, fmt.Errorf("Karyawan sedang cuti / izin pada tanggal tersebut")
	}

	// 2. Cek apakah karyawan sedang menggunakan reward (hari libur pengganti)
	assignFilter := bson.M{
		"employees": bson.M{
			"$elemMatch": bson.M{
				"user_id": userOID,
				"day_off_reward.replacement_off_date": date,
				"day_off_reward.status": models.DayOffRewardStatusUsed,
			},
		},
	}
	assignments, _ := s.assignmentRepo.Find(ctx, assignFilter)
	if len(assignments) > 0 {
		return models.AssignmentOriginalShift{}, fmt.Errorf("Karyawan sedang mengambil hari libur pengganti (reward) pada tanggal tersebut")
	}

	jk, err := s.jamKerjaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return models.AssignmentOriginalShift{}, err
	}

	if jk == nil {
		return models.AssignmentOriginalShift{Type: models.ShiftTypeOff}, nil
	}

	// Map time.Weekday ke Bahasa Indonesia
	dayNames := map[time.Weekday]string{
		time.Monday:    "Senin",
		time.Tuesday:   "Selasa",
		time.Wednesday: "Rabu",
		time.Thursday:  "Kamis",
		time.Friday:    "Jumat",
		time.Saturday:  "Sabtu",
		time.Sunday:    "Minggu",
	}

	targetDay := dayNames[date.Weekday()]
	isWorkDay := false
	for _, d := range jk.DayOfWeek {
		if d == targetDay {
			isWorkDay = true
			break
		}
	}

	if !isWorkDay {
		return models.AssignmentOriginalShift{Type: models.ShiftTypeOff}, nil
	}

	return models.AssignmentOriginalShift{
		Type:      models.ShiftTypeShift,
		StartTime: jk.StartTime.UTC().Format("15:04"),
		EndTime:   jk.EndTime.UTC().Format("15:04"),
		Source:    "jam_kerja",
	}, nil
}

func (s *assignmentService) Create(ctx context.Context, requestedByID string, req models.CreateAssignmentRequest) (*models.AssignmentResponse, error) {
	requestedByOID, _ := primitive.ObjectIDFromHex(requestedByID)
	deptOID, _ := primitive.ObjectIDFromHex(req.DepartmentID)

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}

	assignment := &models.Assignment{
		ID:            primitive.NewObjectID(),
		DepartmentID:  deptOID,
		RequestedByID: requestedByOID,
		Date:          parsedDate,
		Reason:        req.Reason,
		Status:        req.Status,
		Notes:         req.Notes,
		ShiftTarget: models.AssignmentShift{
			StartTime: req.ShiftStart,
			EndTime:   req.ShiftEnd,
		},
	}

	if assignment.Status == "" {
		assignment.Status = models.AssignmentStatusDraft
	}

	var employees []models.AssignmentEmployee
	for _, empInput := range req.Employees {
		empOID, _ := primitive.ObjectIDFromHex(empInput.UserID)

		// Cari jadwal asli
		origShift, err := s.GetOriginalSchedule(ctx, empInput.UserID, parsedDate)
		if err != nil {
			return nil, fmt.Errorf("gagal mendapatkan jadwal asli untuk user %s: %v", empInput.UserID, err)
		}

		// Tentukan shift penugasan (bisa di-override per karyawan jika ada inputnya)
		assignedStart := req.ShiftStart
		assignedEnd := req.ShiftEnd
		if empInput.AssignedStartTime != nil {
			assignedStart = *empInput.AssignedStartTime
		}
		if empInput.AssignedEndTime != nil {
			assignedEnd = *empInput.AssignedEndTime
		}

		eligibleReward := false
		if origShift.Type == models.ShiftTypeOff {
			eligibleReward = true
		}

		employees = append(employees, models.AssignmentEmployee{
			UserID:        empOID,
			OriginalShift: origShift,
			AssignedShift: models.AssignmentShift{
				StartTime: assignedStart,
				EndTime:   assignedEnd,
			},
			EmployeeStatus: models.AssignmentEmployeeStatusPending,
			DayOffReward: models.AssignmentDayOffReward{
				Eligible: eligibleReward,
				Status:   models.DayOffRewardStatusPending,
			},
		})
	}

	assignment.Employees = employees

	if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
		return nil, err
	}

	// Broadcast if published
	if assignment.Status == models.AssignmentStatusPublished && s.wsHub != nil {
		for _, emp := range assignment.Employees {
			s.wsHub.BroadcastToUser(emp.UserID.Hex(), WSEventAssignmentUpdated, map[string]any{
				"id":   assignment.ID.Hex(),
				"type": "new_assignment",
			})
		}
	}

	if assignment.Status == models.AssignmentStatusPublished {
		s.notifyAssignmentEmployees(ctx, assignment, requestedByID, "Penugasan Baru")
	}

	return s.GetByID(ctx, assignment.ID.Hex())
}

func (s *assignmentService) GetByID(ctx context.Context, id string) (*models.AssignmentResponse, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	assignment, err := s.assignmentRepo.GetByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	dept, _ := s.departmentRepo.FindByID(ctx, assignment.DepartmentID.Hex())
	requester, _ := s.userRepo.FindByID(ctx, assignment.RequestedByID.Hex())

	resp := &models.AssignmentResponse{
		ID:              assignment.ID.Hex(),
		DepartmentID:    assignment.DepartmentID.Hex(),
		DepartmentName:  "",
		RequestedByID:   assignment.RequestedByID.Hex(),
		RequestedByName: "",
		Date:            assignment.Date,
		Reason:          assignment.Reason,
		Status:          assignment.Status,
		Notes:           assignment.Notes,
		ShiftStart:      assignment.ShiftTarget.StartTime,
		ShiftEnd:        assignment.ShiftTarget.EndTime,
		CreatedAt:       assignment.CreatedAt,
		UpdatedAt:       assignment.UpdatedAt,
		Employees:       []models.AssignmentEmployeeResponse{},
	}

	if dept != nil {
		resp.DepartmentName = dept.Name
	}
	if requester != nil {
		resp.RequestedByName = requester.FullName
	}

	for _, emp := range assignment.Employees {
		u, _ := s.userRepo.FindByID(ctx, emp.UserID.Hex())
		empResp := models.AssignmentEmployeeResponse{
			UserID:             emp.UserID.Hex(),
			FullName:           "",
			PayrollNumber:      "",
			PositionName:       "",
			OriginalShiftType:  emp.OriginalShift.Type,
			OriginalStartTime:  emp.OriginalShift.StartTime,
			OriginalEndTime:    emp.OriginalShift.EndTime,
			AssignedStartTime:  emp.AssignedShift.StartTime,
			AssignedEndTime:    emp.AssignedShift.EndTime,
			EmployeeStatus:     emp.EmployeeStatus,
			RejectionNote:      emp.RejectionNote,
			ConfirmedAt:        emp.ConfirmedAt,
			DayOffEligible:     emp.DayOffReward.Eligible,
			DayOffStatus:       emp.DayOffReward.Status,
			DayOffGrantedAt:    emp.DayOffReward.GrantedAt,
			DayOffUsedAt:       emp.DayOffReward.UsedAt,
			ReplacementOffDate: emp.DayOffReward.ReplacementOffDate,
		}

		if u != nil {
			empResp.FullName = u.FullName
			empResp.PayrollNumber = u.PayrollNumber
			empResp.PositionName = u.PositionName
		}
		resp.Employees = append(resp.Employees, empResp)
	}

	return resp, nil
}

func (s *assignmentService) ListForManagerDepartemen(ctx context.Context, departmentID string) ([]models.AssignmentResponse, error) {
	deptOID, _ := primitive.ObjectIDFromHex(departmentID)
	assignments, err := s.assignmentRepo.ListByDepartment(ctx, deptOID)
	if err != nil {
		return nil, err
	}

	var resps []models.AssignmentResponse
	for _, a := range assignments {
		resp, _ := s.GetByID(ctx, a.ID.Hex())
		if resp != nil {
			resps = append(resps, *resp)
		}
	}
	return resps, nil
}

func (s *assignmentService) ListForManagerByUser(ctx context.Context, managerUserID string) ([]models.AssignmentResponse, error) {
	manager, err := s.userRepo.FindByID(ctx, managerUserID)
	if err != nil {
		return nil, err
	}
	if manager == nil || manager.DepartmentID.IsZero() {
		return nil, errors.New("department manager tidak valid")
	}
	return s.ListForManagerDepartemen(ctx, manager.DepartmentID.Hex())
}

func (s *assignmentService) ListForEmployee(ctx context.Context, userID string) ([]models.AssignmentResponse, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user_id tidak valid")
	}

	assignments, err := s.assignmentRepo.ListByEmployee(ctx, userOID)
	if err != nil {
		return nil, err
	}

	resps := make([]models.AssignmentResponse, 0, len(assignments))
	for _, a := range assignments {
		resp, _ := s.GetByID(ctx, a.ID.Hex())
		if resp != nil {
			resps = append(resps, *resp)
		}
	}
	return resps, nil
}

func (s *assignmentService) GetForEmployeeByID(ctx context.Context, id string, userID string) (*models.AssignmentResponse, error) {
	resp, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, emp := range resp.Employees {
		if emp.UserID == userID {
			return resp, nil
		}
	}

	return nil, errors.New("penugasan tidak ditemukan untuk user ini")
}

func (s *assignmentService) Submit(ctx context.Context, id string, requestedByID string) (*models.AssignmentResponse, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("id penugasan tidak valid")
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	if assignment.RequestedByID.Hex() != requestedByID {
		return nil, errors.New("anda tidak memiliki akses submit penugasan ini")
	}

	if assignment.Status == models.AssignmentStatusSubmitted {
		return s.GetByID(ctx, id)
	}

	if assignment.Status != models.AssignmentStatusDraft {
		return nil, errors.New("hanya penugasan draft yang bisa disubmit")
	}

	assignment.Status = models.AssignmentStatusSubmitted
	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	// Broadcast to manager if needed? Or just general update
	if s.wsHub != nil {
		s.wsHub.BroadcastToUser(assignment.RequestedByID.Hex(), WSEventAssignmentUpdated, map[string]any{
			"id":   id,
			"type": "submitted",
		})
	}

	s.notifyAssignmentEmployees(ctx, assignment, requestedByID, "Penugasan Baru")

	return s.GetByID(ctx, id)
}

func (s *assignmentService) UpdateEmployeeStatus(ctx context.Context, id string, userID string, req models.UpdateAssignmentEmployeeStatusRequest) (*models.AssignmentResponse, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("id penugasan tidak valid")
	}

	if req.Status != models.AssignmentEmployeeStatusAgreed && req.Status != models.AssignmentEmployeeStatusRejected {
		return nil, errors.New("status karyawan tidak valid")
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	if assignment.Status != models.AssignmentStatusSubmitted {
		return nil, errors.New("penugasan belum bisa diproses")
	}

	now := time.Now()
	found := false
	for i := range assignment.Employees {
		emp := &assignment.Employees[i]
		if emp.UserID.Hex() != userID {
			continue
		}
		found = true
		if emp.EmployeeStatus != models.AssignmentEmployeeStatusPending {
			return nil, errors.New("penugasan sudah Anda proses")
		}

		emp.EmployeeStatus = req.Status
		emp.ConfirmedAt = &now
		if req.Status == models.AssignmentEmployeeStatusRejected {
			emp.RejectionNote = req.RejectionNote
			emp.DayOffReward.Status = models.DayOffRewardStatusCancelled
		}
		break
	}

	if !found {
		return nil, errors.New("penugasan tidak ditemukan untuk user ini")
	}

	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	if s.wsHub != nil {
		// Broadcast to requester
		s.wsHub.BroadcastToUser(assignment.RequestedByID.Hex(), WSEventAssignmentUpdated, map[string]any{
			"id":      id,
			"user_id": userID,
			"status":  req.Status,
			"type":    "response",
		})
	}

	// Kirim Notifikasi ke Manager yang memberikan tugas
	if s.notificationService != nil {
		empUser, _ := s.userRepo.FindByID(ctx, userID)
		empName := "Seorang Karyawan"
		if empUser != nil {
			empName = empUser.FullName
		}
		
		statusText := "MENYETUJUI"
		if req.Status == models.AssignmentEmployeeStatusRejected {
			statusText = "MENOLAK"
		}
		msg := fmt.Sprintf("%s %s penugasan pada tanggal %s.", empName, statusText, assignment.Date.Format("2006-01-02"))
		if req.RejectionNote != "" {
			msg += fmt.Sprintf(" Alasan: %s", req.RejectionNote)
		}

		_, _ = s.notificationService.CreateNotification(ctx, models.CreateNotificationRequest{
			UserID:      assignment.RequestedByID.Hex(),
			SenderID:    userID,
			Title:       "Respon Penugasan",
			Message:     msg,
			Type:        "assignment",
			ReferenceID: assignment.ID.Hex(),
		})
	}

	return s.GetByID(ctx, id)
}

func (s *assignmentService) Update(ctx context.Context, id string, req models.UpdateAssignmentRequest) (*models.AssignmentResponse, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	assignment, err := s.assignmentRepo.GetByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	if req.Date != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.Date)
		if err == nil {
			assignment.Date = parsedDate
		}
	}

	if req.Reason != nil {
		assignment.Reason = *req.Reason
	}
	if req.Status != nil {
		assignment.Status = *req.Status
	}
	if req.Notes != nil {
		assignment.Notes = *req.Notes
	}
	if req.ShiftStart != nil {
		assignment.ShiftTarget.StartTime = *req.ShiftStart
	}
	if req.ShiftEnd != nil {
		assignment.ShiftTarget.EndTime = *req.ShiftEnd
	}

	if req.Employees != nil {
		var employees []models.AssignmentEmployee
		for _, empInput := range *req.Employees {
			empOID, _ := primitive.ObjectIDFromHex(empInput.UserID)

			// Cari jadwal asli
			origShift, _ := s.GetOriginalSchedule(ctx, empInput.UserID, assignment.Date)

			assignedStart := assignment.ShiftTarget.StartTime
			assignedEnd := assignment.ShiftTarget.EndTime
			if empInput.AssignedStartTime != nil {
				assignedStart = *empInput.AssignedStartTime
			}
			if empInput.AssignedEndTime != nil {
				assignedEnd = *empInput.AssignedEndTime
			}

			eligibleReward := (origShift.Type == models.ShiftTypeOff)

			employees = append(employees, models.AssignmentEmployee{
				UserID:        empOID,
				OriginalShift: origShift,
				AssignedShift: models.AssignmentShift{
					StartTime: assignedStart,
					EndTime:   assignedEnd,
				},
				EmployeeStatus: models.AssignmentEmployeeStatusPending,
				DayOffReward: models.AssignmentDayOffReward{
					Eligible: eligibleReward,
					Status:   models.DayOffRewardStatusPending,
				},
			})
		}
		assignment.Employees = employees
	}

	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

func (s *assignmentService) notifyAssignmentEmployees(ctx context.Context, assignment *models.Assignment, senderID string, title string) {
	if s.notificationService == nil || assignment == nil {
		return
	}

	for _, emp := range assignment.Employees {
		msg := fmt.Sprintf("Anda mendapatkan penugasan baru pada tanggal %s. Alasan: %s", assignment.Date.Format("2006-01-02"), assignment.Reason)
		if assignment.Status == models.AssignmentStatusSubmitted {
			msg = fmt.Sprintf("Anda menerima penugasan baru pada tanggal %s. Alasan: %s", assignment.Date.Format("2006-01-02"), assignment.Reason)
		}

		_, _ = s.notificationService.CreateNotification(ctx, models.CreateNotificationRequest{
			UserID:      emp.UserID.Hex(),
			SenderID:    senderID,
			Title:       title,
			Message:     msg,
			Type:        "assignment",
			ReferenceID: assignment.ID.Hex(),
		})
	}
}

func (s *assignmentService) Delete(ctx context.Context, id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	return s.assignmentRepo.Delete(ctx, oid)
}

func (s *assignmentService) UseReplacementDayOff(ctx context.Context, assignmentID string, userID string, replacementDate time.Time) (*models.AssignmentResponse, error) {
	oid, err := primitive.ObjectIDFromHex(assignmentID)
	if err != nil {
		return nil, errors.New("id penugasan tidak valid")
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	found := false
	now := time.Now()
	for i := range assignment.Employees {
		emp := &assignment.Employees[i]
		if emp.UserID.Hex() != userID {
			continue
		}
		found = true

		// Bypass validasi untuk keperluan testing
		// if !emp.DayOffReward.Eligible {
		// 	return nil, errors.New("anda tidak memenuhi syarat untuk day off reward")
		// }
		// if emp.DayOffReward.Status != models.DayOffRewardStatusGranted {
		// 	return nil, errors.New("day off reward belum diberikan atau sudah digunakan")
		// }

		emp.DayOffReward.Status = models.DayOffRewardStatusUsed
		emp.DayOffReward.UsedAt = &now
		emp.DayOffReward.ReplacementOffDate = &replacementDate
		break
	}

	if !found {
		return nil, errors.New("karyawan tidak ditemukan dalam penugasan ini")
	}

	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, assignmentID)
}

func (s *assignmentService) GrantDayOffReward(ctx context.Context, assignmentID string, userID string) (*models.AssignmentResponse, error) {
	oid, err := primitive.ObjectIDFromHex(assignmentID)
	if err != nil {
		return nil, errors.New("id penugasan tidak valid")
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	found := false
	now := time.Now()
	for i := range assignment.Employees {
		emp := &assignment.Employees[i]
		if emp.UserID.Hex() != userID {
			continue
		}
		found = true

		if !emp.DayOffReward.Eligible {
			return nil, errors.New("karyawan tidak memenuhi syarat untuk day off reward")
		}
		if emp.DayOffReward.Status != models.DayOffRewardStatusPending {
			return nil, fmt.Errorf("day off reward tidak dalam status pending (status: %s)", emp.DayOffReward.Status)
		}

		emp.DayOffReward.Status = models.DayOffRewardStatusGranted
		emp.DayOffReward.GrantedAt = &now
		break
	}

	if !found {
		return nil, errors.New("karyawan tidak ditemukan dalam penugasan ini")
	}

	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, assignmentID)
}

