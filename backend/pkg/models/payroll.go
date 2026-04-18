// pkg/models/payroll.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payroll struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID        primitive.ObjectID `json:"user_id" bson:"user_id"`
	Month         int                `json:"month" bson:"month"`
	Year          int                `json:"year" bson:"year"`
	BasicSalary   string             `json:"basic_salary" bson:"basic_salary"`

	TotalDaysPresent string `json:"total_days_present" bson:"total_days_present"` // ✅ renamed
	TotalLeave       string `json:"total_leave" bson:"total_leave"`               // ✅ renamed
	TotalPermission  string `json:"total_permission" bson:"total_permission"`     // ✅ renamed

	NetSalary     string `json:"net_salary" bson:"net_salary"`
	PaymentStatus string `json:"payment_status" bson:"payment_status"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type CreatePayrollRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	Month   int    `json:"month" binding:"required"`
	Year    int    `json:"year" binding:"required"`

	BasicSalary string `json:"basic_salary" binding:"required"`

	TotalDaysPresent string `json:"total_days_present" binding:"required"` // ✅ renamed
	TotalLeave       string `json:"total_leave" binding:"required"`        // ✅ renamed
	TotalPermission  string `json:"total_permission" binding:"required"`   // ✅ renamed

	NetSalary     string `json:"net_salary" binding:"required"`
	PaymentStatus string `json:"payment_status" binding:"required"`
}

type UpdatePayrollRequest struct {
	Month *int `json:"month,omitempty"`
	Year  *int `json:"year,omitempty"`

	BasicSalary string `json:"basic_salary,omitempty"`

	TotalDaysPresent string `json:"total_days_present,omitempty"` // ✅ renamed
	TotalLeave       string `json:"total_leave,omitempty"`        // ✅ renamed
	TotalPermission  string `json:"total_permission,omitempty"`   // ✅ renamed

	NetSalary     string `json:"net_salary,omitempty"`
	PaymentStatus string `json:"payment_status,omitempty"`
}

type PayrollResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`

	Month int `json:"month"`
	Year  int `json:"year"`

	BasicSalary string `json:"basic_salary"`

	TotalDaysPresent string `json:"total_days_present"` // ✅ renamed
	TotalLeave       string `json:"total_leave"`        // ✅ renamed
	TotalPermission  string `json:"total_permission"`   // ✅ renamed

	NetSalary     string    `json:"net_salary"`
	PaymentStatus string    `json:"payment_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (p *Payroll) ToResponse() PayrollResponse {
	resp := PayrollResponse{
		ID:               p.ID.Hex(),
		Month:            p.Month,
		Year:             p.Year,
		BasicSalary:      p.BasicSalary,
		TotalDaysPresent: p.TotalDaysPresent,
		TotalLeave:       p.TotalLeave,
		TotalPermission:  p.TotalPermission,
		NetSalary:        p.NetSalary,
		PaymentStatus:    p.PaymentStatus,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}

	if !p.UserID.IsZero() {
		resp.UserID = p.UserID.Hex()
	}

	return resp
}