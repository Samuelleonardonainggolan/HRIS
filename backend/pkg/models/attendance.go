// pkg/models/attendance.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AttendanceStatus string

const (
	StatusOnTime   AttendanceStatus = "On Time"
	StatusLate     AttendanceStatus = "Late"
	StatusAbsent   AttendanceStatus = "Absent"
	StatusOvertime AttendanceStatus = "Overtime"
)

type Attendance struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`
	Date             time.Time          `json:"date" bson:"date"`
	ClockInTime      *time.Time         `json:"clock_in_time,omitempty" bson:"clock_in_time,omitempty"`
	ClockOutTime     *time.Time         `json:"clock_out_time,omitempty" bson:"clock_out_time,omitempty"`
	BreakStartTime   *time.Time         `json:"break_start_time,omitempty" bson:"break_start_time,omitempty"`
	BreakEndTime     *time.Time         `json:"break_end_time,omitempty" bson:"break_end_time,omitempty"`
	ClockInPhoto     string             `json:"clock_in_photo,omitempty" bson:"clock_in_photo,omitempty"`
	ClockOutPhoto    string             `json:"clock_out_photo,omitempty" bson:"clock_out_photo,omitempty"`
	ClockInLocation  GeoLocation        `json:"clock_in_location" bson:"clock_in_location"`
	ClockOutLocation GeoLocation        `json:"clock_out_location,omitempty" bson:"clock_out_location,omitempty"`
	Status           AttendanceStatus   `json:"status" bson:"status"`
	WorkHours        float64            `json:"work_hours" bson:"work_hours"`
	OvertimeHours    float64            `json:"overtime_hours" bson:"overtime_hours"`
	FaceSimilarity   float64            `json:"face_similarity,omitempty" bson:"face_similarity,omitempty"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}

type GeoLocation struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	Address   string  `json:"address,omitempty" bson:"address,omitempty"`
}

type AttendanceRequest struct {
	UserID    string  `json:"user_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

type AttendanceResponse struct {
	ID             string           `json:"id"`
	Date           string           `json:"date"`
	ClockInTime    string           `json:"clock_in_time"`
	ClockOutTime   string           `json:"clock_out_time"`
	BreakStartTime string           `json:"break_start_time,omitempty"`
	BreakEndTime   string           `json:"break_end_time,omitempty"`
	Status         AttendanceStatus `json:"status"`
	WorkHours      float64          `json:"work_hours"`
	OvertimeHours  float64          `json:"overtime_hours"`
	FaceSimilarity float64          `json:"face_similarity,omitempty"`
}

type MonthlyAttendanceResponse struct {
	Month         string               `json:"month"`
	Year          int                  `json:"year"`
	TotalDays     int                  `json:"total_days"`
	TotalHours    float64              `json:"total_hours"`
	OvertimeHours float64              `json:"overtime_hours"`
	Records       []AttendanceResponse `json:"records"`
}
