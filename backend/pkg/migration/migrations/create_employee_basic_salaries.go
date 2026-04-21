// pkg/migration/migrations/016_create_employee_basic_salaries.go
package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateEmployeeBasicSalaries() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 16 // ✅ sesuaikan urutan migration Anda
	name := "create_employee_basic_salaries"
	description := "Create employee_basic_salaries collection and indexes (basic_salary, effective_from, is_active)"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("employee_basic_salaries")

		// ✅ Unique active salary per user (partial unique index)
		uniqActivePerUser := mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().
				SetName("uniq_active_salary_per_user").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"is_active": true}),
		}

		indexModels := []mongo.IndexModel{
			uniqActivePerUser,
			{
				Keys:    bson.D{{Key: "user_id", Value: 1}},
				Options: options.Index().SetName("idx_user_id"),
			},
			{
				Keys:    bson.D{{Key: "is_active", Value: 1}},
				Options: options.Index().SetName("idx_is_active"),
			},
			{
				Keys:    bson.D{{Key: "effective_from", Value: -1}},
				Options: options.Index().SetName("idx_effective_from"),
			},
			{
				Keys:    bson.D{{Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_created_at"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("employee_basic_salaries").Drop(ctx)
	}

	return version, name, description, up, down
}