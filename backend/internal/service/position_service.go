package service

import (
	"context"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
)

type PositionService interface {
	GetAllPositions(ctx context.Context, departmentID string) ([]models.PositionResponse, error)
	GetPositionByID(ctx context.Context, id string) (*models.PositionResponse, error)
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
