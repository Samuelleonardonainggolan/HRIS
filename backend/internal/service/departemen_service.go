// internal/service/department_service.go
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DepartmentService interface {
	CreateDepartment(ctx context.Context, req models.CreateDepartmentRequest) (*models.DepartmentResponse, error)
	GetDepartmentByID(ctx context.Context, id string) (*models.DepartmentResponse, error)
	GetAllDepartments(ctx context.Context) ([]models.DepartmentResponse, error)
	UpdateDepartment(ctx context.Context, id string, req models.UpdateDepartmentRequest) (*models.DepartmentResponse, error)
	DeleteDepartment(ctx context.Context, id string) error
}

type departmentService struct {
	departmentRepo repository.DepartmentRepository
	userRepo       repository.UserRepository
}

func NewDepartmentService(
	departmentRepo repository.DepartmentRepository,
	userRepo repository.UserRepository,
) DepartmentService {
	return &departmentService{
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
	}
}

func (s *departmentService) CreateDepartment(ctx context.Context, req models.CreateDepartmentRequest) (*models.DepartmentResponse, error) {
	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("nama departemen wajib diisi")
	}

	// Check if department name already exists
	existing, err := s.departmentRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("departemen dengan nama ini sudah ada")
	}

	// Check if code already exists (if provided)
	if req.Code != "" {
		existingCode, err := s.departmentRepo.FindByCode(ctx, req.Code)
		if err != nil {
			return nil, err
		}
		if existingCode != nil {
			return nil, errors.New("kode departemen sudah digunakan")
		}
	}

	// Set default icon if not provided
	if req.Icon == "" {
		req.Icon = "🏢"
	}

	// Create department
	department := &models.Department{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		TotalStaff:  0,
		IsActive:    true,
	}

	// Handle Manager ID
	if req.ManagerID != "" {
		managerOID, err := primitive.ObjectIDFromHex(req.ManagerID)
		if err != nil {
			return nil, errors.New("manager ID tidak valid")
		}

		// Verify manager exists
		manager, err := s.userRepo.FindByID(ctx, req.ManagerID)
		if err != nil {
			return nil, errors.New("manager tidak ditemukan")
		}

		if !manager.IsActive {
			return nil, errors.New("manager tidak aktif")
		}

		department.ManagerID = managerOID
		department.ManagerName = manager.FullName
	}

	err = s.departmentRepo.Create(ctx, department)
	if err != nil {
		return nil, errors.New("gagal membuat departemen")
	}

	response := department.ToResponse()
	return &response, nil
}

func (s *departmentService) GetDepartmentByID(ctx context.Context, id string) (*models.DepartmentResponse, error) {
	department, err := s.departmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := department.ToResponse()
	return &response, nil
}

func (s *departmentService) GetAllDepartments(ctx context.Context) ([]models.DepartmentResponse, error) {
	departments, err := s.departmentRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]models.DepartmentResponse, len(departments))
	for i, dept := range departments {
		responses[i] = dept.ToResponse()
	}

	return responses, nil
}

func (s *departmentService) UpdateDepartment(ctx context.Context, id string, req models.UpdateDepartmentRequest) (*models.DepartmentResponse, error) {
	// Check if department exists
	existing, err := s.departmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new name already used by another department
	if req.Name != "" && req.Name != existing.Name {
		existingName, err := s.departmentRepo.FindByName(ctx, req.Name)
		if err != nil {
			return nil, err
		}
		if existingName != nil && existingName.ID.Hex() != id {
			return nil, errors.New("nama departemen sudah digunakan")
		}
	}

	// Check if new code already used by another department
	if req.Code != "" && req.Code != existing.Code {
		existingCode, err := s.departmentRepo.FindByCode(ctx, req.Code)
		if err != nil {
			return nil, err
		}
		if existingCode != nil && existingCode.ID.Hex() != id {
			return nil, errors.New("kode departemen sudah digunakan")
		}
	}

	// Update department
	err = s.departmentRepo.Update(ctx, id, &req)
	if err != nil {
		return nil, errors.New("gagal memperbarui departemen")
	}

	// Get updated department
	updated, err := s.departmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := updated.ToResponse()
	return &response, nil
}

func (s *departmentService) DeleteDepartment(ctx context.Context, id string) error {
	// Check if department exists
	_, err := s.departmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Check if department has employees

	err = s.departmentRepo.Delete(ctx, id)
	if err != nil {
		return errors.New("gagal menghapus departemen")
	}

	return nil
}