// pkg/migration/migrations/010_create_pengajuan_izin_cuti.go
package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreatePengajuanIzinCuti() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 10
	name := "create_leave_request"
	description := "Create leave_request collection and seed sample requests"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("leave_request") // ✅ renamed

		// Indexes
		indexModels := []mongo.IndexModel{
			{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetName("idx_user_id")},
			{Keys: bson.D{{Key: "request_type_id", Value: 1}}, Options: options.Index().SetName("idx_request_type_id")}, // ✅ renamed
			{Keys: bson.D{{Key: "kepala_departemen_id", Value: 1}}, Options: options.Index().SetName("idx_kepala_departemen_id")},
			{Keys: bson.D{{Key: "manager_hr_id", Value: 1}}, Options: options.Index().SetName("idx_manager_hr_id")},
			{Keys: bson.D{{Key: "status_kepala_departemen", Value: 1}}, Options: options.Index().SetName("idx_status_kepala_departemen")},
			{Keys: bson.D{{Key: "status_manager_hr", Value: 1}}, Options: options.Index().SetName("idx_status_manager_hr")},
			{Keys: bson.D{{Key: "final_status", Value: 1}}, Options: options.Index().SetName("idx_final_status")}, // ✅ renamed
			{Keys: bson.D{{Key: "start_date", Value: 1}}, Options: options.Index().SetName("idx_start_date")},       // ✅ renamed
			{Keys: bson.D{{Key: "end_date", Value: 1}}, Options: options.Index().SetName("idx_end_date")},           // ✅ renamed

			// Optional unique guard (prevent duplicated request on same range)
			{
				Keys: bson.D{
					{Key: "user_id", Value: 1},
					{Key: "request_type_id", Value: 1}, // ✅ renamed
					{Key: "start_date", Value: 1},      // ✅ renamed
					{Key: "end_date", Value: 1},        // ✅ renamed
				},
				Options: options.Index().
					SetUnique(true).
					SetName("uniq_user_type_date_range"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// ================
		// Seeder (optional)
		// ================
		type userMin struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		cur, err := db.Collection("users").Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(2))
		if err != nil {
			return fmt.Errorf("failed to query users for seed: %w", err)
		}
		var users []userMin
		if err := cur.All(ctx, &users); err != nil {
			return fmt.Errorf("failed to decode users for seed: %w", err)
		}
		if len(users) == 0 {
			return nil
		}

		requesterID := users[0].ID
		approverID := users[0].ID
		if len(users) > 1 {
			approverID = users[1].ID
		}

		// ambil salah satu request_type untuk contoh
		var tipe models.RequestType
		err = db.Collection("request_type").
			FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"type_name": 1})).
			Decode(&tipe)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return fmt.Errorf("failed to get request_type for seed: %w", err)
		}

		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day()-2, 8, 0, 0, 0, now.Location())
		end := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, now.Location())

		seed := []interface{}{
			models.LeaveRequest{
				ID:                     primitive.NewObjectID(),
				UserID:                 requesterID,
				RequestTypeID:          tipe.ID,
				TypeName:               tipe.TypeName,
				StartDate:              start,
				EndDate:                end,
				DaysTotal:              3,
				Reason:                 "Contoh pengajuan untuk testing persetujuan izin/cuti.",
				DocumentURL:            "",
				LeaveBalanceID:         nil,
				StatusKepalaDepartemen: models.StatusPending,
				KepalaDepartemenID:     approverID,
				ManagerHRID:            approverID,
				StatusManagerHR:        models.StatusPending,
				FinalStatus:            models.StatusPending,
				CreatedAt:              now,
				UpdatedAt:              now,
			},
		}

		_, err = collection.InsertMany(ctx, seed)
		if err != nil {
			return fmt.Errorf("failed to insert leave_request seed: %w", err)
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("leave_request").Drop(ctx) // ✅ renamed
	}

	return version, name, description, up, down
}