// internal/service/attendance_service.go
package service

import (
	"bufio"
	"bytes"
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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// wib adalah timezone WIB (UTC+7).
// Seluruh logika bisnis (status Late, overtime, dll.) harus menggunakan
// waktu WIB agar sesuai dengan jam kerja di Indonesia.
var wib = time.FixedZone("WIB", 7*60*60)

// ── Interface ─────────────────────────────────────────────────────────────────

type AttendanceService interface {
	ClockIn(ctx context.Context, userID string, latitude, longitude float64, address string, geofenceID primitive.ObjectID, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error)
	ClockOut(ctx context.Context, userID string, latitude, longitude float64, address string, geofenceID primitive.ObjectID, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error)
	StartBreak(ctx context.Context, userID string) (*models.Attendance, error)
	EndBreak(ctx context.Context, userID string) (*models.Attendance, error)
	GetTodayAttendance(ctx context.Context, userID string) (*models.Attendance, error)
	GetMonthlyAttendance(ctx context.Context, userID string, month, year int) (*models.MonthlyAttendanceResponse, error)
	ProcessAttendanceWithFace(ctx context.Context, userID string, photo []byte, filename string, latitude, longitude float64, recordType string, verifyOnly bool) (*AttendanceProcessResult, error)
	ValidateClockInWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error)
	ValidateClockOutWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error)
	GetScheduleInfo(ctx context.Context, userID string) (*ScheduleInfoResponse, error)
	GetWorkScheduleInfo(ctx context.Context, userID string) (*WorkScheduleInfo, error)
	// Manager-only: rekap presensi & export CSV
	GetManagerAttendance(ctx context.Context, from, to time.Time, departmentName, q string, page, pageSize int64) (*models.ManagerAttendanceListResponse, error)
	ExportManagerAttendanceCSV(ctx context.Context, from, to time.Time, departmentName, q string) ([]byte, string, error)
	ExportManagerAttendanceCSVStream(ctx context.Context, from, to time.Time, departmentName, q string) (io.ReadCloser, string, error)
}

// ── Response structs ──────────────────────────────────────────────────────────

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
	ClockOutWindow string `json:"clock_out_window"` // HH:mm onwards atau HH:mm - HH:mm
	CanClockIn     bool   `json:"can_clock_in"`
	CanClockOut    bool   `json:"can_clock_out"`
	Message        string `json:"message"`
}

type BreakWindow struct {
	StartHour int
	EndHour   int
}

var allowedBreakWindows = []BreakWindow{
	{StartHour: 7, EndHour: 8},
	{StartHour: 12, EndHour: 13},
	{StartHour: 19, EndHour: 20},
}

func formatBreakWindowRange(w BreakWindow) string {
	return fmt.Sprintf("%02d:00-%02d:00", w.StartHour, w.EndHour)
}

func formatAllowedBreakWindows() string {
	parts := make([]string, 0, len(allowedBreakWindows))
	for _, w := range allowedBreakWindows {
		parts = append(parts, formatBreakWindowRange(w))
	}
	return strings.Join(parts, ", ")
}

func isInAllowedBreakWindow(nowWIB time.Time) (bool, string) {
	for _, w := range allowedBreakWindows {
		start := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), w.StartHour, 0, 0, 0, wib)
		end := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), w.EndHour, 0, 0, 0, wib)
		if !nowWIB.Before(start) && nowWIB.Before(end) {
			return true, formatBreakWindowRange(w)
		}
	}
	return false, ""
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

// ── Struct implementasi ───────────────────────────────────────────────────────
// Menggabungkan semua dependency:
//   - V1: officeLat/officeLng/radiusMeters sebagai fallback statis
//   - V2: geofenceRepo untuk validasi dinamis dari database

type attendanceService struct {
	db                *mongo.Database
	attendanceRepo    repository.AttendanceRepository
	breakTimeRepo     repository.BreakTimeRepository
	userRepo          repository.UserRepository
	faceEmbeddingRepo repository.FaceEmbeddingRepository
	jamKerjaRepo      repository.JamKerjaRepository
	geofenceRepo      repository.GeofenceRepository // dari V2: dynamic geofence
	faceClient        *faceclient.Client
	// dari V1: fallback statis jika DB geofence kosong
	officeLat    float64
	officeLng    float64
	radiusMeters float64
}

