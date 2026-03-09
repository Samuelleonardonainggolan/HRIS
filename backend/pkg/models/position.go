// pkg/models/position.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Position represents a job position/jabatan in the organization
type Position struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code         string             `json:"code" bson:"code"`
	Name         string             `json:"name" bson:"name"`
	DepartmentID primitive.ObjectID `json:"department_id" bson:"department_id"`
	Level        int                `json:"level" bson:"level"` // 1=Staff, 2=Supervisor, 3=Manager, 4=Director
	Description  string             `json:"description" bson:"description"`
	Requirements string             `json:"requirements" bson:"requirements"`
	SalaryRange  SalaryRange        `json:"salary_range" bson:"salary_range"`
	IsActive     bool               `json:"is_active" bson:"is_active"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// SalaryRange represents salary range for a position
type SalaryRange struct {
	Min      float64 `json:"min" bson:"min"`
	Max      float64 `json:"max" bson:"max"`
	Currency string  `json:"currency" bson:"currency"`
}

// CreatePositionRequest represents request to create position
type CreatePositionRequest struct {
	Code         string      `json:"code" binding:"required"`
	Name         string      `json:"name" binding:"required"`
	DepartmentID string      `json:"department_id" binding:"required"`
	Level        int         `json:"level" binding:"required"`
	Description  string      `json:"description"`
	Requirements string      `json:"requirements"`
	SalaryRange  SalaryRange `json:"salary_range"`
}

// UpdatePositionRequest represents request to update position
type UpdatePositionRequest struct {
	Code         string       `json:"code,omitempty"`
	Name         string       `json:"name,omitempty"`
	DepartmentID string       `json:"department_id,omitempty"`
	Level        int          `json:"level,omitempty"`
	Description  string       `json:"description,omitempty"`
	Requirements string       `json:"requirements,omitempty"`
	SalaryRange  *SalaryRange `json:"salary_range,omitempty"`
	IsActive     *bool        `json:"is_active,omitempty"`
}

// PositionResponse represents position response
type PositionResponse struct {
	ID           string      `json:"id"`
	Code         string      `json:"code"`
	Name         string      `json:"name"`
	DepartmentID string      `json:"department_id"`
	Level        int         `json:"level"`
	LevelName    string      `json:"level_name"`
	Description  string      `json:"description"`
	Requirements string      `json:"requirements"`
	SalaryRange  SalaryRange `json:"salary_range"`
	IsActive     bool        `json:"is_active"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// ToResponse converts Position to PositionResponse
func (p *Position) ToResponse() PositionResponse {
	return PositionResponse{
		ID:           p.ID.Hex(),
		Code:         p.Code,
		Name:         p.Name,
		DepartmentID: p.DepartmentID.Hex(),
		Level:        p.Level,
		LevelName:    GetLevelName(p.Level),
		Description:  p.Description,
		Requirements: p.Requirements,
		SalaryRange:  p.SalaryRange,
		IsActive:     p.IsActive,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// GetLevelName returns level name based on level number
func GetLevelName(level int) string {
	switch level {
	case 1:
		return "Staff"
	case 2:
		return "Supervisor"
	case 3:
		return "Manager"
	case 4:
		return "Director"
	case 5:
		return "C-Level"
	default:
		return "Unknown"
	}
}

// Position level constants
const (
	LevelStaff      = 1
	LevelSupervisor = 2
	LevelManager    = 3
	LevelDirector   = 4
	LevelCLevel     = 5
)