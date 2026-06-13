// internal/service/attendance_service.go
package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ProcessAttendanceWithFace(ctx context.Context, userID string, photo []byte, filename string, latitude, longitude float64, recordType string, verifyOnly bool, liveness string) (*AttendanceProcessResult, error)
	ValidateClockInWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error)
	ValidateClockOutWindow(ctx context.Context, userID string) (bool, *models.JamKerja, error)
	GetScheduleInfo(ctx context.Context, userID string) (*ScheduleInfoResponse, error)
	GetWorkScheduleInfo(ctx context.Context, userID string) (*WorkScheduleInfo, error)
	SetWSHub(hub *WSHub) // inject WSHub untuk real-time broadcast
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
	{StartHour: 12, EndHour: 13},
	{StartHour: 18, EndHour: 19},
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
	Success           bool                   `json:"success"`
	Message           string                 `json:"message"`
	FaceSimilarity    float64                `json:"face_similarity"`
	LocationValid     bool                   `json:"location_valid"`
	Distance          float64                `json:"distance_m"`
	Attendance        *models.Attendance     `json:"attendance,omitempty"`
	Face              *faceclient.FaceResult `json:"face,omitempty"`
	Geo               *faceclient.GeoResult  `json:"geo,omitempty"`
	IsClockInAllowed  bool                   `json:"is_clock_in_allowed,omitempty"`
	IsClockOutAllowed bool                   `json:"is_clock_out_allowed,omitempty"`
	ClockInWindow     string                 `json:"clock_in_window,omitempty"`
	ClockOutWindow    string                 `json:"clock_out_window,omitempty"`
	NextWindowOpen    string                 `json:"next_window_open,omitempty"`
	WorkScheduleFound bool                   `json:"work_schedule_found,omitempty"`
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
	overtimeRepo      repository.OvertimeRequestRepository
	faceClient        *faceclient.Client
	wsHub             *WSHub // untuk broadcast real-time events
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
	overtimeRepo repository.OvertimeRequestRepository,
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
		overtimeRepo:      overtimeRepo,
		faceClient:        faceClient,
		wsHub:             nil, // set via SetWSHub jika diperlukan
		// Fallback statis (Labersa Hotel) digunakan jika tidak ada geofence aktif di DB
		officeLat:    2.3561,
		officeLng:    99.1431,
		radiusMeters: 10000,
	}
}