// ── Constructor ───────────────────────────────────────────────────────────────

func NewAttendanceService(
	db *mongo.Database,
	attendanceRepo repository.AttendanceRepository,
	breakTimeRepo repository.BreakTimeRepository,
	userRepo repository.UserRepository,
	faceEmbeddingRepo repository.FaceEmbeddingRepository,
	jamKerjaRepo repository.JamKerjaRepository,
	geofenceRepo repository.GeofenceRepository,
	faceClient *faceclient.Client,
) AttendanceService {
	return &attendanceService{
		db:                db,
		attendanceRepo:    attendanceRepo,
		breakTimeRepo:     breakTimeRepo,
		userRepo:          userRepo,
		faceEmbeddingRepo: faceEmbeddingRepo,
		jamKerjaRepo:      jamKerjaRepo,
		geofenceRepo:      geofenceRepo,
		faceClient:        faceClient,
		// Fallback statis (Labersa Hotel) digunakan jika tidak ada geofence aktif di DB
		officeLat:    2.3561,
		officeLng:    99.1431,
		radiusMeters: 10000,
	}
}

func (s *attendanceService) hasApprovedLeaveToday(ctx context.Context, userID primitive.ObjectID) (*models.LeaveRequest, error) {
	if s.db == nil {
		return nil, nil
	}

	nowWIB := time.Now().In(wib)
	startOfDayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib)
	endOfDayWIB := startOfDayWIB.Add(24 * time.Hour)

	filter := bson.M{
		"user_id":      userID,
		"final_status": models.StatusApproved,
		"start_date":   bson.M{"$lte": endOfDayWIB.UTC()},
		"end_date":     bson.M{"$gte": startOfDayWIB.UTC()},
	}

	var leave models.LeaveRequest
	err := s.db.Collection("leave_request").FindOne(ctx, filter).
		Decode(&leave)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &leave, nil
}

func blockedAttendanceMessage(leave *models.LeaveRequest) string {
	if leave == nil {
		return "Anda memiliki pengajuan izin/cuti yang sudah disetujui pada hari ini"
	}

	title := strings.TrimSpace(leave.TypeName)
	if title == "" {
		title = "izin/cuti"
	}

	return fmt.Sprintf("Tidak dapat melakukan attendance karena %s Anda sudah disetujui pada hari ini", title)
}

// ═══════════════════════════════════════════════════════════════════════════════
//  JADWAL KERJA (JamKerja)
// ═══════════════════════════════════════════════════════════════════════════════

// GetWorkScheduleInfo — info jadwal + window hari ini untuk dashboard Flutter
func (s *attendanceService) GetWorkScheduleInfo(ctx context.Context, userID string) (*WorkScheduleInfo, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	jamKerja, err := s.jamKerjaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if jamKerja == nil {
		jamKerja = s.getDefaultJamKerja(userObjID)
	}

	waktuMulaiStr := jamKerja.StartTime.Format("15:04")
	waktuSelesaiStr := jamKerja.EndTime.Format("15:04")

	info := &WorkScheduleInfo{
		UserID:       userID,
		HariKerja:    jamKerja.DayOfWeek,
		WaktuMulai:   waktuMulaiStr,
		WaktuSelesai: waktuSelesaiStr,
		Aktif:        jamKerja.IsActive,
	}

	nowWIB := time.Now().In(wib)
	dayName := s.getDayName(nowWIB.Weekday())
	isWorkDay := s.isWorkDay(dayName, jamKerja.DayOfWeek)
	leaveToday, err := s.hasApprovedLeaveToday(ctx, userObjID)
	if err != nil {
		return nil, err
	}

	startTimeToday := s.extractTimeForToday(nowWIB, jamKerja.StartTime)
	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
	clockInWindowOpen := startTimeToday.Add(-15 * time.Minute)
	clockOutWindowClose := endTimeToday.Add(30 * time.Minute)

	canClockIn := isWorkDay && !nowWIB.Before(clockInWindowOpen) && !nowWIB.After(endTimeToday)
	canClockOut := isWorkDay && !nowWIB.Before(endTimeToday) && !nowWIB.After(clockOutWindowClose)

	message := s.buildScheduleMessage(isWorkDay, canClockIn, canClockOut, nowWIB, clockInWindowOpen, clockOutWindowClose)
	if leaveToday != nil {
		canClockIn = false
		canClockOut = false
		message = blockedAttendanceMessage(leaveToday)
	}

	info.TodaySchedule = &TodaySchedule{
		IsWorkDay:      isWorkDay,
		ClockInWindow:  clockInWindowOpen.Format("15:04") + " - " + endTimeToday.Format("15:04"),
		ClockOutWindow: endTimeToday.Format("15:04") + " - " + clockOutWindowClose.Format("15:04"),
		CanClockIn:     canClockIn,
		CanClockOut:    canClockOut,
		Message:        message,
	}

	return info, nil
}

