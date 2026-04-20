// internal/service/attendance_service.go
package service

import (
	"bytes"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/faceclient"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var wib = time.FixedZone("WIB", 7*60*60)

type AttendanceService interface {
	ClockIn(ctx context.Context, userID string, latitude, longitude float64, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error)
	ClockOut(ctx context.Context, userID string, latitude, longitude float64, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error)
	GetTodayAttendance(ctx context.Context, userID string) (*models.Attendance, error)
	GetMonthlyAttendance(ctx context.Context, userID string, month, year int) (*models.MonthlyAttendanceResponse, error)
	ProcessAttendanceWithFace(ctx context.Context, userID string, photo []byte, filename string, latitude, longitude float64, recordType string, verifyOnly bool) (*AttendanceProcessResult, error)
	ValidateClockInWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error)
	ValidateClockOutWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error)
	GetScheduleInfo(ctx context.Context, userID string) (*ScheduleInfoResponse, error)
	GetWorkScheduleInfo(ctx context.Context, userID string) (*WorkScheduleInfo, error)
	GetManagerAttendance(ctx context.Context, from, to time.Time, departmentName, q string, page, pageSize int64) (*models.ManagerAttendanceListResponse, error)
	ExportManagerAttendanceCSV(ctx context.Context, from, to time.Time, departmentName, q string) ([]byte, string, error)
	ExportManagerAttendanceCSVStream(ctx context.Context, from, to time.Time, departmentName, q string) (io.ReadCloser, string, error)
}

type WorkScheduleInfo struct {
	UserID        string         `json:"user_id"`
	HariKerja     []string       `json:"hari_kerja"`
	WaktuMulai    string         `json:"waktu_mulai"`   // HH:mm
	WaktuSelesai  string         `json:"waktu_selesai"` // HH:mm
	Aktif         bool           `json:"aktif"`
	TodaySchedule *TodaySchedule `json:"today_schedule,omitempty"`
}

type TodaySchedule struct {
	IsWorkDay      bool   `json:"is_work_day"`
	ClockInWindow  string `json:"clock_in_window"`  // HH:mm - HH:mm
	ClockOutWindow string `json:"clock_out_window"` // HH:mm onwards
	CanClockIn     bool   `json:"can_clock_in"`
	CanClockOut    bool   `json:"can_clock_out"`
	Message        string `json:"message"`
}

type attendanceService struct {
	attendanceRepo    repository.AttendanceRepository
	userRepo          repository.UserRepository
	faceEmbeddingRepo repository.FaceEmbeddingRepository
	jamKerjaRepo      repository.JamKerjaRepository
	faceClient        *faceclient.Client
	officeLat         float64
	officeLng         float64
	radiusMeters      float64
}

type AttendanceProcessResult struct {
	Success           bool               `json:"success"`
	Message           string             `json:"message"`
	FaceSimilarity    float64            `json:"face_similarity"`
	LocationValid     bool               `json:"location_valid"`
	Distance          float64            `json:"distance_m"`
	Attendance        *models.Attendance `json:"attendance,omitempty"`
	IsClockInAllowed  bool               `json:"is_clock_in_allowed,omitempty"`
	IsClockOutAllowed bool               `json:"is_clock_out_allowed,omitempty"`
	ClockInWindow     string             `json:"clock_in_window,omitempty"`
	ClockOutWindow    string             `json:"clock_out_window,omitempty"`
	NextWindowOpen    string             `json:"next_window_open,omitempty"`
	WorkScheduleFound bool               `json:"work_schedule_found,omitempty"`
}

func NewAttendanceService(
	attendanceRepo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	faceEmbeddingRepo repository.FaceEmbeddingRepository,
	jamKerjaRepo repository.JamKerjaRepository,
	faceClient *faceclient.Client,
) AttendanceService {
	return &attendanceService{
		attendanceRepo:    attendanceRepo,
		userRepo:          userRepo,
		faceEmbeddingRepo: faceEmbeddingRepo,
		jamKerjaRepo:      jamKerjaRepo,
		faceClient:        faceClient,
		officeLat:         2.3561,
		officeLng:         99.1431,
		radiusMeters:      10000,
	}
}

