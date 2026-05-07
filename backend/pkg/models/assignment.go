package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// =============================================================================
// Subdocuments
// =============================================================================

type AssignmentShift struct {
	StartTime string `json:"start_time" bson:"start_time"` // "HH:mm"
	EndTime   string `json:"end_time" bson:"end_time"`     // "HH:mm"
}

type AssignmentOriginalShift struct {
	Type      string `json:"type" bson:"type"`                     // "shift" | "off"
	StartTime string `json:"start_time,omitempty" bson:"start_time,omitempty"` // "HH:mm" if type="shift"
	EndTime   string `json:"end_time,omitempty" bson:"end_time,omitempty"`     // "HH:mm" if type="shift"
	Source    string `json:"source,omitempty" bson:"source,omitempty"`         // optional: "working_hours"|"manual"|"schedule"
}

type AssignmentDayOffReward struct {
	Eligible           bool       `json:"eligible" bson:"eligible"`
	Status             string     `json:"status" bson:"status"` // none|pending|granted|used|cancelled
	Description        string     `json:"description,omitempty" bson:"description,omitempty"`
	GrantedAt          *time.Time `json:"granted_at,omitempty" bson:"granted_at,omitempty"`
	UsedAt             *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
	ReplacementOffDate *time.Time `json:"replacement_off_date,omitempty" bson:"replacement_off_date,omitempty"`
}

type AssignmentEmployee struct {
	UserID         primitive.ObjectID      `json:"user_id" bson:"user_id"`
	OriginalShift  AssignmentOriginalShift  `json:"original_shift" bson:"original_shift"`
	AssignedShift  AssignmentShift          `json:"assigned_shift" bson:"assigned_shift"`
	EmployeeStatus string                  `json:"employee_status" bson:"employee_status"` // pending|agreed|rejected|cancelled
	RejectionNote  string                  `json:"rejection_note,omitempty" bson:"rejection_note,omitempty"`
	ConfirmedAt    *time.Time              `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
	DayOffReward   AssignmentDayOffReward  `json:"day_off_reward" bson:"day_off_reward"`
}

// =============================================================================
// Main document
// =============================================================================

type Assignment struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DepartmentID  primitive.ObjectID `json:"department_id" bson:"department_id"`
	RequestedByID primitive.ObjectID `json:"requested_by_id" bson:"requested_by_id"`

	Date   time.Time `json:"date" bson:"date"`     // only date
	Reason string    `json:"reason" bson:"reason"` // reason/event/occupancy

	Status string `json:"status" bson:"status"` // draft|submitted|published|cancelled
	Notes  string `json:"notes,omitempty" bson:"notes,omitempty"`

	ShiftTarget AssignmentShift `json:"shift_target" bson:"shift_target"`

	Employees []AssignmentEmployee `json:"employees" bson:"employees"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// =============================================================================
// Constants
// =============================================================================

const (
	AssignmentStatusDraft     = "draft"
	AssignmentStatusSubmitted = "submitted"
	AssignmentStatusPublished = "published"
	AssignmentStatusCancelled = "cancelled"
)

const (
	ShiftTypeShift = "shift"
	ShiftTypeOff   = "off"
)

const (
	AssignmentEmployeeStatusPending   = "pending"
	AssignmentEmployeeStatusAgreed    = "agreed"
	AssignmentEmployeeStatusRejected  = "rejected"
	AssignmentEmployeeStatusCancelled = "cancelled"
)

const (
	DayOffRewardStatusNone      = "none"
	DayOffRewardStatusPending   = "pending"
	DayOffRewardStatusGranted   = "granted"
	DayOffRewardStatusUsed      = "used"
	DayOffRewardStatusCancelled = "cancelled"
)

// =============================================================================
// Request Models
// =============================================================================

type AssignmentEmployeeInput struct {
	UserID string `json:"user_id" binding:"required"`

	// optional: jika Anda ingin override shift per karyawan
	AssignedStartTime *string `json:"assigned_start_time,omitempty"` // "HH:mm"
	AssignedEndTime   *string `json:"assigned_end_time,omitempty"`   // "HH:mm"
}

type CreateAssignmentRequest struct {
	DepartmentID  string                  `json:"department_id" binding:"required"`
	Date          string                  `json:"date" binding:"required"` // YYYY-MM-DD
	ShiftStart    string                  `json:"start_time" binding:"required"`
	ShiftEnd      string                  `json:"end_time" binding:"required"`
	Reason        string                  `json:"reason" binding:"required"`
	Status        string                  `json:"status,omitempty"` // draft/submitted
	Employees     []AssignmentEmployeeInput `json:"employees" binding:"required,min=1"`
	Notes         string                  `json:"notes,omitempty"`
}

type UpdateAssignmentRequest struct {
	Date       *string `json:"date,omitempty"` // YYYY-MM-DD
	ShiftStart *string `json:"start_time,omitempty"`
	ShiftEnd   *string `json:"end_time,omitempty"`
	Reason     *string `json:"reason,omitempty"`
	Status     *string `json:"status,omitempty"`
	Notes      *string `json:"notes,omitempty"`

	Employees *[]AssignmentEmployeeInput `json:"employees,omitempty"`
}

// karyawan setuju / tolak
type UpdateAssignmentEmployeeStatusRequest struct {
	Status        string `json:"status" binding:"required"` // agreed|rejected
	RejectionNote string `json:"rejection_note,omitempty"`
}

// optional: ketika karyawan/HR menetapkan tanggal off pengganti (jika dipakai)
type UseReplacementDayOffRequest struct {
	ReplacementOffDate string `json:"replacement_off_date" binding:"required"` // YYYY-MM-DD
}

// =============================================================================
// Response Models
// =============================================================================

type AssignmentEmployeeResponse struct {
	UserID         string     `json:"user_id"`
	FullName       string     `json:"full_name"`
	PayrollNumber  string     `json:"payroll_number"`
	PositionName   string     `json:"position_name"`

	OriginalShiftType string `json:"original_shift_type"`
	OriginalStartTime string `json:"original_start_time,omitempty"`
	OriginalEndTime   string `json:"original_end_time,omitempty"`

	AssignedStartTime string `json:"assigned_start_time"`
	AssignedEndTime   string `json:"assigned_end_time"`

	EmployeeStatus string     `json:"employee_status"`
	RejectionNote  string     `json:"rejection_note,omitempty"`
	ConfirmedAt    *time.Time `json:"confirmed_at,omitempty"`

	DayOffEligible           bool       `json:"day_off_eligible"`
	DayOffStatus             string     `json:"day_off_status"`
	DayOffGrantedAt          *time.Time `json:"day_off_granted_at,omitempty"`
	DayOffUsedAt             *time.Time `json:"day_off_used_at,omitempty"`
	ReplacementOffDate       *time.Time `json:"replacement_off_date,omitempty"`
}

type AssignmentResponse struct {
	ID             string    `json:"id"`
	DepartmentID   string    `json:"department_id"`
	DepartmentName string    `json:"department_name"`

	RequestedByID   string `json:"requested_by_id"`
	RequestedByName string `json:"requested_by_name"`

	Date   time.Time `json:"date"`
	Reason string    `json:"reason"`

	Status string `json:"status"`
	Notes  string `json:"notes,omitempty"`

	ShiftStart string `json:"start_time"`
	ShiftEnd   string `json:"end_time"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Employees []AssignmentEmployeeResponse `json:"employees"`
}