// pkg/migration/migrations/007_create_jam_kerja.go
package migrations

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateJamKerja() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 7
	name := "create_jam_kerja"
	description := "Create jam_kerja collection, indexes, and seed from attendances"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("jam_kerja")

		// =========================
		// 1) INDEXES
		// =========================

		// Unique: 1 attendance => 1 jam_kerja
		uniqAttendance := mongo.IndexModel{
			Keys: bson.D{{Key: "attendance_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_attendance_id"),
		}

		// Query helpers
		idxUserStart := mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "start_time", Value: 1},
			},
			Options: options.Index().SetName("idx_user_start_time"),
		}

		idxUserAttendance := mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "attendance_id", Value: 1},
			},
			Options: options.Index().SetName("idx_user_attendance"),
		}

		_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			uniqAttendance,
			idxUserStart,
			idxUserAttendance,
		})
		if err != nil {
			return err
		}

		// =========================
		// 2) SEEDER (FROM attendances)
		// =========================
		attendanceColl := db.Collection("attendances")

		// Ambil semua attendance (kalau ingin batasi, bisa filter date range)
		cur, err := attendanceColl.Find(
			ctx,
			bson.M{}, // or: bson.M{"status": bson.M{"$in": []string{"present", "late"}}}
			options.Find().SetProjection(bson.M{"_id": 1, "user_id": 1, "date": 1}),
		)
		if err != nil {
			return err
		}
		defer cur.Close(ctx)

		type attendanceSeed struct {
			ID     primitive.ObjectID `bson:"_id"`
			UserID primitive.ObjectID `bson:"user_id"`
			Date   time.Time          `bson:"date"`
		}

		now := time.Now()

		// Default jam kerja: 09:00 - 18:00 di tanggal attendance
		defaultStartHour := 9
		defaultEndHour := 18

		var writes []mongo.WriteModel

		for cur.Next(ctx) {
			var a attendanceSeed
			if err := cur.Decode(&a); err != nil {
				return err
			}

			// Normalisasi: start/end harus “tanggal attendance” + jam default
			y, m, d := a.Date.Date()
			loc := a.Date.Location()
			if loc == nil {
				loc = time.Local
			}

			startTime := time.Date(y, m, d, defaultStartHour, 0, 0, 0, loc)
			endTime := time.Date(y, m, d, defaultEndHour, 0, 0, 0, loc)

			filter := bson.M{"attendance_id": a.ID}

			update := bson.M{
				"$setOnInsert": bson.M{
					"user_id":       a.UserID,
					"attendance_id": a.ID,
					"start_time":    startTime,
					"end_time":      endTime,
					"is_active":     true,
					"created_at":    now,
					"updated_at":    now,
				},
			}

			writes = append(
				writes,
				mongo.NewUpdateOneModel().
					SetFilter(filter).
					SetUpdate(update).
					SetUpsert(true),
			)
		}

		if err := cur.Err(); err != nil {
			return err
		}

		if len(writes) > 0 {
			_, err := collection.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false))
			if err != nil {
				return err
			}
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("jam_kerja").Drop(ctx)
	}

	return version, name, description, up, down
}
