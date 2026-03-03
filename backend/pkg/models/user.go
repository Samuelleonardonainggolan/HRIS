// pkg/models/user.go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Role constants
const (
    RoleManagerHR    = "manager_hr"
    RoleManagerDept  = "manager_departemen"
    RoleAdminDept    = "admin_departemen"
    RoleStaf         = "staf"
)

type User struct {
    ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    NIK          string             `json:"nik" bson:"nik"`                    // Nomor Induk Karyawan
    Email        string             `json:"email" bson:"email" binding:"required,email"`
    Password     string             `json:"-" bson:"password"`
    FullName     string             `json:"full_name" bson:"full_name" binding:"required"`
    Role         string             `json:"role" bson:"role"`                  // manager_hr, manager_departemen, admin_departemen, staf
    Department   string             `json:"department" bson:"department"`      // IT, HR, Finance, Marketing, etc
    Position     string             `json:"position" bson:"position"`          // Jabatan
    Avatar       string             `json:"avatar,omitempty" bson:"avatar,omitempty"`
    Phone        string             `json:"phone,omitempty" bson:"phone,omitempty"`
    Address      string             `json:"address,omitempty" bson:"address,omitempty"`
    JoinDate     time.Time          `json:"join_date" bson:"join_date"`
    IsActive     bool               `json:"is_active" bson:"is_active"`
    CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// Department Model
type Department struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name        string             `json:"name" bson:"name"`                   // IT, HR, Finance, etc
    Code        string             `json:"code" bson:"code"`                   // IT001, HR001, etc
    Description string             `json:"description" bson:"description"`
    ManagerID   primitive.ObjectID `json:"manager_id" bson:"manager_id"`       // Manager Departemen
    CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Request Models
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
    NIK        string `json:"nik" binding:"required"`
    Email      string `json:"email" binding:"required,email"`
    Password   string `json:"password" binding:"required,min=6"`
    FullName   string `json:"full_name" binding:"required"`
    Role       string `json:"role" binding:"required,oneof=manager_hr manager_departemen admin_departemen staf"`
    Department string `json:"department" binding:"required"`
    Position   string `json:"position" binding:"required"`
    Phone      string `json:"phone"`
    Address    string `json:"address"`
}

type UpdateProfileRequest struct {
    FullName   string `json:"full_name"`
    Phone      string `json:"phone"`
    Address    string `json:"address"`
    Avatar     string `json:"avatar"`
    Department string `json:"department"`
    Position   string `json:"position"`
}

type UpdateUserRequest struct {
    FullName   string `json:"full_name"`
    Role       string `json:"role" binding:"oneof=manager_hr manager_departemen admin_departemen staf"`
    Department string `json:"department"`
    Position   string `json:"position"`
    Phone      string `json:"phone"`
    Address    string `json:"address"`
    IsActive   bool   `json:"is_active"`
}

type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

// Response Models
type LoginResponse struct {
    User         *User  `json:"user"`
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

// Permission helper
func (u *User) HasPermission(permission string) bool {
    permissions := map[string][]string{
        RoleManagerHR: {
            "user:create", "user:read", "user:update", "user:delete",
            "attendance:approve", "attendance:view_all",
            "department:manage", "report:view_all",
        },
        RoleManagerDept: {
            "user:read", "user:update_dept",
            "attendance:approve_dept", "attendance:view_dept",
            "report:view_dept",
        },
        RoleAdminDept: {
            "user:read", "user:create_staf", "user:update_staf",
            "attendance:view_dept", "attendance:input",
        },
        RoleStaf: {
            "attendance:submit", "profile:view", "profile:update",
        },
    }

    rolePermissions, exists := permissions[u.Role]
    if !exists {
        return false
    }

    for _, p := range rolePermissions {
        if p == permission {
            return true
        }
    }
    return false
}

// Check if user can access department data
func (u *User) CanAccessDepartment(department string) bool {
    if u.Role == RoleManagerHR {
        return true // Manager HR can access all departments
    }
    return u.Department == department
}