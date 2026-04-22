package repository

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BreakTimeRepository interface {
	Create(ctx context.Context, breakTime *models.BreakTime) error
	FindTodayByUserID(ctx context.Context, userID string) (*models.BreakTime, error)
	FindActiveTodayByUserID(ctx context.Context, userID string) (*models.BreakTime, error)
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
	breakTime.CreatedAt = time.Now()
	breakTime.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, breakTime)
	return err
}

func (r *breakTimeRepository) FindTodayByUserID(ctx context.Context, userID string) (*models.BreakTime, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	now := time.Now().In(wib)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, wib)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := bson.M{
		"user_id": userObjID.Hex(),
		"date": bson.M{
			"$gte": startOfDay.UTC(),
			"$lt":  endOfDay.UTC(),
		},
	}

	var breakTime models.BreakTime
	err = r.collection.FindOne(ctx, filter).Decode(&breakTime)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &breakTime, nil
}

func (r *breakTimeRepository) FindActiveTodayByUserID(ctx context.Context, userID string) (*models.BreakTime, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	now := time.Now().In(wib)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, wib)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := bson.M{
		"user_id": userObjID.Hex(),
		"date": bson.M{
			"$gte": startOfDay.UTC(),
			"$lt":  endOfDay.UTC(),
		},
		"status": "ONGOING",
	}

	var breakTime models.BreakTime
	err = r.collection.FindOne(ctx, filter).Decode(&breakTime)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &breakTime, nil
}

func (r *breakTimeRepository) UpdateEnd(ctx context.Context, id primitive.ObjectID, endTime time.Time, status string) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"end_time": endTime, "status": status, "updated_at": time.Now()}})
	return err
}
