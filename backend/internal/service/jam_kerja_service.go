// internal/service/jam_kerja_service.go
package service

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JamKerjaService interface {
	ListJamKerja(ctx context.Context) ([]models.JamKerjaListRowResponse, error)
	GetJamKerjaByUserID(ctx context.Context, userID string) (*models.JamKerjaDetailResponse, error)
	UpdateJamKerjaByUserID(ctx context.Context, userID string, req models.UpdateJamKerjaRequest) (*models.JamKerjaDetailResponse, error)
	CreateJamKerja(ctx context.Context, req models.CreateJamKerjaRequest) (*models.JamKerjaDetailResponse, error)
	SearchAvailableEmployees(ctx context.Context, q string) ([]models.AvailableEmployeeResponse, error)
}

type jamKerjaService struct {
	jamKerjaRepo repository.JamKerjaRepository
	userRepo     repository.UserRepository
}

func NewJamKerjaService(
	jamKerjaRepo repository.JamKerjaRepository,
	userRepo repository.UserRepository,
) JamKerjaService {
	return &jamKerjaService{
		jamKerjaRepo: jamKerjaRepo,
		userRepo:     userRepo,
	}
}

// ====== DTO untuk UI ======
/*
UI butuh:
id (user id untuk tombol Atur Jam Kerja),
name, nik, department, position,
workDays (badge: Senin - Jumat / Senin - Sabtu / Shift), startTime, endTime
*/

type WorkDaysLabel string

const (
	WorkDaysSeninJumat WorkDaysLabel = "Senin - Jumat"
	WorkDaysSeninSabtu WorkDaysLabel = "Senin - Sabtu"
	WorkDaysShift      WorkDaysLabel = "Shift"
)

var validHari = map[string]bool{
	"Senin": true, "Selasa": true, "Rabu": true, "Kamis": true, "Jumat": true, "Sabtu": true, "Minggu": true,
}

var hhmmRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

func parseHHMMToToday(v string) (time.Time, error) {
	if !hhmmRe.MatchString(v) {
		return time.Time{}, errors.New("format waktu harus HH:mm")
	}

	now := time.Now()
	y, m, d := now.Date()
	loc := now.Location()
	if loc == nil {
		loc = time.Local
	}

	h := int((v[0]-'0')*10 + (v[1] - '0'))
	min := int((v[3]-'0')*10 + (v[4] - '0'))

	return time.Date(y, m, d, h, min, 0, 0, loc), nil
}

func defaultJamKerja() *models.JamKerja {
	now := time.Now()
	y, m, d := now.Date()
	loc := now.Location()
	if loc == nil {
		loc = time.Local
	}

	return &models.JamKerja{
		HariKerja:    []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat"},
		WaktuMulai:   time.Date(y, m, d, 9, 0, 0, 0, loc),
		WaktuSelesai: time.Date(y, m, d, 18, 0, 0, 0, loc),
		Aktif:        true,
	}
}

