// internal/service/work_schedule_service.go
package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkScheduleService interface {
	ListForManagerDepartment(ctx context.Context, managerUserID string) ([]models.WorkScheduleRowResponse, error)
	UpsertForManagerDepartment(ctx context.Context, managerUserID string, userID string, req models.UpsertWorkScheduleRequest) (*models.WorkScheduleRowResponse, error)
}

type workScheduleService struct {
	workScheduleRepo *repository.WorkScheduleRepoMongo
	departmentRepo   repository.DepartmentRepository
	userRepo         repository.UserRepository
}

func NewWorkScheduleService(
	workScheduleRepo *repository.WorkScheduleRepoMongo,
	departmentRepo repository.DepartmentRepository,
	userRepo repository.UserRepository,
) WorkScheduleService {
	return &workScheduleService{
		workScheduleRepo: workScheduleRepo,
		departmentRepo:   departmentRepo,
		userRepo:         userRepo,
	}
}

var hhmmRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

func parseHHMMToToday(t string) (time.Time, error) {
	if !hhmmRe.MatchString(t) {
		return time.Time{}, errors.New("format jam harus HH:mm")
	}
	now := time.Now()
	// bikin time date hari ini, jam dari string
	h := int((t[0]-'0')*10 + (t[1] - '0'))
	m := int((t[3]-'0')*10 + (t[4] - '0'))
	return time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location()), nil
}

func defaultWorkDays() []string {
	return []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat"}
}

func (s *workScheduleService) ListForManagerDepartment(ctx context.Context, managerUserID string) ([]models.WorkScheduleRowResponse, error) {
	// butuh method repo: FindByManagerID
	dept, err := s.departmentRepo.FindByManagerID(ctx, managerUserID)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, errors.New("departemen manager tidak ditemukan")
	}

	// butuh method repo: FindByDepartmentID
	users, err := s.userRepo.FindByDepartmentID(ctx, dept.ID.Hex())
	if err != nil {
		return nil, err
	}

	out := make([]models.WorkScheduleRowResponse, 0, len(users))
	for _, u := range users {
		ws, err := s.workScheduleRepo.FindByUserID(ctx, u.ID)
		if err != nil {
			return nil, err
		}

		workDays := defaultWorkDays()
		start := "08:00"
		end := "17:00"
		isActive := true

		if ws != nil {
			if len(ws.WorkDays) > 0 {
				workDays = ws.WorkDays
			}
			start = ws.StartTime.Format("15:04")
			end = ws.EndTime.Format("15:04")
			isActive = ws.IsActive
		}

		out = append(out, models.WorkScheduleRowResponse{
			UserID:     u.ID.Hex(),
			Name:       u.FullName,
			NIK:        u.PayrollNumber, // sesuaikan field user Anda
			Department: dept.Name,
			Position:   u.PositionName,  // sesuaikan field user Anda
			WorkDays:   workDays,
			StartTime:  start,
			EndTime:    end,
			IsActive:   isActive,
		})
	}

	return out, nil
}

func (s *workScheduleService) UpsertForManagerDepartment(ctx context.Context, managerUserID string, userID string, req models.UpsertWorkScheduleRequest) (*models.WorkScheduleRowResponse, error) {
	if len(req.WorkDays) == 0 {
		return nil, errors.New("hari kerja wajib diisi")
	}

	startT, err := parseHHMMToToday(req.StartTime)
	if err != nil {
		return nil, err
	}
	endT, err := parseHHMMToToday(req.EndTime)
	if err != nil {
		return nil, err
	}

	dept, err := s.departmentRepo.FindByManagerID(ctx, managerUserID)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, errors.New("departemen manager tidak ditemukan")
	}

	targetUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	// pastikan 1 departemen
	if targetUser.DepartmentID.Hex() != dept.ID.Hex() {
		return nil, errors.New("tidak diizinkan mengatur jadwal karyawan dari departemen lain")
	}

	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// AttendanceID: dari screenshot ada field ini.
	// Kalau Anda punya 1 attendance settings global, Anda bisa isi dari config/collection lain.
	// Untuk sementara: biarkan Zero ObjectID kalau belum dipakai.
	ws := &models.WorkSchedule{
		UserID:       userOID,
		AttendanceID: primitive.NilObjectID,
		StartTime:    startT,
		EndTime:      endT,
		WorkDays:     req.WorkDays,
		IsActive:     isActive,
	}

	if err := s.workScheduleRepo.UpsertByUserID(ctx, userOID, ws); err != nil {
		return nil, err
	}

	// Response
	return &models.WorkScheduleRowResponse{
		UserID:     targetUser.ID.Hex(),
		Name:       targetUser.FullName,
		NIK:        targetUser.PayrollNumber,
		Department: dept.Name,
		Position:   targetUser.PositionName,
		WorkDays:   req.WorkDays,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		IsActive:   isActive,
	}, nil
}