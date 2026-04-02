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

type KategoriPengajuan struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	NamaKategori string             `bson:"nama_kategori"`
}

func CreateKategoriPengajuan() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 8
	name := "create_kategori_pengajuan"
	description := "Create kategori_pengajuan collection and seed initial categories"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("kategori_pengajuan")

		// Create indexes
		indexModels := []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "nama_kategori", Value: 1}},
				Options: options.Index().
					SetUnique(true).
					SetName("uniq_nama_kategori"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// Seed categories (sesuaikan dengan kebutuhan Anda)
		categories := []interface{}{
			KategoriPengajuan{
				ID:           primitive.NewObjectID(),
				NamaKategori: "Izin",
			},
			KategoriPengajuan{
				ID:           primitive.NewObjectID(),
				NamaKategori: "Cuti",
			},
		}

		// InsertMany bisa gagal kalau migration dijalankan ulang.
		// Namun migration manager Anda biasanya track applied version,
		// jadi ini aman selama version unik.
		_, err = collection.InsertMany(ctx, categories)
		if err != nil {
			return fmt.Errorf("failed to insert kategori_pengajuan seed: %w", err)
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("kategori_pengajuan").Drop(ctx)
	}

	return version, name, description, up, down
}
