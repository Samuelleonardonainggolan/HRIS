// pkg/database/repository/face_embedding_repository.go
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

type FaceEmbeddingRepository interface {
	Create(ctx context.Context, embedding *models.FaceEmbedding) error
	FindByUserID(ctx context.Context, userID string) (*models.FaceEmbedding, error)
	Update(ctx context.Context, userID string, faceEmbedding []float64) error
	Delete(ctx context.Context, userID string) error
	DeactivateByUserID(ctx context.Context, userID string) error
}

type faceEmbeddingRepository struct {
	collection *mongo.Collection
}

func NewFaceEmbeddingRepository(db *mongo.Database) FaceEmbeddingRepository {
	return &faceEmbeddingRepository{
		collection: db.Collection("face_embeddings"),
	}
}

func (r *faceEmbeddingRepository) Create(ctx context.Context, embedding *models.FaceEmbedding) error {
	embedding.ID = primitive.NewObjectID()
	embedding.CreatedAt = time.Now()
	embedding.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, embedding)
	return err
}

func (r *faceEmbeddingRepository) FindByUserID(ctx context.Context, userID string) (*models.FaceEmbedding, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var embedding models.FaceEmbedding
	err = r.collection.FindOne(ctx, bson.M{"user_id": objectID, "is_active": true}).Decode(&embedding)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found, not an error
		}
		return nil, err
	}

	return &embedding, nil
}

func (r *faceEmbeddingRepository) Update(ctx context.Context, userID string, faceEmbedding []float64) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	update := bson.M{
		"$set": bson.M{
			"face_embedding":  faceEmbedding,
			"last_updated_at": time.Now(),
			"updated_at":      time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"user_id": objectID, "is_active": true}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("face embedding not found")
	}

	return nil
}

func (r *faceEmbeddingRepository) Delete(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"user_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("face embedding not found")
	}

	return nil
}

func (r *faceEmbeddingRepository) DeactivateByUserID(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"user_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("face embedding not found")
	}

	return nil
}