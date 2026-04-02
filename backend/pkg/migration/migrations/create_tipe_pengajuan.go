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
	name := "create_tipe_pengajuan"
	description := "Create tipe_pengajuan collection and seed initial types"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("tipe_pengajuan")

		// Create indexes
		indexModels := []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "nama_tipe", Value: 1}},
				Options: options.Index().
					SetUnique(true).
					SetName("uniq_nama_tipe"),
			},
			{
				Keys: bson.D{{Key: "kategori_pengajuan_id", Value: 1}},
				Options: options.Index().
					SetName("idx_kategori_pengajuan_id"),
			},
			{
				Keys: bson.D{{Key: "potong_kuota", Value: 1}},
				Options: options.Index().
					SetName("idx_potong_kuota"),
			},
			{
				Keys: bson.D{{Key: "wajib_lampiran", Value: 1}},
				Options: options.Index().
					SetName("idx_wajib_lampiran"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// Fetch kategori_pengajuan: "Izin" dan "Cuti"
		var izin models.KategoriPengajuan
		err = db.Collection("kategori_pengajuan").FindOne(ctx, bson.M{"nama_kategori": "Izin"}).Decode(&izin)
		if err != nil {
			return fmt.Errorf("kategori_pengajuan 'Izin' not found. Make sure migration kategori_pengajuan ran: %w", err)
		}

		var cuti models.KategoriPengajuan
		err = db.Collection("kategori_pengajuan").FindOne(ctx, bson.M{"nama_kategori": "Cuti"}).Decode(&cuti)
		if err != nil {
			return fmt.Errorf("kategori_pengajuan 'Cuti' not found. Make sure migration kategori_pengajuan ran: %w", err)
		}

		// Seed tipe_pengajuan
		types := []interface{}{
			// IZIN
			models.TipePengajuan{
				ID:                  primitive.NewObjectID(),
				NamaTipe:            "Izin Sakit",
				KategoriPengajuanID: izin.ID.Hex(),
				NamaKategori:        izin.NamaKategori,
				PotongKuota:         false,
				WajibLampiran:       true, // surat dokter (umumnya)
			},
			models.TipePengajuan{
				ID:                  primitive.NewObjectID(),
				NamaTipe:            "Izin Pribadi",
				KategoriPengajuanID: izin.ID.Hex(),
				NamaKategori:        izin.NamaKategori,
				PotongKuota:         false,
				WajibLampiran:       false,
			},
			models.TipePengajuan{
				ID:                  primitive.NewObjectID(),
				NamaTipe:            "Izin Khusus",
				KategoriPengajuanID: izin.ID.Hex(),
				NamaKategori:        izin.NamaKategori,
				PotongKuota:         false,
				WajibLampiran:       false,
			},

			// CUTI
			models.TipePengajuan{
				ID:                  primitive.NewObjectID(),
				NamaTipe:            "Cuti Tahunan",
				KategoriPengajuanID: cuti.ID.Hex(),
				NamaKategori:        cuti.NamaKategori,
				PotongKuota:         true,
				WajibLampiran:       false, // sesuai kebutuhan Anda: cuti bisa tanpa lampiran
			},
			models.TipePengajuan{
				ID:                  primitive.NewObjectID(),
				NamaTipe:            "Cuti Khusus",
				KategoriPengajuanID: cuti.ID.Hex(),
				NamaKategori:        cuti.NamaKategori,
				PotongKuota:         false, // bisa true/false tergantung rule
				WajibLampiran:       true,  // contoh: cuti nikah/duka biasanya perlu dokumen
			},
		}

		_, err = collection.InsertMany(ctx, types)
		if err != nil {
			return fmt.Errorf("failed to insert tipe_pengajuan seed: %w", err)
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("tipe_pengajuan").Drop(ctx)
	}

	return version, name, description, up, down
}
