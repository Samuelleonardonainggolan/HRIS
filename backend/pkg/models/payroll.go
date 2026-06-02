// pkg/models/payroll.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payroll struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	Month int `json:"month" bson:"month"`
	Year  int `json:"year" bson:"year"`

	// base salary & final
	BasicSalary string `json:"basic_salary" bson:"basic_salary"`
	NetSalary   string `json:"net_salary" bson:"net_salary"`

	// numeric mirror (recommended for calc/audit)
	BasicSalaryValue int64 `json:"basic_salary_value" bson:"basic_salary_value"`
	NetSalaryValue   int64 `json:"net_salary_value" bson:"net_salary_value"`

	// attendance summary
	TotalDaysPresent string `json:"total_days_present" bson:"total_days_present"`
	LateMinutesTotal int    `json:"late_minutes_total" bson:"late_minutes_total"`
	AbsentDays       int    `json:"absent_days" bson:"absent_days"`

	// overtime summary
	OvertimeHoursApproved float64 `json:"overtime_hours_approved" bson:"overtime_hours_approved"`
	OvertimeHoursRedeemed float64 `json:"overtime_hours_redeemed" bson:"overtime_hours_redeemed"`
	OvertimeHoursPaid     float64 `json:"overtime_hours_paid" bson:"overtime_hours_paid"`

	// calculation constants (audit)
	MonthlyHoursDivisor int `json:"monthly_hours_divisor" bson:"monthly_hours_divisor"` // 173
	WorkdaysDivisor     int `json:"workdays_divisor" bson:"workdays_divisor"`           // 26
	MinutesPerWorkday   int `json:"minutes_per_workday" bson:"minutes_per_workday"`     // 480 (atau dari jadwal)

	// breakdown (string + numeric)
	OvertimePay      string `json:"overtime_pay" bson:"overtime_pay"`
	OvertimePayValue int64  `json:"overtime_pay_value" bson:"overtime_pay_value"`

	LateDeduction      string `json:"late_deduction" bson:"late_deduction"`
	LateDeductionValue int64  `json:"late_deduction_value" bson:"late_deduction_value"`

	AbsentDeduction      string `json:"absent_deduction" bson:"absent_deduction"`
	AbsentDeductionValue int64  `json:"absent_deduction_value" bson:"absent_deduction_value"`

	OtherEarnings      string `json:"other_earnings" bson:"other_earnings"`
	OtherEarningsValue int64  `json:"other_earnings_value" bson:"other_earnings_value"`

	OtherDeductions      string `json:"other_deductions" bson:"other_deductions"`
	OtherDeductionsValue int64  `json:"other_deductions_value" bson:"other_deductions_value"`

	Status string `json:"status" bson:"status"` // draft|pending|approved|paid

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type CreatePayrollRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Month  int    `json:"month" binding:"required"`
	Year   int    `json:"year" binding:"required"`

	BasicSalary string `json:"basic_salary" binding:"required"`
	NetSalary   string `json:"net_salary" binding:"required"`

	TotalDaysPresent string `json:"total_days_present" binding:"required"`

	LateMinutesTotal int `json:"late_minutes_total"`
	AbsentDays       int `json:"absent_days"`

	OvertimeHoursApproved float64 `json:"overtime_hours_approved"`
	OvertimeHoursRedeemed float64 `json:"overtime_hours_redeemed"`
	OvertimeHoursPaid     float64 `json:"overtime_hours_paid"`

	MonthlyHoursDivisor int `json:"monthly_hours_divisor"` // default 173
	WorkdaysDivisor     int `json:"workdays_divisor"`       // default 26
	MinutesPerWorkday   int `json:"minutes_per_workday"`    // default 480

	OvertimePay     string `json:"overtime_pay"`
	LateDeduction   string `json:"late_deduction"`
	AbsentDeduction string `json:"absent_deduction"`

	OtherEarnings   string `json:"other_earnings"`
	OtherDeductions string `json:"other_deductions"`

	Status string `json:"status,omitempty"`
}

