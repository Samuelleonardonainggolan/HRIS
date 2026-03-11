// pkg/database/repository/face_embedding_repository.go
package repository

import (
	"context"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FaceEmbeddingRepository interface {
	Create(ctx context.Context, faceEmbedding *models.FaceEmbedding) error
	FindByUserID(ctx context.Context, userID string) (*models.FaceEmbedding, error)
	Update(ctx context.Context, faceEmbedding *models.FaceEmbedding) error
	Delete(ctx context.Context, userID string) error
}

type faceEmbeddingRepository struct {
	collection *mongo.Collection
}

func NewFaceEmbeddingRepository(db *mongo.Database) FaceEmbeddingRepository {
	return &faceEmbeddingRepository{
		collection: db.Collection("face_embeddings"),
	}
}

func (r *faceEmbeddingRepository) Create(ctx context.Context, faceEmbedding *models.FaceEmbedding) error {
	_, err := r.collection.InsertOne(ctx, faceEmbedding)
	return err
}

func (r *faceEmbeddingRepository) FindByUserID(ctx context.Context, userID string) (*models.FaceEmbedding, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var faceEmbedding models.FaceEmbedding
	err = r.collection.FindOne(ctx, bson.M{"user_id": objID, "is_active": true}).Decode(&faceEmbedding)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &faceEmbedding, nil
}

func (r *faceEmbeddingRepository) Update(ctx context.Context, faceEmbedding *models.FaceEmbedding) error {
	filter := bson.M{"_id": faceEmbedding.ID}
	update := bson.M{"$set": faceEmbedding}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *faceEmbeddingRepository) Delete(ctx context.Context, userID string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": objID}, bson.M{"$set": bson.M{"is_active": false}})
	return err
}
