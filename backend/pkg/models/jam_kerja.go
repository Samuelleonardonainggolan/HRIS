// pkg/models/jam_kerja.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JamKerja struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`

	// ✅ renamed fields (db)
	DayOfWeek []string  `json:"day_of_week" bson:"day_of_week"` // ["Senin",...]
	StartTime time.Time `json:"start_time" bson:"start_time"`
	EndTime   time.Time `json:"end_time" bson:"end_time"`
	IsActive  bool      `json:"is_active" bson:"is_active"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type JamKerjaListRowResponse struct {
	ID         string   `json:"id"` // user_id (dipakai di frontend sebagai id)
	Name       string   `json:"name"`
	NIK        string   `json:"nik"`
	Department string   `json:"department"`
	Position   string   `json:"position"`
	DayOfWeek []string  `json:"day_of_week"`
	WorkDays   string   `json:"workDays"`  // "Senin - Jumat" | "Senin - Sabtu" | "Shift"
	StartTime  string   `json:"startTime"` // "08:00"
	EndTime    string   `json:"endTime"`   // "17:00"
}
	
type JamKerjaDetailResponse struct {
	UserID       string   `json:"user_id"`
	Name         string   `json:"name"`
	NIK          string   `json:"nik"`
	Department   string   `json:"department"`
	Position     string   `json:"position"`
	DayOfWeek []string  `json:"day_of_week"`
	StartTime  string   `json:"start_time"` // "08:00"
	EndTime    string   `json:"end_time"`   // "17:00"
	IsActive  bool      `json:"is_active"`
}

type UpdateJamKerjaRequest struct {
	DayOfWeek []string  `json:"day_of_week" binding:"required,min=1"`
	StartTime  string   `json:"start_time" binding:"required"`
	EndTime    string   `json:"end_time" binding:"required"`
	IsActive  *bool    `json:"is_active,omitempty"`
}

type CreateJamKerjaRequest struct {
	UserID       string   `json:"user_id" binding:"required"`
	DayOfWeek []string  `json:"day_of_week" binding:"required,min=1"`
	StartTime  string   `json:"start_time" binding:"required"`   // "HH:mm"
	EndTime    string   `json:"end_time" binding:"required"` // "HH:mm"
	IsActive  *bool    `json:"is_active,omitempty"`
}

type AvailableEmployeeResponse struct {
	ID         string `json:"id"`
	FullName   string `json:"full_name"`
	NIK        string `json:"nik"`
	Department string `json:"department_name"`
	Position   string `json:"position_name"`
}
