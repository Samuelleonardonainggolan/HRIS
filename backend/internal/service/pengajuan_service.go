// internal/service/pengajuan_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PengajuanService interface mendefinisikan semua operasi pengajuan.
type PengajuanService interface {
	GetAllTipePengajuan(ctx context.Context) ([]models.RequestTypeResponse, error)
	CreatePengajuan(ctx context.Context, req CreatePengajuanRequest) (*models.PengajuanIzinCutiResponse, error)
	GetPengajuanByUser(ctx context.Context, userID string) ([]models.PengajuanIzinCutiResponse, error)
	UpdatePengajuan(ctx context.Context, userID, pengajuanID string, req UpdatePengajuanRequest) (*models.PengajuanIzinCutiResponse, error)
	CancelPengajuan(ctx context.Context, userID, pengajuanID string) error
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

// UpdatePengajuanRequest adalah request untuk mengubah pengajuan milik user.
type UpdatePengajuanRequest struct {
	TipePengajuanID *string `json:"tipe_pengajuan_id,omitempty"`
	TanggalMulai    *string `json:"tanggal_mulai,omitempty"`   // format: yyyy-MM-dd
	TanggalSelesai  *string `json:"tanggal_selesai,omitempty"` // format: yyyy-MM-dd
	TotalHari       *int    `json:"total_hari,omitempty"`
	Alasan          *string `json:"alasan,omitempty"`
	DokumenURL      *string `json:"dokumen_url,omitempty"`
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

	var requester models.User
	err = s.db.Collection("users").FindOne(ctx, bson.M{"_id": userObjID}).Decode(&requester)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, err
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
	tanggalMulai, err := time.Parse(layout, req.TanggalMulai)
	if err != nil {
		return nil, errors.New("format tanggal_mulai tidak valid (gunakan yyyy-MM-dd)")
	}
	tanggalSelesai, err := time.Parse(layout, req.TanggalSelesai)
	if err != nil {
		return nil, errors.New("format tanggal_selesai tidak valid (gunakan yyyy-MM-dd)")
	}
	tanggalMulai = dateOnlyUTC(tanggalMulai)
	tanggalSelesai = dateOnlyUTC(tanggalSelesai)

	if tanggalSelesai.Before(tanggalMulai) {
		return nil, errors.New("tanggal_selesai tidak boleh sebelum tanggal_mulai")
	}
	if isSameCalendarDate(tanggalMulai, tanggalSelesai) {
		return nil, errors.New("tanggal_mulai dan tanggal_selesai tidak boleh sama")
	}

	if err := validateRequestLeadTime(tipe.TypeName, tanggalMulai); err != nil {
		return nil, err
	}
	if err := s.ensureNoDateConflict(ctx, userObjID, tanggalMulai, tanggalSelesai, primitive.NilObjectID); err != nil {
		return nil, err
	}

	var dept models.Department
	_ = s.db.Collection("departments").FindOne(ctx, bson.M{"_id": requester.DepartmentID}).Decode(&dept)

	var managerHR struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	_ = s.db.Collection("users").FindOne(
		ctx,
		bson.M{"role": models.RoleManagerHR},
	).Decode(&managerHR)

	kepalaDepartemenID := dept.ManagerID
	if kepalaDepartemenID.IsZero() {
		kepalaDepartemenID = managerHR.ID
	}
	if kepalaDepartemenID.IsZero() {
		kepalaDepartemenID = requester.ID
	}

	managerHRID := managerHR.ID
	if managerHRID.IsZero() {
		managerHRID = kepalaDepartemenID
	}

	now := time.Now()
	pengajuan := &models.LeaveRequest{
		ID:                     primitive.NewObjectID(),
		UserID:                 userObjID,
		RequestTypeID:          tipeObjID,
		TypeName:               tipe.TypeName,
		StartDate:              tanggalMulai,
		EndDate:                tanggalSelesai,
		DaysTotal:              req.TotalHari,
		Reason:                 req.Alasan,
		DocumentURL:            req.DokumenURL,
		StatusKepalaDepartemen: models.StatusPending,
		KepalaDepartemenID:     kepalaDepartemenID,
		ManagerHRID:            managerHRID,
		StatusManagerHR:        models.StatusPending,
		FinalStatus:            models.StatusPending,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	_, err = s.db.Collection("leave_request").InsertOne(ctx, pengajuan)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("pengajuan sudah ada untuk rentang tanggal tersebut")
		}
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

	cursor, err := s.db.Collection("leave_request").Find(ctx, bson.M{"user_id": userObjID})
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

func (s *pengajuanServiceImpl) UpdatePengajuan(ctx context.Context, userID, pengajuanID string, req UpdatePengajuanRequest) (*models.PengajuanIzinCutiResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	pengajuanObjID, err := primitive.ObjectIDFromHex(pengajuanID)
	if err != nil {
		return nil, errors.New("ID pengajuan tidak valid")
	}

	var current models.LeaveRequest
	err = s.db.Collection("leave_request").FindOne(ctx, bson.M{
		"_id":     pengajuanObjID,
		"user_id": userObjID,
	}).Decode(&current)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan tidak ditemukan")
		}
		return nil, err
	}

	if strings.ToUpper(current.FinalStatus) != models.StatusPending {
		return nil, errors.New("pengajuan sudah diproses, tidak dapat diubah")
	}

	set := bson.M{"updated_at": time.Now()}

	if req.TipePengajuanID != nil {
		tipeID := strings.TrimSpace(*req.TipePengajuanID)
		if tipeID != "" {
			tipeObjID, err := primitive.ObjectIDFromHex(tipeID)
			if err != nil {
				return nil, errors.New("tipe pengajuan ID tidak valid")
			}

			var tipe models.RequestType
			err = s.db.Collection("request_type").FindOne(ctx, bson.M{"_id": tipeObjID}).Decode(&tipe)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return nil, errors.New("tipe pengajuan tidak ditemukan")
				}
				return nil, err
			}

			set["request_type_id"] = tipeObjID
			set["type_name"] = tipe.TypeName
			current.TypeName = tipe.TypeName
		}
	}

	startDate := current.StartDate
	endDate := current.EndDate
	const layout = "2006-01-02"

	if req.TanggalMulai != nil {
		tanggalMulai := strings.TrimSpace(*req.TanggalMulai)
		if tanggalMulai != "" {
			parsed, err := time.Parse(layout, tanggalMulai)
			if err != nil {
				return nil, errors.New("format tanggal_mulai tidak valid (gunakan yyyy-MM-dd)")
			}
			startDate = dateOnlyUTC(parsed)
			set["start_date"] = parsed
		}
	}

	if req.TanggalSelesai != nil {
		tanggalSelesai := strings.TrimSpace(*req.TanggalSelesai)
		if tanggalSelesai != "" {
			parsed, err := time.Parse(layout, tanggalSelesai)
			if err != nil {
				return nil, errors.New("format tanggal_selesai tidak valid (gunakan yyyy-MM-dd)")
			}
			endDate = dateOnlyUTC(parsed)
			set["end_date"] = parsed
		}
	}

	if endDate.Before(startDate) {
		return nil, errors.New("tanggal_selesai tidak boleh sebelum tanggal_mulai")
	}
	if isSameCalendarDate(startDate, endDate) {
		return nil, errors.New("tanggal_mulai dan tanggal_selesai tidak boleh sama")
	}

	if err := validateRequestLeadTime(current.TypeName, startDate); err != nil {
		return nil, err
	}
	if err := s.ensureNoDateConflict(ctx, userObjID, startDate, endDate, pengajuanObjID); err != nil {
		return nil, err
	}

	if req.TotalHari != nil {
		if *req.TotalHari <= 0 {
			return nil, errors.New("total_hari harus lebih dari 0")
		}
		set["days_total"] = *req.TotalHari
	}

	if req.Alasan != nil {
		set["reason"] = strings.TrimSpace(*req.Alasan)
	}

	if req.DokumenURL != nil {
		set["document_url"] = strings.TrimSpace(*req.DokumenURL)
	}

	_, err = s.db.Collection("leave_request").UpdateOne(
		ctx,
		bson.M{"_id": pengajuanObjID, "user_id": userObjID, "final_status": models.StatusPending},
		bson.M{"$set": set},
	)
	if err != nil {
		return nil, err
	}

	var updated models.LeaveRequest
	err = s.db.Collection("leave_request").FindOne(ctx, bson.M{"_id": pengajuanObjID}).Decode(&updated)
	if err != nil {
		return nil, err
	}

	resp := updated.ToResponse()
	return &resp, nil
}

