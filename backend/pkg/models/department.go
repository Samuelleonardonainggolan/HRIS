// pkg/models/department.go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Department struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Code        string             `json:"code" bson:"code"`
    Name        string             `json:"name" bson:"name"`
    Description string             `json:"description" bson:"description"`
    ManagerID   primitive.ObjectID `json:"manager_id,omitempty" bson:"manager_id,omitempty"`
    IsActive    bool               `json:"is_active" bson:"is_active"`
    CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type Position struct {
    ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    DepartmentID   primitive.ObjectID `json:"department_id" bson:"department_id"`
    Code           string             `json:"code" bson:"code"`
    Name           string             `json:"name" bson:"name"`
    Level          int                `json:"level" bson:"level"`
    Grade          string             `json:"grade" bson:"grade"`
    Description    string             `json:"description" bson:"description"`
    Responsibilities []string         `json:"responsibilities,omitempty" bson:"responsibilities,omitempty"`
    Requirements   []string           `json:"requirements,omitempty" bson:"requirements,omitempty"`
    SalaryRange    SalaryRange        `json:"salary_range" bson:"salary_range"`
    IsActive       bool               `json:"is_active" bson:"is_active"`
    CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

type SalaryRange struct {
    Min      int    `json:"min" bson:"min"`
    Max      int    `json:"max" bson:"max"`
    Currency string `json:"currency" bson:"currency"`
}

// Response types with populated data
type DepartmentWithManager struct {
    Department
    Manager *User `json:"manager,omitempty"`
}

type PositionWithDepartment struct {
    Position
    Department *Department `json:"department,omitempty"`
}