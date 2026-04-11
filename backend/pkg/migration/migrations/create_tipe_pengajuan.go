// pkg/migration/migrations/009_create_tipe_pengajuan.go
package migrations

import (
	"context"
	"fmt"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateTipePengajuan() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 9
	name := "create_request_type"
	description := "Create request_type collection and seed initial request types"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("request_type") // ✅ renamed (sesuai gambar)

		// Create indexes
		indexModels := []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "type_name", Value: 1}}, // ✅ renamed
				Options: options.Index().
					SetUnique(true).
					SetName("uniq_type_name"),
			},
			{
				Keys: bson.D{{Key: "request_category_id", Value: 1}}, // ✅ renamed
				Options: options.Index().
					SetName("idx_request_category_id"),
			},
			{
				Keys: bson.D{{Key: "quota_deduction", Value: 1}}, // ✅ renamed
				Options: options.Index().
					SetName("idx_quota_deduction"),
			},
			{
				Keys: bson.D{{Key: "attachment_required", Value: 1}}, // ✅ renamed
				Options: options.Index().
					SetName("idx_attachment_required"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// Fetch request_category: "Izin" dan "Cuti"
		var izin models.RequestCategory
		err = db.Collection("request_category"). // ✅ renamed collection
							FindOne(ctx, bson.M{"category_name": "Izin"}).Decode(&izin)
		if err != nil {
			return fmt.Errorf("request_category 'Izin' not found. Make sure migration request_category ran: %w", err)
		}

		var cuti models.RequestCategory
		err = db.Collection("request_category"). // ✅ renamed collection
							FindOne(ctx, bson.M{"category_name": "Cuti"}).Decode(&cuti)
		if err != nil {
			return fmt.Errorf("request_category 'Cuti' not found. Make sure migration request_category ran: %w", err)
		}

		// Seed request_type
		types := []interface{}{
			// IZIN
			models.RequestType{
				ID:                primitive.NewObjectID(),
				TypeName:          "Izin Sakit",
				RequestCategoryID: izin.ID.Hex(),
				CategoryName:      izin.CategoryName,
				QuotaDeduction:    false,
				AttachmentRequired:true,
			},
			models.RequestType{
				ID:                primitive.NewObjectID(),
				TypeName:          "Izin Menikah",
				RequestCategoryID: izin.ID.Hex(),
				CategoryName:      izin.CategoryName,
				QuotaDeduction:    false,
				AttachmentRequired:false,
			},
			models.RequestType{
				ID:                primitive.NewObjectID(),
				TypeName:          "Izin Baptisan / Akikah",
				RequestCategoryID: izin.ID.Hex(),
				CategoryName:      izin.CategoryName,
				QuotaDeduction:    false,
				AttachmentRequired:false,
			},
			models.RequestType{
				ID:                primitive.NewObjectID(),
				TypeName:          "Izin Anak Manikah",
				RequestCategoryID: izin.ID.Hex(),
				CategoryName:      izin.CategoryName,
				QuotaDeduction:    false,
				AttachmentRequired:false,
			},
			models.RequestType{
				ID:                primitive.NewObjectID(),
				TypeName:          "Izin Dukacita Keluarga Kandung",
				RequestCategoryID: izin.ID.Hex(),
				CategoryName:      izin.CategoryName,
				QuotaDeduction:    false,
				AttachmentRequired:false,
			},

			// CUTI
			models.RequestType{
				ID:                primitive.NewObjectID(),
				TypeName:          "Cuti Tahunan",
				RequestCategoryID: cuti.ID.Hex(),
				CategoryName:      cuti.CategoryName,
				QuotaDeduction:    true,
				AttachmentRequired:false,
			},
			models.RequestType{
				ID:                primitive.NewObjectID(),
				TypeName:          "Cuti Khusus",
				RequestCategoryID: cuti.ID.Hex(),
				CategoryName:      cuti.CategoryName,
				QuotaDeduction:    false,
				AttachmentRequired:true,
			},
		}

		_, err = collection.InsertMany(ctx, types)
		if err != nil {
			return fmt.Errorf("failed to insert request_type seed: %w", err)
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("request_type").Drop(ctx) // ✅ renamed
	}

	return version, name, description, up, down
}