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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EmployeeBasicSalaryService interface {
	List(ctx context.Context, q string, department string, active *bool) ([]models.EmployeeBasicSalaryListItem, error)
	Create(ctx context.Context, req models.CreateEmployeeBasicSalaryRequest) (*models.EmployeeBasicSalaryResponse, error)
	GetActiveByUserID(ctx context.Context, userID string) (*models.EmployeeBasicSalaryResponse, error)
	UpdateActiveByUserID(ctx context.Context, userID string, req models.UpdateEmployeeBasicSalaryRequest) (*models.EmployeeBasicSalaryResponse, error)
	DeactivateActiveByUserID(ctx context.Context, userID string) error
	ListAvailableEmployees(ctx context.Context, q string) ([]models.AvailableEmployeeForBasicSalary, error)

	GetLatestByUserID(ctx context.Context, userID string) (*models.EmployeeBasicSalaryResponse, error)
UpdateBySalaryID(ctx context.Context, salaryID string, req models.UpdateEmployeeBasicSalaryRequest) (*models.EmployeeBasicSalaryResponse, error)
}

type employeeBasicSalaryService struct {
	salaryRepo repository.EmployeeBasicSalaryRepository
	userRepo   repository.UserRepository
}

func NewEmployeeBasicSalaryService(
	salaryRepo repository.EmployeeBasicSalaryRepository,
	userRepo repository.UserRepository,
) EmployeeBasicSalaryService {
	return &employeeBasicSalaryService{
		salaryRepo: salaryRepo,
		userRepo:   userRepo,
	}
}

func (s *employeeBasicSalaryService) GetLatestByUserID(ctx context.Context, userID string) (*models.EmployeeBasicSalaryResponse, error) {
	sal, err := s.salaryRepo.FindLatestByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if sal == nil {
		return nil, errors.New("gaji pokok tidak ditemukan")
	}
	resp := sal.ToResponse()
	return &resp, nil
}

func (s *employeeBasicSalaryService) UpdateBySalaryID(ctx context.Context, salaryID string, req models.UpdateEmployeeBasicSalaryRequest) (*models.EmployeeBasicSalaryResponse, error) {
	if req.BasicSalary != nil && *req.BasicSalary <= 0 {
		return nil, errors.New("basic_salary harus lebih dari 0")
	}

	// update fields basic_salary/is_active
	if err := s.salaryRepo.UpdateByID(ctx, salaryID, &req); err != nil {
		return nil, err
	}

	// update effective_from if provided
	if req.EffectiveFrom != nil && strings.TrimSpace(*req.EffectiveFrom) != "" {
		eff, err := parseDateYYYYMMDD(*req.EffectiveFrom)
		if err != nil {
			return nil, err
		}
		if err := s.salaryRepo.UpdateEffectiveFromByID(ctx, salaryID, eff); err != nil {
			return nil, err
		}
	}

	updated, err := s.salaryRepo.FindByID(ctx, salaryID)
	if err != nil {
		return nil, err
	}
	resp := updated.ToResponse()
	return &resp, nil
}

func (s *employeeBasicSalaryService) ListAvailableEmployees(ctx context.Context, q string) ([]models.AvailableEmployeeForBasicSalary, error) {
  qq := strings.ToLower(strings.TrimSpace(q))
  if len(qq) < 2 {
    return []models.AvailableEmployeeForBasicSalary{}, nil
  }

  // Ambil semua user aktif (idealnya ada repo method khusus search)
  users, err := s.userRepo.FindAll(ctx) // <-- sesuaikan dengan repo Anda
  if err != nil {
    return nil, err
  }

  // filter by q (name/payroll/dept/pos)
  filtered := make([]models.User, 0)
  for _, u := range users {
    if !u.IsActive {
      continue
    }
    blob := strings.ToLower(u.FullName + " " + u.PayrollNumber + " " + u.DepartmentName + " " + u.PositionName)
    if strings.Contains(blob, qq) {
      filtered = append(filtered, u)
    }
  }

  // cek siapa yang sudah punya salary aktif
  ids := make([]primitive.ObjectID, 0, len(filtered))
  for _, u := range filtered {
    ids = append(ids, u.ID)
  }

  hasActive, err := s.salaryRepo.FindActiveByUserIDs(ctx, ids)
  if err != nil {
    return nil, err
  }

  // hanya yang belum punya salary aktif
  out := make([]models.AvailableEmployeeForBasicSalary, 0)
  for _, u := range filtered {
    if hasActive[u.ID] {
      continue
    }
    out = append(out, models.AvailableEmployeeForBasicSalary{
      ID:             u.ID.Hex(),
      FullName:       u.FullName,
      PayrollNumber:  u.PayrollNumber,
      DepartmentName: u.DepartmentName,
      PositionName:   u.PositionName,
    })
  }

  // limit 20
  if len(out) > 20 {
    out = out[:20]
  }

  return out, nil
}



func parseDateYYYYMMDD(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return time.Time{}, errors.New("format tanggal harus YYYY-MM-DD")
	}
	return t, nil
}

