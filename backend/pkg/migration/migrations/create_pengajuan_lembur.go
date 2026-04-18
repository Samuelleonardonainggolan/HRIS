// pkg/migration/migrations/014_create_overtime_request.go
package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateOvertimeRequest() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 14 // ✅ sesuaikan urutan migration Anda
	name := "create_overtime_request"
	description := "Create overtime_request collection and indexes"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("overtime_request")

		indexModels := []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}},
				Options: options.Index().SetName("idx_user_id"),
			},
			{
				Keys:    bson.D{{Key: "date", Value: 1}},
				Options: options.Index().SetName("idx_date"),
			},
			{
				Keys:    bson.D{{Key: "start_time", Value: 1}},
				Options: options.Index().SetName("idx_start_time"),
			},
			{
				Keys:    bson.D{{Key: "end_time", Value: 1}},
				Options: options.Index().SetName("idx_end_time"),
			},
			{
				Keys:    bson.D{{Key: "status_kepala_departemen", Value: 1}},
				Options: options.Index().SetName("idx_status_kepala_departemen"),
			},
			{
				Keys:    bson.D{{Key: "kepala_departemen_id", Value: 1}},
				Options: options.Index().SetName("idx_kepala_departemen_id"),
			},
			{
				Keys:    bson.D{{Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_created_at"),
			},

			// Optional unique guard: 1 overtime per user per date+time range
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
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("overtime_request").Drop(ctx)
	}

	return version, name, description, up, down
}