func (s *attendanceService) GetManagerAttendance(ctx context.Context, from, to time.Time, departmentName, q string, page, pageSize int64) (*models.ManagerAttendanceListResponse, error) {
	rows, total, statusCounts, err := s.attendanceRepo.FindManagerAttendance(ctx, from, to, departmentName, q, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]models.ManagerAttendanceRecord, 0, len(rows))
	for _, r := range rows {
		clockIn := "--:--"
		if r.ClockInTime != nil {
			clockIn = r.ClockInTime.In(wib).Format("15:04")
		}
		clockOut := "--:--"
		if r.ClockOutTime != nil {
			clockOut = r.ClockOutTime.In(wib).Format("15:04")
		}

		location := formatGeoLocation(r.ClockInLocation)

		items = append(items, models.ManagerAttendanceRecord{
			ID:             r.ID.Hex(),
			UserID:         r.UserID.Hex(),
			FullName:       r.User.FullName,
			Email:          r.User.Email,
			PayrollNumber:  r.User.PayrollNumber,
			DepartmentName: r.User.DepartmentName,
			PositionName:   r.User.PositionName,
			Date:           r.Date.In(wib).Format("2006-01-02"),
			ClockInTime:    clockIn,
			ClockOutTime:   clockOut,
			Status:         mapAttendanceStatusToUI(r.Status),
			Location:       location,
		})
	}

	summary := buildManagerAttendanceSummary(statusCounts)
	summary.TotalRecords = total

	return &models.ManagerAttendanceListResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Summary:  summary,
	}, nil
}

func (s *attendanceService) ExportManagerAttendanceCSV(ctx context.Context, from, to time.Time, departmentName, q string) ([]byte, string, error) {
	rows, err := s.attendanceRepo.FindManagerAttendanceExport(ctx, from, to, departmentName, q)
	if err != nil {
		return nil, "", err
	}

	buf := &bytes.Buffer{}
	buf.WriteString("date,payroll_number,full_name,email,department,clock_in,clock_out,status,location\n")
	for _, r := range rows {
		dateValue := r.DateStr
		if dateValue == "" {
			dateValue = r.Date.In(wib).Format("02/01/2006")
		}
		clockIn := ""
		if r.ClockInTime != nil {
			clockIn = r.ClockInTime.In(wib).Format("15:04")
		}
		clockOut := ""
		if r.ClockOutTime != nil {
			clockOut = r.ClockOutTime.In(wib).Format("15:04")
		}
		location := formatGeoLocation(r.ClockInLocation)

		line := fmt.Sprintf(
			"%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			dateValue,
			escapeCSV(r.User.PayrollNumber),
			escapeCSV(r.User.FullName),
			escapeCSV(r.User.Email),
			escapeCSV(r.User.DepartmentName),
			escapeCSV(clockIn),
			escapeCSV(clockOut),
			escapeCSV(mapAttendanceStatusToUI(r.Status)),
			escapeCSV(location),
		)
		buf.WriteString(line)
	}

	filename := fmt.Sprintf("presensi_%s_%s.csv", from.In(wib).Format("20060102"), to.In(wib).Add(-time.Nanosecond).Format("20060102"))
	return buf.Bytes(), filename, nil
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\n\r\"") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}

func mapAttendanceStatusToUI(status models.AttendanceStatus) string {
	switch status {
	case models.StatusOnTime:
		return "HADIR"
	case models.StatusLate:
		return "TELAT"
	case models.StatusOvertime:
		return "HADIR"
	case models.StatusAbsent:
		return "ALFA"
	default:
		return "ALFA"
	}
}