// GetScheduleInfo — info lengkap termasuk window, untuk handler /attendance/schedule
func (s *attendanceService) GetScheduleInfo(ctx context.Context, userID string) (*ScheduleInfoResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	jamKerja, err := s.jamKerjaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if jamKerja == nil {
		jamKerja = s.getDefaultJamKerja(userObjID)
	}

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
	isWorkDay := s.isWorkDay(dayName, jamKerja.DayOfWeek)
	leaveToday, err := s.hasApprovedLeaveToday(ctx, userObjID)
	if err != nil {
		return nil, err
	}

	startTimeToday := s.extractTimeForToday(nowWIB, jamKerja.StartTime)
	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
	clockInWindowOpen := startTimeToday.Add(-15 * time.Minute)
	clockOutWindowClose := endTimeToday.Add(30 * time.Minute)

	canClockIn := isWorkDay && !nowWIB.Before(clockInWindowOpen) && !nowWIB.After(endTimeToday)
	canClockOut := isWorkDay && !nowWIB.Before(endTimeToday) && !nowWIB.After(clockOutWindowClose)

	message := s.buildScheduleMessage(isWorkDay, canClockIn, canClockOut, nowWIB, clockInWindowOpen, clockOutWindowClose)
	if leaveToday != nil {
		canClockIn = false
		canClockOut = false
		message = blockedAttendanceMessage(leaveToday)
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

// ValidateClockInWindow — cek apakah sekarang dalam window clock-in
func (s *attendanceService) ValidateClockInWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, nil, errors.New("invalid user ID format")
	}

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
	leaveToday, err := s.hasApprovedLeaveToday(ctx, userObjID)
	if err != nil {
		return false, jamKerja, err
	}
	if leaveToday != nil {
		return false, jamKerja, errors.New(blockedAttendanceMessage(leaveToday))
	}

	if !s.isWorkDay(dayName, jamKerja.DayOfWeek) {
		return false, jamKerja, errors.New("hari ini bukan hari kerja")
	}

	startTimeToday := s.extractTimeForToday(nowWIB, jamKerja.StartTime)
	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
	clockInWindowOpen := startTimeToday.Add(-15 * time.Minute)

	isInWindow := !nowWIB.Before(clockInWindowOpen) && !nowWIB.After(endTimeToday)
	return isInWindow, jamKerja, nil
}

// ValidateClockOutWindow — cek apakah sekarang dalam window clock-out
func (s *attendanceService) ValidateClockOutWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, nil, errors.New("invalid user ID format")
	}

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
	leaveToday, err := s.hasApprovedLeaveToday(ctx, userObjID)
	if err != nil {
		return false, jamKerja, err
	}
	if leaveToday != nil {
		return false, jamKerja, errors.New(blockedAttendanceMessage(leaveToday))
	}

	if !s.isWorkDay(dayName, jamKerja.DayOfWeek) {
		return false, jamKerja, errors.New("hari ini bukan hari kerja")
	}

	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
	clockOutWindowClose := endTimeToday.Add(30 * time.Minute)

	isInWindow := !nowWIB.Before(endTimeToday) && !nowWIB.After(clockOutWindowClose)
	return isInWindow, jamKerja, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
//  CLOCK IN / CLOCK OUT
// ═══════════════════════════════════════════════════════════════════════════════

