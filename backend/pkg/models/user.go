// pkg/models/user.go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// User roles constants
const (
    RoleManagerHR   = "manager_hr"
    RoleManagerDept = "manager_departemen"
    RoleAdminDept   = "admin_departemen"
    RoleStaf        = "staf"
)

type User struct {
    ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    NIK          string             `json:"nik" bson:"nik"`
    Email        string             `json:"email" bson:"email"`
    Password     string             `json:"-" bson:"password"`
    FullName     string             `json:"full_name" bson:"full_name"`
    Role         string             `json:"role" bson:"role"`
    DepartmentID primitive.ObjectID `json:"department_id" bson:"department_id"`
    PositionID   primitive.ObjectID `json:"position_id" bson:"position_id"`
    Avatar       string             `json:"avatar,omitempty" bson:"avatar,omitempty"`
    Phone        string             `json:"phone,omitempty" bson:"phone,omitempty"`
    Address      string             `json:"address,omitempty" bson:"address,omitempty"`
    JoinDate     time.Time          `json:"join_date" bson:"join_date"`
    IsActive     bool               `json:"is_active" bson:"is_active"`
    CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateUserRequest - Request for creating new user
type CreateUserRequest struct {
    NIK          string `json:"nik" binding:"required"`
    Email        string `json:"email" binding:"required,email"`
    Password     string `json:"password" binding:"required,min=8"`
    FullName     string `json:"full_name" binding:"required"`
    Role         string `json:"role" binding:"required"`
    DepartmentID string `json:"department_id" binding:"required"`
    PositionID   string `json:"position_id" binding:"required"`
    Phone        string `json:"phone,omitempty"`
    Address      string `json:"address,omitempty"`
}

// UpdateProfileRequest - Request for updating user profile
type UpdateProfileRequest struct {
    FullName string `json:"full_name,omitempty"`
    Phone    string `json:"phone,omitempty"`
    Address  string `json:"address,omitempty"`
    Avatar   string `json:"avatar,omitempty"`
}

// UpdateUserRequest - Request for admin to update user
type UpdateUserRequest struct {
    FullName     string `json:"full_name,omitempty"`
    DepartmentID string `json:"department_id,omitempty"`
    PositionID   string `json:"position_id,omitempty"`
    Phone        string `json:"phone,omitempty"`
    Address      string `json:"address,omitempty"`
    IsActive     *bool  `json:"is_active,omitempty"`
}

// ChangePasswordRequest - Request for changing password
type ChangePasswordRequest struct {
    OldPassword string `json:"old_password" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
}

// UserResponse - User data without password
type UserResponse struct {
    ID           string    `json:"id"`
    NIK          string    `json:"nik"`
    Email        string    `json:"email"`
    FullName     string    `json:"full_name"`
    Role         string    `json:"role"`
    DepartmentID string    `json:"department_id"`
    PositionID   string    `json:"position_id"`
    Avatar       string    `json:"avatar,omitempty"`
    Phone        string    `json:"phone,omitempty"`
    Address      string    `json:"address,omitempty"`
    JoinDate     time.Time `json:"join_date"`
    IsActive     bool      `json:"is_active"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// ToResponse - Convert User to UserResponse
func (u *User) ToResponse() UserResponse {
    return UserResponse{
        ID:           u.ID.Hex(),
        NIK:          u.NIK,
        Email:        u.Email,
        FullName:     u.FullName,
        Role:         u.Role,
        DepartmentID: u.DepartmentID.Hex(),
        PositionID:   u.PositionID.Hex(),
        Avatar:       u.Avatar,
        Phone:        u.Phone,
        Address:      u.Address,
        JoinDate:     u.JoinDate,
        IsActive:     u.IsActive,
        CreatedAt:    u.CreatedAt,
        UpdatedAt:    u.UpdatedAt,
    }
}

// UserWithDetails - User with populated department and position
type UserWithDetails struct {
    User
    Department *Department `json:"department,omitempty" bson:"department,omitempty"`
    Position   *Position   `json:"position,omitempty" bson:"position,omitempty"`
}

// UserListResponse - Response for list of users
type UserListResponse struct {
    ID         string `json:"id"`
    NIK        string `json:"nik"`
    Email      string `json:"email"`
    FullName   string `json:"full_name"`
    Role       string `json:"role"`
    Department string `json:"department"`
    Position   string `json:"position"`
    IsActive   bool   `json:"is_active"`
}