// pkg/migration/migrations/008_create_kategori_pengajuan.go
package migrations

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RequestCategory struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	CategoryName string             `bson:"category_name"`
}

func CreateKategoriPengajuan() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 8
	name := "create_request_category"
	description := "Create request_category collection and seed initial categories"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("request_category") // ✅ renamed collection

		// Create indexes
		indexModels := []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "category_name", Value: 1}}, // ✅ renamed field
				Options: options.Index().
					SetUnique(true).
					SetName("uniq_category_name"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// Seed categories
		categories := []interface{}{
			RequestCategory{
				ID:           primitive.NewObjectID(),
				CategoryName: "Izin",
			},
			RequestCategory{
				ID:           primitive.NewObjectID(),
				CategoryName: "Cuti",
			},
		}

		_, err = collection.InsertMany(ctx, categories)
		if err != nil {
			return fmt.Errorf("failed to insert request_category seed: %w", err)
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("request_category").Drop(ctx) // ✅ renamed
	}

	return version, name, description, up, down
}