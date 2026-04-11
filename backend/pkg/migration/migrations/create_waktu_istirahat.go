// pkg/migration/migrations/011_create_break_time.go
package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateWaktuIstirahat() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 11
	name := "create_break_time"
	description := "Create break_time collection and indexes"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("break_time") // ✅ renamed collection

		indexModels := []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}},
				Options: options.Index().SetName("idx_user_id"),
			},
			{
				Keys:    bson.D{{Key: "date", Value: 1}}, // ✅ new field
				Options: options.Index().SetName("idx_date"),
			},
			{
				Keys:    bson.D{{Key: "start_time", Value: 1}}, // ✅ renamed
				Options: options.Index().SetName("idx_start_time"),
			},
			{
				Keys:    bson.D{{Key: "end_time", Value: 1}}, // ✅ renamed
				Options: options.Index().SetName("idx_end_time"),
			},
			{
				Keys:    bson.D{{Key: "status", Value: 1}},
				Options: options.Index().SetName("idx_status"),
			},
			// Optional: mencegah duplikat break yang sama untuk user yang sama (di tanggal yang sama)
			{
				Keys: bson.D{
					{Key: "user_id", Value: 1},
					{Key: "date", Value: 1},
					{Key: "start_time", Value: 1},
					{Key: "end_time", Value: 1},
				},
				Options: options.Index().
					SetUnique(true).
					SetName("uniq_user_date_start_end"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// OPTIONAL seed (hapus kalau tidak perlu seed)
		now := time.Now()
		y, m, d := now.Date()
		loc := now.Location()
		if loc == nil {
			loc = time.Local
		}

		// date: simpan tanggal saja (00:00:00)
		seedDate := time.Date(y, m, d, 0, 0, 0, 0, loc)

		seed := []interface{}{
			models.BreakTime{
				ID:        primitive.NewObjectID(),
				UserID:    "seed-user-1",
				Date:      seedDate,
				StartTime: now.Add(-2 * time.Hour),
				EndTime:   now.Add(-90 * time.Minute),
				Status:    "aktif",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		_, _ = collection.InsertMany(ctx, seed) // optional: abaikan error seed jika tidak mau strict
		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("break_time").Drop(ctx) // ✅ renamed
	}

	return version, name, description, up, down
}