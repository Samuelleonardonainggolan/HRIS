// pkg/database/repository/face_embedding_repository.go
package repository

import (
	"context"
	"errors"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FaceEmbeddingRepository interface {
	Create(ctx context.Context, faceEmbedding *models.FaceEmbedding) error

	// existing
	FindByUserID(ctx context.Context, userID string) (*models.FaceEmbedding, error)

	// ✅ new
	FindByID(ctx context.Context, id string) (*models.FaceEmbedding, error)
	FindAll(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.FaceEmbedding, error)
	FindLatestByUserID(ctx context.Context, userID string) (*models.FaceEmbedding, error)

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

// FindByUserID keeps your current behavior: only active embedding
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

// ✅ FindLatestByUserID: ambil embedding terbaru (aktif atau tidak)
func (r *faceEmbeddingRepository) FindLatestByUserID(ctx context.Context, userID string) (*models.FaceEmbedding, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var out models.FaceEmbedding
	err = r.collection.FindOne(
		ctx,
		bson.M{"user_id": uid},
		options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}}),
	).Decode(&out)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

// ✅ FindByID: untuk ambil detail berdasarkan embedding id
func (r *faceEmbeddingRepository) FindByID(ctx context.Context, id string) (*models.FaceEmbedding, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid face embedding ID")
	}

	var out models.FaceEmbedding
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("face embedding not found")
		}
		return nil, err
	}
	return &out, nil
}

// ✅ FindAll: list untuk halaman persetujuan
func (r *faceEmbeddingRepository) FindAll(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.FaceEmbedding, error) {
	cur, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []models.FaceEmbedding
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
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