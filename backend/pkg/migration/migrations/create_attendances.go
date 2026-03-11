// pkg/migration/migrations/006_create_attendances.go
package migrations

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateAttendances() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 6
	name := "create_attendances"
	description := "Create attendances collection with indexes and seed sample data"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("attendances")

		// Create indexes
		indexModels := []mongo.IndexModel{
			// Compound unique index for user_id + date
			{
				Keys: bson.D{
					{Key: "user_id", Value: 1},
					{Key: "date", Value: 1},
				},
				Options: options.Index().
					SetUnique(true).
					SetName("idx_user_date_unique"),
			},
			// Index for user_id queries
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}},
				Options: options.Index().SetName("idx_user_id"),
			},
			// Index for date queries
			{
				Keys:    bson.D{{Key: "date", Value: -1}},
				Options: options.Index().SetName("idx_date"),
			},
			// Index for status queries
			{
				Keys:    bson.D{{Key: "status", Value: 1}},
				Options: options.Index().SetName("idx_status"),
			},
			// Compound index for user + date range queries
			{
				Keys: bson.D{
					{Key: "user_id", Value: 1},
					{Key: "date", Value: -1},
				},
				Options: options.Index().SetName("idx_user_date_range"),
			},
			// Index for created_at (for sorting)
			{
				Keys:    bson.D{{Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_created_at"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return err
		}

		log.Println("   ✅ Attendance indexes created")

		// Get all users for seeding
		var users []models.User
		userCursor, err := db.Collection("users").Find(ctx, bson.M{"role": bson.M{"$ne": "manager_hr"}})
		if err != nil {
			log.Println("   ⚠️  No users found, skipping attendance seeding")
			return nil
		}
		defer userCursor.Close(ctx)

		if err = userCursor.All(ctx, &users); err != nil {
			return err
		}

		if len(users) == 0 {
			log.Println("   ℹ️  No users to seed attendance data")
			return nil
		}

		// Seed sample attendance data for current month
		attendances := []interface{}{}
		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		// Create attendance for last 30 days
		for i := 0; i < 30; i++ {
			date := startOfMonth.AddDate(0, 0, -i)

			// Skip weekends for demo (Saturday = 6, Sunday = 0)
			if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
				continue
			}

			// Create attendance for each user
			for userIdx, user := range users {
				// Random attendance patterns
				shouldBePresent := i%10 != 0 // 90% attendance rate

				if !shouldBePresent {
					// Absent record
					attendances = append(attendances, models.Attendance{
						ID:         primitive.NewObjectID(),
						UserID:     user.ID,
						Date:       date,
						Status:     models.StatusAbsent,
						WorkHours:  0,
						CreatedAt:  date,
						UpdatedAt:  date,
					})
					continue
				}

				// Present record
				isLate := i%5 == 0 // 20% late rate

				// Clock in time (8:00 AM or 8:15 AM if late)
				clockInHour := 8
				clockInMinute := 0
				status := models.StatusOnTime

				if isLate {
					clockInMinute = 15 + (i % 30) // Late by 15-45 minutes
					status = models.StatusLate
				}

				clockInTime := time.Date(date.Year(), date.Month(), date.Day(), clockInHour, clockInMinute, 0, 0, date.Location())

				// Clock out time (5:00 PM base)
				clockOutHour := 17
				clockOutMinute := 0
				overtimeHours := 0.0

				// Some days have overtime
				if i%7 == 0 {
					clockOutHour = 19 // Work until 7 PM
					overtimeHours = 2.0
				}

				clockOutTime := time.Date(date.Year(), date.Month(), date.Day(), clockOutHour, clockOutMinute, 0, 0, date.Location())

				// Calculate work hours
				workDuration := clockOutTime.Sub(clockInTime)
				workHours := workDuration.Hours() - 1.0 // Subtract 1 hour lunch break

				// Example location (Jakarta area) - you can randomize this
				clockInLocation := models.GeoLocation{
					Latitude:  -6.2088 + (float64(i%10) * 0.0001),  // Slight variation
					Longitude: 106.8456 + (float64(i%10) * 0.0001), // Slight variation
					Address:   "Hotel Labosta, Jakarta",
				}

				clockOutLocation := models.GeoLocation{
					Latitude:  -6.2088 + (float64(i%10) * 0.0001),
					Longitude: 106.8456 + (float64(i%10) * 0.0001),
					Address:   "Hotel Labosta, Jakarta",
				}

				// ✅ Generate photo URLs (simulated)
				// Format: /uploads/attendance/USERID_DATE_clockin.jpg
				dateStr := date.Format("20060102")
				clockInPhoto := fmt.Sprintf("/uploads/attendance/%s_%s_clockin.jpg", user.ID.Hex(), dateStr)
				clockOutPhoto := fmt.Sprintf("/uploads/attendance/%s_%s_clockout.jpg", user.ID.Hex(), dateStr)

				// Face similarity (95-99%)
				faceSimilarity := 0.95 + (float64((userIdx+i)%5) * 0.01)

				attendances = append(attendances, models.Attendance{
					ID:               primitive.NewObjectID(),
					UserID:           user.ID,
					Date:             date,
					ClockInTime:      &clockInTime,
					ClockOutTime:     &clockOutTime,
					ClockInPhoto:     clockInPhoto,  // ✅ Added
					ClockOutPhoto:    clockOutPhoto, // ✅ Added
					ClockInLocation:  clockInLocation,
					ClockOutLocation: clockOutLocation,
					Status:           status,
					WorkHours:        workHours,
					OvertimeHours:    overtimeHours,
					FaceSimilarity:   faceSimilarity,
					CreatedAt:        clockInTime,
					UpdatedAt:        clockOutTime,
				})
			}
		}

		if len(attendances) > 0 {
			_, err = collection.InsertMany(ctx, attendances)
			if err != nil {
				return err
			}
			log.Printf("   ✅ Seeded %d attendance records for %d users (with photos)", len(attendances), len(users))
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("attendances").Drop(ctx)
	}

	return version, name, description, up, down
}