// pkg/migration/migrations/00X_create_waktu_istirahat.go
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
	version := 11 // TODO: ganti sesuai urutan migration Anda (mis. 3, 4, dst)
	name := "create_waktu_istirahat"
	description := "Create waktu_istirahat collection and indexes"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("waktu_istirahat")

		indexModels := []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "user_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "waktu_mulai", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "waktu_selesai", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "status", Value: 1}},
			},
			// Optional: mencegah duplikat break yang sama untuk user yang sama
			{
				Keys: bson.D{
					{Key: "user_id", Value: 1},
					{Key: "waktu_mulai", Value: 1},
					{Key: "waktu_selesai", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// OPTIONAL seed (hapus kalau tidak perlu seed)
		seed := []interface{}{
			models.WaktuIstirahat{
				ID:           primitive.NewObjectID(),
				UserID:       "seed-user-1",
				WaktuMulai:   time.Now().Add(-2 * time.Hour),
				WaktuSelesai: time.Now().Add(-90 * time.Minute),
				Status:       "aktif",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}

		_, _ = collection.InsertMany(ctx, seed) // optional: abaikan error seed jika tidak mau strict
		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("waktu_istirahat").Drop(ctx)
	}

	return version, name, description, up, down
}