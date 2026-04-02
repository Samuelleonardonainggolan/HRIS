// pkg/models/user.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role constants
const (
	RoleManagerHR         = "manager_hr"
	RoleManagerDepartemen = "manager_departemen"
	RoleAdminDepartemen   = "admin_departemen"
	RoleStaf              = "staf"
	RoleAccountant        = "accountant"
)

// User represents main user authentication and profile table
type User struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PayrollNumber    string             `json:"payroll_number" bson:"payroll_number"`
	Email            string             `json:"email" bson:"email"`
	Password         string             `json:"-" bson:"password"`
	FullName         string             `json:"full_name" bson:"full_name"`
	BirthDate        time.Time          `json:"birth_date" bson:"birth_date"`
	Religion         string             `json:"religion" bson:"religion"`
	LastEducation    string             `json:"last_education" bson:"last_education"`
	YearEnrolled     string             `json:"year_enrolled" bson:"year_enrolled"`
	EmploymentStatus string             `json:"employment_status" bson:"employment_status"`
	DepartmentID     primitive.ObjectID `json:"department_id" bson:"department_id"`
	DepartmentName   string             `json:"department_name" bson:"department_name"`
	PositionID       primitive.ObjectID `json:"position_id" bson:"position_id"`
	PositionName     string             `json:"position_name" bson:"position_name"`
	Phone            string             `json:"phone" bson:"phone"`
	Address          string             `json:"address" bson:"address"`
	Role             string             `json:"role" bson:"role"`
	IsActive         bool               `json:"is_active" bson:"is_active"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}

// UserResponse represents user response without password
type UserResponse struct {
	ID               string    `json:"id"`
	PayrollNumber    string    `json:"payroll_number"`
	Email            string    `json:"email"`
	FullName         string    `json:"full_name"`
	BirthDate        time.Time `json:"birth_date"`
	Religion         string    `json:"religion"`
	LastEducation    string    `json:"last_education"`
	YearEnrolled     string    `json:"year_enrolled"`
	EmploymentStatus string    `json:"employment_status"`
	DepartmentID     string    `json:"department_id"`
	DepartmentName   string    `json:"department_name"`
	PositionID       string    `json:"position_id"`
	PositionName     string    `json:"position_name"`
	Phone            string    `json:"phone"`
	Address          string    `json:"address"`
	Role             string    `json:"role"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateUserRequest represents request to create user/employee
type CreateUserRequest struct {
	PayrollNumber    string `json:"payroll_number" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=8"`
	FullName         string `json:"full_name" binding:"required"`
	BirthDate        string `json:"birth_date" binding:"required"`
	Religion         string `json:"religion" binding:"required"`
	LastEducation    string `json:"last_education" binding:"required"`
	YearEnrolled     string `json:"year_enrolled" binding:"required"`
	EmploymentStatus string `json:"employment_status" binding:"required"`
	DepartmentID     string `json:"department_id" binding:"required"`
	PositionID       string `json:"position_id" binding:"required"`
	Phone            string `json:"phone" binding:"required"`
	Address          string `json:"address" binding:"required"`
	Role             string `json:"role" binding:"required"`
}

type CreateEmployeeRequest struct {
	PayrollNumber    string `json:"payroll_number,omitempty"`
	NIK              string `json:"nik,omitempty"`
	Email            string `json:"email,omitempty"`
	OfficeEmail      string `json:"office_email,omitempty"`
	Password         string `json:"password,omitempty"`
	FullName         string `json:"full_name,omitempty"`
	BirthDate        string `json:"birth_date,omitempty"`
	Religion         string `json:"religion,omitempty"`
	LastEducation    string `json:"last_education,omitempty"`
	YearEnrolled     string `json:"year_enrolled,omitempty"`
	EmploymentStatus string `json:"employment_status,omitempty"`
	DepartmentID     string `json:"department_id,omitempty"`
	PositionID       string `json:"position_id,omitempty"`
	Phone            string `json:"phone,omitempty"`
	PhoneNumber      string `json:"phone_number,omitempty"`
	Address          string `json:"address,omitempty"`
	Role             string `json:"role,omitempty"`
	IsActive         *bool  `json:"is_active,omitempty"`
}

// UpdateUserRequest represents request to update user
type UpdateUserRequest struct {
	PayrollNumber    string `json:"payroll_number,omitempty"`
	FullName         string `json:"full_name,omitempty"`
	BirthDate        string `json:"birth_date,omitempty"`
	Religion         string `json:"religion,omitempty"`
	LastEducation    string `json:"last_education,omitempty"`
	YearEnrolled     string `json:"year_enrolled,omitempty"`
	EmploymentStatus string `json:"employment_status,omitempty"`
	DepartmentID     string `json:"department_id,omitempty"`
	DepartmentName   string `json:"-"`
	PositionID       string `json:"position_id,omitempty"`
	PositionName     string `json:"-"`
	Phone            string `json:"phone,omitempty"`
	Address          string `json:"address,omitempty"`
	IsActive         *bool  `json:"is_active,omitempty"`
}

// ChangePasswordRequest untuk POST /profile/change-password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:               u.ID.Hex(),
		PayrollNumber:    u.PayrollNumber,
		Email:            u.Email,
		FullName:         u.FullName,
		BirthDate:        u.BirthDate,
		Religion:         u.Religion,
		LastEducation:    u.LastEducation,
		YearEnrolled:     u.YearEnrolled,
		EmploymentStatus: u.EmploymentStatus,
		DepartmentID:     u.DepartmentID.Hex(),
		DepartmentName:   u.DepartmentName,
		PositionID:       u.PositionID.Hex(),
		PositionName:     u.PositionName,
		Phone:            u.Phone,
		Address:          u.Address,
		Role:             u.Role,
		IsActive:         u.IsActive,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}