func (s *attendanceService) ClockIn(ctx context.Context, userID string, latitude, longitude float64, address string, geofenceID primitive.ObjectID, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	existing, _ := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if existing != nil && existing.ClockInTime != nil {
		return nil, errors.New("sudah melakukan clock in hari ini")
	}
	leaveToday, err := s.hasApprovedLeaveToday(ctx, userObjID)
	if err != nil {
		return nil, err
	}
	if leaveToday != nil {
		return nil, errors.New(blockedAttendanceMessage(leaveToday))
	}

	isInWindow, jamKerja, err := s.ValidateClockInWindow(ctx, userID)
	if err != nil {
		return nil, errors.New("tidak dapat clock in: " + err.Error())
	}
	if !isInWindow {
		nowWIB := time.Now().In(wib)
		startTimeToday := s.extractTimeForToday(nowWIB, jamKerja.StartTime)
		windowOpen := startTimeToday.Add(-15 * time.Minute)
		endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
		var message string
		if nowWIB.Before(windowOpen) {
			message = "jendela clock in belum dibuka (buka pada " + windowOpen.Format("15:04") + " WIB)"
		} else if nowWIB.After(endTimeToday) {
			message = "jendela clock in sudah tutup (buka kembali besok jam " + windowOpen.Format("15:04") + " WIB)"
		}
		return nil, errors.New(message)
	}

	now := time.Now().In(wib)
	location := models.GeoLocation{Latitude: latitude, Longitude: longitude, Address: address}

	// Status: Tepat Waktu jika clock-in <= startTime + 1 menit toleransi
	startTimeToday := s.extractTimeForToday(now, jamKerja.StartTime)
	lateThreshold := startTimeToday.Add(1 * time.Minute)
	status := models.StatusOnTime
	if now.After(lateThreshold) {
		status = models.StatusLate
	}

	fmt.Printf("⏰ Clock In Status:\n  Start: %s | Late threshold: %s | Submit: %s | Status: %s\n",
		startTimeToday.Format("15:04:05"), lateThreshold.Format("15:04:05"),
		now.Format("15:04:05"), status)

	attendance := &models.Attendance{
		ID:              primitive.NewObjectID(),
		UserID:          userObjID,
		Date:            now,
		ClockInTime:     &now,
		ClockInPhoto:    filename,
		ClockInLocation: location,
		Status:          status,
		GeofenceID:      geofenceID,
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

func (s *attendanceService) ClockOut(ctx context.Context, userID string, latitude, longitude float64, address string, geofenceID primitive.ObjectID, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	existing, _ := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if existing != nil && existing.ClockOutTime != nil {
		return nil, errors.New("sudah melakukan clock out hari ini")
	}
	leaveToday, err := s.hasApprovedLeaveToday(ctx, userObjID)
	if err != nil {
		return nil, err
	}
	if leaveToday != nil {
		return nil, errors.New(blockedAttendanceMessage(leaveToday))
	}

	isInWindow, jamKerja, err := s.ValidateClockOutWindow(ctx, userID)
	if err != nil {
		return nil, errors.New("tidak dapat clock out: " + err.Error())
	}
	if !isInWindow {
		nowWIB := time.Now().In(wib)
		endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)
		clockOutWindowClose := endTimeToday.Add(30 * time.Minute)
		var message string
		if nowWIB.Before(endTimeToday) {
			message = "jendela clock out belum dibuka (buka pada " + endTimeToday.Format("15:04") + " WIB)"
		} else if nowWIB.After(clockOutWindowClose) {
			message = "jendela clock out sudah tutup (buka kembali besok jam " + endTimeToday.Format("15:04") + " WIB)"
		}
		return nil, errors.New(message)
	}

	if existing == nil || existing.ClockInTime == nil {
		return nil, errors.New("belum melakukan clock in hari ini")
	}

	now := time.Now().In(wib)
	location := models.GeoLocation{Latitude: latitude, Longitude: longitude, Address: address}

	workDuration := now.Sub(*existing.ClockInTime)
	workHours := workDuration.Hours()
	overtimeHours := 0.0
	if workHours > 9.0 {
		overtimeHours = workHours - 9.0
	}

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

	if err = s.attendanceRepo.UpdateClockOut(ctx, existing.ID, now, filename, location); err != nil {
		return nil, err
	}
	if err = s.attendanceRepo.UpdateWorkHours(ctx, existing.ID, existing.WorkHours, existing.OvertimeHours, existing.Status); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *attendanceService) StartBreak(ctx context.Context, userID string) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	attendance, err := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if err != nil {
		return nil, err
	}
	if attendance == nil || attendance.ClockInTime == nil {
		return nil, errors.New("belum clock in hari ini")
	}
	if attendance.ClockOutTime != nil {
		return nil, errors.New("tidak dapat memulai istirahat setelah clock out")
	}
	if attendance.BreakStartTime != nil && attendance.BreakEndTime == nil {
		return nil, errors.New("sedang dalam sesi istirahat")
	}
	if attendance.BreakStartTime != nil && attendance.BreakEndTime != nil {
		return nil, errors.New("istirahat hari ini sudah digunakan")
	}

	nowWIB := time.Now().In(wib)
	inWindow, _ := isInAllowedBreakWindow(nowWIB)
	if !inWindow {
		return nil, fmt.Errorf("break hanya dapat dimulai pada jam %s WIB", formatAllowedBreakWindows())
	}

	activeBreak, err := s.breakTimeRepo.FindActiveTodayByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if activeBreak != nil {
		return nil, errors.New("sedang dalam sesi istirahat")
	}

	breakRecord := &models.BreakTime{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Date:      time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib),
		StartTime: nowWIB,
		Status:    "ONGOING",
		CreatedAt: nowWIB,
		UpdatedAt: nowWIB,
	}
	if err := s.breakTimeRepo.Create(ctx, breakRecord); err != nil {
		return nil, err
	}

	if err := s.attendanceRepo.UpdateBreakStart(ctx, attendance.ID, nowWIB); err != nil {
		return nil, err
	}

	attendance.BreakStartTime = &nowWIB
	attendance.BreakEndTime = nil
	attendance.UpdatedAt = nowWIB
	return attendance, nil
}

