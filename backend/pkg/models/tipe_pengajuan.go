// pkg/models/tipe_pengajuan.go
package models

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RequestType represents request type (mis: Cuti Tahunan, Izin Sakit)
type RequestType struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TypeName           string             `json:"type_name" bson:"type_name"`
	RequestCategoryID  string             `json:"request_category_id" bson:"request_category_id"`
	CategoryName       string             `json:"category_name" bson:"category_name"`
	QuotaDeduction     bool               `json:"quota_deduction" bson:"quota_deduction"`
	PotongKuota        bool               `json:"potong_kuota,omitempty" bson:"potong_kuota,omitempty"`
	AttachmentRequired bool               `json:"attachment_required" bson:"attachment_required"`
}

// CreateRequestTypeRequest represents request to create request type
type CreateRequestTypeRequest struct {
	TypeName           string `json:"type_name" binding:"required"`
	RequestCategoryID  string `json:"request_category_id" binding:"required"`
	QuotaDeduction     bool   `json:"quota_deduction"`
	AttachmentRequired bool   `json:"attachment_required"`
}

// UpdateRequestTypeRequest represents request to update request type
type UpdateRequestTypeRequest struct {
	TypeName           string `json:"type_name,omitempty"`
	RequestCategoryID  string `json:"request_category_id,omitempty"`
	CategoryName       string `json:"category_name,omitempty"`
	QuotaDeduction     *bool  `json:"quota_deduction,omitempty"`
	AttachmentRequired *bool  `json:"attachment_required,omitempty"`
}

// RequestTypeResponse represents response for request type
type RequestTypeResponse struct {
	ID                 string `json:"id"`
	TypeName           string `json:"type_name"`
	RequestCategoryID  string `json:"request_category_id"`
	CategoryName       string `json:"category_name"`
	QuotaDeduction     bool   `json:"quota_deduction"`
	AttachmentRequired bool   `json:"attachment_required"`
}

// ToResponse converts RequestType to response
func (t *RequestType) ToResponse() RequestTypeResponse {
	quotaDeduction := t.QuotaDeduction || t.PotongKuota || isQuotaDeductingCategory(t.CategoryName)
	return RequestTypeResponse{
		ID:                 t.ID.Hex(),
		TypeName:           t.TypeName,
		RequestCategoryID:  t.RequestCategoryID,
		CategoryName:       t.CategoryName,
		QuotaDeduction:     quotaDeduction,
		AttachmentRequired: t.AttachmentRequired,
	}
}

func isQuotaDeductingCategory(categoryName string) bool {
	switch strings.ToLower(strings.TrimSpace(categoryName)) {
	case "izin", "cuti":
		return true
	default:
		return false
	}
}
