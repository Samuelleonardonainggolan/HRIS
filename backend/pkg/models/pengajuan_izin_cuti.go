// pkg/models/pengajuan_izin_cuti.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeaveRequest struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID        primitive.ObjectID `json:"user_id" bson:"user_id"`
	RequestTypeID primitive.ObjectID `json:"request_type_id" bson:"request_type_id"`
	TypeName      string             `json:"type_name" bson:"type_name"`

	StartDate time.Time `json:"start_date" bson:"start_date"`
	EndDate   time.Time `json:"end_date" bson:"end_date"`
	DaysTotal int       `json:"days_total" bson:"days_total"`
	Reason    string    `json:"reason" bson:"reason"`

	DocumentURL    string              `json:"document_url,omitempty" bson:"document_url,omitempty"`
	LeaveBalanceID *primitive.ObjectID `json:"leave_balance_id,omitempty" bson:"leave_balance_id,omitempty"`

	StatusKepalaDepartemen string             `json:"status_kepala_departemen" bson:"status_kepala_departemen"`
	KepalaDepartemenID     primitive.ObjectID `json:"kepala_departemen_id" bson:"kepala_departemen_id"`

	ManagerHRID     primitive.ObjectID `json:"manager_hr_id" bson:"manager_hr_id"`
	StatusManagerHR string             `json:"status_manager_hr" bson:"status_manager_hr"`
	FinalStatus     string             `json:"final_status" bson:"final_status"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type CreatePengajuanIzinCutiRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	RequestTypeID string `json:"request_type_id" binding:"required"`
	StartDate     string `json:"start_date" binding:"required"`
	EndDate       string `json:"end_date" binding:"required"`
	DaysTotal     int    `json:"days_total" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
	DocumentURL   string `json:"document_url,omitempty"`
}

type UpdatePengajuanIzinCutiRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	DaysTotal *int       `json:"days_total,omitempty"`
	Reason    string     `json:"reason,omitempty"`
	DocumentURL string   `json:"document_url,omitempty"`

	StatusKepalaDepartemen string `json:"status_kepala_departemen,omitempty"`
	StatusManagerHR        string `json:"status_manager_hr,omitempty"`
	FinalStatus            string `json:"final_status,omitempty"`
}

type PengajuanIzinCutiResponse struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	RequestTypeID string `json:"request_type_id"`
	TypeName      string `json:"type_name"`

	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	DaysTotal int       `json:"days_total"`
	Reason    string    `json:"reason"`

	DocumentURL    string `json:"document_url,omitempty"`
	LeaveBalanceID string `json:"leave_balance_id,omitempty"`

	StatusKepalaDepartemen string `json:"status_kepala_departemen"`
	KepalaDepartemenID     string `json:"kepala_departemen_id"`

	ManagerHRID     string `json:"manager_hr_id"`
	StatusManagerHR string `json:"status_manager_hr"`
	FinalStatus     string `json:"final_status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PengajuanIzinCutiApprovalEmployeeResponse struct {
	ID             string `json:"id"`
	PayrollNumber  string `json:"payroll_number"`
	FullName       string `json:"full_name"`
	DepartmentName string `json:"department_name"`
	PositionName   string `json:"position_name"`
}

type PengajuanIzinCutiApprovalResponse struct {
	Pengajuan PengajuanIzinCutiResponse                  `json:"pengajuan"`
	Employee  *PengajuanIzinCutiApprovalEmployeeResponse `json:"employee,omitempty"`
}

func (p *LeaveRequest) ToResponse() PengajuanIzinCutiResponse {
	var bal string
	if p.LeaveBalanceID != nil {
		bal = p.LeaveBalanceID.Hex()
	}

	return PengajuanIzinCutiResponse{
		ID:                     p.ID.Hex(),
		UserID:                 p.UserID.Hex(),
		RequestTypeID:          p.RequestTypeID.Hex(),
		TypeName:               p.TypeName,
		StartDate:              p.StartDate,
		EndDate:                p.EndDate,
		DaysTotal:              p.DaysTotal,
		Reason:                 p.Reason,
		DocumentURL:            p.DocumentURL,
		LeaveBalanceID:         bal,
		StatusKepalaDepartemen: p.StatusKepalaDepartemen,
		KepalaDepartemenID:     p.KepalaDepartemenID.Hex(),
		ManagerHRID:            p.ManagerHRID.Hex(),
		StatusManagerHR:        p.StatusManagerHR,
		FinalStatus:            p.FinalStatus,
		CreatedAt:              p.CreatedAt,
		UpdatedAt:              p.UpdatedAt,
	}
}

// Status constants (tetap)
const (
	StatusPending  = "PENDING"
	StatusApproved = "APPROVED"
	StatusRejected = "REJECTED"
)