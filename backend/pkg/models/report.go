package models

import (
	"time"
)

type AttendanceActivityReportType string

const (
	ReportTypeLate            AttendanceActivityReportType = "late"
	ReportTypeMissingClockOut AttendanceActivityReportType = "missing_clock_out"
	ReportTypeLeave           AttendanceActivityReportType = "leave"
	ReportTypePermission      AttendanceActivityReportType = "permission"
	ReportTypeOvertime        AttendanceActivityReportType = "overtime"
)

type AttendanceActivityReportRow struct {
	ID             string                       `json:"id"`
	Date           time.Time                    `json:"date"`
	DayLabel       string                       `json:"day_label"`
	EmployeeName   string                       `json:"employee_name"`
	EmployeeNIK    string                       `json:"employee_nik"`
	EmployeeInitials string                     `json:"employee_initials"`
	DepartmentName string                       `json:"department_name"`
	PositionName   string                       `json:"position_name"`
	Type           AttendanceActivityReportType `json:"type"`
	
	// Time fields
	ScheduledIn    string `json:"scheduled_in,omitempty"`
	ActualIn       string `json:"actual_in,omitempty"`
	ScheduledOut   string `json:"scheduled_out,omitempty"`
	ActualOut      string `json:"actual_out,omitempty"`
	
	// Overtime fields
	OvertimeStart  string  `json:"overtime_start,omitempty"`
	OvertimeEnd    string  `json:"overtime_end,omitempty"`
	OvertimeHours  float64 `json:"overtime_hours,omitempty"`
	
	// Leave/Permission fields
	DateRange      string `json:"date_range,omitempty"`
	
	// Computed details
	LateMinutes    int    `json:"late_minutes,omitempty"`
	
	ApprovalStatus string `json:"approval_status,omitempty"`
	Impact         string `json:"impact"`
	Note           string `json:"note,omitempty"`
}

type AttendanceActivitySummaryItem struct {
	Events  int `json:"events"`
	Unique  int `json:"unique"`
	Hours   float64 `json:"hours,omitempty"`
}

type AttendanceActivityReportSummary struct {
	Late       AttendanceActivitySummaryItem `json:"late"`
	Missing    AttendanceActivitySummaryItem `json:"missing"`
	Leave      AttendanceActivitySummaryItem `json:"leave"`
	Permission AttendanceActivitySummaryItem `json:"permission"`
	Overtime   AttendanceActivitySummaryItem `json:"overtime"`
	Total      int                           `json:"total"`
}

type AttendanceActivityReportResponse struct {
	Rows    []AttendanceActivityReportRow   `json:"rows"`
	Summary AttendanceActivityReportSummary `json:"summary"`
}