func (s *attendanceService) EndBreak(ctx context.Context, userID string) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	attendance, err := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if err != nil {
		return nil, err
	}
	if attendance == nil || attendance.ClockInTime == nil {
		return nil, errors.New("belum clock in hari ini")
	}
	if attendance.BreakStartTime == nil {
		return nil, errors.New("belum memulai istirahat")
	}
	if attendance.BreakEndTime != nil {
		return nil, errors.New("istirahat sudah selesai")
	}

	breakRecord, err := s.breakTimeRepo.FindActiveTodayByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if breakRecord == nil {
		return nil, errors.New("belum memulai istirahat")
	}

	nowWIB := time.Now().In(wib)
	inWindow, _ := isInAllowedBreakWindow(nowWIB)
	if !inWindow {
		return nil, fmt.Errorf("break hanya dapat diakhiri pada jam %s WIB", formatAllowedBreakWindows())
	}

	if err := s.breakTimeRepo.UpdateEnd(ctx, breakRecord.ID, nowWIB, "DONE"); err != nil {
		return nil, err
	}
	if err := s.attendanceRepo.UpdateBreakEnd(ctx, attendance.ID, nowWIB); err != nil {
		return nil, err
	}

	attendance.BreakEndTime = &nowWIB
	attendance.UpdatedAt = nowWIB
	return attendance, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
//  GET TODAY / MONTHLY
// ═══════════════════════════════════════════════════════════════════════════════

func (s *attendanceService) GetTodayAttendance(ctx context.Context, userID string) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	attendance, err := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if err != nil || attendance == nil {
		return attendance, err
	}

	if attendance.BreakStartTime == nil || attendance.BreakEndTime == nil {
		breakRecord, breakErr := s.breakTimeRepo.FindTodayByUserID(ctx, userID)
		if breakErr == nil && breakRecord != nil {
			if attendance.BreakStartTime == nil {
				start := breakRecord.StartTime
				if !start.IsZero() {
					attendance.BreakStartTime = &start
				}
			}
			if attendance.BreakEndTime == nil && !breakRecord.EndTime.IsZero() {
				end := breakRecord.EndTime
				attendance.BreakEndTime = &end
			}
		}
	}

	return attendance, nil
}

