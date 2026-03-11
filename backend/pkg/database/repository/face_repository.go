// pkg/database/repository/face_repository.go
package repository

import (
	"context"
	"errors"
	"math"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FaceRepository struct {
	collection     *mongo.Collection
	userCollection *mongo.Collection
}

func NewFaceRepository(db *mongo.Database) *FaceRepository {
	return &FaceRepository{
		collection:     db.Collection("face_embeddings"),
		userCollection: db.Collection("users"),
	}
}

// Create face embedding
func (r *FaceRepository) Create(face *models.FaceEmbedding) error {
	ctx := context.Background()
	_, err := r.collection.InsertOne(ctx, face)
	return err
}

// Find active face by user ID
func (r *FaceRepository) FindActiveByUserID(userID primitive.ObjectID) (*models.FaceEmbedding, error) {
	ctx := context.Background()
	filter := bson.M{
		"user_id":   userID,
		"is_active": true,
	}

	var face models.FaceEmbedding
	err := r.collection.FindOne(ctx, filter).Decode(&face)
	if err != nil {
		return nil, err
	}
	return &face, nil
}

// FindMatchingFace - Find user by face embedding using cosine similarity
func (r *FaceRepository) FindMatchingFace(embedding []float32, threshold float64) (*models.FaceEmbedding, *models.User, float64, error) {
	ctx := context.Background()

	// Get all active face embeddings
	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, nil, 0, err
	}
	defer cursor.Close(ctx)

	var bestMatch *models.FaceEmbedding
	var highestSimilarity float64 = -1

	for cursor.Next(ctx) {
		var face models.FaceEmbedding
		if err := cursor.Decode(&face); err != nil {
			continue
		}

		// Pastikan face.FaceEmbedding adalah []float32
		// Calculate cosine similarity
		similarity := r.cosineSimilarity(embedding, face.FaceEmbedding)

		if similarity > highestSimilarity {
			highestSimilarity = similarity
			bestMatch = &face
		}
	}

	// Check if similarity meets threshold
	if highestSimilarity < threshold || bestMatch == nil {
		return nil, nil, 0, errors.New("no matching face found")
	}

	// Get user details
	var user models.User
	err = r.userCollection.FindOne(ctx, bson.M{"_id": bestMatch.UserID}).Decode(&user)
	if err != nil {
		return nil, nil, 0, err
	}

	return bestMatch, &user, highestSimilarity, nil
}

// Cosine similarity calculation - menerima []float32
func (r *FaceRepository) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Update face embedding
func (r *FaceRepository) Update(face *models.FaceEmbedding) error {
	ctx := context.Background()
	filter := bson.M{"_id": face.ID}
	update := bson.M{"$set": face}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete face (soft delete)
func (r *FaceRepository) Delete(userID primitive.ObjectID) error {
	ctx := context.Background()
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"is_active": false}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
