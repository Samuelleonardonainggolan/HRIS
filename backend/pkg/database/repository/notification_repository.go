// pkg/database/repository/notification_repository.go
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

type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) error
	FindByID(ctx context.Context, id string) (*models.Notification, error)
	FindByUserID(ctx context.Context, userID string, limit int) ([]models.Notification, error)
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
}

type notificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) NotificationRepository {
	return &notificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *notificationRepository) Create(ctx context.Context, n *models.Notification) error {
	n.ID = primitive.NewObjectID()
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	n.IsRead = false
	_, err := r.collection.InsertOne(ctx, n)
	return err
}

func (r *notificationRepository) FindByID(ctx context.Context, id string) (*models.Notification, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid notification ID")
	}

	var n models.Notification
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&n)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}
	return &n, nil
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID string, limit int) ([]models.Notification, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": uid}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Notification
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, errors.New("invalid user ID")
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"user_id": uid, "is_read": false})
	return int(count), err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid notification ID")
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"is_read": true, "updated_at": time.Now()}})
	return err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	_, err = r.collection.UpdateMany(ctx, bson.M{"user_id": uid, "is_read": false}, bson.M{"$set": bson.M{"is_read": true, "updated_at": time.Now()}})
	return err
}

func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid notification ID")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
