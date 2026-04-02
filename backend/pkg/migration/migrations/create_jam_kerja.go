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
	description := "Create jam_kerja collection, indexes, and seed unique per user"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("jam_kerja")

		// =========================
		// 1) INDEXES
		// =========================
		uniqUser := mongo.IndexModel{
			Keys: bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_user_id"),
		}

		idxActive := mongo.IndexModel{
			Keys:    bson.D{{Key: "aktif", Value: 1}},
			Options: options.Index().SetName("idx_aktif"),
		}

		// optional untuk filter by hari kerja
		idxHari := mongo.IndexModel{
			Keys:    bson.D{{Key: "hari_kerja", Value: 1}},
			Options: options.Index().SetName("idx_hari_kerja"),
		}

		_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			uniqUser,
			idxActive,
			idxHari,
		})
		if err != nil {
			return err
		}

		// =========================
		// 2) SEEDER (UNIQUE PER USER)
		// =========================
		userColl := db.Collection("users")

		cur, err := userColl.Find(
			ctx,
			bson.M{"is_active": true}, // optional: hanya user aktif
			options.Find().SetProjection(bson.M{"_id": 1}),
		)
		if err != nil {
			return err
		}
		defer cur.Close(ctx)

		type userSeed struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		now := time.Now()

		// default: Senin-Jumat, 09:00-18:00 (tanpa tanggal spesifik)
		defaultHari := []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat"}

		// pakai tanggal "hari ini" untuk menyimpan jam, karena type date butuh full datetime
		// nanti saat dipakai cukup ambil HH:mm
		y, m, d := now.Date()
		loc := now.Location()
		if loc == nil {
			loc = time.Local
		}

		defaultMulai := time.Date(y, m, d, 9, 0, 0, 0, loc)
		defaultSelesai := time.Date(y, m, d, 18, 0, 0, 0, loc)

		var writes []mongo.WriteModel

		for cur.Next(ctx) {
			var u userSeed
			if err := cur.Decode(&u); err != nil {
				return err
			}

			filter := bson.M{"user_id": u.ID}

			update := bson.M{
				"$setOnInsert": bson.M{
					"user_id":       u.ID,
					"hari_kerja":    defaultHari,
					"waktu_mulai":   defaultMulai,
					"waktu_selesai": defaultSelesai,
					"aktif":         true,
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