// internal/service/payroll_service.go
package service

import (
	"context"
	"errors"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
)

type PayrollService interface {
	GetPayrollForEmployee(ctx context.Context, userID string, month, year int) (*models.PayrollResponse, error)
}

type payrollService struct {
	payrollRepo repository.PayrollRepository
}

func NewPayrollService(payrollRepo repository.PayrollRepository) PayrollService {
	return &payrollService{
		payrollRepo: payrollRepo,
	}
}

func (s *payrollService) GetPayrollForEmployee(ctx context.Context, userID string, month, year int) (*models.PayrollResponse, error) {
	payroll, err := s.payrollRepo.FindByUserAndMonthYear(ctx, userID, month, year)
	if err != nil {
		return nil, err
	}
	if payroll == nil {
		return nil, errors.New("slip gaji tidak ditemukan untuk periode ini")
	}

	resp := payroll.ToResponse()
	return &resp, nil
}
