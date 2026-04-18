// pkg/models/leave_balance.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeaveBalance struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID         primitive.ObjectID `json:"user_id" bson:"user_id"`
	Year           int                `json:"year" bson:"year"`
	TotalKuota     int                `json:"total_kuota" bson:"total_kuota"`
	UsedKuota      int                `json:"used_kuota" bson:"used_kuota"`
	RemainingKuota int                `json:"remaining_kuota" bson:"remaining_kuota"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

type CreateLeaveBalanceRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Year       int    `json:"year" binding:"required"`
	TotalKuota int    `json:"total_kuota" binding:"required"`
	UsedKuota  int    `json:"used_kuota"`
	// RemainingKuota optional: bisa dihitung TotalKuota - UsedKuota di service
	RemainingKuota *int `json:"remaining_kuota,omitempty"`
}

type UpdateLeaveBalanceRequest struct {
	Year       *int `json:"year,omitempty"`
	TotalKuota *int `json:"total_kuota,omitempty"`
	UsedKuota  *int `json:"used_kuota,omitempty"`
	RemainingKuota *int `json:"remaining_kuota,omitempty"`
}

type LeaveBalanceResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Year           int       `json:"year"`
	TotalKuota     int       `json:"total_kuota"`
	UsedKuota      int       `json:"used_kuota"`
	RemainingKuota int       `json:"remaining_kuota"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (l *LeaveBalance) ToResponse() LeaveBalanceResponse {
	resp := LeaveBalanceResponse{
		ID:             l.ID.Hex(),
		Year:           l.Year,
		TotalKuota:     l.TotalKuota,
		UsedKuota:      l.UsedKuota,
		RemainingKuota: l.RemainingKuota,
		CreatedAt:      l.CreatedAt,
		UpdatedAt:      l.UpdatedAt,
	}

	if !l.UserID.IsZero() {
		resp.UserID = l.UserID.Hex()
	}

	return resp
}