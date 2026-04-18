// pkg/models/overtime_request.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OvertimeRequest struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID             primitive.ObjectID `json:"user_id" bson:"user_id"`
	Date               time.Time          `json:"date" bson:"date"`
	StartTime           time.Time          `json:"start_time" bson:"start_time"`
	EndTime             time.Time          `json:"end_time" bson:"end_time"`
	Total              string             `json:"total" bson:"total"`

	StatusKepalaDepartemen string             `json:"status_kepala_departemen" bson:"status_kepala_departemen"`
	KepalaDepartemenID     primitive.ObjectID `json:"kepala_departemen_id" bson:"kepala_departemen_id"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type CreateOvertimeRequestRequest struct {
	UserID string `json:"user_id" binding:"required"`

	// date bisa "YYYY-MM-DD"
	Date string `json:"date" binding:"required"`

	// start_time/end_time bisa "HH:mm" atau ISO string tergantung implementasi handler/service
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`

	Total string `json:"total" binding:"required"`

	KepalaDepartemenID string `json:"kepala_departemen_id" binding:"required"`
}

type UpdateOvertimeRequestRequest struct {
	Date      *time.Time `json:"date,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Total     string     `json:"total,omitempty"`

	StatusKepalaDepartemen string `json:"status_kepala_departemen,omitempty"`
	KepalaDepartemenID     string `json:"kepala_departemen_id,omitempty"`
}

type OvertimeRequestResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`

	Date      time.Time `json:"date"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Total     string    `json:"total"`

	StatusKepalaDepartemen string `json:"status_kepala_departemen"`
	KepalaDepartemenID     string `json:"kepala_departemen_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (o *OvertimeRequest) ToResponse() OvertimeRequestResponse {
	resp := OvertimeRequestResponse{
		ID:                    o.ID.Hex(),
		Date:                  o.Date,
		StartTime:              o.StartTime,
		EndTime:                o.EndTime,
		Total:                 o.Total,
		StatusKepalaDepartemen: o.StatusKepalaDepartemen,
		CreatedAt:             o.CreatedAt,
		UpdatedAt:             o.UpdatedAt,
	}

	if !o.UserID.IsZero() {
		resp.UserID = o.UserID.Hex()
	}
	if !o.KepalaDepartemenID.IsZero() {
		resp.KepalaDepartemenID = o.KepalaDepartemenID.Hex()
	}

	return resp
}