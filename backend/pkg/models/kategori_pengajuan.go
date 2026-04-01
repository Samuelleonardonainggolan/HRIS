// pkg/models/kategori_pengajuan.go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// KategoriPengajuan represents kategori pengajuan (mis: Izin, Cuti)
type KategoriPengajuan struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NamaKategori string             `json:"nama_kategori" bson:"nama_kategori"`
}

// CreateKategoriPengajuanRequest represents request to create kategori pengajuan
type CreateKategoriPengajuanRequest struct {
	NamaKategori string `json:"nama_kategori" binding:"required"`
}

// UpdateKategoriPengajuanRequest represents request to update kategori pengajuan
type UpdateKategoriPengajuanRequest struct {
	NamaKategori string `json:"nama_kategori,omitempty"`
}

// KategoriPengajuanResponse represents response for kategori pengajuan
type KategoriPengajuanResponse struct {
	ID           string `json:"id"`
	NamaKategori string `json:"nama_kategori"`
}

// ToResponse converts KategoriPengajuan to KategoriPengajuanResponse
func (k *KategoriPengajuan) ToResponse() KategoriPengajuanResponse {
	return KategoriPengajuanResponse{
		ID:           k.ID.Hex(),
		NamaKategori: k.NamaKategori,
	}
}