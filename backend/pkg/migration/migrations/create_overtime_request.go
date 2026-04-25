// pkg/migration/migrations/00X_create_overtime_requests.go
package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateOvertimeRequests() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	// ✅ sesuaikan nomor version dengan urutan migration Anda
	version := 14
	name := "create_overtime_requests"
	description := "Create overtime_requests collection and indexes"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("overtime_requests")

		indexes := []mongo.IndexModel{
			// Query cepat: history lembur per user per tanggal
			{
				Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "date", Value: -1}},
				Options: options.Index().
					SetName("idx_user_date"),
			},
			// Approval Kepala Departemen
			{
				Keys: bson.D{{Key: "status_kepala_departemen", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().
					SetName("idx_status_kadep_created"),
			},
			// Approval Manager HR
			{
				Keys: bson.D{{Key: "status_manager_hr", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().
					SetName("idx_status_hr_created"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexes)
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("overtime_requests").Drop(ctx)
	}

	return version, name, description, up, down
}