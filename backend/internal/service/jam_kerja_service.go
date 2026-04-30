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
	ListJamKerjaMyDepartment(ctx context.Context, departmentName, q, position string) ([]models.JamKerjaListRowResponse, error)
	GetJamKerjaByUserID(ctx context.Context, userID string) (*models.JamKerjaDetailResponse, error)
	UpdateJamKerjaByUserID(ctx context.Context, userID string, req models.UpdateJamKerjaRequest) (*models.JamKerjaDetailResponse, error)
	UpdateJamKerjaByUserIDForManager(ctx context.Context, actorRole, actorDepartment, userID string, req models.UpdateJamKerjaRequest) (*models.JamKerjaDetailResponse, error)
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

var hariOrder = []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"}

func formatHariKerja(hari []string) string {
	set := map[string]bool{}
	for _, h := range hari {
		set[h] = true
	}

	out := make([]string, 0, 7)
	for _, h := range hariOrder {
		if set[h] {
			out = append(out, h)
		}
	}
	if len(out) == 0 {
		return "-"
	}
	if len(out) == 7 {
		return "Senin - Minggu"
	}
	return strings.Join(out, ", ")
}

var validHari = map[string]bool{
	"Senin": true, "Selasa": true, "Rabu": true, "Kamis": true, "Jumat": true, "Sabtu": true, "Minggu": true,
}

var hhmmRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

func parseHHMMToToday(v string) (time.Time, error) {
	if !hhmmRe.MatchString(v) {
		return time.Time{}, errors.New("format waktu harus HH:mm")
	}

	now := time.Now().UTC()
	y, m, d := now.Date()

	h := int((v[0]-'0')*10 + (v[1] - '0'))
	min := int((v[3]-'0')*10 + (v[4] - '0'))

	return time.Date(y, m, d, h, min, 0, 0, time.UTC), nil
}

func defaultJamKerja() *models.JamKerja {
	now := time.Now().UTC()
	y, m, d := now.Date()

	return &models.JamKerja{
		DayOfWeek: []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat"},
		StartTime: time.Date(y, m, d, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(y, m, d, 18, 0, 0, 0, time.UTC),
		IsActive:  true,
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
	users, err := s.userRepo.FindAll(ctx)
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
			out = append(out, models.JamKerjaListRowResponse{
				ID:         u.ID.Hex(),
				Name:       u.FullName,
				NIK:        u.PayrollNumber,
				Avatar:     u.Avatar,
				Department: u.DepartmentName,
				Position:   u.PositionName,
				DayOfWeek:  []string{},
				WorkDays:   "Belum Diatur",
				StartTime:  "-",
				EndTime:    "-",
			})
		} else {
			out = append(out, models.JamKerjaListRowResponse{
				ID:         u.ID.Hex(),
				Name:       u.FullName,
				NIK:        u.PayrollNumber,
				Avatar:     u.Avatar,
				Department: u.DepartmentName,
				Position:   u.PositionName,
				DayOfWeek:  jk.DayOfWeek,                    // ✅ map
				WorkDays:   formatHariKerja(jk.DayOfWeek),   // ✅ map
				StartTime:  jk.StartTime.UTC().Format("15:04"), // ✅ map + UTC
				EndTime:    jk.EndTime.UTC().Format("15:04"),   // ✅ map + UTC
			})
		}
	}

	return out, nil
}

