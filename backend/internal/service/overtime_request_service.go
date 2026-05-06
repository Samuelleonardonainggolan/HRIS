package service

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OvertimeRequestService interface {
	// Management (Manager HR / Manager Dept)
	ListOvertimeRequests(ctx context.Context, filter bson.M) ([]models.OvertimeRequestResponse, error)
	GetOvertimeRequestByID(ctx context.Context, id string) (*models.OvertimeRequestResponse, error)
	CreateOvertimeRequest(ctx context.Context, requestedByID string, req models.CreateOvertimeRequestRequest) (*models.OvertimeRequestResponse, error)
	UpdateOvertimeRequest(ctx context.Context, id string, req models.UpdateOvertimeRequestRequest) (*models.OvertimeRequestResponse, error)
	DeleteOvertimeRequest(ctx context.Context, id string) error

	// Employee Actions
	UpdateEmployeeStatus(ctx context.Context, overtimeID string, userID string, req models.UpdateEmployeeStatusRequest) error
	GetEmployeeOvertimeHistory(ctx context.Context, userID string) ([]models.OvertimeRequestResponse, error)

	// Legacy/Compat methods (to minimize handler changes)
	ListForManagerHR(ctx context.Context, status string, search string) ([]models.OvertimeRequestResponse, error)
	ListForKepalaDepartemen(ctx context.Context, status string, search string, kepalaUserID string) ([]models.OvertimeRequestResponse, error)

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

func (s *overtimeRequestService) CreateOvertimeRequest(ctx context.Context, requestedByID string, req models.CreateOvertimeRequestRequest) (*models.OvertimeRequestResponse, error) {
	reqOID, err := primitive.ObjectIDFromHex(requestedByID)
	if err != nil {
		return nil, errors.New("requested_by_id tidak valid")
	}
	deptOID, err := primitive.ObjectIDFromHex(req.DepartmentID)
	if err != nil {
		return nil, errors.New("department_id tidak valid")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid (YYYY-MM-DD)")
	}

	employees := make([]models.OvertimeEmployee, 0, len(req.Employees))
	for _, emp := range req.Employees {
		empOID, err := primitive.ObjectIDFromHex(emp.UserID)
		if err != nil {
			continue
		}
		employees = append(employees, models.OvertimeEmployee{
			UserID:         empOID,
			EmployeeStatus: models.EmployeeStatusPending,
		})
	}

	status := req.Status
	if status == "" {
		status = models.StatusDraft
	}

	overtime := &models.OvertimeRequest{
		DepartmentID:  deptOID,
		RequestedByID: reqOID,
		Date:          date,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Reason:        req.Reason,
		Status:        status,
		Employees:     employees,
	}

	created, err := s.overtimeRepo.Create(ctx, overtime)
	if err != nil {
		return nil, err
	}

	return s.toResponse(ctx, created), nil
}

func (s *overtimeRequestService) GetOvertimeRequestByID(ctx context.Context, id string) (*models.OvertimeRequestResponse, error) {
	req, err := s.overtimeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, req), nil
}

func (s *overtimeRequestService) ListOvertimeRequests(ctx context.Context, filter bson.M) ([]models.OvertimeRequestResponse, error) {
	requests, err := s.overtimeRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := make([]models.OvertimeRequestResponse, 0, len(requests))
	for _, r := range requests {
		resp = append(resp, *s.toResponse(ctx, &r))
	}
	return resp, nil
}

func (s *overtimeRequestService) UpdateOvertimeRequest(ctx context.Context, id string, req models.UpdateOvertimeRequestRequest) (*models.OvertimeRequestResponse, error) {
	updates := bson.M{}
	if req.Date != nil {
		date, _ := time.Parse("2006-01-02", *req.Date)
		updates["date"] = date
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	if req.Reason != nil {
		updates["reason"] = *req.Reason
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.LetterURL != nil {
		updates["letter_url"] = *req.LetterURL
	}

	updated, err := s.overtimeRepo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	return s.toResponse(ctx, updated), nil
}

func (s *overtimeRequestService) DeleteOvertimeRequest(ctx context.Context, id string) error {
	return s.overtimeRepo.Delete(ctx, id)
}

func (s *overtimeRequestService) UpdateEmployeeStatus(ctx context.Context, overtimeID string, userID string, req models.UpdateEmployeeStatusRequest) error {
	return s.overtimeRepo.UpdateEmployeeStatus(ctx, overtimeID, userID, req.Status, req.RejectionNote)
}

func (s *overtimeRequestService) GetEmployeeOvertimeHistory(ctx context.Context, userID string) ([]models.OvertimeRequestResponse, error) {
	uoid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return []models.OvertimeRequestResponse{}, nil
	}

	filter := bson.M{
		"$or": []bson.M{
			{"employees.user_id": uoid},
			{"requested_by_id": uoid},
		},
	}
	return s.ListOvertimeRequests(ctx, filter)
}

// ─── Legacy/Compat methods ────────────────────────────────────────────────

func (s *overtimeRequestService) ListForManagerHR(ctx context.Context, status string, search string) ([]models.OvertimeRequestResponse, error) {
	filter := bson.M{}
	if status != "" && status != "ALL" {
		filter["status"] = status
	}

	return s.ListOvertimeRequests(ctx, filter)
}

func (s *overtimeRequestService) ListForKepalaDepartemen(ctx context.Context, status string, search string, kepalaUserID string) ([]models.OvertimeRequestResponse, error) {
	kepala, _ := s.userRepo.FindByID(ctx, kepalaUserID)
	filter := bson.M{"department_id": kepala.DepartmentID}

	return s.ListOvertimeRequests(ctx, filter)
}

// ─── Internal Helpers ─────────────────────────────────────────────────────

func (s *overtimeRequestService) toResponse(ctx context.Context, r *models.OvertimeRequest) *models.OvertimeRequestResponse {
	empResp := make([]models.OvertimeEmployeeResponse, 0, len(r.Employees))
	requestedByName := ""
	departmentName := ""
	if requester, err := s.userRepo.FindByID(ctx, r.RequestedByID.Hex()); err == nil && requester != nil {
		requestedByName = requester.FullName
		departmentName = requester.DepartmentName
	}

	// Fetch user details for employees
	userIDs := make([]string, 0, len(r.Employees))
	for _, e := range r.Employees {
		userIDs = append(userIDs, e.UserID.Hex())
	}

	users, _ := s.userRepo.FindByIDs(ctx, userIDs)
	userMap := make(map[string]models.User)
	for _, u := range users {
		userMap[u.ID.Hex()] = u
	}

	for _, e := range r.Employees {
		u := userMap[e.UserID.Hex()]
		empResp = append(empResp, models.OvertimeEmployeeResponse{
			UserID:         e.UserID.Hex(),
			FullName:       u.FullName,
			PayrollNumber:  u.PayrollNumber,
			EmployeeStatus: e.EmployeeStatus,
			RejectionNote:  e.RejectionNote,
			ConfirmedAt:    e.ConfirmedAt,
		})
	}

	return &models.OvertimeRequestResponse{
		ID:              r.ID.Hex(),
		DepartmentID:    r.DepartmentID.Hex(),
		DepartmentName:  departmentName,
		RequestedByID:   r.RequestedByID.Hex(),
		RequestedByName: requestedByName,
		Date:            r.Date,
		StartTime:       r.StartTime,
		EndTime:         r.EndTime,
		Reason:          r.Reason,
		Status:          r.Status,
		Notes:           r.Notes,
		LetterURL:       r.LetterURL,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		Employees:       empResp,
	}
}