func formatGeoLocation(loc models.GeoLocation) string {
	if loc.Latitude != 0 || loc.Longitude != 0 {
		return fmt.Sprintf("%.5f, %.5f", loc.Latitude, loc.Longitude)
	}
	return "Unrecorded"
}
func (s *attendanceService) ExportManagerAttendanceCSVStream(ctx context.Context, from, to time.Time, departmentName, q string) (io.ReadCloser, string, error) {
	pr, pw := io.Pipe()
	filename := fmt.Sprintf("presensi_%s_%s.csv", from.In(wib).Format("20060102"), to.In(wib).Add(-time.Nanosecond).Format("20060102"))

	go func() {
		bw := bufio.NewWriterSize(pw, 64*1024)
		_, writeErr := bw.WriteString("date,payroll_number,full_name,email,department,clock_in,clock_out,status,location\n")
		if writeErr != nil {
			_ = pw.CloseWithError(writeErr)
			return
		}
		if err := bw.Flush(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		cursor, err := s.attendanceRepo.FindManagerAttendanceExportCursor(ctx, from, to, departmentName, q)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var r models.ManagerAttendanceAggRow
			if err := cursor.Decode(&r); err != nil {
				_ = pw.CloseWithError(err)
				return
			}

			dateValue := r.DateStr
			if dateValue == "" {
				dateValue = r.Date.In(wib).Format("02/01/2006")
			}
			clockIn := ""
			if r.ClockInTime != nil {
				clockIn = r.ClockInTime.In(wib).Format("15:04")
			}
			clockOut := ""
			if r.ClockOutTime != nil {
				clockOut = r.ClockOutTime.In(wib).Format("15:04")
			}
			location := formatGeoLocation(r.ClockInLocation)

			line := fmt.Sprintf(
				"%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
				dateValue,
				escapeCSV(r.User.PayrollNumber),
				escapeCSV(r.User.FullName),
				escapeCSV(r.User.Email),
				escapeCSV(r.User.DepartmentName),
				escapeCSV(clockIn),
				escapeCSV(clockOut),
				escapeCSV(mapAttendanceStatusToUI(r.Status)),
				escapeCSV(location),
			)

			if _, err := bw.WriteString(line); err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}

		if err := cursor.Err(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		if err := bw.Flush(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_ = pw.Close()
	}()

	return pr, filename, nil
}


func buildManagerAttendanceSummary(statusCounts map[string]int64) models.ManagerAttendanceSummary {
	tepatWaktu := statusCounts[string(models.StatusOnTime)]
	terlambat := statusCounts[string(models.StatusLate)]
	alfa := statusCounts[string(models.StatusAbsent)]
	izin := int64(0)
	denom := tepatWaktu + terlambat + izin + alfa
	kehadiran := tepatWaktu + terlambat
	pct := float64(0)
	if denom > 0 {
		pct = (float64(kehadiran) / float64(denom)) * 100
	}
	return models.ManagerAttendanceSummary{
		TepatWaktu:        tepatWaktu,
		Terlambat:         terlambat,
		IzinSakit:         izin,
		Alfa:              alfa,
		TotalKehadiranPct: math.Round(pct*10) / 10,
	}
}

// ✅ Get work schedule info untuk dashboard
func (s *attendanceService) GetWorkScheduleInfo(ctx context.Context, userID string) (*WorkScheduleInfo, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// ✅ Cari jam kerja berdasarkan user ID
	jamKerja, err := s.jamKerjaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if jamKerja == nil {
		jamKerja = s.getDefaultJamKerja(userObjID)
	}

	// ✅ Format waktu dari time.Time ke string "HH:mm"
	waktuMulaiStr := jamKerja.StartTime.Format("15:04")
	waktuSelesaiStr := jamKerja.EndTime.Format("15:04")

	info := &WorkScheduleInfo{
		UserID:       userID,
		HariKerja:    jamKerja.DayOfWeek,
		WaktuMulai:   waktuMulaiStr,
		WaktuSelesai: waktuSelesaiStr,
		Aktif:        jamKerja.IsActive,
	}

	// Ambil info hari ini
	nowWIB := time.Now().In(wib)
	dayName := s.getDayName(nowWIB.Weekday())

	isWorkDay := false
	for _, day := range jamKerja.DayOfWeek {
		if day == dayName {
			isWorkDay = true
			break
		}
	}

	startTimeToday := s.extractTimeForToday(nowWIB, jamKerja.StartTime)
	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
	clockInWindowOpen := startTimeToday.Add(-15 * time.Minute)
	clockInWindowClose := startTimeToday

	canClockIn := isWorkDay && !nowWIB.Before(clockInWindowOpen) && !nowWIB.After(clockInWindowClose)
	canClockOut := isWorkDay && !nowWIB.Before(endTimeToday)

	message := ""
	if !isWorkDay {
		message = "Hari ini bukan hari kerja"
	} else if !canClockIn {
		if nowWIB.Before(clockInWindowOpen) {
			message = "Clock in dibuka pada " + clockInWindowOpen.Format("15:04") + " WIB"
		} else {
			message = "Clock in sudah ditutup. Buka kembali besok jam " + clockInWindowOpen.Format("15:04") + " WIB"
		}
	}

	info.TodaySchedule = &TodaySchedule{
		IsWorkDay:      isWorkDay,
		ClockInWindow:  clockInWindowOpen.Format("15:04") + " - " + clockInWindowClose.Format("15:04"),
		ClockOutWindow: endTimeToday.Format("15:04") + " WIB onwards",
		CanClockIn:     canClockIn,
		CanClockOut:    canClockOut,
		Message:        message,
	}

	return info, nil
}

// ✅ Validasi window clock in
func (s *attendanceService) ValidateClockInWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, nil, errors.New("invalid user ID format")
	}

	// ✅ Cari jam kerja berdasarkan user ID
	jamKerja, err := s.jamKerjaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	if jamKerja == nil {
		jamKerja = s.getDefaultJamKerja(userObjID)
	}

	if !jamKerja.IsActive {
		return false, jamKerja, errors.New("jadwal kerja tidak aktif")
	}

	nowWIB := time.Now().In(wib)
	dayName := s.getDayName(nowWIB.Weekday())

	isWorkDay := false
	for _, day := range jamKerja.DayOfWeek {
		if day == dayName {
			isWorkDay = true
			break
		}
	}

	if !isWorkDay {
		return false, jamKerja, errors.New("hari ini bukan hari kerja")
	}

	startTimeToday := s.extractTimeForToday(nowWIB, jamKerja.StartTime)
	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
	clockInWindowOpen := startTimeToday.Add(-15 * time.Minute)

	isInWindow := !nowWIB.Before(clockInWindowOpen) && !nowWIB.After(endTimeToday)

	return isInWindow, jamKerja, nil
}

