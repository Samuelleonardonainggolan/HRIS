// pkg/database/repository/attendance_repository.go
package repository

import (
	"context"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// wib adalah timezone WIB (UTC+7).
// MongoDB menyimpan semua waktu dalam UTC. Konversi ke WIB wajib dilakukan
// sebelum mem-format waktu menjadi string yang dikirim ke client (Flutter).
var wib = time.FixedZone("WIB", 7*60*60)

// AttendanceRepository interface
type AttendanceRepository interface {
	Create(ctx context.Context, attendance *models.Attendance) error
	FindTodayByUserID(ctx context.Context, userID primitive.ObjectID) (*models.Attendance, error)
	UpdateClockOut(ctx context.Context, id primitive.ObjectID, clockOutTime time.Time, photo string, location models.GeoLocation) error
	UpdateWorkHours(ctx context.Context, id primitive.ObjectID, workHours, overtimeHours float64, status models.AttendanceStatus) error
	FindByUserIDAndMonth(ctx context.Context, userID primitive.ObjectID, year, month int) ([]models.Attendance, error)
	GetMonthlySummary(ctx context.Context, userID primitive.ObjectID, year, month int) (*models.MonthlyAttendanceResponse, error)
	CreateIndexes(ctx context.Context) error
}

// attendanceRepository struct implementing the interface
type attendanceRepository struct {
	collection *mongo.Collection
}

// NewAttendanceRepository creates a new attendance repository
func NewAttendanceRepository(db *mongo.Database) AttendanceRepository {
	return &attendanceRepository{
		collection: db.Collection("attendances"),
	}
}

// Create implements AttendanceRepository.Create
func (r *attendanceRepository) Create(ctx context.Context, attendance *models.Attendance) error {
	attendance.Date = attendance.Date.UTC()
	attendance.CreatedAt = time.Now().UTC()
	attendance.UpdatedAt = time.Now().UTC()
	_, err := r.collection.InsertOne(ctx, attendance)
	return err
}

// FindTodayByUserID implements AttendanceRepository.FindTodayByUserID
func (r *attendanceRepository) FindTodayByUserID(ctx context.Context, userID primitive.ObjectID) (*models.Attendance, error) {
	// ✅ FIX: Hitung start/end hari berdasarkan WIB agar query MongoDB konsisten
	// dengan hari yang dimaksud pengguna, bukan hari UTC yang bisa beda 7 jam.
	nowWIB := time.Now().In(wib)
	startOfDayWIB := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wib)
	endOfDayWIB := startOfDayWIB.Add(24 * time.Hour)

	// Konversi ke UTC untuk query MongoDB (MongoDB menyimpan dalam UTC)
	startOfDayUTC := startOfDayWIB.UTC()
	endOfDayUTC := endOfDayWIB.UTC()

	filter := bson.M{
		"user_id": userID,
		"date": bson.M{
			"$gte": startOfDayUTC,
			"$lt":  endOfDayUTC,
		},
	}

	var attendance models.Attendance
	err := r.collection.FindOne(ctx, filter).Decode(&attendance)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &attendance, nil
}

// UpdateClockOut implements AttendanceRepository.UpdateClockOut
func (r *attendanceRepository) UpdateClockOut(ctx context.Context, id primitive.ObjectID, clockOutTime time.Time, photo string, location models.GeoLocation) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"clock_out_time":     clockOutTime,
			"clock_out_photo":    photo,
			"clock_out_location": location,
			"updated_at":         time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateWorkHours implements AttendanceRepository.UpdateWorkHours
func (r *attendanceRepository) UpdateWorkHours(ctx context.Context, id primitive.ObjectID, workHours, overtimeHours float64, status models.AttendanceStatus) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"work_hours":     workHours,
			"overtime_hours": overtimeHours,
			"status":         status,
			"updated_at":     time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// FindByUserIDAndMonth implements AttendanceRepository.FindByUserIDAndMonth
func (r *attendanceRepository) FindByUserIDAndMonth(ctx context.Context, userID primitive.ObjectID, year, month int) ([]models.Attendance, error) {
	// ✅ FIX: Gunakan WIB untuk menentukan range bulan, lalu konversi ke UTC untuk query MongoDB.
	startOfMonthWIB := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, wib)
	endOfMonthWIB := startOfMonthWIB.AddDate(0, 1, 0)

	filter := bson.M{
		"user_id": userID,
		"date": bson.M{
			"$gte": startOfMonthWIB.UTC(),
			"$lt":  endOfMonthWIB.UTC(),
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "date", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var attendances []models.Attendance
	if err = cursor.All(ctx, &attendances); err != nil {
		return nil, err
	}
	return attendances, nil
}

// GetMonthlySummary implements AttendanceRepository.GetMonthlySummary
func (r *attendanceRepository) GetMonthlySummary(ctx context.Context, userID primitive.ObjectID, year, month int) (*models.MonthlyAttendanceResponse, error) {
	attendances, err := r.FindByUserIDAndMonth(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}

	var totalHours, overtimeHours float64
	var records []models.AttendanceResponse

	for _, att := range attendances {
		totalHours += att.WorkHours
		overtimeHours += att.OvertimeHours

		// ✅ FIX: Konversi semua timestamp dari UTC ke WIB sebelum format string.
		// Ini adalah root cause dari masalah jam beda 7 jam (UTC vs WIB).
		clockInStr := "--:--"
		if att.ClockInTime != nil {
			clockInStr = att.ClockInTime.In(wib).Format("15:04")
		}
		clockOutStr := "--:--"
		if att.ClockOutTime != nil {
			clockOutStr = att.ClockOutTime.In(wib).Format("15:04")
		}

		// ✅ FIX: Format tanggal juga dalam WIB agar konsisten
		records = append(records, models.AttendanceResponse{
			ID:            att.ID.Hex(),
			Date:          att.Date.In(wib).Format("2006-01-02"),
			ClockInTime:   clockInStr,
			ClockOutTime:  clockOutStr,
			Status:        att.Status,
			WorkHours:     att.WorkHours,
			OvertimeHours: att.OvertimeHours,
		})
	}

	return &models.MonthlyAttendanceResponse{
		Month:         time.Month(month).String(),
		Year:          year,
		TotalDays:     len(attendances),
		TotalHours:    totalHours,
		OvertimeHours: overtimeHours,
		Records:       records,
	}, nil
}

// CreateIndexes implements AttendanceRepository.CreateIndexes
func (r *attendanceRepository) CreateIndexes(ctx context.Context) error {
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "date", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexModels)
	return err
}
