// pkg/models/employee_basic_salary.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmployeeBasicSalary struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	BasicSalary  int64              `json:"basic_salary" bson:"basic_salary"`
	EffectiveFrom time.Time         `json:"effective_from" bson:"effective_from"`
	IsActive     bool               `json:"is_active" bson:"is_active"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// Request create (tanggal dikirim string agar mudah dari form date input)
type CreateEmployeeBasicSalaryRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	BasicSalary   int64  `json:"basic_salary" binding:"required"`
	EffectiveFrom string `json:"effective_from" binding:"required"` // "YYYY-MM-DD"
	IsActive      *bool  `json:"is_active,omitempty"`
}

type UpdateEmployeeBasicSalaryRequest struct {
	BasicSalary   *int64  `json:"basic_salary,omitempty"`
	EffectiveFrom *string `json:"effective_from,omitempty"` // "YYYY-MM-DD"
	IsActive      *bool   `json:"is_active,omitempty"`
}

type EmployeeBasicSalaryResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	BasicSalary   int64     `json:"basic_salary"`
	EffectiveFrom time.Time `json:"effective_from"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (e *EmployeeBasicSalary) ToResponse() EmployeeBasicSalaryResponse {
	resp := EmployeeBasicSalaryResponse{
		ID:            e.ID.Hex(),
		BasicSalary:   e.BasicSalary,
		EffectiveFrom: e.EffectiveFrom,
		IsActive:      e.IsActive,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}

	if !e.UserID.IsZero() {
		resp.UserID = e.UserID.Hex()
	}

	return resp
}