// ✅ Validasi window clock out
func (s *attendanceService) ValidateClockOutWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, nil, errors.New("invalid user ID format")
	}

	// ✅ Cari jam kerja berdasarkan user ID
	jamKerja, err := s.jamKerjaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	if jamKerja == nil {
		jamKerja = s.getDefaultJamKerja(userObjID)
	}

	if !jamKerja.IsActive {
		return false, jamKerja, errors.New("jadwal kerja tidak aktif")
	}

	nowWIB := time.Now().In(wib)
	dayName := s.getDayName(nowWIB.Weekday())

	isWorkDay := false
	for _, day := range jamKerja.DayOfWeek {
		if day == dayName {
			isWorkDay = true
			break
		}
	}

	if !isWorkDay {
		return false, jamKerja, errors.New("hari ini bukan hari kerja")
	}

	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
	clockOutWindowClose := endTimeToday.Add(30 * time.Minute)

	isInWindow := !nowWIB.Before(endTimeToday) && !nowWIB.After(clockOutWindowClose)

	return isInWindow, jamKerja, nil
}

func (s *attendanceService) ClockIn(ctx context.Context, userID string, latitude, longitude float64, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Cek sudah clock in hari ini
	existing, _ := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if existing != nil && existing.ClockInTime != nil {
		return nil, errors.New("sudah melakukan clock in hari ini")
	}

	// ✅ Validasi window clock in
	isInWindow, jamKerja, err := s.ValidateClockInWindow(ctx, userID)
	if err != nil {
		return nil, errors.New("tidak dapat clock in: " + err.Error())
	}

	if !isInWindow {
		nowWIB := time.Now().In(wib)
		startTimeToday := s.extractTimeForToday(nowWIB, jamKerja.StartTime)
		windowOpen := startTimeToday.Add(-15 * time.Minute)
		endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)

		message := ""
		if nowWIB.Before(windowOpen) {
			message = "jendela clock in belum dibuka (buka pada " + windowOpen.Format("15:04") + " WIB)"
		} else if nowWIB.After(endTimeToday) {
			message = "jendela clock in sudah tutup (buka kembali besok jam " + windowOpen.Format("15:04") + " WIB)"
		}
		return nil, errors.New(message)
	}

	now := time.Now().In(wib)

	location := models.GeoLocation{
		Latitude:  latitude,
		Longitude: longitude,
	}

	// Tentukan status berdasarkan waktu saat submit dengan toleransi keterlambatan 1 menit.
	startTimeToday := s.extractTimeForToday(now, jamKerja.StartTime)
	lateThreshold := startTimeToday.Add(1 * time.Minute)

	status := models.StatusOnTime
	if now.After(lateThreshold) {
		status = models.StatusLate
	}

	fmt.Printf("⏰ Clock In Status Determination:\n  Start Time: %s\n  Late Threshold(+1m): %s\n  Submit Time: %s\n  Status: %s\n",
		startTimeToday.Format("15:04:05"), lateThreshold.Format("15:04:05"), now.Format("15:04:05"), status)

	attendance := &models.Attendance{
		ID:              primitive.NewObjectID(),
		UserID:          userObjID,
		Date:            now,
		ClockInTime:     &now,
		ClockInPhoto:    filename,
		ClockInLocation: location,
		Status:          status,
		WorkHours:       0,
		OvertimeHours:   0,
		FaceSimilarity:  faceSimilarity,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err = s.attendanceRepo.Create(ctx, attendance)
	if err != nil {
		return nil, err
	}

	return attendance, nil
}

