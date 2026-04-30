// pkg/models/jam_kerja.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JamKerja struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

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
	Avatar     string   `json:"avatar,omitempty"`
	Department string   `json:"department"`
	Position   string   `json:"position"`
	DayOfWeek  []string `json:"day_of_week"`
	WorkDays   string   `json:"work_days"`  // "Senin - Jumat" | "Senin - Sabtu" | "Shift"
	StartTime  string   `json:"start_time"` // "08:00"
	EndTime    string   `json:"end_time"`   // "17:00"
}

type JamKerjaDetailResponse struct {
	UserID     string   `json:"user_id"`
	Name       string   `json:"name"`
	NIK        string   `json:"nik"`
	Department string   `json:"department"`
	Position   string   `json:"position"`
	DayOfWeek  []string `json:"day_of_week"`
	StartTime  string   `json:"start_time"` // "08:00"
	EndTime    string   `json:"end_time"`   // "17:00"
	IsActive   bool     `json:"is_active"`
}

type UpdateJamKerjaRequest struct {
	// ✅ format baru
	DayOfWeek []string `json:"day_of_week"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	IsActive  *bool    `json:"is_active,omitempty"`

	// ✅ format lama (biar tidak break)
	HariKerja    []string `json:"hari_kerja,omitempty"`
	WaktuMulai   string   `json:"waktu_mulai,omitempty"`
	WaktuSelesai string   `json:"waktu_selesai,omitempty"`
	Aktif        *bool    `json:"aktif,omitempty"`
}

type CreateJamKerjaRequest struct {
	UserID string `json:"user_id" binding:"required"`

	// ✅ format baru
	DayOfWeek []string `json:"day_of_week"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	IsActive  *bool    `json:"is_active,omitempty"`

	// ✅ format lama (biar tidak break)
	HariKerja    []string `json:"hari_kerja,omitempty"`
	WaktuMulai   string   `json:"waktu_mulai,omitempty"`
	WaktuSelesai string   `json:"waktu_selesai,omitempty"`
	Aktif        *bool    `json:"aktif,omitempty"`
}

type AvailableEmployeeResponse struct {
	ID         string `json:"id"`
	FullName   string `json:"full_name"`
	NIK        string `json:"nik"`
	Department string `json:"department_name"`
	Position   string `json:"position_name"`
}
