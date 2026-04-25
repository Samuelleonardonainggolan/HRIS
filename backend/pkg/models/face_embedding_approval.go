// pkg/models/face_embedding_approval.go
package models

import "time"

type FaceEmbeddingApprovalItem struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	FaceImageURL string    `json:"face_image_url,omitempty"`

	DepartmentName string `json:"department_name,omitempty"`
	PositionName   string `json:"position_name,omitempty"`
	FullName       string `json:"full_name,omitempty"`
	PayrollNumber  string `json:"payroll_number,omitempty"`
	Email          string `json:"email,omitempty"`

	RegisteredAt time.Time `json:"registered_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	IsActive     bool `json:"is_active"`
	IsFirstLogin bool `json:"is_first_login"`
}