func (s *attendanceService) GetMonthlyAttendance(ctx context.Context, userID string, month, year int) (*models.MonthlyAttendanceResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	summary, err := s.attendanceRepo.GetMonthlySummary(ctx, userObjID, year, month)
	if err != nil || summary == nil {
		return summary, err
	}

	breakRecords, breakErr := s.breakTimeRepo.FindByUserIDAndMonth(ctx, userID, year, month)
	if breakErr != nil || len(breakRecords) == 0 {
		return summary, nil
	}

	breakByDate := make(map[string]models.BreakTime)
	for _, br := range breakRecords {
		dateKey := br.Date.In(wib).Format("2006-01-02")
		if _, exists := breakByDate[dateKey]; !exists {
			breakByDate[dateKey] = br
		}
	}

	for i := range summary.Records {
		rec := &summary.Records[i]
		if rec.BreakStartTime != "" && rec.BreakEndTime != "" {
			continue
		}
		br, ok := breakByDate[rec.Date]
		if !ok {
			continue
		}
		if rec.BreakStartTime == "" && !br.StartTime.IsZero() {
			rec.BreakStartTime = br.StartTime.In(wib).Format("15:04")
		}
		if rec.BreakEndTime == "" && !br.EndTime.IsZero() {
			rec.BreakEndTime = br.EndTime.In(wib).Format("15:04")
		}
	}

	return summary, nil
}

func (s *attendanceService) resolveUserFromIdentifier(ctx context.Context, identifier string) (*models.User, string, error) {
	id := strings.TrimSpace(identifier)
	if id == "" {
		return nil, "", errors.New("user tidak ditemukan")
	}

	// Primary path: token already contains Mongo ObjectID.
	if _, err := primitive.ObjectIDFromHex(id); err == nil {
		user, findErr := s.userRepo.FindByID(ctx, id)
		if findErr == nil && user != nil {
			return user, user.ID.Hex(), nil
		}
	}

	// Backward compatibility: allow old tokens that stored email/payroll as user_id.
	if user, err := s.userRepo.FindByEmail(ctx, id); err == nil && user != nil {
		return user, user.ID.Hex(), nil
	}
	if user, err := s.userRepo.FindByPayrollNumber(ctx, id); err == nil && user != nil {
		return user, user.ID.Hex(), nil
	}

	return nil, "", errors.New("user tidak ditemukan")
}

// ═══════════════════════════════════════════════════════════════════════════════
//  PROCESS ATTENDANCE WITH FACE
//  Gabungan: geofence dinamis dari DB (V2) + fallback statis (V1)
// ═══════════════════════════════════════════════════════════════════════════════