func workDaysLabelFromHari(hari []string) WorkDaysLabel {
	// urutkan agar stabil
	h := append([]string{}, hari...)
	sort.Strings(h)

	// basic mapping untuk badge seperti UI Anda
	isMonFri := len(h) == 5 &&
		containsAll(h, []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat"})
	if isMonFri {
		return WorkDaysSeninJumat
	}

	isMonSat := len(h) == 6 &&
		containsAll(h, []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"})
	if isMonSat {
		return WorkDaysSeninSabtu
	}

	return WorkDaysShift
}

func containsAll(h []string, needed []string) bool {
	set := map[string]bool{}
	for _, x := range h {
		set[x] = true
	}
	for _, n := range needed {
		if !set[n] {
			return false
		}
	}
	return true
}

// ====== Models request/response (buat file terpisah di pkg/models jika mau) ======

func (s *jamKerjaService) ListJamKerja(ctx context.Context) ([]models.JamKerjaListRowResponse, error) {
	users, err := s.userRepo.FindAll(ctx) // sesuaikan jika Anda ingin hanya staff
	if err != nil {
		return nil, err
	}

	out := make([]models.JamKerjaListRowResponse, 0, len(users))
	for _, u := range users {
		jk, err := s.jamKerjaRepo.FindByUserID(ctx, u.ID.Hex())
		if err != nil {
			return nil, err
		}

		if jk == nil {
			jk = defaultJamKerja()
		}

		out = append(out, models.JamKerjaListRowResponse{
			ID:         u.ID.Hex(),
			Name:       u.FullName,
			NIK:        u.PayrollNumber,
			Department: u.DepartmentName,
			Position:   u.PositionName,
			WorkDays:   string(workDaysLabelFromHari(jk.HariKerja)),
			StartTime:  jk.WaktuMulai.Format("15:04"),
			EndTime:    jk.WaktuSelesai.Format("15:04"),
		})
	}

	return out, nil
}

func (s *jamKerjaService) GetJamKerjaByUserID(ctx context.Context, userID string) (*models.JamKerjaDetailResponse, error) {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	jk, err := s.jamKerjaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if jk == nil {
		jk = defaultJamKerja()
	}

	return &models.JamKerjaDetailResponse{
		UserID:       u.ID.Hex(),
		Name:         u.FullName,
		NIK:          u.PayrollNumber,
		Department:   u.DepartmentName,
		Position:     u.PositionName,
		HariKerja:    jk.HariKerja,
		WaktuMulai:   jk.WaktuMulai.Format("15:04"),
		WaktuSelesai: jk.WaktuSelesai.Format("15:04"),
		Aktif:        jk.Aktif,
	}, nil
}

func (s *jamKerjaService) UpdateJamKerjaByUserID(ctx context.Context, userID string, req models.UpdateJamKerjaRequest) (*models.JamKerjaDetailResponse, error) {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	if len(req.HariKerja) == 0 {
		return nil, errors.New("hari kerja wajib diisi")
	}
	for _, h := range req.HariKerja {
		hh := strings.TrimSpace(h)
		if !validHari[hh] {
			return nil, errors.New("hari kerja tidak valid")
		}
	}

	startT, err := parseHHMMToToday(req.WaktuMulai)
	if err != nil {
		return nil, err
	}
	endT, err := parseHHMMToToday(req.WaktuSelesai)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	aktif := true
	if req.Aktif != nil {
		aktif = *req.Aktif
	}

	jk := &models.JamKerja{
		HariKerja:    req.HariKerja,
		WaktuMulai:   startT,
		WaktuSelesai: endT,
		Aktif:        aktif,
	}

	if err := s.jamKerjaRepo.UpsertByUserID(ctx, userID, jk); err != nil {
		return nil, err
	}

	return &models.JamKerjaDetailResponse{
		UserID:       u.ID.Hex(),
		Name:         u.FullName,
		NIK:          u.PayrollNumber,
		Department:   u.DepartmentName,
		Position:     u.PositionName,
		HariKerja:    req.HariKerja,
		WaktuMulai:   req.WaktuMulai,
		WaktuSelesai: req.WaktuSelesai,
		Aktif:        aktif,
	}, nil
}

func (s *jamKerjaService) CreateJamKerja(ctx context.Context, req models.CreateJamKerjaRequest) (*models.JamKerjaDetailResponse, error) {
	_, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	// pastikan user ada
	u, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	// cek apakah sudah ada jam kerja untuk user ini (karena unique)
	existing, err := s.jamKerjaRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("jam kerja untuk karyawan ini sudah ada, silakan gunakan fitur atur/update")
	}

	if len(req.HariKerja) == 0 {
		return nil, errors.New("hari kerja wajib diisi")
	}
	for _, h := range req.HariKerja {
		hh := strings.TrimSpace(h)
		if !validHari[hh] {
			return nil, errors.New("hari kerja tidak valid")
		}
	}

	startT, err := parseHHMMToToday(req.WaktuMulai)
	if err != nil {
		return nil, err
	}
	endT, err := parseHHMMToToday(req.WaktuSelesai)
	if err != nil {
		return nil, err
	}

	aktif := true
	if req.Aktif != nil {
		aktif = *req.Aktif
	}

	userOID, _ := primitive.ObjectIDFromHex(req.UserID)

	jk := &models.JamKerja{
		UserID:       userOID,
		HariKerja:    req.HariKerja,
		WaktuMulai:   startT,
		WaktuSelesai: endT,
		Aktif:        aktif,
	}

	if err := s.jamKerjaRepo.Create(ctx, jk); err != nil {
		return nil, err
	}

	return &models.JamKerjaDetailResponse{
		UserID:       u.ID.Hex(),
		Name:         u.FullName,
		NIK:          u.PayrollNumber,
		Department:   u.DepartmentName,
		Position:     u.PositionName,
		HariKerja:    req.HariKerja,
		WaktuMulai:   req.WaktuMulai,
		WaktuSelesai: req.WaktuSelesai,
		Aktif:        aktif,
	}, nil
}

func (s *jamKerjaService) SearchAvailableEmployees(ctx context.Context, q string) ([]models.AvailableEmployeeResponse, error) {
	usedIDs, err := s.jamKerjaRepo.GetAllUserIDs(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.FindActiveExcludeIDsWithSearch(ctx, usedIDs, q)
	if err != nil {
		return nil, err
	}

	out := make([]models.AvailableEmployeeResponse, 0, len(users))
	for _, u := range users {
		out = append(out, models.AvailableEmployeeResponse{
			ID:         u.ID.Hex(),
			FullName:   u.FullName,
			NIK:        u.PayrollNumber,
			Department: u.DepartmentName,
			Position:   u.PositionName,
		})
	}
	return out, nil
}