// SetWSHub mengatur WebSocket hub untuk broadcast real-time events.
func (s *attendanceService) SetWSHub(hub *WSHub) {
	s.wsHub = hub
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

func (s *attendanceService) resolveEffectiveJamKerjaForToday(ctx context.Context, userID primitive.ObjectID, jamKerja *models.JamKerja) (*models.JamKerja, bool, error) {
	if jamKerja == nil {
		return nil, false, nil
	}
	if s.db == nil {
		return jamKerja, false, nil
	}

	nowWIB := time.Now().In(wib)
	startOfDayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib)
	endOfDayWIB := startOfDayWIB.Add(24 * time.Hour)

	// 1. Cek apakah hari ini adalah hari libur pengganti (Reward dari Penugasan)
	rewardFilter := bson.M{
		"employees": bson.M{"$elemMatch": bson.M{
			"user_id":               userID,
			"day_off_reward.status": models.DayOffRewardStatusUsed,
			"day_off_reward.replacement_off_date": bson.M{
				"$gte": startOfDayWIB.UTC(),
				"$lt":  endOfDayWIB.UTC(),
			},
		}},
	}

	var rewardAssignment models.Assignment
	err := s.db.Collection("assignments").FindOne(ctx, rewardFilter).Decode(&rewardAssignment)
	if err == nil {
		// Ditemukan reward untuk hari ini -> user LIBUR
		effective := *jamKerja
		effective.DayOfWeek = []string{} // Kosongkan agar isWorkDay menjadi false
		effective.IsActive = false
		return &effective, true, nil
	}

	// 2. Cek apakah ada penugasan (kerja) untuk hari ini
	filter := bson.M{
		"date": bson.M{
			"$gte": startOfDayWIB.UTC(),
			"$lt":  endOfDayWIB.UTC(),
		},
		"status": bson.M{"$in": bson.A{models.AssignmentStatusSubmitted, models.AssignmentStatusPublished}},
		"employees": bson.M{"$elemMatch": bson.M{
			"user_id":         userID,
			"employee_status": models.AssignmentEmployeeStatusAgreed,
		}},
	}

	var assignment models.Assignment
	err = s.db.Collection("assignments").FindOne(ctx, filter, options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})).Decode(&assignment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return jamKerja, false, nil
		}
		return nil, false, err
	}

	for _, employee := range assignment.Employees {
		if employee.UserID != userID || employee.EmployeeStatus != models.AssignmentEmployeeStatusAgreed {
			continue
		}

		startShift := strings.TrimSpace(employee.AssignedShift.StartTime)
		endShift := strings.TrimSpace(employee.AssignedShift.EndTime)
		if startShift == "" || endShift == "" {
			return jamKerja, false, nil
		}

		startTime, startErr := time.ParseInLocation("15:04", startShift, wib)
		if startErr != nil {
			return nil, false, fmt.Errorf("invalid assigned_shift start_time: %w", startErr)
		}
		endTime, endErr := time.ParseInLocation("15:04", endShift, wib)
		if endErr != nil {
			return nil, false, fmt.Errorf("invalid assigned_shift end_time: %w", endErr)
		}

		effective := *jamKerja
		effective.StartTime = startTime
		effective.EndTime = endTime
		effective.DayOfWeek = []string{s.getDayName(nowWIB.Weekday())}
		effective.IsActive = true

		return &effective, true, nil
	}

	return jamKerja, false, nil
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
	effectiveJamKerja, _, err := s.resolveEffectiveJamKerjaForToday(ctx, userObjID, jamKerja)
	if err != nil {
		return nil, err
	}
	jamKerja = effectiveJamKerja

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

	// Cek reward Early Leave (Pulang Cepat)
	reductionOut, _ := s.getTimeOffReductionForToday(ctx, userObjID, "early_out")
	if reductionOut > 0 {
		endTimeToday = endTimeToday.Add(-time.Duration(reductionOut * float64(time.Hour)))
		info.WaktuSelesai = endTimeToday.Format("15:04")
	}

	// Cek reward Late In (Masuk Terlambat)
	reductionIn, _ := s.getTimeOffReductionForToday(ctx, userObjID, "late_in")
	if reductionIn > 0 {
		startTimeToday = startTimeToday.Add(time.Duration(reductionIn * float64(time.Hour)))
		info.WaktuMulai = startTimeToday.Format("15:04")
	}

	clockInWindowOpen := startTimeToday.Add(-15 * time.Minute)
	// Clock out window: tutup 6 jam setelah jam pulang (bukan 30 menit)
	clockOutWindowClose := endTimeToday.Add(6 * time.Hour)

	// Cek apakah user sudah clock in hari ini untuk canClockOut
	hasClockIn := false
	if todayAtt, _ := s.attendanceRepo.FindTodayByUserID(ctx, userObjID); todayAtt != nil && todayAtt.ClockInTime != nil {
		hasClockIn = true
	}

	canClockIn := isWorkDay && !nowWIB.Before(clockInWindowOpen) && !nowWIB.After(endTimeToday)
	// Clock out bisa dilakukan sejak clock in hingga 6 jam setelah jam pulang
	canClockOut := isWorkDay && hasClockIn && !nowWIB.After(clockOutWindowClose)

	message := s.buildScheduleMessage(isWorkDay, canClockIn, canClockOut, nowWIB, clockInWindowOpen, clockOutWindowClose)
	if leaveToday != nil {
		canClockIn = false
		canClockOut = false
		message = blockedAttendanceMessage(leaveToday)
	}

	info.TodaySchedule = &TodaySchedule{
		IsWorkDay:      isWorkDay,
		ClockInWindow:  clockInWindowOpen.Format("15:04") + " - " + endTimeToday.Format("15:04"),
		ClockOutWindow: "Sejak Clock In - " + clockOutWindowClose.Format("15:04"),
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
	effectiveJamKerja, _, err := s.resolveEffectiveJamKerjaForToday(ctx, userObjID, jamKerja)
	if err != nil {
		return nil, err
	}
	jamKerja = effectiveJamKerja

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

	// Cek reward Early Leave (Pulang Cepat)
	reductionOut, _ := s.getTimeOffReductionForToday(ctx, userObjID, "early_out")
	if reductionOut > 0 {
		endTimeToday = endTimeToday.Add(-time.Duration(reductionOut * float64(time.Hour)))
		info.WaktuSelesai = endTimeToday.Format("15:04")
	}

	// Cek reward Late In (Masuk Terlambat)
	reductionIn, _ := s.getTimeOffReductionForToday(ctx, userObjID, "late_in")
	if reductionIn > 0 {
		startTimeToday = startTimeToday.Add(time.Duration(reductionIn * float64(time.Hour)))
		info.WaktuMulai = startTimeToday.Format("15:04")
	}

	clockInWindowOpen := startTimeToday.Add(-15 * time.Minute)
	// Clock out window: tutup 6 jam setelah jam pulang
	clockOutWindowClose := endTimeToday.Add(6 * time.Hour)

	// Cek apakah user sudah clock in hari ini
	hasClockIn := false
	if todayAtt, _ := s.attendanceRepo.FindTodayByUserID(ctx, userObjID); todayAtt != nil && todayAtt.ClockInTime != nil {
		hasClockIn = true
	}

	canClockIn := isWorkDay && !nowWIB.Before(clockInWindowOpen) && !nowWIB.After(endTimeToday)
	// Clock out bisa dilakukan sejak clock in hingga 6 jam setelah jam pulang
	canClockOut := isWorkDay && hasClockIn && !nowWIB.After(clockOutWindowClose)

	message := s.buildScheduleMessage(isWorkDay, canClockIn, canClockOut, nowWIB, clockInWindowOpen, clockOutWindowClose)
	if leaveToday != nil {
		canClockIn = false
		canClockOut = false
		message = blockedAttendanceMessage(leaveToday)
	}

	info.TodaySchedule = &TodayScheduleInfoResponse{
		IsWorkDay:      isWorkDay,
		ClockInWindow:  clockInWindowOpen.Format("15:04") + " - " + endTimeToday.Format("15:04"),
		ClockOutWindow: "Sejak Clock In - " + clockOutWindowClose.Format("15:04"),
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
	effectiveJamKerja, _, err := s.resolveEffectiveJamKerjaForToday(ctx, userObjID, jamKerja)
	if err != nil {
		return false, nil, err
	}
	jamKerja = effectiveJamKerja
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
	// Cek reward Late In (Masuk Terlambat)
	reductionIn, _ := s.getTimeOffReductionForToday(ctx, userObjID, "late_in")
	if reductionIn > 0 {
		startTimeToday = startTimeToday.Add(time.Duration(reductionIn * float64(time.Hour)))
	}

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
	effectiveJamKerja, _, err := s.resolveEffectiveJamKerjaForToday(ctx, userObjID, jamKerja)
	if err != nil {
		return false, nil, err
	}
	jamKerja = effectiveJamKerja
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

	// ValidateClockOutWindow: window buka sejak clock in, tutup 6 jam setelah jam pulang
	endTimeToday := s.extractTimeForToday(nowWIB, jamKerja.EndTime)

	// Cek reward Early Leave (Pulang Cepat)
	reduction, _ := s.getTimeOffReductionForToday(ctx, userObjID, "early_out")
	if reduction > 0 {
		endTimeToday = endTimeToday.Add(-time.Duration(reduction * float64(time.Hour)))
	}

	// Clock out window tutup 6 jam setelah jam pulang kerja
	clockOutWindowClose := endTimeToday.Add(6 * time.Hour)

	// Cek apakah user sudah clock in hari ini
	existingAtt, _ := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if existingAtt == nil || existingAtt.ClockInTime == nil {
		return false, jamKerja, errors.New("belum melakukan clock in hari ini")
	}

	// Window terbuka: sejak clock in, sampai 6 jam setelah jam pulang
	isInWindow := !nowWIB.After(clockOutWindowClose)
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
		clockOutWindowClose := endTimeToday.Add(6 * time.Hour)
		var message string
		if nowWIB.After(clockOutWindowClose) {
			message = "jendela clock out sudah tutup (buka kembali setelah clock in besok)"
		} else {
			message = "clock out hanya dapat dilakukan setelah clock in"
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

	// Cek reward Early Leave (Pulang Cepat) & Late In (Masuk Terlambat)
	redOut, _ := s.getTimeOffReductionForToday(ctx, userObjID, "early_out")
	redIn, _ := s.getTimeOffReductionForToday(ctx, userObjID, "late_in")

	if redOut > 0 || redIn > 0 {
		workHours += (redOut + redIn)
		// Mark rewards as used in background
		go s.markTimeOffRewardsAsUsed(context.Background(), userObjID)
	}

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
		return nil, fmt.Errorf("istirahat hanya dapat dimulai pada jam %s WIB", formatAllowedBreakWindows())
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
		return nil, fmt.Errorf("istirahat hanya dapat diakhiri pada jam %s WIB", formatAllowedBreakWindows())
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

	// Cari semua overtime reward time_off di bulan ini
	startOfMonthWIB := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, wib)
	endOfMonthWIB := startOfMonthWIB.AddDate(0, 1, 0)

	rewardFilter := bson.M{
		"employees": bson.M{
			"$elemMatch": bson.M{
				"user_id":            userObjID,
				"reward.reward_type": models.OvertimeRewardTypeTimeOff,
				"reward.reward_date": bson.M{
					"$gte": startOfMonthWIB,
					"$lt":  endOfMonthWIB,
				},
				"reward.status": bson.M{"$in": []string{models.OvertimeRewardStatusGranted, models.OvertimeRewardStatusUsed}},
			},
		},
	}

	rewardsInfoByDate := make(map[string]string)
	if s.overtimeRepo != nil {
		requests, _ := s.overtimeRepo.Find(ctx, rewardFilter)
		for _, req := range requests {
			for _, emp := range req.Employees {
				if emp.UserID == userObjID && emp.Reward.RewardType == models.OvertimeRewardTypeTimeOff && emp.Reward.RewardDate != nil {
					rd := emp.Reward.RewardDate.In(wib).Format("2006-01-02")
					hours := req.GetDurationHours()
					var label string
					if emp.Reward.RewardOption == "early_out" {
						label = fmt.Sprintf("Pulang Cepat %.1f jam (Reward Lembur)", hours)
					} else {
						label = fmt.Sprintf("Masuk Siang %.1f jam (Reward Lembur)", hours)
					}

					if existing, ok := rewardsInfoByDate[rd]; ok {
						rewardsInfoByDate[rd] = existing + ", " + label
					} else {
						rewardsInfoByDate[rd] = label
					}
				}
			}
		}
	}

	for i := range summary.Records {
		rec := &summary.Records[i]
		if info, ok := rewardsInfoByDate[rec.Date]; ok {
			rec.RewardInfo = info
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
	liveness string,
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

	// 3. Panggil FastAPI /face/verify untuk validasi wajah, liveness, spoofing, aksesoris, dan similarity sekaligus!
	thr := 0.75
	verifyResp, err := s.faceClient.VerifyFace(resolvedUserID, faceEmbedding.FaceEmbedding, photo, filename, liveness, &thr)
	if err != nil {
		// Clean up error message if it is from the face service (e.g., FastAPI validation)
		errMsg := err.Error()
		if strings.Contains(errMsg, "face service error") {
			parts := strings.SplitN(errMsg, ": ", 2)
			if len(parts) > 1 {
				errMsg = parts[1]
				// If it's a FastAPI detail JSON, let's clean it up
				if strings.Contains(errMsg, `{"detail":`) {
					var detail struct {
						Detail string `json:"detail"`
					}
					if json.Unmarshal([]byte(errMsg), &detail) == nil {
						errMsg = detail.Detail
					}
				}
			}
		}
		return &AttendanceProcessResult{
			Success:        false,
			Message:        errMsg,
			FaceSimilarity: 0.0,
			LocationValid:  locationValid,
			Distance:       distance,
		}, nil
	}

	// 4. Periksa hasil verifikasi
	faceResult := &faceclient.FaceResult{
		Matched:    verifyResp.Matched,
		Similarity: verifyResp.Similarity,
		Confidence: verifyResp.Confidence,
		Threshold:  verifyResp.Threshold,
		RealScore:  verifyResp.RealScore,
		SpoofScore: verifyResp.SpoofScore,
		Message:    verifyResp.Message,
	}

	geoResult := &faceclient.GeoResult{
		IsValid:   locationValid,
		DistanceM: distance,
		Message:   locationMsg,
	}

	if !verifyResp.Matched {
		return &AttendanceProcessResult{
			Success:        false,
			Message:        verifyResp.Message,
			FaceSimilarity: verifyResp.Similarity,
			LocationValid:  locationValid,
			Distance:       distance,
			Face:           faceResult,
			Geo:            geoResult,
		}, nil
	}

	similarity := verifyResp.Similarity

	// 5. verifyOnly=true → hanya return hasil, tidak simpan ke DB
	if verifyOnly {
		fmt.Printf("✅ [VERIFY ONLY] user=%s\n", resolvedUserID)
		return &AttendanceProcessResult{
			Success:        true,
			Message:        "Verifikasi berhasil - klik Konfirmasi untuk menyimpan",
			FaceSimilarity: similarity,
			LocationValid:  locationValid,
			Distance:       distance,
			Face:           faceResult,
			Geo:            geoResult,
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

	// 8. Broadcast real-time event ke client Flutter yang sedang subscribe
	if s.wsHub != nil {
		s.wsHub.BroadcastToUser(resolvedUserID, WSEventAttendanceUpdated, map[string]any{
			"record_type": recordType,
			"status":      attendance.Status,
			"message":     actionMsg + " berhasil dicatat",
		})
		s.wsHub.BroadcastToUser(resolvedUserID, WSEventStatsUpdated, map[string]any{
			"reason": "attendance_changed",
		})
	}

	return &AttendanceProcessResult{
		Success: true, Message: actionMsg + " berhasil dicatat",
		FaceSimilarity: similarity,
		LocationValid:  locationValid,
		Distance:       distance,
		Attendance:     attendance,
		Face:           faceResult,
		Geo:            geoResult,
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

		dateStr := r.Date.In(wib).Format("2006-01-02")
		if r.Date.IsZero() {
			dateStr = from.In(wib).Format("2006-01-02")
		}

		status := mapAttendanceStatusToUI(r.Status)
		// Requirement 2 & 3: Priority leave status
		if r.LeaveRequest != nil {
			status = strings.ToUpper(r.LeaveRequest.TypeName)
			if status == "" {
				status = "IZIN/CUTI"
			}
		} else if r.ID.IsZero() {
			status = "BELUM ABSENSI"
		}

		idStr := r.ID.Hex()
		if r.ID.IsZero() {
			idStr = fmt.Sprintf("absent_%s_%s", r.UserID.Hex(), r.Date.Format("20060102"))
		}

		items = append(items, models.ManagerAttendanceRecord{
			ID:             idStr,
			UserID:         r.UserID.Hex(),
			FullName:       r.User.FullName,
			Email:          r.User.Email,
			PayrollNumber:  r.User.PayrollNumber,
			DepartmentName: r.User.DepartmentName,
			PositionName:   r.User.PositionName,
			Date:           dateStr,
			ClockInTime:    clockIn,
			ClockOutTime:   clockOut,
			Status:         status,
			Location:       location,
			Avatar:         r.User.Avatar,
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

func (s *attendanceService) getTimeOffReductionForToday(ctx context.Context, userID primitive.ObjectID, option string) (float64, error) {
	if s.overtimeRepo == nil {
		return 0, nil
	}

	nowWIB := time.Now().In(wib)
	// Kita bandingkan hanya tanggal (YYYY-MM-DD)
	startOfDayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib)
	endOfDayWIB := startOfDayWIB.Add(24 * time.Hour)

	// Cari overtime yang reward_date nya hari ini dan statusnya GRANTED/USED
	filter := bson.M{
		"employees": bson.M{
			"$elemMatch": bson.M{
				"user_id":              userID,
				"reward.reward_type":   models.OvertimeRewardTypeTimeOff,
				"reward.reward_option": option,
				"reward.reward_date": bson.M{
					"$gte": startOfDayWIB,
					"$lt":  endOfDayWIB,
				},
				"reward.status": bson.M{"$in": []string{models.OvertimeRewardStatusGranted, models.OvertimeRewardStatusUsed}},
			},
		},
	}

	requests, err := s.overtimeRepo.Find(ctx, filter)
	if err != nil {
		return 0, err
	}

	totalReduction := 0.0
	for _, req := range requests {
		for _, emp := range req.Employees {
			if emp.UserID == userID && emp.Reward.RewardType == models.OvertimeRewardTypeTimeOff && emp.Reward.RewardOption == option {
				// Cek apakah reward_date nya memang hari ini (double check)
				if emp.Reward.RewardDate != nil {
					rd := emp.Reward.RewardDate.In(wib)
					if rd.Year() == nowWIB.Year() && rd.Month() == nowWIB.Month() && rd.Day() == nowWIB.Day() {
						totalReduction += req.GetDurationHours()
					}
				}
			}
		}
	}

	return totalReduction, nil
}

func (s *attendanceService) markTimeOffRewardsAsUsed(ctx context.Context, userID primitive.ObjectID) {
	if s.overtimeRepo == nil {
		return
	}

	nowWIB := time.Now().In(wib)
	startOfDayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib)
	endOfDayWIB := startOfDayWIB.Add(24 * time.Hour)

	filter := bson.M{
		"employees": bson.M{
			"$elemMatch": bson.M{
				"user_id":            userID,
				"reward.reward_type": models.OvertimeRewardTypeTimeOff,
				"reward.reward_date": bson.M{
					"$gte": startOfDayWIB,
					"$lt":  endOfDayWIB,
				},
				"reward.status": models.OvertimeRewardStatusGranted,
			},
		},
	}

	requests, _ := s.overtimeRepo.Find(ctx, filter)
	for _, req := range requests {
		for _, emp := range req.Employees {
			if emp.UserID == userID && emp.Reward.RewardType == models.OvertimeRewardTypeTimeOff && emp.Reward.Status == models.OvertimeRewardStatusGranted {
				// Update status to USED
				reward := emp.Reward
				reward.Status = models.OvertimeRewardStatusUsed
				now := time.Now()
				reward.UsedAt = &now
				s.overtimeRepo.UpdateEmployeeReward(ctx, req.ID.Hex(), userID.Hex(), reward)
			}
		}
	}
}