func (s *jamKerjaService) ListJamKerjaMyDepartment(ctx context.Context, departmentName, q, position string) ([]models.JamKerjaListRowResponse, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	query := strings.ToLower(strings.TrimSpace(q))
	posFilter := strings.TrimSpace(position)
	dept := strings.TrimSpace(departmentName)

	out := make([]models.JamKerjaListRowResponse, 0)
	for _, u := range users {
		if strings.TrimSpace(u.DepartmentName) != dept {
			continue
		}
		if posFilter != "" && posFilter != "all" {
			if strings.TrimSpace(u.PositionName) != posFilter {
				continue
			}
		}
		if query != "" {
			name := strings.ToLower(u.FullName)
			nik := strings.ToLower(u.PayrollNumber)
			if !strings.Contains(name, query) && !strings.Contains(nik, query) {
				continue
			}
		}

		jk, err := s.jamKerjaRepo.FindByUserID(ctx, u.ID.Hex())
		if err != nil {
			return nil, err
		}
		if jk == nil {
			out = append(out, models.JamKerjaListRowResponse{
				ID:         u.ID.Hex(),
				Name:       u.FullName,
				NIK:        u.PayrollNumber,
				Avatar:     u.Avatar,
				Department: u.DepartmentName,
				Position:   u.PositionName,
				DayOfWeek:  []string{},
				WorkDays:   "Belum Diatur",
				StartTime:  "-",
				EndTime:    "-",
			})
		} else {
			out = append(out, models.JamKerjaListRowResponse{
				ID:         u.ID.Hex(),
				Name:       u.FullName,
				NIK:        u.PayrollNumber,
				Avatar:     u.Avatar,
				Department: u.DepartmentName,
				Position:   u.PositionName,
				DayOfWeek:  jk.DayOfWeek,
				WorkDays:   string(workDaysLabelFromHari(jk.DayOfWeek)),
				StartTime:  jk.StartTime.UTC().Format("15:04"),
				EndTime:    jk.EndTime.UTC().Format("15:04"),
			})
		}
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
		return &models.JamKerjaDetailResponse{
			UserID:       u.ID.Hex(),
			Name:         u.FullName,
			NIK:          u.PayrollNumber,
			Department:   u.DepartmentName,
			Position:     u.PositionName,
			DayOfWeek:    []string{},
			StartTime:    "",
			EndTime:      "",
			IsActive:     false,
		}, nil
	}

	return &models.JamKerjaDetailResponse{
		UserID:       u.ID.Hex(),
		Name:         u.FullName,
		NIK:          u.PayrollNumber,
		Department:   u.DepartmentName,
		Position:     u.PositionName,
		DayOfWeek:    jk.DayOfWeek,                    // ✅ map
		StartTime:   jk.StartTime.UTC().Format("15:04"), // ✅ map + UTC
		EndTime:     jk.EndTime.UTC().Format("15:04"),   // ✅ map + UTC
		IsActive:     jk.IsActive,                     // ✅ map
	}, nil
}

func (s *jamKerjaService) UpdateJamKerjaByUserID(ctx context.Context, userID string, req models.UpdateJamKerjaRequest) (*models.JamKerjaDetailResponse, error) {
	_, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	if len(req.DayOfWeek) == 0 {
		return nil, errors.New("hari kerja wajib diisi")
	}
	for _, h := range req.DayOfWeek {
		hh := strings.TrimSpace(h)
		if !validHari[hh] {
			return nil, errors.New("hari kerja tidak valid")
		}
	}

	startT, err := parseHHMMToToday(req.StartTime)
	if err != nil {
		return nil, err
	}
	endT, err := parseHHMMToToday(req.EndTime)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("karyawan tidak ditemukan")
	}

	aktif := true
	if req.IsActive != nil {
		aktif = *req.IsActive
	}

	jk := &models.JamKerja{
		DayOfWeek:    req.DayOfWeek,
		StartTime:   startT,
		EndTime:     endT,
		IsActive:     aktif,
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
		DayOfWeek:    req.DayOfWeek,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsActive:     aktif,
	}, nil
}

func (s *jamKerjaService) UpdateJamKerjaByUserIDForManager(ctx context.Context, actorRole, actorDepartment, userID string, req models.UpdateJamKerjaRequest) (*models.JamKerjaDetailResponse, error) {
	actorRole = strings.TrimSpace(actorRole)
	actorDepartment = strings.TrimSpace(actorDepartment)

	if actorRole == models.RoleManagerDepartemen {
		u, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return nil, errors.New("karyawan tidak ditemukan")
		}
		if strings.TrimSpace(u.DepartmentName) != actorDepartment {
			return nil, errors.New("akses ditolak: hanya dapat mengatur jam kerja departemen Anda")
		}
	}

	return s.UpdateJamKerjaByUserID(ctx, userID, req)
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

	if len(req.DayOfWeek) == 0 {
		return nil, errors.New("hari kerja wajib diisi")
	}
	for _, h := range req.DayOfWeek {
		hh := strings.TrimSpace(h)
		if !validHari[hh] {
			return nil, errors.New("hari kerja tidak valid")
		}
	}

	startT, err := parseHHMMToToday(req.StartTime)
	if err != nil {
		return nil, err
	}
	endT, err := parseHHMMToToday(req.EndTime)
	if err != nil {
		return nil, err
	}

	aktif := true
	if req.IsActive != nil {
		aktif = *req.IsActive
	}

	userOID, _ := primitive.ObjectIDFromHex(req.UserID)

	jk := &models.JamKerja{
		UserID:       userOID,
		DayOfWeek:    req.DayOfWeek,
		StartTime:   startT,
		EndTime:     endT,
		IsActive:     aktif,
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
		DayOfWeek:    req.DayOfWeek,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsActive:     aktif,
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
