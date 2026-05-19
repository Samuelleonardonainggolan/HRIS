// pkg/migration/migrations/012_create_payrolls.go
package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreatePayrolls() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 19 // ✅ sesuaikan urutan migration Anda
	name := "create_payrolls"
	description := "Create payrolls collection and indexes"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("payrolls")

		// 1) Unique payroll per user per month/year
		uniq := mongo.IndexModel{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "month", Value: 1},
				{Key: "year", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_user_month_year"),
		}

		// 2) Index for filtering
		idxUser := mongo.IndexModel{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_user_id"),
		}
		idxMonthYear := mongo.IndexModel{
			Keys: bson.D{
				{Key: "year", Value: 1},
				{Key: "month", Value: 1},
			},
			Options: options.Index().SetName("idx_year_month"),
		}
		idxPaymentStatus := mongo.IndexModel{
			Keys:    bson.D{{Key: "payment_status", Value: 1}},
			Options: options.Index().SetName("idx_payment_status"),
		}
		idxCreatedAt := mongo.IndexModel{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		}

		// ✅ indexes for renamed fields
		idxTotalDaysPresent := mongo.IndexModel{
			Keys:    bson.D{{Key: "total_days_present", Value: 1}},
			Options: options.Index().SetName("idx_total_days_present"),
		}
		idxTotalLeave := mongo.IndexModel{
			Keys:    bson.D{{Key: "total_leave", Value: 1}},
			Options: options.Index().SetName("idx_total_leave"),
		}
		idxTotalPermission := mongo.IndexModel{
			Keys:    bson.D{{Key: "total_permission", Value: 1}},
			Options: options.Index().SetName("idx_total_permission"),
		}

		_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			uniq,
			idxUser,
			idxMonthYear,
			idxPaymentStatus,
			idxCreatedAt,
			idxTotalDaysPresent,
			idxTotalLeave,
			idxTotalPermission,
		})
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("payrolls").Drop(ctx)
	}

	return version, name, description, up, down
}