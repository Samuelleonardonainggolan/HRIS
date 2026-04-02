// pkg/database/repository/work_schedule_repo_mongo.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodrv "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WorkScheduleRepoMongo struct {
	col *mongodrv.Collection
}

func NewWorkScheduleRepoMongo(db *mongodrv.Database) *WorkScheduleRepoMongo {
	return &WorkScheduleRepoMongo{
		col: db.Collection("work_schedules"), // ✅ ganti kalau nama collection Anda beda
	}
}

func (r *WorkScheduleRepoMongo) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateMany(ctx, []mongodrv.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index()},
		{Keys: bson.D{{Key: "attendance_id", Value: 1}}, Options: options.Index()},
		{Keys: bson.D{{Key: "is_active", Value: 1}}, Options: options.Index()},
		// 1 user 1 schedule (opsional tapi recommended)
		{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	return err
}

func (r *WorkScheduleRepoMongo) FindByUserID(ctx context.Context, userID primitive.ObjectID) (*models.WorkSchedule, error) {
	var out models.WorkSchedule
	err := r.col.FindOne(ctx, bson.M{"user_id": userID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongodrv.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *WorkScheduleRepoMongo) UpsertByUserID(ctx context.Context, userID primitive.ObjectID, ws *models.WorkSchedule) error {
	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"user_id":        userID,
			"attendance_id":  ws.AttendanceID,
			"start_time":     ws.StartTime,
			"end_time":       ws.EndTime,
			"work_days":      ws.WorkDays, // ✅ tambahan
			"is_active":      ws.IsActive,
			"updated_at":     now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		update,
		options.Update().SetUpsert(true),
	)
	return err
}