func (s *attendanceService) GetScheduleInfo(ctx context.Context, userID string) (*ScheduleInfoResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// ✅ Cari jam kerja berdasarkan user ID
	jamKerja, err := s.jamKerjaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if jamKerja == nil {
		jamKerja = s.getDefaultJamKerja(userObjID)
	}

	// ✅ Format waktu dari time.Time ke string "HH:mm"
	waktuMulaiStr := jamKerja.StartTime.Format("15:04")
	waktuSelesaiStr := jamKerja.EndTime.Format("15:04")

	info := &ScheduleInfoResponse{
		UserID:       userID,
		HariKerja:    jamKerja.DayOfWeek,
		WaktuMulai:   waktuMulaiStr,
		WaktuSelesai: waktuSelesaiStr,
		Aktif:        jamKerja.IsActive,
	}

	nowWIB := time.Now().In(wib)
	dayName := s.getDayName(nowWIB.Weekday())

	isWorkDay := false
	for _, day := range jamKerja.DayOfWeek {
		if day == dayName {
			isWorkDay = true
			break
		}
	}

	startTimeToday := s.extractTimeForToday(nowWIB, jamKerja.StartTime)
	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
	clockInWindowOpen := startTimeToday.Add(-15 * time.Minute)

	// ✅ Clock out window: dari waktu_selesai hingga +30 menit
	clockOutWindowClose := endTimeToday.Add(30 * time.Minute)

	canClockIn := isWorkDay && !nowWIB.Before(clockInWindowOpen) && !nowWIB.After(endTimeToday)
	canClockOut := isWorkDay && !nowWIB.Before(endTimeToday) && !nowWIB.After(clockOutWindowClose)

	message := ""
	if !isWorkDay {
		message = "Hari ini bukan hari kerja"
	} else if !canClockIn && !canClockOut {
		if nowWIB.Before(clockInWindowOpen) {
			message = "Clock in dibuka pada " + clockInWindowOpen.Format("15:04") + " WIB"
		} else if nowWIB.After(endTimeToday) && nowWIB.After(clockOutWindowClose) {
			message = "CLOCK IN"
		}
	}

	info.TodaySchedule = &TodayScheduleInfoResponse{
		IsWorkDay:      isWorkDay,
		ClockInWindow:  clockInWindowOpen.Format("15:04") + " - " + endTimeToday.Format("15:04"),
		ClockOutWindow: endTimeToday.Format("15:04") + " - " + clockOutWindowClose.Format("15:04"),
		CanClockIn:     canClockIn,
		CanClockOut:    canClockOut,
		Message:        message,
	}

	return info, nil
}

