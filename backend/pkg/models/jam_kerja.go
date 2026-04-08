// pkg/models/jam_kerja.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JamKerja struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	HariKerja    []string           `json:"hari_kerja" bson:"hari_kerja"` // ["Senin",...]
	WaktuMulai   time.Time          `json:"waktu_mulai" bson:"waktu_mulai"`
	WaktuSelesai time.Time          `json:"waktu_selesai" bson:"waktu_selesai"`
	Aktif        bool               `json:"aktif" bson:"aktif"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

type JamKerjaListRowResponse struct {
	ID         string `json:"id"` // user_id (dipakai di frontend sebagai id)
	Name       string `json:"name"`
	NIK        string `json:"nik"`
	Department string `json:"department"`
	Position   string `json:"position"`
	HariKerja  []string `json:"hari_kerja"`
	WorkDays   string `json:"workDays"`  // "Senin - Jumat" | "Senin - Sabtu" | "Shift"
	StartTime  string `json:"startTime"` // "08:00"
	EndTime    string `json:"endTime"`   // "17:00"
}

type JamKerjaDetailResponse struct {
	UserID       string   `json:"user_id"`
	Name         string   `json:"name"`
	NIK          string   `json:"nik"`
	Department   string   `json:"department"`
	Position     string   `json:"position"`
	HariKerja    []string `json:"hari_kerja"`
	WaktuMulai   string   `json:"waktu_mulai"`   // "HH:mm"
	WaktuSelesai string   `json:"waktu_selesai"` // "HH:mm"
	Aktif        bool     `json:"aktif"`
}

type UpdateJamKerjaRequest struct {
	HariKerja    []string `json:"hari_kerja" binding:"required,min=1"`
	WaktuMulai   string   `json:"waktu_mulai" binding:"required"`
	WaktuSelesai string   `json:"waktu_selesai" binding:"required"`
	Aktif        *bool    `json:"aktif,omitempty"`
}

type CreateJamKerjaRequest struct {
	UserID       string   `json:"user_id" binding:"required"`
	HariKerja    []string `json:"hari_kerja" binding:"required,min=1"`
	WaktuMulai   string   `json:"waktu_mulai" binding:"required"`   // "HH:mm"
	WaktuSelesai string   `json:"waktu_selesai" binding:"required"` // "HH:mm"
	Aktif        *bool    `json:"aktif,omitempty"`
}

type AvailableEmployeeResponse struct {
	ID            string `json:"id"`
	FullName      string `json:"full_name"`
	NIK           string `json:"nik"`
	Department    string `json:"department_name"`
	Position      string `json:"position_name"`
}