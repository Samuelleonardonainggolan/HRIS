	// pkg/models/pengajuan_izin_cuti.go
	package models

	import (
		"time"

		"go.mongodb.org/mongo-driver/bson/primitive"
	)

	type PengajuanIzinCuti struct {
		ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
		UserID          primitive.ObjectID `json:"user_id" bson:"user_id"`
		TipePengajuanID primitive.ObjectID `json:"tipe_pengajuan_id" bson:"tipe_pengajuan_id"`
		NamaTipe        string             `json:"nama_tipe" bson:"nama_tipe"`

		TanggalMulai   time.Time `json:"tanggal_mulai" bson:"tanggal_mulai"`
		TanggalSelesai time.Time `json:"tanggal_selesai" bson:"tanggal_selesai"`
		TotalHari      int       `json:"total_hari" bson:"total_hari"`
		Alasan         string    `json:"alasan" bson:"alasan"`

		DokumenURL  string              `json:"dokumen_url,omitempty" bson:"dokumen_url,omitempty"`
		KuotaCutiID *primitive.ObjectID `json:"kuota_cuti_id,omitempty" bson:"kuota_cuti_id,omitempty"`

		StatusKepalaDepartemen string             `json:"status_kepala_departemen" bson:"status_kepala_departemen"`
		KepalaDepartemenID     primitive.ObjectID `json:"kepala_departemen_id" bson:"kepala_departemen_id"`

		ManagerHRID     primitive.ObjectID `json:"manager_hr_id" bson:"manager_hr_id"`
		StatusManagerHR string             `json:"status_manager_hr" bson:"status_manager_hr"`
		StatusFinal     string             `json:"status_final" bson:"status_final"`

		CreatedAt time.Time `json:"created_at" bson:"created_at"`
		UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	}

	type CreatePengajuanIzinCutiRequest struct {
		UserID          string `json:"user_id" binding:"required"`
		TipePengajuanID string `json:"tipe_pengajuan_id" binding:"required"`
		TanggalMulai    string `json:"tanggal_mulai" binding:"required"`
		TanggalSelesai  string `json:"tanggal_selesai" binding:"required"`
		TotalHari       int    `json:"total_hari" binding:"required"`
		Alasan          string `json:"alasan" binding:"required"`
		DokumenURL      string `json:"dokumen_url,omitempty"`
	}

	type UpdatePengajuanIzinCutiRequest struct {
		TanggalMulai   *time.Time `json:"tanggal_mulai,omitempty"`
		TanggalSelesai *time.Time `json:"tanggal_selesai,omitempty"`
		TotalHari      *int       `json:"total_hari,omitempty"`
		Alasan         string     `json:"alasan,omitempty"`
		DokumenURL     string     `json:"dokumen_url,omitempty"`

		StatusKepalaDepartemen string `json:"status_kepala_departemen,omitempty"`
		StatusManagerHR        string `json:"status_manager_hr,omitempty"`
		StatusFinal            string `json:"status_final,omitempty"`
	}

	type PengajuanIzinCutiResponse struct {
		ID              string `json:"id"`
		UserID          string `json:"user_id"`
		TipePengajuanID string `json:"tipe_pengajuan_id"`
		NamaTipe        string `json:"nama_tipe"`

		TanggalMulai   time.Time `json:"tanggal_mulai"`
		TanggalSelesai time.Time `json:"tanggal_selesai"`
		TotalHari      int       `json:"total_hari"`
		Alasan         string    `json:"alasan"`

		DokumenURL  string `json:"dokumen_url,omitempty"`
		KuotaCutiID string `json:"kuota_cuti_id,omitempty"`

		StatusKepalaDepartemen string `json:"status_kepala_departemen"`
		KepalaDepartemenID     string `json:"kepala_departemen_id"`

		ManagerHRID     string `json:"manager_hr_id"`
		StatusManagerHR string `json:"status_manager_hr"`
		StatusFinal     string `json:"status_final"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	type PengajuanIzinCutiApprovalEmployeeResponse struct {
		ID             string `json:"id"`
		PayrollNumber  string `json:"payroll_number"`
		FullName       string `json:"full_name"`
		DepartmentName string `json:"department_name"`
		PositionName   string `json:"position_name"`
	}

	type PengajuanIzinCutiApprovalResponse struct {
		Pengajuan PengajuanIzinCutiResponse                  `json:"pengajuan"`
		Employee  *PengajuanIzinCutiApprovalEmployeeResponse `json:"employee,omitempty"`
	}

	func (p *PengajuanIzinCuti) ToResponse() PengajuanIzinCutiResponse {
		var kuota string
		if p.KuotaCutiID != nil {
			kuota = p.KuotaCutiID.Hex()
		}

		return PengajuanIzinCutiResponse{
			ID:                     p.ID.Hex(),
			UserID:                 p.UserID.Hex(),
			TipePengajuanID:        p.TipePengajuanID.Hex(),
			NamaTipe:               p.NamaTipe,
			TanggalMulai:           p.TanggalMulai,
			TanggalSelesai:         p.TanggalSelesai,
			TotalHari:              p.TotalHari,
			Alasan:                 p.Alasan,
			DokumenURL:             p.DokumenURL,
			KuotaCutiID:            kuota,
			StatusKepalaDepartemen: p.StatusKepalaDepartemen,
			KepalaDepartemenID:     p.KepalaDepartemenID.Hex(),
			ManagerHRID:            p.ManagerHRID.Hex(),
			StatusManagerHR:        p.StatusManagerHR,
			StatusFinal:            p.StatusFinal,
			CreatedAt:              p.CreatedAt,
			UpdatedAt:              p.UpdatedAt,
		}
	}

	// Status constants (opsional, biar konsisten)
	const (
		StatusPending  = "PENDING"
		StatusApproved = "APPROVED"
		StatusRejected = "REJECTED"
	)