func (s *attendanceService) ProcessAttendanceWithFace(
	ctx context.Context,
	userID string,
	photo []byte,
	filename string,
	latitude, longitude float64,
	recordType string,
	verifyOnly bool,
) (*AttendanceProcessResult, error) {

	// 1. Validasi lokasi — coba geofence dinamis dari DB terlebih dahulu.
	//    Jika tidak ada geofence aktif, fallback ke koordinat statis (V1).
	user, resolvedUserID, err := s.resolveUserFromIdentifier(ctx, userID)
	if err != nil || user == nil {
		return &AttendanceProcessResult{
			Success: false, Message: "data user tidak ditemukan",
			LocationValid: false, Distance: 0,
		}, nil
	}
	leaveToday, err := s.hasApprovedLeaveToday(ctx, user.ID)
	if err != nil {
		return &AttendanceProcessResult{
			Success: false, Message: err.Error(),
			LocationValid: false, Distance: 0,
		}, nil
	}
	if leaveToday != nil {
		return &AttendanceProcessResult{
			Success: false, Message: blockedAttendanceMessage(leaveToday),
			LocationValid: false, Distance: 0,
		}, nil
	}

	locationValid, distance, locationMsg, address, geofenceID := s.validateLocation(ctx, user, latitude, longitude)
	if !locationValid {
		return &AttendanceProcessResult{
			Success: false, Message: locationMsg,
			LocationValid: false, Distance: distance,
		}, nil
	}

	// 2. Cek embedding wajah
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, resolvedUserID)
	if err != nil || faceEmbedding == nil {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Wajah belum terdaftar. Silakan daftarkan wajah Anda terlebih dahulu.",
			LocationValid: locationValid, Distance: distance,
		}, nil
	}
	if len(faceEmbedding.FaceEmbedding) == 0 {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Data embedding wajah tidak valid. Silakan daftarkan ulang wajah Anda.",
			LocationValid: locationValid, Distance: distance,
		}, nil
	}

	// 3. Ekstrak embedding foto saat ini
	currentEmbedding, err := s.faceClient.ExtractEmbedding(resolvedUserID, photo, filename)
	if err != nil {
		return &AttendanceProcessResult{
			Success: false, Message: "Gagal memproses foto: " + err.Error(),
			LocationValid: locationValid, Distance: distance,
		}, nil
	}
	if len(currentEmbedding) == 0 {
		return &AttendanceProcessResult{
			Success: false, Message: "Tidak ada wajah terdeteksi dalam foto",
			LocationValid: locationValid, Distance: distance,
		}, nil
	}

	// 4. Hitung similarity
	//    threshold 0.75 (V2 lebih ketat) — sesuaikan jika perlu
	similarity := s.cosineSimilarity(currentEmbedding, faceEmbedding.FaceEmbedding)
	const threshold = 0.75
	if similarity < threshold {
		return &AttendanceProcessResult{
			Success:        false,
			Message:        fmt.Sprintf("Wajah tidak cocok (similarity: %.1f%%, min: %.1f%%)", similarity*100, threshold*100),
			FaceSimilarity: similarity, LocationValid: locationValid, Distance: distance,
		}, nil
	}

	// 5. verifyOnly=true → hanya return hasil, tidak simpan ke DB
	if verifyOnly {
		fmt.Printf("✅ [VERIFY ONLY] user=%s\n", resolvedUserID)
		return &AttendanceProcessResult{
			Success:        true,
			Message:        "Verifikasi berhasil - klik Konfirmasi untuk menyimpan",
			FaceSimilarity: similarity, LocationValid: locationValid, Distance: distance,
		}, nil
	}

	// 6. Simpan ke database
	fmt.Printf("💾 [SUBMIT] %s untuk user=%s\n", recordType, resolvedUserID)
	var attendance *models.Attendance
	if recordType == "clock_in" {
		attendance, err = s.ClockIn(ctx, resolvedUserID, latitude, longitude, address, geofenceID, photo, filename, similarity)
	} else {
		attendance, err = s.ClockOut(ctx, resolvedUserID, latitude, longitude, address, geofenceID, photo, filename, similarity)
	}
	if err != nil {
		return &AttendanceProcessResult{
			Success: false, Message: err.Error(),
			FaceSimilarity: similarity, LocationValid: locationValid, Distance: distance,
		}, nil
	}

	// 7. Update stats embedding
	nowUTC := time.Now()
	faceEmbedding.LastVerifiedAt = &nowUTC
	faceEmbedding.VerificationCount++
	faceEmbedding.UpdatedAt = nowUTC
	_ = s.faceEmbeddingRepo.Update(ctx, faceEmbedding)

	actionMsg := "Clock In"
	if recordType == "clock_out" {
		actionMsg = "Clock Out"
	}
	fmt.Printf("✅ [SUCCESS] %s disimpan — Status: %s\n", actionMsg, attendance.Status)

	return &AttendanceProcessResult{
		Success: true, Message: actionMsg + " berhasil dicatat",
		FaceSimilarity: similarity, LocationValid: locationValid, Distance: distance,
		Attendance: attendance,
	}, nil
}

