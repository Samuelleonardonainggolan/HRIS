// internal/service/user_service.go
package service

import (
	"context"
	"errors"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/auth"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	CreateEmployee(ctx context.Context, req models.CreateEmployeeRequest) (*models.UserResponse, *string, error)
	GetAllEmployees(ctx context.Context) ([]models.UserResponse, error)
	GetEmployeesMyDepartment(ctx context.Context, managerUserID string) ([]models.UserResponse, error)
	GetEmployeeByID(ctx context.Context, id string) (*models.UserResponse, error)
	UpdateEmployee(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteEmployee(ctx context.Context, id string) error
	ImportEmployees(ctx context.Context, employees []models.CreateEmployeeRequest) (int, []string, error)
}

type userService struct {
	userRepo       repository.UserRepository
	departmentRepo repository.DepartmentRepository
	positionRepo   repository.PositionRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	departmentRepo repository.DepartmentRepository,
	positionRepo repository.PositionRepository,
) UserService {
	return &userService{
		userRepo:       userRepo,
		departmentRepo: departmentRepo,
		positionRepo:   positionRepo,
	}
}

func (s *userService) CreateEmployee(ctx context.Context, req models.CreateEmployeeRequest) (*models.UserResponse, *string, error) {
	payrollNumber := strings.TrimSpace(req.PayrollNumber)
	if payrollNumber == "" {
		payrollNumber = strings.TrimSpace(req.NIK)
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		email = strings.TrimSpace(req.OfficeEmail)
	}

	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		phone = strings.TrimSpace(req.PhoneNumber)
	}

	fullName := strings.TrimSpace(req.FullName)
	birthDateStr := strings.TrimSpace(req.BirthDate)
	religion := strings.TrimSpace(req.Religion)
	lastEducation := strings.TrimSpace(req.LastEducation)
	yearEnrolled := strings.TrimSpace(req.YearEnrolled)
	employmentStatus := strings.TrimSpace(req.EmploymentStatus)
	departmentID := strings.TrimSpace(req.DepartmentID)
	positionID := strings.TrimSpace(req.PositionID)
	address := strings.TrimSpace(req.Address)
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = models.RoleStaf
	}

	if payrollNumber == "" {
		return nil, nil, errors.New("payroll_number is required")
	}
	if email == "" {
		return nil, nil, errors.New("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, nil, errors.New("invalid email")
	}
	if fullName == "" {
		return nil, nil, errors.New("full_name is required")
	}
	if birthDateStr == "" {
		return nil, nil, errors.New("birth_date is required")
	}
	if religion == "" {
		return nil, nil, errors.New("religion is required")
	}
	if lastEducation == "" {
		return nil, nil, errors.New("last_education is required")
	}
	if yearEnrolled == "" {
		return nil, nil, errors.New("year_enrolled is required")
	}
	if employmentStatus == "" {
		return nil, nil, errors.New("employment_status is required")
	}
	if departmentID == "" {
		return nil, nil, errors.New("department_id is required")
	}
	if positionID == "" {
		return nil, nil, errors.New("position_id is required")
	}
	if phone == "" {
		return nil, nil, errors.New("phone is required")
	}
	if address == "" {
		return nil, nil, errors.New("address is required")
	}

	existingEmail, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingEmail != nil {
		return nil, nil, errors.New("email already registered")
	}

	existingPayroll, err := s.userRepo.FindByPayrollNumber(ctx, payrollNumber)
	if err != nil {
		return nil, nil, err
	}
	if existingPayroll != nil {
		return nil, nil, errors.New("payroll_number already used")
	}

	department, err := s.departmentRepo.FindByID(ctx, departmentID)
	if err != nil || department == nil {
		return nil, nil, errors.New("department not found")
	}

	position, err := s.positionRepo.FindByID(ctx, positionID)
	if err != nil || position == nil {
		return nil, nil, errors.New("position not found")
	}

	if position.DepartmentID != department.ID {
		return nil, nil, errors.New("position does not belong to department")
	}

	birthDate, err := time.Parse("2006-01-02", birthDateStr)
	if err != nil {
		birthDate, err = time.Parse(time.RFC3339, birthDateStr)
	}
	if err != nil {
		return nil, nil, errors.New("invalid birth_date format")
	}

	rawPassword := strings.TrimSpace(req.Password)
	var tempPassword *string
	if rawPassword == "" {
		defaultPassword := "Password123"
		rawPassword = defaultPassword
		tempPassword = &defaultPassword
	} else if len(rawPassword) < 8 {
		return nil, nil, errors.New("password must be at least 8 characters")
	}

	hashedPassword, err := auth.HashPassword(rawPassword)
	if err != nil {
		return nil, nil, errors.New("failed to hash password")
	}

	now := time.Now()
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := &models.User{
		ID:               primitive.NewObjectID(),
		PayrollNumber:    payrollNumber,
		Email:            email,
		Password:         hashedPassword,
		FullName:         fullName,
		BirthDate:        birthDate,
		Religion:         religion,
		LastEducation:    lastEducation,
		YearEnrolled:     yearEnrolled,
		EmploymentStatus: employmentStatus,
		DepartmentID:     department.ID,
		DepartmentName:   department.Name,
		PositionID:       position.ID,
		PositionName:     position.Name,
		Phone:            phone,
		Address:          address,
		Role:             role,
		IsActive:         isActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	response := user.ToResponse()
	return &response, tempPassword, nil
}

func (s *userService) GetAllEmployees(ctx context.Context) ([]models.UserResponse, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	return responses, nil
}

func (s *userService) GetEmployeesMyDepartment(ctx context.Context, managerUserID string) ([]models.UserResponse, error) {
	manager, err := s.userRepo.FindByID(ctx, managerUserID)
	if err != nil {
		return nil, err
	}
	if manager == nil {
		return nil, errors.New("user not found")
	}
	if manager.DepartmentID.IsZero() {
		return nil, errors.New("department not set")
	}

	users, err := s.userRepo.FindByDepartment(ctx, manager.DepartmentID.Hex())
	if err != nil {
		return nil, err
	}

	responses := make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	return responses, nil
}

func (s *userService) GetEmployeeByID(ctx context.Context, id string) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("employee not found")
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) UpdateEmployee(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	if req.DepartmentID != "" {
		dept, err := s.departmentRepo.FindByID(ctx, req.DepartmentID)
		if err != nil || dept == nil {
			return nil, errors.New("department not found")
		}
		req.DepartmentName = dept.Name
	}
	if req.PositionID != "" {
		pos, err := s.positionRepo.FindByID(ctx, req.PositionID)
		if err != nil || pos == nil {
			return nil, errors.New("position not found")
		}
		req.PositionName = pos.Name
		if req.DepartmentID != "" {
			deptOID, deptErr := primitive.ObjectIDFromHex(req.DepartmentID)
			if deptErr != nil {
				return nil, errors.New("invalid department_id")
			}
			if pos.DepartmentID != deptOID {
				return nil, errors.New("position does not belong to department")
			}
		}
	}

	err := s.userRepo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) DeleteEmployee(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) ImportEmployees(ctx context.Context, employees []models.CreateEmployeeRequest) (int, []string, error) {
	if len(employees) == 0 {
		return 0, nil, errors.New("employees is required")
	}

	created := 0
	var failures []string
	for i, emp := range employees {
		_, _, err := s.CreateEmployee(ctx, emp)
		if err != nil {
			failures = append(failures, "index "+strconv.Itoa(i)+": "+err.Error())
			continue
		}
		created++
	}

	return created, failures, nil
}
