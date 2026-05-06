package migrations

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateOvertimeRequests() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 14 // atau sesuaikan dengan nomor migration Anda
	name := "create_overtime_requests"
	description := "Create overtime_requests collection and indexes (support employess status)"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("overtime_requests")

		indexes := []mongo.IndexModel{
			// Percepat cari/riwayat lembur per karyawan (subdoc)
			{
				Keys: bson.D{{Key: "employees.user_id", Value: 1}},
				Options: options.Index().
					SetName("idx_employees_userid"),
			},
			// Filter per departemen, status, tanggal
			{
				Keys: bson.D{{Key: "department_id", Value: 1}, {Key: "date", Value: -1}},
				Options: options.Index().
					SetName("idx_dept_date"),
			},
			// Untuk filter submitted/published/draft
			{
				Keys: bson.D{{Key: "status", Value: 1}, {Key: "date", Value: -1}},
				Options: options.Index().
					SetName("idx_status_date"),
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