func (s *employeeBasicSalaryService) List(ctx context.Context, q string, department string, active *bool) ([]models.EmployeeBasicSalaryListItem, error) {
	filter := bson.M{}
	if active != nil {
		filter["is_active"] = *active
	}

	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	salaries, err := s.salaryRepo.FindAll(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	qq := strings.ToLower(strings.TrimSpace(q))
	dept := strings.TrimSpace(department)

	out := make([]models.EmployeeBasicSalaryListItem, 0, len(salaries))

	for _, sal := range salaries {
		user, err := s.userRepo.FindByID(ctx, sal.UserID.Hex())
		if err != nil || user == nil {
			continue
		}

		// filter department name (UI uses DepartmentName)
		if dept != "" && dept != "Semua Departemen" {
			if user.DepartmentName != dept {
				continue
			}
		}

		// q search: name, payroll_number, department_name, position_name
		if qq != "" {
			blob := strings.ToLower(
				user.FullName + " " +
					user.PayrollNumber + " " +
					user.DepartmentName + " " +
					user.PositionName,
			)
			if !strings.Contains(blob, qq) {
				continue
			}
		}

		out = append(out, models.EmployeeBasicSalaryListItem{
			ID:            sal.ID.Hex(),
			UserID:        sal.UserID.Hex(),
			FullName:      user.FullName,
			PayrollNumber: user.PayrollNumber,
			Department:    user.DepartmentName,
			Position:      user.PositionName,
			BasicSalary:   sal.BasicSalary,
			EffectiveFrom: sal.EffectiveFrom,
			IsActive:      sal.IsActive,
		})
	}

	return out, nil
}

func (s *employeeBasicSalaryService) Create(ctx context.Context, req models.CreateEmployeeBasicSalaryRequest) (*models.EmployeeBasicSalaryResponse, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return nil, errors.New("user_id wajib diisi")
	}
	if req.BasicSalary <= 0 {
		return nil, errors.New("basic_salary harus lebih dari 0")
	}

	// validate user exists
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil || user == nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}
	if !user.IsActive {
		return nil, errors.New("karyawan tidak aktif")
	}

	eff, err := parseDateYYYYMMDD(req.EffectiveFrom)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// enforce: only one active record per user
	if isActive {
		existing, err := s.salaryRepo.FindActiveByUserID(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			_ = s.salaryRepo.DeactivateByID(ctx, existing.ID.Hex())
		}
	}

	userOID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, errors.New("user_id tidak valid")
	}

	doc := &models.EmployeeBasicSalary{
		UserID:        userOID,
		BasicSalary:   req.BasicSalary,
		EffectiveFrom: eff,
		IsActive:      isActive,
	}

	if err := s.salaryRepo.Create(ctx, doc); err != nil {
		// if unique index triggers, return friendly
		return nil, errors.New("gagal menyimpan gaji pokok")
	}

	resp := doc.ToResponse()
	return &resp, nil
}

func (s *employeeBasicSalaryService) GetActiveByUserID(ctx context.Context, userID string) (*models.EmployeeBasicSalaryResponse, error) {
	sal, err := s.salaryRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if sal == nil {
		return nil, errors.New("gaji pokok aktif tidak ditemukan")
	}
	resp := sal.ToResponse()
	return &resp, nil
}

func (s *employeeBasicSalaryService) UpdateActiveByUserID(ctx context.Context, userID string, req models.UpdateEmployeeBasicSalaryRequest) (*models.EmployeeBasicSalaryResponse, error) {
	activeSal, err := s.salaryRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if activeSal == nil {
		return nil, errors.New("gaji pokok aktif tidak ditemukan")
	}

	if req.BasicSalary != nil && *req.BasicSalary <= 0 {
		return nil, errors.New("basic_salary harus lebih dari 0")
	}

	// update basic_salary and is_active
	if err := s.salaryRepo.UpdateByID(ctx, activeSal.ID.Hex(), &req); err != nil {
		return nil, err
	}

	// update effective_from if provided
	if req.EffectiveFrom != nil && strings.TrimSpace(*req.EffectiveFrom) != "" {
		eff, err := parseDateYYYYMMDD(*req.EffectiveFrom)
		if err != nil {
			return nil, err
		}
		if err := s.salaryRepo.UpdateEffectiveFromByID(ctx, activeSal.ID.Hex(), eff); err != nil {
			return nil, err
		}
	}

	updated, err := s.salaryRepo.FindByID(ctx, activeSal.ID.Hex())
	if err != nil {
		return nil, err
	}
	resp := updated.ToResponse()
	return &resp, nil
}

func (s *employeeBasicSalaryService) DeactivateActiveByUserID(ctx context.Context, userID string) error {
	activeSal, err := s.salaryRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if activeSal == nil {
		return errors.New("gaji pokok aktif tidak ditemukan")
	}
	return s.salaryRepo.DeactivateByID(ctx, activeSal.ID.Hex())
}