func (s *pengajuanServiceImpl) CancelPengajuan(ctx context.Context, userID, pengajuanID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("user ID tidak valid")
	}

	pengajuanObjID, err := primitive.ObjectIDFromHex(pengajuanID)
	if err != nil {
		return errors.New("ID pengajuan tidak valid")
	}

	res, err := s.db.Collection("leave_request").UpdateOne(
		ctx,
		bson.M{"_id": pengajuanObjID, "user_id": userObjID, "final_status": models.StatusPending},
		bson.M{"$set": bson.M{
			"status_kepala_departemen": models.StatusRejected,
			"status_manager_hr":        models.StatusRejected,
			"final_status":             models.StatusRejected,
			"updated_at":               time.Now(),
		}},
	)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("pengajuan tidak ditemukan atau sudah diproses")
	}

	return nil
}

func validateRequestLeadTime(typeName string, startDate time.Time) error {
	if isSickLeave(typeName) {
		return nil
	}

	loc := time.FixedZone("WIB", 7*60*60)
	now := time.Now().In(loc)
	minStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 2)
	startAtDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)

	if startAtDay.Before(minStart) {
		return errors.New("pengajuan hanya boleh diajukan minimal H-2 (kecuali izin sakit)")
	}

	return nil
}

func isSickLeave(typeName string) bool {
	n := strings.ToLower(strings.TrimSpace(typeName))
	return strings.Contains(n, "sakit")
}

func dateOnlyUTC(t time.Time) time.Time {
	local := t.In(time.FixedZone("WIB", 7*60*60))
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func isSameCalendarDate(a, b time.Time) bool {
	aLocal := a.In(time.FixedZone("WIB", 7*60*60))
	bLocal := b.In(time.FixedZone("WIB", 7*60*60))
	return aLocal.Year() == bLocal.Year() && aLocal.Month() == bLocal.Month() && aLocal.Day() == bLocal.Day()
}

func (s *pengajuanServiceImpl) ensureNoDateConflict(
	ctx context.Context,
	userID primitive.ObjectID,
	startDate time.Time,
	endDate time.Time,
	excludeID primitive.ObjectID,
) error {
	filter := bson.M{
		"user_id":      userID,
		"final_status": bson.M{"$ne": models.StatusRejected},
		"start_date":   bson.M{"$lte": endDate},
		"end_date":     bson.M{"$gte": startDate},
	}
	if !excludeID.IsZero() {
		filter["_id"] = bson.M{"$ne": excludeID}
	}

	count, err := s.db.Collection("leave_request").CountDocuments(ctx, filter)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("sudah ada pengajuan pada tanggal tersebut")
	}

	return nil
}