func (s *attendanceService) ClockOut(ctx context.Context, userID string, latitude, longitude float64, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Cek sudah clock out hari ini
	existing, _ := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if existing != nil && existing.ClockOutTime != nil {
		return nil, errors.New("sudah melakukan clock out hari ini")
	}

	// ✅ Validasi window clock out
	isInWindow, jamKerja, err := s.ValidateClockOutWindow(ctx, userID)
	if err != nil {
		return nil, errors.New("tidak dapat clock out: " + err.Error())
	}

	if !isInWindow {
		nowWIB := time.Now().In(wib)
		endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
		clockOutWindowClose := endTimeToday.Add(30 * time.Minute)

		message := ""
		if nowWIB.Before(endTimeToday) {
			message = "jendela clock out belum dibuka (buka pada " + endTimeToday.Format("15:04") + " WIB)"
		} else if nowWIB.After(clockOutWindowClose) {
			message = "jendela clock out sudah tutup (buka kembali besok jam " + endTimeToday.Format("15:04") + " WIB)"
		}
		return nil, errors.New(message)
	}

	// Cek sudah ada clock in
	if existing == nil || existing.ClockInTime == nil {
		return nil, errors.New("belum melakukan clock in hari ini")
	}

	now := time.Now().In(wib)

	location := models.GeoLocation{
		Latitude:  latitude,
		Longitude: longitude,
	}

	// Calculate work hours
	workDuration := now.Sub(*existing.ClockInTime)
	workHours := workDuration.Hours()

	// Calculate overtime (lebih dari 9 jam)
	overtimeHours := 0.0
	if workHours > 9.0 {
		overtimeHours = workHours - 9.0
	}

	// Update status jika ada overtime
	status := existing.Status
	if overtimeHours > 0 {
		status = models.StatusOvertime
	}

	existing.WorkHours = workHours
	existing.OvertimeHours = overtimeHours
	existing.Status = status
	existing.ClockOutTime = &now
	existing.ClockOutPhoto = filename
	existing.ClockOutLocation = location
	existing.FaceSimilarity = faceSimilarity
	existing.UpdatedAt = now

	// Update in database
	err = s.attendanceRepo.UpdateClockOut(ctx, existing.ID, now, filename, location)
	if err != nil {
		return nil, err
	}

	// Update work hours
	err = s.attendanceRepo.UpdateWorkHours(ctx, existing.ID, existing.WorkHours, existing.OvertimeHours, existing.Status)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *attendanceService) GetTodayAttendance(ctx context.Context, userID string) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	return s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
}

func (s *attendanceService) GetMonthlyAttendance(ctx context.Context, userID string, month, year int) (*models.MonthlyAttendanceResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	return s.attendanceRepo.GetMonthlySummary(ctx, userObjID, year, month)
}

func (s *attendanceService) ProcessAttendanceWithFace(
	ctx context.Context,
	userID string,
	photo []byte,
	filename string,
	latitude, longitude float64,
	recordType string,
	verifyOnly bool,
) (*AttendanceProcessResult, error) {

	// 1. Validate location
	distance := s.calculateDistance(latitude, longitude, s.officeLat, s.officeLng)
	locationValid := distance <= s.radiusMeters

	if !locationValid {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Anda berada di luar area kantor (jarak: " + formatDistance(distance) + "m, max: " + formatDistance(s.radiusMeters) + "m)",
			LocationValid: false,
			Distance:      distance,
		}, nil
	}

	// 2. Get user's face embedding
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil || faceEmbedding == nil {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Wajah belum terdaftar. Silakan daftarkan wajah Anda terlebih dahulu.",
			LocationValid: locationValid,
			Distance:      distance,
		}, nil
	}

	if len(faceEmbedding.FaceEmbedding) == 0 {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Data embedding wajah tidak valid. Silakan daftarkan ulang wajah Anda.",
			LocationValid: locationValid,
			Distance:      distance,
		}, nil
	}

	// 3. Extract embedding dari foto
	currentEmbedding, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Gagal memproses foto: " + err.Error(),
			LocationValid: locationValid,
			Distance:      distance,
		}, nil
	}

	if len(currentEmbedding) == 0 {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Tidak ada wajah terdeteksi dalam foto",
			LocationValid: locationValid,
			Distance:      distance,
		}, nil
	}

	// 4. Calculate similarity
	similarity := s.cosineSimilarity(currentEmbedding, faceEmbedding.FaceEmbedding)

	const threshold = 0.60
	faceValid := similarity >= threshold

	if !faceValid {
		return &AttendanceProcessResult{
			Success:        false,
			Message:        "Wajah tidak cocok dengan data terdaftar (similarity: " + formatFloat(similarity*100) + "%, min: " + formatFloat(threshold*100) + "%)",
			FaceSimilarity: similarity,
			LocationValid:  locationValid,
			Distance:       distance,
		}, nil
	}

	// ✅ Jika verifyOnly=true, HANYA return hasil verifikasi TANPA simpan database
	if verifyOnly {
		fmt.Printf("✅ [VERIFY ONLY] Verifikasi wajah & lokasi berhasil untuk user %s - Menunggu konfirmasi dari user\n", userID)
		return &AttendanceProcessResult{
			Success:        true,
			Message:        "Verifikasi berhasil - Silakan klik tombol konfirmasi untuk menyimpan ke database",
			FaceSimilarity: similarity,
			LocationValid:  locationValid,
			Distance:       distance,
		}, nil
	}

	// ✅ Jika verifyOnly=false, lakukan verifikasi + SIMPAN ke database
	fmt.Printf("💾 [SUBMIT] Menyimpan absensi %s untuk user %s ke database\n", recordType, userID)

	// 5. Process attendance dan simpan ke database
	var attendance *models.Attendance
	if recordType == "clock_in" {
		attendance, err = s.ClockIn(ctx, userID, latitude, longitude, photo, filename, similarity)
	} else {
		attendance, err = s.ClockOut(ctx, userID, latitude, longitude, photo, filename, similarity)
	}

	if err != nil {
		return &AttendanceProcessResult{
			Success:        false,
			Message:        err.Error(),
			FaceSimilarity: similarity,
			LocationValid:  locationValid,
			Distance:       distance,
		}, nil
	}

	// 6. Update face verification stats
	nowUTC := time.Now()
	faceEmbedding.LastVerifiedAt = &nowUTC
	faceEmbedding.VerificationCount++
	faceEmbedding.UpdatedAt = nowUTC
	_ = s.faceEmbeddingRepo.Update(ctx, faceEmbedding)

	actionMsg := "Clock In"
	if recordType == "clock_out" {
		actionMsg = "Clock Out"
	}

	fmt.Printf("✅ [SUCCESS] %s berhasil disimpan ke database - Status: %s\n", actionMsg, attendance.Status)
	return &AttendanceProcessResult{
		Success:        true,
		Message:        actionMsg + " berhasil dicatat",
		FaceSimilarity: similarity,
		LocationValid:  locationValid,
		Distance:       distance,
		Attendance:     attendance,
	}, nil
}

