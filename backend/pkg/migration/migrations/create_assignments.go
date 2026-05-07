package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateAssignments() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	// ✅ sesuaikan version sesuai urutan migration Anda
	version := 18
	name := "create_assignments"
	description := "Create assignments collection and indexes (employee confirm + replacement day off reward)"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("assignments")

		indexes := []mongo.IndexModel{
			// Query cepat: list penugasan per departemen + tanggal
			{
				Keys: bson.D{{Key: "department_id", Value: 1}, {Key: "date", Value: -1}},
				Options: options.Index().
					SetName("idx_department_date"),
			},
			// Filter status (draft/submitted/published/cancelled)
			{
				Keys: bson.D{{Key: "status", Value: 1}, {Key: "date", Value: -1}},
				Options: options.Index().
					SetName("idx_status_date"),
			},
			// Query cepat: "penugasan saya" per karyawan
			{
				Keys: bson.D{{Key: "employees.user_id", Value: 1}, {Key: "date", Value: -1}},
				Options: options.Index().
					SetName("idx_employees_userid_date"),
			},
			// Query cepat: penugasan yang dibuat oleh kepala departemen tertentu
			{
				Keys: bson.D{{Key: "requested_by_id", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().
					SetName("idx_requested_by_created"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexes)
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("assignments").Drop(ctx)
	}

	return version, name, description, up, down
}