// validateLocation — coba geofence DB dulu, fallback ke koordinat statis
func (s *attendanceService) validateLocation(ctx context.Context, user *models.User, latitude, longitude float64) (bool, float64, string, string, primitive.ObjectID) {
	activeGeofences, err := s.geofenceRepo.FindActive(ctx)
	if err == nil && len(activeGeofences) > 0 {
		// Cari geofence yang berlaku untuk user ini
		applicable := make([]models.Geofence, 0)
		for _, g := range activeGeofences {
			if s.geofenceAppliesToUser(user, &g) {
				applicable = append(applicable, g)
			}
		}

		if len(applicable) > 0 {
			closestDist := math.MaxFloat64
			var matched *models.Geofence
			for i := range applicable {
				g := &applicable[i]
				d := s.calculateDistance(latitude, longitude, g.Latitude, g.Longitude)
				if d < closestDist {
					closestDist = d
				}
				if d <= float64(g.Radius) && (matched == nil ||
					d < s.calculateDistance(latitude, longitude, matched.Latitude, matched.Longitude)) {
					matched = g
				}
			}
			if matched != nil {
				return true, closestDist, "", matched.Name, matched.ID
			}
			return false, closestDist,
				fmt.Sprintf("Anda berada di luar area geofence (jarak terdekat: %s)", formatDistance(closestDist)), "", primitive.NilObjectID
		}
	}

	// Fallback: gunakan koordinat statis jika tidak ada geofence DB aktif
	dist := s.calculateDistance(latitude, longitude, s.officeLat, s.officeLng)
	if dist <= s.radiusMeters {
		return true, dist, "", "Kantor Pusat", primitive.NilObjectID
	}
	return false, dist,
		"Anda berada di luar area kantor (jarak: " + formatDistance(dist) + "m, max: " + formatDistance(s.radiusMeters) + "m)", "", primitive.NilObjectID
}

func (s *attendanceService) geofenceAppliesToUser(user *models.User, geofence *models.Geofence) bool {
	switch geofence.AppliesTo {
	case "", "all":
		return true
	case "departments":
		for _, deptID := range geofence.DepartmentIDs {
			if deptID == user.DepartmentID {
				return true
			}
		}
		return false
	case "positions":
		for _, posID := range geofence.PositionIDs {
			if posID == user.PositionID {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
//  MANAGER: REKAP PRESENSI + CSV EXPORT
// ═══════════════════════════════════════════════════════════════════════════════

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
			Avatar: r.User.Avatar,
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

// ═══════════════════════════════════════════════════════════════════════════════
//  HELPER METHODS
// ═══════════════════════════════════════════════════════════════════════════════

// isWorkDay cek apakah dayName ada dalam workDays
func (s *attendanceService) isWorkDay(dayName string, workDays []string) bool {
	for _, d := range workDays {
		if d == dayName {
			return true
		}
	}
	return false
}

// buildScheduleMessage membangun pesan status jadwal
func (s *attendanceService) buildScheduleMessage(isWorkDay, canClockIn, canClockOut bool, now, clockInOpen, clockOutClose time.Time) string {
	if !isWorkDay {
		return "Hari ini bukan hari kerja"
	}
	if canClockIn || canClockOut {
		return ""
	}
	if now.Before(clockInOpen) {
		return "Clock in dibuka pada " + clockInOpen.Format("15:04") + " WIB"
	}
	if now.After(clockOutClose) {
		return "CLOCK IN"
	}
	return ""
}

// extractTimeForToday: ambil jam dari scheduleTime, tempel ke tanggal baseTime
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

func (s *attendanceService) isWithinAllowedBreakWindow(nowWIB time.Time) bool {
	for _, window := range allowedBreakWindows {
		start := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), window.StartHour, 0, 0, 0, wib)
		end := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), window.EndHour, 0, 0, 0, wib)
		if !nowWIB.Before(start) && nowWIB.Before(end) {
			return true
		}
	}
	return false
}

func (s *attendanceService) formatAllowedBreakWindows() string {
	parts := make([]string, 0, len(allowedBreakWindows))
	for _, window := range allowedBreakWindows {
		parts = append(parts, fmt.Sprintf("%02d:00-%02d:00", window.StartHour, window.EndHour))
	}
	return strings.Join(parts, ", ")
}

// getDefaultJamKerja: Senin-Jumat 08:00-17:00 jika user belum punya jadwal
func (s *attendanceService) getDefaultJamKerja(userID primitive.ObjectID) *models.JamKerja {
	now := time.Now().In(wib)
	waktuMulai := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, wib)
	waktuSelesai := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, wib)
	return &models.JamKerja{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		DayOfWeek: []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat"},
		StartTime: waktuMulai,
		EndTime:   waktuSelesai,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s *attendanceService) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (s *attendanceService) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ── Utility functions ─────────────────────────────────────────────────────────

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
	if loc.Address != "" {
		return loc.Address
	}
	if loc.Latitude != 0 || loc.Longitude != 0 {
		return fmt.Sprintf("%.5f, %.5f", loc.Latitude, loc.Longitude)
	}
	return "Unrecorded"
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

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\n\r\"") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}
