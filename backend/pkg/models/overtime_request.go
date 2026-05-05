package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==========================================
// Subdocument for each employee
type OvertimeEmployee struct {
	UserID         primitive.ObjectID `json:"user_id" bson:"user_id"`
	EmployeeStatus string             `json:"employee_status" bson:"employee_status"` // pending|agreed|rejected
	RejectionNote  string             `json:"rejection_note,omitempty" bson:"rejection_note,omitempty"` // opsional, wajib jika rejected
	ConfirmedAt    *time.Time         `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
}

// ==========================================
type OvertimeRequest struct {
	ID             primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	DepartmentID   primitive.ObjectID   `json:"department_id" bson:"department_id"`          // pengaju
	RequestedByID  primitive.ObjectID   `json:"requested_by_id" bson:"requested_by_id"`      // kepala departemen yang submit
	Date           time.Time            `json:"date" bson:"date"`                            // hanya tanggal, jam di bawah
	StartTime      string               `json:"start_time" bson:"start_time"`                // format "HH:mm"
	EndTime        string               `json:"end_time" bson:"end_time"`                    // format "HH:mm"
	Reason         string               `json:"reason" bson:"reason"`
	Status         string               `json:"status" bson:"status"`                        // draft|submitted|published
	Notes          string               `json:"notes,omitempty" bson:"notes,omitempty"`      // catatan HR opsional
	LetterURL      string               `json:"letter_url,omitempty" bson:"letter_url,omitempty"` // file/link SPKL opsional
	CreatedAt      time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" bson:"updated_at"`

	Employees      []OvertimeEmployee   `json:"employees" bson:"employees"`                 // array karyawan lembur
}

// ─── Constants ────────────────────────────────────────────────────────────

const (
	StatusDraft     = "draft"
	StatusSubmitted = "submitted"
	StatusPublished = "published"

	EmployeeStatusPending  = "pending"
	EmployeeStatusAgreed   = "agreed"
	EmployeeStatusRejected = "rejected"
)

// ─── Request Models ───────────────────────────────────────────────────────

type OvertimeEmployeeInput struct {
	UserID string `json:"user_id" binding:"required"`
}

type CreateOvertimeRequestRequest struct {
	DepartmentID string                  `json:"department_id" binding:"required"`
	Date         string                  `json:"date" binding:"required"` // YYYY-MM-DD
	StartTime    string                  `json:"start_time" binding:"required"`
	EndTime      string                  `json:"end_time" binding:"required"`
	Reason       string                  `json:"reason" binding:"required"`
	Employees    []OvertimeEmployeeInput `json:"employees" binding:"required"`
}

type UpdateOvertimeRequestRequest struct {
	Date      *string                  `json:"date,omitempty"`
	StartTime *string                  `json:"start_time,omitempty"`
	EndTime   *string                  `json:"end_time,omitempty"`
	Reason    *string                  `json:"reason,omitempty"`
	Status    *string                  `json:"status,omitempty"`
	Employees *[]OvertimeEmployeeInput `json:"employees,omitempty"`
}

type UpdateEmployeeStatusRequest struct {
	Status        string `json:"status" binding:"required"` // agreed|rejected
	RejectionNote string `json:"rejection_note,omitempty"`
}

// ─── Response Models ──────────────────────────────────────────────────────

type OvertimeEmployeeResponse struct {
	UserID         string     `json:"user_id"`
	FullName       string     `json:"full_name"`
	PayrollNumber  string     `json:"payroll_number"`
	EmployeeStatus string     `json:"employee_status"`
	RejectionNote  string     `json:"rejection_note,omitempty"`
	ConfirmedAt    *time.Time `json:"confirmed_at,omitempty"`
}

type OvertimeRequestResponse struct {
	ID            string                     `json:"id"`
	DepartmentID  string                     `json:"department_id"`
	DepartmentName string                    `json:"department_name"`
	RequestedByID string                     `json:"requested_by_id"`
	Date          time.Time                  `json:"date"`
	StartTime     string                     `json:"start_time"`
	EndTime       string                     `json:"end_time"`
	Reason        string                     `json:"reason"`
	Status        string                     `json:"status"`
	Notes         string                     `json:"notes,omitempty"`
	LetterURL     string                     `json:"letter_url,omitempty"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
	Employees     []OvertimeEmployeeResponse `json:"employees"`
}

// Legacy support or internal helper
type OvertimeApprovalResponse struct {
	Overtime OvertimeRequestResponse `json:"overtime"`
}