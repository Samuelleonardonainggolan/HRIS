// internal/service/attendance_service.go
package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/faceclient"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AttendanceService interface {
	ClockIn(ctx context.Context, userID string, latitude, longitude float64, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error)
	ClockOut(ctx context.Context, userID string, latitude, longitude float64, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error)
	GetTodayAttendance(ctx context.Context, userID string) (*models.Attendance, error)
	GetMonthlyAttendance(ctx context.Context, userID string, month, year int) (*models.MonthlyAttendanceResponse, error)
	ProcessAttendanceWithFace(ctx context.Context, userID string, photo []byte, filename string, latitude, longitude float64, recordType string) (*AttendanceProcessResult, error)
}

type attendanceService struct {
	attendanceRepo    repository.AttendanceRepository
	userRepo          repository.UserRepository          // Gunakan interface
	faceEmbeddingRepo repository.FaceEmbeddingRepository // Gunakan interface
	faceClient        *faceclient.Client
	officeLat         float64
	officeLng         float64
	radiusMeters      float64
}

type AttendanceProcessResult struct {
	Success        bool               `json:"success"`
	Message        string             `json:"message"`
	FaceSimilarity float64            `json:"face_similarity"`
	LocationValid  bool               `json:"location_valid"`
	Distance       float64            `json:"distance_m"`
	Attendance     *models.Attendance `json:"attendance,omitempty"`
}

func NewAttendanceService(
	attendanceRepo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	faceEmbeddingRepo repository.FaceEmbeddingRepository,
	faceClient *faceclient.Client,
) AttendanceService {
	return &attendanceService{
		attendanceRepo:    attendanceRepo,
		userRepo:          userRepo,
		faceEmbeddingRepo: faceEmbeddingRepo,
		faceClient:        faceClient,
		officeLat:         2.3561,  // IT Del Sitoluama
		officeLng:         99.1431, // IT Del Sitoluama
		radiusMeters:      10000,   // 10 km radius
	}
}

func (s *attendanceService) ClockIn(ctx context.Context, userID string, latitude, longitude float64, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Check if already clocked in today
	existing, _ := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if existing != nil && existing.ClockInTime != nil {
		return nil, errors.New("already clocked in today")
	}

	now := time.Now()
	location := models.GeoLocation{
		Latitude:  latitude,
		Longitude: longitude,
	}

	// Calculate status based on time
	status := models.StatusOnTime
	standardStart := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
	if now.After(standardStart.Add(15 * time.Minute)) {
		status = models.StatusLate
	}

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

func (s *attendanceService) ClockOut(ctx context.Context, userID string, latitude, longitude float64, photo []byte, filename string, faceSimilarity float64) (*models.Attendance, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Find today's attendance
	attendance, err := s.attendanceRepo.FindTodayByUserID(ctx, userObjID)
	if err != nil {
		return nil, errors.New("no clock in record found for today")
	}

	if attendance.ClockOutTime != nil {
		return nil, errors.New("already clocked out today")
	}

	now := time.Now()
	location := models.GeoLocation{
		Latitude:  latitude,
		Longitude: longitude,
	}

	// Calculate work hours
	if attendance.ClockInTime != nil {
		workDuration := now.Sub(*attendance.ClockInTime)
		workHours := workDuration.Hours()

		// Calculate overtime (more than 9 hours including 1 hour break)
		overtimeHours := 0.0
		if workHours > 9.0 {
			overtimeHours = workHours - 9.0
		}

		// Update status if worked overtime
		status := attendance.Status
		if overtimeHours > 0 {
			status = models.StatusOvertime
		}

		attendance.WorkHours = workHours
		attendance.OvertimeHours = overtimeHours
		attendance.Status = status
	}

	attendance.ClockOutTime = &now
	attendance.ClockOutPhoto = filename
	attendance.ClockOutLocation = location
	attendance.FaceSimilarity = faceSimilarity
	attendance.UpdatedAt = now

	// Update in database
	err = s.attendanceRepo.UpdateClockOut(ctx, attendance.ID, now, filename, location)
	if err != nil {
		return nil, err
	}

	// Update work hours
	err = s.attendanceRepo.UpdateWorkHours(ctx, attendance.ID, attendance.WorkHours, attendance.OvertimeHours, attendance.Status)
	if err != nil {
		return nil, err
	}

	return attendance, nil
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
) (*AttendanceProcessResult, error) {

	// 1. Validate location
	distance := s.calculateDistance(latitude, longitude, s.officeLat, s.officeLng)
	locationValid := distance <= s.radiusMeters

	// 2. Get user's face embedding from database
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil || faceEmbedding == nil {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Face not registered. Please register your face first.",
			LocationValid: locationValid,
			Distance:      distance,
		}, nil
	}

	// 3. Extract embedding from current photo
	currentEmbedding, err := s.faceClient.ExtractEmbedding(userID, photo, filename)
	if err != nil {
		return &AttendanceProcessResult{
			Success:       false,
			Message:       "Failed to extract face: " + err.Error(),
			LocationValid: locationValid,
			Distance:      distance,
		}, nil
	}

	// 4. Calculate similarity between stored and current embedding
	similarity := s.cosineSimilarity(currentEmbedding, faceEmbedding.FaceEmbedding)
	faceValid := similarity >= 0.6 // Threshold 60%

	if !faceValid {
		return &AttendanceProcessResult{
			Success:        false,
			Message:        "Face does not match registered face",
			FaceSimilarity: similarity,
			LocationValid:  locationValid,
			Distance:       distance,
		}, nil
	}

	if !locationValid {
		return &AttendanceProcessResult{
			Success:        false,
			Message:        "Location outside allowed radius",
			FaceSimilarity: similarity,
			LocationValid:  locationValid,
			Distance:       distance,
		}, nil
	}

	// 5. Process attendance
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
	now := time.Now()
	faceEmbedding.LastVerifiedAt = &now
	faceEmbedding.VerificationCount++
	faceEmbedding.UpdatedAt = now
	s.faceEmbeddingRepo.Update(ctx, faceEmbedding)

	return &AttendanceProcessResult{
		Success:        true,
		Message:        "Attendance processed successfully",
		FaceSimilarity: similarity,
		LocationValid:  locationValid,
		Distance:       distance,
		Attendance:     attendance,
	}, nil
}

func (s *attendanceService) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // Earth radius in meters
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