type UpdatePayrollRequest struct {
	Month *int `json:"month,omitempty"`
	Year  *int `json:"year,omitempty"`

	BasicSalary string `json:"basic_salary,omitempty"`
	NetSalary   string `json:"net_salary,omitempty"`

	TotalDaysPresent string `json:"total_days_present,omitempty"`

	LateMinutesTotal *int `json:"late_minutes_total,omitempty"`
	AbsentDays       *int `json:"absent_days,omitempty"`

	OvertimeHoursApproved *float64 `json:"overtime_hours_approved,omitempty"`
	OvertimeHoursRedeemed *float64 `json:"overtime_hours_redeemed,omitempty"`
	OvertimeHoursPaid     *float64 `json:"overtime_hours_paid,omitempty"`

	MonthlyHoursDivisor *int `json:"monthly_hours_divisor,omitempty"`
	WorkdaysDivisor     *int `json:"workdays_divisor,omitempty"`
	MinutesPerWorkday   *int `json:"minutes_per_workday,omitempty"`

	OvertimePay     string `json:"overtime_pay,omitempty"`
	LateDeduction   string `json:"late_deduction,omitempty"`
	AbsentDeduction string `json:"absent_deduction,omitempty"`

	OtherEarnings   string `json:"other_earnings,omitempty"`
	OtherDeductions string `json:"other_deductions,omitempty"`

	Status *string `json:"status,omitempty"`
}

type PayrollResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`

	Month int `json:"month"`
	Year  int `json:"year"`

	BasicSalary string `json:"basic_salary"`
	NetSalary   string `json:"net_salary"`

	BasicSalaryValue int64 `json:"basic_salary_value"`
	NetSalaryValue   int64 `json:"net_salary_value"`

	TotalDaysPresent string `json:"total_days_present"`
	LateMinutesTotal int    `json:"late_minutes_total"`
	AbsentDays       int    `json:"absent_days"`

	OvertimeHoursApproved float64 `json:"overtime_hours_approved"`
	OvertimeHoursRedeemed float64 `json:"overtime_hours_redeemed"`
	OvertimeHoursPaid     float64 `json:"overtime_hours_paid"`

	MonthlyHoursDivisor int `json:"monthly_hours_divisor"`
	WorkdaysDivisor     int `json:"workdays_divisor"`
	MinutesPerWorkday   int `json:"minutes_per_workday"`

	OvertimePay     string `json:"overtime_pay"`
	LateDeduction   string `json:"late_deduction"`
	AbsentDeduction string `json:"absent_deduction"`

	OvertimePayValue     int64 `json:"overtime_pay_value"`
	LateDeductionValue   int64 `json:"late_deduction_value"`
	AbsentDeductionValue int64 `json:"absent_deduction_value"`

	OtherEarnings   string `json:"other_earnings"`
	OtherDeductions string `json:"other_deductions"`

	OtherEarningsValue   int64 `json:"other_earnings_value"`
	OtherDeductionsValue int64 `json:"other_deductions_value"`

	Status string `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Payroll) ToResponse() PayrollResponse {
	resp := PayrollResponse{
		ID:                  p.ID.Hex(),
		Month:               p.Month,
		Year:                p.Year,
		BasicSalary:         p.BasicSalary,
		NetSalary:           p.NetSalary,
		BasicSalaryValue:    p.BasicSalaryValue,
		NetSalaryValue:      p.NetSalaryValue,
		TotalDaysPresent:    p.TotalDaysPresent,
		LateMinutesTotal:    p.LateMinutesTotal,
		AbsentDays:          p.AbsentDays,
		OvertimeHoursApproved: p.OvertimeHoursApproved,
		OvertimeHoursRedeemed: p.OvertimeHoursRedeemed,
		OvertimeHoursPaid:     p.OvertimeHoursPaid,
		MonthlyHoursDivisor: p.MonthlyHoursDivisor,
		WorkdaysDivisor:     p.WorkdaysDivisor,
		MinutesPerWorkday:   p.MinutesPerWorkday,
		OvertimePay:         p.OvertimePay,
		LateDeduction:       p.LateDeduction,
		AbsentDeduction:     p.AbsentDeduction,
		OvertimePayValue:    p.OvertimePayValue,
		LateDeductionValue:  p.LateDeductionValue,
		AbsentDeductionValue: p.AbsentDeductionValue,
		OtherEarnings:       p.OtherEarnings,
		OtherDeductions:     p.OtherDeductions,
		OtherEarningsValue:  p.OtherEarningsValue,
		OtherDeductionsValue: p.OtherDeductionsValue,
		Status:              p.Status,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}

	if !p.UserID.IsZero() {
		resp.UserID = p.UserID.Hex()
	}
	return resp
}

