package service

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PengajuanService interface {
	GetTipePengajuan(ctx context.Context) ([]models.TipePengajuanResponse, error)
	GetMyPengajuan(ctx context.Context, userID string) ([]models.PengajuanIzinCutiResponse, error)
	CreatePengajuan(ctx context.Context, userID string, req CreatePengajuanRequest) (*models.PengajuanIzinCutiResponse, error)
}

type CreatePengajuanRequest struct {
	TipePengajuanID string
	TanggalMulai    string
	TanggalSelesai  string
	TotalHari       int
	Alasan          string
	DokumenURL      string
}

type pengajuanService struct {
	db *mongo.Database
}

func NewPengajuanService(db *mongo.Database) PengajuanService {
	return &pengajuanService{db: db}
}

func (s *pengajuanService) GetTipePengajuan(ctx context.Context) ([]models.TipePengajuanResponse, error) {
	coll := s.db.Collection("tipe_pengajuan")
	cur, err := coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "nama_tipe", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var items []models.TipePengajuan
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}

	out := make([]models.TipePengajuanResponse, 0, len(items))
	for i := range items {
		out = append(out, items[i].ToResponse())
	}
	return out, nil
}

func (s *pengajuanService) GetMyPengajuan(ctx context.Context, userID string) ([]models.PengajuanIzinCutiResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	coll := s.db.Collection("pengajuan_izin_cuti")
	cur, err := coll.Find(ctx, bson.M{"user_id": uid}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var items []models.PengajuanIzinCuti
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}

	out := make([]models.PengajuanIzinCutiResponse, 0, len(items))
	for i := range items {
		out = append(out, items[i].ToResponse())
	}
	return out, nil
}

func (s *pengajuanService) CreatePengajuan(ctx context.Context, userID string, req CreatePengajuanRequest) (*models.PengajuanIzinCutiResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	tipeID, err := primitive.ObjectIDFromHex(req.TipePengajuanID)
	if err != nil {
		return nil, errors.New("tipe_pengajuan_id tidak valid")
	}

	start, err := parseTimeFlexible(req.TanggalMulai)
	if err != nil {
		return nil, errors.New("tanggal_mulai tidak valid")
	}
	end, err := parseTimeFlexible(req.TanggalSelesai)
	if err != nil {
		return nil, errors.New("tanggal_selesai tidak valid")
	}

	var tipe models.TipePengajuan
	_ = s.db.Collection("tipe_pengajuan").FindOne(ctx, bson.M{"_id": tipeID}).Decode(&tipe)

	now := time.Now()
	doc := models.PengajuanIzinCuti{
		ID:                   primitive.NewObjectID(),
		UserID:               uid,
		TipePengajuanID:      tipeID,
		NamaTipe:             tipe.NamaTipe,
		TanggalMulai:         start,
		TanggalSelesai:       end,
		TotalHari:            req.TotalHari,
		Alasan:               req.Alasan,
		DokumenURL:           req.DokumenURL,
		KuotaCutiID:          nil,
		StatusKepalaDepartemen: models.StatusPending,
		KepalaDepartemenID:     primitive.NilObjectID,
		ManagerHRID:            primitive.NilObjectID,
		StatusManagerHR:        models.StatusPending,
		StatusFinal:            models.StatusPending,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	if _, err := s.db.Collection("pengajuan_izin_cuti").InsertOne(ctx, doc); err != nil {
		return nil, err
	}

	resp := doc.ToResponse()
	return &resp, nil
}

func parseTimeFlexible(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("invalid time")
}

