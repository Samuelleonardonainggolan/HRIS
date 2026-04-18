// pkg/migration/migrations/013_create_leave_balance.go
package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateLeaveBalance() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 13 // ✅ sesuaikan urutan migration Anda
	name := "create_leave_balance"
	description := "Create leave_balance collection and indexes"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("leave_balance")

		uniqUserYear := mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "year", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_user_year"),
		}

		idxUser := mongo.IndexModel{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_user_id"),
		}

		idxYear := mongo.IndexModel{
			Keys:    bson.D{{Key: "year", Value: 1}},
			Options: options.Index().SetName("idx_year"),
		}

		idxRemaining := mongo.IndexModel{
			Keys:    bson.D{{Key: "remaining_kuota", Value: 1}},
			Options: options.Index().SetName("idx_remaining_kuota"),
		}

		idxCreatedAt := mongo.IndexModel{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		}

		_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			uniqUserYear,
			idxUser,
			idxYear,
			idxRemaining,
			idxCreatedAt,
		})
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("leave_balance").Drop(ctx)
	}

	return version, name, description, up, down
}