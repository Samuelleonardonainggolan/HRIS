package repository

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BreakTimeRepository interface {
	Create(ctx context.Context, breakTime *models.BreakTime) error
	FindTodayByUserID(ctx context.Context, userID string) (*models.BreakTime, error)
	FindActiveTodayByUserID(ctx context.Context, userID string) (*models.BreakTime, error)
	FindByUserIDAndMonth(ctx context.Context, userID string, year, month int) ([]models.BreakTime, error)
	UpdateEnd(ctx context.Context, id primitive.ObjectID, endTime time.Time, status string) error
}

type breakTimeRepository struct {
	collection *mongo.Collection
}

func NewBreakTimeRepository(db *mongo.Database) BreakTimeRepository {
	return &breakTimeRepository{collection: db.Collection("break_time")}
}

func (r *breakTimeRepository) Create(ctx context.Context, breakTime *models.BreakTime) error {
	if breakTime == nil {
		return errors.New("break time tidak valid")
	}
	now := time.Now().In(wib)
	if breakTime.CreatedAt.IsZero() {
		breakTime.CreatedAt = now
	}
	breakTime.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, breakTime)
	return err
}

func (r *breakTimeRepository) FindTodayByUserID(ctx context.Context, userID string) (*models.BreakTime, error) {
	now := time.Now().In(wib)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, wib)
	endOfDay := startOfDay.Add(24 * time.Hour)

	userFilter, err := buildBreakUserFilter(userID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"$and": bson.A{
			userFilter,
			bson.M{
				"date": bson.M{
					"$gte": startOfDay.UTC(),
					"$lt":  endOfDay.UTC(),
				},
			},
		},
	}

	var breakTime models.BreakTime
	err = r.collection.FindOne(ctx, filter, options.FindOne().SetSort(bson.D{{Key: "start_time", Value: -1}})).Decode(&breakTime)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &breakTime, nil
}

func (r *breakTimeRepository) FindActiveTodayByUserID(ctx context.Context, userID string) (*models.BreakTime, error) {
	now := time.Now().In(wib)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, wib)
	endOfDay := startOfDay.Add(24 * time.Hour)

	userFilter, err := buildBreakUserFilter(userID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"$and": bson.A{
			userFilter,
			bson.M{
				"date": bson.M{
					"$gte": startOfDay.UTC(),
					"$lt":  endOfDay.UTC(),
				},
			},
			bson.M{"status": "ONGOING"},
		},
	}

	var breakTime models.BreakTime
	err = r.collection.FindOne(ctx, filter, options.FindOne().SetSort(bson.D{{Key: "start_time", Value: -1}})).Decode(&breakTime)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &breakTime, nil
}

func (r *breakTimeRepository) FindByUserIDAndMonth(ctx context.Context, userID string, year, month int) ([]models.BreakTime, error) {
	userFilter, err := buildBreakUserFilter(userID)
	if err != nil {
		return nil, err
	}

	startOfMonthWIB := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, wib)
	endOfMonthWIB := startOfMonthWIB.AddDate(0, 1, 0)

	filter := bson.M{
		"$and": bson.A{
			userFilter,
			bson.M{
				"date": bson.M{
					"$gte": startOfMonthWIB.UTC(),
					"$lt":  endOfMonthWIB.UTC(),
				},
			},
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.BreakTime
	if err := cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *breakTimeRepository) UpdateEnd(ctx context.Context, id primitive.ObjectID, endTime time.Time, status string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"end_time": endTime, "status": status, "updated_at": time.Now().In(wib)}})
	return err
}

func buildBreakUserFilter(userID string) (bson.M, error) {
	if userID == "" {
		return nil, errors.New("invalid user ID format")
	}

	if oid, err := primitive.ObjectIDFromHex(userID); err == nil {
		return bson.M{"$or": bson.A{bson.M{"user_id": userID}, bson.M{"user_id": oid}}}, nil
	}

	return bson.M{"user_id": userID}, nil
}
