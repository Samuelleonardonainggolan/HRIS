// pkg/migration/migrations/020_create_notifications.go
package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateNotifications() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 20
	name := "create_notifications"
	description := "Create notifications collection and indexes"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("notifications")

		// 1) Index on user_id for filtering user notifications
		idxUser := mongo.IndexModel{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_user_id"),
		}

		// 2) Compound index on user_id and is_read to fetch unread/read quickly
		idxUserIsRead := mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "is_read", Value: 1},
			},
			Options: options.Index().SetName("idx_user_id_is_read"),
		}

		// 3) Index on created_at in descending order for feed sorting
		idxCreatedAt := mongo.IndexModel{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		}

		_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			idxUser,
			idxUserIsRead,
			idxCreatedAt,
		})
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("notifications").Drop(ctx)
	}

	return version, name, description, up, down
}
