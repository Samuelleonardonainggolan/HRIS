// internal/service/pengajuan_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PengajuanService interface mendefinisikan semua operasi pengajuan.
type PengajuanService interface {
	GetAllTipePengajuan(ctx context.Context) ([]models.RequestTypeResponse, error)
	CreatePengajuan(ctx context.Context, req CreatePengajuanRequest) (*models.PengajuanIzinCutiResponse, error)
	GetPengajuanByUser(ctx context.Context, userID string) ([]models.PengajuanIzinCutiResponse, error)
}

// CreatePengajuanRequest adalah request untuk membuat pengajuan baru.
type CreatePengajuanRequest struct {
	UserID          string `json:"user_id"`
	TipePengajuanID string `json:"tipe_pengajuan_id" binding:"required"`
	TanggalMulai    string `json:"tanggal_mulai" binding:"required"`   // format: yyyy-MM-dd
	TanggalSelesai  string `json:"tanggal_selesai" binding:"required"` // format: yyyy-MM-dd
	TotalHari       int    `json:"total_hari" binding:"required"`
	Alasan          string `json:"alasan" binding:"required"`
	DokumenURL      string `json:"dokumen_url,omitempty"`
}

// pengajuanService adalah implementasi konkret dari PengajuanService.
type pengajuanService struct {
	tipePengajuanCol interface {
		FindAll(ctx context.Context) ([]models.RequestType, error)
	}
	pengajuanCol interface {
		Create(ctx context.Context, p *models.LeaveRequest) error
		FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.LeaveRequest, error)
		FindTipeByID(ctx context.Context, id primitive.ObjectID) (*models.RequestType, error)
	}
	db *mongo.Database
}

// NewPengajuanService membuat service pengajuan baru.
// Karena repository belum tentu ada, kita buat implementasi langsung ke MongoDB
// via db parameter untuk kesederhanaan integrasi.
func NewPengajuanService(db *mongo.Database) PengajuanService {
	return &pengajuanServiceImpl{db: db}
}

// ── Implementasi konkret menggunakan MongoDB langsung ────────────────────────

type pengajuanServiceImpl struct {
	db *mongo.Database
}

func (s *pengajuanServiceImpl) GetAllTipePengajuan(ctx context.Context) ([]models.RequestTypeResponse, error) {
	col := s.db.Collection("request_type")

	cursor, err := col.Find(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tipes []models.RequestType
	if err = cursor.All(ctx, &tipes); err != nil {
		return nil, err
	}

	result := make([]models.RequestTypeResponse, 0, len(tipes))
	for _, t := range tipes {
		result = append(result, t.ToResponse())
	}
	return result, nil
}

func (s *pengajuanServiceImpl) CreatePengajuan(ctx context.Context, req CreatePengajuanRequest) (*models.PengajuanIzinCutiResponse, error) {
	// Validasi user ID
	userObjID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	// Validasi tipe pengajuan ID
	tipeObjID, err := primitive.ObjectIDFromHex(req.TipePengajuanID)
	if err != nil {
		return nil, errors.New("tipe pengajuan ID tidak valid")
	}

	// Ambil detail tipe pengajuan
	var tipe models.RequestType
	err = s.db.Collection("request_type").FindOne(ctx, map[string]interface{}{"_id": tipeObjID}).Decode(&tipe)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("tipe pengajuan tidak ditemukan")
		}
		return nil, err
	}

	// Parse tanggal
	const layout = "2006-01-02"
	tanggalMulai, err := time.ParseInLocation(layout, req.TanggalMulai, time.FixedZone("WIB", 7*60*60))
	if err != nil {
		return nil, errors.New("format tanggal_mulai tidak valid (gunakan yyyy-MM-dd)")
	}
	tanggalSelesai, err := time.ParseInLocation(layout, req.TanggalSelesai, time.FixedZone("WIB", 7*60*60))
	if err != nil {
		return nil, errors.New("format tanggal_selesai tidak valid (gunakan yyyy-MM-dd)")
	}

	if tanggalSelesai.Before(tanggalMulai) {
		return nil, errors.New("tanggal_selesai tidak boleh sebelum tanggal_mulai")
	}

	// Cari kepala departemen dan manager HR (pakai approver pertama sebagai fallback)
	// TODO: sesuaikan dengan struktur organisasi backend Anda
	var approver struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	_ = s.db.Collection("users").FindOne(ctx, map[string]interface{}{"role": "manager_hr"}).Decode(&approver)
	if approver.ID.IsZero() {
		// fallback ke user pertama jika tidak ada manager HR
		_ = s.db.Collection("users").FindOne(ctx, map[string]interface{}{}).Decode(&approver)
	}

	now := time.Now()
	pengajuan := &models.LeaveRequest{
		ID:                     primitive.NewObjectID(),
		UserID:                 userObjID,
		RequestTypeID:        	tipeObjID,
		TypeName:               tipe.TypeName,
		StartDate:           	tanggalMulai,
		EndDate:         		tanggalSelesai,
		DaysTotal:              req.TotalHari,
		Reason:                 req.Alasan,
		DocumentURL:            req.DokumenURL,
		StatusKepalaDepartemen: models.StatusPending,
		KepalaDepartemenID:     approver.ID,
		ManagerHRID:            approver.ID,
		StatusManagerHR:        models.StatusPending,
		FinalStatus:            models.StatusPending,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	_, err = s.db.Collection("pengajuan_izin_cuti").InsertOne(ctx, pengajuan)
	if err != nil {
		return nil, err
	}

	resp := pengajuan.ToResponse()
	return &resp, nil
}

func (s *pengajuanServiceImpl) GetPengajuanByUser(ctx context.Context, userID string) ([]models.PengajuanIzinCutiResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	cursor, err := s.db.Collection("pengajuan_izin_cuti").Find(ctx, map[string]interface{}{"user_id": userObjID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.LeaveRequest
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	result := make([]models.PengajuanIzinCutiResponse, 0, len(list))
	for _, p := range list {
		result = append(result, p.ToResponse())
	}
	return result, nil
}
