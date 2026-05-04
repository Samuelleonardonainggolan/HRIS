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