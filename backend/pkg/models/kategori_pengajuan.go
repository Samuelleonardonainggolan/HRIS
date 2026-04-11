// pkg/models/kategori_pengajuan.go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// RequestCategory represents request category (mis: Izin, Cuti)
// (sebelumnya: KategoriPengajuan)
type RequestCategory struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CategoryName string             `json:"category_name" bson:"category_name"`
}

// CreateRequestCategoryRequest represents request to create request category
type CreateRequestCategoryRequest struct {
	CategoryName string `json:"category_name" binding:"required"`
}

// UpdateRequestCategoryRequest represents request to update request category
type UpdateRequestCategoryRequest struct {
	CategoryName string `json:"category_name,omitempty"`
}

// RequestCategoryResponse represents response for request category
type RequestCategoryResponse struct {
	ID           string `json:"id"`
	CategoryName string `json:"category_name"`
}

// ToResponse converts RequestCategory to RequestCategoryResponse
func (k *RequestCategory) ToResponse() RequestCategoryResponse {
	return RequestCategoryResponse{
		ID:           k.ID.Hex(),
		CategoryName: k.CategoryName,
	}
}