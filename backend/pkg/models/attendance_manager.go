package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ManagerAttendanceUserAgg struct {
	FullName        string `bson:"full_name" json:"full_name"`
	Email           string `bson:"email" json:"email"`
	PayrollNumber   string `bson:"payroll_number" json:"payroll_number"`
	DepartmentName  string `bson:"department_name" json:"department_name"`
	PositionName    string `bson:"position_name" json:"position_name"`
	Avatar          string `bson:"avatar" json:"avatar"`
}

type ManagerAttendanceAggRow struct {
	ID              primitive.ObjectID  `bson:"_id"`
	UserID          primitive.ObjectID  `bson:"user_id"`
	Date            time.Time           `bson:"date"`
	DateStr         string              `bson:"date_str,omitempty"`
	ClockInTime     *time.Time          `bson:"clock_in_time,omitempty"`
	ClockOutTime    *time.Time          `bson:"clock_out_time,omitempty"`
	ClockInLocation GeoLocation         `bson:"clock_in_location"`
	Status          AttendanceStatus    `bson:"status"`
	User            ManagerAttendanceUserAgg `bson:"user"`
}

type ManagerAttendanceRecord struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	PayrollNumber  string `json:"payroll_number"`
	DepartmentName string `json:"department_name"`
	PositionName   string `json:"position_name"`
	Date           string `json:"date"`
	ClockInTime    string `json:"clock_in_time"`
	ClockOutTime   string `json:"clock_out_time"`
	Status         string `json:"status"`
	Location       string `json:"location"`
	Avatar         string `json:"avatar,omitempty"`
}

type ManagerAttendanceSummary struct {
	TotalRecords       int64   `json:"total_records"`
	TepatWaktu         int64   `json:"tepat_waktu"`
	Terlambat          int64   `json:"terlambat"`
	IzinSakit          int64   `json:"izin_sakit"`
	Alfa               int64   `json:"alfa"`
	TotalKehadiranPct  float64 `json:"total_kehadiran_pct"`
}

type ManagerAttendanceListResponse struct {
	Items    []ManagerAttendanceRecord `json:"items"`
	Page     int64                    `json:"page"`
	PageSize int64                    `json:"page_size"`
	Total    int64                    `json:"total"`
	Summary  ManagerAttendanceSummary  `json:"summary"`
}
