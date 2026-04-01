// pkg/models/tipe_pengajuan.go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// TipePengajuan represents tipe pengajuan (mis: Cuti Tahunan, Izin Sakit)
type TipePengajuan struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NamaTipe          string             `json:"nama_tipe" bson:"nama_tipe"`
	KategoriPengajuanID string           `json:"kategori_pengajuan_id" bson:"kategori_pengajuan_id"`
	NamaKategori      string             `json:"nama_kategori" bson:"nama_kategori"`
	PotongKuota       bool               `json:"potong_kuota" bson:"potong_kuota"`
	WajibLampiran     bool               `json:"wajib_lampiran" bson:"wajib_lampiran"`
}

// CreateTipePengajuanRequest represents request to create tipe pengajuan
type CreateTipePengajuanRequest struct {
	NamaTipe            string `json:"nama_tipe" binding:"required"`
	KategoriPengajuanID string `json:"kategori_pengajuan_id" binding:"required"`
	PotongKuota         bool   `json:"potong_kuota"`
	WajibLampiran       bool   `json:"wajib_lampiran"`
}

// UpdateTipePengajuanRequest represents request to update tipe pengajuan
type UpdateTipePengajuanRequest struct {
	NamaTipe            string `json:"nama_tipe,omitempty"`
	KategoriPengajuanID string `json:"kategori_pengajuan_id,omitempty"`
	NamaKategori        string `json:"nama_kategori,omitempty"`
	PotongKuota         *bool  `json:"potong_kuota,omitempty"`
	WajibLampiran       *bool  `json:"wajib_lampiran,omitempty"`
}

// TipePengajuanResponse represents response for tipe pengajuan
type TipePengajuanResponse struct {
	ID                  string `json:"id"`
	NamaTipe            string `json:"nama_tipe"`
	KategoriPengajuanID string `json:"kategori_pengajuan_id"`
	NamaKategori        string `json:"nama_kategori"`
	PotongKuota         bool   `json:"potong_kuota"`
	WajibLampiran       bool   `json:"wajib_lampiran"`
}

// ToResponse converts TipePengajuan to response
func (t *TipePengajuan) ToResponse() TipePengajuanResponse {
	return TipePengajuanResponse{
		ID:                  t.ID.Hex(),
		NamaTipe:            t.NamaTipe,
		KategoriPengajuanID: t.KategoriPengajuanID,
		NamaKategori:        t.NamaKategori,
		PotongKuota:         t.PotongKuota,
		WajibLampiran:       t.WajibLampiran,
	}
}