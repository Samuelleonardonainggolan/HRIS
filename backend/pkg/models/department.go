// pkg/models/department.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Department represents a department in the organization
type Department struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code        string             `json:"code" bson:"code"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Icon        string             `json:"icon" bson:"icon"`
	ManagerID   primitive.ObjectID `json:"manager_id,omitempty" bson:"manager_id,omitempty"`
	ManagerName string             `json:"manager_name,omitempty" bson:"manager_name,omitempty"`
	TotalStaff  int                `json:"total_staff" bson:"total_staff"`
	IsActive    bool               `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateDepartmentRequest represents request to create department
type CreateDepartmentRequest struct {
	Code        string `json:"code" bson:"code"`
	Name        string `json:"name" binding:"required" bson:"name"`
	Description string `json:"description" bson:"description"`
	Icon        string `json:"icon" bson:"icon"`
	ManagerID   string `json:"manager_id,omitempty"`
}

// UpdateDepartmentRequest represents request to update department
type UpdateDepartmentRequest struct {
	Code        string `json:"code,omitempty" bson:"code,omitempty"`
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Icon        string `json:"icon,omitempty" bson:"icon,omitempty"`
	ManagerID   string `json:"manager_id,omitempty"`
}

// DepartmentResponse represents department response
type DepartmentResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	ManagerID   string    `json:"manager_id,omitempty"`
	ManagerName string    `json:"manager_name,omitempty"`
	TotalStaff  int       `json:"total_staff"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts Department to DepartmentResponse
func (d *Department) ToResponse() DepartmentResponse {
	response := DepartmentResponse{
		ID:          d.ID.Hex(),
		Code:        d.Code,
		Name:        d.Name,
		Description: d.Description,
		Icon:        d.Icon,
		ManagerName: d.ManagerName,
		TotalStaff:  d.TotalStaff,
		IsActive:    d.IsActive,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}

	// Only include manager_id if it's not zero
	if !d.ManagerID.IsZero() {
		response.ManagerID = d.ManagerID.Hex()
	}

	return response
}