// CalculateOvertimePay calculates overtime bonus based on 1.5x first hour and 2x subsequent hours.
// Divisor is usually 173.
func (p *Payroll) CalculateOvertimePay(hours float64) int64 {
	if hours <= 0 {
		return 0
	}

	divisor := int64(p.MonthlyHoursDivisor)
	if divisor == 0 {
		divisor = 173 // Default divisor
	}

	basis := float64(p.BasicSalaryValue) / float64(divisor)
	
	var totalFactor float64
	if hours <= 1 {
		totalFactor = hours * 1.5
	} else {
		// First hour is 1.5, the rest is 2.0
		totalFactor = 1.5 + (hours-1)*2.0
	}

	return int64(basis * totalFactor)
}

// CalculateLateDeduction calculates deduction for late clock-in.
// Divisor is usually 26 workdays, then divided by minutes per workday (e.g. 480).
func (p *Payroll) CalculateLateDeduction(lateMinutes int) int64 {
	if lateMinutes <= 0 {
		return 0
	}

	workdaysDivisor := int64(p.WorkdaysDivisor)
	if workdaysDivisor == 0 {
		workdaysDivisor = 26
	}

	minutesDivisor := int64(p.MinutesPerWorkday)
	if minutesDivisor == 0 {
		minutesDivisor = 480 // 8 hours * 60 minutes
	}

	// Rate per minute = (Basic Salary / 26) / 480
	minuteRate := float64(p.BasicSalaryValue) / float64(workdaysDivisor) / float64(minutesDivisor)
	
	return int64(minuteRate * float64(lateMinutes))
}

// CalculateAbsentDeduction calculates deduction for absence (mangkir).
// Formula: (Basic Salary / 26) * absentDays
func (p *Payroll) CalculateAbsentDeduction(absentDays int) int64 {
	if absentDays <= 0 {
		return 0
	}

	workdaysDivisor := int64(p.WorkdaysDivisor)
	if workdaysDivisor == 0 {
		workdaysDivisor = 26
	}

	dailyRate := float64(p.BasicSalaryValue) / float64(workdaysDivisor)
	
	return int64(dailyRate * float64(absentDays))
}

// RecalculateNetSalary updates the NetSalaryValue based on basic salary, overtime, and deductions.
func (p *Payroll) RecalculateNetSalary() {
	p.NetSalaryValue = p.BasicSalaryValue +
		p.OvertimePayValue +
		p.OtherEarningsValue -
		p.LateDeductionValue -
		p.AbsentDeductionValue -
		p.OtherDeductionsValue
}

// PayrollRecord adalah ringkasan payroll untuk daftar/tabel (digunakan oleh accountant).
type PayrollRecord struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	Initials   string `json:"initials"`
	Position   string `json:"position"`
	Department string `json:"department"`

	BasicSalary int64 `json:"basicSalary"`
	Bonus10     int64 `json:"bonus10"`    // legacy, bisa 0
	Overtime    int64 `json:"overtime"`
	Deduction   int64 `json:"deduction"`
	NetTotal    int64 `json:"netTotal"`

	Status string `json:"status"`
	Month  int    `json:"month"`
	Year   int    `json:"year"`
}

// PayrollDetailResponse adalah detail payroll lengkap termasuk info karyawan.
type PayrollDetailResponse struct {
	PayrollResponse

	// Info karyawan (di-join dari collection users)
	EmployeeName     string `json:"employee_name"`
	EmployeePosition string `json:"employee_position"`
	EmployeeDept     string `json:"employee_department"`
	PayrollNumber    string `json:"payroll_number"`
}