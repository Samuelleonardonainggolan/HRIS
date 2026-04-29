// pkg/models/overtime_request.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OvertimeRequest represents overtime request submission with multi-level approval
type OvertimeRequest struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	// Stored as MongoDB DateTime
	Date      time.Time `json:"date" bson:"date"`
	StartTime time.Time `json:"start_time" bson:"start_time"`
	EndTime   time.Time `json:"end_time" bson:"end_time"`

	Reason string `json:"reason" bson:"reason"`
	Total  string `json:"total" bson:"total"`

	StatusKepalaDepartemen string             `json:"status_kepala_departemen" bson:"status_kepala_departemen"`
	KepalaDepartemenID     primitive.ObjectID `json:"kepala_departemen_id,omitempty" bson:"kepala_departemen_id,omitempty"`

	StatusManagerHR string             `json:"status_manager_hr" bson:"status_manager_hr"`
	ManagerHRID     primitive.ObjectID `json:"manager_hr_id,omitempty" bson:"manager_hr_id,omitempty"`

	FinalStatus string `json:"final_status" bson:"final_status"`

	RejectionReasonKepalaDept string `json:"rejection_reason_kepala_dept,omitempty" bson:"rejection_reason_kepala_dept,omitempty"`
	RejectionReasonManagerHR  string `json:"rejection_reason_manager_hr,omitempty" bson:"rejection_reason_manager_hr,omitempty"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// CreateOvertimeRequestRequest represents request to create overtime request
type CreateOvertimeRequestRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	Date      string `json:"date" binding:"required"`       // YYYY-MM-DD recommended
	StartTime string `json:"start_time" binding:"required"` // ISO or HH:MM (depends on your UI parsing)
	EndTime   string `json:"end_time" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
	Total     string `json:"total,omitempty"` // optional (can be computed)
}

// UpdateOvertimeRequestRequest represents request to update overtime request
type UpdateOvertimeRequestRequest struct {
	Date      *string `json:"date,omitempty"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
	Reason    *string `json:"reason,omitempty"`
	Total     *string `json:"total,omitempty"`

	StatusKepalaDepartemen *string `json:"status_kepala_departemen,omitempty"`
	KepalaDepartemenID     *string `json:"kepala_departemen_id,omitempty"`

	StatusManagerHR *string `json:"status_manager_hr,omitempty"`
	ManagerHRID     *string `json:"manager_hr_id,omitempty"`

	FinalStatus *string `json:"final_status,omitempty"`
}

// RejectOvertimeRequest represents rejection request
type RejectOvertimeRequest struct {
	RejectionReason string `json:"rejection_reason" binding:"required"`
}

// OvertimeRequestResponse represents response
type OvertimeRequestResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`

	Date      time.Time `json:"date"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	Reason string `json:"reason"`
	Total  string `json:"total"`

	StatusKepalaDepartemen string `json:"status_kepala_departemen"`
	KepalaDepartemenID     string `json:"kepala_departemen_id,omitempty"`

	StatusManagerHR string `json:"status_manager_hr"`
	ManagerHRID     string `json:"manager_hr_id,omitempty"`

	FinalStatus string `json:"final_status"`

	RejectionReasonKepalaDept string `json:"rejection_reason_kepala_dept,omitempty"`
	RejectionReasonManagerHR  string `json:"rejection_reason_manager_hr,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OvertimeApprovalEmployeeResponse represents employee info in approval response
type OvertimeApprovalEmployeeResponse struct {
	ID             string `json:"id"`
	PayrollNumber  string `json:"payroll_number"`
	FullName       string `json:"full_name"`
	DepartmentName string `json:"department_name"`
	PositionName   string `json:"position_name"`
}

// OvertimeApprovalResponse represents full approval response
type OvertimeApprovalResponse struct {
	Overtime OvertimeRequestResponse           `json:"overtime"`
	Employee *OvertimeApprovalEmployeeResponse `json:"employee,omitempty"`
}

// ToResponse converts OvertimeRequest to OvertimeRequestResponse
func (o *OvertimeRequest) ToResponse() OvertimeRequestResponse {
	resp := OvertimeRequestResponse{
		ID:     o.ID.Hex(),
		UserID: o.UserID.Hex(),

		Date:      o.Date,
		StartTime: o.StartTime,
		EndTime:   o.EndTime,

		Reason: o.Reason,
		Total:  o.Total,

		StatusKepalaDepartemen: o.StatusKepalaDepartemen,
		StatusManagerHR:        o.StatusManagerHR,
		FinalStatus:            o.FinalStatus,

		RejectionReasonKepalaDept: o.RejectionReasonKepalaDept,
		RejectionReasonManagerHR:  o.RejectionReasonManagerHR,

		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}

	if !o.KepalaDepartemenID.IsZero() {
		resp.KepalaDepartemenID = o.KepalaDepartemenID.Hex()
	}
	if !o.ManagerHRID.IsZero() {
		resp.ManagerHRID = o.ManagerHRID.Hex()
	}

	return resp
}
