package service

import (
	"context"
	"errors"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PositionService interface {
	GetAllPositions(ctx context.Context, departmentID string) ([]models.PositionResponse, error)
	GetPositionByID(ctx context.Context, id string) (*models.PositionResponse, error)
	UpdatePosition(ctx context.Context, id string, req models.UpdatePositionRequest) (*models.PositionResponse, error)
	CreatePosition(ctx context.Context, req models.CreatePositionRequest) (*models.PositionResponse, error)
	DeletePosition(ctx context.Context, id string) error
}

type positionService struct {
	positionRepo repository.PositionRepository
}

func NewPositionService(positionRepo repository.PositionRepository) PositionService {
	return &positionService{positionRepo: positionRepo}
}

func (s *positionService) GetAllPositions(ctx context.Context, departmentID string) ([]models.PositionResponse, error) {
	var positions []models.Position
	var err error

	if departmentID != "" {
		positions, err = s.positionRepo.FindByDepartment(ctx, departmentID)
	} else {
		positions, err = s.positionRepo.FindAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	responses := make([]models.PositionResponse, len(positions))
	for i, pos := range positions {
		responses[i] = pos.ToResponse()
	}
	return responses, nil
}

func (s *positionService) GetPositionByID(ctx context.Context, id string) (*models.PositionResponse, error) {
	position, err := s.positionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	response := position.ToResponse()
	return &response, nil
}

func (s *positionService) CreatePosition(ctx context.Context, req models.CreatePositionRequest) (*models.PositionResponse, error) {
	deptOID, err := primitive.ObjectIDFromHex(req.DepartmentID)
	if err != nil {
		return nil, errors.New("department_id tidak valid")
	}

	position := &models.Position{
		Code:         req.Code,
		Name:         req.Name,
		DepartmentID: deptOID,
		Level:        req.Level,
		Description:  req.Description,
		Requirements: req.Requirements,
		SalaryRange:  req.SalaryRange,
	}

	if err := s.positionRepo.Create(ctx, position); err != nil {
		return nil, err
	}

	resp := position.ToResponse()
	return &resp, nil
}

func (s *positionService) DeletePosition(ctx context.Context, id string) error {
	return s.positionRepo.Delete(ctx, id)
}

func (s *positionService) UpdatePosition(ctx context.Context, id string, req models.UpdatePositionRequest) (*models.PositionResponse, error) {
	_, err := s.positionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.positionRepo.Update(ctx, id, &req); err != nil {
		return nil, err
	}

	updated, err := s.positionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := updated.ToResponse()
	return &resp, nil
}
