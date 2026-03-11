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
	attendance.CreatedAt = time.Now()
	attendance.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, attendance)
	return err
}

// FindTodayByUserID implements AttendanceRepository.FindTodayByUserID
func (r *attendanceRepository) FindTodayByUserID(ctx context.Context, userID primitive.ObjectID) (*models.Attendance, error) {
	startOfDay := time.Now().Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := bson.M{
		"user_id": userID,
		"date": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
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
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	filter := bson.M{
		"user_id": userID,
		"date": bson.M{
			"$gte": startOfMonth,
			"$lt":  endOfMonth,
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

		clockInStr := "--:--"
		if att.ClockInTime != nil {
			clockInStr = att.ClockInTime.Format("15:04")
		}
		clockOutStr := "--:--"
		if att.ClockOutTime != nil {
			clockOutStr = att.ClockOutTime.Format("15:04")
		}

		records = append(records, models.AttendanceResponse{
			ID:            att.ID.Hex(),
			Date:          att.Date.Format("2006-01-02"),
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
