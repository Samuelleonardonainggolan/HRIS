// pkg/migration/migrations/00X_create_jam_kerja.go
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
	// Sesuaikan nomor version sesuai urutan migration Anda
	version := 7
	name := "create_jam_kerja"
	description := "Create jam_kerja collection, indexes, and seed default schedules for all active users"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("jam_kerja")

		// 1) Indexes
		uniqueIndex := mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "day_of_week", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_user_day_of_week"),
		}

		userActiveIndex := mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "is_active", Value: 1},
			},
			Options: options.Index().SetName("idx_user_active"),
		}

		_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			uniqueIndex,
			userActiveIndex,
		})
		if err != nil {
			return err
		}

		// 2) Seed: buat default jam kerja untuk semua user aktif (yang sudah ada)
		// Asumsi field users: _id (ObjectId) dan is_active (bool)
		userColl := db.Collection("users")

		cur, err := userColl.Find(ctx, bson.M{"is_active": true}, options.Find().SetProjection(bson.M{"_id": 1}))
		if err != nil {
			return err
		}
		defer cur.Close(ctx)

		// Default: Senin–Jumat 09:00–18:00
		days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

		// Pakai tanggal dummy agar tipe tetap date (sesuai tabel)
		// 2000-01-01 09:00 dan 18:00 (date-nya tidak penting, yang penting jam-nya)
		start := time.Date(2000, 1, 1, 9, 0, 0, 0, time.Local)
		end := time.Date(2000, 1, 1, 18, 0, 0, 0, time.Local)

		now := time.Now()

		type userIDOnly struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		var bulk []mongo.WriteModel

		for cur.Next(ctx) {
			var u userIDOnly
			if err := cur.Decode(&u); err != nil {
				return err
			}

			for _, day := range days {
				// Upsert supaya aman kalau migration dijalankan ulang
				filter := bson.M{
					"user_id":     u.ID,
					"day_of_week": day,
				}

				update := bson.M{
					"$setOnInsert": bson.M{
						"user_id":     u.ID,
						"day_of_week": day,
						"start_time":  start,
						"end_time":    end,
						"is_active":   true,
						"created_at":  now,
						"updated_at":  now,
					},
				}

				model := mongo.NewUpdateOneModel().
					SetFilter(filter).
					SetUpdate(update).
					SetUpsert(true)

				bulk = append(bulk, model)
			}
		}

		if err := cur.Err(); err != nil {
			return err
		}

		if len(bulk) > 0 {
			_, err := collection.BulkWrite(ctx, bulk, options.BulkWrite().SetOrdered(false))
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