package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeviceTokenRepository interface {
	Save(ctx context.Context, userID string, token string) error
	FindByUserID(ctx context.Context, userID string) ([]string, error)
	Delete(ctx context.Context, token string) error
}

type deviceTokenRepo struct {
	collection *mongo.Collection
}

func NewDeviceTokenRepository(db *mongo.Database) DeviceTokenRepository {
	return &deviceTokenRepo{collection: db.Collection("device_tokens")}
}

func (r *deviceTokenRepo) Save(ctx context.Context, userID string, token string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	doc := bson.M{
		"user_id":    uid,
		"token":      token,
		"created_at": time.Now(),
	}

	// Upsert to avoid duplicates
	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": uid, "token": token}, bson.M{"$set": doc}, nil)
	return err
}

func (r *deviceTokenRepo) FindByUserID(ctx context.Context, userID string) ([]string, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": uid})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	tokens := []string{}
	for cursor.Next(ctx) {
		var d struct {
			Token string `bson:"token"`
		}
		if err := cursor.Decode(&d); err == nil {
			tokens = append(tokens, d.Token)
		}
	}
	return tokens, nil
}

func (r *deviceTokenRepo) Delete(ctx context.Context, token string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"token": token})
	return err
}
