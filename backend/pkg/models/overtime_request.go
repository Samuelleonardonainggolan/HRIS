package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ==========================================
// Subdocument for reward info
type OvertimeReward struct {
	RewardType    string     `json:"reward_type" bson:"reward_type"`                     // money|time_off
	RewardOption  string     `json:"reward_option,omitempty" bson:"reward_option,omitempty"` // early_out|late_in
	RewardDate    *time.Time `json:"reward_date,omitempty" bson:"reward_date,omitempty"`   // Tanggal klaim reward (terutama untuk time_off)
	RewardNominal float64    `json:"reward_nominal,omitempty" bson:"reward_nominal,omitempty"` // Nominal uang lembur
	Status        string     `json:"status" bson:"status"`                               // none|pending|granted|used
	GrantedAt     *time.Time `json:"granted_at,omitempty" bson:"granted_at,omitempty"`
	UsedAt        *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
}

// ==========================================
// Subdocument for each employee
type OvertimeEmployee struct {
	UserID         primitive.ObjectID `json:"user_id" bson:"user_id"`
	EmployeeStatus string             `json:"employee_status" bson:"employee_status"`                   // pending|agreed|rejected
	RejectionNote  string             `json:"rejection_note,omitempty" bson:"rejection_note,omitempty"` // opsional, wajib jika rejected
	LetterURL      string             `json:"letter_url,omitempty" bson:"letter_url,omitempty"`         // link SPKL per karyawan
	ConfirmedAt    *time.Time         `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
	Reward         OvertimeReward     `json:"reward" bson:"reward"`
}

// ==========================================
type OvertimeRequest struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DepartmentID  primitive.ObjectID `json:"department_id" bson:"department_id"`     // pengaju
	RequestedByID primitive.ObjectID `json:"requested_by_id" bson:"requested_by_id"` // kepala departemen yang submit
	Date          time.Time          `json:"date" bson:"date"`                       // hanya tanggal, jam di bawah
	StartTime     string             `json:"start_time" bson:"start_time"`           // format "HH:mm"
	EndTime       string             `json:"end_time" bson:"end_time"`               // format "HH:mm"
	Reason        string             `json:"reason" bson:"reason"`
	Status        string             `json:"status" bson:"status"`                             // draft|submitted|published
	Notes         string             `json:"notes,omitempty" bson:"notes,omitempty"`           // catatan HR opsional
	LetterURL     string             `json:"letter_url,omitempty" bson:"letter_url,omitempty"` // file/link SPKL opsional
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`

	Employees []OvertimeEmployee `json:"employees" bson:"employees"` // array karyawan lembur
}

func (r *OvertimeRequest) GetDurationHours() float64 {
	start, err1 := time.Parse("15:04", r.StartTime)
	end, err2 := time.Parse("15:04", r.EndTime)
	if err1 != nil || err2 != nil {
		return 0
	}
	duration := end.Sub(start)
	if duration < 0 {
		duration += 24 * time.Hour
	}
	return duration.Hours()
}

// ─── Constants ────────────────────────────────────────────────────────────

const (
	StatusDraft     = "draft"
	StatusSubmitted = "submitted"
	StatusPublished = "published"

	EmployeeStatusPending  = "pending"
	EmployeeStatusAgreed   = "agreed"
	EmployeeStatusRejected = "rejected"

	OvertimeRewardTypeMoney   = "money"
	OvertimeRewardTypeTimeOff = "time_off"

	OvertimeRewardStatusNone    = "none"
	OvertimeRewardStatusPending = "pending"
	OvertimeRewardStatusGranted = "granted"
	OvertimeRewardStatusUsed    = "used"
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
	Status       string                  `json:"status,omitempty"`
	Employees    []OvertimeEmployeeInput `json:"employees" binding:"required"`
}

type UpdateOvertimeRequestRequest struct {
	Date      *string                  `json:"date,omitempty"`
	StartTime *string                  `json:"start_time,omitempty"`
	EndTime   *string                  `json:"end_time,omitempty"`
	Reason    *string                  `json:"reason,omitempty"`
	Status    *string                  `json:"status,omitempty"`
	Employees *[]OvertimeEmployeeInput `json:"employees,omitempty"`
	Notes     *string                  `json:"notes,omitempty"`
	LetterURL *string                  `json:"letter_url,omitempty"`
}

type UpdateEmployeeStatusRequest struct {
	Status        string `json:"status" binding:"required"` // agreed|rejected
	RejectionNote string `json:"rejection_note,omitempty"`
}

type ClaimOvertimeRewardRequest struct {
	RewardType   string `json:"reward_type" binding:"required"` // money|time_off
	RewardOption string `json:"reward_option,omitempty"`        // early_out|late_in
	RewardDate   string `json:"reward_date,omitempty"`          // YYYY-MM-DD
}

// ─── Response Models ──────────────────────────────────────────────────────

type OvertimeEmployeeResponse struct {
	UserID         string         `json:"user_id"`
	FullName       string         `json:"full_name"`
	PayrollNumber  string         `json:"payroll_number"`
	PositionName   string         `json:"position_name"`
	EmployeeStatus string         `json:"employee_status"`
	RejectionNote  string         `json:"rejection_note,omitempty"`
	LetterURL      string         `json:"letter_url,omitempty"`
	ConfirmedAt    *time.Time     `json:"confirmed_at,omitempty"`
	Reward         OvertimeReward `json:"reward"`
}

type OvertimeRequestResponse struct {
	ID              string                     `json:"id"`
	DepartmentID    string                     `json:"department_id"`
	DepartmentName  string                     `json:"department_name"`
	RequestedByID   string                     `json:"requested_by_id"`
	RequestedByName string                     `json:"requested_by_name"`
	Date            time.Time                  `json:"date"`
	StartTime       string                     `json:"start_time"`
	EndTime         string                     `json:"end_time"`
	Reason          string                     `json:"reason"`
	Status          string                     `json:"status"`
	Notes           string                     `json:"notes,omitempty"`
	LetterURL       string                     `json:"letter_url,omitempty"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
	Employees       []OvertimeEmployeeResponse `json:"employees"`
}

// Legacy support or internal helper
type OvertimeApprovalResponse struct {
	Overtime OvertimeRequestResponse `json:"overtime"`
}