// ─── Helper Methods ───────────────────────────────────────────────────────────

func (s *attendanceService) extractTimeForToday(baseTime time.Time, scheduleTime time.Time) time.Time {
	return time.Date(
		baseTime.Year(), baseTime.Month(), baseTime.Day(),
		scheduleTime.Hour(), scheduleTime.Minute(), scheduleTime.Second(),
		0, wib,
	)
}

func (s *attendanceService) getDayName(wd time.Weekday) string {
	days := map[time.Weekday]string{
		time.Monday:    "Senin",
		time.Tuesday:   "Selasa",
		time.Wednesday: "Rabu",
		time.Thursday:  "Kamis",
		time.Friday:    "Jumat",
		time.Saturday:  "Sabtu",
		time.Sunday:    "Minggu",
	}
	return days[wd]
}

// ✅ Perbaikan: getDefaultJamKerja sekarang menerima parameter userID
func (s *attendanceService) getDefaultJamKerja(userID primitive.ObjectID) *models.JamKerja {
	now := time.Now().In(wib)
	// Set default waktu: 08:00 - 17:00
	waktuMulai := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, wib)
	waktuSelesai := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, wib)

	return &models.JamKerja{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		DayOfWeek:    []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat"},
		StartTime:   waktuMulai,
		EndTime:     waktuSelesai,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (s *attendanceService) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (s *attendanceService) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func formatDistance(d float64) string {
	if d < 1000 {
		return fmt.Sprintf("%.0f", d)
	}
	return fmt.Sprintf("%.1f km", d/1000)
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.1f", f)
}

type ScheduleInfoResponse struct {
	UserID        string                     `json:"user_id"`
	HariKerja     []string                   `json:"hari_kerja"`
	WaktuMulai    string                     `json:"waktu_mulai"`
	WaktuSelesai  string                     `json:"waktu_selesai"`
	Aktif         bool                       `json:"aktif"`
	TodaySchedule *TodayScheduleInfoResponse `json:"today_schedule,omitempty"`
}

type TodayScheduleInfoResponse struct {
	IsWorkDay      bool   `json:"is_work_day"`
	ClockInWindow  string `json:"clock_in_window"`
	ClockOutWindow string `json:"clock_out_window"`
	CanClockIn     bool   `json:"can_clock_in"`
	CanClockOut    bool   `json:"can_clock_out"`
	Message        string `json:"message"`
}
