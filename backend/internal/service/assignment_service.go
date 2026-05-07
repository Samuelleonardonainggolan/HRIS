package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssignmentService interface {
	Create(ctx context.Context, requestedByID string, req models.CreateAssignmentRequest) (*models.AssignmentResponse, error)
	GetByID(ctx context.Context, id string) (*models.AssignmentResponse, error)
	ListForManagerDepartemen(ctx context.Context, departmentID string) ([]models.AssignmentResponse, error)
	Update(ctx context.Context, id string, req models.UpdateAssignmentRequest) (*models.AssignmentResponse, error)
	Delete(ctx context.Context, id string) error
	
	// Helper untuk mendapatkan jadwal asli karyawan pada tanggal tertentu
	GetOriginalSchedule(ctx context.Context, userID string, date time.Time) (models.AssignmentOriginalShift, error)
}

type assignmentService struct {
	assignmentRepo repository.AssignmentRepository
	userRepo       repository.UserRepository
	jamKerjaRepo   repository.JamKerjaRepository
	departmentRepo repository.DepartmentRepository
}

func NewAssignmentService(
	assignmentRepo repository.AssignmentRepository,
	userRepo repository.UserRepository,
	jamKerjaRepo repository.JamKerjaRepository,
	departmentRepo repository.DepartmentRepository,
) AssignmentService {
	return &assignmentService{
		assignmentRepo: assignmentRepo,
		userRepo:       userRepo,
		jamKerjaRepo:   jamKerjaRepo,
		departmentRepo: departmentRepo,
	}
}

func (s *assignmentService) GetOriginalSchedule(ctx context.Context, userID string, date time.Time) (models.AssignmentOriginalShift, error) {
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
			UserID:            emp.UserID.Hex(),
			FullName:          "",
			PayrollNumber:     "",
			PositionName:      "",
			OriginalShiftType: emp.OriginalShift.Type,
			OriginalStartTime: emp.OriginalShift.StartTime,
			OriginalEndTime:   emp.OriginalShift.EndTime,
			AssignedStartTime: emp.AssignedShift.StartTime,
			AssignedEndTime:   emp.AssignedShift.EndTime,
			EmployeeStatus:    emp.EmployeeStatus,
			RejectionNote:     emp.RejectionNote,
			ConfirmedAt:       emp.ConfirmedAt,
			DayOffEligible:    emp.DayOffReward.Eligible,
			DayOffStatus:      emp.DayOffReward.Status,
			DayOffGrantedAt:   emp.DayOffReward.GrantedAt,
			DayOffUsedAt:      emp.DayOffReward.UsedAt,
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

func (s *assignmentService) Delete(ctx context.Context, id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	return s.assignmentRepo.Delete(ctx, oid)
}
